// Copyright 2020-2025 Consensys Incorporated
// Licensed under the Apache License, Version 2.0. See the LICENSE file for details.

package utils

import (
	"math/bits"
	"runtime"
	"unsafe"

	"github.com/consensys/gnark-crypto/parallel"
)

// BitReverseOption configures [BitReverse] and [BitReverseCopy]. It works on
// the config by value, so that options do not make the config escape to the
// heap.
type BitReverseOption func(bitReverseConfig) bitReverseConfig

type bitReverseConfig struct {
	nbTasks int
}

// WithNbTasks lets [BitReverse] and [BitReverseCopy] split the permutation
// over up to nbTasks goroutines. Like fft.WithNbTasks, nbTasks < 1 means one
// task (serial) and nbTasks is capped at 512; pass runtime.GOMAXPROCS(0) to
// use every core.
// The default is one task. Small inputs use fewer tasks (each gets at least
// 256 KiB).
func WithNbTasks(nbTasks int) BitReverseOption {
	nbTasks = min(max(nbTasks, 1), 512)
	return func(c bitReverseConfig) bitReverseConfig {
		c.nbTasks = nbTasks
		return c
	}
}

func bitReverseNbTasks(opts []BitReverseOption) int {
	cfg := bitReverseConfig{nbTasks: 1}
	for _, opt := range opts {
		cfg = opt(cfg)
	}
	return cfg.nbTasks
}

const (
	// bitReverseMaxLogTile caps the tile side at B = 64.
	bitReverseMaxLogTile = 6
	// bitReverseBytesPerTask is the least data a goroutine is given.
	bitReverseBytesPerTask = 256 << 10
	// bitReverseMinTiledBytes is the smallest slice handled by COBRA-L1; below
	// it the slice fits in cache and the naive loop is used. One threshold for
	// all element sizes, measured on amd64 and arm64.
	bitReverseMinTiledBytes = 128 << 10
	// bitReverseLargeSliceBytes is the slice size from which 16- to 96-byte
	// elements use the largest tile (see bitReverseTiling).
	bitReverseLargeSliceBytes = 64 << 20
)

// bitReverseTileBufBytes is the size budget of one task's B×B tile buffer: a
// fraction of the L1d, since the kernels write into the buffer with strided
// stores. It is half of the L1d of the smallest common cores: 32 KiB on arm64
// (64 KiB on Neoverse), 16 KiB elsewhere (32 KiB on AMD Zen). runtime.GOARCH
// is a constant, so the branch is resolved at compile time.
func bitReverseTileBufBytes() int {
	if runtime.GOARCH == "arm64" {
		return 32 << 10
	}
	return 16 << 10
}

// bitReverseSerialInPlace reports whether in-place BitReverse uses COBRA-L1
// on a single task for elements of elemSize bytes. On arm64 only elements of
// at most 4 bytes do: larger ones only win from 256 KiB there, so they keep
// the naive loop. Elsewhere every element size does.
func bitReverseSerialInPlace(elemSize uintptr) bool {
	return runtime.GOARCH != "arm64" || elemSize <= 4
}

// bitReverseTiling reports whether COBRA-L1 applies to n elements of type T,
// and if so the number of tasks (1 to nbTasks) and log2 of the tile side.
//
// Together with BitReverse and BitReverseCopy, the full dispatch is:
//
//	nbTasks = WithNbTasks(n) clamped to [1, 512], 1 by default
//	slice < 128 KiB                        → naive
//	q = largest q <= 6 with B²·elemSize <= bitReverseTileBufBytes()
//	    (16 KiB, or 32 KiB on arm64)
//	16 <= elemSize <= 96, slice >= 64 MiB  → q = 6
//	logN < 2q (no full B×B tile)           → naive
//	tasks = max(1, min(nbTasks, slice/256 KiB, tiles))
//	BitReverseCopy                         → COBRA-L1(tasks, q)
//	BitReverse: tasks > 1 or bitReverseSerialInPlace(elemSize)
//	    (false only on arm64 for elements over 4 bytes)
//	                                       → COBRA-L1(tasks, q), else naive
func bitReverseTiling[T any](n, nbTasks int) (tasks, logTile int, ok bool) {
	elemSize := int(unsafe.Sizeof(*new(T)))
	// The byte size is computed in uint64: n*elemSize in int overflows for
	// slices of 2 GiB or more on 32-bit platforms.
	bytes := uint64(n) * uint64(elemSize)
	if elemSize == 0 || bytes < bitReverseMinTiledBytes {
		return 1, 0, false
	}
	logTile = bitReverseLogTile(elemSize)
	// On slices far beyond the caches, each far jump costs a DRAM access, and
	// fewer jumps per element (a bigger tile) pay off more than keeping the
	// tile in L1. For 16- to 96-byte elements, q = 6 is then 1.1–1.4× faster
	// on AMD Zen 4 (where the L1 budget's smaller q lost to COBRA-9) and equal
	// or faster on a Xeon 8488C. The upper bound (the largest size measured)
	// keeps each q = 6 tile buffer at 384 KiB or less, two of them in place;
	// larger elements keep the budget's q.
	if elemSize >= 16 && elemSize <= 96 && bytes >= bitReverseLargeSliceBytes {
		logTile = bitReverseMaxLogTile
	}
	logN := bits.TrailingZeros(uint(n))
	// The slice must hold at least one full B×B tile (logN >= 2q).
	if logN < 2*logTile {
		return 1, 0, false
	}
	tasks = max(1, min(nbTasks, int(bytes/bitReverseBytesPerTask), 1<<(logN-2*logTile)))
	return tasks, logTile, true
}

// bitReverseLogTile returns the largest q in [1, bitReverseMaxLogTile] such
// that a 2^q × 2^q tile of elemSize-byte elements fits bitReverseTileBufBytes().
func bitReverseLogTile(elemSize int) int {
	logTile := bitReverseMaxLogTile
	for logTile > 1 && (1<<(2*logTile))*elemSize > bitReverseTileBufBytes() {
		logTile--
	}
	return logTile
}

// COBRA-L1 is COBRA (Carter and Gatlin, 1998; see bitReverseCobraInPlace)
// with tiles small enough for L1, split over goroutines.
//
// With the index split as i = a|c|b (a and b of q bits),
//
//	bitrev(a|c|b) = rev(b)|rev(c)|rev(a),
//
// so the B×B elements sharing c (the tile of c, B contiguous runs of B) map to
// the tile of rev(c), transposed. Each tile goes through a buffer, so that
// both reads and writes are contiguous runs.

// bitReverseCopyCobraL1 writes dst[bitrev(i)] = src[i] with COBRA-L1 on tasks
// goroutines. Tiles are disjoint, so they are split over the tasks.
func bitReverseCopyCobraL1[T any](dst, src []T, tasks, logTile int) {
	n := len(src)
	logN := bits.TrailingZeros(uint(n))
	q := logTile
	B := 1 << q
	mid := logN - 2*q
	runShift := mid + q // a selects the source run, rev(b) the destination run

	var revQ [1 << bitReverseMaxLogTile]int
	for x := range B {
		revQ[x] = int(bits.Reverse64(uint64(x)) >> (64 - q))
	}

	parallel.Execute(1<<mid, func(start, end int) {
		buf := make([]T, B*B)
		for c := start; c < end; c++ {
			// For mid = 0 the shift is by 64, which gives 0 for a uint64.
			rc := int(bits.Reverse64(uint64(c)) >> (64 - mid))
			for a := range B {
				run := src[a<<runShift|c<<q:][:B]
				ra := revQ[a]
				// Indexed rather than ranged: faster for array element types.
				for b := range run {
					buf[revQ[b]<<q|ra] = run[b]
				}
			}
			for rb := range B {
				copy(dst[rb<<runShift|rc<<q:][:B], buf[rb<<q:][:B])
			}
		}
	}, tasks)
}

// bitReverseCobraInPlaceL1 applies the bit-reversal permutation to v in place
// with COBRA-L1 on tasks goroutines. The tiles of c and rev(c) are swapped
// together by a single task (see ownsPair), so tasks do not race.
func bitReverseCobraInPlaceL1[T any](v []T, tasks, logTile int) {
	n := len(v)
	logN := bits.TrailingZeros(uint(n))
	q := logTile
	B := 1 << q
	mid := logN - 2*q
	runShift := mid + q

	var revQ [1 << bitReverseMaxLogTile]int
	for x := range B {
		revQ[x] = int(bits.Reverse64(uint64(x)) >> (64 - q))
	}

	// load transposes the tile of c into buf, laid out as the tile of rev(c).
	load := func(buf []T, c int) {
		for a := range B {
			run := v[a<<runShift|c<<q:][:B]
			ra := revQ[a]
			for b := range run {
				buf[revQ[b]<<q|ra] = run[b]
			}
		}
	}
	store := func(c int, buf []T) {
		for rb := range B {
			copy(v[rb<<runShift|c<<q:][:B], buf[rb<<q:][:B])
		}
	}

	// ownsPair reports whether c, rather than rc = rev(c) != c, handles the
	// pair {c, rc}. Exactly one of them does. Using min(c, rc) alone would
	// give most pairs to the tasks with low c; XOR-ing with the parity of
	// popcount(c), the same for c and rc, balances them.
	ownsPair := func(c, rc int) bool {
		return (c < rc) != (bits.OnesCount(uint(c))&1 == 1)
	}

	parallel.Execute(1<<mid, func(start, end int) {
		buf := make([]T, 2*B*B)
		bufC, bufRC := buf[:B*B], buf[B*B:]
		for c := start; c < end; c++ {
			rc := int(bits.Reverse64(uint64(c)) >> (64 - mid))
			switch {
			case c == rc:
				load(bufC, c)
				store(c, bufC)
			case ownsPair(c, rc):
				load(bufC, c)
				load(bufRC, rc)
				store(rc, bufC)
				store(c, bufRC)
			}
		}
	}, tasks)
}
