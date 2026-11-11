// Copyright 2020-2025 Consensys Incorporated
// Licensed under the Apache License, Version 2.0. See the LICENSE file for details.

package utils

// bitReverseTileBufBytes: half of the 64 KiB L1d of Neoverse (Graviton) cores;
// Apple cores have 128 KiB. Not yet measured on arm64.
const bitReverseTileBufBytes = 32 << 10

// bitReverseSerialInPlaceMaxElemSize: on an Apple M4, single-task in-place
// tiling was 2–4× faster than naive for 4-byte elements; larger elements are
// not measured with the current tile sizes, so they keep the naive path.
const bitReverseSerialInPlaceMaxElemSize = 4
