// Copyright 2020-2026 Consensys Incorporated
// Licensed under the Apache License, Version 2.0. See the LICENSE file for details.

package fft

import (
	"testing"

	"github.com/consensys/gnark-crypto/field/koalabear"
	fext "github.com/consensys/gnark-crypto/field/koalabear/extensions"
	"github.com/stretchr/testify/require"
)

// TestExt6SmallKernelSpecializations checks that the fused m=4/m=2/m=1 stages
// (VectorE6.Radix8DIT / Radix8DIF) match the generic per-stage butterflies,
// on whichever path this build uses (CI also runs with -tags=purego).
func TestExt6SmallKernelSpecializations(t *testing.T) {
	domain := NewDomain(1 << 9)

	// generic stages with half-size m on each block of 2m elements
	ditStage := func(a []fext.E6, tw []koalabear.Element, m int) {
		for off := 0; off < len(a); off += 2 * m {
			innerDITWithTwiddlesExt6(a[off:off+2*m], tw, 0, m, m)
		}
	}
	difStage := func(a []fext.E6, tw []koalabear.Element, m int) {
		for off := 0; off < len(a); off += 2 * m {
			innerDIFWithTwiddlesExt6(a[off:off+2*m], tw, 0, m, m)
		}
	}

	randomTwiddles := func(n int) []koalabear.Element {
		tw := make([]koalabear.Element, n)
		tw[0].SetOne()
		for i := 1; i < n; i++ {
			tw[i].MustSetRandom()
		}
		return tw
	}

	tw1 := randomTwiddles(1) // {1}

	run := func(t *testing.T) {
		t.Helper()
		for _, n := range []int{8, 16, 64, 72, 128, 200, 512} {
			for i := 0; i < 16; i++ {
				a := make([]fext.E6, n)
				for j := range a {
					a[j].MustSetRandom()
				}
				if i == 2 {
					// boundary values: every coordinate q-1 or 0
					var m1 koalabear.Element
					m1.SetOne().Neg(&m1)
					for j := range a {
						if j%3 == 0 {
							a[j].SetZero()
						} else {
							a[j].B0.A0, a[j].B0.A1 = m1, m1
							a[j].B1.A0, a[j].B1.A1 = m1, m1
							a[j].B2.A0, a[j].B2.A1 = m1, m1
						}
					}
				}
				tw4, tw2 := domain.twiddles[6], domain.twiddles[7]
				if i%2 == 1 {
					tw4, tw2 = randomTwiddles(4), randomTwiddles(2)
				}

				ref := append([]fext.E6(nil), a...)
				got := append([]fext.E6(nil), a...)
				ditStage(ref, tw1, 1)
				ditStage(ref, tw2, 2)
				ditStage(ref, tw4, 4)
				fext.VectorE6(got).Radix8DIT(tw4, tw2)
				require.Equal(t, ref, got, "DIT n=%d", n)

				ref = append(ref[:0], a...)
				got = append(got[:0], a...)
				difStage(ref, tw4, 4)
				difStage(ref, tw2, 2)
				difStage(ref, tw1, 1)
				fext.VectorE6(got).Radix8DIF(tw4, tw2)
				require.Equal(t, ref, got, "DIF n=%d", n)
			}
		}
	}

	run(t)
}
