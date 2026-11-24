// Copyright 2020 Consensys Incorporated
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package amd64

import (
	"github.com/consensys/bavard/amd64"
)

// generateRadix8VecE6 emits vectorRadix8DIT_E6_avx512 / vectorRadix8DIF_E6_avx512:
//
//	func name(a *E6, twiddles *fr.Element, N uint64)
//
// Each runs the m=4, m=2 and m=1 FFT stages on every block of 8 consecutive
// E6 elements, keeping the block in registers between stages. DIT runs
// m=1 → m=2 → m=4 (multiply the odd half by its twiddle, then butterfly); DIF
// runs m=4 → m=2 → m=1 (butterfly, then multiply).
//
// A block is 48 fr lanes = 3 zmm (v0, v1, v2); E6 k occupies global lanes
// 6k..6k+5. At distance d (in E6 units) lane 6k+c pairs with lane 6(k^d)+c.
// Each zmm is butterflied against a "partner" zmm holding the lanes it pairs
// with, then add and sub results are blended by a per-zmm lane mask:
//
//	d=4: partners are 256-bit halves ([v1.hi,v2.lo], [v2.hi,v0.lo], [v0.hi,v1.lo]),
//	     built with VSHUFI64X2.
//	d=2: maskPermE6Stride2_{0,1,2} VPERMI2D tables.
//	d=1: maskPermE6Pair{0,1A,2} tables (shared with vectorButterflyPair_E6_avx512);
//	     the zmm1 partner needs lanes of v0, v1 and v2, so v0 and v2 are first
//	     merged into one zmm (their needed lanes don't overlap).
//
// twiddles points to 4 precomputed zmm (64 fr), built by the Go caller since
// the twiddles are the same for every block:
//
//	[0] m=4 multipliers for v1, [1] m=4 multipliers for v2,
//	[2] m=2 multipliers for v1, [3] m=2 multipliers for v2,
//
// with Montgomery one in lanes that are not multiplied. v0 never needs a twiddle.
//
// Precondition: N % 8 == 0.
func (_f *FFAmd64) generateRadix8VecE6(dit bool) {
	name := "vectorRadix8DIF_E6_avx512"
	if dit {
		name = "vectorRadix8DIT_E6_avx512"
	}
	const argSize = 3 * 8
	// 3 args + 6 table addresses
	stackSize := _f.StackSize(9, 0, 0)
	registers := _f.FnHeader(name, stackSize, argSize, amd64.DX, amd64.AX)
	defer _f.AssertCleanStack(stackSize, 0)
	f := &fieldHelper{FFAmd64: _f, registers: &registers}

	addrA := registers.Pop()
	addrTw := registers.Pop()
	N := registers.Pop()

	f.loadQ()
	f.loadQInvNeg()
	// K3: low dword of each qword, used by mul.
	f.MOVQ(uint64(0x5555), amd64.AX)
	f.KMOVD(amd64.AX, amd64.K3)

	// d=1 add-select masks (lanes holding the first element of a pair).
	kP0, kP1, kP2 := amd64.K1, amd64.K2, amd64.K4
	// d=2 sub-select masks (lanes holding the second element of a pair).
	// kS2v1 is also the d=4 add-select mask for v1, and kS2v2 merges v0 and
	// v2 for the d=1 partner of v1.
	kS2v0, kS2v1, kS2v2 := amd64.K5, amd64.K6, amd64.K7
	for _, m := range []struct {
		v uint64
		k amd64.MaskRegister
	}{
		{0xf03f, kP0}, {0x3f03, kP1}, {0x03f0, kP2},
		{0xf000, kS2v0}, {0x00ff, kS2v1}, {0xfff0, kS2v2},
	} {
		f.MOVQ(m.v, amd64.AX)
		f.KMOVD(amd64.AX, m.k)
	}

	addrP1 := [3]amd64.Register{registers.Pop(), registers.Pop(), registers.Pop()}
	addrP2 := [3]amd64.Register{registers.Pop(), registers.Pop(), registers.Pop()}
	for i, s := range []string{"·maskPermE6Pair0+0(SB)", "·maskPermE6Pair1A+0(SB)", "·maskPermE6Pair2+0(SB)"} {
		f.MOVQ(s, addrP1[i])
	}
	for i, s := range []string{"·maskPermE6Stride2_0+0(SB)", "·maskPermE6Stride2_1+0(SB)", "·maskPermE6Stride2_2+0(SB)"} {
		f.MOVQ(s, addrP2[i])
	}

	f.MOVQ("a+0(FP)", addrA)
	f.MOVQ("twiddles+8(FP)", addrTw)
	f.MOVQ("N+16(FP)", N)
	f.SHRQ("$3", N)

	w4v1, w4v2 := registers.PopV(), registers.PopV()
	w2v1, w2v2 := registers.PopV(), registers.PopV()
	f.VMOVDQU32(addrTw.At(0), w4v1)
	f.VMOVDQU32(addrTw.At(8), w4v2)
	f.VMOVDQU32(addrTw.At(16), w2v1)
	f.VMOVDQU32(addrTw.At(24), w2v2)

	// v holds the 3 data zmm; t is a pool of 3 temporaries the stages rotate
	// through (each stage writes its outputs into the partner registers and
	// swaps them with v). The loop body is straight-line, so register roles
	// only need to agree at the load and the store.
	v := [3]amd64.VectorRegister{registers.PopV(), registers.PopV(), registers.PopV()}
	p := [3]amd64.VectorRegister{registers.PopV(), registers.PopV(), registers.PopV()}
	vAdd := registers.PopV()
	vSub := registers.PopV()

	// out = k ? (v + p) : (p - v)   when addSel
	// out = k ? (p - v) : (v + p)   otherwise
	// out is written into p.
	bfly := func(v, p amd64.VectorRegister, k amd64.MaskRegister, addSel bool) {
		f.add(v, p, vAdd)
		f.sub(p, v, vSub)
		if addSel {
			f.VPBLENDMD(vAdd, vSub, p, k)
		} else {
			f.VPBLENDMD(vSub, vAdd, p, k)
		}
	}
	swap := func() { v, p = p, v }

	stage4 := func() {
		f.VSHUFI64X2("$0x4e", v[2], v[1], p[0]) // [v1.hi, v2.lo]
		f.VSHUFI64X2("$0x4e", v[0], v[2], p[1]) // [v2.hi, v0.lo]
		f.VSHUFI64X2("$0x4e", v[1], v[0], p[2]) // [v0.hi, v1.lo]
		// v0 is entirely first elements, v2 entirely second elements.
		f.add(v[0], p[0], p[0])
		bfly(v[1], p[1], kS2v1, true)
		f.sub(p[2], v[2], p[2])
		swap()
	}
	stage2 := func() {
		f.VMOVDQU32(addrP2[0].At(0), p[0])
		f.VPERMI2D(v[1], v[0], p[0])
		f.VMOVDQU32(addrP2[1].At(0), p[1])
		f.VPERMI2D(v[2], v[0], p[1])
		f.VMOVDQU32(addrP2[2].At(0), p[2])
		f.VPERMI2D(v[2], v[1], p[2])
		bfly(v[0], p[0], kS2v0, false)
		bfly(v[1], p[1], kS2v1, false)
		bfly(v[2], p[2], kS2v2, false)
		swap()
	}
	stage1 := func() {
		// merge v0 lanes 4..15 with v2 lanes 0..3 (vSub is free here)
		f.VPBLENDMD(v[0], v[2], vSub, kS2v2)
		f.VMOVDQU32(addrP1[0].At(0), p[0])
		f.VPERMI2D(v[1], v[0], p[0])
		f.VMOVDQU32(addrP1[1].At(0), p[1])
		f.VPERMI2D(v[1], vSub, p[1])
		f.VMOVDQU32(addrP1[2].At(0), p[2])
		f.VPERMI2D(v[2], v[1], p[2])
		bfly(v[0], p[0], kP0, true)
		bfly(v[1], p[1], kP1, true)
		bfly(v[2], p[2], kP2, true)
		swap()
	}
	// v[i] *= w, in place via p[i] then swapped back.
	mulTw := func(w1, w2 amd64.VectorRegister) {
		f.mul(v[1], w1, p[1], true)
		f.mul(v[2], w2, p[2], true)
		v[1], p[1] = p[1], v[1]
		v[2], p[2] = p[2], v[2]
	}

	f.Loop(N, func() {
		f.VMOVDQU32(addrA.At(0), v[0])
		f.VMOVDQU32(addrA.At(8), v[1])
		f.VMOVDQU32(addrA.At(16), v[2])

		if dit {
			stage1()
			mulTw(w2v1, w2v2)
			stage2()
			mulTw(w4v1, w4v2)
			stage4()
		} else {
			stage4()
			mulTw(w4v1, w4v2)
			stage2()
			mulTw(w2v1, w2v2)
			stage1()
		}

		f.VMOVDQU32(v[0], addrA.At(0))
		f.VMOVDQU32(v[1], addrA.At(8))
		f.VMOVDQU32(v[2], addrA.At(16))
		f.ADDQ("$192", addrA)
	})

	f.RETAVX()
}
