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

// scalarVecOp selects the kernel emitted by generateScalarVecF31.
type scalarVecOp int

const (
	negVecOp           scalarVecOp = iota // res[i] = -a[i]
	scalarAddVecOp                        // res[i] = a[i] + c
	scalarSubFromVecOp                    // res[i] = c - a[i]
)

// generateScalarVecF31 emits one of the unary / broadcast-scalar kernels on
// base F31 vectors:
//
//	func negVec(res, a *Element, n uint64)              // res[i] = -a[i]
//	func scalarAddVec(res, a, c *Element, n uint64)     // res[i] = a[i] + c
//	func scalarSubFromVec(res, a, c *Element, n uint64) // res[i] = c - a[i]
//
// where n is the number of 16-element blocks. The scalar operand (*c, or zero
// for negVec) is broadcast once before the loop, so callers may pass a pointer
// into res; the loop then runs one addMod / subMod per block: for canonical a,
// -a = min(0 - a, q - a) (as uint32).
func (_f *FFAmd64) generateScalarVecF31(op scalarVecOp) {
	var name string
	argSize := 4 * 8
	switch op {
	case negVecOp:
		name = "negVec"
		argSize = 3 * 8
		_f.Comment("negVec(res, a *Element, n uint64) res[0...n] = -a[0...n]")
	case scalarAddVecOp:
		name = "scalarAddVec"
		_f.Comment("scalarAddVec(res, a, c *Element, n uint64) res[0...n] = a[0...n] + c")
	case scalarSubFromVecOp:
		name = "scalarSubFromVec"
		_f.Comment("scalarSubFromVec(res, a, c *Element, n uint64) res[0...n] = c - a[0...n]")
	}
	_f.Comment("n is the number of blocks of 16 elements to process")

	registers := _f.FnHeader(name, 0, argSize, amd64.AX)
	defer _f.AssertCleanStack(0, 0)
	f := &fieldHelper{FFAmd64: _f, registers: &registers}

	addrRes := registers.Pop()
	addrA := registers.Pop()
	n := registers.Pop()

	f.loadQ()
	f.MOVQ("res+0(FP)", addrRes)
	f.MOVQ("a+8(FP)", addrA)

	// vC is the broadcast scalar operand: zero for negVec.
	vC := registers.PopV()
	if op == negVecOp {
		f.MOVQ("n+16(FP)", n)
		f.VPXORQ(vC, vC, vC, "c = 0")
	} else {
		addrC := registers.Pop()
		f.MOVQ("c+16(FP)", addrC)
		f.MOVQ("n+24(FP)", n)
		f.VPBROADCASTD(addrC.At(0), vC, "broadcast c")
		registers.Push(addrC)
	}

	vA := registers.PopV()
	f.Loop(n, func() {
		f.VMOVDQU32(addrA.At(0), vA)
		switch op {
		case scalarAddVecOp:
			f.add(vA, vC, vA)
		default:
			f.sub(vC, vA, vA)
		}
		f.VMOVDQU32(vA, addrRes.At(0))

		f.ADDQ("$64", addrA)
		f.ADDQ("$64", addrRes)
	})

	f.RET()
}
