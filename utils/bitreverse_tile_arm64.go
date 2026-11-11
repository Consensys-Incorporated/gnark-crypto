// Copyright 2020-2025 Consensys Software Inc.
// Licensed under the Apache License, Version 2.0. See the LICENSE file for details.

package utils

// bitReverseTileBufBytes: half of the 64 KiB L1d of Neoverse (Graviton) cores;
// Apple cores have 128 KiB. Not yet measured on arm64.
const bitReverseTileBufBytes = 32 << 10
