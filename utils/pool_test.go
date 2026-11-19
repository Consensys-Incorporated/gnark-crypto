// Copyright 2020-2026 Consensys Incorporated
// Licensed under the Apache License, Version 2.0. See the LICENSE file for details.

package utils

import (
	"runtime"
	"slices"
	"sync"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestPoolMakeDump(t *testing.T) {
	pool := NewPool[uint64]()

	for _, c := range []struct{ n, cap int }{
		{1, 1}, {2, 2}, {3, 4}, {4, 4}, {5, 8}, {7, 8}, {8, 8}, {9, 16}, {1023, 1024}, {1024, 1024}, {1025, 2048},
	} {
		s := pool.Make(c.n)
		require.Len(t, s, c.n)
		require.Equal(t, c.cap, cap(s), "n = %d", c.n)
		pool.Dump(s)
	}

	a := pool.Make(3)
	for i := range a {
		a[i] = uint64(i + 1)
	}
	c := pool.Clone(a)
	require.Equal(t, a, c)
	c[0] = 42
	require.Equal(t, uint64(1), a[0], "Clone must not alias its input")
	pool.Dump(a, c)

	require.Nil(t, pool.Make(0))
	pool.Dump(nil)
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
	// a slice reused after a Dump can be dumped again
	b := pool.Make(4)
	pool.Dump(b)
}

func TestPoolConcurrentMakeDump(t *testing.T) {
	const nbGoroutines, nbIterations, maxN = 8, 500, 256
	pool := NewPool[uint64](maxN)
	pool.StartStats()

	var wg sync.WaitGroup
	for g := range nbGoroutines {
		wg.Go(func() {
			for i := range nbIterations {
				want := slices.Repeat([]uint64{uint64(g)<<32 | uint64(i)}, 1+(g+i)%maxN)
				s := pool.Make(len(want))
				copy(s, want)
				runtime.Gosched()
				if !slices.Equal(s, want) {
					t.Errorf("goroutine %d, iteration %d: slice overwritten", g, i)
				}
				pool.Dump(s)
			}
		})
	}
	wg.Wait()

	used := 0
	for _, s := range pool.stats.subPools {
		require.Zero(t, s.InUse)
		used += s.Used
	}
	require.Equal(t, nbGoroutines*nbIterations, used)
	require.Empty(t, pool.stats.allocatedFor)
}
