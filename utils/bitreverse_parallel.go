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
// goroutine costs more than it saves; below that size the serial path is used.
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
	// bitReverseTileRowBytes bounds one row of a tile: the tile side B is the
	// largest power of two (at most 2^bitReverseMaxLogTile) with
	// B·sizeof(T) <= bitReverseTileRowBytes, i.e. B = 32 for 4-byte elements
	// and B = 8 for 32-byte elements. A row then spans a few cache lines and
	// the B×B tile buffers stay in L1.
	bitReverseTileRowBytes = 256
	bitReverseMaxLogTile   = 5
	// bitReverseBytesPerTask is the least data a goroutine is given.
	bitReverseBytesPerTask = 256 << 10
)

// bitReverseParallelTasks returns the number of goroutines to use for a
// parallel bit reversal of n elements of type T, and log2 of the tile side.
// It returns nbTasks <= 1 when the serial path should be used instead.
func bitReverseParallelTasks[T any](n, nbTasks int) (tasks, logTile int) {
	if nbTasks <= 1 {
		return 1, 0
	}
	elemSize := int(unsafe.Sizeof(*new(T)))
	if elemSize == 0 {
		return 1, 0
	}
	logTile = bitReverseMaxLogTile
	for logTile > 1 && (1<<logTile)*elemSize > bitReverseTileRowBytes {
		logTile--
	}
	logN := bits.TrailingZeros(uint(n))
	if logN < 2*logTile {
		return 1, 0
	}
	tasks = min(nbTasks, n*elemSize/bitReverseBytesPerTask, 1<<(logN-2*logTile))
	return tasks, logTile
}

// The parallel bit reversals below split the logN index bits as
// i = a|c|b, with a and b of q = logTile bits each and c of mid = logN-2q bits,
// so that bitrev(i) = rev(b)|rev(c)|rev(a). For a fixed middle c, the B×B
// elements (a, b) form B contiguous runs of B elements (the tile of c), and
// their images form the tile of rev(c). Each tile is transposed through a
// small L1-resident buffer, so every source and destination cache line is
// touched in one burst; storing straight from the tile would make the B
// destination runs, which are a power-of-two stride apart, evict each other.

// bitReverseCopyParallel writes dst[bitrev(i)] = src[i]. Tiles of distinct c
// are disjoint and are spread over the tasks.
func bitReverseCopyParallel[T any](dst, src []T, tasks, logTile int) {
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
				for b, x := range run {
					buf[revQ[b]<<q|ra] = x
				}
			}
			for rb := range B {
				copy(dst[rb<<runShift|rc<<q:][:B], buf[rb<<q:][:B])
			}
		}
	}, tasks)
}

// bitReversePermuteParallel applies the bit-reversal permutation to v in place.
// The tiles of c and rev(c) are exchanged by the task that owns
// min(c, rev(c)); distinct pairs are disjoint, so tasks do not race.
func bitReversePermuteParallel[T any](v []T, tasks, logTile int) {
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
			for b, x := range run {
				buf[revQ[b]<<q|ra] = x
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
