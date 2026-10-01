package vortex

import (
	"fmt"
	"math/big"
	"sync"
	"unsafe"

	"github.com/consensys/gnark-crypto/field/mamabear"
	fext "github.com/consensys/gnark-crypto/field/mamabear/extensions"
	"github.com/consensys/gnark-crypto/parallel"
)

// Proof is an opening proof.
type Proof struct {
	// UAlpha is the random linear combination of the encoded rows.
	UAlpha []fext.E3
	// OpenedColumns is the list of columns that have been opened.
	OpenedColumns [][]mamabear.Element
	// MerkleProofOpenedColumns is the list of Merkle proofs for the opened columns.
	MerkleProofOpenedColumns []MerkleProof
}

// ProverState stores the state of the prover in the Vortex protocol.
type ProverState struct {
	// Params are the parameters provided to the prover.
	Params *Params
	// EncodedMatrix is computed by the prover during the commitment time.
	EncodedMatrix []mamabear.Element
	// HashedColumns are the hashes of the encoded matrix columns.
	HashedColumns []mamabear.Element
	// MerkleTree is the Merkle tree of the hashed columns.
	MerkleTree *MerkleTree
	// Ualpha is the linear combination of the rows of the encoded matrix.
	Ualpha []fext.E3
}

// GetCommitment returns the short commitment to the input matrix.
func (ps *ProverState) GetCommitment() Hash {
	return ps.MerkleTree.Levels[0][0]
}

// Commit returns the commitment to the input matrix (provided row-by-row).
func Commit(p *Params, input [][]mamabear.Element) (*ProverState, error) {
	sizeCodeWord := p.SizeCodeWord()

	// 1. Encode the input matrix.
	codewords := make([]mamabear.Element, len(input)*sizeCodeWord)
	parallel.Execute(len(input), func(start, end int) {
		for i := start; i < end; i++ {
			p.EncodeReedSolomon(input[i], codewords[i*sizeCodeWord:i*sizeCodeWord+sizeCodeWord])
		}
	})

	// 2. Compute the hashes of the encoded matrix (column-wise).
	hashedColumns := transversalHash(codewords, p.Key, p.SizeCodeWord(), p.Conf.columnHash)

	// 3. Compute the Merkle tree leaves by hashing hashedColumns using the Merkle hash function.
	merkleLeaves := make([]Hash, sizeCodeWord)
	sizeBatch := len(hashedColumns) / sizeCodeWord
	nbBytes := mamabear.Bytes
	parallel.Execute(sizeCodeWord, func(start, end int) {
		if p.Conf.merkleHashFunc == nil {
			for i := start; i < end; i++ {
				sStart := sizeBatch * i
				sEnd := sStart + sizeBatch
				merkleLeaves[i] = HashPoseidon2(hashedColumns[sStart:sEnd])
			}
		} else {
			h := p.Conf.merkleHashFunc()
			for i := start; i < end; i++ {
				h.Reset()
				sStart := sizeBatch * i
				sEnd := sStart + sizeBatch
				for j := sStart; j < sEnd; j++ {
					h.Write(hashedColumns[j].Marshal())
				}
				curHash := h.Sum(nil)
				for j := range 4 {
					merkleLeaves[i][j].SetBytes(curHash[nbBytes*j : nbBytes*j+nbBytes])
				}
			}
		}
	})

	return &ProverState{
		Params:        p,
		EncodedMatrix: codewords,
		HashedColumns: hashedColumns,
		MerkleTree:    BuildMerkleTree(merkleLeaves, p.Conf.merkleHashFunc),
	}, nil
}

// OpenLinComb performs the "UAlpha" part of the proof computation.
func (ps *ProverState) OpenLinComb(alpha fext.E3) {

	codewords := ps.EncodedMatrix

	N := ps.Params.SizeCodeWord()
	nbCodewords := len(codewords) / N
	_ualpha := make([]fext.E3, ps.Params.SizeCodeWord())
	var lock sync.Mutex
	parallel.Execute(nbCodewords, func(start, end int) {
		ualpha := make(fext.Vector, ps.Params.SizeCodeWord())
		alphaPow := new(fext.E3).SetOne()
		alphaPow.Exp(alpha, big.NewInt(int64(start)))
		for i := start; i < end; i++ {
			ualpha.MulAccByElement(codewords[i*N:i*N+N], alphaPow)
			alphaPow.Mul(alphaPow, &alpha)
		}

		// Reinterpret []E3 as []mamabear.Element for the accumulation (E3 = 3 × mamabear.Element).
		M := len(ualpha) * 3
		vUalpha := mamabear.Vector(unsafe.Slice((*mamabear.Element)(unsafe.Pointer(&ualpha[0])), M))
		_vUalpha := mamabear.Vector(unsafe.Slice((*mamabear.Element)(unsafe.Pointer(&_ualpha[0])), M))

		lock.Lock()
		_vUalpha.Add(_vUalpha, vUalpha)
		lock.Unlock()
	})

	ps.Ualpha = _ualpha
}

// OpenColumns sets the OpenedColumns field of the proof.
func (ps *ProverState) OpenColumns(selectedColumns []int) (*Proof, error) {

	var (
		numSelectedColumns       = len(selectedColumns)
		openedColumns            = make([][]mamabear.Element, numSelectedColumns)
		merkleProofOpenedColumns = make([]MerkleProof, numSelectedColumns)
		encodedMatrix            = ps.EncodedMatrix
		err                      error
	)
	for i, col := range selectedColumns {
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
func getTransposedColumn(codewords []mamabear.Element, col int, sizeCodeWord int) []mamabear.Element {
	colBuffer := make([]mamabear.Element, len(codewords)/sizeCodeWord)
	for row := range colBuffer {
		colBuffer[row] = codewords[row*sizeCodeWord+col]
	}
	return colBuffer
}
