// Copyright 2020-2025 Consensys Software Inc.
// Licensed under the Apache License, Version 2.0. See the LICENSE file for details.

package utils

import (
	"math/bits"
	"runtime"
	"unsafe"

	"github.com/consensys/gnark-crypto/parallel"
)

// BitReverseOption configures [BitReverse] and [BitReverseCopy].
type BitReverseOption func(*bitReverseConfig)

type bitReverseConfig struct {
	nbTasks int
}

// WithNbTasks lets [BitReverse] and [BitReverseCopy] split the permutation
// over up to nbTasks goroutines. nbTasks <= 0 means runtime.GOMAXPROCS(0).
//
// The default is a single task. Callers that already parallelise one level up
// (e.g. one permutation per column, inside a parallel loop over the columns)
// should keep the default or pass their share of the CPUs.
//
// Fewer tasks than requested are used on small inputs, where spawning a
// goroutine costs more than it saves (each task gets at least 256 KiB).
func WithNbTasks(nbTasks int) BitReverseOption {
	return func(c *bitReverseConfig) {
		if nbTasks <= 0 {
			nbTasks = runtime.GOMAXPROCS(0)
		}
		c.nbTasks = nbTasks
	}
}

func bitReverseNbTasks(opts []BitReverseOption) int {
	cfg := bitReverseConfig{nbTasks: 1}
	for _, opt := range opts {
		opt(&cfg)
	}
	return cfg.nbTasks
}

const (
	// bitReverseMaxLogTile caps the tile side at B = 64.
	bitReverseMaxLogTile = 6
	// bitReverseBytesPerTask is the least data a goroutine is given.
	bitReverseBytesPerTask = 256 << 10
	// bitReverseMinTiledBytes is the smallest slice the COBRA-L1 kernels handle;
	// below it the slice stays in cache and the naive loop is good enough
	// (see bitReverseTiling). On a Xeon 8488C, 4-byte elements already gain
	// from tiling at 32 KiB, while 24- and 32-byte elements tie with naive up
	// to the L2 size; 128 KiB was tuned on an Apple M4, where lower made small
	// copies slower.
	bitReverseMinTiledBytes = 128 << 10
)

// bitReverseTileBufBytes (bitreverse_tile_*.go) is the L1 budget for the B×B
// tile buffer of one task. The kernels scatter into the buffer with strided
// stores, which become expensive as soon as the buffer leaves L1, so the
// budget is a fraction of a conservative L1d size per architecture.
//
// On a Xeon 8488C (48 KiB L1d), single-task copy, the 16 KiB amd64 budget
// picks q = 6 for 4-byte elements, q = 5 for 8 bytes and q = 4 for 24- and
// 32-byte ones. Against the previous q = 5 / 5 / 3 this is 4% faster for
// uint32 at 2^22, 12% for E6 at 2^21, and 6–18% in place for 24- and 32-byte
// elements at 2^21–2^23. One step past the budget, 4- and 8-byte elements are
// 3–5× slower.

// bitReverseTiling reports whether the COBRA-L1 kernels apply to n elements of
// type T (ok), and if so how many goroutines to use (at most nbTasks, possibly
// 1) and log2 of the tile side. Tiling applies to slices of at least
// bitReverseMinTiledBytes whose length allows a full tile.
func bitReverseTiling[T any](n, nbTasks int) (tasks, logTile int, ok bool) {
	elemSize := int(unsafe.Sizeof(*new(T)))
	// Size threshold: this is what decides tiled or naive. The naive loop
	// refetches each destination cache line once per element it holds (up to
	// 16 times for 4-byte elements). In a small slice that line is still in L1
	// or L2 when the loop comes back to it, so the revisits are cheap, and
	// tiling would save little while still paying its per-tile costs (loop
	// setup, the buffer).
	if elemSize == 0 || n*elemSize < bitReverseMinTiledBytes {
		return 1, 0, false
	}
	logTile = bitReverseLogTile(elemSize)
	logN := bits.TrailingZeros(uint(n))
	// Correctness condition, not a size threshold: the kernels split the index
	// as a|c|b with q = logTile bits each for a and b, which needs
	// logN >= 2q, i.e. the slice must hold at least one full B×B tile
	// (n >= B²). Since a tile is at most bitReverseTileBufBytes, any slice
	// that passed the size check above holds many tiles; this only rejects
	// elements larger than a few KiB, where B drops to 2.
	if logN < 2*logTile {
		return 1, 0, false
	}
	tasks = max(1, min(nbTasks, n*elemSize/bitReverseBytesPerTask, 1<<(logN-2*logTile)))
	return tasks, logTile, true
}

// bitReverseLogTile returns log2 of the tile side for elements of elemSize
// bytes: the largest q in [1, bitReverseMaxLogTile] with
// B²·elemSize <= bitReverseTileBufBytes. In place, a task holds two such
// buffers but uses them one after the other; measured, it does best with the
// same q as copy (uint32: 15–20% faster at q = 6 than at q = 5).
func bitReverseLogTile(elemSize int) int {
	logTile := bitReverseMaxLogTile
	for logTile > 1 && (1<<(2*logTile))*elemSize > bitReverseTileBufBytes {
		logTile--
	}
	return logTile
}

// COBRA-L1: the COBRA bit reversal (Carter and Gatlin, "Towards an Optimal
// Bit-Reversal Permutation Program", 1998; see bitReverseCobraInPlace for the
// references) with tiles sized to stay in L1, spread over goroutines.
//
// Reversing a bit string split into three fields is reversing each field and
// swapping the outer two. With the logN index bits split as i = a|c|b, a and b
// of q = logTile bits and c of mid = logN-2q bits,
//
//	bitrev(a|c|b) = rev(b)|rev(c)|rev(a).
//
// For a fixed middle c, the B×B elements (a, b), B = 2^q, form B contiguous
// runs of B elements (the tile of c), and their images form the tile of
// rev(c): element (a, b) goes to row rev(b), column rev(a), i.e. the tile is
// transposed with both coordinates reversed. Each tile goes through a small
// buffer, so every source and destination cache line is touched in one burst;
// storing straight from the tile would make the B destination runs, which are
// a power-of-two stride apart, evict each other.
//
// Differences with bitReverseCobra / bitReverseCopyCobra (same algorithm):
//   - the tile side is the largest 2^q (q <= 6) whose B×B buffer fits
//     bitReverseTileBufBytes, so the buffer stays in L1, where bitReverseCobra
//     uses a side of up to 2^9 (a buffer of up to 2^18 elements, in L2/L3);
//   - the transpose happens when filling the buffer (strided writes into it,
//     then contiguous copies out) rather than when emptying it;
//   - tiles of distinct c are disjoint, so they are split over the tasks; in
//     place, the tiles of c and rev(c) are exchanged together by one task.

// bitReverseCopyCobraL1 writes dst[bitrev(i)] = src[i] with COBRA-L1, on
// tasks goroutines (1 runs it on the calling goroutine). Tiles of distinct c
// are disjoint and are spread over the tasks.
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
			rc := 0
			if mid > 0 {
				rc = int(bits.Reverse64(uint64(c)) >> (64 - mid))
			}
			for a := range B {
				run := src[a<<runShift|c<<q:][:B]
				ra := revQ[a]
				// Index run rather than ranging over its values: for array
				// element types the range value goes through a stack slot
				// with overlapping moves, which stalls store forwarding.
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
// with COBRA-L1, on tasks goroutines. The tiles of c and rev(c) are exchanged
// by the task that owns min(c, rev(c)); distinct pairs are disjoint, so tasks
// do not race.
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
			for b := range run { // indexed; see bitReverseCopyCobraL1
				buf[revQ[b]<<q|ra] = run[b]
			}
		}
	}
	store := func(c int, buf []T) {
		for rb := range B {
			copy(v[rb<<runShift|c<<q:][:B], buf[rb<<q:][:B])
		}
	}

	parallel.Execute(1<<mid, func(start, end int) {
		buf := make([]T, 2*B*B)
		bufC, bufRC := buf[:B*B], buf[B*B:]
		for c := start; c < end; c++ {
			rc := 0
			if mid > 0 {
				rc = int(bits.Reverse64(uint64(c)) >> (64 - mid))
			}
			switch {
			case c == rc:
				load(bufC, c)
				store(c, bufC)
			case c < rc:
				load(bufC, c)
				load(bufRC, rc)
				store(rc, bufC)
				store(c, bufRC)
			}
		}
	}, tasks)
}
