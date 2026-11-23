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

// mulAccOp selects the operation of generateMulAccVecF31.
type mulAccOp int

const (
	mulAccVecOp       mulAccOp = iota // res[i] += a[i] * b[i]
	scalarMulAccVecOp                 // res[i] += a[i] * c
	scalarMulAddVecOp                 // res[i] = res[i] * c + b[i]
)

// generateMulAccVecF31 emits one of the fused multiply-add kernels on base F31
// vectors. All three share the signature
//
//	func name(res, x, y *Element, n uint64)
//
// where n is the number of 16-element blocks, and:
//
//	mulAccVec:       res[i] += x[i] * y[i]
//	scalarMulAccVec: res[i] += x[i] * (*y)
//	scalarMulAddVec: res[i] = res[i] * (*y) + x[i]
//
// The scalar *y is broadcast once before the loop, so the caller may pass a
// pointer into res only if it copied it first (the Go wrappers do).
func (_f *FFAmd64) generateMulAccVecF31(op mulAccOp) {
	var name string
	switch op {
	case mulAccVecOp:
		name = "mulAccVec"
		_f.Comment("mulAccVec(res, a, b *Element, n uint64) res[0...n] += a[0...n] * b[0...n]")
	case scalarMulAccVecOp:
		name = "scalarMulAccVec"
		_f.Comment("scalarMulAccVec(res, a, c *Element, n uint64) res[0...n] += a[0...n] * c")
	case scalarMulAddVecOp:
		name = "scalarMulAddVec"
		_f.Comment("scalarMulAddVec(res, b, c *Element, n uint64) res[0...n] = res[0...n] * c + b[0...n]")
	}
	_f.Comment("n is the number of blocks of 16 elements to process")

	const argSize = 4 * 8
	stackSize := _f.StackSize(_f.NbWords*4+2, 0, 0)
	registers := _f.FnHeader(name, stackSize, argSize, amd64.DX, amd64.AX)
	defer _f.AssertCleanStack(stackSize, 0)
	f := &fieldHelper{FFAmd64: _f, registers: &registers}

	addrRes := registers.Pop()
	addrX := registers.Pop()
	addrY := registers.Pop()
	n := registers.Pop()

	f.loadQ()
	f.loadQInvNeg()
	f.MOVQ(uint64(0b0101010101010101), amd64.AX)
	f.KMOVD(amd64.AX, amd64.K3)

	f.MOVQ("res+0(FP)", addrRes)
	f.MOVQ("a+8(FP)", addrX)
	f.MOVQ("b+16(FP)", addrY)
	f.MOVQ("n+24(FP)", n)

	vX := registers.PopV()
	vY := registers.PopV()
	vR := registers.PopV()
	if op != mulAccVecOp {
		f.VPBROADCASTD(addrY.At(0), vY)
	}

	f.Loop(n, func() {
		f.VMOVDQU32(addrX.At(0), vX)
		switch op {
		case mulAccVecOp:
			f.VMOVDQU32(addrY.At(0), vY)
			f.mul(vX, vY, vX, true)
			f.VMOVDQU32(addrRes.At(0), vR)
			f.add(vR, vX, vR)
		case scalarMulAccVecOp:
			f.mul(vX, vY, vX, true)
			f.VMOVDQU32(addrRes.At(0), vR)
			f.add(vR, vX, vR)
		case scalarMulAddVecOp:
			f.VMOVDQU32(addrRes.At(0), vR)
			f.mul(vR, vY, vR, true)
			f.add(vR, vX, vR)
		}
		f.VMOVDQU32(vR, addrRes.At(0))

		f.ADDQ("$64", addrX)
		if op == mulAccVecOp {
			f.ADDQ("$64", addrY)
		}
		f.ADDQ("$64", addrRes)
	})

	f.RETAVX()
}

// generateInnerProdByElementE6 emits
//
//	func vectorInnerProductByElement_E6_avx512(res *[48]uint64, a *E6, b *fr.Element, N uint64)
//
// which accumulates ∑ a[i]·b[i] for an E6 vector a and a base vector b. Each
// iteration broadcasts 8 base scalars over the 6 fr lanes of their E6 with the
// VPERMD tables of e6FusedMulPrologue, multiplies them by the 48 lanes of 8 E6,
// and, as innerProdVec does, skips the final Montgomery reduction: each 64-bit
// lane product is reduced to (x·y + m·q) / 2³² < 2q and added, zero-extended,
// to 6 uint64 accumulators (even and odd dword lanes of each of the 3 zmm).
// Nothing is stored in the loop.
//
// On return res[16g + 8h + j] holds the partial sum of dword lane 16g + 2j + h
// of the 48-lane window, i.e. of coordinate (16g + 2j + h) % 6; the caller
// reduces and folds. Each accumulator grows by < 2³² per iteration, so this is
// exact for N < 2³⁵. Precondition: N % 8 == 0.
func (_f *FFAmd64) generateInnerProdByElementE6() {
	const argSize = 4 * 8
	stackSize := _f.StackSize(_f.NbWords*4+2, 0, 0)
	registers := _f.FnHeader("vectorInnerProductByElement_E6_avx512", stackSize, argSize, amd64.DX, amd64.AX)
	defer _f.AssertCleanStack(stackSize, 0)
	f := &fieldHelper{FFAmd64: _f, registers: &registers}

	addrRes := registers.Pop()
	addrA := registers.Pop()
	addrB := registers.Pop()
	N := registers.Pop()

	vMask := [3]amd64.VectorRegister{}
	vMask[0], vMask[1], vMask[2] = f.e6FusedMulPrologue()

	f.MOVQ("res+0(FP)", addrRes)
	f.MOVQ("a+8(FP)", addrA)
	f.MOVQ("b+16(FP)", addrB)
	f.MOVQ("N+24(FP)", N)
	f.SHRQ("$3", N)

	var accEven, accOdd [3]amd64.VectorRegister
	for g := range 3 {
		accEven[g] = registers.PopV()
		accOdd[g] = registers.PopV()
		f.VPXORQ(accEven[g], accEven[g], accEven[g])
		f.VPXORQ(accOdd[g], accOdd[g], accOdd[g])
	}

	vBase := registers.PopV()
	vX := registers.PopV() // a lanes
	vY := registers.PopV() // broadcast b lanes
	xOdd := registers.PopV()
	yOdd := registers.PopV()
	p0 := registers.PopV()
	p1 := registers.PopV()
	m0 := registers.PopV()
	m1 := registers.PopV()

	f.Loop(N, func() {
		// 8 base scalars (32 bytes) into the low ymm half of vBase
		f.VMOVDQU32(addrB.At(0), vBase.Y())
		for g := range 3 {
			f.VPERMD(vBase, vMask[g], vY)
			f.VMOVDQU32(addrA.At(8*g), vX)

			f.VPSRLQ("$32", vX, xOdd)
			f.VPSRLQ("$32", vY, yOdd)
			f.VPMULUDQ(vX, vY, p0)
			f.VPMULUDQ(xOdd, yOdd, p1)
			f.VPMULUDQ(p0, f.qInvNeg, m0)
			f.VPMULUDQ(p1, f.qInvNeg, m1)
			f.VPMULUDQ(m0, f.qd, m0)
			f.VPMULUDQ(m1, f.qd, m1)
			f.VPADDQ(p0, m0, p0)
			f.VPSRLQ("$32", p0, p0)
			f.VPADDQ(p0, accEven[g], accEven[g])
			f.VPADDQ(p1, m1, p1)
			f.VPSRLQ("$32", p1, p1)
			f.VPADDQ(p1, accOdd[g], accOdd[g])
		}
		f.ADDQ("$192", addrA) // 8 E6
		f.ADDQ("$32", addrB)  // 8 fr.Element
	})

	for g := range 3 {
		f.VMOVDQU64(accEven[g], addrRes.At(16*g))
		f.VMOVDQU64(accOdd[g], addrRes.At(16*g+8))
	}

	f.RETAVX()
}
