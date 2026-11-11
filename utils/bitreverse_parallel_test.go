// Copyright 2020-2025 Consensys Incorporated
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

// testBitReverseCobraL1Kernels checks the COBRA-L1 kernels against the naive
// permutation for every size, tile size and task count, bypassing the size
// heuristic of the public API.
func testBitReverseCobraL1Kernels[T comparable](t *testing.T, gen func(i int) T, maxLogN int) {
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
				bitReverseCopyCobraL1(dst, src, tasks, logTile)
				if !slices.Equal(dst, want) {
					t.Fatalf("copy: 2^%d, logTile=%d, tasks=%d: mismatch", logN, logTile, tasks)
				}

				got := slices.Clone(src)
				bitReverseCobraInPlaceL1(got, tasks, logTile)
				if !slices.Equal(got, want) {
					t.Fatalf("in-place: 2^%d, logTile=%d, tasks=%d: mismatch", logN, logTile, tasks)
				}
			}
		}
	}
}

func TestBitReverseCobraL1Kernels(t *testing.T) {
	t.Run("uint32", func(t *testing.T) {
		testBitReverseCobraL1Kernels(t, func(i int) uint32 { return uint32(i) }, 14)
	})
	t.Run("[6]uint32", func(t *testing.T) {
		testBitReverseCobraL1Kernels(t, func(i int) [6]uint32 { return [6]uint32{uint32(i), uint32(^i)} }, 12)
	})
	t.Run("[4]uint64", func(t *testing.T) {
		testBitReverseCobraL1Kernels(t, func(i int) [4]uint64 { return [4]uint64{uint64(i), 0, 0, uint64(^i)} }, 12)
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

func TestWithNbTasksBounds(t *testing.T) {
	// Same convention as fft.WithNbTasks: < 1 means serial, capped at 512.
	for in, want := range map[int]int{-3: 1, 0: 1, 1: 1, 7: 7, 512: 512, 10000: 512} {
		if got := bitReverseNbTasks([]BitReverseOption{WithNbTasks(in)}); got != want {
			t.Errorf("WithNbTasks(%d): %d tasks, want %d", in, got, want)
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

// testBitReverseCopyRouting checks BitReverseCopy and BitReverse against the naive kernel on
// both sides of the tiling threshold (the power of two just under it and the
// one at it), on a tiny slice, and on a large 2^21 slice.
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
				t.Fatalf("copy: 2^%d, nbTasks=%d: mismatch", logN, nbTasks)
			}
			copy(dst, src)
			BitReverse(dst, WithNbTasks(nbTasks))
			if !slices.Equal(dst, want) {
				t.Fatalf("in-place: 2^%d, nbTasks=%d: mismatch", logN, nbTasks)
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

// BenchmarkBitReverseCopyKernels compares the serial copy kernels; it documents
// why BitReverseCopy uses COBRA-L1 on one task. benchE6 mirrors the nested
// struct of a koalabear/babybear E6, so the numbers match the real type.
// (Before the COBRA-L1 kernels copied tile elements by index, flat arrays such
// as [3]uint64 or [6]uint32 hit a store-forwarding stall and favoured COBRA
// with 2^9 tiles; they no longer do.) Pin it to a core for stable numbers, e.g.
// taskset -c 2 go test -run '^$' -bench BitReverseCopyKernels -count 5 ./utils/
func BenchmarkBitReverseCopyKernels(b *testing.B) {
	logNs := []int{20, 21, 22, 23}
	benchBitReverseCopyKernels[uint32](b, "uint32", logNs)
	benchBitReverseCopyKernels[uint64](b, "uint64", logNs)
	benchBitReverseCopyKernels[benchE6](b, "E6", logNs)
	benchBitReverseCopyKernels[[4]uint64](b, "[4]uint64", logNs)
}

type benchE2 struct{ A0, A1 [1]uint32 }
type benchE6 struct{ B0, B1, B2 benchE2 }

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
		b.Run(fmt.Sprintf("%s/2^%d/cobra9", name, logN), func(b *testing.B) {
			for b.Loop() {
				bitReverseCopyCobra(dst, src)
			}
		})
		b.Run(fmt.Sprintf("%s/2^%d/cobraL1", name, logN), func(b *testing.B) {
			for b.Loop() {
				bitReverseCopyCobraL1(dst, src, 1, logTile)
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
		for _, nbTasks := range []int{1, 8, 32, runtime.GOMAXPROCS(0)} {
			label := fmt.Sprint(nbTasks)
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

// TestBitReverseNoAlloc checks that option handling adds no allocation: with or
// without options, the public entry points do not allocate on the naive path
// (small slices). An option taking a pointer to the config would move it to the
// heap on every call.
//
// The COBRA-L1 path (slices of 128 KiB or more) still allocates on every call,
// and this test does not cover it: with one task, 3 allocations per
// BitReverseCopy and 6 per BitReverse (the tile buffers, whose size is not a
// constant, and the bit-reversal table and closures that capture it), and 3
// more per extra task for BitReverseCopy (Go 1.26, amd64).
func TestBitReverseNoAlloc(t *testing.T) {
	v := make([]uint64, 16)
	dst := make([]uint64, 16)
	for name, f := range map[string]func(){
		"BitReverse":                 func() { BitReverse(v) },
		"BitReverse/WithNbTasks":     func() { BitReverse(v, WithNbTasks(4)) },
		"BitReverseCopy":             func() { BitReverseCopy(dst, v) },
		"BitReverseCopy/WithNbTasks": func() { BitReverseCopy(dst, v, WithNbTasks(0)) },
	} {
		if n := testing.AllocsPerRun(100, f); n != 0 {
			t.Errorf("%s: %v allocations per call, want 0", name, n)
		}
	}
}
