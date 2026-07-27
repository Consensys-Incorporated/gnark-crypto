// Copyright 2020 ConsenSys Software Inc.
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

func (f *FFAmd64) generateAddVecF31() {
	f.Comment("addVec(res, a, b *Element, n uint64) res[0...n] = a[0...n] + b[0...n]")
	f.Comment("n is the number of blocks of 16 elements to process")

	const argSize = 4 * 8
	stackSize := f.StackSize(f.NbWords*2+4, 0, 0)
	registers := f.FnHeader("addVec", stackSize, argSize, amd64.AX)
	defer f.AssertCleanStack(stackSize, 0)

	// registers & labels we need
	addrA := registers.Pop()
	addrB := registers.Pop()
	addrRes := registers.Pop()
	len := registers.Pop()

	// AVX512 registers
	a := registers.PopV()
	b := registers.PopV()
	t := registers.PopV()
	q := registers.PopV()

	// load q in Z3
	f.MOVD("$const_q", amd64.AX)
	f.VPBROADCASTD(amd64.AX, q)

	// load arguments
	f.MOVQ("res+0(FP)", addrRes)
	f.MOVQ("a+8(FP)", addrA)
	f.MOVQ("b+16(FP)", addrB)
	f.MOVQ("n+24(FP)", len)

	f.Loop(len, func() {
		// a = a + b
		f.VMOVDQU32(addrA.At(0), a)
		f.VMOVDQU32(addrB.At(0), b)
		f.VPADDD(a, b, a, "a = a + b")
		// t = a - q
		f.VPSUBD(q, a, t, "t = a - q")
		// b = min(t, a)
		f.VPMINUD(a, t, b, "b = min(t, a)")

		// move b to res
		f.VMOVDQU32(b, addrRes.At(0), "res = b")

		f.Comment("increment pointers to visit next element")
		f.ADDQ("$64", addrA)
		f.ADDQ("$64", addrB)
		f.ADDQ("$64", addrRes)
	})

	f.RET()

	f.Push(&registers, addrA, addrB, addrRes, len)

}

func (f *FFAmd64) generateSubVecF31() {
	f.Comment("subVec(res, a, b *Element, n uint64) res[0...n] = a[0...n] - b[0...n]")
	f.Comment("n is the number of blocks of 16 elements to process")

	const argSize = 4 * 8
	stackSize := f.StackSize(f.NbWords*2+4, 0, 0)
	registers := f.FnHeader("subVec", stackSize, argSize, amd64.AX)
	defer f.AssertCleanStack(stackSize, 0)

	// registers & labels we need
	addrA := registers.Pop()
	addrB := registers.Pop()
	addrRes := registers.Pop()
	len := registers.Pop()

	// AVX512 registers
	a := registers.PopV()
	b := registers.PopV()
	t := registers.PopV()
	q := registers.PopV()

	// load q in Z3
	f.MOVD("$const_q", amd64.AX)
	f.VPBROADCASTD(amd64.AX, q)

	// load arguments
	f.MOVQ("res+0(FP)", addrRes)
	f.MOVQ("a+8(FP)", addrA)
	f.MOVQ("b+16(FP)", addrB)
	f.MOVQ("n+24(FP)", len)

	f.Loop(len, func() {

		// a = a - b
		f.VMOVDQU32(addrA.At(0), a)
		f.VMOVDQU32(addrB.At(0), b)

		f.VPSUBD(b, a, a, "a = a - b")

		// t = a + q
		f.VPADDD(q, a, t, "t = a + q")

		// b = min(t, a)
		f.VPMINUD(a, t, b, "b = min(t, a)")

		// move b to res
		f.VMOVDQU32(b, addrRes.At(0), "res = b")

		f.Comment("increment pointers to visit next element")
		f.ADDQ("$64", addrA)
		f.ADDQ("$64", addrB)
		f.ADDQ("$64", addrRes)

	})

	f.RET()

	f.Push(&registers, addrA, addrB, addrRes, len)

}

// sumVec res = sum(a[0...n])
func (f *FFAmd64) generateSumVecF31() {
	f.Comment("sumVec(res *uint64, a *[]uint32, n uint64) res = sum(a[0...n])")
	f.Comment("n is the number of blocks of 16 elements to process")
	const argSize = 3 * 8
	stackSize := f.StackSize(f.NbWords*3+2, 0, 0)
	registers := f.FnHeader("sumVec", stackSize, argSize, amd64.DX, amd64.AX)
	defer f.AssertCleanStack(stackSize, 0)

	f.WriteLn(`
	// We load 8 31bits values at a time and accumulate them into an accumulator of
	// 8 quadwords (64bits). The caller then needs to reduce the result mod q.
	// We can safely accumulate ~2**33 31bits values into a single accumulator.
	// That gives us a maximum of 2**33 * 8 = 2**36 31bits values to sum safely.
	`)

	// registers & labels we need
	addrA := registers.Pop()
	addrT := registers.Pop()
	len := registers.Pop()

	// AVX512 registers
	a1 := registers.PopV()
	a2 := registers.PopV()
	acc1 := registers.PopV()
	acc2 := registers.PopV()

	// load arguments
	f.MOVQ("t+0(FP)", addrT)
	f.MOVQ("a+8(FP)", addrA)
	f.MOVQ("n+16(FP)", len)

	// zeroize the accumulators
	f.VXORPS(acc1, acc1, acc1, "acc1 = 0")
	f.VMOVDQA64(acc1, acc2, "acc2 = 0")

	f.Loop(len, func() {
		// 1 cache line is typically 64 bytes, so we maintain 2 accumulators
		f.VPMOVZXDQ(addrA.At(0), a1, "load 8 31bits values in a1")
		f.VPMOVZXDQ(addrA.At(4), a2, "load 8 31bits values in a2")

		f.VPADDQ(a1, acc1, acc1, "acc1 += a1")
		f.VPADDQ(a2, acc2, acc2, "acc2 += a2")

		f.Comment("increment pointers to visit next element")
		f.ADDQ("$64", addrA)
	})

	// store t into res
	f.VPADDQ(acc1, acc2, acc1, "acc1 += acc2")
	f.VMOVDQU64(acc1, addrT.At(0), "res = acc1")

	f.RET()

	f.Push(&registers, addrA, addrT, len)
}

func (f *FFAmd64) generateSumVecSmallF31(size int) {
	// size must be 16 or 24
	if size != 16 && size != 24 {
		panic("size must be 16 or 24")
	}
	fName := fmt.Sprintf("sumVec%d_AVX512", size)
	const argSize = 2 * 8
	stackSize := f.StackSize(f.NbWords*3+2, 0, 0)
	registers := f.FnHeader(fName, stackSize, argSize, amd64.DX, amd64.AX)
	defer f.AssertCleanStack(stackSize, 0)

	// registers & labels we need
	addrA := registers.Pop()
	addrT := registers.Pop()

	// AVX512 registers
	a1 := registers.PopV()
	a2 := registers.PopV()
	a3 := registers.PopV()

	// load arguments
	f.MOVQ("t+0(FP)", addrT)
	f.MOVQ("a+8(FP)", addrA)

	f.VPMOVZXDQ(addrA.AtD(0), a1)
	f.VPMOVZXDQ(addrA.AtD(8), a2)
	if size == 24 {
		f.VPMOVZXDQ(addrA.AtD(16), a3)
		f.VPADDQ(a3, a2, a2)
	}

	f.VPADDQ(a1, a2, a1)
	f.VEXTRACTI64X4(1, a1, a2.Y())
	f.VPADDQ(a1.Y(), a2.Y(), a1.Y())
	f.VEXTRACTI64X2(1, a1.Y(), a2.X())
	f.VPADDQ(a1.X(), a2.X(), a1.X())

	f.PEXTRQ(0, a1.X(), amd64.AX)
	f.PEXTRQ(1, a1.X(), amd64.DX)
	f.ADDQ(amd64.DX, amd64.AX)
	f.MOVQ(amd64.AX, addrT.At(0))

	f.RET()

	f.Push(&registers, addrA, addrT)
}

// mulVec res = a * b
func (f *FFAmd64) generateMulVecF31() {
	f.Comment("mulVec(res, a, b *Element, n uint64) res[0...n] = a[0...n] * b[0...n]")
	f.Comment("n is the number of blocks of 16 elements to process")
	const argSize = 4 * 8
	stackSize := f.StackSize(f.NbWords*2+4, 0, 0)
	registers := f.FnHeader("mulVec", stackSize, argSize, amd64.AX)
	defer f.AssertCleanStack(stackSize, 0)

	f.Comment("code inspired by Plonky3: https://github.com/Plonky3/Plonky3/blob/36e619f3c6526ee86e2e5639a24b3224e1c1700f/monty-31/src/x86_64_avx512/packing.rs#L319")

	// registers & labels we need
	addrA := registers.Pop()
	addrB := registers.Pop()
	addrRes := registers.Pop()
	len := registers.Pop()

	// AVX512 registers
	q := registers.PopV()
	qInvNeg := registers.PopV()

	// load q in Z3
	f.MOVD("$const_q", amd64.AX)
	f.VPBROADCASTD(amd64.AX, q)
	f.MOVD("$const_qInvNeg", amd64.AX)
	f.VPBROADCASTD(amd64.AX, qInvNeg)

	// load arguments
	f.MOVQ("res+0(FP)", addrRes)
	f.MOVQ("a+8(FP)", addrA)
	f.MOVQ("b+16(FP)", addrB)
	f.MOVQ("n+24(FP)", len)

	f.MOVQ(uint64(0b0101010101010101), amd64.AX)
	f.KMOVD(amd64.AX, amd64.K3)

	f.Loop(len, func() {

		a := registers.PopV()
		b := registers.PopV()
		b1 := registers.PopV()
		aOdd := registers.PopV()
		bOdd := registers.PopV()
		b0 := registers.PopV()
		PL0 := registers.PopV()
		PL1 := registers.PopV()
		// a = a * b
		f.VMOVDQU32(addrA.At(0), a)
		f.VMOVDQU32(addrB.At(0), b)

		f.VPSRLQ("$32", a, aOdd) // keep high 32 bits
		f.VPSRLQ("$32", b, bOdd) // keep high 32 bits

		f.VPMULUDQ(a, b, b0)
		f.VPMULUDQ(aOdd, bOdd, b1)
		f.VPMULUDQ(b0, qInvNeg, PL0)
		f.VPMULUDQ(b1, qInvNeg, PL1)

		f.VPMULUDQ(PL0, q, PL0)
		f.VPMULUDQ(PL1, q, PL1)

		f.VPADDQ(b0, PL0, b0)
		f.VPADDQ(b1, PL1, b1)

		f.VMOVSHDUPk(b0, amd64.K3, b1)

		f.VPSUBD(q, b1, PL1)
		f.VPMINUD(b1, PL1, b1)

		f.VMOVDQU32(b1, addrRes.At(0), "res = P")

		f.Comment("increment pointers to visit next element")
		f.ADDQ("$64", addrA)
		f.ADDQ("$64", addrB)
		f.ADDQ("$64", addrRes)

	})
	f.RET()

	f.Push(&registers, addrA, addrB, addrRes, len)

}

// scalarMulVec res = a * b
func (f *FFAmd64) generateScalarMulVecF31() {
	f.Comment("scalarMulVec(res, a, b *Element, n uint64) res[0...n] = a[0...n] * b")
	f.Comment("n is the number of blocks of 16 elements to process")
	const argSize = 4 * 8
	stackSize := f.StackSize(f.NbWords*2+4, 0, 0)
	registers := f.FnHeader("scalarMulVec", stackSize, argSize, amd64.AX)
	defer f.AssertCleanStack(stackSize, 0)

	// registers & labels we need
	addrA := registers.Pop()
	addrB := registers.Pop()
	addrRes := registers.Pop()
	len := registers.Pop()

	// AVX512 registers
	a := registers.PopV()
	b := registers.PopV()
	b1 := registers.PopV()
	aOdd := registers.PopV()
	b0 := registers.PopV()
	PL0 := registers.PopV()
	PL1 := registers.PopV()
	q := registers.PopV()
	qInvNeg := registers.PopV()

	// load q in Z3
	f.MOVD("$const_q", amd64.AX)
	f.VPBROADCASTD(amd64.AX, q)
	f.MOVD("$const_qInvNeg", amd64.AX)
	f.VPBROADCASTD(amd64.AX, qInvNeg)

	f.MOVQ(uint64(0b0101010101010101), amd64.AX)
	f.KMOVD(amd64.AX, amd64.K3)

	// load arguments
	f.MOVQ("res+0(FP)", addrRes)
	f.MOVQ("a+8(FP)", addrA)
	f.MOVQ("b+16(FP)", addrB)
	f.MOVQ("n+24(FP)", len)

	f.VPBROADCASTD(addrB.At(0), b)

	f.Loop(len, func() {

		// a = a * b
		f.VMOVDQU32(addrA.At(0), a)

		f.VPSRLQ("$32", a, aOdd) // keep high 32 bits
		f.VPMULUDQ(a, b, b0)
		f.VPMULUDQ(aOdd, b, b1)
		f.VPMULUDQ(b0, qInvNeg, PL0)
		f.VPMULUDQ(b1, qInvNeg, PL1)

		f.VPMULUDQ(PL0, q, PL0)
		f.VPMULUDQ(PL1, q, PL1)

		f.VPADDQ(b0, PL0, b0)
		f.VPADDQ(b1, PL1, b1)

		f.VMOVSHDUPk(b0, amd64.K3, b1)

		f.VPSUBD(q, b1, PL1)
		f.VPMINUD(b1, PL1, b1)

		f.VMOVDQU32(b1, addrRes.At(0), "res = P")

		f.Comment("increment pointers to visit next element")
		f.ADDQ("$64", addrA)
		f.ADDQ("$64", addrRes)
	})

	f.RET()

	f.Push(&registers, addrA, addrB, addrRes, len)
}

// innerProdVec res = sum(a * b)
func (f *FFAmd64) generateInnerProdVecF31() {
	f.Comment("innerProdVec(t *uint64, a,b *[]uint32, n uint64) res = sum(a[0...n] * b[0...n])")
	f.Comment("n is the number of blocks of 16 elements to process")
	const argSize = 4 * 8
	stackSize := f.StackSize(f.NbWords*4+2, 0, 0)
	registers := f.FnHeader("innerProdVec", stackSize, argSize, amd64.DX, amd64.AX)
	defer f.AssertCleanStack(stackSize, 0)

	f.WriteLn(`
	// Similar to mulVec; we do most of the montgomery multiplication but don't do
	// the final reduction. We accumulate the result like in sumVec and let the caller
	// reduce mod q.
	`)

	// registers & labels we need
	addrA := registers.Pop()
	addrB := registers.Pop()
	addrT := registers.Pop()
	len := registers.Pop()

	// AVX512 registers
	q := registers.PopV()
	qInvNeg := registers.PopV()

	// load q in Z3
	f.MOVD("$const_q", amd64.AX)
	f.VPBROADCASTD(amd64.AX, q)
	f.MOVD("$const_qInvNeg", amd64.AX)
	f.VPBROADCASTD(amd64.AX, qInvNeg)

	// load arguments
	f.MOVQ("t+0(FP)", addrT)
	f.MOVQ("a+8(FP)", addrA)
	f.MOVQ("b+16(FP)", addrB)
	f.MOVQ("n+24(FP)", len)

	a := registers.PopV()
	b := registers.PopV()
	b1 := registers.PopV()
	aOdd := registers.PopV()
	bOdd := registers.PopV()
	b0 := registers.PopV()
	PL0 := registers.PopV()
	PL1 := registers.PopV()

	acc0 := registers.PopV()
	acc1 := registers.PopV()

	f.MOVQ(uint64(0b0101010101010101), amd64.AX)
	f.KMOVD(amd64.AX, amd64.K3)

	// zeroize the accumulators
	f.VXORPS(acc0, acc0, acc0, "acc0 = 0")
	f.VMOVDQA64(acc0, acc1, "acc1 = 0")

	f.Loop(len, func() {

		// a = a * b
		f.VMOVDQU32(addrA.At(0), a)
		f.VMOVDQU32(addrB.At(0), b)

		f.VPSRLQ("$32", a, aOdd) // keep high 32 bits
		f.VPSRLQ("$32", b, bOdd) // keep high 32 bits

		f.VPMULUDQ(a, b, b0)
		f.VPMULUDQ(aOdd, bOdd, b1)
		f.VPMULUDQ(b0, qInvNeg, PL0)
		f.VPMULUDQ(b1, qInvNeg, PL1)

		f.VPMULUDQ(PL0, q, PL0)
		f.VPMULUDQ(PL1, q, PL1)

		f.VPADDQ(b0, PL0, b0)
		f.VPSRLQ("$32", b0, b0)
		f.VPADDQ(b0, acc0, acc0)
		f.VPADDQ(b1, PL1, b1)
		f.VPSRLQ("$32", b1, b1)
		f.VPADDQ(b1, acc1, acc1)

		f.Comment("increment pointers to visit next element")
		f.ADDQ("$64", addrA)
		f.ADDQ("$64", addrB)
	})

	// store t into res
	f.VPADDQ(acc1, acc0, acc1, "acc1 += acc0")
	f.VMOVDQU64(acc1, addrT.At(0), "res = acc1")

	f.RET()

	f.Push(&registers, addrA, addrB, addrT, len)
}

func (f *FFAmd64) generateSumVecE4() {
	// func vectorSum_avx512(res *[4]uint64, a *E4, N uint64)
	const argSize = 3 * 8
	stackSize := f.StackSize(f.NbWords*4+2, 0, 0)
	registers := f.FnHeader("vectorSum_avx512", stackSize, argSize, amd64.DX, amd64.AX)
	defer f.AssertCleanStack(stackSize, 0)

	addrRes := registers.Pop()
	addrA := registers.Pop()
	N := registers.Pop()

	f.MOVQ("res+0(FP)", addrRes)
	f.MOVQ("a+8(FP)", addrA)
	f.MOVQ("N+16(FP)", N)

	// here we load 2 E4 at a time, we zero extend into zmm register of qwords
	// last step, we fold the results into a single YMM register, then return 4 uint64 that the caller
	// reduces mod q.
	vSum := registers.PopV()
	vA := registers.PopV()

	// zeroize vSum
	f.VXORPS(vSum, vSum, vSum, "vSum = 0")

	// N % 2 == 0 (pre condition checked by caller)
	// divide N by 2
	f.SHRQ("$1", N)

	f.Loop(N, func() {
		f.VPMOVZXDQ(addrA.At(0), vA, "load 2 E4 into vA")
		f.VPADDQ(vA, vSum, vSum, "vSum += vA")
		f.ADDQ("$32", addrA)
	})

	vT := registers.PopV()
	f.VEXTRACTI64X4(1, vSum, vT.Y())
	f.VPADDQ(vT.Y(), vSum.Y(), vSum.Y())

	f.VMOVDQU64(vSum.Y(), addrRes.At(0))

	f.RET()
}

func (_f *FFAmd64) generateAddVecE4() {
	// func vectorAdd_avx512(res, a, b *E4, N uint64)
	const argSize = 4 * 8
	stackSize := _f.StackSize(_f.NbWords*4+2, 0, 0)
	registers := _f.FnHeader("vectorAdd_avx512", stackSize, argSize, amd64.DX, amd64.AX)
	defer _f.AssertCleanStack(stackSize, 0)
	f := &fieldHelper{FFAmd64: _f, registers: &registers}

	addrRes := registers.Pop()
	addrA := registers.Pop()
	addrB := registers.Pop()
	N := registers.Pop()

	f.loadQ()

	f.MOVQ("res+0(FP)", addrRes)
	f.MOVQ("a+8(FP)", addrA)
	f.MOVQ("b+16(FP)", addrB)
	f.MOVQ("N+24(FP)", N)

	// N % 4 == 0 (pre condition checked by caller)
	// divide N by 4
	f.SHRQ("$2", N)

	// each e4 is 4 uint32; so we work with blocks of 4 E4 == 16 uint32 == 1 zmm vector
	va := registers.PopV()
	vb := registers.PopV()
	result := registers.PopV()

	f.Loop(N, func() {

		// load args
		f.VMOVDQU32(addrA.At(0), va)
		f.VMOVDQU32(addrB.At(0), vb)

		f.add(va, vb, result)

		// save result
		f.VMOVDQU32(result, addrRes.At(0))

		// increment result by 16uint32
		f.ADDQ("$64", addrRes)
		f.ADDQ("$64", addrB)
		f.ADDQ("$64", addrA)
	})

	f.RET()
}

func (_f *FFAmd64) generateSubVecE4() {
	// func vectorSub_avx512(res, a, b *E4, N uint64)

	const argSize = 4 * 8
	stackSize := _f.StackSize(_f.NbWords*4+2, 0, 0)
	registers := _f.FnHeader("vectorSub_avx512", stackSize, argSize, amd64.DX, amd64.AX)
	defer _f.AssertCleanStack(stackSize, 0)

	f := &fieldHelper{FFAmd64: _f, registers: &registers}

	addrRes := registers.Pop()
	addrA := registers.Pop()
	addrB := registers.Pop()
	N := registers.Pop()

	f.loadQ()
	// f.MOVD("$const_q", amd64.AX)
	// f.VPBROADCASTD(amd64.AX, qd)

	f.MOVQ("res+0(FP)", addrRes)
	f.MOVQ("a+8(FP)", addrA)
	f.MOVQ("b+16(FP)", addrB)
	f.MOVQ("N+24(FP)", N)

	// N % 4 == 0 (pre condition checked by caller)
	// divide N by 4
	f.SHRQ("$2", N)

	// each e4 is 4 uint32; so we work with blocks of 4 E4 == 16 uint32 == 1 zmm vector
	va := registers.PopV()
	vb := registers.PopV()
	result := registers.PopV()

	f.Loop(N, func() {

		// load args
		f.VMOVDQU32(addrA.At(0), va)
		f.VMOVDQU32(addrB.At(0), vb)

		f.sub(va, vb, result)

		// save result
		f.VMOVDQU32(result, addrRes.At(0))

		// increment result by 16uint32
		f.ADDQ("$64", addrRes)
		f.ADDQ("$64", addrB)
		f.ADDQ("$64", addrA)
	})

	f.RET()
}
func (_f *FFAmd64) generateButterflyPairVecE4() {
	// func vectorButterflyPair_avx512(a *E4, N uint64)
	const argSize = 2 * 8
	stackSize := _f.StackSize(_f.NbWords*2+2, 0, 0)
	registers := _f.FnHeader("vectorButterflyPair_avx512", stackSize, argSize, amd64.DX, amd64.AX)
	defer _f.AssertCleanStack(stackSize, 0)
	f := &fieldHelper{FFAmd64: _f, registers: &registers}

	addrA := registers.Pop()
	N := registers.Pop()

	f.loadQ()

	f.MOVQ("a+0(FP)", addrA)
	f.MOVQ("N+8(FP)", N)

	// N % 4 == 0 (pre condition checked by caller)
	// divide N by 4
	f.SHRQ("$2", N)

	// so we load elements 4 by 4.
	// that is, we end up with va == [a0,a1,a2,a3, b0,b1,b2,b3, c0,c1,c2,c3, d0,d1,d2,d3]
	// and we want to compute
	// butterfly between (a,b) and (c,d)
	// so we first need to rearrange va into
	// vb = [b0,b1,b2,b3, a0,a1,a2,a3, d0,d1,d2,d3, c0,c1,c2,c3]
	// and compute the butterfly between va and vb
	// and merge the results
	// note: this is not optimal we should iterate on larger blocks, it will divide by 2 number of ops.
	va := registers.PopV()
	vb := registers.PopV()
	vTmp := registers.PopV()
	resultA := registers.PopV()
	resultB := registers.PopV()

	f.MOVQ(uint64(0b00_11_00_11), amd64.AX)
	f.KMOVQ(amd64.AX, amd64.K2)

	addrVInterleaveIndices := registers.Pop()
	vInterleaveIndices := registers.PopV()
	f.MOVQ("·vInterleaveIndices+0(SB)", addrVInterleaveIndices)
	f.VMOVDQU64(addrVInterleaveIndices.At(0), vInterleaveIndices)

	// goes from
	// in0 = [a0 a1 a2 a3 a4 a5 a6 a7 a8 a9 a10 a11 a12 a13 a14 a15]
	// in1 = [b0 b1 b2 b3 b4 b5 b6 b7 b8 b9 b10 b11 b12 b13 b14 b15]
	// to
	// in0 = [a0 a1 a2 a3 b0 b1 b2 b3 a8 a9 a10 a11 b8 b9 b10 b11]
	// in1 = [a4 a5 a6 a7 b4 b5 b6 b7 a12 a13 a14 a15 b12 b13 b14 b15]
	permute4x4 := f.Define("permute4x4", 4, func(args ...any) {
		x := args[0]
		y := args[1]
		vInterleaveIndices := args[2]
		tmp := args[3]
		f.VMOVDQA64(vInterleaveIndices, tmp)
		f.VPERMI2Q(y, x, tmp)
		f.VPBLENDMQ(x, tmp, x, amd64.K2)
		f.VPBLENDMQ(tmp, y, y, amd64.K2)
	}, true)

	f.Loop(N, func() {

		// load args
		f.VMOVDQU32(addrA.At(0), va)
		f.VMOVDQA32(va, vb)

		permute4x4(va, vb, vInterleaveIndices, vTmp)

		// a' = a + b
		f.add(va, vb, resultA)
		// b' = a - b
		f.sub(va, vb, resultB)

		// merge the results
		permute4x4(resultA, resultB, vInterleaveIndices, vTmp)

		// save result
		f.VMOVDQU32(resultA, addrA.At(0))

		// increment result by 16uint32
		f.ADDQ("$64", addrA)
	})

	f.RET()

}

func (_f *FFAmd64) generateButterflyPairVecE6() {
	// func vectorButterflyPair_E6_avx512(a *E6, N uint64)
	//
	// Precondition: N % 8 == 0; processes 8 E6 (= 48 fr lanes = 3 zmm) per
	// outer iteration. For each zmm we build the partner lane vector with
	// VPERMI2D and blend add/sub results into the original E6 order.
	const argSize = 2 * 8
	stackSize := _f.StackSize(_f.NbWords*2+2, 0, 0)
	registers := _f.FnHeader("vectorButterflyPair_E6_avx512", stackSize, argSize, amd64.DX, amd64.AX)
	defer _f.AssertCleanStack(stackSize, 0)
	f := &fieldHelper{FFAmd64: _f, registers: &registers}

	addrA := registers.Pop()
	N := registers.Pop()

	f.loadQ()

	f.MOVQ("a+0(FP)", addrA)
	f.MOVQ("N+8(FP)", N)

	// N % 8 == 0 (precondition); we process 8 E6 per outer iter.
	f.SHRQ("$3", N)

	f.MOVQ(uint64(0xf03f), amd64.AX)
	f.KMOVD(amd64.AX, amd64.K1)
	f.MOVQ(uint64(0x3f03), amd64.AX)
	f.KMOVD(amd64.AX, amd64.K2)
	f.MOVQ(uint64(0x03f0), amd64.AX)
	f.KMOVD(amd64.AX, amd64.K3)
	f.MOVQ(uint64(0x3c00), amd64.AX)
	f.KMOVD(amd64.AX, amd64.K4)

	addrMask := registers.Pop()
	vMask0 := registers.PopV()
	vMask1A := registers.PopV()
	vMask1B := registers.PopV()
	vMask2 := registers.PopV()

	f.MOVQ("·maskPermE6Pair0+0(SB)", addrMask)
	f.VMOVDQU32(addrMask.At(0), vMask0)
	f.MOVQ("·maskPermE6Pair1A+0(SB)", addrMask)
	f.VMOVDQU32(addrMask.At(0), vMask1A)
	f.MOVQ("·maskPermE6Pair1B+0(SB)", addrMask)
	f.VMOVDQU32(addrMask.At(0), vMask1B)
	f.MOVQ("·maskPermE6Pair2+0(SB)", addrMask)
	f.VMOVDQU32(addrMask.At(0), vMask2)

	v0 := registers.PopV()
	v1 := registers.PopV()
	v2 := registers.PopV()
	p0 := registers.PopV()
	p1 := registers.PopV()
	p1Cross := registers.PopV()
	p2 := registers.PopV()
	add0 := registers.PopV()
	add1 := registers.PopV()
	add2 := registers.PopV()
	sub := registers.PopV()

	f.Loop(N, func() {
		f.VMOVDQU32(addrA.At(0), v0)
		f.VMOVDQU32(addrA.At(8), v1)
		f.VMOVDQU32(addrA.At(16), v2)

		f.VMOVDQA32(vMask0, p0)
		f.VPERMI2D(v1, v0, p0)

		f.VMOVDQA32(vMask1A, p1)
		f.VPERMI2D(v1, v0, p1)
		f.VMOVDQA32(vMask1B, p1Cross)
		f.VPERMI2D(v2, v1, p1Cross)
		f.VPBLENDMD(p1Cross, p1, p1, amd64.K4)

		f.VMOVDQA32(vMask2, p2)
		f.VPERMI2D(v2, v1, p2)

		f.add(v0, p0, add0)
		f.sub(p0, v0, sub)
		f.VPBLENDMD(add0, sub, add0, amd64.K1)

		f.add(v1, p1, add1)
		f.sub(p1, v1, sub)
		f.VPBLENDMD(add1, sub, add1, amd64.K2)

		f.add(v2, p2, add2)
		f.sub(p2, v2, sub)
		f.VPBLENDMD(add2, sub, add2, amd64.K3)

		f.VMOVDQU32(add0, addrA.At(0))
		f.VMOVDQU32(add1, addrA.At(8))
		f.VMOVDQU32(add2, addrA.At(16))

		f.ADDQ("$192", addrA)
	})

	f.RET()
}

func (_f *FFAmd64) generateButterflyVecE4() {
	// func vectorButterfly_avx512(a, b *E4, N uint64)

	const argSize = 3 * 8
	stackSize := _f.StackSize(_f.NbWords*2+2, 0, 0)
	registers := _f.FnHeader("vectorButterfly_avx512", stackSize, argSize, amd64.DX, amd64.AX)
	defer _f.AssertCleanStack(stackSize, 0)
	f := &fieldHelper{FFAmd64: _f, registers: &registers}

	addrA := registers.Pop()
	addrB := registers.Pop()
	N := registers.Pop()

	f.loadQ()

	f.MOVQ("a+0(FP)", addrA)
	f.MOVQ("b+8(FP)", addrB)
	f.MOVQ("N+16(FP)", N)

	// N % 4 == 0 (pre condition checked by caller)
	// divide N by 4
	f.SHRQ("$2", N)

	// each e4 is 4 uint32; so we work with blocks of 4 E4 == 16 uint32 == 1 zmm vector
	va := registers.PopV()
	vb := registers.PopV()
	resultA := registers.PopV()
	resultB := registers.PopV()

	f.Loop(N, func() {

		// load args
		f.VMOVDQU32(addrA.At(0), va)
		f.VMOVDQU32(addrB.At(0), vb)

		// a' = a + b
		f.add(va, vb, resultA)
		// b' = a - b
		f.sub(va, vb, resultB)

		// save result
		f.VMOVDQU32(resultA, addrA.At(0))
		f.VMOVDQU32(resultB, addrB.At(0))

		// increment result by 16uint32
		f.ADDQ("$64", addrA)
		f.ADDQ("$64", addrB)
	})

	f.RET()
}

// e6FusedMulPrologue emits the setup shared by every per-E6 vector op that
// broadcasts a single fr scalar across the 6 fr lanes of one E6: load q and
// qInvNeg, set K3 = 0x55 (low-bit lane mask used by mul), and load the three
// maskPermDE6_{0,1,2} VPERMD index tables. Returns the three mask registers.
// Caller is responsible for loading N from FP and dividing by 8 (one zmm holds
// the broadcast lanes for 8 E6).
func (f *fieldHelper) e6FusedMulPrologue() (vMask0, vMask1, vMask2 amd64.VectorRegister) {
	f.loadQ()
	f.loadQInvNeg()
	f.MOVQ(uint64(0b01_01_01_01_01_01_01_01), amd64.AX)
	f.KMOVD(amd64.AX, amd64.K3)

	vMask0 = f.registers.PopV()
	vMask1 = f.registers.PopV()
	vMask2 = f.registers.PopV()
	addrMask := f.registers.Pop()
	f.MOVQ("·maskPermDE6_0+0(SB)", addrMask)
	f.VMOVDQU32(addrMask.At(0), vMask0)
	f.MOVQ("·maskPermDE6_1+0(SB)", addrMask)
	f.VMOVDQU32(addrMask.At(0), vMask1)
	f.MOVQ("·maskPermDE6_2+0(SB)", addrMask)
	f.VMOVDQU32(addrMask.At(0), vMask2)
	f.registers.Push(addrMask)
	return
}

func (_f *FFAmd64) generateMulVecElementE6() {
	// func vectorMulByElement_E6_avx512(res, a *E6, b *fr.Element, N uint64)
	//
	// Precondition: N % 8 == 0; processes 8 E6 (= 48 fr lanes = 3 zmm) per
	// outer iteration. Each scalar b[i] is broadcast across the 6 fr lanes of
	// E6[i] via three precomputed VPERMD index tables.

	const argSize = 4 * 8
	stackSize := _f.StackSize(_f.NbWords*4+2, 0, 0)

	registers := _f.FnHeader("vectorMulByElement_E6_avx512", stackSize, argSize, amd64.DX, amd64.AX)
	defer _f.AssertCleanStack(stackSize, 0)
	f := &fieldHelper{FFAmd64: _f, registers: &registers}

	addrRes := registers.Pop()
	addrA := registers.Pop()
	addrB := registers.Pop()
	N := registers.Pop()

	vMask0, vMask1, vMask2 := f.e6FusedMulPrologue()

	f.MOVQ("res+0(FP)", addrRes)
	f.MOVQ("a+8(FP)", addrA)
	f.MOVQ("b+16(FP)", addrB)
	f.MOVQ("N+24(FP)", N)
	f.SHRQ("$3", N)

	va := registers.PopV()
	vb := registers.PopV()
	vbExp := registers.PopV()
	vRes := registers.PopV()

	process := func(offset int, mask amd64.VectorRegister) {
		f.VMOVDQU32(addrA.At(offset), va)
		f.VPERMD(vb, mask, vbExp)
		f.mul(va, vbExp, vRes, true)
		f.VMOVDQU32(vRes, addrRes.At(offset))
	}

	f.Loop(N, func() {
		// load 8 scalars (= 32 bytes) of b into the low ymm half of vb
		f.VMOVDQU32(addrB.At(0), vb.Y())
		process(0, vMask0)
		process(8, vMask1)
		process(16, vMask2)
		f.ADDQ("$192", addrA)
		f.ADDQ("$192", addrRes)
		f.ADDQ("$32", addrB)
	})

	f.RET()
}

func (_f *FFAmd64) generateMulAccVecElementE6() {
	// func vectorMulAccByElement_E6_avx512(dst, a *E6, b *fr.Element, N uint64)
	//
	// dst[i] += a[i] * b[i]. Precondition: N % 8 == 0; processes 8 E6 (= 48 fr
	// lanes = 3 zmm) per outer iteration. Each scalar b[i] is broadcast across
	// the 6 fr lanes of E6[i] via three precomputed VPERMD index tables.

	const argSize = 4 * 8
	stackSize := _f.StackSize(_f.NbWords*4+2, 0, 0)

	registers := _f.FnHeader("vectorMulAccByElement_E6_avx512", stackSize, argSize, amd64.DX, amd64.AX)
	defer _f.AssertCleanStack(stackSize, 0)
	f := &fieldHelper{FFAmd64: _f, registers: &registers}

	addrDst := registers.Pop()
	addrA := registers.Pop()
	addrB := registers.Pop()
	N := registers.Pop()

	vMask0, vMask1, vMask2 := f.e6FusedMulPrologue()

	f.MOVQ("dst+0(FP)", addrDst)
	f.MOVQ("a+8(FP)", addrA)
	f.MOVQ("b+16(FP)", addrB)
	f.MOVQ("N+24(FP)", N)
	f.SHRQ("$3", N)

	va := registers.PopV()
	vb := registers.PopV()
	vbExp := registers.PopV()
	vRes := registers.PopV()
	vAcc := registers.PopV()

	process := func(offset int, mask amd64.VectorRegister) {
		f.VMOVDQU32(addrA.At(offset), va)
		f.VPERMD(vb, mask, vbExp)
		f.mul(va, vbExp, vRes, true)
		f.VMOVDQU32(addrDst.At(offset), vAcc)
		f.add(vAcc, vRes, vRes)
		f.VMOVDQU32(vRes, addrDst.At(offset))
	}

	f.Loop(N, func() {
		// load 8 scalars (= 32 bytes) of b into the low ymm half of vb
		f.VMOVDQU32(addrB.At(0), vb.Y())
		process(0, vMask0)
		process(8, vMask1)
		process(16, vMask2)
		f.ADDQ("$192", addrA)
		f.ADDQ("$192", addrDst)
		f.ADDQ("$32", addrB)
	})

	f.RET()
}

func (_f *FFAmd64) generateDITWithTwiddlesVecE6() {
	// func vectorDITWithTwiddles_E6_avx512(a0, a1 *E6, twiddles *fr.Element, N uint64)
	//
	// Computes a1 *= twiddles followed by Butterfly(a0, a1). Precondition:
	// N % 8 == 0; processes 8 E6 (= 48 fr lanes = 3 zmm) per outer iteration.
	const argSize = 4 * 8
	stackSize := _f.StackSize(_f.NbWords*4+2, 0, 0)

	registers := _f.FnHeader("vectorDITWithTwiddles_E6_avx512", stackSize, argSize, amd64.DX, amd64.AX)
	defer _f.AssertCleanStack(stackSize, 0)
	f := &fieldHelper{FFAmd64: _f, registers: &registers}

	addrA0 := registers.Pop()
	addrA1 := registers.Pop()
	addrTwiddles := registers.Pop()
	N := registers.Pop()

	vMask0, vMask1, vMask2 := f.e6FusedMulPrologue()

	f.MOVQ("a0+0(FP)", addrA0)
	f.MOVQ("a1+8(FP)", addrA1)
	f.MOVQ("twiddles+16(FP)", addrTwiddles)
	f.MOVQ("N+24(FP)", N)
	f.SHRQ("$3", N)

	vTwiddles := registers.PopV()
	vScale := registers.PopV()
	vY := registers.PopV()
	vYScaled := registers.PopV()
	vX := registers.PopV()
	vXOut := registers.PopV()
	vYOut := registers.PopV()

	process := func(offset int, mask amd64.VectorRegister) {
		f.VMOVDQU32(addrA1.At(offset), vY)
		f.VPERMD(vTwiddles, mask, vScale)
		f.mul(vY, vScale, vYScaled, true)
		f.VMOVDQU32(addrA0.At(offset), vX)
		f.add(vX, vYScaled, vXOut)
		f.sub(vX, vYScaled, vYOut)
		f.VMOVDQU32(vXOut, addrA0.At(offset))
		f.VMOVDQU32(vYOut, addrA1.At(offset))
	}

	f.Loop(N, func() {
		f.VMOVDQU32(addrTwiddles.At(0), vTwiddles.Y())
		process(0, vMask0)
		process(8, vMask1)
		process(16, vMask2)
		f.ADDQ("$192", addrA0)
		f.ADDQ("$192", addrA1)
		f.ADDQ("$32", addrTwiddles)
	})

	f.RET()
}

func (_f *FFAmd64) generateDIFWithTwiddlesVecE6() {
	// func vectorDIFWithTwiddles_E6_avx512(a0, a1 *E6, twiddles *fr.Element, N uint64)
	//
	// Computes Butterfly(a0, a1) followed by a1 *= twiddles. Precondition:
	// N % 8 == 0; processes 8 E6 (= 48 fr lanes = 3 zmm) per outer iteration.
	const argSize = 4 * 8
	stackSize := _f.StackSize(_f.NbWords*4+2, 0, 0)

	registers := _f.FnHeader("vectorDIFWithTwiddles_E6_avx512", stackSize, argSize, amd64.DX, amd64.AX)
	defer _f.AssertCleanStack(stackSize, 0)
	f := &fieldHelper{FFAmd64: _f, registers: &registers}

	addrA0 := registers.Pop()
	addrA1 := registers.Pop()
	addrTwiddles := registers.Pop()
	N := registers.Pop()

	vMask0, vMask1, vMask2 := f.e6FusedMulPrologue()

	f.MOVQ("a0+0(FP)", addrA0)
	f.MOVQ("a1+8(FP)", addrA1)
	f.MOVQ("twiddles+16(FP)", addrTwiddles)
	f.MOVQ("N+24(FP)", N)
	f.SHRQ("$3", N)

	vTwiddles := registers.PopV()
	vScale := registers.PopV()
	vX := registers.PopV()
	vY := registers.PopV()
	vXOut := registers.PopV()
	vYOut := registers.PopV()
	vYScaled := registers.PopV()

	process := func(offset int, mask amd64.VectorRegister) {
		f.VMOVDQU32(addrA0.At(offset), vX)
		f.VMOVDQU32(addrA1.At(offset), vY)
		f.add(vX, vY, vXOut)
		f.sub(vX, vY, vYOut)
		f.VMOVDQU32(vXOut, addrA0.At(offset))
		f.VPERMD(vTwiddles, mask, vScale)
		f.mul(vYOut, vScale, vYScaled, true)
		f.VMOVDQU32(vYScaled, addrA1.At(offset))
	}

	f.Loop(N, func() {
		f.VMOVDQU32(addrTwiddles.At(0), vTwiddles.Y())
		process(0, vMask0)
		process(8, vMask1)
		process(16, vMask2)
		f.ADDQ("$192", addrA0)
		f.ADDQ("$192", addrA1)
		f.ADDQ("$32", addrTwiddles)
	})

	f.RET()
}

func (_f *FFAmd64) generateMulVecElementE4() {
	// func vectorMulByElement_avx512(res, a *E4, b *fr.Element, N uint64)

	const argSize = 4 * 8
	stackSize := _f.StackSize(_f.NbWords*4+2, 0, 0)

	registers := _f.FnHeader("vectorMulByElement_avx512", stackSize, argSize, amd64.DX, amd64.AX)
	defer _f.AssertCleanStack(stackSize, 0)
	f := &fieldHelper{FFAmd64: _f, registers: &registers}

	addrRes := registers.Pop()
	addrA := registers.Pop()
	addrB := registers.Pop()
	N := registers.Pop()

	f.loadQ()
	f.loadQInvNeg()

	f.MOVQ(uint64(0b01_01_01_01_01_01_01_01), amd64.AX)
	f.KMOVD(amd64.AX, amd64.K3)

	f.MOVQ("res+0(FP)", addrRes)
	f.MOVQ("a+8(FP)", addrA)
	f.MOVQ("b+16(FP)", addrB)
	f.MOVQ("N+24(FP)", N)

	va, vb, vRes := registers.PopV(), registers.PopV(), registers.PopV()

	// load the mask we use to duplicate the 4 uint32 of b into a zmm register
	// maskPermD = [0,0,0,0, 1,1,1,1, 2,2,2,2, 3,3,3,3]
	// so that if b = [b0,b1,b2,b3], then after permutation we get
	// [b0,b0,b0,b0, b1,b1,b1,b1, b2,b2,b2,b2, b3,b3,b3,b3]
	// which is what we need to multiply with a (4 E4 at a time)
	addrMaskPermD := registers.Pop()
	vMaskPermD := registers.PopV()
	f.MOVQ("·maskPermD+0(SB)", addrMaskPermD)
	f.VMOVDQU32(addrMaskPermD.At(0), vMaskPermD)

	// code here is very similar to vector::Mul() (base)
	// the only thing is we advance the iterators on a (on E4) and b (on Element) at
	// different speeds, and need to load b a bit differently.

	// N % 4 == 0 (pre condition checked by caller)
	// divide N by 4
	f.SHRQ("$2", N)

	f.Loop(N, func() {
		// load a
		f.VMOVDQU32(addrA.At(0), va)
		f.VMOVDQU32(addrB.At(0), vb.X()) // need only 4 of them

		// now vb has [b0, b1, b2, b3, 0, 0, ..., 0]
		// but we want [b0, b0, b0, b0, b1, b1, b1, b1, b2, b2, b2, b2, b3, b3, b3, b3]
		f.VPERMD(vb, vMaskPermD, vb)

		// now we can mul
		f.mul(va, vb, vRes, true)

		// save result
		f.VMOVDQU32(vRes, addrRes.At(0))

		// increment result by 16uint32 (4 E4)
		f.ADDQ("$64", addrRes)
		f.ADDQ("$64", addrA)
		// increment b by 16 bytes (4 base element)
		f.ADDQ("$16", addrB)
	})

	f.RET()
}

type e4VecOp int

const (
	e4VecMul e4VecOp = iota
	e4VecScalarMul
	e4VecInnerProd
)

func (_f *FFAmd64) generateMulVecE4(op e4VecOp) {
	// the code for Mul, ScalarMul and InnerProduct are very similar;
	// we load 16 E4 at a time, transpose them to ZMM vectors
	// perform the mul.
	// for scalarMul, we load the second operand only once.
	// for innerProduct, we don't transpose back but instead accumulate into other zmm registers
	// to return the sum.
	const argSize = 4 * 8
	stackSize := _f.StackSize(_f.NbWords*4+2, 0, 0)
	var name string
	switch op {
	case e4VecMul:
		name = "vectorMul_avx512"
	case e4VecScalarMul:
		name = "vectorScalarMul_avx512"
	case e4VecInnerProd:
		name = "vectorInnerProduct_avx512"
	default:
		panic("invalid E4VecOp")
	}
	registers := _f.FnHeader(name, stackSize, argSize, amd64.DX, amd64.AX)
	defer _f.AssertCleanStack(stackSize, 0)
	f := &fieldHelper{FFAmd64: _f, registers: &registers}

	addrRes := registers.Pop()
	addrA := registers.Pop()
	addrB := registers.Pop()
	N := registers.Pop()

	f.loadQ()
	f.loadQInvNeg()

	f.MOVQ(uint64(0b01_01_01_01_01_01_01_01), amd64.AX)
	f.KMOVD(amd64.AX, amd64.K3)

	f.MOVQ("res+0(FP)", addrRes)
	f.MOVQ("a+8(FP)", addrA)
	f.MOVQ("b+16(FP)", addrB)
	f.MOVQ("N+24(FP)", N)

	// transpose stuff
	maskFFFF := registers.Pop()
	addrIndexGather4 := registers.Pop()
	vIndexGather := registers.PopV()
	f.MOVQ("$0xffffffffffffffff", maskFFFF)
	f.MOVQ("·indexGather4+0(SB)", addrIndexGather4)
	f.VMOVDQU32(addrIndexGather4.At(0), vIndexGather)

	z := registers.PopVN(8)
	var accInnerProd []amd64.VectorRegister
	if op == e4VecInnerProd {
		// zeroize accInnerProd
		accInnerProd = registers.PopVN(4)
		f.VXORPS(accInnerProd[0], accInnerProd[0], accInnerProd[0])
		for i := 1; i < 4; i++ {
			f.VMOVDQA64(accInnerProd[0], accInnerProd[i])
		}
	}

	if op == e4VecScalarMul {
		// we load b only once.
		for i := 4; i < 8; i++ {
			f.MOVD(addrB.AtD(i-4), amd64.AX)
			f.VPBROADCASTD(amd64.AX, z[i])
		}
	}

	// N % 16 == 0 (pre condition checked by caller)
	// divide N by 16
	f.SHRQ("$4", N)
	f.Loop(N, func() {
		switch op {
		case e4VecMul, e4VecInnerProd:
			// first we fetch 16 E4 values from a and from b, and then we transpose them
			// into 2*4 vectors of 16 dwords
			// such that z0 == [a[0].B0.A0, a[1].B0.A0, ...]
			for i := range 4 {
				// interleave ops for better throughput
				f.KMOVD(maskFFFF, amd64.K1)
				f.KMOVD(maskFFFF, amd64.K2)
				f.VPGATHERDD(i*4, addrA, vIndexGather, 4, amd64.K1, z[i])
				f.VPGATHERDD(i*4, addrB, vIndexGather, 4, amd64.K2, z[i+4])
			}
		case e4VecScalarMul:
			// for scalar mul, we don't need to fetch b.
			for i := range 4 {
				// interleave ops for better throughput
				f.KMOVD(maskFFFF, amd64.K1)
				f.VPGATHERDD(i*4, addrA, vIndexGather, 4, amd64.K1, z[i])
			}
		}

		// note that is implementation is not overoptimized but readable as it follows
		// strictly the purego "single" E4 mul, flattened.

		// perform the mul:
		// Inline E2.Add(&x.B0, &x.B1)
		// var a0, a1 fr.Element
		// a0.Add(&zmm0, &zmm2)
		// a1.Add(&zmm1, &zmm3)
		a0, a1 := registers.PopV(), registers.PopV()
		f.add(z[0], z[2], a0)
		f.add(z[1], z[3], a1)

		// Inline E2.Add(&y.B0, &y.B1)
		// var b0, b1 fr.Element
		// b0.Add(&zmm4, &zmm6)
		// b1.Add(&zmm5, &zmm7)
		b0, b1 := registers.PopV(), registers.PopV()
		f.add(z[4], z[6], b0)
		f.add(z[5], z[7], b1)

		// Inline E2.Mul(&x.B0, &y.B0)
		// var dA0, dA1 fr.Element
		// {
		// E2.Mul(x.B0, y.B0)
		// 	var  t1, t2 fr.Element
		// 	dA1.Add(&zmm0, &zmm1)
		// 	t1.Add(&zmm4, &zmm5)
		// 	dA0.Mul(&zmm0, &zmm4)
		// 	t2.Mul(&zmm1, &zmm5)
		// 	dA1.Mul(&dA1, &t1)
		// 	dA0.Add(&dA0, &t2)
		// 	dA1.Sub(&dA1, &dA0)
		// 	dA0.Add(&dA0, &t2).Add(&dA0, &t2)
		// }
		e2Mul := func(xA0, xA1, yA0, yA1 amd64.VectorRegister) (rA0, rA1 amd64.VectorRegister) {
			rA0, rA1 = registers.PopV(), registers.PopV()

			t1, t2, t3 := registers.PopV(), registers.PopV(), registers.PopV()

			// rA1 = xA0 + xA1
			f.add(xA0, xA1, rA1)
			// t1 = yA0 + yA1
			f.addNoReduce(yA0, yA1, t1)

			// rA0 = xA0 * yA0
			// t2 = xA1 * yA1
			f.mul(xA0, yA0, rA0, true)
			f.mul(xA1, yA1, t2, true)

			// rA1 = rA1 * t1
			f.mul(rA1, t1, rA1, true)

			// TODO @gbotrel: here this work only for koalabear (mulByNonResidue == 3)
			// update this part with the correct logic for other fields if needed

			// t3 = rA0 + t2
			f.add(rA0, t2, t3)
			// rA1 = rA1 - t3
			f.sub(rA1, t3, rA1)

			// rA0 = t3 + t2 + t2
			f.add(t3, t2, rA0)
			f.add(rA0, t2, rA0)

			registers.PushV(t1, t2, t3)

			return
		}
		dA0, dA1 := e2Mul(z[0], z[1], z[4], z[5])

		// Inline E2.Mul(&x.B1, &y.B1)
		cA0, cA1 := e2Mul(z[2], z[3], z[6], z[7])

		// Inline E2.Mul(&a, &b)
		aMbA0, aMbA1 := e2Mul(a0, a1, b0, b1)

		// Inline E2.Add(&d, &c)
		// var bcA0, bcA1 fr.Element
		// bcA0.Add(&dA0, &cA0)
		// bcA1.Add(&dA1, &cA1)
		bcA0, bcA1 := registers.PopV(), registers.PopV()
		f.add(dA0, cA0, bcA0)
		f.add(dA1, cA1, bcA1)

		// Inline E2.Add(&a, &b)
		// var abA0, abA1 fr.Element
		// abA0.Add(&a0, &b0)
		// abA1.Add(&a1, &b1)
		abA0, abA1 := registers.PopV(), registers.PopV()
		f.add(a0, b0, abA0)
		f.add(a1, b1, abA1)

		// let's use z[0:4] for z result, we don't need them anymore.
		// z.B1 = a - (d + c)
		// z.B1.A0.Sub(&aMbA0, &bcA0)
		// z.B1.A1.Sub(&aMbA1, &bcA1)
		f.sub(aMbA0, bcA0, z[2])
		f.sub(aMbA1, bcA1, z[3])

		// z.B0 = MulByNonResidue(c) + d
		// MulByNonResidue: (A0, A1) -> (A1*3, A0)
		// fr.MulBy3(&cA1)
		// z.B0.A0.Add(&cA1, &dA0)
		// z.B0.A1.Add(&cA0, &dA1)
		f.add(cA1, cA1, a0)  // 2x
		f.add(cA1, dA0, dA0) // 1x
		f.add(cA0, dA1, z[1])
		f.add(a0, dA0, z[0]) // 3x

		switch op {
		case e4VecMul, e4VecScalarMul:
			// transpose result back
			for i := range z[:4] {
				f.KMOVD(maskFFFF, amd64.K1)
				f.VPSCATTERDD(i*4, addrRes, vIndexGather, 4, amd64.K1, z[i])
			}

			// increment result by 16*4uint32 (16*E4)
			f.ADDQ("$256", addrRes)
			f.ADDQ("$256", addrA)
			if op == e4VecMul {
				f.ADDQ("$256", addrB)
			}
		case e4VecInnerProd:
			// here we have 4 zmm vectors that we need to accumulate;
			// we use z[4:7] as temps;
			// z[0] contains the result of the product r[0].B0.A0, r[1].B0.A0, ...
			// z[1] --> r[0].B0.A1, r[1].B0.A1, ...
			// so the idea is we accumulate these coordinates in accInnerProd
			// (in mul and scalar mul we transpose the result back, here we skip this step.)
			// and let the caller reduce mod q.
			for i := range 4 {
				f.VEXTRACTI64X4(1, z[i], z[i+4].Y())
				f.add(z[i], z[i+4], z[i+4], fY)
				f.VEXTRACTI64X2(1, z[i+4].Y(), z[i].X())
				f.VPADDD(z[i+4].X(), z[i].X(), z[i].X())
				f.VPMOVZXDQ(z[i].X(), z[i+4].Y())
				f.VPADDQ(z[i+4].Y(), accInnerProd[i].Y(), accInnerProd[i].Y())
			}

			f.ADDQ("$256", addrA)
			f.ADDQ("$256", addrB)
		}
	})

	if op == e4VecInnerProd {
		// we have 4 * 8 uint64 to store in result from the accumulators;
		f.VMOVDQU64(accInnerProd[0].Y(), addrRes.At(0))
		f.VMOVDQU64(accInnerProd[1].Y(), addrRes.At(8))
		f.VMOVDQU64(accInnerProd[2].Y(), addrRes.At(16))
		f.VMOVDQU64(accInnerProd[3].Y(), addrRes.At(24))
	}

	f.RET()
}

func (_f *FFAmd64) generateInnerProdByElementE4() {
	// func vectorInnerProductByElement_avx512(res, a *E4, b *fr.Element, N uint64)

	const argSize = 4 * 8
	stackSize := _f.StackSize(_f.NbWords*4+2, 0, 0)
	registers := _f.FnHeader("vectorInnerProductByElement_avx512", stackSize, argSize, amd64.DX, amd64.AX)
	defer _f.AssertCleanStack(stackSize, 0)

	f := &fieldHelper{FFAmd64: _f, registers: &registers}

	addrRes := registers.Pop()
	addrA := registers.Pop()
	addrB := registers.Pop()
	N := registers.Pop()

	va, vb, vRes, vAcc := registers.PopV(), registers.PopV(), registers.PopV(), registers.PopV()

	f.loadQ()
	f.loadQInvNeg()

	f.MOVQ(uint64(0b01_01_01_01_01_01_01_01), amd64.AX)
	f.KMOVD(amd64.AX, amd64.K3)

	f.MOVQ("res+0(FP)", addrRes)
	f.MOVQ("a+8(FP)", addrA)
	f.MOVQ("b+16(FP)", addrB)
	f.MOVQ("N+24(FP)", N)

	// N % 4 == 0 (pre condition checked by caller)
	// divide N by 4
	f.SHRQ("$2", N)

	// code here is similar to MulVecElementE4, see this for comments
	// we advance the iterators on a (on E4) and b (on Element) at
	// different speeds, and need to load b a bit differently.
	addrMaskPermD := registers.Pop()
	vMaskPermD := registers.PopV()
	f.MOVQ("·maskPermD+0(SB)", addrMaskPermD)
	f.VMOVDQU32(addrMaskPermD.At(0), vMaskPermD)

	// zero-out vAcc
	f.VXORPS(vAcc, vAcc, vAcc)

	f.Loop(N, func() {
		// load a
		f.VMOVDQU32(addrA.At(0), va)
		f.VMOVDQU32(addrB.At(0), vb.X()) // need only 4 of them

		// now vb has [b0, b1, b2, b3, 0, 0, ..., 0]
		// but we want [b0, b0, b0, b0, b1, b1, b1, b1, b2, b2, b2, b2, b3, b3, b3, b3]
		f.VPERMD(vb, vMaskPermD, vb)

		// now we can mul
		// note that we could avoid reduction here, and maintain 2 accumulators and do a lazy
		// reduction after the loop, but in practice this does not seem to help perf.
		f.mul(va, vb, vRes, true)

		// accumulate
		f.add(vRes, vAcc, vAcc)

		// increment a by 4*4 uint32 (4 * E4)
		f.ADDQ("$64", addrA)
		// increment b by 4 uint32 (4 base element)
		f.ADDQ("$16", addrB)
	})

	// now the result is in vAcc, with 4 E4. We want to reduce that to 1 E4.
	// that is vAcc contains [a0, a1, a2, a3, b0, b1, b2, b3, c0, c1, c2, c3, d0, d1, d2, d3]
	// and we want [a0+b0+c0+d0, a1+b1+c1+d1, a2+b2+c2+d2, a3+b3+c3+d3]
	v0 := registers.PopV()
	f.VEXTRACTI64X4(1, vAcc, v0.Y())
	f.add(vAcc, v0, v0, fY)
	f.VEXTRACTI64X2(1, v0.Y(), vAcc.X())
	f.add(v0, vAcc, vAcc, fX)

	// store result
	f.VMOVDQU32(vAcc.X(), addrRes.At(0))

	f.RET()

}

func (_f *FFAmd64) generateMulAccByElement() {
	// func mulAccByElement_avx512(alpha *E4, scale *fr.Element, res *E4, N uint64)

	const argSize = 4 * 8
	stackSize := _f.StackSize(_f.NbWords*4+2, 0, 0)
	registers := _f.FnHeader("mulAccByElement_avx512", stackSize, argSize, amd64.DX, amd64.AX)
	defer _f.AssertCleanStack(stackSize, 0)

	f := &fieldHelper{FFAmd64: _f, registers: &registers}

	f.loadQ()
	f.loadQInvNeg()

	addrAlpha := registers.Pop()
	addrScale := registers.Pop()
	addrRes := registers.Pop()
	N := registers.Pop()

	// prepare the mask used for the merging mul results
	f.MOVQ(uint64(0b01_01_01_01_01_01_01_01), amd64.AX)
	f.KMOVD(amd64.AX, amd64.K3)

	f.MOVQ("alpha+0(FP)", addrAlpha)
	f.MOVQ("scale+8(FP)", addrScale)
	f.MOVQ("res+16(FP)", addrRes)
	f.MOVQ("N+24(FP)", N)

	// var tmp E4
	// for i := 0; i < N; i++ {
	// 	tmp.MulByElement(alpha, &scale[i])
	// 	res[i].Add(&res[i], &tmp)
	// }

	// alpha is an E4, so it is 4 uint32
	// we load it into a XMM register and repeat it to a YMM, then to a ZMM
	alpha := registers.PopV()
	alphaOdd := registers.PopV()
	result := registers.PopV()
	s0 := registers.PopV()
	s1 := registers.PopV()
	s2 := registers.PopV()
	s3 := registers.PopV()
	acc := registers.PopV()

	f.VMOVDQU32(addrAlpha.At(0), alpha.X())
	f.VINSERTI64X2(1, alpha.X(), alpha.Y(), alpha.Y())
	f.VINSERTI64X4(1, alpha.Y(), alpha.Z(), alpha.Z())

	f.VPSRLQ("$32", alpha, alphaOdd) // keep high 32 bits

	// N % 4 == 0 (pre condition checked by caller)
	// divide N by 4
	f.SHRQ("$2", N)

	f.Loop(N, func() {

		// load result
		f.VMOVDQU32(addrRes.At(0), result)

		// load scale
		f.VPBROADCASTD(addrScale.AtD(0), s0.X())
		f.VPBROADCASTD(addrScale.AtD(1), s1.X())
		f.VPBROADCASTD(addrScale.AtD(2), s2.X())
		f.VPBROADCASTD(addrScale.AtD(3), s3.X())

		f.VINSERTI64X2(1, s1.X(), s0.Y(), s0.Y())
		f.VINSERTI64X2(1, s3.X(), s2.Y(), s2.Y())
		f.VINSERTI64X4(1, s2.Y(), s0.Z(), s0.Z())

		// computes c = a * b mod q
		// a and b can be in [0, 2q)
		f.mul(s0, alpha, acc, true)

		f.add(result, acc, result)

		// save result
		f.VMOVDQU32(result, addrRes.At(0))

		// increment result by 16uint32
		f.ADDQ("$64", addrRes)
		f.ADDQ("$16", addrScale)
	})

	f.RET()
}
