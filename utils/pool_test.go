// Copyright 2020-2026 Consensys Incorporated
// Licensed under the Apache License, Version 2.0. See the LICENSE file for details.

package utils

import (
	"math"
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

func TestPoolMakeTooLargePanicsRecoverably(t *testing.T) {
	pool := NewPool[uint64]()
	require.Panics(t, func() { pool.Make(math.MaxInt/2 + 2) })
	// the lock was released: the pool is still usable
	require.Len(t, pool.Make(1), 1)
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

func TestPoolDumpWhilePausedUntracks(t *testing.T) {
	pool := NewPool[int]()
	pool.StartTrackingUndumped()
	s := pool.Make(5)
	pool.PauseTrackingUndumped()
	pool.Dump(s)
	require.Empty(t, pool.undumped)
}

func TestPoolConcurrentMakeDump(t *testing.T) {
	const nbGoroutines, nbIterations, maxN = 8, 500, 256
	pool := NewPool[uint64](maxN)
	pool.StartTrackingUndumped()

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

	require.Empty(t, pool.undumped)
}
