// Copyright 2026 Consensys Software Inc.
// Licensed under the Apache License, Version 2.0. See the LICENSE file for details.

package bls12381

import (
	"fmt"
	"math"
	"testing"

	"github.com/consensys/gnark-crypto/ecc/bls12-381/fr"
)

// The base selector is a local closure, so reproduce its exact arithmetic here.
// Candidate selection calls the PR functions themselves. These values describe
// unsplit inputs; public MultiExp can select other windows after splitting.
func TestMSMBenchmarkSelection(t *testing.T) {
	for _, n := range msmBenchInts(t, "MSM_BENCH_SIZES", []int{1024, 1536, 2048, 4096, 8192, 16384, 32768}) {
		oldC := uint64(0)
		minimum := math.MaxFloat64
		for _, c := range msmBenchWindows {
			cost := float64((fr.Bits+1)*(n+(1<<c))) / float64(c)
			if cost < minimum {
				minimum, oldC = cost, uint64(c)
			}
		}
		fmt.Printf("selection,bls12-381,G1,%d,%d,%d\n", n, oldC, bestCG1(n))
		fmt.Printf("selection,bls12-381,G2,%d,%d,%d\n", n, oldC, bestCG2(n))
	}
}
