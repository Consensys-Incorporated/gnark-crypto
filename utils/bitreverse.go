// Copyright 2020-2025 Consensys Software Inc.
// Licensed under the Apache License, Version 2.0. See the LICENSE file for details.

package utils

import (
	"math/bits"
	"runtime"
	"unsafe" // used for a heuristic on size of generic type, non critical.
)

// BitReverse applies the bit-reversal permutation to v.
// len(v) must be a power of 2
func BitReverse[T any](v []T) {
	checkBitReverseLength(len(v))

	if runtime.GOARCH == "arm64" || len(v) < (1<<21) || unsafe.Sizeof(v[0]) < 8 {
		bitReverseNaive(v)
	} else {
		bitReverseCobra(v)
	}
}

// BitReverseNaive applies the bit-reversal permutation to v using direct
// swaps. len(v) must be a power of two.
func BitReverseNaive[T any](v []T) {
	checkBitReverseLength(len(v))
	bitReverseNaive(v)
}

// BitReverseCobra applies the cache-friendly COBRA bit-reversal permutation
// to v. len(v) must be a power of two.
func BitReverseCobra[T any](v []T) {
	checkBitReverseLength(len(v))
	if len(v) == 1 {
		return
	}
	bitReverseCobra(v)
}

// BitReverseCopy writes dst[bitrev(i)] = src[i]. The slices must have the
// same power-of-two length and must not overlap. It uses the same architecture,
// input-size, and element-size heuristic as [BitReverse].
func BitReverseCopy[T any](dst, src []T) {
	checkBitReverseCopy(dst, src)
	if runtime.GOARCH == "arm64" || len(src) < (1<<21) || unsafe.Sizeof(src[0]) < 8 {
		bitReverseCopyNaive(dst, src)
	} else {
		bitReverseCopyCobra(dst, src)
	}
}

// BitReverseCopyNaive writes dst[bitrev(i)] = src[i] using direct indexed
// copies. The slices must have the same power-of-two length and must not
// overlap.
func BitReverseCopyNaive[T any](dst, src []T) {
	checkBitReverseCopy(dst, src)
	bitReverseCopyNaive(dst, src)
}

// BitReverseCopyCobra writes dst[bitrev(i)] = src[i] using cache-friendly
// COBRA tiling. The slices must have the same power-of-two length and must not
// overlap.
func BitReverseCopyCobra[T any](dst, src []T) {
	checkBitReverseCopy(dst, src)
	bitReverseCopyCobra(dst, src)
}

func checkBitReverseLength(n int) {
	if bits.OnesCount64(uint64(n)) != 1 {
		panic("length must be a power of two")
	}
}

func checkBitReverseCopy[T any](dst, src []T) {
	if len(dst) != len(src) {
		panic("source and destination lengths must match")
	}
	checkBitReverseLength(len(src))
	if slicesOverlap(dst, src) {
		panic("source and destination must not overlap")
	}
}

func slicesOverlap[T any](a, b []T) bool {
	elementSize := unsafe.Sizeof(*new(T))
	if elementSize == 0 || len(a) == 0 || len(b) == 0 {
		return false
	}
	aStart := uintptr(unsafe.Pointer(unsafe.SliceData(a)))
	bStart := uintptr(unsafe.Pointer(unsafe.SliceData(b)))
	aEnd := aStart + uintptr(len(a))*elementSize
	bEnd := bStart + uintptr(len(b))*elementSize
	return aStart < bEnd && bStart < aEnd
}

func bitReverseCopyNaive[T any](dst, src []T) {
	n := uint64(len(src))
	nn := uint64(64 - bits.TrailingZeros64(n))
	for i := range n {
		dst[bits.Reverse64(i)>>nn] = src[i]
	}
}

func bitReverseCopyCobra[T any](dst, src []T) {
	if len(src) == 1 {
		dst[0] = src[0]
		return
	}

	logN := uint64(bits.Len64(uint64(len(src))) - 1)
	logTileSize := deriveLogTileSize(logN)
	logBLen := logN - 2*logTileSize
	bShift := logBLen + logTileSize
	tileSize := uint64(1) << logTileSize
	t := make([]T, tileSize*tileSize)

	for b := range uint64(1) << logBLen {
		for a := range tileSize {
			aRev := (bits.Reverse64(a) >> (64 - logTileSize)) << logTileSize
			for c := range tileSize {
				t[aRev|c] = src[(a<<bShift)|(b<<logTileSize)|c]
			}
		}

		bRev := (bits.Reverse64(b) >> (64 - logBLen)) << logTileSize
		for c := range tileSize {
			base := ((bits.Reverse64(c) >> (64 - logTileSize)) << bShift) | bRev
			for aRev := range tileSize {
				dst[base|aRev] = t[(aRev<<logTileSize)|c]
			}
		}
	}
}

// bitReverseNaive applies the bit-reversal permutation to v.
// len(v) must be a power of 2
func bitReverseNaive[T any](v []T) {
	n := uint64(len(v))
	nn := uint64(64 - bits.TrailingZeros64(n))

	for i := range n {
		iRev := bits.Reverse64(i) >> nn
		if iRev > i {
			v[i], v[iRev] = v[iRev], v[i]
		}
	}
}

// bitReverseCobraInPlace applies the bit-reversal permutation to v.
// len(v) must be a power of 2
// This is derived from:
//
//   - Towards an Optimal Bit-Reversal Permutation Program
//     Larry Carter and Kang Su Gatlin, 1998
//     https://csaws.cs.technion.ac.il/~itai/Courses/Cache/bit.pdf
//
//   - Practically efficient methods for performing bit-reversed
//     permutation in C++11 on the x86-64 architecture
//     Knauth, Adas, Whitfield, Wang, Ickler, Conrad, Serang, 2017
//     https://arxiv.org/pdf/1708.01873.pdf
//
//   - and more specifically, constantine implementation:
//     https://github.com/mratsim/constantine/blob/d51699248db04e29c7b1ad97e0bafa1499db00b5/constantine/math/polynomials/fft.nim#L205
//     by Mamy Ratsimbazafy (@mratsim).
func bitReverseCobraInPlace[T any](v []T) {
	logN := uint64(bits.Len64(uint64(len(v))) - 1)
	logTileSize := deriveLogTileSize(logN)
	logBLen := logN - 2*logTileSize
	bLen := uint64(1) << logBLen
	bShift := logBLen + logTileSize
	tileSize := uint64(1) << logTileSize

	// rough idea;
	// bit reversal permutation naive implementation may have some cache associativity issues,
	// since we are accessing elements by strides of powers of 2.
	// on large inputs, this is noticeable and can be improved by using a t buffer.
	// idea is for t buffer to be small enough to fit in cache.
	// in the first inner loop, we copy the elements of v into t in a bit-reversed order.
	// in the subsequent inner loops, accesses have much better cache locality than the naive implementation.
	// hence even if we apparently do more work (swaps / copies), we are faster.
	//
	// on arm64 (and particularly on M1 macs), this is not noticeable, and the naive implementation is faster,
	// in most cases.
	// on x86 (and particularly on aws hpc6a) this is noticeable, and the t buffer implementation is faster (up to 3x).
	//
	// optimal choice for the tile size is cache dependent; in theory, we want the t buffer to fit in the L1 cache;
	// in practice, a common size for L1 is 64kb, a field element is 32bytes or more.
	// hence we can fit 2k elements in the L1 cache, which corresponds to a tile size of 2**5 with some margin for cache conflicts.
	//
	// for most sizes of interest, this tile size choice doesn't yield good results;
	// we find that a tile size of 2**9 gives best results for input sizes from 2**21 up to 2**27+.
	t := make([]T, tileSize*tileSize)

	// see https://csaws.cs.technion.ac.il/~itai/Courses/Cache/bit.pdf
	// for a detailed explanation of the algorithm.
	for b := range bLen {

		for a := range tileSize {
			aRev := (bits.Reverse64(a) >> (64 - logTileSize)) << logTileSize
			for c := range tileSize {
				idx := (a << bShift) | (b << logTileSize) | c
				t[aRev|c] = v[idx]
			}
		}

		bRev := (bits.Reverse64(b) >> (64 - logBLen)) << logTileSize

		for c := range tileSize {
			cRev := ((bits.Reverse64(c) >> (64 - logTileSize)) << bShift) | bRev
			for aRev := range tileSize {
				a := bits.Reverse64(aRev) >> (64 - logTileSize)
				idx := (a << bShift) | (b << logTileSize) | c
				idxRev := cRev | aRev
				if idx < idxRev {
					tIdx := (aRev << logTileSize) | c
					v[idxRev], t[tIdx] = t[tIdx], v[idxRev]
				}
			}
		}

		for a := range tileSize {
			aRev := bits.Reverse64(a) >> (64 - logTileSize)
			for c := range tileSize {
				cRev := (bits.Reverse64(c) >> (64 - logTileSize)) << bShift
				idx := (a << bShift) | (b << logTileSize) | c
				idxRev := cRev | bRev | aRev
				if idx < idxRev {
					tIdx := (aRev << logTileSize) | c
					v[idx], t[tIdx] = t[tIdx], v[idx]
				}
			}
		}
	}
}

func bitReverseCobra[T any](v []T) {
	switch len(v) {
	case 1 << 21:
		bitReverseCobraInPlace_9_21(v)
	case 1 << 22:
		bitReverseCobraInPlace_9_22(v)
	case 1 << 23:
		bitReverseCobraInPlace_9_23(v)
	case 1 << 24:
		bitReverseCobraInPlace_9_24(v)
	case 1 << 25:
		bitReverseCobraInPlace_9_25(v)
	case 1 << 26:
		bitReverseCobraInPlace_9_26(v)
	case 1 << 27:
		bitReverseCobraInPlace_9_27(v)
	default:
		bitReverseCobraInPlace(v)
	}
}

func deriveLogTileSize(logN uint64) uint64 {
	q := uint64(9) // see bitReverseCobraInPlace for more details

	for int(logN)-int(2*q) <= 0 {
		q--
	}

	return q
}

// note these are generated from bitreverse.go.tmpl

// bitReverseCobraInPlace_9_21 applies the bit-reversal permutation to v.
// len(v) must be 1 << 21.
// see bitReverseCobraInPlace for more details; this function is specialized for 9,
// as it declares the t buffer and various constants statically for performance.
func bitReverseCobraInPlace_9_21[T any](v []T) {
	const (
		logTileSize = uint64(9)
		tileSize    = uint64(1) << logTileSize
		logN        = 21
		logBLen     = logN - 2*logTileSize
		bShift      = logBLen + logTileSize
		bLen        = uint64(1) << logBLen
	)

	var t [tileSize * tileSize]T

	for b := range bLen {

		for a := range tileSize {
			aRev := (bits.Reverse64(a) >> 55) << logTileSize
			for c := range tileSize {
				idx := (a << bShift) | (b << logTileSize) | c
				t[aRev|c] = v[idx]
			}
		}

		bRev := (bits.Reverse64(b) >> (64 - logBLen)) << logTileSize

		for c := range tileSize {
			cRev := ((bits.Reverse64(c) >> 55) << bShift) | bRev
			for aRev := range tileSize {
				a := bits.Reverse64(aRev) >> 55
				idx := (a << bShift) | (b << logTileSize) | c
				idxRev := cRev | aRev
				if idx < idxRev {
					tIdx := (aRev << logTileSize) | c
					v[idxRev], t[tIdx] = t[tIdx], v[idxRev]
				}
			}
		}

		for a := range tileSize {
			aRev := bits.Reverse64(a) >> 55
			for c := range tileSize {
				cRev := (bits.Reverse64(c) >> 55) << bShift
				idx := (a << bShift) | (b << logTileSize) | c
				idxRev := cRev | bRev | aRev
				if idx < idxRev {
					tIdx := (aRev << logTileSize) | c
					v[idx], t[tIdx] = t[tIdx], v[idx]
				}
			}
		}
	}

}

// bitReverseCobraInPlace_9_22 applies the bit-reversal permutation to v.
// len(v) must be 1 << 22.
// see bitReverseCobraInPlace for more details; this function is specialized for 9,
// as it declares the t buffer and various constants statically for performance.
func bitReverseCobraInPlace_9_22[T any](v []T) {
	const (
		logTileSize = uint64(9)
		tileSize    = uint64(1) << logTileSize
		logN        = 22
		logBLen     = logN - 2*logTileSize
		bShift      = logBLen + logTileSize
		bLen        = uint64(1) << logBLen
	)

	var t [tileSize * tileSize]T

	for b := range bLen {

		for a := range tileSize {
			aRev := (bits.Reverse64(a) >> 55) << logTileSize
			for c := range tileSize {
				idx := (a << bShift) | (b << logTileSize) | c
				t[aRev|c] = v[idx]
			}
		}

		bRev := (bits.Reverse64(b) >> (64 - logBLen)) << logTileSize

		for c := range tileSize {
			cRev := ((bits.Reverse64(c) >> 55) << bShift) | bRev
			for aRev := range tileSize {
				a := bits.Reverse64(aRev) >> 55
				idx := (a << bShift) | (b << logTileSize) | c
				idxRev := cRev | aRev
				if idx < idxRev {
					tIdx := (aRev << logTileSize) | c
					v[idxRev], t[tIdx] = t[tIdx], v[idxRev]
				}
			}
		}

		for a := range tileSize {
			aRev := bits.Reverse64(a) >> 55
			for c := range tileSize {
				cRev := (bits.Reverse64(c) >> 55) << bShift
				idx := (a << bShift) | (b << logTileSize) | c
				idxRev := cRev | bRev | aRev
				if idx < idxRev {
					tIdx := (aRev << logTileSize) | c
					v[idx], t[tIdx] = t[tIdx], v[idx]
				}
			}
		}
	}

}

// bitReverseCobraInPlace_9_23 applies the bit-reversal permutation to v.
// len(v) must be 1 << 23.
// see bitReverseCobraInPlace for more details; this function is specialized for 9,
// as it declares the t buffer and various constants statically for performance.
func bitReverseCobraInPlace_9_23[T any](v []T) {
	const (
		logTileSize = uint64(9)
		tileSize    = uint64(1) << logTileSize
		logN        = 23
		logBLen     = logN - 2*logTileSize
		bShift      = logBLen + logTileSize
		bLen        = uint64(1) << logBLen
	)

	var t [tileSize * tileSize]T

	for b := range bLen {

		for a := range tileSize {
			aRev := (bits.Reverse64(a) >> 55) << logTileSize
			for c := range tileSize {
				idx := (a << bShift) | (b << logTileSize) | c
				t[aRev|c] = v[idx]
			}
		}

		bRev := (bits.Reverse64(b) >> (64 - logBLen)) << logTileSize

		for c := range tileSize {
			cRev := ((bits.Reverse64(c) >> 55) << bShift) | bRev
			for aRev := range tileSize {
				a := bits.Reverse64(aRev) >> 55
				idx := (a << bShift) | (b << logTileSize) | c
				idxRev := cRev | aRev
				if idx < idxRev {
					tIdx := (aRev << logTileSize) | c
					v[idxRev], t[tIdx] = t[tIdx], v[idxRev]
				}
			}
		}

		for a := range tileSize {
			aRev := bits.Reverse64(a) >> 55
			for c := range tileSize {
				cRev := (bits.Reverse64(c) >> 55) << bShift
				idx := (a << bShift) | (b << logTileSize) | c
				idxRev := cRev | bRev | aRev
				if idx < idxRev {
					tIdx := (aRev << logTileSize) | c
					v[idx], t[tIdx] = t[tIdx], v[idx]
				}
			}
		}
	}

}

// bitReverseCobraInPlace_9_24 applies the bit-reversal permutation to v.
// len(v) must be 1 << 24.
// see bitReverseCobraInPlace for more details; this function is specialized for 9,
// as it declares the t buffer and various constants statically for performance.
func bitReverseCobraInPlace_9_24[T any](v []T) {
	const (
		logTileSize = uint64(9)
		tileSize    = uint64(1) << logTileSize
		logN        = 24
		logBLen     = logN - 2*logTileSize
		bShift      = logBLen + logTileSize
		bLen        = uint64(1) << logBLen
	)

	var t [tileSize * tileSize]T

	for b := range bLen {

		for a := range tileSize {
			aRev := (bits.Reverse64(a) >> 55) << logTileSize
			for c := range tileSize {
				idx := (a << bShift) | (b << logTileSize) | c
				t[aRev|c] = v[idx]
			}
		}

		bRev := (bits.Reverse64(b) >> (64 - logBLen)) << logTileSize

		for c := range tileSize {
			cRev := ((bits.Reverse64(c) >> 55) << bShift) | bRev
			for aRev := range tileSize {
				a := bits.Reverse64(aRev) >> 55
				idx := (a << bShift) | (b << logTileSize) | c
				idxRev := cRev | aRev
				if idx < idxRev {
					tIdx := (aRev << logTileSize) | c
					v[idxRev], t[tIdx] = t[tIdx], v[idxRev]
				}
			}
		}

		for a := range tileSize {
			aRev := bits.Reverse64(a) >> 55
			for c := range tileSize {
				cRev := (bits.Reverse64(c) >> 55) << bShift
				idx := (a << bShift) | (b << logTileSize) | c
				idxRev := cRev | bRev | aRev
				if idx < idxRev {
					tIdx := (aRev << logTileSize) | c
					v[idx], t[tIdx] = t[tIdx], v[idx]
				}
			}
		}
	}

}

// bitReverseCobraInPlace_9_25 applies the bit-reversal permutation to v.
// len(v) must be 1 << 25.
// see bitReverseCobraInPlace for more details; this function is specialized for 9,
// as it declares the t buffer and various constants statically for performance.
func bitReverseCobraInPlace_9_25[T any](v []T) {
	const (
		logTileSize = uint64(9)
		tileSize    = uint64(1) << logTileSize
		logN        = 25
		logBLen     = logN - 2*logTileSize
		bShift      = logBLen + logTileSize
		bLen        = uint64(1) << logBLen
	)

	var t [tileSize * tileSize]T

	for b := range bLen {

		for a := range tileSize {
			aRev := (bits.Reverse64(a) >> 55) << logTileSize
			for c := range tileSize {
				idx := (a << bShift) | (b << logTileSize) | c
				t[aRev|c] = v[idx]
			}
		}

		bRev := (bits.Reverse64(b) >> (64 - logBLen)) << logTileSize

		for c := range tileSize {
			cRev := ((bits.Reverse64(c) >> 55) << bShift) | bRev
			for aRev := range tileSize {
				a := bits.Reverse64(aRev) >> 55
				idx := (a << bShift) | (b << logTileSize) | c
				idxRev := cRev | aRev
				if idx < idxRev {
					tIdx := (aRev << logTileSize) | c
					v[idxRev], t[tIdx] = t[tIdx], v[idxRev]
				}
			}
		}

		for a := range tileSize {
			aRev := bits.Reverse64(a) >> 55
			for c := range tileSize {
				cRev := (bits.Reverse64(c) >> 55) << bShift
				idx := (a << bShift) | (b << logTileSize) | c
				idxRev := cRev | bRev | aRev
				if idx < idxRev {
					tIdx := (aRev << logTileSize) | c
					v[idx], t[tIdx] = t[tIdx], v[idx]
				}
			}
		}
	}

}

// bitReverseCobraInPlace_9_26 applies the bit-reversal permutation to v.
// len(v) must be 1 << 26.
// see bitReverseCobraInPlace for more details; this function is specialized for 9,
// as it declares the t buffer and various constants statically for performance.
func bitReverseCobraInPlace_9_26[T any](v []T) {
	const (
		logTileSize = uint64(9)
		tileSize    = uint64(1) << logTileSize
		logN        = 26
		logBLen     = logN - 2*logTileSize
		bShift      = logBLen + logTileSize
		bLen        = uint64(1) << logBLen
	)

	var t [tileSize * tileSize]T

	for b := range bLen {

		for a := range tileSize {
			aRev := (bits.Reverse64(a) >> 55) << logTileSize
			for c := range tileSize {
				idx := (a << bShift) | (b << logTileSize) | c
				t[aRev|c] = v[idx]
			}
		}

		bRev := (bits.Reverse64(b) >> (64 - logBLen)) << logTileSize

		for c := range tileSize {
			cRev := ((bits.Reverse64(c) >> 55) << bShift) | bRev
			for aRev := range tileSize {
				a := bits.Reverse64(aRev) >> 55
				idx := (a << bShift) | (b << logTileSize) | c
				idxRev := cRev | aRev
				if idx < idxRev {
					tIdx := (aRev << logTileSize) | c
					v[idxRev], t[tIdx] = t[tIdx], v[idxRev]
				}
			}
		}

		for a := range tileSize {
			aRev := bits.Reverse64(a) >> 55
			for c := range tileSize {
				cRev := (bits.Reverse64(c) >> 55) << bShift
				idx := (a << bShift) | (b << logTileSize) | c
				idxRev := cRev | bRev | aRev
				if idx < idxRev {
					tIdx := (aRev << logTileSize) | c
					v[idx], t[tIdx] = t[tIdx], v[idx]
				}
			}
		}
	}

}

// bitReverseCobraInPlace_9_27 applies the bit-reversal permutation to v.
// len(v) must be 1 << 27.
// see bitReverseCobraInPlace for more details; this function is specialized for 9,
// as it declares the t buffer and various constants statically for performance.
func bitReverseCobraInPlace_9_27[T any](v []T) {
	const (
		logTileSize = uint64(9)
		tileSize    = uint64(1) << logTileSize
		logN        = 27
		logBLen     = logN - 2*logTileSize
		bShift      = logBLen + logTileSize
		bLen        = uint64(1) << logBLen
	)

	var t [tileSize * tileSize]T

	for b := range bLen {

		for a := range tileSize {
			aRev := (bits.Reverse64(a) >> 55) << logTileSize
			for c := range tileSize {
				idx := (a << bShift) | (b << logTileSize) | c
				t[aRev|c] = v[idx]
			}
		}

		bRev := (bits.Reverse64(b) >> (64 - logBLen)) << logTileSize

		for c := range tileSize {
			cRev := ((bits.Reverse64(c) >> 55) << bShift) | bRev
			for aRev := range tileSize {
				a := bits.Reverse64(aRev) >> 55
				idx := (a << bShift) | (b << logTileSize) | c
				idxRev := cRev | aRev
				if idx < idxRev {
					tIdx := (aRev << logTileSize) | c
					v[idxRev], t[tIdx] = t[tIdx], v[idxRev]
				}
			}
		}

		for a := range tileSize {
			aRev := bits.Reverse64(a) >> 55
			for c := range tileSize {
				cRev := (bits.Reverse64(c) >> 55) << bShift
				idx := (a << bShift) | (b << logTileSize) | c
				idxRev := cRev | bRev | aRev
				if idx < idxRev {
					tIdx := (aRev << logTileSize) | c
					v[idx], t[tIdx] = t[tIdx], v[idx]
				}
			}
		}
	}

}
