// Copyright 2020-2025 Consensys Incorporated
// Licensed under the Apache License, Version 2.0. See the LICENSE file for details.

package utils

// bitReverseTileBufBytes: half of the 64 KiB L1d of Neoverse (Graviton) cores;
// Apple cores have 128 KiB. Not yet measured on arm64.
const bitReverseTileBufBytes = 32 << 10

// bitReverseSerialInPlaceMaxElemSize: on an Apple M4, single-task in-place
// COBRA-L1 beats naive from 128 KiB for 4-byte elements, but larger elements
// only from 256 KiB (at 128 KiB they are 1.3–2.4× slower than naive). Rather
// than a second in-place threshold, they keep the naive path on one task.
const bitReverseSerialInPlaceMaxElemSize = 4
