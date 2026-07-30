// Copyright 2020-2025 Consensys Software Inc.
// Licensed under the Apache License, Version 2.0. See the LICENSE file for details.

package amd64

import (
	"github.com/consensys/bavard/amd64"
)

func (fq2 *Fq2Amd64) generateMulByNonResidueE2BLS381() {
	// // MulByNonResidue multiplies a E2 by (1,1)
	// func (z *E2) MulByNonResidue(x *E2) *E2 {
	// 	var a fp.Element
	// 	a.Sub(&x.A0, &x.A1)
	// 	z.A1.Add(&x.A0, &x.A1)
	// 	z.A0.Set(&a)
	// 	return z
	// }
	registers := fq2.FnHeader("mulNonResE2", 0, 16)

	a := registers.PopN(fq2.NbWords)
	b := registers.PopN(fq2.NbWords)
	x := registers.Pop()
	tr := amd64.R15  // zero or r
	fq2.XORQ(tr, tr) // set to zero

	fq2.MOVQ("x+8(FP)", x)
	fq2.Mov(x, a) // a = a0

	// a = x.A0 - x.A1
	fq2.Sub(x, a, fq2.NbWords)
	fq2.modReduceAfterSubScratch(tr, a, b)
	// b = x.A0 + x.A1
	fq2.Mov(x, b, fq2.NbWords) // b = a1
	fq2.Add(x, b)

	fq2.MOVQ("res+0(FP)", x)
	fq2.Mov(a, x)
	fq2.ReduceElement(b, concat(a, tr), true)
	fq2.Mov(b, x, 0, fq2.NbWords)

	fq2.RET()
}
