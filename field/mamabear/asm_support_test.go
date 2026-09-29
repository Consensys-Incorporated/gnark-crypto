// Copyright 2020-2026 Consensys Software Inc.
// Licensed under the Apache License, Version 2.0. See the LICENSE file for details.

// This file is hand-written and is NOT produced by go generate.

package mamabear

import (
	"testing"

	"github.com/consensys/gnark-crypto/utils/cpu"
)

// TestAVX512IFMADispatch reports whether this run actually exercised the
// hand-written AVX-512IFMA kernels in element_amd64.s (and, in the fft package,
// kernel_amd64.s).
//
// It never fails. Its only job is to make the distinction visible in a test log,
// because every correctness test in this package passes either way: the vector
// oracles in TestVectorOps compare Vector.Add/Sub/Mul/ScalarMul against the
// per-element Element operations, so they do check the assembly — but only on a
// host that dispatches to it. On anything else they check the generic path twice
// and say nothing about the asm.
//
// Run with -v to see which path was taken.
func TestAVX512IFMADispatch(t *testing.T) {
	if cpu.SupportAVX512IFMA {
		t.Log("AVX-512IFMA present: the vector and FFT assembly WAS exercised by this run")
		return
	}
	t.Log("AVX-512IFMA absent: the vector and FFT assembly was NOT exercised by this run; " +
		"every result here comes from the pure-Go fallback. " +
		"The kernels need a validating run on an Ice Lake+ or Zen4+ host.")
}
