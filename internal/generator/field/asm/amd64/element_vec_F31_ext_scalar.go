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
	"fmt"

	"github.com/consensys/bavard/amd64"
)

// extScalarOp selects the operation of generateScalarAddTiledVec.
type extScalarOp int

const (
	extScalarAdd     extScalarOp = iota // res[i] = a[i] + c
	extScalarSubFrom                    // res[i] = c - a[i]
)

// generateScalarAddTiledVec emits a broadcast-scalar add (or reversed sub)
// kernel for an extension vector over F31.
//
//	func name(res, a *E, cTiled *fr.Element, N uint64)
//
// The extension scalar c is pre-tiled by the caller into nbZmm*16 fr lanes
// (periodic in the extension degree), so it is loaded into nbZmm registers
// once and each iteration is nbZmm contiguous load / addMod (or subMod) /
// store triples. degree is the extension degree; nbZmm*16 must be a multiple
// of it. N is the number of extension elements and must be a multiple of
// nbZmm*16/degree (checked by the caller).
func (_f *FFAmd64) generateScalarAddTiledVec(name string, degree, nbZmm int, op extScalarOp) {
	const lanes = 16
	if (nbZmm*lanes)%degree != 0 {
		panic(fmt.Sprintf("%s: %d lanes are not a multiple of degree %d", name, nbZmm*lanes, degree))
	}
	blockSize := nbZmm * lanes / degree

	const argSize = 4 * 8
	stackSize := _f.StackSize(_f.NbWords*4+2, 0, 0)
	registers := _f.FnHeader(name, stackSize, argSize, amd64.DX, amd64.AX)
	defer _f.AssertCleanStack(stackSize, 0)
	f := &fieldHelper{FFAmd64: _f, registers: &registers}

	switch op {
	case extScalarAdd:
		f.Comment(fmt.Sprintf("res[i] = a[i] + c, %d elements of degree %d per iteration", blockSize, degree))
	case extScalarSubFrom:
		f.Comment(fmt.Sprintf("res[i] = c - a[i], %d elements of degree %d per iteration", blockSize, degree))
	}

	addrRes := registers.Pop()
	addrA := registers.Pop()
	addrC := registers.Pop()
	N := registers.Pop()

	f.loadQ()

	f.MOVQ("res+0(FP)", addrRes)
	f.MOVQ("a+8(FP)", addrA)
	f.MOVQ("cTiled+16(FP)", addrC)
	f.MOVQ("N+24(FP)", N)

	// N is a multiple of blockSize (checked by the caller); count blocks.
	switch blockSize {
	case 4:
		f.SHRQ("$2", N)
	case 8:
		f.SHRQ("$3", N)
	default:
		panic(fmt.Sprintf("%s: unsupported block size %d", name, blockSize))
	}

	// load the tiled scalar once; the loop never touches *cTiled again, so
	// callers may alias c with res.
	vC := make([]amd64.VectorRegister, nbZmm)
	for k := range vC {
		vC[k] = registers.PopV()
		f.VMOVDQU32(addrC.AtD(k*lanes), vC[k])
	}

	vA := registers.PopV()
	vR := registers.PopV()

	f.Loop(N, func() {
		for k := range nbZmm {
			f.VMOVDQU32(addrA.At(k*8), vA)
			switch op {
			case extScalarAdd:
				f.add(vA, vC[k], vR)
			case extScalarSubFrom:
				f.sub(vC[k], vA, vR)
			}
			f.VMOVDQU32(vR, addrRes.At(k*8))
		}
		step := fmt.Sprintf("$%d", nbZmm*64)
		f.ADDQ(step, addrA)
		f.ADDQ(step, addrRes)
	})

	f.RET()
}

// generateSumVecE6 emits
//
//	func vectorSum_E6_avx512(res *[24]uint64, a *E6, N uint64)
//
// It zero-extends 8 E6 (48 uint32 = 6 ymm) per iteration into qword lanes and
// adds them into 3 zmm accumulators: lane p of a 48-lane window belongs to
// coordinate p%6, and since 24%6 == 0 the second half of the window lines up
// with the first. The caller folds res[p] into coordinate p%6 and reduces mod
// q. Each accumulator lane grows by < 2³² per iteration, so the result is exact
// for N < 2³⁴. Precondition: N % 8 == 0.
func (_f *FFAmd64) generateSumVecE6() {
	const argSize = 3 * 8
	stackSize := _f.StackSize(_f.NbWords*4+2, 0, 0)
	registers := _f.FnHeader("vectorSum_E6_avx512", stackSize, argSize, amd64.DX, amd64.AX)
	defer _f.AssertCleanStack(stackSize, 0)

	addrRes := registers.Pop()
	addrA := registers.Pop()
	N := registers.Pop()

	_f.MOVQ("res+0(FP)", addrRes)
	_f.MOVQ("a+8(FP)", addrA)
	_f.MOVQ("N+16(FP)", N)
	_f.SHRQ("$3", N)

	var acc [3]amd64.VectorRegister
	for k := range acc {
		acc[k] = registers.PopV()
		_f.VPXORQ(acc[k], acc[k], acc[k])
	}
	vA := registers.PopV()

	_f.Loop(N, func() {
		for j := range 6 {
			// 8 uint32 (32 bytes) -> 8 uint64
			_f.VPMOVZXDQ(addrA.At(j*4), vA)
			_f.VPADDQ(vA, acc[j%3], acc[j%3])
		}
		_f.ADDQ("$192", addrA)
	})

	for k := range acc {
		_f.VMOVDQU64(acc[k], addrRes.At(k*8))
	}

	_f.RET()
}
