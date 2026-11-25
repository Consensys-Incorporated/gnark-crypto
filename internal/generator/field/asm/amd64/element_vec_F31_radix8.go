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
// m=1 -> m=2 -> m=4 (multiply the second element of each pair by its twiddle,
// then butterfly); DIF runs m=4 -> m=2 -> m=1 (butterfly, then multiply).
//
// A block is 48 fr lanes = 3 packed zmm in memory. It is expanded into 4 zmm
// holding 2 whole E6 each, in lanes 0..11 (lanes 12..15 are don't-care and
// their results are dropped when compressing back):
//
//	R0 = [e0 e1], R1 = [e2 e3], R2 = [e4 e5], R3 = [e6 e7]
//
// so the pairs line up lane by lane:
//
//	m=4: (R0, R2), (R1, R3)  plain add/sub, no permute
//	m=2: (R0, R1), (R2, R3)  plain add/sub, no permute
//	m=1: the two E6 of each register: swap halves (VPERMD), add/sub, blend.
//
// Expansion is 3 VALIGND; compression is 3 VPERMI2D (maskPermE6Pad{0,1,2}).
//
// twiddles points to 3 zmm (48 fr) built by the Go caller, since the twiddles
// are the same for every block: the m=4 multipliers for R2 and R3, then the
// m=2 multiplier for R1 and R3. Lanes that are not multiplied hold one.
//
// Precondition: N % 8 == 0.
func (_f *FFAmd64) generateRadix8VecE6(dit bool) {
	name := "vectorRadix8DIF_E6_avx512"
	if dit {
		name = "vectorRadix8DIT_E6_avx512"
	}
	const argSize = 3 * 8
	// 3 args + 3 table addresses
	stackSize := _f.StackSize(6, 0, 0)
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
	// m=1 add-select: lanes 0..5 hold the first element of the pair.
	kHalf := amd64.K1
	f.MOVQ(uint64(0x003f), amd64.AX)
	f.KMOVD(amd64.AX, kHalf)

	addrC := [3]amd64.Register{registers.Pop(), registers.Pop(), registers.Pop()}
	for i, s := range []string{"·maskPermE6Pad0+0(SB)", "·maskPermE6Pad1+0(SB)", "·maskPermE6Pad2+0(SB)"} {
		f.MOVQ(s, addrC[i])
	}
	vSwap := registers.PopV()
	f.MOVQ("·maskPermE6PadSwap+0(SB)", addrA)
	f.VMOVDQU32(addrA.At(0), vSwap)

	f.MOVQ("a+0(FP)", addrA)
	f.MOVQ("twiddles+8(FP)", addrTw)
	f.MOVQ("N+16(FP)", N)
	f.SHRQ("$3", N)

	w4a, w4b, w2 := registers.PopV(), registers.PopV(), registers.PopV()
	f.VMOVDQU32(addrTw.At(0), w4a)
	f.VMOVDQU32(addrTw.At(8), w4b)
	f.VMOVDQU32(addrTw.At(16), w2)

	// r holds the 4 padded zmm; t is a pool of 4 temporaries. Each stage
	// writes its results into t and swaps the two arrays. The loop body is
	// straight-line, so register roles only need to agree at load and store.
	r := [4]amd64.VectorRegister{registers.PopV(), registers.PopV(), registers.PopV(), registers.PopV()}
	t := [4]amd64.VectorRegister{registers.PopV(), registers.PopV(), registers.PopV(), registers.PopV()}
	vAdd := registers.PopV()
	vSub := registers.PopV()

	// (r[i], r[j]) <- (r[i]+r[j], r[i]-r[j]), written to t[i], t[j]
	pair := func(i, j int) {
		f.add(r[i], r[j], t[i])
		f.sub(r[i], r[j], t[j])
	}
	swapAll := func() { r, t = t, r }
	stage4 := func() { pair(0, 2); pair(1, 3); swapAll() }
	stage2 := func() { pair(0, 1); pair(2, 3); swapAll() }
	stage1 := func() {
		for i := range 4 {
			f.VPERMD(r[i], vSwap, t[i]) // [e_{2i+1} e_{2i}]
			f.add(r[i], t[i], vAdd)
			f.sub(t[i], r[i], vSub)
			f.VPBLENDMD(vAdd, vSub, t[i], kHalf)
		}
		swapAll()
	}
	// r[i] *= w, via t[i] then swapped back.
	mulTw := func(i int, w amd64.VectorRegister) {
		f.mul(r[i], w, t[i], true)
		r[i], t[i] = t[i], r[i]
	}

	f.Loop(N, func() {
		// load the 3 packed zmm into r[0], t[1], t[2], then expand.
		// r[0] = v0 already holds e0 e1 in lanes 0..11.
		f.VMOVDQU32(addrA.At(0), r[0])
		f.VMOVDQU32(addrA.At(8), t[1])
		f.VMOVDQU32(addrA.At(16), t[2])
		// VALIGND $k, lo, hi, dst: dst = (hi:lo) >> 32k
		f.VALIGND("$12", r[0], t[1], r[1]) // v0[12..15], v1[0..11]
		f.VALIGND("$8", t[1], t[2], r[2])  // v1[8..15],  v2[0..7]
		f.VALIGND("$4", t[2], t[2], r[3])  // v2[4..15]

		if dit {
			stage1()
			mulTw(1, w2)
			mulTw(3, w2)
			stage2()
			mulTw(2, w4a)
			mulTw(3, w4b)
			stage4()
		} else {
			stage4()
			mulTw(2, w4a)
			mulTw(3, w4b)
			stage2()
			mulTw(1, w2)
			mulTw(3, w2)
			stage1()
		}

		// compress the 4 padded zmm back into 3 packed zmm
		for i := range 3 {
			f.VMOVDQU32(addrC[i].At(0), t[i])
			f.VPERMI2D(r[i+1], r[i], t[i])
		}
		f.VMOVDQU32(t[0], addrA.At(0))
		f.VMOVDQU32(t[1], addrA.At(8))
		f.VMOVDQU32(t[2], addrA.At(16))
		f.ADDQ("$192", addrA)
	})

	f.RETAVX()
}
