// Copyright 2020-2025 Consensys Software Inc.
// Licensed under the Apache License, Version 2.0. See the LICENSE file for details.

package utils

import (
	"fmt"
	"slices"
	"testing"
)

const maxSizeBitReverse = 1 << 22

func TestBitReverse(t *testing.T) {
	sizes := []int{2, 4, 8, 16, 32, 64, 128, 256, 512, maxSizeBitReverse}

	t.Run("uint32", func(t *testing.T) {
		for _, size := range sizes {
			t.Run(fmt.Sprintf("size=%d", size), func(t *testing.T) {
				// check that bit-reversing twice is identity
				original := make([]uint32, size)
				for i := range original {
					original[i] = uint32(i)
				}
				a := make([]uint32, size)
				copy(a, original)

				BitReverse(a)
				BitReverse(a)

				for i := range a {
					if a[i] != original[i] {
						t.Fatalf("bit-reversing twice is not identity")
					}
				}
			})
		}

		// check that it panics for non-power of 2
		for _, size := range []int{3, 5, 6, 7, 9, 10, 12} {
			a := make([]uint32, size)
			assertPanic(t, func() {
				BitReverse(a)
			})
		}
	})

	t.Run("[4]uint64", func(t *testing.T) {
		for _, size := range sizes {
			t.Run(fmt.Sprintf("size=%d", size), func(t *testing.T) {
				// check that bit-reversing twice is identity
				original := make([][4]uint64, size)
				for i := range original {
					original[i] = [4]uint64{uint64(i), uint64(i) + 1, uint64(i) + 2, uint64(i) + 3}
				}
				a := make([][4]uint64, size)
				copy(a, original)

				BitReverse(a)
				BitReverse(a)

				for i := range a {
					if a[i] != original[i] {
						t.Fatalf("bit-reversing twice is not identity")
					}
				}
			})
		}

		// check that it panics for non-power of 2
		for _, size := range []int{3, 5, 6, 7, 9, 10, 12} {
			a := make([][4]uint64, size)
			assertPanic(t, func() {
				BitReverse(a)
			})
		}
	})
}

func TestBitReverseAlgorithms(t *testing.T) {
	state := uint32(1)
	next := func() uint32 {
		state ^= state << 13
		state ^= state >> 17
		state ^= state << 5
		return state
	}
	for logN := range 23 {
		size := 1 << logN
		src := make([]uint32, size)
		for i := range src {
			src[i] = next()
		}

		want := slices.Clone(src)
		BitReverseNaive(want)

		got := slices.Clone(src)
		BitReverseCobra(got)
		if !slices.Equal(got, want) {
			t.Fatalf("COBRA result differs from naive at size 2^%d", logN)
		}

		for name, copyFn := range map[string]func([]uint32, []uint32){
			"auto":  BitReverseCopy[uint32],
			"naive": BitReverseCopyNaive[uint32],
			"cobra": BitReverseCopyCobra[uint32],
		} {
			dst := make([]uint32, size)
			copyFn(dst, src)
			if !slices.Equal(dst, want) {
				t.Fatalf("%s copy differs from naive at size 2^%d", name, logN)
			}
		}
	}
}

func TestBitReverseInvalidInputs(t *testing.T) {
	for name, f := range map[string]func(){
		"empty in-place":     func() { BitReverseNaive([]uint32{}) },
		"non-power in-place": func() { BitReverseCobra(make([]uint32, 3)) },
		"empty copy":         func() { BitReverseCopy([]uint32{}, []uint32{}) },
		"length mismatch":    func() { BitReverseCopy(make([]uint32, 2), make([]uint32, 1)) },
		"same slice": func() {
			v := make([]uint32, 2)
			BitReverseCopy(v, v)
		},
		"partial overlap": func() {
			v := make([]uint32, 4)
			BitReverseCopy(v[:2], v[1:3])
		},
	} {
		t.Run(name, func(t *testing.T) { assertPanic(t, f) })
	}
}

func BenchmarkBitReverse(b *testing.B) {
	sizes := []int{1 << 8, 1 << 9, 1 << 16, 1 << 21, maxSizeBitReverse}

	b.Run("uint32", func(b *testing.B) {
		for _, size := range sizes {
			a := make([]uint32, size)
			for i := range a {
				a[i] = uint32(i)
			}
			b.Run(fmt.Sprintf("size=%d", size), func(b *testing.B) {
				b.ResetTimer()
				for range b.N {
					BitReverse(a)
				}
			})
		}
	})

	b.Run("[4]uint64", func(b *testing.B) {
		for _, size := range sizes {
			a := make([][4]uint64, size)
			for i := range a {
				a[i] = [4]uint64{uint64(i), uint64(i) + 1, uint64(i) + 2, uint64(i) + 3}
			}
			b.Run(fmt.Sprintf("size=%d", size), func(b *testing.B) {
				b.ResetTimer()
				for range b.N {
					BitReverse(a)
				}
			})
		}
	})
}

func BenchmarkBitReverseAlgorithms(b *testing.B) {
	for _, logN := range []int{20, 21, 22} {
		v := make([]uint32, 1<<logN)
		for name, reverse := range map[string]func([]uint32){
			"naive": BitReverseNaive[uint32],
			"cobra": BitReverseCobra[uint32],
		} {
			b.Run(fmt.Sprintf("in-place/%s/2^%d", name, logN), func(b *testing.B) {
				b.SetBytes(int64(len(v)) * 4)
				for b.Loop() {
					reverse(v)
				}
			})
		}

		dst := make([]uint32, len(v))
		for name, reverse := range map[string]func([]uint32, []uint32){
			"naive": BitReverseCopyNaive[uint32],
			"cobra": BitReverseCopyCobra[uint32],
		} {
			b.Run(fmt.Sprintf("copy/%s/2^%d", name, logN), func(b *testing.B) {
				b.SetBytes(int64(len(v)) * 4)
				for b.Loop() {
					reverse(dst, v)
				}
			})
		}
	}
}

// / test helpers
func assertPanic(t *testing.T, f func()) {
	t.Helper()
	defer func() {
		if r := recover(); r == nil {
			t.Errorf("expected a panic")
		}
	}()
	f()
}
