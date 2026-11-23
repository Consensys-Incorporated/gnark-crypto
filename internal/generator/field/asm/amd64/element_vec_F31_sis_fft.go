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

func (f *FFAmd64) generateFFTDefinesF31() {

	// computes a = a + b and b = a - b,
	// leaves a in [0, q)
	// leaves b in [0,q)
	_ = f.Define("butterflyD1Q", 5, func(args ...any) {
		x := args[0]
		y := args[1]
		qd := args[2]
		b0 := args[3]
		b1 := args[4]
		f.VPADDD(x, y, b0)
		f.VPSUBD(y, x, y)
		f.VPSUBD(qd, b0, x)
		f.VPMINUD(b0, x, x)
		f.VPADDD(qd, y, b1)
		f.VPMINUD(b1, y, y)
	}, true)

	// computes a = a + b and b = a - b,
	// leaves a in [0, q)
	// leaves b in [0,2q)
	butterflyD2Q := f.Define("butterflyD2Q", 5, func(args ...any) {
		x := args[0]
		y := args[1]
		qd := args[2]
		b0 := args[3]
		b1 := args[4]
		f.VPSUBD(y, x, b1)
		f.VPADDD(x, y, b0)
		f.VPADDD(qd, b1, y)
		f.VPSUBD(qd, b0, x)
		f.VPMINUD(b0, x, x)
	}, true)

	// computes a = a + b and b = a - b,
	// leaves a in [0,2q)
	// leaves b in [0,2q)
	_ = f.Define("butterflyD2Q2Q", 4, func(args ...any) {
		x := args[0]
		y := args[1]
		qd := args[2]
		b0 := args[3]
		f.VPSUBD(y, x, b0)
		f.VPADDD(x, y, x)
		f.VPADDD(qd, b0, y)
	}, true)

	// computes a = a * b mod q
	// a and b can be in [0, 2q)
	mulD := f.Define("mulD", 10, func(args ...any) {
		a := args[0]
		b := args[1]
		aOdd := args[2]
		bOdd := args[3]
		b0 := args[4]
		b1 := args[5]
		PL0 := args[6]
		PL1 := args[7]
		q := args[8]
		qInvNeg := args[9]

		f.VPSRLQ("$32", a, aOdd) // keep high 32 bits
		f.VPSRLQ("$32", b, bOdd) // keep high 32 bits

		// VPMULUDQ conveniently ignores the high 32 bits of each QWORD lane
		f.VPMULUDQ(a, b, b0)
		f.VPMULUDQ(aOdd, bOdd, b1)
		f.VPMULUDQ(b0, qInvNeg, PL0)
		f.VPMULUDQ(b1, qInvNeg, PL1)

		f.VPMULUDQ(PL0, q, PL0)
		f.VPADDQ(b0, PL0, b0)

		f.VPMULUDQ(PL1, q, PL1)
		f.VPADDQ(b1, PL1, b1)

		f.VMOVSHDUPk(b0, amd64.K3, b1)

		f.VPSUBD(q, b1, PL1)
		f.VPMINUD(b1, PL1, a)
	}, true)

	// goes from
	// in0 = [a0 a1 a2 a3 a4 a5 a6 a7 a8 a9 a10 a11 a12 a13 a14 a15]
	// in1 = [b0 b1 b2 b3 b4 b5 b6 b7 b8 b9 b10 b11 b12 b13 b14 b15]
	// to
	// in0 = [a0 a1 a2 a3 a4 a5 a6 a7 b0 b1 b2 b3 b4 b5 b6 b7]
	// in1 = [a8 a9 a10 a11 a12 a13 a14 a15 b8 b9 b10 b11 b12 b13 b14 b15]
	_ = f.Define("permute8x8", 3, func(args ...any) {
		x := args[0]
		y := args[1]
		b0 := args[2]
		f.VSHUFI64X2(uint64(0b01_00_11_10), y, x, b0)
		f.VPBLENDMQ(x, b0, x, amd64.K1)
		f.VPBLENDMQ(b0, y, y, amd64.K1)
	}, true)

	// goes from
	// in0 = [a0 a1 a2 a3 a4 a5 a6 a7 a8 a9 a10 a11 a12 a13 a14 a15]
	// in1 = [b0 b1 b2 b3 b4 b5 b6 b7 b8 b9 b10 b11 b12 b13 b14 b15]
	// to
	// in0 = [a0 a1 a2 a3 b0 b1 b2 b3 a8 a9 a10 a11 b8 b9 b10 b11]
	// in1 = [a4 a5 a6 a7 b4 b5 b6 b7 a12 a13 a14 a15 b12 b13 b14 b15]
	_ = f.Define("permute4x4", 4, func(args ...any) {
		x := args[0]
		y := args[1]
		vInterleaveIndices := args[2]
		tmp := args[3]
		f.VMOVDQA64(vInterleaveIndices, tmp)
		f.VPERMI2Q(y, x, tmp)
		f.VPBLENDMQ(x, tmp, x, amd64.K2)
		f.VPBLENDMQ(tmp, y, y, amd64.K2)
	}, true)

	// goes from
	// in0 = [a0 a1 a2 a3 a4 a5 a6 a7 a8 a9 a10 a11 a12 a13 a14 a15]
	// in1 = [b0 b1 b2 b3 b4 b5 b6 b7 b8 b9 b10 b11 b12 b13 b14 b15]
	// to
	// in0 = [a0 a1 b0 b1 a4 a5 b4 b5 a8 a9 b8 b9 a12 a13 b12 b13]
	// in1 = [a2 a3 b2 b3 a6 a7 b6 b7 a10 a11 b10 b11 a14 a15 b14 b15]
	_ = f.Define("permute2x2", 3, func(args ...any) {
		x := args[0]
		y := args[1]
		b0 := args[2]

		f.VSHUFPD(0b01010101, y, x, b0)
		f.VPBLENDMQ(x, b0, x, amd64.K3)
		f.VPBLENDMQ(b0, y, y, amd64.K3)
	}, true)

	// goes from
	// in0 = [a0 a1 a2 a3 a4 a5 a6 a7 a8 a9 a10 a11 a12 a13 a14 a15]
	// in1 = [b0 b1 b2 b3 b4 b5 b6 b7 b8 b9 b10 b11 b12 b13 b14 b15]
	// to
	// in0 = [a0 b0 a2 b2 a4 b4 a6 b6 a8 b8 a10 b10 a12 b12 a14 b14]
	// in1 = [a1 b1 a3 b3 a5 b5 a7 b7 a9 b9 a11 b11 a13 b13 a15 b15]
	_ = f.Define("permute1x1", 3, func(args ...any) {
		x := args[0]
		y := args[1]
		b0 := args[2]

		f.VPSHRDQ("$32", y, x, b0)
		f.VPBLENDMD(x, b0, x, amd64.K3)
		f.VPBLENDMD(b0, y, y, amd64.K3)
	}, true)

	_ = f.Define("load_q", 2, func(args ...any) {
		q := args[0]
		qInv := args[1]
		f.MOVD("$const_q", amd64.AX)
		f.VPBROADCASTD(amd64.AX, q)
		f.MOVD("$const_qInvNeg", amd64.AX)
		f.VPBROADCASTD(amd64.AX, qInv)
	}, true)

	// these masks are used in the permuteNxN functions
	_ = f.Define("load_masks", 0, func(_ ...any) {
		f.MOVQ(uint64(0b0000_1111_0000_1111), amd64.AX)
		f.KMOVQ(amd64.AX, amd64.K1)

		f.MOVQ(uint64(0b00_11_00_11), amd64.AX)
		f.KMOVQ(amd64.AX, amd64.K2)

		f.MOVQ(uint64(0b01_01_01_01_01_01_01_01), amd64.AX)
		f.KMOVD(amd64.AX, amd64.K3)
	}, true)

	_ = f.Define("butterfly_mulD", 11+4, func(args ...any) {
		butterflyD2Q(args[0], args[1], args[2], args[3], args[4])
		mulD(args[5:]...)
	}, true)
}

func (_f *FFAmd64) generateFFTInnerDITF31() {
	f := &fftHelper{_f}

	// func innerDITWithTwiddles(a []Element, twiddles []Element, start, end, m int) {
	// 	for i := start; i < end; i++ {
	// 		a[i+m].Mul(&a[i+m], &twiddles[i])
	// 		Butterfly(&a[i], &a[i+m])
	// 	}
	// }
	const argSize = 5 * 8
	stackSize := f.StackSize(f.NbWords*2+4, 1, 0)

	registers := f.FnHeader("innerDITWithTwiddles_avx512", stackSize, argSize, amd64.AX)
	defer f.AssertCleanStack(stackSize, 0)

	f.Comment("refer to the code generator for comments and documentation.")

	addrA := registers.Pop()
	addrTwiddles := registers.Pop()
	len := registers.Pop()
	m := registers.Pop()
	addrAPlusM := registers.Pop()

	a := registers.PopV()
	am := registers.PopV()
	b0 := registers.PopV()
	b1 := registers.PopV()
	qd := registers.PopV()
	qInvNeg := registers.PopV()
	t0 := registers.PopV()
	aOdd := registers.PopV()
	bOdd := registers.PopV()
	PL0 := registers.PopV()
	PL1 := registers.PopV()

	f.loadQ(qd, qInvNeg)
	f.loadMasks()

	// load arguments
	f.MOVQ("a+0(FP)", addrA)
	f.MOVQ("twiddles+8(FP)", addrTwiddles)
	f.MOVQ("end+24(FP)", len)
	f.MOVQ("m+32(FP)", m)

	// we do only m >= 16;
	// that ensures we can process 16 elements at a time
	// avoid bound checks and edge cases
	f.SHRQ("$4", len, "we are processing 16 elements at a time")

	// offset we want to add to a is m*4bytes
	f.SHLQ("$2", m, "offset = m * 4bytes")

	f.MOVQ(addrA, addrAPlusM)
	f.ADDQ(m, addrAPlusM)

	f.Loop(len, func() {
		f.VMOVDQU32(addrA.At(0), a, "load a[i]")
		f.VMOVDQU32(addrAPlusM.At(0), am, "load a[i+m]")
		f.VMOVDQU32(addrTwiddles.At(0), t0, "load twiddles[i]")

		f.mulD(am, t0, aOdd, bOdd, b0, b1, PL0, PL1, qd, qInvNeg)
		f.butterflyD1Q(a, am, qd, b0, b1)

		// a is ready to be stored, but we need to scale am by twiddles.
		f.VMOVDQU32(a, addrA.At(0), "store a[i]")
		f.VMOVDQU32(am, addrAPlusM.At(0), "store a[i+m]")

		f.ADDQ("$64", addrA)
		f.ADDQ("$64", addrAPlusM)
		f.ADDQ("$64", addrTwiddles)
	})

	f.RETAVX()
}

func (_f *FFAmd64) generateFFTInnerDIFF31() {
	f := &fftHelper{_f}

	// func innerDIFWithTwiddles_avx512(a, twiddles *{{ .FF }}.Element, start, end, m int)
	// 	for i := start; i < end; i++ {
	// 		Butterfly(&a[i], &a[i+m])
	// 		a[i+m].Mul(&a[i+m], &twiddles[i])
	// 	}
	// }
	const argSize = 5 * 8
	stackSize := f.StackSize(f.NbWords*2+4, 1, 0)

	registers := f.FnHeader("innerDIFWithTwiddles_avx512", stackSize, argSize, amd64.AX)
	defer f.AssertCleanStack(stackSize, 0)
	f.Comment("refer to the code generator for comments and documentation.")

	addrA := registers.Pop()
	addrTwiddles := registers.Pop()
	len := registers.Pop()
	m := registers.Pop()
	addrAPlusM := registers.Pop()

	a := registers.PopV()
	am := registers.PopV()
	qd := registers.PopV()
	b0 := registers.PopV()
	qInvNeg := registers.PopV()
	t0 := registers.PopV()

	aOdd := registers.PopV()
	bOdd := registers.PopV()
	b1 := registers.PopV()
	PL0 := registers.PopV()
	PL1 := registers.PopV()

	// load arguments
	f.MOVQ("a+0(FP)", addrA)
	f.MOVQ("twiddles+8(FP)", addrTwiddles)
	f.MOVQ("end+24(FP)", len)
	f.MOVQ("m+32(FP)", m)

	f.loadQ(qd, qInvNeg)
	f.loadMasks()

	// we do only m >= 16;
	// that ensures we can process 16 elements at a time
	// avoid bound checks, and edge cases

	// offset we want to add to a is m*4bytes
	f.SHLQ("$2", m, "offset = m * 4bytes")

	f.MOVQ(addrA, addrAPlusM)
	f.ADDQ(m, addrAPlusM)

	// we are processing elements 16x16 so we divide len by 16
	f.SHRQ("$4", len, "we are processing 16 elements at a time")

	f.Loop(len, func() {
		f.VMOVDQU32(addrA.At(0), a, "load a[i]")
		f.VMOVDQU32(addrAPlusM.At(0), am, "load a[i+m]")

		f.VMOVDQU32(addrTwiddles.At(0), t0, "load twiddles[i]")
		f.butterfly_mulD(a, am, qd, b0, b1,
			am, t0, aOdd, bOdd, b0, b1, PL0, PL1, qd, qInvNeg)

		f.VMOVDQU32(a, addrA.At(0), "store a[i]")
		f.VMOVDQU32(am, addrAPlusM.At(0))

		f.ADDQ("$64", addrA)
		f.ADDQ("$64", addrAPlusM)
		f.ADDQ("$64", addrTwiddles)
	})

	f.RETAVX()
}

func (_f *FFAmd64) generateFFTKernelF31(klog2 int, dif bool) {
	f := &fftHelper{_f}
	if klog2 > 8 || klog2 < 7 {
		panic("not implemented")
	}
	n := 1 << klog2
	name := fmt.Sprintf("kerDIFNP_%d_avx512", n)
	if !dif {
		name = fmt.Sprintf("kerDITNP_%d_avx512", n)
	}
	const argSize = 7 * 8
	stackSize := f.StackSize(f.NbWords*2+4, 1, 0)
	registers := f.FnHeader(name, stackSize, argSize, amd64.AX)
	defer f.AssertCleanStack(stackSize, 0)

	f.Comment("refer to the code generator for comments and documentation.")

	// registers & labels we need
	addrA := registers.Pop()
	addrAPlusM := registers.Pop()
	addrTwiddlesRoot := registers.Pop()
	addrTwiddles := registers.Pop()
	innerLen := registers.Pop()

	// AVX512 registers
	a := registers.PopVN(n / 16)
	qd := registers.PopV()
	qInvNeg := registers.PopV()

	f.loadQ(qd, qInvNeg)
	f.loadMasks()

	f.Comment("load arguments")
	f.MOVQ("a+0(FP)", addrA)
	f.MOVQ("twiddles+24(FP)", addrTwiddlesRoot)
	f.MOVQ("stage+48(FP)", amd64.AX)
	f.IMULQ("$24", amd64.AX)
	f.ADDQ(amd64.AX, addrTwiddlesRoot, "we want twiddles[stage] as starting point")

	// we load the full input in avx512 registers
	// (16x uint32 per register)
	for i := range a {
		f.VMOVDQU32(addrA.AtD(i*16), a[i], fmt.Sprintf("load a[%d]", i))
	}

	if dif {
		vInterleaveIndices := f.generateCoreDIFKernel(n, &registers, addrTwiddlesRoot, a, qd, qInvNeg, true)

		b0 := registers.PopV()

		for i := 0; i < len(a); i += 2 {
			// ok let's say now each pair of vector v0 v1
			// such that
			// v0 = [a0 a2 a4 a6 a8 a10 a12 a14 | b0 b2 b4 b6 b8 b10 b12 b14]
			// v1 = [a1 a3 a5 a7 a9 a11 a13 a15 | b1 b3 b5 b7 b9 b11 b13 b15]

			// we need to repack them; let's do it the naive way for now
			f.VPUNPCKLDQ(a[i+1], a[i], b0)
			f.VPUNPCKHDQ(a[i+1], a[i], a[i+1])
			f.VMOVDQA32(b0, a[i])
			f.permute4x4(a[i], a[i+1], vInterleaveIndices, b0)
			f.permute8x8(a[i], a[i+1], b0)

			// store the result
			f.VMOVDQU32(a[i], addrA.AtD(i*16))
			f.VMOVDQU32(a[i+1], addrA.AtD((i+1)*16))
		}
	} else {
		f.generateCoreDITKernel(n, &registers, addrTwiddlesRoot, a, qd, qInvNeg, true, klog2-2)
		for i := 0; i < len(a); i += 2 {
			// store the result
			f.VMOVDQU32(a[i], addrA.AtD(i*16))
			f.VMOVDQU32(a[i+1], addrA.AtD((i+1)*16))
		}
	}

	f.RETAVX()

	f.Push(&registers, addrA, addrTwiddles, addrAPlusM, innerLen)

}
func (f *fftHelper) generateCoreDIFKernel(n int, registers *amd64.Registers, addrTwiddlesRoot amd64.Register, a []amd64.VectorRegister, qd, qInvNeg amd64.VectorRegister, reduceModQ bool) amd64.VectorRegister {
	t := registers.PopVN(n / 32)
	b0 := registers.PopV()
	b1 := registers.PopV()
	aOdd := registers.PopV()
	bOdd := registers.PopV()
	PL0 := registers.PopV()
	PL1 := registers.PopV()

	addrTwiddles := registers.Pop()

	m := n >> 1
	for m >= 16 {
		// The purego code is roughly:
		// for offset := 0; offset < 256; offset += X {
		// 	innerDIFWithTwiddles(a[offset:offset+X], twiddles[stage+T], 0, m, m)
		// }
		// here we have m >= 16 so we can process 16 elements at a time, filling our ZMM registers
		// of 16 DWORDS.

		// load twiddles
		f.MOVQ(addrTwiddlesRoot.At(0), addrTwiddles)
		nbTwiddles := m
		for i := 0; i < nbTwiddles/16; i++ {
			f.VMOVDQU32(addrTwiddles.AtD(i*16), t[i])
		}
		am := m / 16
		for i := 0; i < len(a); i += n / 16 {
			for j := range am {
				f.butterflyD2Q(a[i+j], a[i+j+am], qd, PL1, b1)
			}
		}
		for i := 0; i < len(a); i += n / 16 {
			for j := range am {
				f.mulD(a[i+j+am], t[j], aOdd, bOdd, b0, b1, PL0, PL1, qd, qInvNeg)
			}
		}

		n >>= 1
		m = n >> 1

		// increment addrTwiddlesRoot to go to the next stage
		f.ADDQ("$24", addrTwiddlesRoot)
	}

	for i := 4; i < len(t); i++ {
		registers.PushV(t[i])
	}

	// here we have m == 8; we can't process
	// vector of 16 elements anymore without some interleaving

	addrVInterleaveIndices := f.Pop(registers)
	vInterleaveIndices := registers.PopV()
	f.MOVQ("·vInterleaveIndices+0(SB)", addrVInterleaveIndices)
	f.VMOVDQU64(addrVInterleaveIndices.At(0), vInterleaveIndices)

	// prepare the twiddles for each stage
	// m == 8
	// we have 8 uint32 twiddles
	// --> we load in the Y part and duplicate in the rest of the ZMM register
	f.MOVQ(addrTwiddlesRoot.At(0), addrTwiddles)
	f.VMOVDQU32(addrTwiddles.At(0), t[0].Y())
	f.VINSERTI64X4(1, t[0].Y(), t[0], t[0])

	// m == 4
	// we have 4 uint32 twiddles
	f.MOVQ(addrTwiddlesRoot.At(3), addrTwiddles)
	f.VMOVDQU32(addrTwiddles.At(0), t[1].X())
	f.VINSERTI64X2(1, t[1].X(), t[1], t[1])
	f.VINSERTI64X2(2, t[1].X(), t[1], t[1])
	f.VINSERTI64X2(3, t[1].X(), t[1], t[1])

	// m == 2
	// we have 2 uint32 twiddles
	f.MOVQ(addrTwiddlesRoot.At(6), addrTwiddles)
	f.VPBROADCASTD(addrTwiddles.AtD(0), t[2])
	f.VPBROADCASTD(addrTwiddles.AtD(1), t[3])
	f.VPBLENDMD(t[2], t[3], t[2], amd64.K3)

	for j := range 3 {
		for i := 0; i < len(a); i += 2 {
			switch j {
			case 0:
				// m == 8
				f.permute8x8(a[i], a[i+1], b0)
			case 1:
				// m == 4
				f.permute4x4(a[i], a[i+1], vInterleaveIndices, b0)
			case 2:
				// m == 2
				f.permute2x2(a[i], a[i+1], b0)
			}

			f.butterflyD2Q(a[i], a[i+1], qd, PL1, b1)
		}
		for i := 0; i < len(a); i += 2 {
			// perf note:
			// we can optimize a bit further here by having a
			// mulD version that takes b and bOdd as fixed inputs;
			// will save couple of high bit extraction
			f.mulD(a[i+1], t[j], aOdd, bOdd, b0, b1, PL0, PL1, qd, qInvNeg)
			if j == 2 {
				// m == 1 --> we do the permute here it gets us better
				// throughput.
				f.permute1x1(a[i], a[i+1], b0)
			}
		}
	}

	// we already did the permute1x1 above; for the last stage we just do a butterfly
	// (twiddles[lastStage][0] == 1)
	for i := 0; i < len(a); i += 2 {
		if reduceModQ {
			// the last butterfly we reduce everything in [0, q)
			f.butterflyD1Q(a[i], a[i+1], qd, b0, b1)
		} else {
			// the last butterfly we reduce everything in [0, 2q)
			// this is useful in SIS when we scale the result afterwards;
			// the last mul will reduce in [0, q)
			f.butterflyD2Q2Q(a[i], a[i+1], qd, b0)
		}
	}

	for i := range 4 {
		registers.PushV(t[i])
	}
	registers.PushV(b0, b1, aOdd, bOdd, PL0, PL1)
	registers.Push(addrVInterleaveIndices, addrTwiddles)

	return vInterleaveIndices
}

func (f *fftHelper) generateCoreDITKernel(n int, registers *amd64.Registers, addrTwiddlesRoot amd64.Register, a []amd64.VectorRegister, qd, qInvNeg amd64.VectorRegister, reduceModQ bool, startStage int) {
	// perf note: this is less optimized than the DIF one and unrolled a bit naively.
	// not on a hot path at the moment.
	// see generateCoreDIFKernel for more comments, it is the same code but in reverse order.
	t := registers.PopVN(4)
	b0 := registers.PopV()
	b1 := registers.PopV()
	aOdd := registers.PopV()
	bOdd := registers.PopV()
	PL0 := registers.PopV()
	PL1 := registers.PopV()

	addrTwiddles := registers.Pop()

	addrVInterleaveIndices := f.Pop(registers)
	vInterleaveIndices := registers.PopV()
	f.MOVQ("·vInterleaveIndices+0(SB)", addrVInterleaveIndices)
	f.VMOVDQU64(addrVInterleaveIndices.At(0), vInterleaveIndices)

	// m == 1
	for i := 0; i < len(a); i += 2 {
		// m == 1
		f.permute1x1(a[i], a[i+1], b0)
		f.butterflyD1Q(a[i], a[i+1], qd, b0, b1)
		f.permute1x1(a[i], a[i+1], b0)
	}

	// m == 2
	f.MOVQ(startStage, amd64.AX)
	f.IMULQ("$24", amd64.AX)
	f.ADDQ(amd64.AX, addrTwiddlesRoot)
	f.MOVQ(addrTwiddlesRoot.At(0), addrTwiddles)
	f.VPBROADCASTD(addrTwiddles.AtD(0), t[2])
	f.VPBROADCASTD(addrTwiddles.AtD(1), t[3])
	f.VPBLENDMD(t[2], t[3], t[2], amd64.K3)
	for i := 0; i < len(a); i += 2 {
		f.permute2x2(a[i], a[i+1], b0)
		f.mulD(a[i+1], t[2], aOdd, bOdd, b0, b1, PL0, PL1, qd, qInvNeg)
		f.butterflyD1Q(a[i], a[i+1], qd, b0, b1)
		f.permute2x2(a[i], a[i+1], b0)
	}

	// m == 4
	f.SUBQ("$24", addrTwiddlesRoot)
	f.MOVQ(addrTwiddlesRoot.At(0), addrTwiddles)
	f.VMOVDQU32(addrTwiddles.At(0), t[1].X())
	f.VINSERTI64X2(1, t[1].X(), t[1], t[1])
	f.VINSERTI64X2(2, t[1].X(), t[1], t[1])
	f.VINSERTI64X2(3, t[1].X(), t[1], t[1])
	for i := 0; i < len(a); i += 2 {
		f.permute4x4(a[i], a[i+1], vInterleaveIndices, b0)
		f.mulD(a[i+1], t[1], aOdd, bOdd, b0, b1, PL0, PL1, qd, qInvNeg)
		f.butterflyD1Q(a[i], a[i+1], qd, b0, b1)
		f.permute4x4(a[i], a[i+1], vInterleaveIndices, b0)
	}

	// m == 8
	f.SUBQ("$24", addrTwiddlesRoot)
	f.MOVQ(addrTwiddlesRoot.At(0), addrTwiddles)
	f.VMOVDQU32(addrTwiddles.At(0), t[0].Y())
	f.VINSERTI64X4(1, t[0].Y(), t[0], t[0])
	for i := 0; i < len(a); i += 2 {
		f.permute8x8(a[i], a[i+1], b0)
		f.mulD(a[i+1], t[0], aOdd, bOdd, b0, b1, PL0, PL1, qd, qInvNeg)
		f.butterflyD1Q(a[i], a[i+1], qd, b0, b1)
		f.permute8x8(a[i], a[i+1], b0)
	}

	registers.PushV(t...)
	registers.PushV(vInterleaveIndices)

	t = registers.PopVN(n / 32)

	targetM := n >> 1
	kk := n
	m := 16
	_n := 32

	for m <= targetM {
		// increment addrTwiddlesRoot
		f.SUBQ("$24", addrTwiddlesRoot)
		f.MOVQ(addrTwiddlesRoot.At(0), addrTwiddles)
		nbTwiddles := m
		for i := 0; i < nbTwiddles/16; i++ {
			f.VMOVDQU32(addrTwiddles.AtD(i*16), t[i])
		}
		am := m / 16
		for offset := 0; offset < kk; offset += _n {
			// for offset := 0; offset < 128; offset += n {
			aa := a[offset/16:]
			for i := range am {
				f.mulD(aa[i+am], t[i], aOdd, bOdd, b0, b1, PL0, PL1, qd, qInvNeg)
				f.butterflyD1Q(aa[i], aa[i+am], qd, b0, b1)
			}
		}

		_n <<= 1
		m = _n >> 1

	}

	for i := range t {
		registers.PushV(t[i])
	}
	registers.PushV(b0, b1, aOdd, bOdd, PL0, PL1)
	registers.Push(addrVInterleaveIndices, addrTwiddles)

}

func (_f *FFAmd64) generateSISShuffleF31() {
	// see comments in generateSIS512_16F31
	// this does not need to be optimized as it's not on the hot path.
	f := &fftHelper{_f}
	const argSize = 1 * 3 * 8
	stackSize := f.StackSize(f.NbWords*2+4, 2, 0)
	registers := f.FnHeader("sisShuffle_avx512", stackSize, argSize, amd64.AX, amd64.DI)
	defer f.AssertCleanStack(stackSize, 0)

	addrA := registers.Pop()
	lenA := registers.Pop()

	b0 := registers.PopV()
	a0 := registers.PopV()
	a1 := registers.PopV()

	f.MOVQ("a+0(FP)", addrA)
	f.MOVQ("a_len+8(FP)", lenA)

	// divide len by 32
	f.SHRQ("$5", lenA)

	f.loadMasks()

	addrVInterleaveIndices := registers.Pop()
	vInterleaveIndices := registers.PopV()
	f.MOVQ("·vInterleaveIndices+0(SB)", addrVInterleaveIndices)
	f.VMOVDQU64(addrVInterleaveIndices.At(0), vInterleaveIndices)

	f.Loop(lenA, func() {
		f.VMOVDQU32(addrA.AtD(0), a0, "load a[i]")
		f.VMOVDQU32(addrA.AtD(16), a1, "load a[i+16]")

		// probably a faster way to do this, but let's do it the naive way for now
		// it's not on the hot path.
		// the idea here is to "shuffle" a vector the way the avx512 fft DIF would.
		f.permute8x8(a0, a1, b0)
		f.permute4x4(a0, a1, vInterleaveIndices, b0)
		f.permute2x2(a0, a1, b0)
		f.permute1x1(a0, a1, b0)

		f.VMOVDQU32(a0, addrA.AtD(0), "store a[i]")
		f.VMOVDQU32(a1, addrA.AtD(16), "store a[i+16]")

		f.ADDQ("$128", addrA)
	})
	f.RETAVX()
}

func (_f *FFAmd64) generateSISUnhuffleF31() {
	// see comments in generateSIS512_16F31
	// this does not need to be optimized as it's not on the hot path.
	f := &fftHelper{_f}
	const argSize = 1 * 3 * 8
	stackSize := f.StackSize(f.NbWords*2+4, 2, 0)
	registers := f.FnHeader("sisUnshuffle_avx512", stackSize, argSize, amd64.AX, amd64.DI)
	defer f.AssertCleanStack(stackSize, 0)

	addrA := registers.Pop()
	lenA := registers.Pop()

	b0 := registers.PopV()
	a0 := registers.PopV()
	a1 := registers.PopV()

	f.MOVQ("a+0(FP)", addrA)
	f.MOVQ("a_len+8(FP)", lenA)

	// divide len by 32
	f.SHRQ("$5", lenA)

	f.loadMasks()

	addrVInterleaveIndices := registers.Pop()
	vInterleaveIndices := registers.PopV()
	f.MOVQ("·vInterleaveIndices+0(SB)", addrVInterleaveIndices)
	f.VMOVDQU64(addrVInterleaveIndices.At(0), vInterleaveIndices)

	f.Loop(lenA, func() {
		f.VMOVDQU32(addrA.AtD(0), a0, "load a[i]")
		f.VMOVDQU32(addrA.AtD(16), a1, "load a[i+16]")

		// ok let's say now each pair of vector v0 v1
		// such that
		// v0 = [a0 a2 a4 a6 a8 a10 a12 a14 | b0 b2 b4 b6 b8 b10 b12 b14]
		// v1 = [a1 a3 a5 a7 a9 a11 a13 a15 | b1 b3 b5 b7 b9 b11 b13 b15]

		// we need to repack them; let's do it the naive way for now
		// this is what the FFT avx512 DIF would do as last step;
		// but for SIS we can skip it and do it only once at the end, very useful
		// when hashing lots of vectors.
		f.VPUNPCKLDQ(a1, a0, b0)
		f.VPUNPCKHDQ(a1, a0, a1)
		f.VMOVDQA32(b0, a0)
		f.permute4x4(a0, a1, vInterleaveIndices, b0)
		f.permute8x8(a0, a1, b0)

		f.VMOVDQU32(a0, addrA.AtD(0), "store a[i]")
		f.VMOVDQU32(a1, addrA.AtD(16), "store a[i+16]")

		f.ADDQ("$128", addrA)
	})

	f.RETAVX()

}
func (_f *FFAmd64) generateSIS512_16F31() {
	// this is a specialized unrolled SIS hash for degree = 512 and log2(bound) = 16
	// essentially, the "pure go" algorithm to hash(v), v being a vector of n elements, is as follows:
	// 1. process v in chunks of 256elements
	// 2. for each chunk of 256 elements, do the following:
	//		- load 256 elements from v
	//		- convert from montgomery to regular form
	//		- split the limbs into (k) 512 values; i.e. separate the uint32 into 2 uint16 (this is the "split" part)
	//		- do a FFT(k, fft.DIF, fft.OnCoset())
	//		- multiply by the Ag[i] (i being the chunk index corresponding to the Ag index)
	//		- accumulate the result in res
	//
	// the AVX512 version follows the same logic, but is unrolled and uses AVX512 instructions (duh.)
	// it minimize memory access, unrolls code when possible, leverage ILP (instruction level parallelism)
	// avoid doing reductions mod q when possible, and a flurry of other tricks.
	//
	// the algorithm is as follows:
	// 1. load 256 values from the chunk (uint32);
	// 		- note that we have 2 pointers; 1 at the beginning "x" and 1 at the middle "xm"
	// 		- this enables us to do the first stage of the FFT directly, and save in registers the first half
	// 		for the next stage of the FFT. the second half is stored on the stack.
	// 2. convert from montgomery to regular form
	// 3. split the limbs from the 256 values into 512 values
	// 4. multiply by cosets
	// 5. perform the FFT first stage (512)
	//		- that is butterfly and multiply by twiddles
	// 6. at the end of this first unrolled loop, we have the first 256 values in registers
	// and the second 256 values on the stack.
	// we still need to do the FFT on these 2 halves, then multiply by rag, and accumulate in res.
	//
	// The result is "shuffled", and before calling the FFT inverse, caller need to call sisUnshuffle.
	// This shuffling is due to the fact that we want to process elements in blocks of 16 (use a full AVX512 DWORD register)
	// And the FFT after a certain stage works with a stride of 8, then 4, then 2, then 1.
	// So we need to shuffle the elements; we could unshuffle them here, but since we accumulate our result in res
	// and this can be called multiple times, we prefer to amortize and do the unshuffle only once at the end.

	f := &fftHelper{_f}
	const argSize = 5 * 3 * 8
	stackSize := f.StackSize(f.NbWords*2+4, 1, 256*4)
	registers := f.FnHeader("sis512_16_avx512", stackSize, argSize, amd64.AX)
	sp := amd64.Register("SP")

	f.Comment("refer to the code generator for comments and documentation.")

	addrK256 := registers.Pop()
	addrK256m := registers.Pop()
	addrCosets := registers.Pop()
	addrCosetsm := registers.Pop()
	addrTwiddlesRoot := registers.Pop()
	addrTwiddles := registers.Pop()
	addrRag := registers.Pop()
	addrRes := registers.Pop()

	qd := registers.PopV()
	qInvNeg := registers.PopV()

	// load q and qInvNeg
	f.loadQ(qd, qInvNeg)
	f.loadMasks()

	f.MOVQ("k256+0(FP)", addrK256)
	f.MOVQ("cosets+24(FP)", addrCosets)
	f.MOVQ("twiddles+48(FP)", addrTwiddlesRoot)
	f.MOVQ("rag+72(FP)", addrRag)
	f.MOVQ("res+96(FP)", addrRes)
	f.MOVQ(addrTwiddlesRoot.At(0), addrTwiddles, "twiddles[0]") // stage 0

	f.MOVQ(addrK256, addrK256m)
	f.MOVQ(addrCosets, addrCosetsm)

	// "middle" pointers
	f.ADDQ(512, addrK256m)
	f.ADDQ(1024, addrCosetsm)

	// this convert vectors from montgomery form to regular form
	// it is a mulD(a, 1)
	// see mulD define for documentation.
	fromMontx2 := f.Define("fromMontgomery", 12, func(args ...any) {
		a := args[0]
		b := args[1]

		tmp0 := args[2]
		tmp1 := args[3]
		tmp2 := args[4]
		tmp3 := args[5]
		tmp4 := args[6]
		tmp5 := args[7]
		tmp6 := args[8]
		tmp7 := args[9]

		q := args[10]
		qInvNeg := args[11]

		f.VPSRLQ("$32", a, tmp1) // keep high 32 bits
		f.VPSRLQ("$32", b, tmp5) // keep high 32 bits

		f.VPMULUDQ(a, qInvNeg, tmp2)
		f.VPMULUDQ(tmp1, qInvNeg, tmp3)
		f.VPMULUDQ(b, qInvNeg, tmp6)
		f.VPMULUDQ(tmp5, qInvNeg, tmp7)

		f.VPMULUDQ(tmp2, q, tmp2)
		f.VPMULUDQ(tmp6, q, tmp6)
		f.VPMULUDQ(tmp3, q, tmp3)
		f.VPMULUDQ(tmp7, q, tmp7)

		f.VPANDDkz(a, a, amd64.K3, tmp0) // keep low 32 bits
		f.VPANDDkz(b, b, amd64.K3, tmp4) // keep low 32 bits

		f.VPADDQ(tmp0, tmp2, tmp0)
		f.VPADDQ(tmp4, tmp6, tmp4)
		f.VPADDQ(tmp1, tmp3, tmp1)
		f.VPADDQ(tmp5, tmp7, tmp5)

		f.VMOVSHDUPk(tmp4, amd64.K3, tmp5)
		f.VMOVSHDUPk(tmp0, amd64.K3, tmp1)

		f.VPSUBD(q, tmp1, tmp3)
		f.VPSUBD(q, tmp5, tmp7)
		f.VPMINUD(tmp1, tmp3, a)
		f.VPMINUD(tmp5, tmp7, b)
	}, true)

	// scratch registers
	s := registers.PopVN(10)

	c0 := registers.PopV()
	c1 := registers.PopV()
	am0 := registers.PopV()
	am1 := registers.PopV()

	// we store the first 256 values directly in register
	a := registers.PopVN(16)

	// split the limbs;
	// we have as input
	// r0 = [a0 a1 a2 a3 a4 a5 a6 a7 a8 a9 a10 a11 a12 a13 a14 a15 a16 a17 a18 a19 a20 a21 a22 a23 a24 a25 a26 a27 a28 a29 a30 a31]
	// (r0 is seen as 32 uint16)
	// we want to split it into
	// r0 = [a0 a1 a2 a3 a4 a5 a6 a7 a8 a9 a10 a11 a12 a13 a14 a15]
	// r1 = [a16 a17 a18 a19 a20 a21 a22 a23 a24 a25 a26 a27 a28 a29 a30 a31]
	// (seen as 16 uint32)
	splitLimbs := func(r0, r1 amd64.VectorRegister) {
		f.VEXTRACTI64X4(1, r0, r1.Y())
		f.VPMOVZXWD(r0.Y(), r0)
		f.VPMOVZXWD(r1.Y(), r1)
	}

	// mul1 and mul2 are just helper to make the loop below more readable.
	// they perform a multiplication of a 16 uint32 vector by a 16 uint32 vector
	// second parameter is a memory address.
	mul1 := func(x amd64.VectorRegister, y amd64.Register) {
		f.VMOVDQU32(y, c0)
		f.mulD(x, c0, s[0], s[1], s[2], s[3], s[4], s[5], qd, qInvNeg)
	}
	mul2 := func(x amd64.VectorRegister, y amd64.Register) {
		f.VMOVDQU32(y, c1)
		f.mulD(x, c1, s[6], s[7], s[8], s[9], s[4], s[5], qd, qInvNeg)
	}

	n := 256 / 16 // process 16 uint32 at a time
	for i := 0; i < n/2; i++ {
		// a0 and a1 are going to be kept in registers
		// am0 and am1 are in the second half of our input, we will store them
		// on the stack after the first layer of the FFT and get back to them later.
		a0 := a[i*2]
		a1 := a[i*2+1]

		f.VMOVDQU32(addrK256.AtD(i*16), a0)
		f.VMOVDQU32(addrK256m.AtD(i*16), am0)

		// convert to regular form
		fromMontx2(a0, am0, s[2], s[3], s[4], s[5], s[6], s[7], s[8], s[9], qd, qInvNeg)

		// split the limbs and mul by coset the first half
		splitLimbs(a0, a1)
		mul1(a0, addrCosets.AtD((i*2)*16))
		mul2(a1, addrCosets.AtD((i*2+1)*16))

		// same thing for the second half
		splitLimbs(am0, am1)
		mul1(am0, addrCosetsm.AtD((i*2)*16))
		mul2(am1, addrCosetsm.AtD((i*2+1)*16))

		// now we can do the first layer of the fft
		f.butterflyD2Q(a0, am0, qd, s[2], s[4])
		f.butterflyD2Q(a1, am1, qd, s[3], s[5])
		mul1(am0, addrTwiddles.AtD((i*2)*16))
		mul2(am1, addrTwiddles.AtD((i*2+1)*16))

		// store am0 and am1 on the stack
		f.VMOVDQU32(am0, sp.AtD((i*2)*16))
		f.VMOVDQU32(am1, sp.AtD((i*2+1)*16))
	}

	registers.PushV(s...)
	registers.PushV(c1, am1, c0, am0)

	// next stage of twiddles
	f.ADDQ("$24", addrTwiddlesRoot)
	f.MOVQ(addrTwiddlesRoot, addrTwiddles) // backup addr twiddles for the other half

	// we unroll the FFT 256 here
	// and call it twice by modifying the "pointers" to the data and the twiddles.
	lblDone := f.NewLabel("done")
	lblFFT256 := f.NewLabel("fft256")
	o := registers.Pop()
	f.MOVQ("$2", o)
	f.LABEL(lblFFT256)
	// we do the fft on the first half.
	// a contains k[:256], addrTwiddlesRoot contains twiddles[1]
	_ = f.generateCoreDIFKernel(256, &registers, addrTwiddlesRoot, a, qd, qInvNeg, false)

	c0 = registers.PopV()
	c1 = registers.PopV()
	s = registers.PopVN(10)

	// we can now mul by RAG the result of the FFT256
	for i := 0; i < len(a); i += 2 {
		mul1(a[i], addrRag.AtD(i*16))
		mul2(a[i+1], addrRag.AtD((i+1)*16))
	}

	// accumulate in res
	for i := 0; i < len(a); i += 2 {
		// res += a[i]
		f.VPADDD(addrRes.AtD(i*16), a[i], a[i])
		f.VPADDD(addrRes.AtD((i+1)*16), a[i+1], a[i+1])

		// reduce in [0, q)
		// perf note:
		// probably a sur-optimization (2-3% gain)
		// but we could have res as a slice of uint64
		// not do the mod reduce here and just do it at the end
		// before the FFT inverse.
		// not only do we skip this reduce here, but we also avoid reducing
		// in [0, q) and reduce in [0, 2q) instead in the mul by rag before.
		f.VPSUBD(qd, a[i], s[4])
		f.VPSUBD(qd, a[i+1], s[5])
		f.VPMINUD(s[4], a[i], a[i])
		f.VPMINUD(s[5], a[i+1], a[i+1])

		// store the result
		f.VMOVDQU32(a[i], addrRes.AtD(i*16))
		f.VMOVDQU32(a[i+1], addrRes.AtD((i+1)*16))
	}

	// decrement the count of FFT256 to do
	f.DECQ(o)
	f.TESTQ(o, o)
	f.JEQ(lblDone)

	// here we are not done, so we setup the other half of the fft.
	// put the second half of the FFT256 in a
	// and twiddles[1] in addrTwiddlesRoot
	f.MOVQ(addrTwiddles, addrTwiddlesRoot)
	for i := range a {
		f.VMOVDQU32(sp.AtD(i*16), a[i])
	}
	// increment the pointers to Rag and Res by 1024 (256*4)
	f.ADDQ(1024, addrRag)
	f.ADDQ(1024, addrRes)

	f.JMP(lblFFT256)

	f.LABEL(lblDone)

	f.RETAVX()
}

type fftHelper struct {
	*FFAmd64
}

func (f *fftHelper) loadQ(q, qInv amd64.VectorRegister) {
	loadQ, _ := f.DefineFn("load_q")
	loadQ(q, qInv)
}

func (f *fftHelper) loadMasks() {
	loadMasks, _ := f.DefineFn("load_masks")
	loadMasks()
}

func (f *fftHelper) butterflyD1Q(args ...any) {
	butterflyD1Q, _ := f.DefineFn("butterflyD1Q")
	butterflyD1Q(args...)
}

func (f *fftHelper) butterflyD2Q(args ...any) {
	butterflyD2Q, _ := f.DefineFn("butterflyD2Q")
	butterflyD2Q(args...)
}

func (f *fftHelper) butterflyD2Q2Q(args ...any) {
	butterflyD2Q2Q, _ := f.DefineFn("butterflyD2Q2Q")
	butterflyD2Q2Q(args...)
}

func (f *fftHelper) mulD(args ...any) {
	mulD, _ := f.DefineFn("mulD")
	mulD(args...)
}

func (f *fftHelper) butterfly_mulD(args ...any) {
	butterfly_mulD, _ := f.DefineFn("butterfly_mulD")
	butterfly_mulD(args...)
}

func (f *fftHelper) permute1x1(args ...any) {
	permute1x1, _ := f.DefineFn("permute1x1")
	permute1x1(args...)
}

func (f *fftHelper) permute2x2(args ...any) {
	permute2x2, _ := f.DefineFn("permute2x2")
	permute2x2(args...)
}

func (f *fftHelper) permute4x4(args ...any) {
	permute4x4, _ := f.DefineFn("permute4x4")
	permute4x4(args...)
}

func (f *fftHelper) permute8x8(args ...any) {
	permute8x8, _ := f.DefineFn("permute8x8")
	permute8x8(args...)
}
