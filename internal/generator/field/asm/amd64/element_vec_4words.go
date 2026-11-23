// Copyright 2020-2025 Consensys Incorporated
// Licensed under the Apache License, Version 2.0. See the LICENSE file for details.

package amd64

import (
	"fmt"

	"github.com/consensys/bavard/amd64"
)

// addVec res = a + b
// func addVec(res, a, b *{{.ElementName}}, n uint64)
func (f *FFAmd64) generateAddVecW4() {
	f.Comment("addVec(res, a, b *Element, n uint64) res[0...n] = a[0...n] + b[0...n]")

	const argSize = 4 * 8
	stackSize := f.StackSize(f.NbWords*2+4, 0, 0)
	registers := f.FnHeader("addVec", stackSize, argSize)
	defer f.AssertCleanStack(stackSize, 0)

	// registers & labels we need
	addrA := f.Pop(&registers)
	addrB := f.Pop(&registers)
	addrRes := f.Pop(&registers)
	len := f.Pop(&registers)

	a := f.PopN(&registers)
	t := f.PopN(&registers)

	loop := f.NewLabel("loop")
	done := f.NewLabel("done")

	// load arguments
	f.MOVQ("res+0(FP)", addrRes)
	f.MOVQ("a+8(FP)", addrA)
	f.MOVQ("b+16(FP)", addrB)
	f.MOVQ("n+24(FP)", len)

	f.LABEL(loop)

	f.TESTQ(len, len)
	f.JEQ(done, "n == 0, we are done")

	// a = a + b
	f.LabelRegisters("a", a...)
	f.Mov(addrA, a)
	f.Add(addrB, a)
	f.PREFETCHT0(fmt.Sprintf("2048(%s)", addrA))
	f.PREFETCHT0(fmt.Sprintf("2048(%s)", addrB))

	// reduce a
	f.ReduceElement(a, t, false)

	// save a into res
	f.Mov(a, addrRes)
	f.Comment("increment pointers to visit next element")
	f.ADDQ("$32", addrA)
	f.ADDQ("$32", addrB)
	f.ADDQ("$32", addrRes)
	f.DECQ(len, "decrement n")
	f.JMP(loop)

	f.LABEL(done)

	f.RET()

	f.Push(&registers, a...)
	f.Push(&registers, t...)
	f.Push(&registers, addrA, addrB, addrRes, len)

}

// subVec res = a - b
// func subVec(res, a, b *{{.ElementName}}, n uint64)
func (f *FFAmd64) generateSubVecW4() {
	f.Comment("subVec(res, a, b *Element, n uint64) res[0...n] = a[0...n] - b[0...n]")

	const argSize = 4 * 8
	stackSize := f.StackSize(f.NbWords*2+5, 0, 0)
	registers := f.FnHeader("subVec", stackSize, argSize)
	defer f.AssertCleanStack(stackSize, 0)

	// registers
	addrA := f.Pop(&registers)
	addrB := f.Pop(&registers)
	addrRes := f.Pop(&registers)
	len := f.Pop(&registers)
	zero := f.Pop(&registers)

	a := f.PopN(&registers)
	q := f.PopN(&registers)

	loop := f.NewLabel("loop")
	done := f.NewLabel("done")

	// load arguments
	f.MOVQ("res+0(FP)", addrRes)
	f.MOVQ("a+8(FP)", addrA)
	f.MOVQ("b+16(FP)", addrB)
	f.MOVQ("n+24(FP)", len)

	f.XORQ(zero, zero)

	f.LABEL(loop)

	f.TESTQ(len, len)
	f.JEQ(done, "n == 0, we are done")

	// a = a - b
	f.LabelRegisters("a", a...)
	f.Mov(addrA, a)
	f.Sub(addrB, a)
	f.PREFETCHT0(fmt.Sprintf("2048(%s)", addrA))
	f.PREFETCHT0(fmt.Sprintf("2048(%s)", addrB))

	// reduce a
	f.Comment("reduce (a-b) mod q")
	f.LabelRegisters("q", q...)
	for i := 0; i < f.NbWords; i++ {
		f.MOVQ(fmt.Sprintf("$const_q%d", i), q[i])
	}
	for i := 0; i < f.NbWords; i++ {
		f.CMOVQCC(zero, q[i])
	}
	// add registers (q or 0) to a, and set to result
	f.Comment("add registers (q or 0) to a, and set to result")
	f.Add(q, a)

	// save a into res
	f.Mov(a, addrRes)

	f.Comment("increment pointers to visit next element")
	f.ADDQ("$32", addrA)
	f.ADDQ("$32", addrB)
	f.ADDQ("$32", addrRes)
	f.DECQ(len, "decrement n")
	f.JMP(loop)

	f.LABEL(done)

	f.RET()

	f.Push(&registers, a...)
	f.Push(&registers, q...)
	f.Push(&registers, addrA, addrB, addrRes, len, zero)

}

// sumVec res = sum(a[0...n])
func (f *FFAmd64) generateSumVecW4() {
	f.Comment("sumVec(res, a *Element, n uint64) res = sum(a[0...n])")

	const argSize = 3 * 8
	stackSize := f.StackSize(12, 2, 0)
	registers := f.FnHeader("sumVec", stackSize, argSize, amd64.DX, amd64.AX)
	defer f.AssertCleanStack(stackSize, 0)

	f.WriteLn(`
	// Derived from https://github.com/a16z/vectorized-fields
	// The idea is to use Z registers to accumulate the sum of elements, 8 by 8
	// first, we handle the case where n % 8 != 0
	// then, we loop over the elements 8 by 8 and accumulate the sum in the Z registers
	// finally, we reduce the sum and store it in res
	// 
	// when we move an element of a into a Z register, we use VPMOVZXDQ
	// let's note w0...w3 the 4 64bits words of ai: w0 = ai[0], w1 = ai[1], w2 = ai[2], w3 = ai[3]
	// VPMOVZXDQ(ai, Z0) will result in 
	// Z0= [hi(w3), lo(w3), hi(w2), lo(w2), hi(w1), lo(w1), hi(w0), lo(w0)]
	// with hi(wi) the high 32 bits of wi and lo(wi) the low 32 bits of wi
	// we can safely add 2^32+1 times Z registers constructed this way without overflow
	// since each of this lo/hi bits are moved into a "64bits" slot
	// N = 2^64-1 / 2^32-1 = 2^32+1
	//
	// we then propagate the carry using ADOXQ and ADCXQ
	// r0 = w0l + lo(woh)
	// r1 = carry + hi(woh) + w1l + lo(w1h)
	// r2 = carry + hi(w1h) + w2l + lo(w2h)
	// r3 = carry + hi(w2h) + w3l + lo(w3h)
	// r4 = carry + hi(w3h)
	// we then reduce the sum using a single-word Barrett reduction
	// we pick mu = 2^288 / q; which correspond to 4.5 words max.
	// meaning we must guarantee that r4 fits in 32bits.
	// To do so, we reduce N to 2^32-1 (since r4 receives 2 carries max)
	`)

	// registers & labels we need
	addrA := f.Pop(&registers)
	n := f.Pop(&registers)
	nMod8 := f.Pop(&registers)

	loop := f.NewLabel("loop8by8")
	done := f.NewLabel("done")
	loopSingle := f.NewLabel("loop_single")
	accumulate := f.NewLabel("accumulate")

	// AVX512 registers
	Z0 := amd64.Register("Z0")
	Z1 := amd64.Register("Z1")
	Z2 := amd64.Register("Z2")
	Z3 := amd64.Register("Z3")
	Z4 := amd64.Register("Z4")
	Z5 := amd64.Register("Z5")
	Z6 := amd64.Register("Z6")
	Z7 := amd64.Register("Z7")
	Z8 := amd64.Register("Z8")

	X0 := amd64.Register("X0")

	// load arguments
	f.MOVQ("a+8(FP)", addrA)
	f.MOVQ("n+16(FP)", n)

	f.Comment("initialize accumulators Z0, Z1, Z2, Z3, Z4, Z5, Z6, Z7")
	f.VXORPS(Z0, Z0, Z0)
	f.VMOVDQA64(Z0, Z1)
	f.VMOVDQA64(Z0, Z2)
	f.VMOVDQA64(Z0, Z3)
	f.VMOVDQA64(Z0, Z4)
	f.VMOVDQA64(Z0, Z5)
	f.VMOVDQA64(Z0, Z6)
	f.VMOVDQA64(Z0, Z7)

	// note: we don't need to handle the case n==0; handled by caller already.
	// f.TESTQ(n, n)
	// f.JEQ(done, "n == 0, we are done")

	f.LabelRegisters("n % 8", nMod8)
	f.LabelRegisters("n / 8", n)
	f.MOVQ(n, nMod8)
	f.ANDQ("$7", nMod8) // nMod8 = n % 8
	f.SHRQ("$3", n)     // len = n / 8

	f.LABEL(loopSingle)
	f.TESTQ(nMod8, nMod8)
	f.JEQ(loop, "n % 8 == 0, we are going to loop over 8 by 8")

	f.VPMOVZXDQ("0("+addrA+")", Z8)
	f.VPADDQ(Z8, Z0, Z0)
	f.ADDQ("$32", addrA)

	f.DECQ(nMod8, "decrement nMod8")
	f.JMP(loopSingle)

	f.Push(&registers, nMod8) // we don't need tmp0

	f.LABEL(loop)
	f.TESTQ(n, n)
	f.JEQ(accumulate, "n == 0, we are going to accumulate")

	for i := range 8 {
		r := fmt.Sprintf("Z%d", i+8)
		f.VPMOVZXDQ(fmt.Sprintf("%d*32("+string(addrA)+")", i), r)
	}

	f.WriteLn(fmt.Sprintf("PREFETCHT0 4096(%[1]s)", addrA))
	for i := range 8 {
		r := fmt.Sprintf("Z%d", i)
		f.VPADDQ(fmt.Sprintf("Z%d", i+8), r, r)
	}

	f.Comment("increment pointers to visit next 8 elements")
	f.ADDQ("$256", addrA)
	f.DECQ(n, "decrement n")
	f.JMP(loop)

	f.Push(&registers, n, addrA)

	f.LABEL(accumulate)

	f.Comment("accumulate the 8 Z registers into Z0")
	f.VPADDQ(Z7, Z6, Z6)
	f.VPADDQ(Z6, Z5, Z5)
	f.VPADDQ(Z5, Z4, Z4)
	f.VPADDQ(Z4, Z3, Z3)
	f.VPADDQ(Z3, Z2, Z2)
	f.VPADDQ(Z2, Z1, Z1)
	f.VPADDQ(Z1, Z0, Z0)

	w0l := f.Pop(&registers)
	w0h := f.Pop(&registers)
	w1l := f.Pop(&registers)
	w1h := f.Pop(&registers)
	w2l := f.Pop(&registers)
	w2h := f.Pop(&registers)
	w3l := f.Pop(&registers)
	w3h := f.Pop(&registers)
	low0h := f.Pop(&registers)
	low1h := f.Pop(&registers)
	low2h := f.Pop(&registers)
	low3h := f.Pop(&registers)

	// Propagate carries
	f.Comment("carry propagation")

	f.LabelRegisters("lo(w0)", w0l)
	f.LabelRegisters("hi(w0)", w0h)
	f.LabelRegisters("lo(w1)", w1l)
	f.LabelRegisters("hi(w1)", w1h)
	f.LabelRegisters("lo(w2)", w2l)
	f.LabelRegisters("hi(w2)", w2h)
	f.LabelRegisters("lo(w3)", w3l)
	f.LabelRegisters("hi(w3)", w3h)

	f.VMOVQ(X0, w0l)
	f.VALIGNQ("$1", Z0, Z0, Z0)
	f.VMOVQ(X0, w0h)
	f.VALIGNQ("$1", Z0, Z0, Z0)
	f.VMOVQ(X0, w1l)
	f.VALIGNQ("$1", Z0, Z0, Z0)
	f.VMOVQ(X0, w1h)
	f.VALIGNQ("$1", Z0, Z0, Z0)
	f.VMOVQ(X0, w2l)
	f.VALIGNQ("$1", Z0, Z0, Z0)
	f.VMOVQ(X0, w2h)
	f.VALIGNQ("$1", Z0, Z0, Z0)
	f.VMOVQ(X0, w3l)
	f.VALIGNQ("$1", Z0, Z0, Z0)
	f.VMOVQ(X0, w3h)

	f.LabelRegisters("lo(hi(wo))", low0h)
	f.LabelRegisters("lo(hi(w1))", low1h)
	f.LabelRegisters("lo(hi(w2))", low2h)
	f.LabelRegisters("lo(hi(w3))", low3h)

	type hilo struct {
		hi, lo amd64.Register
	}

	splitLoHi := f.Define("SPLIT_LO_HI", 2, func(args ...any) {
		lo := args[0]
		hi := args[1]
		f.MOVQ(hi, lo)
		f.ANDQ("$0xffffffff", lo)
		f.SHLQ("$32", lo)
		f.SHRQ("$32", hi)
	}, true)

	for _, v := range []hilo{{w0h, low0h}, {w1h, low1h}, {w2h, low2h}, {w3h, low3h}} {
		splitLoHi(v.lo, v.hi)
	}

	f.WriteLn(`
	// r0 = w0l + lo(woh)
	// r1 = carry + hi(woh) + w1l + lo(w1h)
	// r2 = carry + hi(w1h) + w2l + lo(w2h)
	// r3 = carry + hi(w2h) + w3l + lo(w3h)
	// r4 = carry + hi(w3h)
	`)
	f.XORQ(amd64.AX, amd64.AX, "clear the flags")
	f.ADOXQ(low0h, w0l)

	f.ADOXQ(low1h, w1l)
	f.ADCXQ(w0h, w1l)

	f.ADOXQ(low2h, w2l)
	f.ADCXQ(w1h, w2l)

	f.ADOXQ(low3h, w3l)
	f.ADCXQ(w2h, w3l)

	f.ADOXQ(amd64.AX, w3h)
	f.ADCXQ(amd64.AX, w3h)

	r0 := w0l
	r1 := w1l
	r2 := w2l
	r3 := w3l
	r4 := w3h

	r := []amd64.Register{r0, r1, r2, r3, r4}
	f.LabelRegisters("r", r...)
	// we don't need w0h, w1h, w2h anymore
	f.Push(&registers, w0h, w1h, w2h)
	// we don't need the low bits anymore
	f.Push(&registers, low0h, low1h, low2h, low3h)

	// Reduce using single-word Barrett
	mu := f.Pop(&registers)

	f.Comment("reduce using single-word Barrett")
	f.Comment("see see Handbook of Applied Cryptography, Algorithm 14.42.")
	f.LabelRegisters("mu=2^288 / q", mu)
	f.MOVQ(f.mu(), mu)
	f.MOVQ(r3, amd64.AX)
	f.SHRQw("$32", r4, amd64.AX)
	f.MULQ(mu, "high bits of res stored in DX")

	f.MULXQ(f.qAt(0), amd64.AX, mu)
	f.SUBQ(amd64.AX, r0)
	f.SBBQ(mu, r1)

	f.MULXQ(f.qAt(2), amd64.AX, mu)
	f.SBBQ(amd64.AX, r2)
	f.SBBQ(mu, r3)
	f.SBBQ("$0", r4)

	f.MULXQ(f.qAt(1), amd64.AX, mu)
	f.SUBQ(amd64.AX, r1)
	f.SBBQ(mu, r2)

	f.MULXQ(f.qAt(3), amd64.AX, mu)
	f.SBBQ(amd64.AX, r3)
	f.SBBQ(mu, r4)

	// we need up to 2 conditional substractions to be < q
	modReduced := f.NewLabel("modReduced")
	t := f.PopN(&registers)
	f.Mov(r[:4], t) // backup r0 to r3 (our result)

	// sub modulus
	f.SUBQ(f.qAt(0), r0)
	f.SBBQ(f.qAt(1), r1)
	f.SBBQ(f.qAt(2), r2)
	f.SBBQ(f.qAt(3), r3)
	f.SBBQ("$0", r4)

	// if borrow, we go to mod reduced
	f.JCS(modReduced)
	f.Mov(r, t)
	f.SUBQ(f.qAt(0), r0)
	f.SBBQ(f.qAt(1), r1)
	f.SBBQ(f.qAt(2), r2)
	f.SBBQ(f.qAt(3), r3)
	f.SBBQ("$0", r4)

	// if borrow, we skip to the end
	f.JCS(modReduced)
	f.Mov(r, t)

	f.LABEL(modReduced)
	addrRes := mu
	f.MOVQ("res+0(FP)", addrRes)
	f.Mov(t, addrRes)

	f.LABEL(done)

	f.RETAVX()
	f.Push(&registers, mu)
	f.Push(&registers, w0l, w1l, w2l, w3l, w3h)
}

func (f *FFAmd64) generateInnerProductW4() {
	f.Comment("innerProdVec(res, a, b *Element, n uint64) res = sum(a[0...n] * b[0...n])")
	f.Comment("")
	f.Comment("Algorithm: Accumulate 32-bit x 32-bit partial products into 16 Z registers")
	f.Comment("(8 low halves + 8 high halves), then combine into 544-bit result and reduce.")
	f.Comment("")
	f.Comment("Register allocation:")
	f.Comment("  Z0, Z1 = output result (544-bit in Z1:Z0)")
	f.Comment("  Z2 = PPL (partial product low), Z3 = PPH (partial product high)")
	f.Comment("  Z4 = Y (loaded b element), Z5 = LSW (low 32-bit mask)")
	f.Comment("  Z16-Z23 = accL[0-7], Z24-Z31 = accH[0-7]")

	const argSize = 4 * 8
	stackSize := f.StackSize(7, 2, 0)
	registers := f.FnHeader("innerProdVec", stackSize, argSize, amd64.DX, amd64.AX)
	defer f.AssertCleanStack(stackSize, 0)

	// Scalar registers
	addrA := f.Pop(&registers)
	addrB := f.Pop(&registers)
	n := f.Pop(&registers)

	done := f.NewLabel("done")
	accumulate := f.NewLabel("accumulate")

	// Explicit AVX512 register assignments to avoid conflicts
	// Z0, Z1 are reserved for output (544-bit result)
	PPL := amd64.Register("Z2")
	PPH := amd64.Register("Z3")
	Y := amd64.Register("Z4")
	LSW := amd64.Register("Z5")

	// 16 accumulators: use Z16-Z23 for low halves, Z24-Z31 for high halves
	// This keeps them separate from the output and temp registers
	ACC := amd64.Register("Z16") // ACC aliases accL[0]
	accL := []amd64.Register{"Z16", "Z17", "Z18", "Z19", "Z20", "Z21", "Z22", "Z23"}
	accH := []amd64.Register{"Z24", "Z25", "Z26", "Z27", "Z28", "Z29", "Z30", "Z31"}

	f.Comment("Load arguments")
	f.MOVQ("a+8(FP)", addrA)
	f.MOVQ("b+16(FP)", addrB)
	f.MOVQ("n+24(FP)", n)

	f.Comment("Create mask for low dword in each qword: 0x00000000FFFFFFFF")
	f.VPCMPEQB("Y0", "Y0", "Y0")
	f.VPMOVZXDQ("Y0", LSW)

	f.Comment("Initialize all 16 accumulators to zero")
	f.VPXORQ(accL[0], accL[0], accL[0])
	for i := 1; i < 8; i++ {
		f.VMOVDQA64(accL[0], accL[i])
	}
	for i := range 8 {
		f.VMOVDQA64(accL[0], accH[i])
	}

	f.Comment("Main loop: multiply and accumulate partial products")
	f.Comment("For each element pair (a[i], b[i]):")
	f.Comment("  - Load b[i] as 8 dwords zero-extended to qwords")
	f.Comment("  - Multiply each dword of a[i] by all dwords of b[i]")
	f.Comment("  - Split results into low/high 32 bits and accumulate")

	// Define MAC (multiply-accumulate) macro
	mac := f.Define("MAC", 3, func(inputs ...any) {
		opLeft := inputs[0] // memory operand for dword of a
		lo := inputs[1]     // accumulator for low 32 bits
		hi := inputs[2]     // accumulator for high 32 bits

		f.VPMULUDQ_BCST(opLeft, Y, PPL) // PPL = a_dword * b_dwords (64-bit results)
		f.VPSRLQ("$32", PPL, PPH)       // PPH = high 32 bits
		f.VPANDQ(LSW, PPL, PPL)         // PPL = low 32 bits
		f.VPADDQ(PPL, lo, lo)           // accumulate low
		f.VPADDQ(PPH, hi, hi)           // accumulate high
	}, true)

	// MAC_LAST is identical to MAC, but uses a 4-byte broadcast load (VPBROADCASTD)
	// instead of VPMULUDQ.BCST, which loads an 8-byte qword. The final dword of each
	// element sits at offset 28 of a 32-byte element; an 8-byte broadcast there would
	// read 4 bytes past the element. For interior elements those bytes belong to the
	// next element and are harmless; for the last element of the input slice they may
	// cross a page boundary and SIGSEGV.
	macLast := f.Define("MAC_LAST", 3, func(inputs ...any) {
		opLeft := inputs[0]
		lo := inputs[1]
		hi := inputs[2]

		f.VPBROADCASTD(opLeft, PPL) // 4-byte broadcast: PPL[i] = a_dword in all 16 dword lanes
		f.VPMULUDQ(PPL, Y, PPL)     // PPL = a_dword * b_dwords (uses low 32 of each qword)
		f.VPSRLQ("$32", PPL, PPH)
		f.VPANDQ(LSW, PPL, PPL)
		f.VPADDQ(PPL, lo, lo)
		f.VPADDQ(PPH, hi, hi)
	}, true)

	f.Loop(n, func() {
		f.Comment("Load b[i]: 8 dwords -> 8 qwords via zero-extension")
		f.VPMOVZXDQ("0("+addrB+")", Y)

		f.Comment("Multiply each dword of a[i] (8 dwords) by b[i] and accumulate")
		for i := range 7 {
			mac(fmt.Sprintf("%d(%s)", i*4, addrA), accL[i], accH[i])
		}
		// Use a 4-byte broadcast for the final dword (offset 28) to avoid
		// reading past the 32-byte element boundary; see MAC_LAST above.
		macLast(fmt.Sprintf("28(%s)", addrA), accL[7], accH[7])

		f.ADDQ("$32", addrA)
		f.ADDQ("$32", addrB)
	})

	f.LABEL(accumulate)
	f.Comment("Combine 16 partial product accumulators into 544-bit result in Z1:Z0")
	f.Comment("Each accumulator has 8 qwords; we reduce across qwords using VALIGND")

	// Setup masks for VALIGND operations
	f.MOVQ(uint64(0x1555), amd64.AX)
	f.KMOVD(amd64.AX, "K1")
	f.MOVQ(uint64(1), amd64.AX)
	f.KMOVD(amd64.AX, "K2")

	f.Comment("Word 0: lowest 32 bits from accL[0]")
	f.VALIGND_Z("$16", ACC, ACC, "K2", "Z0")
	f.KSHIFTLW("$1", "K2", "K2")

	f.Comment("Propagate carries and add contributions for word 1")
	f.VPSRLQ("$32", ACC, PPL)
	f.VALIGND_Z("$2", ACC, ACC, "K1", ACC)
	f.VPADDQ(PPL, ACC, ACC)
	f.VPANDQ(LSW, accH[0], PPL)
	f.VPADDQ(PPL, ACC, ACC)
	f.VPANDQ(LSW, accL[1], PPL)
	f.VPADDQ(PPL, ACC, ACC)
	f.VALIGNDk("$15", ACC, ACC, "K2", "Z0")
	f.KSHIFTLW("$1", "K2", "K2")

	f.Comment("Words 2-7: combine partial products with carry propagation")
	addPP := f.Define("ADDPP", 5, func(inputs ...any) {
		axH := inputs[0] // accH[x]
		ayL := inputs[1] // accL[y]
		ayH := inputs[2] // accH[y]
		azL := inputs[3] // accL[z]
		idx := inputs[4] // word index

		f.VPSRLQ("$32", ACC, PPL)
		f.VALIGND_Z("$2", ACC, ACC, "K1", ACC)
		f.VPADDQ(PPL, ACC, ACC)
		f.VPSRLQ("$32", axH, axH)
		f.VPADDQ(axH, ACC, ACC)
		f.VPSRLQ("$32", ayL, ayL)
		f.VPADDQ(ayL, ACC, ACC)
		f.VPANDQ(LSW, ayH, PPL)
		f.VPADDQ(PPL, ACC, ACC)
		f.VPANDQ(LSW, azL, PPL)
		f.VPADDQ(PPL, ACC, ACC)
		f.VALIGNDk("$16-"+idx.(amd64.Register), ACC, ACC, "K2", "Z0")
		f.KADDW("K2", "K2", "K2")
	}, true)

	addPP(accH[0], accL[1], accH[1], accL[2], "2")
	addPP(accH[1], accL[2], accH[2], accL[3], "3")
	addPP(accH[2], accL[3], accH[3], accL[4], "4")
	addPP(accH[3], accL[4], accH[4], accL[5], "5")
	addPP(accH[4], accL[5], accH[5], accL[6], "6")
	addPP(accH[5], accL[6], accH[6], accL[7], "7")

	f.Comment("Word 8: final contributions from accH[6], accL[7], accH[7]")
	f.VPSRLQ("$32", ACC, PPL)
	f.VALIGND_Z("$2", ACC, ACC, "K1", ACC)
	f.VPADDQ(PPL, ACC, ACC)
	f.VPSRLQ("$32", accH[6], accH[6])
	f.VPADDQ(accH[6], ACC, ACC)
	f.VPSRLQ("$32", accL[7], accL[7])
	f.VPADDQ(accL[7], ACC, ACC)
	f.VPANDQ(LSW, accH[7], PPL)
	f.VPADDQ(PPL, ACC, ACC)
	f.VALIGNDk("$16-8", ACC, ACC, "K2", "Z0")
	f.KSHIFTLW("$1", "K2", "K2")

	f.Comment("Word 9: remaining from accH[7]")
	f.VPSRLQ("$32", ACC, PPL)
	f.VALIGND_Z("$2", ACC, ACC, "K1", ACC)
	f.VPADDQ(PPL, ACC, ACC)
	f.VPSRLQ("$32", accH[7], accH[7])
	f.VPADDQ(accH[7], ACC, ACC)
	f.VALIGNDk("$16-9", ACC, ACC, "K2", "Z0")
	f.KSHIFTLW("$1", "K2", "K2")

	f.Comment("Words 10-15: propagate remaining carries")
	addPP2 := f.Define("ADDPP2", 1, func(args ...any) {
		idx := args[0]
		f.VPSRLQ("$32", ACC, PPL)
		f.VALIGND_Z("$2", ACC, ACC, "K1", ACC)
		f.VPADDQ(PPL, ACC, ACC)
		f.VALIGNDk("$16-"+idx.(amd64.Register), ACC, ACC, "K2", "Z0")
		f.KSHIFTLW("$1", "K2", "K2")
	}, true)

	for i := 10; i <= 15; i++ {
		addPP2(fmt.Sprintf("%d", i))
	}

	f.Comment("Final carry into Z1")
	f.VPSRLQ("$32", ACC, PPL)
	f.VALIGND_Z("$2", ACC, ACC, "K1", ACC)
	f.VPADDQ(PPL, ACC, ACC)
	f.VMOVDQA64_Z(ACC, "K1", "Z1")

	f.Comment("Montgomery reduction of the 544-bit result")
	f.Comment("Extract low 4 qwords from Z0 into scalar registers")
	T := make([]amd64.Register, 5)
	for i := range 5 {
		T[i] = f.Pop(&registers)
	}
	f.LabelRegisters("T", T...)

	f.VMOVQ("X0", T[1])
	f.VALIGNQ("$1", "Z0", "Z1", "Z0")
	f.VMOVQ("X0", T[2])
	f.VALIGNQ("$1", "Z0", "Z0", "Z0")
	f.VMOVQ("X0", T[3])
	f.VALIGNQ("$1", "Z0", "Z0", "Z0")
	f.VMOVQ("X0", T[4])
	f.VALIGNQ("$1", "Z0", "Z0", "Z0")
	f.XORQ(T[0], T[0])

	f.Comment("4 rounds of Montgomery reduction")
	PH := f.Pop(&registers)
	PL := amd64.AX

	for round := range 4 {
		f.Comment(fmt.Sprintf("Montgomery reduction round %d", round+1))
		src := T[1+round%4]
		f.MOVQ(f.qInv0(), amd64.DX)
		f.MULXQ(src, amd64.DX, PH)

		// Add q * m to T (4-word addition with carry)
		f.MULXQ(f.qAt(0), PL, PH)
		f.ADDQ(PL, T[(1+round)%5])
		f.ADCQ(PH, T[(2+round)%5])
		f.MULXQ(f.qAt(2), PL, PH)
		f.ADCQ(PL, T[(3+round)%5])
		f.ADCQ(PH, T[(4+round)%5])
		f.ADCQ("$0", T[(0+round)%5])
		f.MULXQ(f.qAt(1), PL, PH)
		f.ADDQ(PL, T[(2+round)%5])
		f.ADCQ(PH, T[(3+round)%5])
		f.MULXQ(f.qAt(3), PL, PH)
		f.ADCQ(PL, T[(4+round)%5])
		f.ADCQ(PH, T[(0+round)%5])
		f.ADCQ("$0", T[(1+round)%5])
	}

	f.Comment("Add remaining 5 qwords from Z0")
	for i := range 5 {
		f.VMOVQ("X0", PL)
		if i == 0 {
			f.ADDQ(PL, T[0])
		} else {
			f.ADCQ(PL, T[i])
		}
		if i < 4 {
			f.VALIGNQ("$1", "Z0", "Z0", "Z0")
		}
	}

	f.Comment("Barrett reduction; see Handbook of Applied Cryptography, Algorithm 14.42.")
	f.MOVQ(T[3], amd64.AX)
	f.SHRQw("$32", T[4], amd64.AX)
	f.MOVQ(f.mu(), amd64.DX)
	f.MULQ(amd64.DX)

	f.Comment("Subtract k*q from T")
	f.MULXQ(f.qAt(0), PL, PH)
	f.SUBQ(PL, T[0])
	f.SBBQ(PH, T[1])
	f.MULXQ(f.qAt(2), PL, PH)
	f.SBBQ(PL, T[2])
	f.SBBQ(PH, T[3])
	f.SBBQ("$0", T[4])
	f.MULXQ(f.qAt(1), PL, PH)
	f.SUBQ(PL, T[1])
	f.SBBQ(PH, T[2])
	f.MULXQ(f.qAt(3), PL, PH)
	f.SBBQ(PL, T[3])
	f.SBBQ(PH, T[4])

	f.Comment("Conditional subtraction: up to 2 subtractions needed to get result < q")
	addrRes := f.Pop(&registers)
	f.MOVQ("res+0(FP)", addrRes)
	result := T[:4]
	f.Mov(result, addrRes)

	for range 2 {
		f.SUBQ(f.qAt(0), T[0])
		f.SBBQ(f.qAt(1), T[1])
		f.SBBQ(f.qAt(2), T[2])
		f.SBBQ(f.qAt(3), T[3])
		f.SBBQ("$0", T[4])
		f.JCS(done)
		f.Mov(result, addrRes)
	}

	f.LABEL(done)
	f.RETAVX()
}
