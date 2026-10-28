// Copyright 2020-2026 Consensys Software Inc.
// Licensed under the Apache License, Version 2.0. See the LICENSE file for details.

package hash_test

import (
	"testing"

	gh "github.com/consensys/gnark-crypto/hash"
	"github.com/stretchr/testify/require"
)

//go:noinline
func absorbCore[F, D any](h gh.FieldHasher[F, D], es []F) D {
	h.Reset()
	for _, e := range es {
		h.WriteElement(e)
	}
	_ = h.SumDigest()
	h.PadBlock()
	return h.SumDigest()
}
func TestMerkleDamgardCoreAllocations(t *testing.T) {
	es := []uint64{1, 2, 3, 4, 5}
	element := gh.NewMerkleDamgardElement(elementCompressor{}, uint64(0))
	array := gh.NewMerkleDamgardArray[uint64](arrayCompressor[[4]uint64]{}, [4]uint64{})
	var eh gh.FieldHasher[uint64, uint64] = element
	var ah gh.FieldHasher[uint64, [4]uint64] = array
	require.Zero(t, testing.AllocsPerRun(100, func() { _ = absorbCore(eh, es) }))
	require.Zero(t, testing.AllocsPerRun(100, func() { _ = absorbCore(ah, es) }))
}
