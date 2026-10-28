// Copyright 2020-2026 Consensys Software Inc.
// Licensed under the Apache License, Version 2.0. See the LICENSE file for details.

package hash

// ArrayBlock is the set of supported fixed-size field blocks. Named array
// types are supported. These lengths cover the small-field Poseidon2 widths.
type ArrayBlock[F any] interface {
	~[4]F | ~[6]F | ~[8]F | ~[12]F
}

// MerkleDamgardElement hashes one element per block into an element digest.
// It implements FieldHasher[F, F]. It does not encode the message length or
// provide domain separation. Instances are mutable and not concurrency-safe.
// The zero value must not be used; construct it with NewMerkleDamgardElement.
type MerkleDamgardElement[F any] struct {
	f                FieldCompressor[F]
	iv, state, block F
}

// NewMerkleDamgardElement constructs an element-block hash with chaining state
// iv. Reset restores iv. The compressor must satisfy FieldCompressor's aliasing
// contract: the destination aliases the left input during absorption.
func NewMerkleDamgardElement[F any](f FieldCompressor[F], iv F) *MerkleDamgardElement[F] {
	return &MerkleDamgardElement[F]{f: f, iv: iv, state: iv}
}

func (h *MerkleDamgardElement[F]) Reset()        { h.state = h.iv }
func (h *MerkleDamgardElement[F]) BlockLen() int { return 1 }

// PadBlock does nothing: each element is a complete block.
func (h *MerkleDamgardElement[F]) PadBlock()    {}
func (h *MerkleDamgardElement[F]) SumDigest() F { return h.state }
func (h *MerkleDamgardElement[F]) WriteElement(e F) {
	// Keep the input in the object so an indirect compression call cannot
	// force an allocation of a per-call local element.
	h.block = e
	h.f.CompressField(&h.state, &h.state, &h.block)
}
func (h *MerkleDamgardElement[F]) WriteElements(es ...F) {
	for i := range es {
		h.WriteElement(es[i])
	}
}

// SetStateDigest replaces the chaining state. It does not change the initial
// state restored by Reset. This is not a snapshot of the absorbed message.
func (h *MerkleDamgardElement[F]) SetStateDigest(state F) { h.state = state }

// MerkleDamgardArray hashes array-sized blocks of elements into an array digest.
// It implements FieldHasher[F, D]. Incomplete blocks are padded with zeros on
// the left; callers must encode length/domain separation when this ambiguity
// matters. Instances are mutable and not concurrency-safe, including SumDigest.
// The zero value must not be used; construct it with NewMerkleDamgardArray.
type MerkleDamgardArray[F any, D ArrayBlock[F]] struct {
	f              FieldCompressor[D]
	iv, state, buf D
	n              int
	// Interface-visible temporaries belong to the object, avoiding escaping
	// local arrays in SumDigest and PadBlock. They are not chaining state.
	padded, result D
}

// NewMerkleDamgardArray constructs an array-block hash with chaining state iv.
// Each input block and the digest contain len(iv) elements. Reset restores iv.
func NewMerkleDamgardArray[F any, D ArrayBlock[F]](f FieldCompressor[D], iv D) *MerkleDamgardArray[F, D] {
	return &MerkleDamgardArray[F, D]{f: f, iv: iv, state: iv}
}
func (h *MerkleDamgardArray[F, D]) Reset()        { h.state = h.iv; h.n = 0 }
func (h *MerkleDamgardArray[F, D]) BlockLen() int { return len(h.buf) }
func (h *MerkleDamgardArray[F, D]) WriteElement(e F) {
	h.buf[h.n] = e
	h.n++
	if h.n == len(h.buf) {
		h.f.CompressField(&h.state, &h.state, &h.buf)
		h.n = 0
	}
}
func (h *MerkleDamgardArray[F, D]) WriteElements(es ...F) {
	for len(es) != 0 {
		n := min(len(h.buf)-h.n, len(es))
		// Union-of-array type parameters permit indexing but not slicing.
		for i := range n {
			h.buf[h.n+i] = es[i]
		}
		h.n += n
		es = es[n:]
		if h.n == len(h.buf) {
			h.f.CompressField(&h.state, &h.state, &h.buf)
			h.n = 0
		}
	}
}
func (h *MerkleDamgardArray[F, D]) pad() {
	var zero F
	for i := range len(h.padded) - h.n {
		h.padded[i] = zero
	}
	for i := range h.n {
		h.padded[len(h.padded)-h.n+i] = h.buf[i]
	}
}

// PadBlock commits the incomplete block, left-padding it with zero elements.
// With no pending elements it does nothing.
func (h *MerkleDamgardArray[F, D]) PadBlock() {
	if h.n == 0 {
		return
	}
	h.pad()
	h.f.CompressField(&h.state, &h.state, &h.padded)
	h.n = 0
}

// SumDigest returns the padded digest without changing the chaining state or
// pending input. Only internal scratch storage is changed.
func (h *MerkleDamgardArray[F, D]) SumDigest() D {
	if h.n == 0 {
		return h.state
	}
	h.pad()
	h.f.CompressField(&h.result, &h.state, &h.padded)
	return h.result
}

// SetStateDigest replaces the chaining state and discards pending input. It
// does not change the initial state restored by Reset. A digest alone cannot
// restore an unfinished block; callers must commit it with PadBlock first.
func (h *MerkleDamgardArray[F, D]) SetStateDigest(state D) { h.state = state; h.n = 0 }
