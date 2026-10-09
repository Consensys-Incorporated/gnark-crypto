// Copyright 2020-2026 Consensys Software Inc.
// Licensed under the Apache License, Version 2.0. See the LICENSE file for details.

package utils

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestPoolMakeDump(t *testing.T) {
	pool := NewPool[uint64](4, 16)

	a := pool.Make(3)
	require.Len(t, a, 3)
	b := pool.Make(10)
	require.Len(t, b, 10)
	require.GreaterOrEqual(t, cap(b), 10)

	for i := range a {
		a[i] = uint64(i + 1)
	}
	c := pool.Clone(a)
	require.Equal(t, a, c)
	c[0] = 42
	require.Equal(t, uint64(1), a[0], "Clone must not alias its input")

	pool.Dump(a, b, c)
}

func TestPoolDumpForeignSlicePanics(t *testing.T) {
	pool := NewPool[int](8)
	require.Panics(t, func() { pool.Dump(make([]int, 4)) })
}

func TestPoolDumpTwicePanics(t *testing.T) {
	pool := NewPool[int](8)
	a := pool.Make(4)
	pool.Dump(a)
	require.Panics(t, func() { pool.Dump(a) })
}

func TestPoolReusesDumpedSlice(t *testing.T) {
	pool := NewPool[int](8)
	a := pool.Make(4)
	pool.Dump(a)
	// re-allocating the same backing array after a Dump must not trip the
	// "re-allocated non-dumped slice" check
	b := pool.Make(4)
	pool.Dump(b)
}
