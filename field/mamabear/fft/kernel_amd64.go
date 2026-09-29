//go:build !purego

// Copyright 2020-2026 Consensys Software Inc.
// Licensed under the Apache License, Version 2.0. See the LICENSE file for details.

// This file is hand-written and is NOT produced by go generate, unlike the rest
// of this package; see config.Field.HandwrittenVectorASMAMD64.

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

// The two _avx512 kernels ignore their start argument: each processes
// ⌊end/8⌋ blocks of 8 elements from the base pointer it is given, pairing
// a[j] with a[j+m] and applying twiddles[j].
//
// That is not the contract the callers in fft.go have. The parallel butterfly
// path fans out over parallel.Execute, handing each goroutine a distinct
// [start, end) sub-range of the same backing array, so passing &a[0] and
// letting the kernel run [0, end) makes every goroutine reprocess the low part
// of the range — overlapping writes, non-deterministic results.
//
// Rather than teach the assembly about start, shift the window into it: pass
// &a[start] and &twiddles[start] and a length of end-start. The kernel's
// a[j]/a[j+m] pairing and twiddles[j] indexing are all relative to the base it
// receives, and m is unchanged, so the offset window computes exactly
// [start, end). The ragged tail (end-start not a multiple of 8) goes to the
// generic path, which is also where a window too short to fill one block goes.
//
// Note the kernels multiply every lane by its twiddle, including lane 0, while
// the generic implementations special-case start == 0 with an untwiddled
// butterfly. Those agree because twiddles[0] is 1.

func innerDIFWithTwiddles(a []mamabear.Element, twiddles []mamabear.Element, start, end, m int) {
	n := end - start
	if !cpu.SupportAVX512IFMA || m < 8 || n < 8 {
		innerDIFWithTwiddlesGeneric(a, twiddles, start, end, m)
		return
	}
	tail := n % 8
	innerDIFWithTwiddles_avx512(&a[start], &twiddles[start], 0, n-tail, m)
	if tail != 0 {
		innerDIFWithTwiddlesGeneric(a, twiddles, end-tail, end, m)
	}
}

func innerDITWithTwiddles(a []mamabear.Element, twiddles []mamabear.Element, start, end, m int) {
	n := end - start
	if !cpu.SupportAVX512IFMA || m < 8 || n < 8 {
		innerDITWithTwiddlesGeneric(a, twiddles, start, end, m)
		return
	}
	tail := n % 8
	innerDITWithTwiddles_avx512(&a[start], &twiddles[start], 0, n-tail, m)
	if tail != 0 {
		innerDITWithTwiddlesGeneric(a, twiddles, end-tail, end, m)
	}
}

// kerDIFNP_32 and kerDITNP_32 have no IFMA kernel: every stage of a 32-point
// transform has m < 8, so the SIMD inner loop would never be taken. They exist
// because the generated fft.go dispatches to them; they forward to the generic
// implementation that kernel_purego.go would otherwise provide.
func kerDIFNP_32(a []mamabear.Element, twiddles [][]mamabear.Element, stage int) {
	kerDIFNP_32generic(a, twiddles, stage)
}

func kerDITNP_32(a []mamabear.Element, twiddles [][]mamabear.Element, stage int) {
	kerDITNP_32generic(a, twiddles, stage)
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
