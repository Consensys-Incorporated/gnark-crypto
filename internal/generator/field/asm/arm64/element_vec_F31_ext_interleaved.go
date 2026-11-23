// Copyright 2020-2026 Consensys Incorporated
// Licensed under the Apache License, Version 2.0. See the LICENSE file for details.

package arm64

import (
	"fmt"
	"slices"

	"github.com/consensys/bavard/arm64"
)

// ilvKernel describes a lane-wise E6 kernel that never transposes. Each
// iteration handles 4 E6 = 24 contiguous fr, i.e. 6 chunks of 4 lanes; for
// chunk k, body receives:
//
//   - in: one register per E6 input stream, chunk k of 4 E6;
//
//   - x: the 4 base-field elements x0..x3 of the iteration (one per E6), spread
//     to match chunk k (if a base stream is given):
//
//     [x0 x0 x0 x0] [x0 x0 x1 x1] [x1 x1 x1 x1] [x2 x2 x2 x2] [x2 x2 x3 x3] [x3 x3 x3 x3]
//
//     or, if embed, only the lanes of coordinate 0 kept (the others zeroed):
//
//     [x0 0 0 0] [0 0 x1 0] [0 0 0 0] [x2 0 0 0] [0 0 x3 0] [0 0 0 0]
//
//   - t, tMu: chunk k of the tiled E6 scalar (period 6, so 3 distinct
//     registers) and t·μ (if a tile stream is given).
//
// body returns one register per output stream, stored as chunk k. Input,
// output and x registers are released after the store; body releases any
// other register it pops.
type ilvKernel struct {
	name string
	args []string // Go parameter names, in order (all pointers; N follows)
	in   []string // E6 input streams
	out  []string // E6 output streams
	x    string   // base-field stream
	tile string   // tiled scalar (≥ 12 fr)

	embed bool
	body  func(k int, in []vReg, x, t, tMu vReg) []vReg
}

func (g *e6Neon) generateInterleavedE6() {
	ks := []ilvKernel{
		{name: "vectorMulByElement_E6_neon", args: []string{"res", "a", "b"}, in: []string{"a"}, out: []string{"res"}, x: "b",
			body: func(_ int, in []vReg, x, _, _ vReg) []vReg {
				g.mul(in[0], x, in[0])
				return in
			}},
		{name: "vectorMulAccByElement_E6_neon", args: []string{"res", "a", "b"}, in: []string{"a", "res"}, out: []string{"res"}, x: "b",
			body: func(_ int, in []vReg, x, _, _ vReg) []vReg {
				g.mul(in[0], x, in[0])
				g.add(in[0], in[1], in[0])
				return in[:1]
			}},
		{name: "vectorScalarMulBase_E6_neon", args: []string{"res", "a", "b"}, out: []string{"res"}, x: "a", tile: "b",
			body: func(_ int, _ []vReg, x, t, tMu vReg) []vReg {
				g.mul(x, t, x, tMu)
				return []vReg{x}
			}},
		{name: "vectorScalarMulAccBase_E6_neon", args: []string{"res", "a", "b"}, in: []string{"res"}, out: []string{"res"}, x: "a", tile: "b",
			body: func(_ int, in []vReg, x, t, tMu vReg) []vReg {
				g.mul(x, t, x, tMu)
				g.add(in[0], x, in[0])
				return in
			}},
		{name: "vectorScalarAdd_E6_neon", args: []string{"res", "a", "cTiled"}, in: []string{"a"}, out: []string{"res"}, tile: "cTiled",
			body: func(_ int, in []vReg, _, t, _ vReg) []vReg {
				g.add(in[0], t, in[0])
				return in
			}},
		{name: "vectorScalarSubFrom_E6_neon", args: []string{"res", "a", "cTiled"}, in: []string{"a"}, out: []string{"res"}, tile: "cTiled",
			body: func(_ int, in []vReg, _, t, _ vReg) []vReg {
				g.sub(t, in[0], in[0])
				return in
			}},
		{name: "vectorAddElement_E6_neon", args: []string{"res", "a", "b"}, in: []string{"a"}, out: []string{"res"}, x: "b", embed: true,
			body: func(_ int, in []vReg, x, _, _ vReg) []vReg {
				g.add(in[0], x, in[0])
				return in
			}},
		{name: "vectorSubElement_E6_neon", args: []string{"res", "a", "b"}, in: []string{"a"}, out: []string{"res"}, x: "b", embed: true,
			body: func(_ int, in []vReg, x, _, _ vReg) []vReg {
				g.sub(in[0], x, in[0])
				return in
			}},
		// res = b − a, b embedded
		{name: "vectorSubFromElement_E6_neon", args: []string{"res", "a", "b"}, in: []string{"a"}, out: []string{"res"}, x: "b", embed: true,
			body: func(_ int, in []vReg, x, _, _ vReg) []vReg {
				g.sub(x, in[0], in[0])
				return in
			}},
		{name: "vectorScalarAddBase_E6_neon", args: []string{"res", "a", "cTiled"}, out: []string{"res"}, x: "a", tile: "cTiled", embed: true,
			body: func(_ int, _ []vReg, x, t, _ vReg) []vReg {
				g.add(t, x, x)
				return []vReg{x}
			}},
		{name: "vectorScalarSubFromBase_E6_neon", args: []string{"res", "a", "cTiled"}, out: []string{"res"}, x: "a", tile: "cTiled", embed: true,
			body: func(_ int, _ []vReg, x, t, _ vReg) []vReg {
				g.sub(t, x, x)
				return []vReg{x}
			}},
		// (a0, a1) = (a0 + a1, a0 − a1)
		{name: "vectorButterfly_E6_neon", args: []string{"a0", "a1"}, in: []string{"a0", "a1"}, out: []string{"a0", "a1"},
			body: func(_ int, in []vReg, _, _, _ vReg) []vReg {
				s := g.pop()
				g.add(in[0], in[1], s)
				g.sub(in[0], in[1], in[1])
				return []vReg{s, in[1]}
			}},
		// a1 *= tw; (a0, a1) = (a0 + a1, a0 − a1)
		{name: "vectorDITWithTwiddles_E6_neon", args: []string{"a0", "a1", "twiddles"}, in: []string{"a0", "a1"}, out: []string{"a0", "a1"}, x: "twiddles",
			body: func(_ int, in []vReg, x, _, _ vReg) []vReg {
				s := g.pop()
				g.mul(in[1], x, in[1])
				g.add(in[0], in[1], s)
				g.sub(in[0], in[1], in[1])
				return []vReg{s, in[1]}
			}},
		// (a0, a1) = (a0 + a1, (a0 − a1)·tw)
		{name: "vectorDIFWithTwiddles_E6_neon", args: []string{"a0", "a1", "twiddles"}, in: []string{"a0", "a1"}, out: []string{"a0", "a1"}, x: "twiddles",
			body: func(_ int, in []vReg, x, _, _ vReg) []vReg {
				s := g.pop()
				g.add(in[0], in[1], s)
				g.sub(in[0], in[1], in[1])
				g.mul(in[1], x, in[1])
				return []vReg{s, in[1]}
			}},
	}
	for _, k := range ks {
		g.generateIlv(k)
	}
}

// xPattern sets pat to chunk k of the base-field elements in x (see
// ilvKernel); masks are the lane-0 and lane-2 masks when embedding.
func (g *e6Neon) xPattern(x, pat vReg, k int, embed bool, mask0, mask2 vReg) {
	if embed && (k == 2 || k == 5) {
		g.ins("VEOR %s.B16, %s.B16, %s.B16", pat, pat, pat)
		return
	}
	switch k {
	case 0, 2, 3, 5:
		g.ins("VDUP %s.S[%d], %s.S4", x, (k*4)/6, pat)
	case 1:
		g.ins("VZIP1 %s.S4, %s.S4, %s.S4", x, x, pat)
	case 4:
		g.ins("VZIP2 %s.S4, %s.S4, %s.S4", x, x, pat)
	}
	if embed {
		m := mask0
		if k == 1 || k == 4 {
			m = mask2
		}
		g.ins("VAND %s.B16, %s.B16, %s.B16", m, pat, pat)
	}
}

// laneMask returns a register with lane l all ones and the others zero.
func (g *e6Neon) laneMask(l int) vReg {
	m := g.pop()
	g.ins("VEOR %s.B16, %s.B16, %s.B16", m, m, m)
	g.ins("MOVD $0xffffffff, %s", gpTmp)
	g.ins("VMOV %s, %s.S[%d]", gpTmp, m, l)
	return m
}

func (g *e6Neon) generateIlv(k ilvKernel) {
	ptrs, n := g.header(k.name, "lane-wise kernel, see generateIlv in the generator", k.args...)
	ptr := func(name string) arm64.Register {
		i := slices.Index(k.args, name)
		if i < 0 {
			panic(fmt.Sprintf("%s: unknown argument %s", k.name, name))
		}
		return ptrs[i]
	}

	// E6 streams both read and written get a separate write pointer.
	spare := []arm64.Register{gpPtr0, gpPtr1, gpPtr2}
	inPtrs := make([]arm64.Register, len(k.in))
	for i, s := range k.in {
		inPtrs[i] = ptr(s)
	}
	outPtrs := make([]arm64.Register, len(k.out))
	for i, s := range k.out {
		outPtrs[i] = ptr(s)
		if slices.Contains(k.in, s) {
			outPtrs[i], spare = spare[0], spare[1:]
			g.ins("MOVD %s, %s", ptr(s), outPtrs[i])
		}
	}

	var mask0, mask2 vReg
	if k.embed {
		mask0, mask2 = g.laneMask(0), g.laneMask(2)
	}
	var tile, tileMu []vReg
	if k.tile != "" {
		tile, tileMu = g.popN(3), g.popN(3)
		for j := range 3 {
			g.ins("VLD1.P 16(%s), [%s.S4]", ptr(k.tile), tile[j])
			g.VMUL_S4(tile[j], vMu, tileMu[j], "t·μ")
		}
	}

	g.Loop(n, func() {
		var x vReg
		if k.x != "" {
			x = g.pop()
			g.ins("VLD1.P 16(%s), [%s.S4]", ptr(k.x), x)
		}
		for c := range 6 {
			in := g.popN(len(k.in))
			for i, r := range in {
				g.ins("VLD1.P 16(%s), [%s.S4]", inPtrs[i], r)
			}
			var pat, t, tMu vReg
			if x != "" {
				pat = g.pop()
				g.xPattern(x, pat, c, k.embed, mask0, mask2)
			}
			if tile != nil {
				t, tMu = tile[c%3], tileMu[c%3]
			}
			out := k.body(c, in, pat, t, tMu)
			for i, r := range out {
				g.ins("VST1.P [%s.S4], 16(%s)", r, outPtrs[i])
			}
			var rel []vReg
			for _, r := range slices.Concat(in, out, []vReg{pat}) {
				if r != "" && !slices.Contains(rel, r) {
					rel = append(rel, r)
				}
			}
			g.push(rel...)
		}
		if x != "" {
			g.push(x)
		}
	})

	if k.embed {
		g.push(mask0, mask2)
	}
	if tile != nil {
		g.push(tile...)
		g.push(tileMu...)
	}
	g.footer()
}
