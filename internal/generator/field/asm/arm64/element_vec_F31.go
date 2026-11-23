package arm64

import (
	"fmt"

	"github.com/consensys/bavard/arm64"
)

// modAddF31 emits dst = a + b mod q on 4-lane registers, for a, b in [0, q). q holds the broadcast modulus and t is
// scratch. dst and t may each be a or b, but must differ from each other and from q.
func (f *FFArm64) modAddF31(dst, a, b, q, t arm64.VectorRegister) {
	f.VADD(a.S4(), b.S4(), dst.S4(), fmt.Sprintf("%s = %s + %s", dst, a, b))
	f.VSUB(q.S4(), dst.S4(), t.S4(), fmt.Sprintf("%s = %s - %s", t, dst, q))
	f.VUMIN(t.S4(), dst.S4(), dst.S4(), fmt.Sprintf("%s = min(%s, %s)", dst, t, dst))
}

// modSubF31 emits dst = a - b mod q on 4-lane registers, for a, b in [0, q). q holds the broadcast modulus and t is
// scratch. dst and t may each be a or b, but must differ from each other and from q.
func (f *FFArm64) modSubF31(dst, a, b, q, t arm64.VectorRegister) {
	f.VSUB(b.S4(), a.S4(), dst.S4(), fmt.Sprintf("%s = %s - %s", dst, a, b))
	f.VADD(dst.S4(), q.S4(), t.S4(), fmt.Sprintf("%s = %s + %s", t, dst, q))
	f.VUMIN(t.S4(), dst.S4(), dst.S4(), fmt.Sprintf("%s = min(%s, %s)", dst, t, dst))
}

// montMulF31 emits d = a * b * R^-1 mod p on 4-lane registers, with R = 2^32, for a, b in [0, p). p holds the
// broadcast modulus and mu the broadcast p^-1 mod 2^32; cHi, q, qpHi and underflow are scratch. d may be a or b, which
// are not read after d is written, and must differ from the other registers.
func (f *FFArm64) montMulF31(d, a, b, p, mu, cHi, q, qpHi, underflow arm64.VectorRegister) {
	// Step 1: c_hi = (2 * a * b) >> 32 using SQDMULH
	// SQDMULH computes (2*a*b) >> 32 with signed saturation
	f.VSQDMULH(a, b, cHi, fmt.Sprintf("%s = (2*%s*%s) >> 32", cHi, a, b))

	// Step 2: q = (a * b * mu) mod 2^32
	// First compute temp = a * b (low 32 bits)
	// Then q = temp * mu (low 32 bits)
	f.VMUL_S4(a, b, q, fmt.Sprintf("%s = %s * %s (low 32 bits)", q, a, b))
	f.VMUL_S4(q, mu, q, fmt.Sprintf("%s = %s * %s (low 32 bits)", q, q, mu))

	f.montReduceF31(d, cHi, q, p, qpHi, underflow)
}

// montReduceF31 emits the second half of a Montgomery product on 4-lane registers: given cHi = (2 * a * b) >> 32 and
// q = (a * b * mu) mod 2^32, it sets d = a * b * R^-1 mod p. p holds the broadcast modulus; qpHi and underflow are
// scratch. d must differ from the other registers.
func (f *FFArm64) montReduceF31(d, cHi, q, p, qpHi, underflow arm64.VectorRegister) {
	// Step 3: qp_hi = (2 * q * P) >> 32 using SQDMULH
	f.VSQDMULH(q, p, qpHi, fmt.Sprintf("%s = (2*%s*%s) >> 32", qpHi, q, p))

	// Step 4: d = (c_hi - qp_hi) / 2 using SHSUB
	// This computes the halving subtraction which accounts for the 2x factor in SQDMULH
	f.VSHSUB(cHi, qpHi, d, fmt.Sprintf("%s = (%s - %s) / 2", d, cHi, qpHi))

	// Step 5: Canonicalize - if d < 0 (negative), add P
	// underflow mask: all 1s if qp_hi > c_hi (which means d is negative), else 0
	f.VCMLT(cHi, qpHi, underflow, fmt.Sprintf("%s = (%s < %s) ? all 1s : 0", underflow, cHi, qpHi))

	// d = d - underflow * P = d + P if underflow (since underflow is -1 when true)
	f.VMLS(underflow, p, d, fmt.Sprintf("%s = %s - %s * %s (adds p when %s < 0)", d, d, underflow, p, d))
}

func (f *FFArm64) generateAddVecF31() {
	f.Comment("addVec(res, a, b *Element, n uint64)")
	f.Comment("n is the number of blocks of 4 uint32 to process")
	registers := f.FnHeader("addVec", 0, 32)
	defer f.AssertCleanStack(0, 0)

	// registers
	resPtr := registers.Pop()
	aPtr := registers.Pop()
	bPtr := registers.Pop()
	n := registers.Pop()

	// load arguments
	f.LDP("res+0(FP)", resPtr, aPtr)
	f.LDP("b+16(FP)", bPtr, n)

	a0 := registers.PopV()
	b0 := registers.PopV()
	q := registers.PopV()

	f.VMOVS("$const_q", q)
	f.VDUP(q.SAt(0), q.S4(), "broadcast q into "+string(q))

	const offset = 4 * 4 // 4 uint32 = 16 bytes per vector

	f.Loop(n, func() {
		f.VLD1_P(offset, aPtr, a0.S4())
		f.VLD1_P(offset, bPtr, b0.S4())

		f.modAddF31(b0, a0, b0, q, a0)
		f.VST1_P(b0.S4(), resPtr, offset, fmt.Sprintf("res = %s", b0))
	})

	registers.Push(resPtr, aPtr, bPtr, n)
	registers.PushV(a0, b0, q)

	f.RET()

}

func (f *FFArm64) generateMulVecF31() {
	f.Comment("mulVec(res, a, b *Element, n uint64)")
	f.Comment("n is the number of blocks of 4 uint32 to process")
	f.Comment("")
	f.Comment("Algorithm from plonky3 using SQDMULH for efficient Montgomery reduction:")
	f.Comment("For inputs a, b in [0, P), compute a*b*R^-1 mod P where R = 2^32")
	f.Comment("  1. c_hi = (2 * a * b) >> 32  using SQDMULH")
	f.Comment("  2. q = (a * b * mu) mod 2^32")
	f.Comment("  3. qp_hi = (2 * q * P) >> 32 using SQDMULH")
	f.Comment("  4. d = (c_hi - qp_hi) / 2 using SHSUB")
	f.Comment("  5. if d < 0: d += P")
	registers := f.FnHeader("mulVec", 0, 32)
	defer f.AssertCleanStack(0, 0)

	// registers
	resPtr := registers.Pop()
	aPtr := registers.Pop()
	bPtr := registers.Pop()
	n := registers.Pop()

	// load arguments
	f.LDP("res+0(FP)", resPtr, aPtr)
	f.LDP("b+16(FP)", bPtr, n)

	// Explicit registers for Montgomery multiplication using SQDMULH
	a := arm64.V0
	b := arm64.V1
	cHi := arm64.V2
	q := arm64.V3
	qpHi := arm64.V4
	d := arm64.V5
	p := arm64.V6
	mu := arm64.V7
	underflow := arm64.V8

	// Load constants
	// P (as signed for SQDMULH - values < 2^31 are safe)
	f.VMOVS("$const_q", p)
	f.VDUP(p.SAt(0), p.S4(), "broadcast P")

	// MU
	tmp := registers.Pop()
	f.MOVD("$const_mu", tmp)
	f.VDUP(tmp, mu.S4(), "broadcast MU")
	registers.Push(tmp)

	f.Loop(n, func() {
		const offset = 4 * 4 // we process 4 uint32 at a time

		f.VLD1_P(offset, aPtr, a.S4())
		f.VLD1_P(offset, bPtr, b.S4())

		f.montMulF31(d, a, b, p, mu, cHi, q, qpHi, underflow)

		f.VST1_P(d.S4(), resPtr, offset, fmt.Sprintf("res = %s", d))
	})

	registers.Push(resPtr, aPtr, bPtr, n)

	f.RET()

}

func (f *FFArm64) generateSubVecF31() {
	f.Comment("subVec(res, a, b *Element, n uint64)")
	f.Comment("n is the number of blocks of 4 uint32 to process")
	registers := f.FnHeader("subVec", 0, 32)
	defer f.AssertCleanStack(0, 0)

	// registers
	resPtr := registers.Pop()
	aPtr := registers.Pop()
	bPtr := registers.Pop()
	n := registers.Pop()

	// load arguments
	f.LDP("res+0(FP)", resPtr, aPtr)
	f.LDP("b+16(FP)", bPtr, n)

	a0 := registers.PopV()
	b0 := registers.PopV()
	q := registers.PopV()

	f.VMOVS("$const_q", q)
	f.VDUP(q.SAt(0), q.S4(), "broadcast q into "+string(q))

	const offset = 4 * 4 // we process 4 uint32 at a time

	f.Loop(n, func() {
		f.VLD1_P(offset, aPtr, a0.S4())
		f.VLD1_P(offset, bPtr, b0.S4())

		f.modSubF31(b0, a0, b0, q, a0)
		f.VST1_P(b0.S4(), resPtr, offset, fmt.Sprintf("res = %s", b0))
	})

	registers.Push(resPtr, aPtr, bPtr, n)
	registers.PushV(a0, b0, q)

	f.RET()

}

func (f *FFArm64) generateSumVecF31() {
	f.Comment("sumVec(t *uint64, a *[]uint32, n uint64) res = sum(a[0...n])")
	f.Comment("n is the number of blocks of 16 uint32 to process")
	f.Comment("")
	f.Comment("Uses UADALP (Unsigned Add and Accumulate Long Pairwise) to efficiently")
	f.Comment("sum 32-bit elements into 64-bit accumulators in a single instruction.")
	registers := f.FnHeader("sumVec", 0, 3*8)
	defer f.AssertCleanStack(0, 0)

	// registers
	aPtr := registers.Pop()
	tPtr := registers.Pop()
	n := registers.Pop()

	// We load 4 vectors of 4 x uint32 = 16 elements per iteration
	a0 := registers.PopV()
	a1 := registers.PopV()
	a2 := registers.PopV()
	a3 := registers.PopV()
	acc0V := registers.PopV()
	acc1V := registers.PopV()

	f.Comment("zeroing accumulators")
	f.VMOVQ_cst(0, 0, acc0V)
	f.VMOVQ_cst(0, 0, acc1V)

	acc0 := acc0V.D2()
	acc1 := acc1V.D2()

	// load arguments
	f.LDP("t+0(FP)", tPtr, aPtr)
	f.MOVD("n+16(FP)", n)

	const offset = 16 * 4 // 16 uint32 = 64 bytes per iteration

	f.Loop(n, func() {
		// Load 16 elements (4 vectors of 4 uint32)
		f.VLD1_P_Multi(offset, aPtr, a0, a1, a2, a3)

		// Add vectors pairwise: a0 += a1, a2 += a3
		f.VADD(a0.S4(), a1.S4(), a0.S4(), fmt.Sprintf("%s += %s", a0, a1))
		f.VADD(a2.S4(), a3.S4(), a2.S4(), fmt.Sprintf("%s += %s", a2, a3))

		// UADALP: pairwise add adjacent 32-bit elements, widen to 64-bit, and accumulate
		// This replaces VUSHLL + VADD with a single instruction
		f.VUADALP(a0, acc0V, fmt.Sprintf("%s += pairwise_widen(%s)", acc0V, a0))
		f.VUADALP(a2, acc1V, fmt.Sprintf("%s += pairwise_widen(%s)", acc1V, a2))
	})

	// Final reduction: combine the two accumulators
	f.VADD(acc0, acc1, acc0, fmt.Sprintf("%s += %s", acc0V, acc1V))

	f.VST1_P(acc0, tPtr, 0, "store accumulator")

	registers.Push(aPtr, tPtr, n)
	registers.PushV(a0, a1, a2, a3, acc0V, acc1V)

	f.RET()

}

func (f *FFArm64) generateScalarMulVecF31() {
	f.Comment("scalarMulVec(res, a, b *Element, n uint64) res[0...n] = a[0...n] * b")
	f.Comment("n is the number of blocks of 4 uint32 to process")
	f.Comment("")
	f.Comment("Algorithm from plonky3 using SQDMULH for efficient Montgomery reduction")
	registers := f.FnHeader("scalarMulVec", 0, 32)
	defer f.AssertCleanStack(0, 0)

	// registers
	resPtr := registers.Pop()
	aPtr := registers.Pop()
	bPtr := registers.Pop()
	n := registers.Pop()

	// load arguments
	f.LDP("res+0(FP)", resPtr, aPtr)
	f.LDP("b+16(FP)", bPtr, n)

	// Explicit registers for Montgomery multiplication using SQDMULH
	a := arm64.V0
	b := arm64.V1
	cHi := arm64.V2
	q := arm64.V3
	qpHi := arm64.V4
	d := arm64.V5
	p := arm64.V6
	mu := arm64.V7
	underflow := arm64.V8
	muB := arm64.V9 // precomputed mu * b

	// Load constants
	// P (as signed for SQDMULH - values < 2^31 are safe)
	f.VMOVS("$const_q", p)
	f.VDUP(p.SAt(0), p.S4(), "broadcast P")

	// MU
	tmp := registers.Pop()
	f.MOVD("$const_mu", tmp)
	f.VDUP(tmp, mu.S4(), "broadcast MU")

	// Load scalar b and broadcast
	f.MOVWU(bPtr.At2(0), tmp)
	f.VDUP(tmp, b.S4(), "broadcast scalar b")

	// Precompute mu * b for reuse in the loop
	f.VMUL_S4(mu, b, muB, fmt.Sprintf("%s = %s * %s (precomputed)", muB, mu, b))
	registers.Push(tmp)

	f.Loop(n, func() {
		const offset = 4 * 4 // we process 4 uint32 at a time

		f.VLD1_P(offset, aPtr, a.S4())

		// Step 1: c_hi = (2 * a * b) >> 32 using SQDMULH
		f.VSQDMULH(a, b, cHi, fmt.Sprintf("%s = (2*%s*%s) >> 32", cHi, a, b))

		// Step 2: q = (a * b * mu) mod 2^32 = a * (mu * b) mod 2^32
		// Using precomputed muB = mu * b
		f.VMUL_S4(a, muB, q, fmt.Sprintf("%s = %s * %s (low 32 bits)", q, a, muB))

		f.montReduceF31(d, cHi, q, p, qpHi, underflow)

		f.VST1_P(d.S4(), resPtr, offset, fmt.Sprintf("res = %s", d))
	})

	registers.Push(resPtr, aPtr, bPtr, n)

	f.RET()
}

func (f *FFArm64) generateInnerProdVecF31() {
	f.Comment("innerProdVec(t *uint64, a, b *[]uint32, n uint64) res = sum(a[0...n] * b[0...n])")
	f.Comment("n is the number of blocks of 4 uint32 to process")
	f.Comment("We do montgomery multiplication and accumulate the reduced results.")
	f.Comment("")
	f.Comment("Algorithm from plonky3 using SQDMULH for efficient Montgomery reduction")
	registers := f.FnHeader("innerProdVec", 0, 32)
	defer f.AssertCleanStack(0, 0)

	// registers
	tPtr := registers.Pop()
	aPtr := registers.Pop()
	bPtr := registers.Pop()
	n := registers.Pop()

	// load arguments
	f.LDP("t+0(FP)", tPtr, aPtr)
	f.LDP("b+16(FP)", bPtr, n)

	// Explicit registers for Montgomery multiplication using SQDMULH
	a := arm64.V0
	b := arm64.V1
	cHi := arm64.V2
	q := arm64.V3
	qpHi := arm64.V4
	d := arm64.V5
	p := arm64.V6
	mu := arm64.V7
	underflow := arm64.V8

	// Accumulators (64-bit)
	acc0 := arm64.V9.D2()
	acc1 := arm64.V10.D2()

	// Temp registers for widening
	tmp0 := arm64.V11
	tmp1 := arm64.V12

	// Load constants
	// P (as signed for SQDMULH - values < 2^31 are safe)
	f.VMOVS("$const_q", p)
	f.VDUP(p.SAt(0), p.S4(), "broadcast P")

	// MU
	tmp := registers.Pop()
	f.MOVD("$const_mu", tmp)
	f.VDUP(tmp, mu.S4(), "broadcast MU")
	registers.Push(tmp)

	// Zero accumulators
	f.VMOVQ_cst(0, 0, arm64.V9)
	f.VMOVQ_cst(0, 0, arm64.V10)

	f.Loop(n, func() {
		const offset = 4 * 4 // we process 4 uint32 at a time

		f.VLD1_P(offset, aPtr, a.S4())
		f.VLD1_P(offset, bPtr, b.S4())

		f.montMulF31(d, a, b, p, mu, cHi, q, qpHi, underflow)

		// Extend to 64 bits and accumulate
		f.VUSHLL(0, d.S2(), tmp0.D2(), fmt.Sprintf("%s = extend(%s[0,1])", tmp0, d))
		f.VUSHLL2(0, d.S4(), tmp1.D2(), fmt.Sprintf("%s = extend(%s[2,3])", tmp1, d))

		// Accumulate
		f.VADD(tmp0.D2(), acc0, acc0, fmt.Sprintf("%s += %s", baseReg(acc0), tmp0))
		f.VADD(tmp1.D2(), acc1, acc1, fmt.Sprintf("%s += %s", baseReg(acc1), tmp1))
	})

	// Combine accumulators
	f.VADD(acc0, acc1, acc0, fmt.Sprintf("%s += %s", baseReg(acc0), baseReg(acc1)))

	// Store result
	f.VST1_P(acc0, tPtr, 16, "store accumulator")

	registers.Push(tPtr, aPtr, bPtr, n)

	f.RET()
}
