// Copyright 2020-2026 Consensys Software Inc.
// Licensed under the Apache License, Version 2.0. See the LICENSE file for details.

// Package element_49b contains shared AVX-512IFMA assembly for 49-bit prime fields
// with Montgomery constant R = 2^52.
//
// The assembly is parameterized by the following constants, which must be
// declared in the importing package so the Go assembler resolves them via go_asm.h:
//
//	const_q       — the prime p (uint64)
//	const_qInvNeg — −p^{-1} mod 2^52 (uint64)
//
// This package exists only to force go mod vendor to include the .s files.
package element_49b

// Dummy constants so this package compiles standalone (overridden by importers).
const (
	DUMMY   = 0
	q       = 0
	qInvNeg = 0
)
