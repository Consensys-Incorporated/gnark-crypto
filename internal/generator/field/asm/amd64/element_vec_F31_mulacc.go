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

	f.RET()
}

// generateScalarMulBaseE4 emits
//
//	func vectorScalarMulBase_avx512(res *E4, a *fr.Element, c *E4, N uint64)
//
// which sets res[i] = a[i] * c for a base-field vector a and a fixed E4 scalar
// c. Each coordinate of the result is a base product, so 16 base elements are
// loaded contiguously, multiplied by each broadcast coordinate of c, and the
// four results scattered back with stride 4. Precondition: N % 16 == 0.
func (_f *FFAmd64) generateScalarMulBaseE4() {
	const argSize = 4 * 8
	stackSize := _f.StackSize(_f.NbWords*4+2, 0, 0)
	registers := _f.FnHeader("vectorScalarMulBase_avx512", stackSize, argSize, amd64.DX, amd64.AX)
	defer _f.AssertCleanStack(stackSize, 0)
	f := &fieldHelper{FFAmd64: _f, registers: &registers}

	addrRes := registers.Pop()
	addrA := registers.Pop()
	addrC := registers.Pop()
	N := registers.Pop()

	f.loadQ()
	f.loadQInvNeg()
	f.MOVQ(uint64(0b0101010101010101), amd64.AX)
	f.KMOVD(amd64.AX, amd64.K3)

	f.MOVQ("res+0(FP)", addrRes)
	f.MOVQ("a+8(FP)", addrA)
	f.MOVQ("c+16(FP)", addrC)
	f.MOVQ("N+24(FP)", N)

	maskFFFF := registers.Pop()
	addrIndexGather4 := registers.Pop()
	vIndexGather := registers.PopV()
	f.MOVQ("$0xffffffffffffffff", maskFFFF)
	f.MOVQ("·indexGather4+0(SB)", addrIndexGather4)
	f.VMOVDQU32(addrIndexGather4.At(0), vIndexGather)

	vC := registers.PopVN(4)
	for k := range vC {
		f.VPBROADCASTD(addrC.AtD(k), vC[k])
	}
	vA := registers.PopV()
	vR := registers.PopV()

	f.SHRQ("$4", N)
	f.Loop(N, func() {
		f.VMOVDQU32(addrA.At(0), vA)
		for k := range 4 {
			f.mul(vA, vC[k], vR, true)
			f.KMOVD(maskFFFF, amd64.K1)
			f.VPSCATTERDD(k*4, addrRes, vIndexGather, 4, amd64.K1, vR)
		}
		f.ADDQ("$64", addrA)    // 16 fr.Element
		f.ADDQ("$256", addrRes) // 16 E4
	})

	f.RET()
}
