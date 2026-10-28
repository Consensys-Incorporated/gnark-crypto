// Copyright 2020-2026 Consensys Software Inc.
// Licensed under the Apache License, Version 2.0. See the LICENSE file for details.

package hash_test

import (
	"testing"

	gh "github.com/consensys/gnark-crypto/hash"
	"github.com/stretchr/testify/require"
)

// Arithmetic compressors expose block ordering and state transitions without
// relying on a particular cryptographic permutation.
type elementCompressor struct{}

func (elementCompressor) CompressField(dst, left, right *uint64) { *dst = 3*(*left) + *right }

type arrayCompressor[D gh.ArrayBlock[uint64]] struct{}

func (arrayCompressor[D]) CompressField(dst, left, right *D) {
	for i := range len(*dst) {
		(*dst)[i] = 3*(*left)[i] + (*right)[i]
	}
}

type namedBlock [6]uint64

func TestMerkleDamgardElementState(t *testing.T) {
	h := gh.NewMerkleDamgardElement(elementCompressor{}, uint64(7))
	h.WriteElement(2)
	h.WriteElements(4, 5)
	require.Equal(t, uint64(224), h.SumDigest())
	h.PadBlock()
	require.Equal(t, uint64(224), h.SumDigest())
	h.SetStateDigest(11)
	h.WriteElement(6)
	require.Equal(t, uint64(39), h.SumDigest())
	h.Reset()
	h.WriteElement(6)
	require.Equal(t, uint64(27), h.SumDigest())
}

func checkArrayState[D gh.ArrayBlock[uint64]](t *testing.T) {
	t.Helper()
	var iv D
	for i := range len(iv) {
		iv[i] = uint64(i + 1)
	}
	h := gh.NewMerkleDamgardArray[uint64](arrayCompressor[D]{}, iv)
	// A sum must not commit the incomplete block or overwrite its pending data.
	h.WriteElement(5)
	first := h.SumDigest()
	for i := range len(iv) - 1 {
		require.Equal(t, 3*iv[i], first[i])
	}
	require.Equal(t, 3*iv[len(iv)-1]+5, first[len(iv)-1])
	h.WriteElement(7)
	second := h.SumDigest()
	require.Equal(t, 3*iv[len(iv)-2]+5, second[len(iv)-2])
	require.Equal(t, 3*iv[len(iv)-1]+7, second[len(iv)-1])
	require.Equal(t, second, h.SumDigest())
	h.PadBlock()
	h.PadBlock()
	require.Equal(t, second, h.SumDigest())
	// PadBlock establishes the next block boundary rather than merely returning
	// a digest. One full block following it starts from the committed state.
	input := make([]uint64, len(iv))
	for i := range input {
		input[i] = uint64(i + 9)
	}
	h.WriteElements(input...)
	got := h.SumDigest()
	for i := range len(got) {
		require.Equal(t, 3*second[i]+input[i], got[i])
	}
	// State replacement must discard a pending block, and leave the IV intact.
	h.WriteElement(99)
	h.SetStateDigest(iv)
	h.WriteElements(input...)
	got = h.SumDigest()
	for i := range len(got) {
		require.Equal(t, 3*iv[i]+input[i], got[i])
	}
	h.WriteElement(77)
	h.Reset()
	require.Equal(t, iv, h.SumDigest())
	// Bulk writes spanning several blocks agree with individual writes for every
	// boundary, including a tail. Named arrays use the same indexing semantics.
	input = make([]uint64, 3*len(iv)+1)
	for i := range input {
		input[i] = uint64(i + 1)
	}
	for split := range len(input) + 1 {
		h.Reset()
		h.WriteElements(input[:split]...)
		_ = h.SumDigest()
		h.WriteElements(input[split:]...)
		single := gh.NewMerkleDamgardArray[uint64](arrayCompressor[D]{}, iv)
		for _, x := range input {
			single.WriteElement(x)
		}
		require.Equal(t, single.SumDigest(), h.SumDigest(), "split %d", split)
	}
}
func TestMerkleDamgardArrayState(t *testing.T) {
	t.Run("four", checkArrayState[[4]uint64])
	t.Run("six-named", checkArrayState[namedBlock])
	t.Run("eight", checkArrayState[[8]uint64])
	t.Run("twelve", checkArrayState[[12]uint64])
}
