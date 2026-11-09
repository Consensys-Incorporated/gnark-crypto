// Copyright 2020-2025 Consensys Software Inc.
// Licensed under the Apache License, Version 2.0. See the LICENSE file for details.

package utils

import (
	"fmt"
	"math/bits"
	"runtime"
	"slices"
	"testing"
	"unsafe"
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

// testBitReverseCopyRouting checks BitReverseCopy against the naive kernel on
// both sides of every switch point: tiny, the power of two just under the
// tiling threshold and the one at it, and 2^21 where COBRA takes over for
// non-power-of-two element sizes.
func testBitReverseCopyRouting[T comparable](t *testing.T, gen func(i int) T) {
	elemSize := int(unsafe.Sizeof(*new(T)))
	logAt := bits.Len(uint(bitReverseMinTiledBytes/elemSize - 1)) // smallest 2^k with 2^k·elemSize >= threshold
	for _, logN := range []int{5, logAt - 1, logAt, 21} {
		n := 1 << logN
		src := make([]T, n)
		for i := range src {
			src[i] = gen(i)
		}
		want := make([]T, n)
		bitReverseCopyNaive(want, src)
		dst := make([]T, n)
		for _, nbTasks := range []int{0, 1, 2, 64} {
			clear(dst)
			BitReverseCopy(dst, src, WithNbTasks(nbTasks))
			if !slices.Equal(dst, want) {
				t.Fatalf("2^%d, nbTasks=%d: mismatch", logN, nbTasks)
			}
		}
	}
}

func TestBitReverseCopyRouting(t *testing.T) {
	t.Run("uint32", func(t *testing.T) {
		testBitReverseCopyRouting(t, func(i int) uint32 { return uint32(i) })
	})
	t.Run("uint64", func(t *testing.T) {
		testBitReverseCopyRouting(t, func(i int) uint64 { return uint64(i) })
	})
	t.Run("[3]uint64", func(t *testing.T) {
		testBitReverseCopyRouting(t, func(i int) [3]uint64 { return [3]uint64{uint64(i), 0, uint64(^i)} })
	})
	t.Run("[6]uint32", func(t *testing.T) {
		testBitReverseCopyRouting(t, func(i int) [6]uint32 { return [6]uint32{uint32(i), 0, 0, 0, 0, uint32(^i)} })
	})
	t.Run("[4]uint64", func(t *testing.T) {
		testBitReverseCopyRouting(t, func(i int) [4]uint64 { return [4]uint64{uint64(i), 0, 0, uint64(^i)} })
	})
}

func TestBitReverseUseCobra(t *testing.T) {
	notArm64 := runtime.GOARCH != "arm64"
	for _, tc := range []struct {
		n        int
		elemSize uintptr
		want     bool
	}{
		{1 << 21, 24, notArm64},
		{1 << 21, 32, false},
		{1 << 20, 24, false},
		{1 << 21, 4, false},
	} {
		if got := bitReverseUseCobra(tc.n, tc.elemSize); got != tc.want {
			t.Errorf("bitReverseUseCobra(2^%d, %d) = %v, want %v", bits.TrailingZeros(uint(tc.n)), tc.elemSize, got, tc.want)
		}
	}
}

// BenchmarkBitReverseCopyKernels compares the serial copy kernels; it documents
// the routing in BitReverseCopy. Pin it to a core for stable numbers, e.g.
// taskset -c 2 go test -run '^$' -bench BitReverseCopyKernels -count 5 ./utils/
func BenchmarkBitReverseCopyKernels(b *testing.B) {
	logNs := []int{20, 21, 22, 23}
	benchBitReverseCopyKernels[uint32](b, "uint32", logNs)
	benchBitReverseCopyKernels[uint64](b, "uint64", logNs)
	benchBitReverseCopyKernels[[3]uint64](b, "[3]uint64", logNs)
	benchBitReverseCopyKernels[[4]uint64](b, "[4]uint64", logNs)
}

func benchBitReverseCopyKernels[T any](b *testing.B, name string, logNs []int) {
	logTile := bitReverseLogTile(int(unsafe.Sizeof(*new(T))))
	for _, logN := range logNs {
		src := make([]T, 1<<logN)
		dst := make([]T, 1<<logN)
		b.Run(fmt.Sprintf("%s/2^%d/naive", name, logN), func(b *testing.B) {
			for b.Loop() {
				bitReverseCopyNaive(dst, src)
			}
		})
		b.Run(fmt.Sprintf("%s/2^%d/cobra", name, logN), func(b *testing.B) {
			for b.Loop() {
				bitReverseCopyCobra(dst, src)
			}
		})
		b.Run(fmt.Sprintf("%s/2^%d/tiled1", name, logN), func(b *testing.B) {
			for b.Loop() {
				bitReverseCopyParallel(dst, src, 1, logTile)
			}
		})
	}
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
