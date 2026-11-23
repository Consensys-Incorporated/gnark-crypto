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
	"slices"

	"github.com/consensys/bavard/amd64"
)

// extMixedOp selects the operation of generateMixedAddVecExt.
type extMixedOp int

const (
	extMixedAdd     extMixedOp = iota // res[i] = a[i] + b[i]
	extMixedSub                       // res[i] = a[i] - b[i]
	extMixedSubFrom                   // res[i] = b[i] - a[i]
)

// generateMixedAddVecExt emits an extension ± base-field kernel over F31. With
// tiled == false it is element-wise:
//
//	func name(res, a *E, b *fr.Element, N uint64)
//
// a and res are extension vectors (degree fr coordinates each, laid out
// contiguously) and b is a base-field vector. With tiled == true the extension
// operand is a broadcast scalar c, pre-tiled by the caller into nbZmm*16 fr
// lanes (as in generateScalarAddTiledVec) and loaded into registers once, and
// the base-field vector is a:
//
//	func name(res *E, a *fr.Element, cTiled *fr.Element, N uint64)
//
// computing res[i] = c + a[i] (extMixedAdd) or c - a[i] (extMixedSub).
//
// In both forms, embedding a base-field element in the extension
// only touches the first coordinate of res[i], so each iteration expands
// nbZmm*16/degree contiguous b values with zero-masked VPEXPANDD straight into
// the first-coordinate lanes of nbZmm registers (all other lanes are zeroed)
// and runs one addMod / subMod per register against a. A zero lane leaves the
// other coordinates of a unchanged (add, sub) or negates them (sub-from), so
// no gather and no index table is needed.
//
// For SubFrom, a is the subtrahend: res[i] = b[i] - a[i]. N is the number of
// extension elements and must be a multiple of nbZmm*16/degree (checked by the
// caller).
func (_f *FFAmd64) generateMixedAddVecExt(name string, degree, nbZmm int, op extMixedOp, tiled bool) {
	const lanes = 16
	if (nbZmm*lanes)%degree != 0 {
		panic(fmt.Sprintf("%s: %d lanes are not a multiple of degree %d", name, nbZmm*lanes, degree))
	}
	blockSize := nbZmm * lanes / degree

	// For each zmm k: the mask of first-coordinate lanes and the dword offset
	// of its first b value (number of first-coordinate lanes before it).
	masks := make([]uint16, nbZmm)
	offsets := make([]int, nbZmm)
	seen := 0
	for k := range nbZmm {
		offsets[k] = seen
		for l := range lanes {
			if (k*lanes+l)%degree == 0 {
				masks[k] |= 1 << l
				seen++
			}
		}
	}
	if seen != blockSize {
		panic(fmt.Sprintf("%s: expanded %d base elements, want %d", name, seen, blockSize))
	}

	// one opmask register per distinct mask (K1, K2, ...), in order of first use.
	kRegs := []amd64.MaskRegister{amd64.K1, amd64.K2, amd64.K3, amd64.K4}
	maskReg := make([]amd64.MaskRegister, nbZmm)
	var distinct []uint16
	for k, m := range masks {
		idx := slices.Index(distinct, m)
		if idx < 0 {
			idx = len(distinct)
			if idx == len(kRegs) {
				panic(fmt.Sprintf("%s: too many distinct masks", name))
			}
			distinct = append(distinct, m)
		}
		maskReg[k] = kRegs[idx]
	}

	const argSize = 4 * 8
	stackSize := _f.StackSize(_f.NbWords*4+2, 0, 0)
	registers := _f.FnHeader(name, stackSize, argSize, amd64.DX, amd64.AX)
	defer _f.AssertCleanStack(stackSize, 0)
	f := &fieldHelper{FFAmd64: _f, registers: &registers}

	// ext / base name the extension and base-field operands in the comments
	// and frame arguments.
	ext, base, extArg, baseArg := "a[i]", "b[i]", "a+8(FP)", "b+16(FP)"
	if tiled {
		if op == extMixedSubFrom {
			panic(fmt.Sprintf("%s: SubFrom is not supported with a tiled scalar", name))
		}
		ext, base, extArg, baseArg = "c", "a[i]", "cTiled+16(FP)", "a+8(FP)"
	}
	desc := fmt.Sprintf("%s in the base field, %d elements of degree %d per iteration", base, blockSize, degree)
	switch op {
	case extMixedAdd:
		f.Comment(fmt.Sprintf("res[i] = %s + %s, %s", ext, base, desc))
	case extMixedSub:
		f.Comment(fmt.Sprintf("res[i] = %s - %s, %s", ext, base, desc))
	case extMixedSubFrom:
		f.Comment(fmt.Sprintf("res[i] = %s - %s, %s", base, ext, desc))
	}

	addrRes := registers.Pop()
	addrA := registers.Pop() // extension operand (vector, or tiled scalar)
	addrB := registers.Pop() // base-field operand
	N := registers.Pop()

	f.loadQ()

	f.MOVQ("res+0(FP)", addrRes)
	f.MOVQ(extArg, addrA)
	f.MOVQ(baseArg, addrB)
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

	for i, m := range distinct {
		f.MOVQ(fmt.Sprintf("$0x%04x", m), amd64.AX)
		f.KMOVW(amd64.AX, kRegs[i])
	}

	// with a tiled scalar, load it once; the loop never reads *cTiled again,
	// so callers may alias c with res.
	var vC []amd64.VectorRegister
	if tiled {
		vC = make([]amd64.VectorRegister, nbZmm)
		for k := range vC {
			vC[k] = registers.PopV()
			f.VMOVDQU32(addrA.AtD(k*lanes), vC[k])
		}
	}

	// with a tiled scalar, vC holds the scalar operand and vA is not needed.
	var vA amd64.VectorRegister
	if !tiled {
		vA = registers.PopV()
	}
	vB := registers.PopV()
	vR := registers.PopV()

	f.Loop(N, func() {
		for k := range nbZmm {
			vA := vA
			if tiled {
				vA = vC[k]
			} else {
				f.VMOVDQU32(addrA.At(k*8), vA)
			}
			// zero-masked expand: the first-coordinate lanes get the next
			// b values, all other lanes are zeroed.
			f.WriteLn(fmt.Sprintf("\tVPEXPANDD.Z %s, %s, %s", addrB.AtD(offsets[k]), maskReg[k], vB))
			switch op {
			case extMixedAdd:
				f.add(vA, vB, vR)
			case extMixedSub:
				f.sub(vA, vB, vR)
			case extMixedSubFrom:
				f.sub(vB, vA, vR)
			}
			f.VMOVDQU32(vR, addrRes.At(k*8))
		}
		step := fmt.Sprintf("$%d", nbZmm*64)
		if !tiled {
			f.ADDQ(step, addrA)
		}
		f.ADDQ(step, addrRes)
		f.ADDQ(fmt.Sprintf("$%d", blockSize*4), addrB)
	})

	f.RET()
}
