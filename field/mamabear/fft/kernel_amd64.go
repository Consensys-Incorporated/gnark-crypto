//go:build !purego

// Copyright 2020-2026 Consensys Software Inc.
// Licensed under the Apache License, Version 2.0. See the LICENSE file for details.

package fft

import (
	"github.com/consensys/gnark-crypto/field/mamabear"
	"github.com/consensys/gnark-crypto/utils/cpu"
)

// Package-level constants exposed to the assembler via go_asm.h.
// q = p = 2^49 − 2^34 + 1; qInvNeg = −p^{−1} mod 2^52.
const q = uint64(562932773552129)
const qInvNeg = uint64(562932773552127)

// butterfly sets a = a + b (mod q) and b = a - b (mod q).
func butterfly(a, b *mamabear.Element) {
	t := *a
	a.Add(a, b)
	b.Sub(&t, b)
}

//go:noescape
func innerDIFWithTwiddles_avx512(a, twiddles *mamabear.Element, start, end, m int)

//go:noescape
func innerDITWithTwiddles_avx512(a, twiddles *mamabear.Element, start, end, m int)

// innerDIFWithTwiddles dispatches to the AVX-512IFMA kernel when supported
// and m >= 8 (at least one full 8-element ZMM block per half-array).
// start is always 0 in the SIMD path; the generic fallback handles start != 0.
func innerDIFWithTwiddles(a []mamabear.Element, twiddles []mamabear.Element, start, end, m int) {
	if !cpu.SupportAVX512IFMA || m < 8 {
		innerDIFWithTwiddlesGeneric(a, twiddles, start, end, m)
		return
	}
	innerDIFWithTwiddles_avx512(&a[0], &twiddles[0], start, end, m)
}

// innerDITWithTwiddles dispatches to the AVX-512IFMA kernel when supported
// and m >= 8.
func innerDITWithTwiddles(a []mamabear.Element, twiddles []mamabear.Element, start, end, m int) {
	if !cpu.SupportAVX512IFMA || m < 8 {
		innerDITWithTwiddlesGeneric(a, twiddles, start, end, m)
		return
	}
	innerDITWithTwiddles_avx512(&a[0], &twiddles[0], start, end, m)
}

// kerDIFNP_256 is an unrolled 256-element DIF kernel.
// Stages with m >= 8 use the AVX-512IFMA inner loops; smaller stages fall back to scalar.
func kerDIFNP_256(a []mamabear.Element, twiddles [][]mamabear.Element, stage int) {
	innerDIFWithTwiddles(a[:256], twiddles[stage+0], 0, 128, 128)
	for offset := 0; offset < 256; offset += 128 {
		innerDIFWithTwiddles(a[offset:offset+128], twiddles[stage+1], 0, 64, 64)
	}
	for offset := 0; offset < 256; offset += 64 {
		innerDIFWithTwiddles(a[offset:offset+64], twiddles[stage+2], 0, 32, 32)
	}
	for offset := 0; offset < 256; offset += 32 {
		innerDIFWithTwiddles(a[offset:offset+32], twiddles[stage+3], 0, 16, 16)
	}
	for offset := 0; offset < 256; offset += 16 {
		innerDIFWithTwiddles(a[offset:offset+16], twiddles[stage+4], 0, 8, 8)
	}
	for offset := 0; offset < 256; offset += 8 {
		innerDIFWithTwiddles(a[offset:offset+8], twiddles[stage+5], 0, 4, 4)
	}
	for offset := 0; offset < 256; offset += 4 {
		innerDIFWithTwiddles(a[offset:offset+4], twiddles[stage+6], 0, 2, 2)
	}
	for offset := 0; offset < 256; offset += 2 {
		butterfly(&a[offset], &a[offset+1])
	}
}

// kerDITNP_256 is an unrolled 256-element DIT kernel.
// Stages with m >= 8 use the AVX-512IFMA inner loops; smaller stages fall back to scalar.
func kerDITNP_256(a []mamabear.Element, twiddles [][]mamabear.Element, stage int) {
	for offset := 0; offset < 256; offset += 2 {
		butterfly(&a[offset], &a[offset+1])
	}
	for offset := 0; offset < 256; offset += 4 {
		innerDITWithTwiddles(a[offset:offset+4], twiddles[stage+6], 0, 2, 2)
	}
	for offset := 0; offset < 256; offset += 8 {
		innerDITWithTwiddles(a[offset:offset+8], twiddles[stage+5], 0, 4, 4)
	}
	for offset := 0; offset < 256; offset += 16 {
		innerDITWithTwiddles(a[offset:offset+16], twiddles[stage+4], 0, 8, 8)
	}
	for offset := 0; offset < 256; offset += 32 {
		innerDITWithTwiddles(a[offset:offset+32], twiddles[stage+3], 0, 16, 16)
	}
	for offset := 0; offset < 256; offset += 64 {
		innerDITWithTwiddles(a[offset:offset+64], twiddles[stage+2], 0, 32, 32)
	}
	for offset := 0; offset < 256; offset += 128 {
		innerDITWithTwiddles(a[offset:offset+128], twiddles[stage+1], 0, 64, 64)
	}
	innerDITWithTwiddles(a[:256], twiddles[stage+0], 0, 128, 128)
}
