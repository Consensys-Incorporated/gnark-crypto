// Copyright 2020-2026 Consensys Software Inc.
// Licensed under the Apache License, Version 2.0. See the LICENSE file for details.

// MamaBear prime field: p = 2^49 − 2^34 + 1
//
// This package implements arithmetic in the MamaBear prime field, a 49-bit prime
// co-designed for the AVX-512IFMA instruction set (VPMADD52LUQ/HUQ).
//
// The Montgomery constant is R = 2^52, matching the 52-bit IFMA operand width.
// This leaves 3 bits of "headroom" (49 + 3 = 52) enabling lazy reduction:
// intermediate sums can be deferred as long as they stay below 2^52 (additions)
// or R + p ≈ 2^52 + 2^49 (after multiplication).
//
// The degree-3 extension field F_{p^3} = F_p[x]/(x^3 − x − 1) gives ~147-bit
// security, well above the 128-bit threshold.
//
// Reference: MamaBearZKP: A Holistic Co-design of Prime Fields and Proving
// Stacks for High-Throughput ZKP on Modern CPUs (Zhang et al., 2026).
//
// Hand-written; see internal/generator/field/ for future codegen integration.
package mamabear
