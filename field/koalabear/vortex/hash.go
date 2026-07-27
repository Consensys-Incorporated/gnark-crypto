package vortex

import (
	"github.com/consensys/gnark-crypto/field/koalabear"
	"github.com/consensys/gnark-crypto/field/koalabear/poseidon2"
)

var (
	// compressPerm stores the parameters of the poseidon2 permutation
	// that we use for merkle trees.
	compressPerm = poseidon2.NewPermutation(16, 6, 21)
	// spongePerm stores the parameters of the poseidon2 permutation
	// that we use for the sponge construction.
	spongePerm = poseidon2.NewPermutation(24, 6, 21)
)

// CompressPoseidon2 runs the Poseidon2 compression function over two hashes
func CompressPoseidon2(a, b Hash) Hash {
	res := Hash{}
	var x [16]koalabear.Element
	copy(x[:], a[:])
	copy(x[8:], b[:])

	// Create a buffer to hold the feed-forward input.
	copy(res[:], x[8:])
	if err := compressPerm.Permutation(x[:]); err != nil {
		// can't error (size is correct)
		panic(err)
	}

	for i := range res {
		res[i].Add(&res[i], &x[8+i])
	}
	return res
}

// CompressPoseidon2x16 runs the Poseidon2 compression function on multiple
// inputs in a SIMD fashion.
// The input matrix is expected to be of size 16 * colSize, where each row corresponds to an input.
// The result is stored in the result slice, which is expected to have a size of 16,
// with each element containing 8 elements (the output size of the compression function).
// This function applies a feed-forward mechanism: for each input, the first 8 elements of the input
// are added to the corresponding output, and then it processes chunks of 8 elements from each row in the matrix.
func CompressPoseidon2x16(matrix []koalabear.Element, colSize int, result []Hash) {
	compressPerm.Compressx16(matrix, colSize, result)
}

// CompressPoseidon2x16Columns is the column-major-input variant of
// CompressPoseidon2x16: the matrix is laid out matrix[col*16+lane] (the 16
// lanes of a given column are contiguous), so the AVX-512 kernel loads each
// Poseidon2 rate coordinate with a contiguous VMOVDQU32 instead of a
// VPGATHERDD. Semantics (feed-forward MD chain, output) are identical to
// CompressPoseidon2x16; only the input layout differs.
func CompressPoseidon2x16Columns(matrix []koalabear.Element, colSize int, result []Hash) {
	compressPerm.Compressx16Columns(matrix, colSize, result)
}

// HashPoseidon2 returns a Poseidon2 hash of an array of field elements. The
// input is zero-padded so it should be used only in the context of fixed
// length hashes to avoid padding attacks.
func HashPoseidon2(x []koalabear.Element) Hash {

	const (
		blockSize = 16
		stateSize = 24
	)
	var (
		res   Hash
		state [stateSize]koalabear.Element
	)

	for i := 0; i < len(x); i += blockSize {
		copy(state[len(res):], x[i:])
		spongePerm.Permutation(state[:])
	}

	copy(res[:], state[:])
	return res
}

func HashPoseidon2x16(sisHashes []koalabear.Element, merkleLeaves []Hash, sisKeySize int) {
	const (
		width       = 16
		p2blockSize = 16
		stateSize   = 24
	)
	if len(merkleLeaves) != width {
		panic("invalid input size")
	}

	var state [stateSize][width]koalabear.Element
	for i := 0; i < sisKeySize; i += p2blockSize {
		// transpose state
		for k := 8; k < stateSize; k++ {
			for j := range width {
				state[k][j] = sisHashes[j*sisKeySize+(k-8)+i]
			}
		}
		spongePerm.Permutation16x24(&state)
	}

	// transpose back the first 8 into merkleLeaves
	for k := range 8 {
		for j := range width {
			merkleLeaves[j][k] = state[k][j]
		}
	}
}
