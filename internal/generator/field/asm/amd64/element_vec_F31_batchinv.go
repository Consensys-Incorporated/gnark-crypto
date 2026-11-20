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

// mulF31 emits res = a * b on 16 packed F31 elements in Montgomery form, as in
// generateMulVecF31; q and qInvNeg hold the broadcast constants and K3 the
// 0b0101... mask. a, b and res must be distinct from the temporaries.
func (f *FFAmd64) mulF31(registers *amd64.Registers, a, b, res, q, qInvNeg amd64.VectorRegister) {
	aOdd := registers.PopV()
	bOdd := registers.PopV()
	p0 := registers.PopV()
	l0 := registers.PopV()
	l1 := registers.PopV()

	f.VPSRLQ("$32", a, aOdd)
	f.VPSRLQ("$32", b, bOdd)
	f.VPMULUDQ(a, b, p0)
	f.VPMULUDQ(aOdd, bOdd, res)
	f.VPMULUDQ(p0, qInvNeg, l0)
	f.VPMULUDQ(res, qInvNeg, l1)
	f.VPMULUDQ(l0, q, l0)
	f.VPMULUDQ(l1, q, l1)
	f.VPADDQ(p0, l0, p0)
	f.VPADDQ(res, l1, res)
	f.VMOVSHDUPk(p0, amd64.K3, res)
	f.VPSUBD(q, res, l1)
	f.VPMINUD(res, l1, res)

	registers.PushV(aOdd, bOdd, p0, l0, l1)
}

// generateBatchInvertVecF31 emits the two row kernels of the interleaved
// Montgomery batch inversion (batchInvertChains), on n blocks of 16 elements;
// a zero element enters its chain as one with a zero running product:
//
//	batchInvFwdVec(acc, prefix, row, a, one *Element, n uint64)
//		prefix[i] = a[i] != 0 ? acc[i] : 0
//		row[i]    = a[i] != 0 ? a[i] : one
//		acc[i]    = acc[i] * row[i]
//
//	batchInvBwdVec(acc, res, prefix, row, one *Element, n uint64)
//		x         = row[i] != 0 ? row[i] : one
//		res[i]    = acc[i] * prefix[i]
//		acc[i]    = acc[i] * x
//
// row may be the very same slice as a (forward) or res (backward): each block is
// loaded before it is stored.
func (f *FFAmd64) generateBatchInvertVecF31(forward bool) {
	name := "batchInvBwdVec"
	if forward {
		name = "batchInvFwdVec"
		f.Comment("batchInvFwdVec(acc, prefix, row, a, one *Element, n uint64)")
		f.Comment("prefix = a != 0 ? acc : 0; row = a != 0 ? a : one; acc *= row")
	} else {
		f.Comment("batchInvBwdVec(acc, res, prefix, row, one *Element, n uint64)")
		f.Comment("x = row != 0 ? row : one; res = acc * prefix; acc *= x")
	}
	f.Comment("n is the number of blocks of 16 elements to process")
	const argSize = 6 * 8
	stackSize := f.StackSize(f.NbWords*2+6, 0, 0)
	registers := f.FnHeader(name, stackSize, argSize, amd64.AX)
	defer f.AssertCleanStack(stackSize, 0)

	addrAcc := registers.Pop()
	addr1 := registers.Pop() // prefix (forward) or res (backward)
	addr2 := registers.Pop() // row (forward) or prefix (backward)
	addr3 := registers.Pop() // a (forward) or row (backward)
	n := registers.Pop()

	q := registers.PopV()
	qInvNeg := registers.PopV()
	one := registers.PopV()

	f.MOVD("$const_q", amd64.AX)
	f.VPBROADCASTD(amd64.AX, q)
	f.MOVD("$const_qInvNeg", amd64.AX)
	f.VPBROADCASTD(amd64.AX, qInvNeg)
	f.MOVQ("one+32(FP)", addrAcc)
	f.MOVL(addrAcc.At(0), amd64.AX)
	f.VPBROADCASTD(amd64.AX, one)

	f.MOVQ("acc+0(FP)", addrAcc)
	if forward {
		f.MOVQ("prefix+8(FP)", addr1)
		f.MOVQ("row+16(FP)", addr2)
		f.MOVQ("a+24(FP)", addr3)
	} else {
		f.MOVQ("res+8(FP)", addr1)
		f.MOVQ("prefix+16(FP)", addr2)
		f.MOVQ("row+24(FP)", addr3)
	}
	f.MOVQ("n+40(FP)", n)

	f.MOVQ(uint64(0b0101010101010101), amd64.AX)
	f.KMOVD(amd64.AX, amd64.K3)

	f.Loop(n, func() {
		x := registers.PopV()
		acc := registers.PopV()
		t := registers.PopV()
		r := registers.PopV()

		f.VMOVDQU32(addr3.At(0), x)
		f.VMOVDQU32(addrAcc.At(0), acc)
		f.VPTESTMD(x, x, amd64.K1, "K1 = nonzero lanes")
		f.VPBLENDMD(x, one, x, amd64.K1, "zeros replaced by one")
		if forward {
			f.VMOVDQA32_Z(acc, amd64.K1, t, "running product, zero for a zero")
			f.VMOVDQU32(t, addr1.At(0))
			f.VMOVDQU32(x, addr2.At(0))
		} else {
			f.VMOVDQU32(addr2.At(0), t)
			f.mulF31(&registers, acc, t, r, q, qInvNeg)
			f.VMOVDQU32(r, addr1.At(0))
		}
		f.mulF31(&registers, acc, x, r, q, qInvNeg)
		f.VMOVDQU32(r, addrAcc.At(0))

		f.ADDQ("$64", addrAcc)
		f.ADDQ("$64", addr1)
		f.ADDQ("$64", addr2)
		f.ADDQ("$64", addr3)

		registers.PushV(x, acc, t, r)
	})
	f.RETAVX()

	registers.PushV(q, qInvNeg, one)
	f.Push(&registers, addrAcc, addr1, addr2, addr3, n)
}
