package hash

import "hash"

// StateStorer allows to store and retrieve the state of a hash function.
type StateStorer interface {
	hash.Hash

	// State retrieves the current state of the hash function. Calling this
	// method should not destroy the current state and allow continue the use of
	// the current hasher.
	State() []byte
	// SetState sets the state of the hash function from a previously stored
	// state retrieved using [StateStorer.State] method.
	SetState(state []byte) error
}

// Compressor is a 2-1 one-way function. It takes two inputs and compresses
// them into one output. The inputs and outputs are all of the same size, which
// is the block size. See [BlockSize].
//
// NB! This is lossy compression, meaning that the output is not guaranteed to
// be unique for different inputs. The function must be stateless, meaning that
// the output is guaranteed to be the same for the same inputs.
// It must furthermore not modify its input, even temporarily.
//
// The Compressor is used in the Merkle-Damgard construction to build a hash
// function.
type Compressor interface {
	// Compress compresses the two inputs into one output. All the inputs and
	// outputs are of the same size, which is the block size. See [BlockSize].
	Compress(left []byte, right []byte) (compressed []byte, err error)
	// BlockSize returns the blocks size.
	BlockSize() int
}

// FieldHasher hashes values of type F into a digest of type D.
type FieldHasher[F, D any] interface {
	// WriteField writes the elements to the hasher. The digest depends on the
	// sequence of the elements and of the calls to PadBlock, and not on how the
	// sequence is split among the calls to WriteField.
	WriteField(...F)
	// PadBlock ends the current block, padding it with zeros on the left to a
	// full block. The elements written afterwards start a new block.
	PadBlock()
	// SumField returns the digest of the elements written so far, where an
	// incomplete block is padded as by PadBlock. It does not change the state of
	// the hasher.
	SumField() D
}

// FieldCompressor is a 2-1 one-way function on field elements, or on blocks of
// them. It is the analogue of [Compressor] on elements. D is a pointer to an
// element, or a slice of elements.
type FieldCompressor[D any] interface {
	// CompressField compresses left and right into dst. dst may alias left or
	// right. It must not modify left or right unless they alias dst.
	CompressField(dst, left, right D)
}
