// Copyright 2020-2026 Consensys Software Inc.
// Licensed under the Apache License, Version 2.0. See the LICENSE file for details.

// Package fft provides Number Theoretic Transform (NTT) over the MamaBear prime field.
//
// The MamaBear prime p = 2^49 − 2^34 + 1 has 2-adicity 34, supporting NTT domains
// up to size 2^34. The package mirrors the structure of field/koalabear/fft.
package fft
