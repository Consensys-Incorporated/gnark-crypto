package hash

import (
	"bytes"
	"fmt"
)

type merkleDamgardHasher struct {
	state []byte
	iv    []byte
	f     Compressor
}

// Write implements hash.Write
func (h *merkleDamgardHasher) Write(p []byte) (n int, err error) {
	blockSize := h.f.BlockSize()
	for len(p) != 0 {
		nWritten := blockSize
		if len(p) < blockSize {
			nWritten = len(p)
			if p, err = cloneLeftPadded(p, blockSize); err != nil {
				panic(err) // this should not be possible
			}
		}
		if h.state, err = h.f.Compress(h.state, p[:blockSize]); err != nil {
			return
		}
		n += nWritten
		p = p[blockSize:]
	}
	return
}

func (h *merkleDamgardHasher) Sum(b []byte) []byte {
	return append(b, h.state...)
}

func (h *merkleDamgardHasher) Reset() {
	h.state = bytes.Clone(h.iv)
}

func (h *merkleDamgardHasher) Size() int {
	return h.f.BlockSize()
}

func (h *merkleDamgardHasher) BlockSize() int {
	return h.f.BlockSize()
}

func (h *merkleDamgardHasher) State() []byte {
	return h.state
}

// SetState sets h's state to state. If len(state) > BlockSize, an error is thrown.
// If len(state) < BlockSize, h's state is set to state, and left padded with zeroes.
// len(state) > BlockSize will result in an error.
func (h *merkleDamgardHasher) SetState(state []byte) error {
	var err error
	h.state, err = cloneLeftPadded(state, h.BlockSize())
	return err
}

// NewMerkleDamgardHasher transforms a 2-1 one-way compression function into a
// hash function using a Merkle-Damgard construction. The resulting hash
// function has a block size equal to the block size of compression function.
//
// NB! The construction does not perform explicit padding on the input data. The
// last block of input data is zero-padded to full block size. This means that
// the construction is not collision resistant for generic data as the digest of
// input and input concatenated with zeros (up to the same number of total
// blocks) is same. For collision resistance the caller should perform explicit
// padding on the input data.
//
// - initialState is provided as initial input to the compression
// function. Its preimage should not be known and thus it should be generated
// using a deterministic method.
// If the given initialState is shorter than the hash block size, it will be zero-padded
// on the left. An oversized initialState will cause a panic.
//
// Deprecated: new users should use [NewMerkleDamgardElement] and [NewMerkleDamgardArray]
// consructors for building Merkle-Damgard hash function construction which works
// directly on field elements, avoiding the byte-conversions. This constructor
// may be removed in future versions.
func NewMerkleDamgardHasher(f Compressor, initialState []byte) StateStorer {
	iv, err := cloneLeftPadded(initialState, f.BlockSize())
	if err != nil {
		panic(err)
	}
	return &merkleDamgardHasher{
		f:     f,
		iv:    iv,
		state: bytes.Clone(iv),
	}
}

// cloneLeftPadded copies b into a new byte slice of size n.
// If len(b) < n, it will be padded on the left.
// len(b) > n will result in an error.
func cloneLeftPadded(b []byte, n int) ([]byte, error) {
	if len(b) > n {
		return nil, fmt.Errorf("state/iv must not exceed the hash block size: %d > %d", len(b), n)
	}
	res := make([]byte, n)
	copy(res[n-len(b):], b)
	return res, nil
}
