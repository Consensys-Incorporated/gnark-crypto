// Copyright 2020-2026 Consensys Incorporated
// Licensed under the Apache License, Version 2.0. See the LICENSE file for details.

package arm64

import (
	"fmt"
	"strings"

	"github.com/consensys/bavard/arm64"
)

// This file generates the fused NEON kernels behind Vector.MulAcc,
// Vector.ScalarMulAcc, Vector.ScalarMulAdd, Vector.FoldPairs and the two row
// steps of the interleaved batch inversion, for 31-bit fields.
//
// Every kernel takes n, the number of blocks of 4 uint32 to process. The main
// loop handles 4 blocks (16 elements) per iteration, emitting each step of the
// computation for the 4 blocks in a row so that independent instructions are
// in flight; a second loop handles the remaining 0..3 blocks one at a time.
//
// Field elements are kept canonical ([0, q)) between steps: the Montgomery
// multiplication (SQDMULH, see generateMulVecF31) needs its inputs < 2^31 and
// returns a canonical result, as do the modular add / sub / halve below.

// vreg returns the vector register Vi.
func vreg(i int) arm64.VectorRegister {
	if i < 0 || i > 31 {
		panic("invalid vector register index")
	}
	return arm64.VectorRegister(fmt.Sprintf("V%d", i))
}

// vgroup returns k consecutive vector registers starting at Vbase, so that a
// group can be loaded / stored with a single multi-register VLD1 / VST1.
func vgroup(base, k int) []arm64.VectorRegister {
	r := make([]arm64.VectorRegister, k)
	for i := range r {
		r[i] = vreg(base + i)
	}
	return r
}

func s4list(regs []arm64.VectorRegister) string {
	parts := make([]string, len(regs))
	for i, r := range regs {
		parts[i] = string(r.S4())
	}
	return strings.Join(parts, ", ")
}

// ldP loads len(regs) vectors of 4 uint32 from ptr and post-increments ptr.
func (f *FFArm64) ldP(ptr arm64.Register, regs []arm64.VectorRegister) {
	if len(regs) == 1 {
		f.VLD1_P(16, ptr, regs[0].S4())
		return
	}
	f.VLD1_P_Multi(16*len(regs), ptr, regs...)
}

// ld loads len(regs) vectors of 4 uint32 from ptr, without moving it.
func (f *FFArm64) ld(ptr arm64.Register, regs []arm64.VectorRegister) {
	f.WriteLn(fmt.Sprintf("    VLD1 (%s), [%s]", ptr, s4list(regs)))
}

// stP stores len(regs) vectors of 4 uint32 to ptr and post-increments ptr.
func (f *FFArm64) stP(regs []arm64.VectorRegister, ptr arm64.Register) {
	if len(regs) == 1 {
		f.VST1_P(regs[0].S4(), ptr, 16)
		return
	}
	f.VST1_P_Multi(16*len(regs), ptr, regs...)
}

// loopBlocks emits the two loops over n blocks of 4 uint32: 4 blocks per
// iteration, then 1. body(k) must process (and advance the pointers by) k
// blocks.
func (f *FFArm64) loopBlocks(n arm64.Register, body func(k int)) {
	lbl4 := f.NewLabel("loop4x")
	lbl1 := f.NewLabel("loop1x")
	done := f.NewLabel("done")

	f.LABEL(lbl4)
	f.CMP("$4", n)
	f.BLT(lbl1)
	f.SUB(4, n, n)
	body(4)
	f.JMP(lbl4)

	f.LABEL(lbl1)
	f.CBZ(n, done)
	f.SUB(1, n, n)
	body(1)
	f.JMP(lbl1)

	f.LABEL(done)
}

// f31Consts holds the broadcast constants of the kernels.
type f31Consts struct {
	p, mu arm64.VectorRegister
}

// loadF31Consts broadcasts q and mu into p and mu.
func (f *FFArm64) loadF31Consts(tmp arm64.Register, p, mu arm64.VectorRegister) f31Consts {
	f.VMOVS("$const_q", p)
	f.VDUP(p.SAt(0), p.S4(), "broadcast P")
	f.MOVD("$const_mu", tmp)
	f.VDUP(tmp, mu.S4(), "broadcast MU")
	return f31Consts{p: p, mu: mu}
}

// broadcastScalar loads the uint32 at ptr into every lane of dst.
func (f *FFArm64) broadcastScalar(ptr, tmp arm64.Register, dst arm64.VectorRegister, comment string) {
	f.MOVWU(ptr.At2(0), tmp)
	f.VDUP(tmp, dst.S4(), comment)
}

// montMul emits d[i] = a[i] * b[i] * 2^-32 mod q for every block i, with
// a[i], b[i] canonical; the result is canonical. It clobbers t0[i] and t1[i].
// d[i] may be a[i] or b[i].
//
// If muB is not nil, b is a broadcast scalar (b[i] all the same register) and
// muB holds mu * b, saving one multiplication per block.
func (f *FFArm64) montMul(c f31Consts, a, b, d, t0, t1 []arm64.VectorRegister, muB *arm64.VectorRegister) {
	k := len(a)
	for i := range k {
		f.VSQDMULH(a[i], b[i], t0[i], "c_hi = (2*a*b) >> 32")
	}
	if muB != nil {
		for i := range k {
			f.VMUL_S4(a[i], *muB, t1[i], "m = a * (mu*b) (low 32 bits)")
		}
	} else {
		for i := range k {
			f.VMUL_S4(a[i], b[i], t1[i], "m = a * b (low 32 bits)")
		}
		for i := range k {
			f.VMUL_S4(t1[i], c.mu, t1[i], "m = m * mu (low 32 bits)")
		}
	}
	for i := range k {
		f.VSQDMULH(t1[i], c.p, t1[i], "qp_hi = (2*m*P) >> 32")
	}
	for i := range k {
		f.VSHSUB(t0[i], t1[i], d[i], "d = (c_hi - qp_hi) / 2")
	}
	for i := range k {
		f.VCMLT(t0[i], t1[i], t0[i], "underflow = c_hi < qp_hi")
	}
	for i := range k {
		f.VMLS(t0[i], c.p, d[i], "d += P if underflow")
	}
}

// modAdd emits d[i] = a[i] + b[i] mod q; it clobbers t[i]. d[i] may be a[i]
// or b[i], but not t[i].
func (f *FFArm64) modAdd(c f31Consts, a, b, d, t []arm64.VectorRegister) {
	k := len(a)
	for i := range k {
		f.VADD(a[i].S4(), b[i].S4(), t[i].S4(), "s = a + b")
	}
	for i := range k {
		f.VSUB(c.p.S4(), t[i].S4(), d[i].S4(), "d = s - P")
	}
	for i := range k {
		f.VUMIN(t[i].S4(), d[i].S4(), d[i].S4(), "d = min(s, s - P)")
	}
}

// modSub emits d[i] = a[i] - b[i] mod q; it clobbers t[i]. d[i] may be a[i]
// or b[i], but not t[i].
func (f *FFArm64) modSub(c f31Consts, a, b, d, t []arm64.VectorRegister) {
	k := len(a)
	for i := range k {
		f.VSUB(b[i].S4(), a[i].S4(), t[i].S4(), "s = a - b")
	}
	for i := range k {
		f.VADD(t[i].S4(), c.p.S4(), d[i].S4(), "d = s + P")
	}
	for i := range k {
		f.VUMIN(t[i].S4(), d[i].S4(), d[i].S4(), "d = min(s, s + P)")
	}
}

func (f *FFArm64) generateMulAccVecF31() {
	f.Comment("mulAccVec(res, a, b *Element, n uint64) res[i] += a[i] * b[i]")
	f.Comment("n is the number of blocks of 4 uint32 to process")
	registers := f.FnHeader("mulAccVec", 0, 32)
	defer f.AssertCleanStack(0, 0)

	resPtr := registers.Pop()
	aPtr := registers.Pop()
	bPtr := registers.Pop()
	n := registers.Pop()
	outPtr := registers.Pop()
	tmp := registers.Pop()

	f.LDP("res+0(FP)", resPtr, aPtr)
	f.LDP("b+16(FP)", bPtr, n)
	f.MOVD(resPtr, outPtr)

	c := f.loadF31Consts(tmp, vreg(30), vreg(31))

	f.loopBlocks(n, func(k int) {
		r, a, b := vgroup(0, k), vgroup(4, k), vgroup(8, k)
		t0, t1 := vgroup(12, k), vgroup(16, k)
		f.ldP(aPtr, a)
		f.ldP(bPtr, b)
		f.ldP(resPtr, r)
		f.montMul(c, a, b, a, t0, t1, nil)
		f.modAdd(c, r, a, t0, t1)
		f.stP(t0, outPtr)
	})

	registers.Push(resPtr, aPtr, bPtr, n, outPtr, tmp)
	f.RET()
}

func (f *FFArm64) generateScalarMulAccVecF31() {
	f.Comment("scalarMulAccVec(res, a, c *Element, n uint64) res[i] += a[i] * c")
	f.Comment("n is the number of blocks of 4 uint32 to process")
	registers := f.FnHeader("scalarMulAccVec", 0, 32)
	defer f.AssertCleanStack(0, 0)

	resPtr := registers.Pop()
	aPtr := registers.Pop()
	cPtr := registers.Pop()
	n := registers.Pop()
	outPtr := registers.Pop()
	tmp := registers.Pop()

	f.LDP("res+0(FP)", resPtr, aPtr)
	f.LDP("c+16(FP)", cPtr, n)
	f.MOVD(resPtr, outPtr)

	c := f.loadF31Consts(tmp, vreg(30), vreg(31))
	cV, muC := vreg(28), vreg(29)
	f.broadcastScalar(cPtr, tmp, cV, "broadcast c")
	f.VMUL_S4(c.mu, cV, muC, "muC = mu * c")

	f.loopBlocks(n, func(k int) {
		r, a := vgroup(0, k), vgroup(4, k)
		t0, t1 := vgroup(12, k), vgroup(16, k)
		cs := make([]arm64.VectorRegister, k)
		for i := range cs {
			cs[i] = cV
		}
		f.ldP(aPtr, a)
		f.ldP(resPtr, r)
		f.montMul(c, a, cs, a, t0, t1, &muC)
		f.modAdd(c, r, a, t0, t1)
		f.stP(t0, outPtr)
	})

	registers.Push(resPtr, aPtr, cPtr, n, outPtr, tmp)
	f.RET()
}

func (f *FFArm64) generateScalarMulAddVecF31() {
	f.Comment("scalarMulAddVec(res, b, c *Element, n uint64) res[i] = res[i] * c + b[i]")
	f.Comment("n is the number of blocks of 4 uint32 to process")
	registers := f.FnHeader("scalarMulAddVec", 0, 32)
	defer f.AssertCleanStack(0, 0)

	resPtr := registers.Pop()
	bPtr := registers.Pop()
	cPtr := registers.Pop()
	n := registers.Pop()
	outPtr := registers.Pop()
	tmp := registers.Pop()

	f.LDP("res+0(FP)", resPtr, bPtr)
	f.LDP("c+16(FP)", cPtr, n)
	f.MOVD(resPtr, outPtr)

	c := f.loadF31Consts(tmp, vreg(30), vreg(31))
	cV, muC := vreg(28), vreg(29)
	f.broadcastScalar(cPtr, tmp, cV, "broadcast c")
	f.VMUL_S4(c.mu, cV, muC, "muC = mu * c")

	f.loopBlocks(n, func(k int) {
		r, b := vgroup(0, k), vgroup(4, k)
		t0, t1 := vgroup(12, k), vgroup(16, k)
		cs := make([]arm64.VectorRegister, k)
		for i := range cs {
			cs[i] = cV
		}
		f.ldP(resPtr, r)
		f.ldP(bPtr, b)
		f.montMul(c, r, cs, r, t0, t1, &muC)
		f.modAdd(c, r, b, t0, t1)
		f.stP(t0, outPtr)
	})

	registers.Push(resPtr, bPtr, cPtr, n, outPtr, tmp)
	f.RET()
}

func (f *FFArm64) generateFoldPairsVecF31() {
	f.Comment("foldPairsVec(res, layer, twiddles, alpha *Element, n uint64)")
	f.Comment("res[j] = (layer[2j] + layer[2j+1]) / 2 + alpha * twiddles[j] * (layer[2j] - layer[2j+1])")
	f.Comment("n is the number of blocks of 4 res elements to process")
	registers := f.FnHeader("foldPairsVec", 0, 40)
	defer f.AssertCleanStack(0, 0)

	resPtr := registers.Pop()
	layerPtr := registers.Pop()
	twPtr := registers.Pop()
	alphaPtr := registers.Pop()
	n := registers.Pop()
	tmp := registers.Pop()

	f.LDP("res+0(FP)", resPtr, layerPtr)
	f.LDP("twiddles+16(FP)", twPtr, alphaPtr)
	f.MOVD("n+32(FP)", n)

	c := f.loadF31Consts(tmp, vreg(30), vreg(31))
	alphaV, muAlpha := vreg(28), vreg(29)
	one, halfP1 := vreg(26), vreg(27)
	f.broadcastScalar(alphaPtr, tmp, alphaV, "broadcast alpha")
	f.VMUL_S4(c.mu, alphaV, muAlpha, "muAlpha = mu * alpha")
	f.MOVD("$1", tmp)
	f.VDUP(tmp, one.S4(), "broadcast 1")
	f.VUSHR(1, c.p.S4(), halfP1.S4(), "halfP1 = P >> 1")
	f.VADD(halfP1.S4(), one.S4(), halfP1.S4(), "halfP1 = (P + 1) / 2")

	f.loopBlocks(n, func(k int) {
		// p[i], q[i] = V(2i), V(2i+1): VLD2 deinterleaves the pairs.
		p := make([]arm64.VectorRegister, k)
		q := make([]arm64.VectorRegister, k)
		for i := range k {
			p[i], q[i] = vreg(2*i), vreg(2*i+1)
		}
		tw, s, t0 := vgroup(8, k), vgroup(12, k), vgroup(16, k)
		alphas := make([]arm64.VectorRegister, k)
		for i := range alphas {
			alphas[i] = alphaV
		}

		for i := range k {
			f.WriteLn(fmt.Sprintf("    VLD2.P 32(%s), [%s, %s]", layerPtr, p[i].S4(), q[i].S4()))
		}
		f.ldP(twPtr, tw)

		// s = p + q
		f.modAdd(c, p, q, s, t0)
		// p = p - q
		f.modSub(c, p, q, p, t0)
		// q = s / 2 = (s >> 1) + (s & 1) * (P + 1) / 2
		for i := range k {
			f.VAND(s[i].B16(), one.B16(), q[i].B16(), "o = s & 1")
		}
		for i := range k {
			f.VMUL_S4(q[i], halfP1, q[i], "o = o * (P + 1) / 2")
		}
		for i := range k {
			f.VUSRA(1, s[i].S4(), q[i].S4(), "o += s >> 1")
		}
		// p = (p - q) * twiddles * alpha
		f.montMul(c, p, tw, p, t0, s, nil)
		f.montMul(c, p, alphas, p, t0, s, &muAlpha)
		// res = s / 2 + p
		f.modAdd(c, q, p, t0, s)
		f.stP(t0, resPtr)
	})

	registers.Push(resPtr, layerPtr, twPtr, alphaPtr, n, tmp)
	f.RET()
}

func (f *FFArm64) generateBatchInvFwdVecF31() {
	f.Comment("batchInvFwdVec(acc, prefix, row, a, one *Element, n uint64)")
	f.Comment("for every lane k:")
	f.Comment("  prefix[k] = a[k] != 0 ? acc[k] : 0")
	f.Comment("  row[k]    = a[k] != 0 ? a[k] : one")
	f.Comment("  acc[k]    = acc[k] * row[k]")
	f.Comment("n is the number of blocks of 4 uint32 to process")
	registers := f.FnHeader("batchInvFwdVec", 0, 48)
	defer f.AssertCleanStack(0, 0)

	accPtr := registers.Pop()
	prefixPtr := registers.Pop()
	rowPtr := registers.Pop()
	aPtr := registers.Pop()
	onePtr := registers.Pop()
	n := registers.Pop()
	accOut := registers.Pop()
	tmp := registers.Pop()

	f.LDP("acc+0(FP)", accPtr, prefixPtr)
	f.LDP("row+16(FP)", rowPtr, aPtr)
	f.LDP("one+32(FP)", onePtr, n)
	f.MOVD(accPtr, accOut)

	c := f.loadF31Consts(tmp, vreg(30), vreg(31))
	one := vreg(29)
	f.broadcastScalar(onePtr, tmp, one, "broadcast one")

	f.loopBlocks(n, func(k int) {
		acc, a, nz := vgroup(0, k), vgroup(4, k), vgroup(8, k)
		t0, t1 := vgroup(12, k), vgroup(16, k)
		f.ldP(aPtr, a)
		f.ldP(accPtr, acc)
		for i := range k {
			f.WriteLn(fmt.Sprintf("    VCMTST %s, %s, %s // nz = a != 0", a[i].S4(), a[i].S4(), nz[i].S4()))
		}
		for i := range k {
			f.VAND(acc[i].B16(), nz[i].B16(), t0[i].B16(), "prefix = nz ? acc : 0")
		}
		f.stP(t0, prefixPtr)
		for i := range k {
			// BIF a, one, nz: insert one into a where nz is false
			f.WriteLn(fmt.Sprintf("    VBIF %s, %s, %s // row = nz ? a : one", nz[i].B16(), one.B16(), a[i].B16()))
		}
		f.stP(a, rowPtr)
		f.montMul(c, acc, a, acc, t0, t1, nil)
		f.stP(acc, accOut)
	})

	registers.Push(accPtr, prefixPtr, rowPtr, aPtr, onePtr, n, accOut, tmp)
	f.RET()
}

func (f *FFArm64) generateBatchInvBwdVecF31() {
	f.Comment("batchInvBwdVec(acc, res, prefix, row, one *Element, n uint64)")
	f.Comment("for every lane k:")
	f.Comment("  res[k] = acc[k] * prefix[k]")
	f.Comment("  acc[k] = acc[k] * (row[k] != 0 ? row[k] : one)")
	f.Comment("n is the number of blocks of 4 uint32 to process")
	registers := f.FnHeader("batchInvBwdVec", 0, 48)
	defer f.AssertCleanStack(0, 0)

	accPtr := registers.Pop()
	resPtr := registers.Pop()
	prefixPtr := registers.Pop()
	rowPtr := registers.Pop()
	onePtr := registers.Pop()
	n := registers.Pop()
	accOut := registers.Pop()
	tmp := registers.Pop()

	f.LDP("acc+0(FP)", accPtr, resPtr)
	f.LDP("prefix+16(FP)", prefixPtr, rowPtr)
	f.LDP("one+32(FP)", onePtr, n)
	f.MOVD(accPtr, accOut)

	c := f.loadF31Consts(tmp, vreg(30), vreg(31))
	one := vreg(29)
	f.broadcastScalar(onePtr, tmp, one, "broadcast one")

	f.loopBlocks(n, func(k int) {
		acc, pre, row := vgroup(0, k), vgroup(4, k), vgroup(8, k)
		t0, t1 := vgroup(12, k), vgroup(16, k)
		// row may be res: it is read before res is written.
		f.ldP(rowPtr, row)
		f.ldP(prefixPtr, pre)
		f.ldP(accPtr, acc)
		for i := range k {
			f.WriteLn(fmt.Sprintf("    VCMTST %s, %s, %s // nz = row != 0", row[i].S4(), row[i].S4(), t0[i].S4()))
		}
		for i := range k {
			f.WriteLn(fmt.Sprintf("    VBIF %s, %s, %s // row = nz ? row : one", t0[i].B16(), one.B16(), row[i].B16()))
		}
		f.montMul(c, acc, pre, pre, t0, t1, nil)
		f.stP(pre, resPtr)
		f.montMul(c, acc, row, acc, t0, t1, nil)
		f.stP(acc, accOut)
	})

	registers.Push(accPtr, resPtr, prefixPtr, rowPtr, onePtr, n, accOut, tmp)
	f.RET()
}
