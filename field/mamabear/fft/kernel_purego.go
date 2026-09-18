// Copyright 2020-2026 Consensys Software Inc.
// Licensed under the Apache License, Version 2.0. See the LICENSE file for details.

package fft

import (
	"github.com/consensys/gnark-crypto/field/mamabear"
)

// butterfly sets a = a + b (mod q) and b = a - b (mod q).
func butterfly(a, b *mamabear.Element) {
	t := *a
	a.Add(a, b)
	b.Sub(&t, b)
}

func innerDIFWithTwiddles(a []mamabear.Element, twiddles []mamabear.Element, start, end, m int) {
	innerDIFWithTwiddlesGeneric(a, twiddles, start, end, m)
}

func innerDITWithTwiddles(a []mamabear.Element, twiddles []mamabear.Element, start, end, m int) {
	innerDITWithTwiddlesGeneric(a, twiddles, start, end, m)
}

func kerDIFNP_256(a []mamabear.Element, twiddles [][]mamabear.Element, stage int) {
	kerDIFNP_256generic(a, twiddles, stage)
}

func kerDITNP_256(a []mamabear.Element, twiddles [][]mamabear.Element, stage int) {
	kerDITNP_256generic(a, twiddles, stage)
}
