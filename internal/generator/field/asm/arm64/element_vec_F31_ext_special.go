// Copyright 2020-2026 Consensys Incorporated
// Licensed under the Apache License, Version 2.0. See the LICENSE file for details.

package arm64

// generateFoldPairsE6 emits
//
//	vectorFoldPairs_E6_neon(res, layer *E6, twiddles *fr.Element, alpha *e6ScalarTable, N)
//
// res[j] = (p + q)/2 + alpha·tw[j]·(p − q), (p, q) = (layer[2j], layer[2j+1]),
// for N outputs (N % 4 == 0), every operation canonical so that the result
// matches the per-element code.
//
// Each iteration reads 4 pairs. LD3.2D of one pair gives, per E2 coefficient
// j, the register [p.re p.im q.re q.im]; two rounds of UZP1/UZP2 (across pairs
// 0,1 and 2,3, then across the halves) transpose them into one register per
// coordinate for the 4 p and the 4 q. The halved sum is stored right away (the
// product needs all the registers), then (p − q)·tw is multiplied by alpha
// (given by its table, see e6Mul) and added to it in memory.
//
// res may be layer[:N]: an iteration reads its 4 pairs before writing its 4
// outputs, which lie below the pairs read by the next iterations.
func (g *e6Neon) generateFoldPairsE6() {
	ptrs, n := g.header("vectorFoldPairs_E6_neon", "see generateFoldPairsE6 in the generator", "res", "layer", "twiddles", "alpha")
	resPtr, layerPtr, twPtr, alphaPtr := ptrs[0], ptrs[1], ptrs[2], ptrs[3]
	accPtr := gpPtr0
	g.ins("MOVD %s, %s", resPtr, accPtr)

	g.Loop(n, func() {
		// u[h][2j], u[h][2j+1]: [p.re q.re p'.re q'.re] and the same for im,
		// coefficient j, for the pairs (2h, 2h+1).
		var u [2][]vReg
		for h := range 2 {
			lo := g.popTriple()
			hi := g.popTriple()
			g.ins("VLD3.P 48(%s), [%s.D2, %s.D2, %s.D2]", layerPtr, lo[0], lo[1], lo[2])
			g.ins("VLD3.P 48(%s), [%s.D2, %s.D2, %s.D2]", layerPtr, hi[0], hi[1], hi[2])
			u[h] = g.popN(6)
			for j := range 3 {
				g.VUZP1(lo[j], hi[j], u[h][2*j])
				g.VUZP2(lo[j], hi[j], u[h][2*j+1])
			}
			g.push(lo[0], lo[1], lo[2], hi[0], hi[1], hi[2])
		}
		P, Q := make([]vReg, 6), make([]vReg, 6)
		for k := range 6 {
			P[k], Q[k] = g.pop(), g.pop()
			g.VUZP1(u[0][k], u[1][k], P[k])
			g.VUZP2(u[0][k], u[1][k], Q[k])
			g.push(u[0][k], u[1][k])
		}

		tw, twMu := g.pop(), g.pop()
		g.ins("VLD1.P 16(%s), [%s.S4]", twPtr, tw)
		g.VMUL_S4(tw, vMu, twMu, "tw·μ")
		ones, hq := g.pop(), g.pop()
		g.broadcast(1, ones, gpTmp)
		g.broadcast((g.q+1)/2, hq, gpTmp)
		for k := range 6 {
			d := g.pop()
			g.sub(P[k], Q[k], d)
			g.add(P[k], Q[k], P[k])
			g.push(Q[k])
			// P[k] /= 2: (x >> 1) + (x & 1)·(q+1)/2
			h := g.pop()
			g.ins("VUSHR $1, %s.S4, %s.S4", P[k], h)
			g.ins("VAND %s.B16, %s.B16, %s.B16", ones, P[k], P[k])
			g.VMUL_S4(P[k], hq, P[k], "(x & 1)·(q+1)/2")
			g.ins("VADD %s.S4, %s.S4, %s.S4", h, P[k], P[k])
			g.push(h)
			g.mul(d, tw, d, twMu)
			Q[k] = d
		}
		g.push(tw, twMu, ones, hq)

		g.store4(resPtr, P, nil)
		g.push(P...)

		c := g.e6Mul(Q, nil, false, alphaPtr, nil)
		g.store4(accPtr, c[:], g.addFrom(accPtr, false))
		g.push(c[:]...)
	})

	g.footer()
}

// generateButterflyPairE6 emits
//
//	vectorButterflyPair_E6_neon(a *E6, N)
//
// (a[2i], a[2i+1]) = (a[2i] + a[2i+1], a[2i] − a[2i+1]), N % 4 == 0. LD3.2D
// of one pair gives, per coefficient, [x y] as two 64-bit lanes; UZP1/UZP2.2D
// across two pairs gather the x and the y, and ZIP1/ZIP2.2D put the sums and
// differences back.
func (g *e6Neon) generateButterflyPairE6() {
	ptrs, n := g.header("vectorButterflyPair_E6_neon", "see generateButterflyPairE6 in the generator", "a")
	rd, wr := ptrs[0], gpPtr0
	g.ins("MOVD %s, %s", rd, wr)

	g.Loop(n, func() {
		A := g.popTriple()
		B := g.popTriple()
		g.ins("VLD3.P 48(%s), [%s.D2, %s.D2, %s.D2]", rd, A[0], A[1], A[2])
		g.ins("VLD3.P 48(%s), [%s.D2, %s.D2, %s.D2]", rd, B[0], B[1], B[2])
		for j := range 3 {
			x, y, s := g.pop(), g.pop(), g.pop()
			g.ins("VUZP1 %s.D2, %s.D2, %s.D2", B[j], A[j], x)
			g.ins("VUZP2 %s.D2, %s.D2, %s.D2", B[j], A[j], y)
			g.add(x, y, s)
			g.sub(x, y, y)
			g.ins("VZIP1 %s.D2, %s.D2, %s.D2", y, s, A[j])
			g.ins("VZIP2 %s.D2, %s.D2, %s.D2", y, s, B[j])
			g.push(x, y, s)
		}
		g.ins("VST3.P [%s.D2, %s.D2, %s.D2], 48(%s)", A[0], A[1], A[2], wr)
		g.ins("VST3.P [%s.D2, %s.D2, %s.D2], 48(%s)", B[0], B[1], B[2], wr)
		g.push(A[0], A[1], A[2], B[0], B[1], B[2])
	})
	g.footer()
}

// generateSumE6 emits
//
//	vectorSum_E6_neon(res *[12]uint64, a *E6, N)
//	vectorInnerProductByElement_E6_neon(res *[12]uint64, a *E6, b *fr.Element, N)
//
// accumulating a[i] (resp. a[i]·b[i]) without transposing: chunks k and k+3
// of the 4 E6 hold the same coordinates, so they are added (< 2q < 2³²) and
// widened into the uint64 accumulators. res[4k+l] is a partial sum of
// coordinate (4k + l) % 6; it grows by < 2³² per iteration.
func (g *e6Neon) generateSumE6(withBase bool) {
	name, args := "vectorSum_E6_neon", []string{"res", "a"}
	if withBase {
		name, args = "vectorInnerProductByElement_E6_neon", []string{"res", "a", "b"}
	}
	ptrs, n := g.header(name, "see generateSumE6 in the generator", args...)
	resPtr, aPtr := ptrs[0], ptrs[1]

	accLo, accHi := g.popN(3), g.popN(3)
	for _, r := range append(accLo, accHi...) {
		g.ins("VEOR %s.B16, %s.B16, %s.B16", r, r, r)
	}

	g.Loop(n, func() {
		var x vReg
		if withBase {
			x = g.pop()
			g.ins("VLD1.P 16(%s), [%s.S4]", ptrs[2], x)
		}
		var hold [3]vReg
		for k := range 6 {
			e := g.pop()
			g.ins("VLD1.P 16(%s), [%s.S4]", aPtr, e)
			if withBase {
				pat := g.pop()
				g.xPattern(x, pat, k, false, "", "")
				g.mul(e, pat, e)
				g.push(pat)
			}
			if k < 3 {
				hold[k] = e
				continue
			}
			h := hold[k-3]
			g.ins("VADD %s.S4, %s.S4, %s.S4", e, h, h)
			g.ins("VUADDW %s.S2, %s.D2, %s.D2", h, accLo[k-3], accLo[k-3])
			g.ins("VUADDW2 %s.S4, %s.D2, %s.D2", h, accHi[k-3], accHi[k-3])
			g.push(e, h)
		}
		if withBase {
			g.push(x)
		}
	})

	for k := range 3 {
		g.ins("VST1.P [%s.D2], 16(%s)", accLo[k], resPtr)
		g.ins("VST1.P [%s.D2], 16(%s)", accHi[k], resPtr)
	}
	g.push(accLo...)
	g.push(accHi...)
	g.footer()
}
