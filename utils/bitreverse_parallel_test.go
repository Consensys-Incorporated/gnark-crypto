// Copyright 2020-2025 Consensys Software Inc.
// Licensed under the Apache License, Version 2.0. See the LICENSE file for details.

package utils

import (
	"fmt"
	"slices"
	"testing"
)

// testBitReverseParallelKernels checks the parallel kernels against the naive
// permutation for every size, tile size and task count, bypassing the size
// heuristic of the public API.
func testBitReverseParallelKernels[T comparable](t *testing.T, gen func(i int) T, maxLogN int) {
	for logN := 2; logN <= maxLogN; logN++ {
		n := 1 << logN
		src := make([]T, n)
		for i := range src {
			src[i] = gen(i)
		}
		want := slices.Clone(src)
		bitReverseNaive(want)

		for logTile := 1; logTile <= bitReverseMaxLogTile && 2*logTile <= logN; logTile++ {
			for _, tasks := range []int{1, 2, 3, 7, 64} {
				dst := make([]T, n)
				bitReverseCopyParallel(dst, src, tasks, logTile)
				if !slices.Equal(dst, want) {
					t.Fatalf("copy: 2^%d, logTile=%d, tasks=%d: mismatch", logN, logTile, tasks)
				}

				got := slices.Clone(src)
				bitReversePermuteParallel(got, tasks, logTile)
				if !slices.Equal(got, want) {
					t.Fatalf("in-place: 2^%d, logTile=%d, tasks=%d: mismatch", logN, logTile, tasks)
				}
			}
		}
	}
}

func TestBitReverseParallelKernels(t *testing.T) {
	t.Run("uint32", func(t *testing.T) {
		testBitReverseParallelKernels(t, func(i int) uint32 { return uint32(i) }, 14)
	})
	t.Run("[6]uint32", func(t *testing.T) {
		testBitReverseParallelKernels(t, func(i int) [6]uint32 { return [6]uint32{uint32(i), uint32(^i)} }, 12)
	})
	t.Run("[4]uint64", func(t *testing.T) {
		testBitReverseParallelKernels(t, func(i int) [4]uint64 { return [4]uint64{uint64(i), 0, 0, uint64(^i)} }, 12)
	})
}

func testBitReverseWithNbTasks[T comparable](t *testing.T, gen func(i int) T, logNs []int) {
	for _, logN := range logNs {
		src := make([]T, 1<<logN)
		for i := range src {
			src[i] = gen(i)
		}
		want := slices.Clone(src)
		bitReverseNaive(want)
		for _, nbTasks := range []int{0, 1, 2, 5, 16} {
			dst := make([]T, len(src))
			BitReverseCopy(dst, src, WithNbTasks(nbTasks))
			if !slices.Equal(dst, want) {
				t.Fatalf("BitReverseCopy: 2^%d, nbTasks=%d: mismatch", logN, nbTasks)
			}
			got := slices.Clone(src)
			BitReverse(got, WithNbTasks(nbTasks))
			if !slices.Equal(got, want) {
				t.Fatalf("BitReverse: 2^%d, nbTasks=%d: mismatch", logN, nbTasks)
			}
		}
	}
}

func TestBitReverseWithNbTasks(t *testing.T) {
	logNs := []int{0, 1, 4, 10, 17, 18, 21}
	if !testing.Short() {
		logNs = append(logNs, 22)
	}
	t.Run("uint32", func(t *testing.T) {
		testBitReverseWithNbTasks(t, func(i int) uint32 { return uint32(i) }, logNs)
	})
	t.Run("[4]uint64", func(t *testing.T) {
		testBitReverseWithNbTasks(t, func(i int) [4]uint64 { return [4]uint64{uint64(i), 0, 0, uint64(^i)} }, logNs)
	})
}

func BenchmarkBitReverseParallel(b *testing.B) {
	benchBitReverseParallel[uint32](b, "uint32", []int{20, 22, 24})
	benchBitReverseParallel[[6]uint32](b, "[6]uint32", []int{20, 22})
	benchBitReverseParallel[[4]uint64](b, "[4]uint64", []int{20, 22})
}

func benchBitReverseParallel[T any](b *testing.B, name string, logNs []int) {
	for _, logN := range logNs {
		src := make([]T, 1<<logN)
		dst := make([]T, 1<<logN)
		for _, nbTasks := range []int{1, 8, 32, 0} {
			label := fmt.Sprint(nbTasks)
			if nbTasks == 0 {
				label = "max"
			}
			b.Run(fmt.Sprintf("%s/copy/2^%d/tasks=%s", name, logN, label), func(b *testing.B) {
				for b.Loop() {
					BitReverseCopy(dst, src, WithNbTasks(nbTasks))
				}
			})
			b.Run(fmt.Sprintf("%s/in-place/2^%d/tasks=%s", name, logN, label), func(b *testing.B) {
				for b.Loop() {
					BitReverse(src, WithNbTasks(nbTasks))
				}
			})
		}
	}
}
