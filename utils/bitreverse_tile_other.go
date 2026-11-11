// Copyright 2020-2025 Consensys Incorporated
// Licensed under the Apache License, Version 2.0. See the LICENSE file for details.

//go:build !arm64

package utils

// bitReverseTileBufBytes: half of the 32 KiB L1d of AMD Zen cores.
const bitReverseTileBufBytes = 16 << 10

// bitReverseSerialInPlaceMaxElemSize: on one task, in-place BitReverse uses
// COBRA-L1 for every element size.
const bitReverseSerialInPlaceMaxElemSize = 1 << 30
