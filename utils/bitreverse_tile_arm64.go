// Copyright 2020-2025 Consensys Incorporated
// Licensed under the Apache License, Version 2.0. See the LICENSE file for details.

package utils

// bitReverseTileBufBytes: half of the 64 KiB L1d of Neoverse cores.
const bitReverseTileBufBytes = 32 << 10

// bitReverseSerialInPlaceMaxElemSize: on one task, in-place BitReverse uses
// COBRA-L1 only for elements up to this size. Larger elements only win from
// 256 KiB on arm64, so they keep the naive loop.
const bitReverseSerialInPlaceMaxElemSize = 4
