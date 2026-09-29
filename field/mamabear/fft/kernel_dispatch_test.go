// Copyright 2020-2026 Consensys Software Inc.
// Licensed under the Apache License, Version 2.0. See the LICENSE file for details.

// Regression coverage for the AVX-512IFMA FFT dispatch honoring arbitrary
// [start, end) windows (cf. the start-window fix), and multi-task
// determinism of the full FFT.

package fft

import (
	"math/rand/v2"
	"testing"

	"github.com/consensys/gnark-crypto/field/mamabear"
	"github.com/consensys/gnark-crypto/utils/cpu"
)

func randPoly(rng *rand.Rand, n int) []mamabear.Element {
	p := make([]mamabear.Element, n)
	for i := range p {
		p[i].SetRandom()
	}
	return p
}

// TestReviewKernelWindowDispatch compares the (possibly asm-dispatched)
// innerDIF/innerDITWithTwiddles wrappers against the generic implementation on
// arbitrary unaligned [start, end) windows, for m >= 8 (SIMD-eligible) and
// windows with ragged tails.
func TestReviewKernelWindowDispatch(t *testing.T) {
	if !cpu.SupportAVX512IFMA {
		t.Skip("AVX-512IFMA absent: asm kernels not exercised")
	}
	rng := rand.New(rand.NewPCG(42, 43))

	for _, m := range []int{8, 16, 32, 64} {
		// backing array: two spans of m
		n := 2 * m
		tw := randPoly(rng, m)
		// generic special-cases j == 0 with an untwiddled butterfly; the asm
		// multiplies lane 0 by twiddles[0]. They agree iff twiddles[0] == 1.
		tw[0].SetOne()
		// windows: full, offset start, ragged end, single block, sub-block
		windows := [][2]int{
			{0, m}, {1, m}, {0, m - 1}, {3, m - 2},
			{0, 8}, {5, 13}, {7, 8}, {0, 7}, {m / 2, m},
		}
		for _, w := range windows {
			start, end := w[0], w[1]
			if start >= end || end > m {
				continue
			}
			aGot := randPoly(rng, n)
			aWant := make([]mamabear.Element, n)
			copy(aWant, aGot)

			innerDIFWithTwiddles(aGot, tw, start, end, m)
			innerDIFWithTwiddlesGeneric(aWant, tw, start, end, m)
			for i := range n {
				if aGot[i] != aWant[i] {
					t.Fatalf("DIF m=%d [%d,%d): index %d: got %d want %d",
						m, start, end, i, aGot[i][0], aWant[i][0])
				}
			}

			copy(aGot, aWant) // reset to same input
			aGot2 := randPoly(rng, n)
			copy(aGot, aGot2)
			copy(aWant, aGot2)
			innerDITWithTwiddles(aGot, tw, start, end, m)
			innerDITWithTwiddlesGeneric(aWant, tw, start, end, m)
			for i := range n {
				if aGot[i] != aWant[i] {
					t.Fatalf("DIT m=%d [%d,%d): index %d: got %d want %d",
						m, start, end, i, aGot[i][0], aWant[i][0])
				}
			}
		}
	}
}

// TestReviewFFTNbTasksDeterminism checks the full FFT gives bit-identical
// results for different task counts — the failure mode of the start-window
// bug was overlapping writes under parallel.Execute.
func TestReviewFFTNbTasksDeterminism(t *testing.T) {
	rng := rand.New(rand.NewPCG(7, 7))
	for _, logN := range []uint64{3, 4, 5, 8, 10, 12, 14} {
		n := uint64(1) << logN
		domain := NewDomain(n)
		src := randPoly(rng, int(n))
		refDIF := make([]mamabear.Element, n)
		copy(refDIF, src)
		domain.FFT(refDIF, DIF, OnCoset(), WithNbTasks(1))
		refDIT := make([]mamabear.Element, n)
		copy(refDIT, src)
		domain.FFT(refDIT, DIT, WithNbTasks(1))
		for _, tasks := range []int{2, 3, 7, 16, 64} {
			got := make([]mamabear.Element, n)
			copy(got, src)
			domain.FFT(got, DIF, OnCoset(), WithNbTasks(tasks))
			for i := range got {
				if got[i] != refDIF[i] {
					t.Fatalf("FFT DIF coset n=%d tasks=%d: index %d differs", n, tasks, i)
				}
			}
			copy(got, src)
			domain.FFT(got, DIT, WithNbTasks(tasks))
			for i := range got {
				if got[i] != refDIT[i] {
					t.Fatalf("FFT DIT n=%d tasks=%d: index %d differs", n, tasks, i)
				}
			}
		}
	}
}
