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

type byteHasher[F, D any] struct {
	core                    FieldHasher[F, D]
	codec                   ByteHasherCodec[F, D]
	setState                func(D)
	block                   []byte
	elements                []F
	stateBytes              []byte
	elementSize, digestSize int
}

// NewByteHasher gives core a separate byte-oriented view. Each Write left-pads
// its final byte block with zeros and ends on a field block boundary. Thus its
// digest depends on Write boundaries, unlike WriteElements. This encoding is
// not collision resistant for arbitrary byte strings: callers must provide
// length encoding or domain separation when needed.
//
// The adapter takes exclusive ownership of core, which must initially be at a
// block boundary; callers must not use core separately. setState must replace
// core's chaining state and discard pending input. The codec and callback must
// describe the same core.
// StateStorer is suitable for Merkle-Damgard cores whose complete chaining state
// is their digest; it is not a general snapshot interface for sponge hashers.
func NewByteHasher[F, D any](core FieldHasher[F, D], codec ByteHasherCodec[F, D], setState func(D)) StateStorer {
	elementSize, digestSize := codec.ElementSize(), codec.DigestSize()
	blockLen := core.BlockLen()
	if elementSize <= 0 || digestSize <= 0 || blockLen <= 0 || setState == nil {
		panic("invalid byte hasher codec, block length or state setter")
	}
	return &byteHasher[F, D]{
		core: core, codec: codec, setState: setState,
		block: make([]byte, blockLen*elementSize), elements: make([]F, blockLen),
		stateBytes: make([]byte, digestSize), elementSize: elementSize, digestSize: digestSize,
	}
}

func (h *byteHasher[F, D]) Write(p []byte) (n int, err error) {
	for len(p) != 0 {
		chunk := p[:min(len(p), len(h.block))]
		input := chunk
		if len(chunk) < len(h.block) {
			clear(h.block)
			copy(h.block[len(h.block)-len(chunk):], chunk)
			input = h.block
		}
		// Parse the whole block before mutating core. Invalid encodings leave
		// the rejected block unabsorbed, including its valid prefix.
		for i := range h.elements {
			h.elements[i], err = h.codec.DecodeElement(input[i*h.elementSize : (i+1)*h.elementSize])
			if err != nil {
				return n, err
			}
		}
		h.core.WriteElements(h.elements...)
		h.core.PadBlock()
		n += len(chunk)
		p = p[len(chunk):]
	}
	return n, nil
}

func (h *byteHasher[F, D]) Sum(b []byte) []byte { return h.codec.AppendDigest(b, h.core.SumDigest()) }
func (h *byteHasher[F, D]) Reset()              { h.core.Reset() }
func (h *byteHasher[F, D]) Size() int           { return h.digestSize }
func (h *byteHasher[F, D]) BlockSize() int      { return len(h.block) }
func (h *byteHasher[F, D]) State() []byte       { return h.Sum(nil) }
func (h *byteHasher[F, D]) SetState(b []byte) error {
	if len(b) > h.digestSize {
		return fmt.Errorf("state must not exceed the hash digest size: %d > %d", len(b), h.digestSize)
	}
	clear(h.stateBytes)
	copy(h.stateBytes[h.digestSize-len(b):], b)
	d, err := h.codec.DecodeDigest(h.stateBytes)
	if err != nil {
		return err
	}
	h.setState(d)
	return nil
}
