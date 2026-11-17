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

// generateFoldPairsVecF31 emits
//
//	func foldPairsVec(res, layer, twiddles, alpha *Element, n uint64)
//
// which computes, for n blocks of 16 outputs,
//
//	res[j] = (layer[2j] + layer[2j+1]) / 2 + alpha · twiddles[j] · (layer[2j] − layer[2j+1])
//
// The 32 layer elements of a block are split into even (p) and odd (q) lanes
// without index tables: VPMOVQD keeps the low dword of each qword (the even
// lanes) and, after VPSRLQ $32, the high dword (the odd lanes). Each block is
// fully loaded before res is written, so res may be layer[:16n].
func (_f *FFAmd64) generateFoldPairsVecF31() {
	_f.Comment("foldPairsVec(res, layer, twiddles, alpha *Element, n uint64)")
	_f.Comment("res[j] = (layer[2j] + layer[2j+1])/2 + alpha * twiddles[j] * (layer[2j] - layer[2j+1])")
	_f.Comment("n is the number of blocks of 16 outputs to process")

	const argSize = 5 * 8
	stackSize := _f.StackSize(_f.NbWords*4+2, 0, 0)
	registers := _f.FnHeader("foldPairsVec", stackSize, argSize, amd64.DX, amd64.AX)
	defer _f.AssertCleanStack(stackSize, 0)
	f := &fieldHelper{FFAmd64: _f, registers: &registers}

	addrRes := registers.Pop()
	addrLayer := registers.Pop()
	addrW := registers.Pop()
	addrAlpha := registers.Pop()
	n := registers.Pop()

	f.loadQ()
	f.loadQInvNeg()
	f.MOVQ(uint64(0b0101010101010101), amd64.AX)
	f.KMOVD(amd64.AX, amd64.K3)

	f.MOVQ("res+0(FP)", addrRes)
	f.MOVQ("layer+8(FP)", addrLayer)
	f.MOVQ("twiddles+16(FP)", addrW)
	f.MOVQ("alpha+24(FP)", addrAlpha)
	f.MOVQ("n+32(FP)", n)

	vAlpha := registers.PopV()
	f.VPBROADCASTD(addrAlpha.At(0), vAlpha)

	z0 := registers.PopV()
	z1 := registers.PopV()
	p := registers.PopV()
	q := registers.PopV()
	t := registers.PopV()
	w := registers.PopV()

	f.Loop(n, func() {
		f.VMOVDQU32(addrLayer.At(0), z0)
		f.VMOVDQU32(addrLayer.At(8), z1)

		// even lanes -> p, odd lanes -> q
		f.VPMOVQD(z0, p.Y())
		f.VPMOVQD(z1, t.Y())
		f.VINSERTI64X4(1, t.Y(), p, p)
		f.VPSRLQ("$32", z0, z0)
		f.VPSRLQ("$32", z1, z1)
		f.VPMOVQD(z0, q.Y())
		f.VPMOVQD(z1, t.Y())
		f.VINSERTI64X4(1, t.Y(), q, q)

		f.sub(p, q, t) // t = p - q
		f.add(p, q, p) // p = p + q
		f.halve(p, p)  // p = (p + q) / 2
		f.VMOVDQU32(addrW.At(0), w)
		f.mul(t, w, t, true)
		f.mul(t, vAlpha, t, true)
		f.add(p, t, p)
		f.VMOVDQU32(p, addrRes.At(0))

		f.ADDQ("$128", addrLayer)
		f.ADDQ("$64", addrW)
		f.ADDQ("$64", addrRes)
	})

	f.RET()
}

// generateFoldPairsVecExt emits the FRI pair-folding kernel for an extension of
// degree d ∈ {4, 6} over F31:
//
//	func name(res, layer *E, twiddles *fr.Element, K *fr.Element, N uint64)
//
// computing res[j] = (p + q)/2 + alpha · twiddles[j] · (p − q) with
// p = layer[2j], q = layer[2j+1], for N (a multiple of 16) outputs.
//
// 16 outputs are processed per iteration in transposed form: the d coordinates
// of the 16 p's and of the 16 q's are gathered into 2d zmm (stride 2d dwords,
// table gatherIdx), so (p+q)/2 and (p−q)·twiddles are lane-wise base-field
// operations, the 16 twiddles being a contiguous load. The multiplication by
// the fixed extension scalar alpha is applied as the d×d matrix K (row-major,
// built by the caller from alpha), out_k = s_k + ∑ᵢ K[k][i]·dᵢ. The d result
// coordinates are scattered back with stride d (table scatterIdx). Every
// gather of a block happens before its scatter, so res may be layer[:N].
func (_f *FFAmd64) generateFoldPairsVecExt(name string, d int, gatherIdx, scatterIdx string) {
	const argSize = 5 * 8
	stackSize := _f.StackSize(_f.NbWords*4+2, 0, 0)
	registers := _f.FnHeader(name, stackSize, argSize, amd64.DX, amd64.AX)
	defer _f.AssertCleanStack(stackSize, 0)
	f := &fieldHelper{FFAmd64: _f, registers: &registers}

	f.Comment(fmt.Sprintf("FRI pair folding over a degree-%d extension, 16 outputs per iteration", d))

	addrRes := registers.Pop()
	addrLayer := registers.Pop()
	addrW := registers.Pop()
	addrK := registers.Pop()
	N := registers.Pop()
	maskFFFF := registers.Pop()
	addrIdx := registers.Pop()

	f.loadQ()
	f.loadQInvNeg()
	f.MOVQ(uint64(0b0101010101010101), amd64.AX)
	f.KMOVD(amd64.AX, amd64.K3)

	f.MOVQ("res+0(FP)", addrRes)
	f.MOVQ("layer+8(FP)", addrLayer)
	f.MOVQ("twiddles+16(FP)", addrW)
	f.MOVQ("K+24(FP)", addrK)
	f.MOVQ("N+32(FP)", N)

	vGather := registers.PopV()
	vScatter := registers.PopV()
	f.MOVQ("$0xffffffffffffffff", maskFFFF)
	f.MOVQ("·"+gatherIdx+"+0(SB)", addrIdx)
	f.VMOVDQU32(addrIdx.At(0), vGather)
	f.MOVQ("·"+scatterIdx+"+0(SB)", addrIdx)
	f.VMOVDQU32(addrIdx.At(0), vScatter)

	p := registers.PopVN(d)
	q := registers.PopVN(d)
	w := registers.PopV()
	t := registers.PopV()
	bc := registers.PopV()

	f.SHRQ("$4", N)
	f.Loop(N, func() {
		for k := range d {
			f.KMOVD(maskFFFF, amd64.K1)
			f.KMOVD(maskFFFF, amd64.K2)
			f.VPGATHERDD(4*k, addrLayer, vGather, 4, amd64.K1, p[k])
			f.VPGATHERDD(4*(d+k), addrLayer, vGather, 4, amd64.K2, q[k])
		}
		f.VMOVDQU32(addrW.At(0), w)

		// p_k <- (p_k + q_k) / 2, q_k <- (p_k - q_k) · twiddles
		for k := range d {
			f.sub(p[k], q[k], t)
			f.add(p[k], q[k], p[k])
			f.halve(p[k], p[k])
			f.mul(t, w, q[k], true)
		}

		// out_k = p_k + ∑ᵢ K[k][i] · q_i, scattered to coordinate k
		for k := range d {
			for i := range d {
				f.VPBROADCASTD(addrK.AtD(k*d+i), bc)
				f.mul(q[i], bc, t, true)
				f.add(p[k], t, p[k])
			}
			f.KMOVD(maskFFFF, amd64.K1)
			f.VPSCATTERDD(4*k, addrRes, vScatter, 4, amd64.K1, p[k])
		}

		f.ADDQ(fmt.Sprintf("$%d", 16*2*d*4), addrLayer)
		f.ADDQ("$64", addrW)
		f.ADDQ(fmt.Sprintf("$%d", 16*d*4), addrRes)
	})

	f.RET()
}
