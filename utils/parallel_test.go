// Copyright 2020-2026 Consensys Incorporated
// Licensed under the Apache License, Version 2.0. See the LICENSE file for details.

package utils

import (
	"fmt"
	"math"
	"sync"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestWorkerPoolSubmitInvalidMinBlock(t *testing.T) {
	// A closed queue makes an attempted submission panic instead of hanging,
	// so this also checks that validation happens before any job is queued.
	pool := &WorkerPool{chJobs: make(chan job)}
	pool.Stop()

	for _, n := range []int{1, 0, -1} {
		for _, minBlock := range []int{0, -1, math.MinInt} {
			t.Run(fmt.Sprintf("n=%d/minBlock=%d", n, minBlock), func(t *testing.T) {
				require.PanicsWithValue(t, "minBlock must be positive", func() {
					pool.Submit(n, func(int, int) {}, minBlock)
				})
			})
		}
	}
}

func TestWorkerPoolSubmit(t *testing.T) {
	tests := []struct {
		name     string
		n        int
		minBlock int
		want     [][2]int
	}{
		{name: "empty", n: 0, minBlock: 1},
		{name: "negative n", n: -1, minBlock: 1},
		{name: "single item", n: 1, minBlock: 1, want: [][2]int{{0, 1}}},
		{name: "unit blocks", n: 3, minBlock: 1, want: [][2]int{{0, 1}, {1, 2}, {2, 3}}},
		{name: "exact blocks", n: 6, minBlock: 2, want: [][2]int{{0, 2}, {2, 4}, {4, 6}}},
		{name: "partial block", n: 7, minBlock: 3, want: [][2]int{{0, 3}, {3, 6}, {6, 7}}},
		{name: "oversized block", n: 3, minBlock: 10, want: [][2]int{{0, 3}}},
	}

	pool := NewWorkerPool()
	t.Cleanup(pool.Stop)

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			var mu sync.Mutex
			var got [][2]int
			wg := pool.Submit(tc.n, func(start, end int) {
				mu.Lock()
				got = append(got, [2]int{start, end})
				mu.Unlock()
			}, tc.minBlock)
			require.NotNil(t, wg)
			wg.Wait()
			require.ElementsMatch(t, tc.want, got)
		})
	}
}
