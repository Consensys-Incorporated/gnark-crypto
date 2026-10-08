package vortex

import (
	"fmt"
	"math/big"
	"sync"
	"unsafe"

	"github.com/consensys/gnark-crypto/field/koalabear"
	fext "github.com/consensys/gnark-crypto/field/koalabear/extensions"
	"github.com/consensys/gnark-crypto/parallel"
)

// Proof is an opening proof
type Proof struct {
	// UAlpha is the random linear combination of the encoded rows
	// of the committed matrix.
	UAlpha []fext.E4
	// OpenedColumns is the list of columns that have been opened
	OpenedColumns [][]koalabear.Element
	// MerkleProof is the list of the Merkle-Proofs for the opened columns
	MerkleProofOpenedColumns []MerkleProof
}

// ProverState stores the state of the prover in the Vortex protocol
// and tracks the internal values.
type ProverState struct {
	// Params are the parameters provided to the prover to commit.
	Params *Params
	// EncodedMatrix is computed by the prover during the commitment
	// time.
	EncodedMatrix []koalabear.Element
	// SisHashes are the SIS hashes of the encoded matrix
	HashedColumns []koalabear.Element
	// MerkleTree is the Merkle tree of the SIS hashes
	MerkleTree *MerkleTree
	// Ualpha is the linear combination of the rows of the encoded matrix
	Ualpha []fext.E4
}

// GetCommitment returns the short commitment to the input matrix
func (ps *ProverState) GetCommitment() Hash {
	return ps.MerkleTree.Levels[0][0]
}

// CommitSis returns the commitment to the input matrix. The
// matrix is provided row-by-row in the input.
func Commit(p *Params, input [][]koalabear.Element) (*ProverState, error) {
	sizeCodeWord := p.SizeCodeWord()

	// 1. Encode the input matrix
	codewords := make([]koalabear.Element, len(input)*sizeCodeWord)
	parallel.Execute(len(input), func(start, end int) {
		for i := start; i < end; i++ {
			p.EncodeReedSolomon(input[i], codewords[i*sizeCodeWord:i*sizeCodeWord+sizeCodeWord])
		}
	})

	// 2. Compute the hashes of the encoded matrix (column-wise). By default, the hash function that is used is SIS.
	hashedColumns := transversalHash(codewords, p.Key, p.SizeCodeWord(), p.Conf.columnHash)

	// 3. Compute the Merkle tree of the SIS hashes using Poseidon2, or the provided hash if needed.
	merkleLeaves := make([]Hash, sizeCodeWord)

	if p.Conf.merkleHashFunc == nil { // in this case, we use poseidon2
		const blockSize = 16
		// if for hashing the columns, we did not use poseidon, then keySize should be interpreted
		// as 8, because in that case, the hashes of the columns are on 32bytes = 8 koalabear elements.
		var sisKeySize int
		if p.Conf.columnHash != nil {
			sisKeySize = 8
		} else {
			sisKeySize = p.Key.Degree
		}
		if sizeCodeWord%blockSize == 0 {
			// we hash by blocks of 16 to leverage optimized SIMD implementation
			// of Poseidon2 which require 16 hashes to be computed independently.
			parallel.Execute(sizeCodeWord/blockSize, func(start, end int) {
				for block := start; block < end; block++ {
					b := block * blockSize
					sStart := b * sisKeySize
					sEnd := sStart + sisKeySize*blockSize
					HashPoseidon2x16(hashedColumns[sStart:sEnd], merkleLeaves[b:b+blockSize], sisKeySize)
				}
			})
		} else {
			// unusual path; it means we have < 16 columns (tiny code words)
			// so we do the hashes one by one.
			for i := range sizeCodeWord {
				sStart := i * sisKeySize
				sEnd := sStart + sisKeySize
				merkleLeaves[i] = HashPoseidon2(hashedColumns[sStart:sEnd])
			}
		}
	} else {
		// in this case, we split hashedColumns in sizeCodeWord blocks of equal size,
		// and we hash them using the provided hash
		sizeBatch := len(hashedColumns) / sizeCodeWord
		nbBytes := koalabear.Bytes
		parallel.Execute(sizeCodeWord, func(start, end int) {
			h := p.Conf.merkleHashFunc()
			for i := start; i < end; i++ {
				sStart := sizeBatch * i
				sEnd := sStart + sizeBatch
				for j := sStart; j < sEnd; j++ {
					h.Write(hashedColumns[j].Marshal())
				}
				curHash := h.Sum(nil)
				for j := range 8 {
					merkleLeaves[i][j].SetBytes(curHash[nbBytes*j : nbBytes*j+nbBytes])
				}
			}
		})
	}

	return &ProverState{
		Params:        p,
		EncodedMatrix: codewords,
		HashedColumns: hashedColumns,
		MerkleTree:    BuildMerkleTree(merkleLeaves, p.Conf.merkleHashFunc),
	}, nil
}

// OpenLinComb performs the "UAlpha" part of the proof computation.
// UAlpha is computed as uAlpha := \sum_i row_i * alpha^i.
func (ps *ProverState) OpenLinComb(alpha fext.E4) {

	codewords := ps.EncodedMatrix

	// We don't use the Horner algorithm because we can save on fext
	// operations using the naive algorithm.
	N := ps.Params.SizeCodeWord()
	nbCodewords := len(codewords) / N
	_ualpha := make([]fext.E4, ps.Params.SizeCodeWord())
	var lock sync.Mutex
	parallel.Execute(nbCodewords, func(start, end int) {
		ualpha := make(fext.VectorE4, ps.Params.SizeCodeWord())
		alphaPow := new(fext.E4).SetOne()
		alphaPow.Exp(alpha, big.NewInt(int64(start)))
		for i := start; i < end; i++ {
			ualpha.MulAccByElement(codewords[i*N:i*N+N], alphaPow)
			alphaPow.Mul(alphaPow, &alpha)
		}

		// using unsafe, we take the address of _ualpha[0] and
		// create a vector of fr.Element of size M starting at _ualpha[0]
		M := len(ualpha) * 4
		vUalpha := koalabear.Vector(unsafe.Slice((*koalabear.Element)(unsafe.Pointer(&ualpha[0])), M))
		_vUalpha := koalabear.Vector(unsafe.Slice((*koalabear.Element)(unsafe.Pointer(&_ualpha[0])), M))

		lock.Lock()
		_vUalpha.Add(_vUalpha, vUalpha)
		lock.Unlock()
	})

	ps.Ualpha = _ualpha
}

// OpenColumns sets the OpenedColumns field of the proof using the provided
// codewords and selected columns.
func (ps *ProverState) OpenColumns(selectedColumns []int) (*Proof, error) {

	var (
		numSelectedColumns       = len(selectedColumns)
		openedColumns            = make([][]koalabear.Element, numSelectedColumns)
		merkleProofOpenedColumns = make([]MerkleProof, numSelectedColumns)
		encodedMatrix            = ps.EncodedMatrix
		err                      error
	)
	for i, col := range selectedColumns {

		// an error here indicates that the user samples integers that are
		// too large.
		if col >= ps.Params.SizeCodeWord() {
			return nil, fmt.Errorf("column index out of range")
		}
		openedColumns[i] = getTransposedColumn(encodedMatrix, col, ps.Params.SizeCodeWord())
		if merkleProofOpenedColumns[i], err = ps.MerkleTree.Open(col); err != nil {
			return nil, fmt.Errorf("error in merkle proof generation: %w", err)
		}
	}

	return &Proof{
		UAlpha:                   ps.Ualpha,
		OpenedColumns:            openedColumns,
		MerkleProofOpenedColumns: merkleProofOpenedColumns,
	}, nil
}

// getTransposedColumn returns the specified column from the codewords matrix.
// It extracts the column at index 'col' from a 2D slice of koalabear.Elements.
func getTransposedColumn(codewords []koalabear.Element, col int, sizeCodeWord int) []koalabear.Element {
	// Create a buffer to store the column elements
	colBuffer := make([]koalabear.Element, len(codewords)/sizeCodeWord)

	// Iterate over each row and extract the element from the specified column
	for row := range colBuffer {
		colBuffer[row] = codewords[row*sizeCodeWord+col]
	}

	// Return the extracted column as a slice
	return colBuffer
}
