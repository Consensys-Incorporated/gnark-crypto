package vortex

import (
	"errors"
	"hash"

	"github.com/consensys/gnark-crypto/field/mamabear"
	"github.com/consensys/gnark-crypto/parallel"
)

// Hash represents a hash as they occur in Merkle trees.
// [4]mamabear.Element = 4 * 8 bytes = 32 bytes (matches SHA-256 output size).
type Hash = [4]mamabear.Element

// MerkleTree represents a Merkle tree.
type MerkleTree struct {
	// Levels collects the nodes of the tree in descending order:
	// The first level has a size of 1 and stores the root of the tree.
	// The last level has a size of 1 << Depth and stores the leaves.
	Levels [][]Hash
}

// MerkleProof is a Merkle proof for membership in the tree.
// It is a list of hashes in ascending order.
type MerkleProof []Hash

// hashNodes computes h(left || right), interpreted as 4 mamabear elements (32 bytes).
func hashNodes(h hash.Hash, left, right [4]mamabear.Element) [4]mamabear.Element {
	h.Reset()
	var res [4]mamabear.Element
	for i := range 4 {
		h.Write(left[i].Marshal())
	}
	for i := range 4 {
		h.Write(right[i].Marshal())
	}
	s := h.Sum(nil)
	for i := range 4 {
		res[i].SetBytes(s[mamabear.Bytes*i : mamabear.Bytes*(i+1)])
	}
	return res
}

// BuildMerkleTree builds a Merkle tree from a list of hashes.
// If the number of leaves is not a power of two, leaves are padded with zero hashes.
// When altHash is nil, poseidon2 is used by default.
func BuildMerkleTree(hashes []Hash, altHash HashConstructor) *MerkleTree {

	var (
		numLeaves    = len(hashes)
		newPow2      = nextPowerOfTwo(numLeaves)
		depth        = log2Ceil(numLeaves)
		paddedHashes = hashes
	)

	if len(hashes) != newPow2 {
		paddedHashes = make([]Hash, newPow2)
		copy(paddedHashes, hashes)
	}

	levels := make([][]Hash, depth+1)
	for i := depth; i >= 0; i-- {
		if i == depth {
			levels[i] = paddedHashes
			continue
		}

		levels[i] = make([]Hash, newPow2>>(depth-i))
		if altHash == nil {
			if len(levels[i]) >= 512 {
				parallel.Execute(len(levels[i]), func(start, end int) {
					for k := start; k < end; k++ {
						left, right := levels[i+1][2*k], levels[i+1][2*k+1]
						levels[i][k] = CompressPoseidon2(left, right)
					}
				})
			} else {
				for k := range levels[i] {
					left, right := levels[i+1][2*k], levels[i+1][2*k+1]
					levels[i][k] = CompressPoseidon2(left, right)
				}
			}
		} else {
			if len(levels[i]) >= 512 {
				parallel.Execute(len(levels[i]), func(start, end int) {
					h := altHash()
					for k := start; k < end; k++ {
						left, right := levels[i+1][2*k], levels[i+1][2*k+1]
						levels[i][k] = hashNodes(h, left, right)
					}
				})
			} else {
				h := altHash()
				for k := range levels[i] {
					left, right := levels[i+1][2*k], levels[i+1][2*k+1]
					levels[i][k] = hashNodes(h, left, right)
				}
			}
		}
	}

	return &MerkleTree{
		Levels: levels,
	}
}

// Open returns the Merkle proof for the element at index i.
func (mt *MerkleTree) Open(i int) (MerkleProof, error) {

	var (
		res       = make(MerkleProof, 0, mt.Depth())
		parentPos = i
		posBound  = 1 << mt.Depth()
	)

	if i >= posBound {
		return nil, errors.New("error: index out of range")
	}

	for level := len(mt.Levels) - 1; level > 0; level-- {
		neighborPos := parentPos ^ 1
		res = append(res, mt.Levels[level][neighborPos])
		parentPos = parentPos >> 1
	}

	if len(res) != mt.Depth() {
		panic("error: incorrect number of hashes")
	}

	return res, nil
}

// Verify checks the validity of a Merkle membership proof.
// When altHash is nil, poseidon2 is used by default.
func (proof MerkleProof) Verify(i int, leaf, root Hash, altHash HashConstructor) error {

	var (
		parentPos = i
		curNode   = leaf
	)

	if altHash != nil {
		nh := altHash()
		for _, h := range proof {
			a, b := curNode, h
			if parentPos&1 == 1 {
				a, b = b, a
			}
			curNode = hashNodes(nh, a, b)
			parentPos = parentPos >> 1
		}
	} else {
		for _, h := range proof {
			a, b := curNode, h
			if parentPos&1 == 1 {
				a, b = b, a
			}
			curNode = CompressPoseidon2(a, b)
			parentPos = parentPos >> 1
		}
	}

	if curNode != root {
		return errors.New("error: invalid proof")
	}

	return nil
}

// Depth returns the depth of the tree. A tree of depth n has 2^n leaves.
func (mt *MerkleTree) Depth() int {
	return len(mt.Levels) - 1
}

// Root returns the root of the tree.
func (mt *MerkleTree) Root() Hash {
	return mt.Levels[0][0]
}

// isPowerOfTwo returns true if n is a power of two.
func isPowerOfTwo[T ~int](n T) bool {
	return n&(n-1) == 0 && n > 0
}

// nextPowerOfTwo returns the next power of two for the given number.
func nextPowerOfTwo[T ~int64 | ~uint64 | ~uintptr | ~int | ~uint](in T) T {
	if in < 0 || uint64(in) > 1<<62 {
		panic("input out of range")
	}
	v := in
	v--
	v |= v >> (1 << 0)
	v |= v >> (1 << 1)
	v |= v >> (1 << 2)
	v |= v >> (1 << 3)
	v |= v >> (1 << 4)
	v |= v >> (1 << 5)
	v++
	return v
}

// log2Floor computes the floored value of Log2.
func log2Floor(a int) int {
	res := 0
	for i := a; i > 1; i = i >> 1 {
		res++
	}
	return res
}

// log2Ceil computes the ceiled value of Log2.
func log2Ceil(a int) int {
	floor := log2Floor(a)
	if a != 1<<floor {
		floor++
	}
	return floor
}

// HashHex returns a hex representation of the hash.
func HashHex(h *Hash) string {
	return "0x" +
		h[0].Text(16) +
		h[1].Text(16) +
		h[2].Text(16) +
		h[3].Text(16)
}
