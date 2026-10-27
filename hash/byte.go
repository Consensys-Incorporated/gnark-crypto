// Copyright 2020-2026 Consensys Software Inc.
// Licensed under the Apache License, Version 2.0. See the LICENSE file for details.

package hash

import "fmt"

// ByteHasherCodec defines the canonical byte encoding of a field hasher's input
// elements and digest. Decode methods receive exactly ElementSize or DigestSize
// bytes, respectively. AppendDigest appends without modifying its input prefix.
// Codec implementations must not retain input byte slices.
type ByteHasherCodec[F, D any] interface {
	ElementSize() int
	DigestSize() int
	DecodeElement([]byte) (F, error)
	AppendDigest([]byte, D) []byte
	DecodeDigest([]byte) (D, error)
}

type byteHasher[F, D any, C ByteHasherCodec[F, D]] struct {
	core interface {
		FieldHasher[F, D]
		SetStateDigest(D)
	}
	codec                   C
	block                   []byte
	elements                []F
	stateBytes              []byte
	elementSize, digestSize int
}

// NewByteHasher gives core a byte view, a [StateStorer] whose bytes are encoded
// by codec. Each Write left-pads its last block of bytes with zeros and ends on a
// block boundary of core, so the digest depends on the calls to Write, unlike on
// the calls to WriteElements of core. The digest is the state, so State and
// SetState round-trip. This encoding is not collision resistant on arbitrary
// bytes: callers must encode the length or separate domains where it matters.
//
// The view takes ownership of core, which must be at a block boundary: callers
// must not use core separately. SetStateDigest must replace the chaining state of
// core and discard its pending elements, so NewByteHasher suits Merkle-Damgard
// hashers, whose state is their digest, and not sponges.
func NewByteHasher[F, D any, C ByteHasherCodec[F, D]](core interface {
	FieldHasher[F, D]
	SetStateDigest(D)
}, codec C) StateStorer {
	elementSize, digestSize, blockLen := codec.ElementSize(), codec.DigestSize(), core.BlockLen()
	if elementSize <= 0 || digestSize <= 0 || blockLen <= 0 {
		panic("hash: invalid codec or block length of a byte hasher")
	}
	return &byteHasher[F, D, C]{
		core:        core,
		codec:       codec,
		block:       make([]byte, blockLen*elementSize),
		elements:    make([]F, blockLen),
		stateBytes:  make([]byte, digestSize),
		elementSize: elementSize,
		digestSize:  digestSize,
	}
}

func (h *byteHasher[F, D, C]) Write(p []byte) (n int, err error) {
	for len(p) != 0 {
		chunk := p[:min(len(p), len(h.block))]
		input := chunk
		if len(chunk) < len(h.block) {
			clear(h.block)
			copy(h.block[len(h.block)-len(chunk):], chunk)
			input = h.block
		}
		// An invalid element rejects its whole block: none of the block's elements
		// is written to core.
		for i := range h.elements {
			h.elements[i], err = h.codec.DecodeElement(input[i*h.elementSize : (i+1)*h.elementSize])
			if err != nil {
				return n, err
			}
		}
		h.core.WriteElements(h.elements...)
		n += len(chunk)
		p = p[len(chunk):]
	}
	return n, nil
}

func (h *byteHasher[F, D, C]) Sum(b []byte) []byte {
	return h.codec.AppendDigest(b, h.core.Sum())
}

func (h *byteHasher[F, D, C]) Reset() { h.core.Reset() }

func (h *byteHasher[F, D, C]) Size() int { return h.digestSize }

func (h *byteHasher[F, D, C]) BlockSize() int { return len(h.block) }

// State returns the marshalled state, which is the digest.
func (h *byteHasher[F, D, C]) State() []byte { return h.Sum(nil) }

// SetState sets the state to state. If len(state) < Size, it is left padded with
// zeros. A state longer than Size, or not the encoding of a digest, results in an
// error.
func (h *byteHasher[F, D, C]) SetState(state []byte) error {
	if len(state) > h.digestSize {
		return fmt.Errorf("state must not exceed the digest size: %d > %d", len(state), h.digestSize)
	}
	clear(h.stateBytes)
	copy(h.stateBytes[h.digestSize-len(state):], state)
	d, err := h.codec.DecodeDigest(h.stateBytes)
	if err != nil {
		return err
	}
	h.core.SetStateDigest(d)
	return nil
}
