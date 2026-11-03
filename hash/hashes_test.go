// Copyright 2020-2026 Consensys Software Inc.
// Licensed under the Apache License, Version 2.0. See the LICENSE file for details.

package hash

import (
	"testing"

	"github.com/stretchr/testify/require"
)

// stubHasher is a field hasher whose digest is a given value.
type stubHasher[F, D any] struct{ digest D }

func (stubHasher[F, D]) Reset()             {}
func (stubHasher[F, D]) WriteElement(F)     {}
func (stubHasher[F, D]) WriteElements(...F) {}
func (stubHasher[F, D]) PadBlock()          {}
func (s stubHasher[F, D]) SumDigest() D     { return s.digest }
func (stubHasher[F, D]) BlockLen() int      { return 1 }

func registerStub[F, D any](r fieldHasherRegistry, name string, digest D) {
	registerFieldHasher(r, name, func() FieldHasher[F, D] { return stubHasher[F, D]{digest} })
}

func TestFieldHasherRegistered(t *testing.T) {
	r := make(fieldHasherRegistry)
	registerStub[int, int](r, "TEST_REGISTERED", 7)

	h, err := newFieldHasher[int, int](r, "TEST_REGISTERED")
	require.NoError(t, err)
	require.Equal(t, 7, h.SumDigest())
}

func TestFieldHasherOverTwoDigestTypes(t *testing.T) {
	r := make(fieldHasherRegistry)
	registerStub[int, int](r, "TEST_TWO_DIGESTS", 1)
	registerStub[int, string](r, "TEST_TWO_DIGESTS", "one")

	h, err := newFieldHasher[int, int](r, "TEST_TWO_DIGESTS")
	require.NoError(t, err)
	require.Equal(t, 1, h.SumDigest())

	hs, err := newFieldHasher[int, string](r, "TEST_TWO_DIGESTS")
	require.NoError(t, err)
	require.Equal(t, "one", hs.SumDigest())
}

func TestFieldHasherNotRegistered(t *testing.T) {
	r := make(fieldHasherRegistry)
	registerStub[int, int](r, "TEST_NOT_REGISTERED", 0)

	_, err := newFieldHasher[int, int](r, "TEST_UNKNOWN")
	require.ErrorContains(t, err, "TEST_UNKNOWN")

	_, err = newFieldHasher[int, bool](r, "TEST_NOT_REGISTERED")
	require.ErrorContains(t, err, "TEST_NOT_REGISTERED")
	require.ErrorContains(t, err, "int")
	require.ErrorContains(t, err, "bool")
}

func TestRegisterFieldHasherTwice(t *testing.T) {
	r := make(fieldHasherRegistry)
	registerStub[int, int](r, "TEST_TWICE", 0)
	require.Panics(t, func() { registerStub[int, int](r, "TEST_TWICE", 1) })

	h, err := newFieldHasher[int, int](r, "TEST_TWICE")
	require.NoError(t, err)
	require.Equal(t, 0, h.SumDigest())
}
