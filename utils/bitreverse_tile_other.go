// Copyright 2020-2025 Consensys Software Inc.
// Licensed under the Apache License, Version 2.0. See the LICENSE file for details.

//go:build !arm64

package utils

// bitReverseTileBufBytes: half of the 32 KiB L1d of AMD Zen cores, a third of
// the 48 KiB L1d of recent Intel cores.
const bitReverseTileBufBytes = 16 << 10

// bitReverseSerialInPlaceMaxElemSize bounds the element size for which
// in-place BitReverse tiles on a single task. On a Xeon 8488C the single-task
// tiled kernel is 2–4× faster than naive (and than COBRA) from 2^18 elements
// for 4- to 32-byte elements, and ties below that, so it is used for all.
const bitReverseSerialInPlaceMaxElemSize = 1 << 30
