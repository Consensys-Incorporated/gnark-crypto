// Copyright 2020-2025 Consensys Software Inc.
// Licensed under the Apache License, Version 2.0. See the LICENSE file for details.

package amd64

import (
	"fmt"
	"io"

	"github.com/consensys/bavard"
	ramd64 "github.com/consensys/bavard/amd64"
	"github.com/consensys/gnark-crypto/internal/generator/config"
	"github.com/consensys/gnark-crypto/internal/generator/field/asm/amd64"
	field "github.com/consensys/gnark-crypto/internal/generator/field/config"
)

// Fq2Amd64 ...
type Fq2Amd64 struct {
	*amd64.FFAmd64
	config config.Curve
	w      io.Writer
	F      *field.Field
}

// NewFq2Amd64 ...
func NewFq2Amd64(w io.Writer, F *field.Field, config config.Curve) *Fq2Amd64 {
	return &Fq2Amd64{
		amd64.NewFFAmd64(w, F.NbWords),
		config,
		w,
		F,
	}
}

// Generate ...
func (fq2 *Fq2Amd64) Generate(forceADXCheck bool) error {
	fq2.WriteLn(bavard.Apache2Header("Consensys Software Inc.", 2020))

	fq2.WriteLn("#include \"textflag.h\"")
	fq2.WriteLn("#include \"funcdata.h\"")
	fq2.WriteLn("#include \"go_asm.h\"")

	fq2.GenerateReduceDefine()
	fq2.GenerateReduceDefine(true)

	fq2.generateAddE2()
	fq2.generateDoubleE2()
	fq2.generateSubE2()
	fq2.generateNegE2()

	if fq2.config.Equal(config.BN254) {
		// BN254 keeps the fused-CIOS mul: at 4 words the whole multiplication
		// is register-resident and measured faster than the lazy-reduction
		// variant (E2Mul 51.6 vs 53.4 ns on Zen 4).
		fq2.generateMulByNonResidueE2BN254()
		fq2.generateMulE2BN254(forceADXCheck)
		fq2.generateSquareE2(forceADXCheck)
	} else if fq2.config.Equal(config.BLS12_381) {
		fq2.generateMulByNonResidueE2BLS381()
		fq2.generateMulE2Lazy(-1, forceADXCheck)
		fq2.generateSquareE2(forceADXCheck)
	} else if fq2.config.Equal(config.BLS12_377) {
		// 𝔽p2 = 𝔽p[u]/(u² + 5)
		fq2.generateMulByNonResidueE2Scalar(-5)
		fq2.generateMulE2Lazy(-5, forceADXCheck)
		fq2.generateSquareE2Scalar(-5, forceADXCheck)
	} else if fq2.config.Equal(config.BLS24_315) {
		// 𝔽p2 = 𝔽p[u]/(u² - 13)
		fq2.generateMulByNonResidueE2Scalar(13)
		fq2.generateMulE2Lazy(13, forceADXCheck)
	} else if fq2.config.Equal(config.BLS24_317) {
		// 𝔽p2 = 𝔽p[u]/(u² + 1), same shape as BLS12-381
		fq2.generateMulByNonResidueE2BLS381()
		fq2.generateMulE2Lazy(-1, forceADXCheck)
		fq2.generateSquareE2(forceADXCheck)
	}
	return nil
}

func (fq2 *Fq2Amd64) generateAddE2() {
	const avoidGlobal = true
	stackSize := fq2.StackSize(fq2.NbWords*2+2, 0, 0)
	registers := fq2.FnHeader("addE2", stackSize, 24)
	defer fq2.AssertCleanStack(stackSize, 0)
	registers.UnsafePush(ramd64.R15) // we are not using global variables

	// registers
	x := registers.Pop()
	y := registers.Pop()
	t := registers.PopN(fq2.NbWords)

	fq2.MOVQ("x+8(FP)", x)

	// move t = x
	fq2.Mov(x, t)

	fq2.MOVQ("y+16(FP)", y)

	// t = t + y = x + y
	fq2.Add(y, t)

	// reduce
	fq2.Reduce(&registers, t, avoidGlobal)

	r := registers.Pop()
	fq2.MOVQ("res+0(FP)", r)
	fq2.Mov(t, r)

	// move x+offset(fq2.NbWords) into t
	fq2.Mov(x, t, fq2.NbWords)

	// add y+offset(fq2.NbWords) into t
	fq2.Add(y, t, fq2.NbWords)

	registers.Push(x)
	registers.Push(y)

	// reduce t into r with offset fq2.NbWords
	fq2.Reduce(&registers, t, avoidGlobal)
	fq2.Mov(t, r, 0, fq2.NbWords)

	fq2.RET()

}

func (fq2 *Fq2Amd64) generateDoubleE2() {
	const avoidGlobal = true
	stackSize := fq2.StackSize(fq2.NbWords*2+2, 0, 0)
	registers := fq2.FnHeader("doubleE2", stackSize, 16)
	registers.UnsafePush(ramd64.R15) // we are not using global variables
	defer fq2.AssertCleanStack(stackSize, 0)

	// registers
	x := registers.Pop()
	r := registers.Pop()
	t := registers.PopN(fq2.NbWords)

	fq2.MOVQ("res+0(FP)", r)
	fq2.MOVQ("x+8(FP)", x)

	fq2.Mov(x, t)
	fq2.Add(t, t)
	fq2.Reduce(&registers, t, avoidGlobal)
	fq2.Mov(t, r)
	fq2.Mov(x, t, fq2.NbWords)
	fq2.Add(t, t)
	fq2.Reduce(&registers, t, avoidGlobal)
	fq2.Mov(t, r, 0, fq2.NbWords)

	fq2.RET()
}

func (fq2 *Fq2Amd64) generateNegE2() {
	stackSize := fq2.StackSize(fq2.NbWords+3, 0, 0)
	registers := fq2.FnHeader("negE2", stackSize, 16)
	defer fq2.AssertCleanStack(stackSize, 0)

	nonZeroA := fq2.NewLabel()
	nonZeroB := fq2.NewLabel()
	B := fq2.NewLabel()

	// registers
	x := registers.Pop()
	r := registers.Pop()
	q := registers.Pop()
	t := registers.PopN(fq2.NbWords)

	fq2.MOVQ("res+0(FP)", r)
	fq2.MOVQ("x+8(FP)", x)

	// t = x
	fq2.Mov(x, t)

	// x = t[0] | ... | t[n]
	fq2.MOVQ(t[0], x)
	for i := 1; i < fq2.NbWords; i++ {
		fq2.ORQ(t[i], x)
	}

	fq2.TESTQ(x, x)

	// if x != 0, we jump to nonzero label
	fq2.JNE(nonZeroA)

	// if x == 0, we set the result to zero and continue
	for i := 0; i < fq2.NbWords; i++ {
		fq2.MOVQ(x, r.At(i))
	}
	fq2.JMP(B)

	fq2.LABEL(nonZeroA)

	// z = x - q
	for i := 0; i < fq2.NbWords; i++ {
		fq2.MOVQ(fq2.F.Q[i], q)
		if i == 0 {
			fq2.SUBQ(t[i], q)
		} else {
			fq2.SBBQ(t[i], q)
		}
		fq2.MOVQ(q, r.At(i))
	}

	fq2.LABEL(B)
	fq2.MOVQ("x+8(FP)", x)
	fq2.Mov(x, t, fq2.NbWords)

	// x = t[0] | ... | t[n]
	fq2.MOVQ(t[0], x)
	for i := 1; i < fq2.NbWords; i++ {
		fq2.ORQ(t[i], x)
	}

	fq2.TESTQ(x, x)

	// if x != 0, we jump to nonzero label
	fq2.JNE(nonZeroB)

	// if x == 0, we set the result to zero and return
	for i := 0; i < fq2.NbWords; i++ {
		fq2.MOVQ(x, r.At(i+fq2.NbWords))
	}
	fq2.RET()

	fq2.LABEL(nonZeroB)

	// z = x - q
	for i := 0; i < fq2.NbWords; i++ {
		fq2.MOVQ(fq2.F.Q[i], q)
		if i == 0 {
			fq2.SUBQ(t[i], q)
		} else {
			fq2.SBBQ(t[i], q)
		}
		fq2.MOVQ(q, r.At(i+fq2.NbWords))
	}

	fq2.RET()

}

func (fq2 *Fq2Amd64) generateSubE2() {
	stackSize := fq2.StackSize(2*fq2.NbWords+1, 0, 0)
	registers := fq2.FnHeader("subE2", stackSize, 24)
	defer fq2.AssertCleanStack(stackSize, 0)

	// registers
	t := registers.PopN(fq2.NbWords)
	xy := registers.Pop()

	zero := ramd64.R15
	fq2.XORQ(zero, zero)

	fq2.MOVQ("x+8(FP)", xy)
	fq2.Mov(xy, t)

	// z = x - y mod q
	// move t = x
	fq2.MOVQ("y+16(FP)", xy)
	fq2.Sub(xy, t)
	fq2.MOVQ("x+8(FP)", xy)

	fq2.modReduceAfterSub(&registers, zero, t)

	r := registers.Pop()
	fq2.MOVQ("res+0(FP)", r)
	fq2.Mov(t, r)
	registers.Push(r)

	fq2.Mov(xy, t, fq2.NbWords)

	// z = x - y mod q
	// move t = x
	fq2.MOVQ("y+16(FP)", xy)
	fq2.Sub(xy, t, fq2.NbWords)

	fq2.modReduceAfterSub(&registers, zero, t)

	r = xy
	fq2.MOVQ("res+0(FP)", r)

	fq2.Mov(t, r, 0, fq2.NbWords)

	fq2.RET()

}

func (fq2 *Fq2Amd64) modReduceAfterSub(registers *ramd64.Registers, zero ramd64.Register, t []ramd64.Register) {
	q := registers.PopN(fq2.NbWords)
	fq2.modReduceAfterSubScratch(zero, t, q)
	registers.Push(q...)
}

func (fq2 *Fq2Amd64) modReduceAfterSubScratch(zero ramd64.Register, t, scratch []ramd64.Register) {
	fq2.Mov(fq2.F.Q, scratch)
	for i := 0; i < fq2.NbWords; i++ {
		fq2.CMOVQCC(zero, scratch[i])
	}
	// add registers (q or 0) to t, and set to result
	fq2.Add(scratch, t)
}

func (fq2 *Fq2Amd64) generateSquareE2(forceCheck bool) {
	// // Square sets z to the E2-product of x,x returns z
	// func (z *E2) Square(x *E2) *E2 {
	// 	// adapted from algo 22 https://eprint.iacr.org/2010/354.pdf
	// 	var a, b fp.Element
	// 	a.Add(&x.A0, &x.A1)
	// 	b.Sub(&x.A0, &x.A1)
	// 	a.Mul(&a, &b)
	// 	b.Mul(&x.A0, &x.A1).Double(&b)
	// 	z.A0.Set(&a)
	// 	z.A1.Set(&b)
	// 	return z
	// }
	const argSize = 16
	minStackSize := fq2.NbWords * 2 * 8 // q stack + a0a1
	stackSize := fq2.StackSize(fq2.NbWords*4-1, 2, minStackSize)
	registers := fq2.FnHeader("squareAdxE2", stackSize, argSize, ramd64.DX, ramd64.AX)
	registers.UnsafePush(ramd64.R15)
	defer fq2.AssertCleanStack(stackSize, minStackSize)
	fq2.WriteLn("NO_LOCAL_POINTERS")

	// check ADX instruction support
	lblNoAdx := fq2.NewLabel()
	if forceCheck {
		fq2.CMPB("·supportAdx(SB)", 1)
		fq2.JNE(lblNoAdx)
	}

	fq2.WriteLn(`
	// z.A0 = (x.A0 + x.A1) * (x.A0 - x.A1)
	// z.A1 = 2 * x.A0 * x.A1
	`)

	// used in the mul operation
	op1 := fq2.PopN(&registers)
	res := fq2.PopN(&registers)

	op1At := func(i int) string {
		return string(op1[i])
	}

	ax := ramd64.AX
	dx := ramd64.DX

	qStack := fq2.PopN(&registers, true)
	// move q to the stack
	for i := 0; i < fq2.NbWords; i++ {
		fq2.MOVQ(fmt.Sprintf("$const_q%d", i), ax)
		fq2.MOVQ(ax, qStack[i])
	}
	fq2.SetQStack(qStack)
	defer fq2.UnsetQStack()

	// b = a0 * a1 * 2

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

	fq2.MOVQ("res+0(FP)", dx)
	fq2.Mov(res, dx, 0, fq2.NbWords)
	fq2.Mov(op1, res)

	// op1 and res both contains x.A1 at this point
	// res+0(FP) (z.A1) must not be referenced.

	// a = a0 + a1
	fq2.Comment("Add(&x.A0, &x.A1)")
	fq2.Add(ax, op1)
	//--> must save on stack
	a0a1 := fq2.PopN(&registers, true)
	fq2.Mov(op1, a0a1)

	zero := dx
	fq2.XORQ(zero, zero)

	// b = a0 - a1
	fq2.Comment("Sub(&x.A0, &x.A1)")
	fq2.Mov(ax, op1)
	fq2.Sub(res, op1)
	fq2.modReduceAfterSubScratch(zero, op1, res) // using res as scratch registers

	// a = a * b
	fq2.MulADX(&registers, op1At, func(i int) string { return string(a0a1[i]) }, res)
	fq2.ReduceElement(res, concat(op1, dx), true)

	fq2.MOVQ("res+0(FP)", ax)
	fq2.Mov(res, ax)

	// result.a0 = a
	fq2.RET()

	fq2.UnsafePush(&registers, a0a1...)
	fq2.UnsafePush(&registers, op1...)
	fq2.UnsafePush(&registers, res...)
	fq2.UnsafePush(&registers, qStack...)

	if forceCheck {
		fq2.LABEL(lblNoAdx)
		fq2.MOVQ("res+0(FP)", ramd64.AX)
		fq2.MOVQ(ramd64.AX, "(SP)")
		fq2.MOVQ("x+8(FP)", ramd64.AX)
		fq2.MOVQ(ramd64.AX, "8(SP)")
		fq2.WriteLn("CALL ·squareGenericE2(SB)")
		fq2.RET()
	}

}
