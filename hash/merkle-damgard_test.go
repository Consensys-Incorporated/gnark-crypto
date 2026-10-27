// Copyright 2020-2026 Consensys Incorporated
// Licensed under the Apache License, Version 2.0. See the LICENSE file for details.

package hash_test

import (
	"testing"

	"github.com/consensys/gnark-crypto/field/koalabear"
	"github.com/consensys/gnark-crypto/hash"
	"github.com/consensys/gnark-crypto/internal/hashtesting"
)

// stubCompressor is a compression function over blocks of any length.
type stubCompressor[D hash.Block[koalabear.Element]] struct{}

// CompressField sets dst[i] = 3·left[i] + right[i].
func (stubCompressor[D]) CompressField(dst, left, right *D) {
	var res D
	for i := range len(res) {
		var t koalabear.Element
		t.Double(&(*left)[i]).Add(&t, &(*left)[i]).Add(&t, &(*right)[i])
		res[i] = t
	}
	*dst = res
}

// byteStub is stubCompressor on the marshalled elements of blocks of type D.
type byteStub[D hash.Block[koalabear.Element]] struct{}

func (byteStub[D]) BlockSize() int {
	var d D
	return len(d) * koalabear.Bytes
}

func (s byteStub[D]) Compress(left, right []byte) ([]byte, error) {
	var x [2]D
	for i, in := range [][]byte{left, right} {
		for j := range len(x[i]) {
			if err := x[i][j].SetBytesCanonical(in[j*koalabear.Bytes : (j+1)*koalabear.Bytes]); err != nil {
				return nil, err
			}
		}
	}
	var dst D
	stubCompressor[D]{}.CompressField(&dst, &x[0], &x[1])
	return digestBytes(dst), nil
}

// codec is the big endian encoding of the elements and digests of koalabear.
type codec[D hash.Block[koalabear.Element]] struct{}

func (codec[D]) ElementSize() int { return koalabear.Bytes }

func (codec[D]) DigestSize() int {
	var d D
	return len(d) * koalabear.Bytes
}

func (codec[D]) DecodeElement(b []byte) (koalabear.Element, error) {
	return koalabear.BigEndian.Element((*[koalabear.Bytes]byte)(b))
}

func (codec[D]) AppendDigest(b []byte, d D) []byte {
	return append(b, digestBytes(d)...)
}

func (c codec[D]) DecodeDigest(b []byte) (D, error) {
	var d D
	var err error
	for i := range len(d) {
		if d[i], err = c.DecodeElement(b[i*koalabear.Bytes : (i+1)*koalabear.Bytes]); err != nil {
			return d, err
		}
	}
	return d, nil
}

func digestBytes[D hash.Block[koalabear.Element]](d D) []byte {
	var b []byte
	for i := range len(d) {
		b = append(b, d[i].Marshal()...)
	}
	return b
}

// TestFieldMerkleDamgardHasher checks the hasher at several block lengths, see
// [hashtesting.CheckMerkleDamgard].
func TestFieldMerkleDamgardHasher(t *testing.T) {
	t.Run("4", testFieldMerkleDamgardHasher[[4]koalabear.Element])
	t.Run("8", testFieldMerkleDamgardHasher[[8]koalabear.Element])
}

func testFieldMerkleDamgardHasher[D hash.Block[koalabear.Element]](t *testing.T) {
	var zero D
	blockLen := len(zero)
	randomElements := func(n int) []koalabear.Element {
		e := make([]koalabear.Element, n)
		koalabear.Vector(e).MustSetRandom()
		return e
	}
	newFieldHasher := func() *hash.FieldMerkleDamgardHasher[koalabear.Element, D] {
		return hash.NewFieldMerkleDamgardHasher[koalabear.Element, D](stubCompressor[D]{}, zero)
	}
	newHasher := func() hash.FieldHasher[koalabear.Element, D] { return newFieldHasher() }
	newByteHasher := func() hash.StateStorer {
		return hash.NewByteHasher[koalabear.Element, D](newFieldHasher(), codec[D]{})
	}
	newRef := func() hash.StateStorer {
		return hash.NewMerkleDamgardHasher(byteStub[D]{}, make([]byte, blockLen*koalabear.Bytes))
	}
	hashtesting.CheckMerkleDamgard[koalabear.Element, *koalabear.Element](t, newHasher, newByteHasher, newRef, digestBytes[D], blockLen, randomElements)
}
