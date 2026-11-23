// Copyright 2020-2026 Consensys Incorporated
// Licensed under the Apache License, Version 2.0. See the LICENSE file for details.

package arm64

import (
	"fmt"

	"github.com/consensys/bavard/arm64"
)

// NEON kernels of the koalabear degree-4 extension vectors. The tower is
// 𝔽_{p^4} = 𝔽_{p^2}[v]/(v² − u), 𝔽_{p^2} = 𝔽_p[u]/(u² − 3); an E4 is stored
// as its 4 fr coordinates (B0.A0, B0.A1, B1.A0, B1.A1), i.e. 16 bytes, so 4
// E4 fill 4 vector registers and LD4.4S / ST4.4S transpose them for free:
// after VLD4, register k holds coordinate k of the 4 E4.

type e4NeonOp int

const (
	e4NeonMul e4NeonOp = iota
	e4NeonMulAcc
	e4NeonInnerProd
	e4NeonScalarMul
	e4NeonScalarMulAcc
	e4NeonScalarMulAdd
	e4NeonScalarMulAddBase

	e4NeonMulByElement
	e4NeonMulAccByElement
	e4NeonScalarMulBase
	e4NeonScalarMulAccBase
	e4NeonInnerProdByElement
)

func (op e4NeonOp) name() string {
	switch op {
	case e4NeonMul:
		return "vectorMul_E4_neon"
	case e4NeonMulAcc:
		return "vectorMulAcc_E4_neon"
	case e4NeonInnerProd:
		return "vectorInnerProduct_E4_neon"
	case e4NeonScalarMul:
		return "vectorScalarMul_E4_neon"
	case e4NeonScalarMulAcc:
		return "vectorScalarMulAcc_E4_neon"
	case e4NeonScalarMulAdd:
		return "vectorScalarMulAdd_E4_neon"
	case e4NeonScalarMulAddBase:
		return "vectorScalarMulAddBase_E4_neon"
	case e4NeonMulByElement:
		return "vectorMulByElement_E4_neon"
	case e4NeonMulAccByElement:
		return "vectorMulAccByElement_E4_neon"
	case e4NeonScalarMulBase:
		return "vectorScalarMulBase_E4_neon"
	case e4NeonScalarMulAccBase:
		return "vectorScalarMulAccBase_E4_neon"
	case e4NeonInnerProdByElement:
		return "vectorInnerProductByElement_E4_neon"
	}
	panic("invalid e4NeonOp")
}

// generateF31ExtE4 emits all the E4 kernels.
func (g *e6Neon) generateF31ExtE4() {
	for _, op := range []e4NeonOp{
		e4NeonMul, e4NeonMulAcc, e4NeonInnerProd,
		e4NeonScalarMul, e4NeonScalarMulAcc, e4NeonScalarMulAdd, e4NeonScalarMulAddBase,
	} {
		g.generateMulE4(op)
	}
	for _, op := range []e4NeonOp{
		e4NeonMulByElement, e4NeonMulAccByElement,
		e4NeonScalarMulBase, e4NeonScalarMulAccBase, e4NeonInnerProdByElement,
	} {
		g.generateMulBaseE4(op)
	}
	g.generateFoldPairsE4()
}

// popQuad returns four consecutive registers, as required by LD4/ST4.
func (g *e6Neon) popQuad() [4]arm64.VectorRegister {
	for i := 2; i+3 < 32; i++ {
		if g.free[i] && g.free[i+1] && g.free[i+2] && g.free[i+3] {
			for j := range 4 {
				g.free[i+j] = false
			}
			return [4]arm64.VectorRegister{vreg(i), vreg(i + 1), vreg(i + 2), vreg(i + 3)}
		}
	}
	panic("e6Neon: no four consecutive vector registers available")
}

func quadList(r [4]arm64.VectorRegister) string {
	return fmt.Sprintf("[%s.S4, %s.S4, %s.S4, %s.S4]", r[0], r[1], r[2], r[3])
}

// e4Scalar holds the 9 operands of the E4 product taken on the scalar side
// (see mulE4), broadcast once outside the loop.
type e4Scalar [9]arm64.VectorRegister

// loadE4Scalar broadcasts the coordinates of the E4 at ptr and precomputes the
// sums used by mulE4.
func (g *e6Neon) loadE4Scalar(ptr, tmp arm64.Register) e4Scalar {
	var s e4Scalar
	for i := range s {
		s[i] = g.pop()
	}
	// y0, y1, y2, y3 → s[0], s[1], s[3], s[4]
	for k, idx := range [4]int{0, 1, 3, 4} {
		g.ins("MOVWU %d(%s), %s", 4*k, ptr, tmp)
		g.ins("VDUP %s, %s.S4", tmp, s[idx])
	}
	g.add(s[0], s[1], s[2]) // y0 + y1
	g.add(s[3], s[4], s[5]) // y2 + y3
	g.add(s[0], s[3], s[6]) // y0 + y2
	g.add(s[1], s[4], s[7]) // y1 + y3
	g.add(s[6], s[7], s[8])
	return s
}

// e2MulE4 sets (re, xr) = (xr + xi·u)(yr + yi·u) with u² = 3, re being a fresh
// register; xi is released.
//
// If ys is empty, yr and yi are consumed: ys = yr + yi is computed in place
// and both are released.
func (g *e6Neon) e2MulE4(xr, xi, yr, yi, ys arm64.VectorRegister) (re arm64.VectorRegister) {
	re = g.pop()
	t := g.pop()
	g.mul(xr, yr, re) // m0
	g.mul(xi, yi, t)  // m1
	g.add(xr, xi, xr)
	g.push(xi)
	consume := ys == ""
	if consume {
		g.add(yr, yi, yr)
		ys = yr
	}
	g.mul(xr, ys, xr) // m2
	if consume {
		g.push(yr, yi)
	}
	g.sub(xr, re, xr)
	g.sub(xr, t, xr) // im = m2 − m0 − m1
	g.add(re, t, re)
	g.add(re, t, re)
	g.add(re, t, re) // re = m0 + 3·m1
	g.push(t)
	return re
}

// mulE4 writes x·y into dst (fresh registers, may be a quad for ST4), x being
// consumed. y is either the 4 coordinates of the second factor (consumed too)
// or, if s is not nil, a broadcast scalar.
//
// Karatsuba over 𝔽_{p^2}, as E4.Mul and the AVX-512 kernels, with X = B0 + B1·v:
//
//	d = X.B0·Y.B0, c = X.B1·Y.B1, a = (X.B0 + X.B1)(Y.B0 + Y.B1)
//	Z.B0 = d + u·c = (d.A0 + 3·c.A1) + (d.A1 + c.A0)·u
//	Z.B1 = a − d − c
//
// 9 base-field products; every intermediate is canonical.
func (g *e6Neon) mulE4(x, y []arm64.VectorRegister, s *e4Scalar, dst [4]arm64.VectorRegister) {
	sx0, sx1 := g.pop(), g.pop()
	g.add(x[0], x[2], sx0)
	g.add(x[1], x[3], sx1)

	var d0, c0, p0 arm64.VectorRegister
	if s != nil {
		d0 = g.e2MulE4(x[0], x[1], s[0], s[1], s[2])
		c0 = g.e2MulE4(x[2], x[3], s[3], s[4], s[5])
		p0 = g.e2MulE4(sx0, sx1, s[6], s[7], s[8])
	} else {
		sy0, sy1 := g.pop(), g.pop()
		g.add(y[0], y[2], sy0)
		g.add(y[1], y[3], sy1)
		d0 = g.e2MulE4(x[0], x[1], y[0], y[1], "")
		c0 = g.e2MulE4(x[2], x[3], y[2], y[3], "")
		p0 = g.e2MulE4(sx0, sx1, sy0, sy1, "")
	}
	d1, c1, p1 := x[0], x[2], sx0

	g.sub(p0, d0, dst[2])
	g.sub(dst[2], c0, dst[2]) // a.A0 − d.A0 − c.A0
	g.sub(p1, d1, dst[3])
	g.sub(dst[3], c1, dst[3]) // a.A1 − d.A1 − c.A1
	g.add(d1, c0, dst[1])     // d.A1 + c.A0
	g.add(d0, c1, dst[0])
	g.add(dst[0], c1, dst[0])
	g.add(dst[0], c1, dst[0]) // d.A0 + 3·c.A1
	g.push(d0, d1, c0, c1, p0, p1)
}

// generateMulE4 emits the E4 × E4 kernels (4 E4 per iteration):
//
//	vectorMul_E4_neon(res, a, b *E4, N uint64)                    res[i]  = a[i]·b[i]
//	vectorMulAcc_E4_neon(res, a, b *E4, N uint64)                 res[i] += a[i]·b[i]
//	vectorInnerProduct_E4_neon(res *[4]uint64, a, b *E4, N uint64)
//	vectorScalarMul_E4_neon(res, a, b *E4, N uint64)              res[i]  = a[i]·b
//	vectorScalarMulAcc_E4_neon(res, a, b *E4, N uint64)           res[i] += a[i]·b
//	vectorScalarMulAdd_E4_neon(res, a, b *E4, N uint64)           res[i]  = res[i]·b + a[i]
//	vectorScalarMulAddBase_E4_neon(res *E4, a *fr.Element, b *E4, N uint64)
//	                                                              res[i]  = res[i]·b + a[i]
//
// VLD4 transposes 4 E4 into 4 coordinate registers, the product is computed
// lane-wise (mulE4) and VST4 transposes back. The Acc / Add variants add in
// the transposed layout before the store.
//
// InnerProduct folds each pair of result coordinates (k, k+1) with ADDP
// (lanes < 2q < 2³²) and UADALP into one 2×uint64 accumulator: res[k]
// receives the (unreduced) sum of coordinate k. Each lane grows by < 4q < 2³³
// per iteration, so no overflow before N ≈ 2³³.
//
// Precondition: N % 4 == 0.
func (g *e6Neon) generateMulE4(op e4NeonOp) {
	scalar := op == e4NeonScalarMul || op == e4NeonScalarMulAcc || op == e4NeonScalarMulAdd || op == e4NeonScalarMulAddBase
	horner := op == e4NeonScalarMulAdd || op == e4NeonScalarMulAddBase

	g.Comment(fmt.Sprintf("%s(res, a, b, N): see generateMulE4 in the generator", op.name()))
	g.ins("")
	g.WriteLn(fmt.Sprintf("TEXT ·%s(SB), NOFRAME|NOSPLIT, $0-32", op.name()))
	g.resetRegs()

	resPtr, aPtr, bPtr, n, tmp := arm64.R0, arm64.R1, arm64.R2, arm64.R3, arm64.R4
	g.ins("MOVD res+0(FP), %s", resPtr)
	g.ins("MOVD a+8(FP), %s", aPtr)
	g.ins("MOVD b+16(FP), %s", bPtr)
	g.ins("MOVD N+24(FP), %s", n)
	g.ins("LSR $2, %s, %s", n, n)
	g.loadConstants(tmp)

	var accs []arm64.VectorRegister
	if op == e4NeonInnerProd {
		accs = g.popN(2)
		for _, r := range accs {
			g.ins("VEOR %s.B16, %s.B16, %s.B16", r, r, r)
		}
	}

	var s *e4Scalar
	if scalar {
		sc := g.loadE4Scalar(bPtr, tmp)
		s = &sc
	}

	g.Loop(n, func() {
		x := g.popQuad()
		if horner {
			// the vector being scaled is res itself
			g.ins("VLD4 (%s), %s", resPtr, quadList(x))
		} else {
			g.ins("VLD4.P 64(%s), %s", aPtr, quadList(x))
		}
		var y []arm64.VectorRegister
		if !scalar {
			yq := g.popQuad()
			g.ins("VLD4.P 64(%s), %s", bPtr, quadList(yq))
			y = yq[:]
		}
		z := g.popQuad()
		g.mulE4(x[:], y, s, z)

		switch op {
		case e4NeonInnerProd:
			t := g.pop()
			for j := range 2 {
				g.VADDP(z[2*j], z[2*j+1], t, "lanes < 2q")
				g.VUADALP(t, accs[j])
			}
			g.push(t)
			g.push(z[:]...)
			return
		case e4NeonMulAcc, e4NeonScalarMulAcc, e4NeonScalarMulAdd:
			old := g.popQuad()
			if op == e4NeonScalarMulAdd {
				g.ins("VLD4.P 64(%s), %s", aPtr, quadList(old))
			} else {
				g.ins("VLD4 (%s), %s", resPtr, quadList(old))
			}
			for k := range 4 {
				g.add(z[k], old[k], z[k])
			}
			g.push(old[:]...)
		case e4NeonScalarMulAddBase:
			// a[i] embeds into the first coordinate only
			t := g.pop()
			g.ins("VLD1.P 16(%s), [%s.S4]", aPtr, t)
			g.add(z[0], t, z[0])
			g.push(t)
		}
		g.ins("VST4.P %s, 64(%s)", quadList(z), resPtr)
		g.push(z[:]...)
	})

	if op == e4NeonInnerProd {
		for j := range 2 {
			g.ins("VST1.P [%s.D2], 16(%s)", accs[j], resPtr)
		}
		g.push(accs...)
	}
	if scalar {
		g.push(s[:]...)
	}
	g.assertCleanRegs()
	g.RET()
	g.WriteLn("")
}

// generateMulBaseE4 emits the E4 × base-field kernels, which never transpose
// (4 E4 = 4 registers per iteration, register k holding E4 k):
//
//	vectorMulByElement_E4_neon(res, a *E4, b *fr.Element, N uint64)             res[i]  = a[i]·b[i]
//	vectorMulAccByElement_E4_neon(res, a *E4, b *fr.Element, N uint64)          res[i] += a[i]·b[i]
//	vectorScalarMulBase_E4_neon(res *E4, a *fr.Element, b *E4, N uint64)        res[i]  = b·a[i]
//	vectorScalarMulAccBase_E4_neon(res *E4, a *fr.Element, b *E4, N uint64)     res[i] += b·a[i]
//	vectorInnerProductByElement_E4_neon(res, a *E4, b *fr.Element, N uint64)    *res = Σ a[i]·b[i]
//
// The base-field factor of E4 k is broadcast with DUP from lane k of one
// 4-scalar load. For the Scalar variants the E4 factor b is one constant
// register (its 4 coordinates), with b·μ precomputed. InnerProductByElement
// keeps two E4 accumulators (lane-wise canonical sums), added and stored at
// the end: *res is a canonical E4.
//
// Precondition: N % 4 == 0.
func (g *e6Neon) generateMulBaseE4(op e4NeonOp) {
	scalar := op == e4NeonScalarMulBase || op == e4NeonScalarMulAccBase
	acc := op == e4NeonMulAccByElement || op == e4NeonScalarMulAccBase
	inner := op == e4NeonInnerProdByElement

	g.Comment(fmt.Sprintf("%s(res, a, b, N): see generateMulBaseE4 in the generator", op.name()))
	g.ins("")
	g.WriteLn(fmt.Sprintf("TEXT ·%s(SB), NOFRAME|NOSPLIT, $0-32", op.name()))
	g.resetRegs()

	resPtr, ePtr, xPtr, n, tmp, rdPtr := arm64.R0, arm64.R1, arm64.R2, arm64.R3, arm64.R4, arm64.R5
	g.ins("MOVD res+0(FP), %s", resPtr)
	if scalar {
		// a is the base-field vector, b the E4 scalar
		g.ins("MOVD a+8(FP), %s", xPtr)
		g.ins("MOVD b+16(FP), %s", ePtr)
	} else {
		g.ins("MOVD a+8(FP), %s", ePtr)
		g.ins("MOVD b+16(FP), %s", xPtr)
	}
	g.ins("MOVD N+24(FP), %s", n)
	g.ins("LSR $2, %s, %s", n, n)
	g.loadConstants(tmp)
	if acc {
		g.ins("MOVD %s, %s", resPtr, rdPtr)
	}

	var tile, tileMu arm64.VectorRegister
	if scalar {
		tile, tileMu = g.pop(), g.pop()
		g.ins("VLD1 (%s), [%s.S4]", ePtr, tile)
		g.VMUL_S4(tile, vMu, tileMu, "b·μ")
	}
	var accs []arm64.VectorRegister
	if inner {
		accs = g.popN(2)
		for _, r := range accs {
			g.ins("VEOR %s.B16, %s.B16, %s.B16", r, r, r)
		}
	}

	g.Loop(n, func() {
		x := g.pop()
		g.ins("VLD1.P 16(%s), [%s.S4]", xPtr, x)
		pat := g.pop()
		e := g.pop()
		for k := range 4 {
			g.ins("VDUP %s.S[%d], %s.S4", x, k, pat)
			if scalar {
				g.mul(pat, tile, e, tileMu)
			} else {
				g.ins("VLD1.P 16(%s), [%s.S4]", ePtr, e)
				g.mul(e, pat, e)
			}
			switch {
			case inner:
				g.add(accs[k%2], e, accs[k%2])
				continue
			case acc:
				old := g.pop()
				g.ins("VLD1.P 16(%s), [%s.S4]", rdPtr, old)
				g.add(e, old, e)
				g.push(old)
			}
			g.ins("VST1.P [%s.S4], 16(%s)", e, resPtr)
		}
		g.push(x, pat, e)
	})

	if inner {
		g.add(accs[0], accs[1], accs[0])
		g.ins("VST1 [%s.S4], (%s)", accs[0], resPtr)
		g.push(accs...)
	}
	if scalar {
		g.push(tile, tileMu)
	}
	g.assertCleanRegs()
	g.RET()
	g.WriteLn("")
}

// generateFoldPairsE4 emits
//
//	vectorFoldPairs_E4_neon(res, layer *E4, tw *fr.Element, alpha *E4, N uint64)
//
// res[j] = (p + q)/2 + alpha·tw[j]·(p − q), p = layer[2j], q = layer[2j+1]
// (see VectorE4.FoldPairs), 4 outputs per iteration. Two VLD4 load layer[2j..2j+7]
// transposed, lanes alternating p and q; UZP1/UZP2 separate them. The halving
// is (s >> 1) + (s & 1)·(q+1)/2. Reads precede the store in each iteration, so
// res may be layer[:N] (res advances half as fast as layer).
//
// Precondition: N % 4 == 0.
func (g *e6Neon) generateFoldPairsE4() {
	const name = "vectorFoldPairs_E4_neon"
	g.Comment(fmt.Sprintf("%s(res, layer, tw, alpha, N): see generateFoldPairsE4 in the generator", name))
	g.ins("")
	g.WriteLn(fmt.Sprintf("TEXT ·%s(SB), NOFRAME|NOSPLIT, $0-40", name))
	g.resetRegs()

	resPtr, lPtr, twPtr, alphaPtr, n, tmp := arm64.R0, arm64.R1, arm64.R2, arm64.R3, arm64.R4, arm64.R5
	g.ins("MOVD res+0(FP), %s", resPtr)
	g.ins("MOVD layer+8(FP), %s", lPtr)
	g.ins("MOVD tw+16(FP), %s", twPtr)
	g.ins("MOVD alpha+24(FP), %s", alphaPtr)
	g.ins("MOVD N+32(FP), %s", n)
	g.ins("LSR $2, %s, %s", n, n)
	g.loadConstants(tmp)

	one, halfQ := g.pop(), g.pop()
	g.ins("MOVD $1, %s", tmp)
	g.ins("VDUP %s, %s.S4", tmp, one)
	g.ins("MOVD $%d, %s", (g.q+1)/2, tmp)
	g.ins("VDUP %s, %s.S4", tmp, halfQ)

	alpha := g.loadE4Scalar(alphaPtr, tmp)

	g.Loop(n, func() {
		lo, hi := g.popQuad(), g.popQuad()
		g.ins("VLD4.P 64(%s), %s", lPtr, quadList(lo))
		g.ins("VLD4.P 64(%s), %s", lPtr, quadList(hi))
		// s[k] = p + q (halved below), d[k] = p − q
		var sum, diff [4]arm64.VectorRegister
		for k := range 4 {
			p := g.pop()
			g.VUZP1(lo[k], hi[k], p)
			g.VUZP2(lo[k], hi[k], lo[k]) // q
			g.push(hi[k])
			sum[k] = g.pop()
			g.add(p, lo[k], sum[k])
			g.sub(p, lo[k], lo[k])
			g.push(p)
			diff[k] = lo[k]
		}
		for k := range 4 {
			t := g.pop()
			g.ins("VAND %s.B16, %s.B16, %s.B16", one, sum[k], t)
			g.VMUL_S4(t, halfQ, t, "(s & 1)·(q+1)/2")
			g.ins("VUSRA $1, %s.S4, %s.S4", sum[k], t)
			g.push(sum[k])
			sum[k] = t
		}
		w, wMu := g.pop(), g.pop()
		g.ins("VLD1.P 16(%s), [%s.S4]", twPtr, w)
		g.VMUL_S4(w, vMu, wMu, "tw·μ")
		for k := range 4 {
			g.mul(diff[k], w, diff[k], wMu)
		}
		g.push(w, wMu)

		z := g.popQuad()
		g.mulE4(diff[:], nil, &alpha, z)
		for k := range 4 {
			g.add(z[k], sum[k], z[k])
		}
		g.push(sum[:]...)
		g.ins("VST4.P %s, 64(%s)", quadList(z), resPtr)
		g.push(z[:]...)
	})

	g.push(alpha[:]...)
	g.push(one, halfQ)
	g.assertCleanRegs()
	g.RET()
	g.WriteLn("")
}
