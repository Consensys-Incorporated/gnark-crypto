// Copyright 2020-2025 Consensys Incorporated
// Licensed under the Apache License, Version 2.0. See the LICENSE file for details.

package utils

import (
	"math/bits"
	"unsafe"
)

// BitReverse applies the bit-reversal permutation to v.
// len(v) must be a power of 2.
//
// Slices of at least 128 KiB use COBRA-L1 (see bitReverseCobraInPlaceL1),
// except for elements over 4 bytes on one task on arm64; smaller slices use
// the naive loop. It runs on the calling goroutine unless [WithBitReverseNbTasks] is
// given.
func BitReverse[T any](v []T, opts ...BitReverseOption) {
	checkBitReverseLength(len(v))
	if tasks, logTile, ok := bitReverseTiling[T](len(v), bitReverseNbTasks(opts)); ok &&
		(tasks > 1 || bitReverseSerialInPlace(unsafe.Sizeof(*new(T)))) {
		bitReverseCobraInPlaceL1(v, tasks, logTile) // tasks may be 1
		return
	}
	bitReverseNaive(v)
}

// BitReverseNaive applies the bit-reversal permutation to v using direct
// swaps. len(v) must be a power of two.
func BitReverseNaive[T any](v []T) {
	checkBitReverseLength(len(v))
	bitReverseNaive(v)
}

// BitReverseCopy writes dst[bitrev(i)] = src[i]. The slices must have the
// same power-of-two length and must not overlap.
//
// Slices of at least 128 KiB use COBRA-L1 (see bitReverseCopyCobraL1);
// smaller slices use the naive loop. It runs on the calling goroutine unless
// [WithBitReverseNbTasks] is given.
func BitReverseCopy[T any](dst, src []T, opts ...BitReverseOption) {
	checkBitReverseCopy(dst, src)
	if tasks, logTile, ok := bitReverseTiling[T](len(src), bitReverseNbTasks(opts)); ok {
		bitReverseCopyCobraL1(dst, src, tasks, logTile) // tasks may be 1
		return
	}
	bitReverseCopyNaive(dst, src)
}

// BitReverseCopyNaive writes dst[bitrev(i)] = src[i] using direct indexed
// copies. The slices must have the same power-of-two length and must not
// overlap.
func BitReverseCopyNaive[T any](dst, src []T) {
	checkBitReverseCopy(dst, src)
	bitReverseCopyNaive(dst, src)
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
