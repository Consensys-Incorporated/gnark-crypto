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
//
// Deprecated: see also [FieldCompressor] for field-aware compression methods
// which avoids field <> byte-slice conversions and operates directly on field
// elements. This interface may be deprecated together with [NewMerkleDamgardHasher]
// method in future versions.
type Compressor interface {
	// Compress compresses the two inputs into one output. All the inputs and
	// outputs are of the same size, which is the block size. See [BlockSize].
	Compress(left []byte, right []byte) (compressed []byte, err error)
	// BlockSize returns the blocks size.
	BlockSize() int
}

// FieldHasher hashes a sequence of elements of type F into a digest of type D.
// Hashers are mutable and must not be shared between concurrent callers.
type FieldHasher[F, D any] interface {
	// Reset restores the initial state.
	Reset()
	// WriteElement writes one element without constructing a variadic slice.
	WriteElement(F)
	// WriteElements is equivalent to WriteElement on each element. The digest
	// does not depend on how writes are split, only on elements and PadBlock calls.
	// Literal argument lists may allocate through interfaces or generic callers;
	// prefer WriteElement there, or expand an existing slice (h.WriteElements(es...)).
	WriteElements(...F)
	// PadBlock ends the current block, left-padding it with zeros. If there is
	// no incomplete block, it does nothing.
	PadBlock()
	// SumDigest returns the digest without changing the state. An incomplete
	// block is padded as by PadBlock.
	SumDigest() D
	// BlockLen returns the number of F elements in an input block, not bytes.
	BlockLen() int
}

// FieldCompressor is a stateless 2-to-1 one-way function over digests of value
// type D (an element or a fixed-size array of elements). Implementations must
// be safe for concurrent use.
type FieldCompressor[D any] interface {
	// CompressField compresses left and right into dst. dst may alias left or
	// right. Inputs must not be modified, even temporarily, unless they alias dst.
	CompressField(dst, left, right *D)
}
