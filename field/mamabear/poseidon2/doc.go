// Copyright 2020-2026 Consensys Software Inc.
// Licensed under the Apache License, Version 2.0. See the LICENSE file for details.

// Package poseidon2 implements the Poseidon2 permutation for the MamaBear field.
//
// Poseidon2 is a cryptographic permutation for algebraic hashes.
// See the [original paper] by Grassi, Khovratovich and Schofnegger for full details.
//
// This is a pure-Go implementation; no SIMD assembly is generated for MamaBear
// (which uses R=2^52 Montgomery arithmetic, incompatible with the F31 AVX-512
// generator used for KoalaBear and BabyBear).
//
// [original paper]: https://eprint.iacr.org/2023/323.pdf
package poseidon2
