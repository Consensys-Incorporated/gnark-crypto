// Copyright 2020-2026 Consensys Software Inc.
// Licensed under the Apache License, Version 2.0. See the LICENSE file for details.

package poseidon2

import (
	"testing"

	fr "github.com/consensys/gnark-crypto/field/mamabear"
	"github.com/stretchr/testify/require"
)

func TestMulMulInternalInPlaceWidth16(t *testing.T) {
	var input, expected [16]fr.Element
	for i := range input {
		input[i].MustSetRandom()
	}
	expected = input

	h := NewPermutation(16, 6, 21)
	h.matMulInternalInPlace(expected[:])

	var sum fr.Element
	sum.Set(&input[0])
	for i := 1; i < h.params.Width; i++ {
		sum.Add(&sum, &input[i])
	}
	for i := range h.params.Width {
		var got fr.Element
		got.Mul(&input[i], &diag16[i]).Add(&got, &sum)
		if !got.Equal(&expected[i]) {
			t.Fatalf("mat mul internal w/ diagonal mismatch at index %d", i)
		}
	}
}

func TestMulMulInternalInPlaceWidth24(t *testing.T) {
	var input, expected [24]fr.Element
	for i := range input {
		input[i].MustSetRandom()
	}
	expected = input

	h := NewPermutation(24, 6, 21)
	h.matMulInternalInPlace(expected[:])

	var sum fr.Element
	sum.Set(&input[0])
	for i := 1; i < h.params.Width; i++ {
		sum.Add(&sum, &input[i])
	}
	for i := range h.params.Width {
		var got fr.Element
		got.Mul(&input[i], &diag24[i]).Add(&got, &sum)
		if !got.Equal(&expected[i]) {
			t.Fatalf("mat mul internal w/ diagonal mismatch at index %d", i)
		}
	}
}

func TestPermutationWidth16(t *testing.T) {
	assert := require.New(t)
	h := NewPermutation(16, 6, 21)

	var a, b [16]fr.Element
	for i := range a {
		a[i].MustSetRandom()
		b[i].Set(&a[i])
	}

	assert.NoError(h.Permutation(a[:]))
	assert.NoError(h.Permutation(b[:]))

	for i := range a {
		assert.True(a[i].Equal(&b[i]), "permutation is not deterministic at index %d", i)
	}

	// result should differ from zero input
	var zero [16]fr.Element
	assert.NoError(h.Permutation(zero[:]))
	allZero := true
	for i := range zero {
		if !zero[i].IsZero() {
			allZero = false
			break
		}
	}
	assert.False(allZero, "permutation of zero should not be all-zero")
}

func TestPermutationWidth24(t *testing.T) {
	assert := require.New(t)
	h := NewPermutation(24, 6, 21)

	var a, b [24]fr.Element
	for i := range a {
		a[i].MustSetRandom()
		b[i].Set(&a[i])
	}

	assert.NoError(h.Permutation(a[:]))
	assert.NoError(h.Permutation(b[:]))

	for i := range a {
		assert.True(a[i].Equal(&b[i]), "permutation is not deterministic at index %d", i)
	}
}

func TestCompress(t *testing.T) {
	assert := require.New(t)
	h := NewPermutation(16, 6, 21)

	n := h.params.Width / 2
	elems := make([]fr.Element, h.params.Width)
	for i := range elems {
		elems[i].MustSetRandom()
	}
	left := make([]byte, 0, n*fr.Bytes)
	right := make([]byte, 0, n*fr.Bytes)
	for i := range n {
		left = append(left, elems[i].Marshal()...)
	}
	for i := range n {
		right = append(right, elems[n+i].Marshal()...)
	}

	out1, err := h.Compress(left, right)
	assert.NoError(err)
	assert.Equal(h.BlockSize(), len(out1))

	out2, err := h.Compress(left, right)
	assert.NoError(err)
	assert.Equal(out1, out2, "Compress is not deterministic")
}

func BenchmarkPermutationWidth16(b *testing.B) {
	h := NewPermutation(16, 6, 21)
	var tmp [16]fr.Element
	for i := range tmp {
		tmp[i].MustSetRandom()
	}
	b.ResetTimer()
	for range b.N {
		h.Permutation(tmp[:])
	}
}

func BenchmarkPermutationWidth24(b *testing.B) {
	h := NewPermutation(24, 6, 21)
	var tmp [24]fr.Element
	for i := range tmp {
		tmp[i].MustSetRandom()
	}
	b.ResetTimer()
	for range b.N {
		h.Permutation(tmp[:])
	}
}
