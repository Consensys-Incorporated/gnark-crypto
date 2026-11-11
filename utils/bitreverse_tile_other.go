// Copyright 2020-2025 Consensys Software Inc.
// Licensed under the Apache License, Version 2.0. See the LICENSE file for details.

//go:build !arm64

package utils

// bitReverseTileBufBytes: half of the 32 KiB L1d of AMD Zen cores, a third of
// the 48 KiB L1d of recent Intel cores.
const bitReverseTileBufBytes = 16 << 10
