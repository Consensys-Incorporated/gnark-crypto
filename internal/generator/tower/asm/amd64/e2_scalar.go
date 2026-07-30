// Copyright 2020-2025 Consensys Software Inc.
// Licensed under the Apache License, Version 2.0. See the LICENSE file for details.

package amd64

import (
	"fmt"
	"math/bits"

	"github.com/consensys/bavard/amd64"
)

// generateMulByNonResidueE2Scalar generates mulNonResE2 for towers
// 𝔽p2 = 𝔽p[u]/(u² - β) with non-residue (0,1), i.e.
//
//	z.A0 = β·x.A1 (mod p)
//	z.A1 = x.A0
//
// β·x.A1 is computed by a double-and-add chain on the bits of |β| (adding
// x.A1 from memory), reducing after each step, negated at the end for β < 0.
func (fq2 *Fq2Amd64) generateMulByNonResidueE2Scalar(beta int64) {
	if beta == 0 || beta == 1 || beta == -1 {
		panic("use the dedicated generators for β = ±1")
	}
	neg := beta < 0
	abs := uint64(beta)
	if neg {
		abs = uint64(-beta)
	}

	registers := fq2.FnHeader("mulNonResE2", 0, 16)

	a := registers.PopN(fq2.NbWords)
	b := registers.PopN(fq2.NbWords)
	x := registers.Pop()
	tr := amd64.R15

	fq2.MOVQ("x+8(FP)", x)
	fq2.Mov(x, a, fq2.NbWords) // a = x.A1

	fq2.Comment(fmt.Sprintf("a = %d * x.A1 (double-and-add, reduced at each step)", abs))
	for i := bits.Len64(abs) - 2; i >= 0; i-- {
		fq2.Add(a, a)
		fq2.ReduceElement(a, concat(b, tr), true)
		if abs&(1<<i) != 0 {
			fq2.Add(x, a, fq2.NbWords)
			fq2.ReduceElement(a, concat(b, tr), true)
		}
	}

	src, tmp := a, b
	if neg {
		fq2.Comment("z.A0 = -a (zero-safe: 0 if a == 0, q - a otherwise)")
		lblZero := fq2.NewLabel()
		lblDone := fq2.NewLabel()
		fq2.MOVQ(a[0], tr)
		for i := 1; i < fq2.NbWords; i++ {
			fq2.ORQ(a[i], tr)
		}
		fq2.TESTQ(tr, tr)
		fq2.JEQ(lblZero)
		for i := 0; i < fq2.NbWords; i++ {
			fq2.MOVQ(fq2.F.Q[i], tr)
			if i == 0 {
				fq2.SUBQ(a[i], tr)
			} else {
				fq2.SBBQ(a[i], tr)
			}
			fq2.MOVQ(tr, b[i])
		}
		fq2.JMP(lblDone)
		fq2.LABEL(lblZero)
		fq2.XORQ(tr, tr)
		for i := 0; i < fq2.NbWords; i++ {
			fq2.MOVQ(tr, b[i])
		}
		fq2.LABEL(lblDone)
		src, tmp = b, a
	}

	// all reads of x happen before any write to res (res may alias x)
	fq2.Mov(x, tmp) // tmp = x.A0

	fq2.MOVQ("res+0(FP)", x)
	fq2.Mov(src, x)                 // z.A0 = β·x.A1
	fq2.Mov(tmp, x, 0, fq2.NbWords) // z.A1 = x.A0

	fq2.RET()
}

// generateSquareE2Scalar generates squareAdxE2 for towers
// 𝔽p2 = 𝔽p[u]/(u² - β) with small odd β < 0:
//
//	z.A1 = 2·x.A0·x.A1
//	z.A0 = (x.A0 + x.A1)·(x.A0 + β·x.A1) - ((1+β)/2)·z.A1 = x.A0² + β·x.A1²
//
// It mirrors generateSquareE2 (β = -1), with the b operand generalized to
// x.A0 - |β|·x.A1 and a final adjustment of -(1+β)/2 additions of z.A1.
func (fq2 *Fq2Amd64) generateSquareE2Scalar(beta int64, forceCheck bool) {
	if beta >= 0 || beta == -1 || beta%2 == 0 {
		panic("generateSquareE2Scalar expects small odd β < -1")
	}
	abs := uint64(-beta)
	k := (1 + beta) / 2 // z.A0 = c0 - k·z.A1; k < 0 here so we add (-k) times

	const argSize = 16
	minStackSize := fq2.NbWords * 3 * 8 // q stack + a0a1 + x.A1 copy
	stackSize := fq2.StackSize(fq2.NbWords*5-1, 2, minStackSize)
	registers := fq2.FnHeader("squareAdxE2", stackSize, argSize, amd64.DX, amd64.AX)
	registers.UnsafePush(amd64.R15)
	defer fq2.AssertCleanStack(stackSize, minStackSize)
	fq2.WriteLn("NO_LOCAL_POINTERS")

	// check ADX instruction support
	lblNoAdx := fq2.NewLabel()
	if forceCheck {
		fq2.CMPB("·supportAdx(SB)", 1)
		fq2.JNE(lblNoAdx)
	}

	fq2.WriteLn(fmt.Sprintf(`
	// z.A1 = 2 * x.A0 * x.A1
	// z.A0 = (x.A0 + x.A1) * (x.A0 - %d·x.A1) + %d·(x.A0·x.A1) = x.A0² - %d·x.A1²
	`, abs, -2*k, abs))

	// used in the mul operation
	op1 := fq2.PopN(&registers)
	res := fq2.PopN(&registers)

	op1At := func(i int) string {
		return string(op1[i])
	}

	ax := amd64.AX
	dx := amd64.DX

	qStack := fq2.PopN(&registers, true)
	// move q to the stack
	for i := 0; i < fq2.NbWords; i++ {
		fq2.MOVQ(fmt.Sprintf("$const_q%d", i), ax)
		fq2.MOVQ(ax, qStack[i])
	}
	fq2.SetQStack(qStack)
	defer fq2.UnsetQStack()

	fq2.Comment("2 * x.A0 * x.A1")
	fq2.MOVQ("x+8(FP)", ax)

	fq2.LabelRegisters("2 * x.A1", op1...)
	fq2.Mov(ax, op1, fq2.NbWords)
	fq2.Add(op1, op1) // op1, no reduce

	fq2.MulADX(&registers, op1At, func(i int) string {
		fq2.MOVQ("x+8(FP)", dx)
		return dx.At(i)
	}, res)
	fq2.ReduceElement(res, concat(op1, dx), true)

	fq2.MOVQ("x+8(FP)", ax)

	fq2.LabelRegisters("x.A1", op1...)
	fq2.Mov(ax, op1, fq2.NbWords)

	// save a copy of x.A1 on the stack: res may alias x, and z.A1 (written
	// below) overwrites x.A1 in that case.
	xA1 := fq2.PopN(&registers, true)
	fq2.Mov(op1, xA1)

	fq2.MOVQ("res+0(FP)", dx)
	fq2.Mov(res, dx, 0, fq2.NbWords)
	fq2.Mov(op1, res)

	// op1 and res both contain x.A1 at this point
	// res+0(FP) (z.A1) and x+8(FP) A1 words must not be referenced.

	// a = a0 + a1
	fq2.Comment("Add(&x.A0, &x.A1)")
	fq2.Add(ax, op1)
	//--> must save on stack
	a0a1 := fq2.PopN(&registers, true)
	fq2.Mov(op1, a0a1)

	// res = |β| * x.A1 (double-and-add from the stack copy, reduced at each step)
	fq2.Comment(fmt.Sprintf("%d * x.A1", abs))
	for i := bits.Len64(abs) - 2; i >= 0; i-- {
		fq2.Add(res, res)
		fq2.ReduceElement(res, concat(op1, dx), true)
		if abs&(1<<i) != 0 {
			fq2.Add(xA1, res)
			fq2.ReduceElement(res, concat(op1, dx), true)
		}
	}

	zero := dx
	fq2.XORQ(zero, zero)

	// b = a0 - |β|·a1
	fq2.Comment(fmt.Sprintf("Sub(&x.A0, %d·&x.A1)", abs))
	fq2.Mov(ax, op1)
	fq2.Sub(res, op1)
	fq2.modReduceAfterSubScratch(zero, op1, res) // using res as scratch registers

	// a = a * b
	fq2.MulADX(&registers, op1At, func(i int) string { return string(a0a1[i]) }, res)
	fq2.ReduceElement(res, concat(op1, dx), true)

	// z.A0 = res + (-k)·z.A1 (z.A1 read back from memory)
	fq2.Comment(fmt.Sprintf("z.A0 = res + %d·z.A1", -k))
	fq2.MOVQ("res+0(FP)", ax)
	for i := int64(0); i < -k; i++ {
		fq2.Add(ax, res, fq2.NbWords)
		fq2.ReduceElement(res, concat(op1, dx), true)
	}
	fq2.Mov(res, ax)

	// result.a0 = a
	fq2.RET()

	fq2.UnsafePush(&registers, xA1...)
	fq2.UnsafePush(&registers, a0a1...)
	fq2.UnsafePush(&registers, op1...)
	fq2.UnsafePush(&registers, res...)
	fq2.UnsafePush(&registers, qStack...)

	if forceCheck {
		fq2.LABEL(lblNoAdx)
		fq2.MOVQ("res+0(FP)", amd64.AX)
		fq2.MOVQ(amd64.AX, "(SP)")
		fq2.MOVQ("x+8(FP)", amd64.AX)
		fq2.MOVQ(amd64.AX, "8(SP)")
		fq2.WriteLn("CALL ·squareGenericE2(SB)")
		fq2.RET()
	}
}
