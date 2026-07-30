// Copyright 2020-2025 Consensys Software Inc.
// Licensed under the Apache License, Version 2.0. See the LICENSE file for details.

package amd64

import (
	"fmt"
	"math/big"

	"github.com/consensys/bavard/amd64"
)

// generateMulE2Lazy generates a complex (Karatsuba) 𝔽p2 multiplication with
// lazy (deferred) Montgomery reduction, for towers 𝔽p2 = 𝔽p[u]/(u² - β) with a
// small integer non-residue β:
//
//	t0 = x.A0 * y.A0                     // 2N-word unreduced product
//	t1 = x.A1 * y.A1                     // 2N-word unreduced product
//	t2 = (x.A0 + x.A1) * (y.A0 + y.A1)   // 2N-word unreduced product; operands < 2p (not reduced)
//	z.A1 = REDC(t2 - t0 - t1)            // = x.A0*y.A1 + x.A1*y.A0 ∈ [0, 2p²)
//	z.A0 = REDC(t0 + β·t1 [+ |β|·p²])    // the |β|·p² offset keeps the value non-negative for β < 0
//
// This performs 2 Montgomery reductions instead of the 3 interleaved (fused
// CIOS) reductions of a schoolbook implementation. Soundness requires
// (|β|+1)·p² < p·2^(64N), i.e. (|β|+1)·p < 2^(64N); the carry folds in
// mulNoReduce and redc additionally require the operands' top word to have a
// spare bit, i.e. bitlen(p) ≤ 64N-2.
func (fq2 *Fq2Amd64) generateMulE2Lazy(beta int64, forceCheck bool) {
	if beta == 0 || beta == 1 {
		panic("beta must be a non-residue")
	}
	if fq2.NbWords < 4 || fq2.NbWords > 6 {
		// the sliding accumulator window needs NbWords+1 registers.
		panic("generateMulE2Lazy supports 4 to 6 words only")
	}
	absBeta := beta
	if absBeta < 0 {
		absBeta = -absBeta
	}
	p := limbsToBig(fq2.F.Q)
	soundness := new(big.Int).Mul(p, big.NewInt(absBeta+1))
	if soundness.BitLen() > 64*fq2.NbWords {
		panic(fmt.Sprintf("lazy reduction unsound: (|β|+1)·p has %d bits > %d", soundness.BitLen(), 64*fq2.NbWords))
	}
	if p.BitLen() > 64*fq2.NbWords-2 {
		panic("lazy reduction requires 2 spare bits in the top word")
	}

	const argSize = 24
	minStackSize := 0
	if forceCheck {
		minStackSize = argSize
	}
	// stack: q (N) + t0, t1, t2 (2N each) + aSum, bSum (N each) = 9N words
	stackSize := fq2.StackSize(amd64.NbRegisters-2+9*fq2.NbWords, 2, minStackSize)
	registers := fq2.FnHeader("mulAdxE2", stackSize, argSize, amd64.DX, amd64.AX)
	registers.UnsafePush(amd64.R15)
	defer fq2.AssertCleanStack(stackSize, minStackSize)

	fq2.WriteLn("NO_LOCAL_POINTERS")

	// check ADX instruction support
	lblNoAdx := fq2.NewLabel()
	if forceCheck {
		fq2.CMPB("·supportAdx(SB)", 1)
		fq2.JNE(lblNoAdx)
	}

	var zA0 string
	switch {
	case beta == -1:
		zA0 = "t0 - t1 + p²"
	case beta < -1:
		zA0 = fmt.Sprintf("t0 - %d·t1 + %d·p²", -beta, -beta)
	default:
		zA0 = fmt.Sprintf("t0 + %d·t1", beta)
	}
	fq2.WriteLn(fmt.Sprintf(`
	// 𝔽p2 = 𝔽p[u]/(u² - (%d)), lazy reduction:
	// t0 = x.A0 * y.A0                     (unreduced)
	// t1 = x.A1 * y.A1                     (unreduced)
	// t2 = (x.A0 + x.A1) * (y.A0 + y.A1)   (unreduced)
	// z.A1 = REDC(t2 - t0 - t1)
	// z.A0 = REDC(%s)
	`, beta, zA0))

	ax := amd64.AX

	// MACC and FOLD_CARRIES are the two instruction sequences that dominate
	// the schoolbook products and the Montgomery reductions below; emitting
	// them as #define macros (reused across the 3 products and 2 reductions)
	// keeps the generated assembly compact. MACC is byte-for-byte the macro
	// MulADX emits, so squareAdxE2 shares this single definition.
	mac := fq2.Define("MACC", 3, func(args ...any) {
		fq2.ADCXQ(args[0], args[1])
		fq2.MULXQ(args[2], amd64.AX, args[0])
		fq2.ADOXQ(amd64.AX, args[1])
	}, true)
	fold := fq2.Define("FOLD_CARRIES", 0, func(_ ...any) {
		// fold the two pending ADCX/ADOX carries into A=BP (cannot overflow:
		// the top word of the operands has a spare bit, checked by the caller)
		fq2.MOVQ(0, amd64.AX)
		fq2.ADCXQ(amd64.AX, amd64.BP)
		fq2.ADOXQ(amd64.AX, amd64.BP)
	}, true)

	qStack := fq2.PopN(&registers, true)
	// move q to the stack
	for i := 0; i < fq2.NbWords; i++ {
		fq2.MOVQ(fmt.Sprintf("$const_q%d", i), ax)
		fq2.MOVQ(ax, qStack[i])
	}
	fq2.SetQStack(qStack)
	defer fq2.UnsetQStack()

	t0 := append(fq2.PopN(&registers, true), fq2.PopN(&registers, true)...)
	t1 := append(fq2.PopN(&registers, true), fq2.PopN(&registers, true)...)
	t2 := append(fq2.PopN(&registers, true), fq2.PopN(&registers, true)...)
	aSum := fq2.PopN(&registers, true)
	bSum := fq2.PopN(&registers, true)

	// aSum = x.A0 + x.A1, bSum = y.A0 + y.A1
	// both < 2p: no reduction needed, no carry out of the top word.
	tmp := registers.PopN(fq2.NbWords)
	fq2.Comment("aSum = x.A0 + x.A1 (no reduction)")
	fq2.MOVQ("x+8(FP)", ax)
	fq2.Mov(ax, tmp)
	fq2.Add(ax, tmp, fq2.NbWords)
	fq2.Mov(tmp, aSum)
	fq2.Comment("bSum = y.A0 + y.A1 (no reduction)")
	fq2.MOVQ("y+16(FP)", ax)
	fq2.Mov(ax, tmp)
	fq2.Add(ax, tmp, fq2.NbWords)
	fq2.Mov(tmp, bSum)
	registers.UnsafePush(tmp...)

	xPtr := registers.Pop()
	yPtr := registers.Pop()
	fq2.MOVQ("x+8(FP)", xPtr)
	fq2.MOVQ("y+16(FP)", yPtr)

	fq2.Comment("t0 = x.A0 * y.A0 (unreduced)")
	fq2.mulNoReduce(&registers,
		func(j int) string { return xPtr.At(j) },
		func(i int) string { return yPtr.At(i) },
		t0, mac, fold)

	fq2.Comment("t1 = x.A1 * y.A1 (unreduced)")
	fq2.mulNoReduce(&registers,
		func(j int) string { return xPtr.At(j + fq2.NbWords) },
		func(i int) string { return yPtr.At(i + fq2.NbWords) },
		t1, mac, fold)

	registers.UnsafePush(xPtr, yPtr)

	fq2.Comment("t2 = aSum * bSum (unreduced)")
	fq2.mulNoReduce(&registers,
		func(j int) string { return string(aSum[j]) },
		func(i int) string { return string(bSum[i]) },
		t2, mac, fold)

	fq2.Comment("t2 = t2 - t0 - t1 = x.A0*y.A1 + x.A1*y.A0 (non-negative)")
	fq2.subDW(t2, t0)
	fq2.subDW(t2, t1)

	switch {
	case beta == -1:
		fq2.Comment("t0 = t0 - t1 + p² (p² offset keeps it non-negative; exact modulo 2^(128N))")
		fq2.subDW(t0, t1)
		fq2.addDWConst(t0, pSquaredTimes(fq2.F.Q, 1))
	case beta < -1:
		fq2.Comment(fmt.Sprintf("t0 = t0 - %d·t1 + %d·p² (offset keeps it non-negative; exact modulo 2^(128N))", absBeta, absBeta))
		fq2.mulDWSmallScalar(&registers, t1, uint64(absBeta))
		fq2.subDW(t0, t1)
		fq2.addDWConst(t0, pSquaredTimes(fq2.F.Q, uint64(absBeta)))
	default: // beta > 1
		fq2.Comment(fmt.Sprintf("t0 = t0 + %d·t1 (non-negative, no offset needed)", beta))
		fq2.mulDWSmallScalar(&registers, t1, uint64(beta))
		fq2.addDW(t0, t1)
	}

	fq2.Comment("z.A1 = REDC(t2)")
	res := fq2.redc(&registers, t2, qStack, mac, fold)
	r := registers.Pop()
	fq2.MOVQ("res+0(FP)", r)
	fq2.Mov(res, r, 0, fq2.NbWords)
	registers.UnsafePush(r)
	registers.UnsafePush(res...)

	fq2.Comment("z.A0 = REDC(t0)")
	res = fq2.redc(&registers, t0, qStack, mac, fold)
	r = registers.Pop()
	fq2.MOVQ("res+0(FP)", r)
	fq2.Mov(res, r)
	registers.UnsafePush(r)
	registers.UnsafePush(res...)

	fq2.RET()

	// No adx
	if forceCheck {
		fq2.LABEL(lblNoAdx)
		fq2.MOVQ("res+0(FP)", amd64.AX)
		fq2.MOVQ(amd64.AX, "(SP)")
		fq2.MOVQ("x+8(FP)", amd64.AX)
		fq2.MOVQ(amd64.AX, "8(SP)")
		fq2.MOVQ("y+16(FP)", amd64.AX)
		fq2.MOVQ(amd64.AX, "16(SP)")
		fq2.WriteLn("CALL ·mulGenericE2(SB)")
		fq2.RET()
	}

	fq2.UnsafePush(&registers, aSum...)
	fq2.UnsafePush(&registers, bSum...)
	fq2.UnsafePush(&registers, t0...)
	fq2.UnsafePush(&registers, t1...)
	fq2.UnsafePush(&registers, t2...)
	fq2.UnsafePush(&registers, qStack...)
}

// mulNoReduce generates a full 2N-word (unreduced) schoolbook product dst = x*y.
// xat(j) must return operands usable as MULXQ sources (register or memory);
// yat(i) is loaded into DX one word at a time. dst must be 2N stack slots.
// Uses AX, DX, BP and pops N+1 registers (a sliding accumulator window):
// after row i the low word t[i] is final and is spilled to dst[i], and the
// freed register becomes the (zeroed) top of the window.
// Requirement: the top word of both operands must be < 2^63 - 1 so that the
// final carry fold of each row cannot overflow (holds for operands < 2p when
// bitlen(p) ≤ 64N-2, checked by the caller).
func (fq2 *Fq2Amd64) mulNoReduce(registers *amd64.Registers, xat, yat func(int) string, dst []amd64.Register, mac, fold func(...any)) {
	N := fq2.NbWords
	ax := amd64.AX
	dx := amd64.DX
	A := amd64.BP

	w := registers.PopN(N + 1)

	// row 0: w[0..N] = x * y[0]
	fq2.MOVQ(yat(0), dx)
	fq2.MULXQ(xat(0), w[0], w[1])
	for j := 1; j < N; j++ {
		fq2.MULXQ(xat(j), ax, w[j+1])
		if j == 1 {
			fq2.ADDQ(ax, w[j])
		} else {
			fq2.ADCQ(ax, w[j])
		}
	}
	fq2.ADCQ(0, w[N])
	fq2.MOVQ(w[0], dst[0])
	head := w[0]
	w = append(append([]amd64.Register{}, w[1:]...), head)

	// rows 1..N-1: add x*y[i] at offset i (dual ADCX/ADOX carry chains)
	for i := 1; i < N; i++ {
		fq2.MOVQ(yat(i), dx)
		// zero the new top of the window; also clears CF and OF
		fq2.XORQ(w[N], w[N])
		fq2.MULXQ(xat(0), ax, A)
		fq2.ADOXQ(ax, w[0])
		for j := 1; j < N; j++ {
			mac(A, w[j], xat(j))
		}
		// fold the two pending carries into A (cannot overflow, top word of x has a spare bit)
		fold()
		fq2.MOVQ(A, w[N])

		fq2.MOVQ(w[0], dst[i])
		head = w[0]
		w = append(append([]amd64.Register{}, w[1:]...), head)
	}

	// upper half
	for i := 0; i < N; i++ {
		fq2.MOVQ(w[i], dst[N+i])
	}

	registers.UnsafePush(w...)
}

// redc generates a standalone Montgomery reduction (SOS): given t, a 2N-word
// value on the stack with t < p·2^(64N), returns N registers holding
// t·2^(-64N) mod p, fully reduced.
func (fq2 *Fq2Amd64) redc(registers *amd64.Registers, t, qStack []amd64.Register, mac, fold func(...any)) []amd64.Register {
	N := fq2.NbWords
	ax := amd64.AX
	dx := amd64.DX
	A := amd64.BP

	w := registers.PopN(N + 1)
	e := registers.Pop()

	// load the low window t[0..N]
	for i := 0; i <= N; i++ {
		fq2.MOVQ(t[i], w[i])
	}
	// e is the ripple carry into the word above the current window top
	fq2.XORQ(e, e)

	for i := 0; i < N; i++ {
		fq2.Comment(fmt.Sprintf("REDC round %d", i))
		// m = w[0] * qInvNeg mod 2^64
		fq2.MOVQ("$const_qInvNeg", dx)
		fq2.IMULQ(w[0], dx)
		// clear CF and OF
		fq2.XORQ(ax, ax)
		// w[0..N-1] += m*q (w[0] zeroes out)
		fq2.MULXQ(qStack[0], ax, A)
		fq2.ADOXQ(ax, w[0])
		for j := 1; j < N; j++ {
			mac(A, w[j], qStack[j])
		}
		// fold the two pending carries into A (cannot overflow, top word of q has spare bits)
		fold()
		// w[N] += A + e; carries ripple into e
		fq2.ADDQ(e, w[N])
		fq2.MOVQ(0, e)
		fq2.ADCQ(0, e)
		fq2.ADDQ(A, w[N])
		fq2.ADCQ(0, e)
		// rotate the window: discard w[0] (now 0), load the next word of t
		head := w[0]
		if i+N+1 < 2*N {
			fq2.MOVQ(t[i+N+1], head)
		}
		w = append(append([]amd64.Register{}, w[1:]...), head)
	}

	// w[0..N-1] < 2p (final e is provably 0); single conditional subtraction
	scratch := []amd64.Register{w[N], e, registers.Pop(), registers.Pop(), registers.Pop(), registers.Pop()}
	fq2.ReduceElement(w[:N], scratch, true)

	registers.UnsafePush(scratch[2:]...)
	registers.UnsafePush(w[N], e)

	return w[:N]
}

// subDW generates a -= b over 2N words (values on the stack), borrow discarded.
func (fq2 *Fq2Amd64) subDW(a, b []amd64.Register) {
	ax := amd64.AX
	for i := 0; i < 2*fq2.NbWords; i++ {
		fq2.MOVQ(a[i], ax)
		if i == 0 {
			fq2.SUBQ(b[i], ax)
		} else {
			fq2.SBBQ(b[i], ax)
		}
		fq2.MOVQ(ax, a[i])
	}
}

// addDW generates a += b over 2N words (values on the stack), carry discarded.
func (fq2 *Fq2Amd64) addDW(a, b []amd64.Register) {
	ax := amd64.AX
	for i := 0; i < 2*fq2.NbWords; i++ {
		fq2.MOVQ(a[i], ax)
		if i == 0 {
			fq2.ADDQ(b[i], ax)
		} else {
			fq2.ADCQ(b[i], ax)
		}
		fq2.MOVQ(ax, a[i])
	}
}

// addDWConst generates a += c over 2N words (a on the stack), carry discarded.
func (fq2 *Fq2Amd64) addDWConst(a []amd64.Register, c []uint64) {
	ax := amd64.AX
	for i := 0; i < 2*fq2.NbWords; i++ {
		fq2.MOVQ(c[i], ax)
		if i == 0 {
			fq2.ADDQ(ax, a[i])
		} else {
			fq2.ADCQ(ax, a[i])
		}
	}
}

// mulDWSmallScalar generates t = scalar * t over 2N words (t on the stack),
// for a small scalar. The caller must guarantee the product fits in 2N words
// (scalar·t < 2^(128N)). Uses AX, DX and pops 2 scratch registers.
func (fq2 *Fq2Amd64) mulDWSmallScalar(registers *amd64.Registers, t []amd64.Register, scalar uint64) {
	ax := amd64.AX
	dx := amd64.DX
	hi0 := registers.Pop()
	hi1 := registers.Pop()

	fq2.MOVQ(scalar, dx)
	fq2.MULXQ(t[0], ax, hi0)
	fq2.MOVQ(ax, t[0])
	for i := 1; i < 2*fq2.NbWords; i++ {
		fq2.MULXQ(t[i], ax, hi1)
		if i == 1 {
			fq2.ADDQ(hi0, ax)
		} else {
			fq2.ADCQ(hi0, ax)
		}
		fq2.MOVQ(ax, t[i])
		hi0, hi1 = hi1, hi0
	}
	// the final high word is zero by the caller's bound.

	registers.UnsafePush(hi0, hi1)
}

// pSquaredTimes returns the 2N little-endian words of k·p².
func pSquaredTimes(q []uint64, k uint64) []uint64 {
	p := limbsToBig(q)
	p.Mul(p, p)
	p.Mul(p, new(big.Int).SetUint64(k))
	if p.BitLen() > 64*2*len(q) {
		panic("k·p² does not fit in 2N words")
	}
	out := make([]uint64, 2*len(q))
	mask := new(big.Int).SetUint64(^uint64(0))
	tmp := new(big.Int).Set(p)
	for i := range out {
		out[i] = new(big.Int).And(tmp, mask).Uint64()
		tmp.Rsh(tmp, 64)
	}
	return out
}

// limbsToBig converts little-endian 64-bit limbs to a big.Int.
func limbsToBig(q []uint64) *big.Int {
	r := new(big.Int)
	for i := len(q) - 1; i >= 0; i-- {
		r.Lsh(r, 64)
		r.Or(r, new(big.Int).SetUint64(q[i]))
	}
	return r
}

func concat(a []amd64.Register, b ...amd64.Register) []amd64.Register {
	return append(a, b...)
}
