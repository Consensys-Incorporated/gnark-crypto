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
	"strconv"

	"github.com/consensys/bavard/amd64"
)

func (f *FFAmd64) generatePoseidon2_F31(params Poseidon2Parameters) {
	fullRounds := params.FullRounds
	partialRounds := params.PartialRounds
	width := params.Width
	rf := fullRounds / 2

	if width != 16 && width != 24 {
		panic("only width 16 and 24 are supported")
	}
	fnName := fmt.Sprintf("permutation%d_avx512", width)

	width24 := width == 24

	const argSize = 2 * 3 * 8
	stackSize := f.StackSize(f.NbWords*2+4, 2, 0)
	registers := f.FnHeader(fnName, stackSize, argSize, amd64.AX, amd64.DX)
	defer f.AssertCleanStack(stackSize, 0)

	addrInput := registers.Pop()
	addrRoundKeys := registers.Pop()
	addrDiagonal := registers.Pop()
	rKey := registers.Pop()

	// constants
	qd := registers.PopV()
	qInvNeg := registers.PopV()

	// state
	b0 := registers.PopV()
	b1 := registers.PopV()

	// temporary registers
	v0 := registers.PopV()
	v1 := registers.PopV()
	t0 := registers.PopV()
	t1 := registers.PopV()
	t2 := registers.PopV()
	t3 := registers.PopV()
	t4 := registers.PopV()
	t5 := registers.PopV()
	aOdd := registers.PopV()
	bOdd := registers.PopV()
	PL0 := registers.PopV()
	PL1 := registers.PopV()
	acc := registers.PopV().Y()
	accShuffled := registers.PopV().Y()

	// diagonal of the matrix
	d0 := registers.PopV()
	d0odd := registers.PopV()
	d1 := registers.PopV()
	d1odd := registers.PopV()

	// prepare the masks used for shuffling the vectors
	f.MOVQ(uint64(0b01_01_01_01_01_01_01_01), amd64.AX)
	f.KMOVD(amd64.AX, amd64.K3)

	f.MOVQ(uint64(0x1), amd64.AX)
	f.KMOVQ(amd64.AX, amd64.K2)

	// load the constants
	f.MOVD("$const_q", amd64.AX)
	f.VPBROADCASTD(amd64.AX, qd)
	f.MOVD("$const_qInvNeg", amd64.AX)
	f.VPBROADCASTD(amd64.AX, qInvNeg)

	f.MOVQ("input+0(FP)", addrInput)
	f.MOVQ("roundKeys+24(FP)", addrRoundKeys)
	// load the 3 * 8 uint32
	f.VMOVDQU32(addrInput.AtD(0), b0)
	if width24 {
		f.VMOVDQU32(addrInput.AtD(16), b1.Y())
		f.MOVQ("·diag24+0(SB)", addrDiagonal)
		f.VMOVDQU32(addrDiagonal.AtD(0), d0)
		f.VMOVDQU32(addrDiagonal.AtD(16), d1.Y())
		f.VPSRLQ("$32", d0, d0odd)
		f.VPSRLQ("$32", d1.Y(), d1odd.Y())
	} else {
		f.MOVQ("·diag16+0(SB)", addrDiagonal)
		f.VMOVDQU32(addrDiagonal.AtD(0), d0)
		f.VPSRLQ("$32", d0, d0odd)
	}

	add := f.Define("add", 5, func(args ...any) {
		a := args[0]
		b := args[1]
		qd := args[2]
		r0 := args[3]
		into := args[4]

		f.VPADDD(a, b, into)
		f.VPSUBD(qd, into, r0)
		f.VPMINUD(into, r0, into)
	}, true)

	matMulM4 := f.Define("mat_mul_m4", 6, func(args ...any) {
		block := args[0]
		t0 := args[1]
		t1 := args[2]
		t2 := args[3]
		qd := args[4]
		r0 := args[5]
		// We multiply by this matrix, each block of 4:
		// (2 3 1 1)
		// (1 2 3 1)
		// (1 1 2 3)
		// (3 1 1 2)
		// so we have
		// s0 = Σ + s0 + 2s1
		// s1 = Σ + s1 + 2s2
		// s2 = Σ + s2 + 2s3
		// s3 = Σ + s3 + 2s0
		// 1. we compute the sum
		// 2. we compute the shifted double(s)
		// 3. we add
		f.VPSHUFD(uint64(0x4e), block, t0)
		add(t0, block, qd, r0, t0)
		f.VPSHUFD(uint64(0xb1), t0, t1)
		add(t0, t1, qd, r0, t0)

		f.VPSHUFD(uint64(0x39), block, t2)
		f.VPSLLD("$1", t2, t2)
		f.VPSUBD(qd, t2, r0)
		f.VPMINUD(t2, r0, t2)

		// compute the sum
		add(block, t0, qd, r0, block)
		add(block, t2, qd, r0, block)
	}, true)

	matMulExternalInPlace := f.Define("mat_mul_external", 0, func(args ...any) {
		matMulM4(b0, t0, t1, t2, qd, t5)
		matMulM4(b1.Y(), t0.Y(), t1.Y(), t2.Y(), qd.Y(), t5.Y())

		// matMulExternalInPlace
		// we need to compute
		// acc[0] = Σ s[i%4]
		// acc[1] = Σ s[(i+1)%4]
		// acc[2] = Σ s[(i+2)%4]
		// acc[3] = Σ s[(i+3)%4]
		f.VEXTRACTI64X4(1, b0, acc)
		add(acc, b0.Y(), qd.Y(), t5.Y(), acc)
		add(acc, b1.Y(), qd.Y(), t5.Y(), acc)

		// we now have a Y register with the 8 elements
		// we permute to compute the desired result duplicated in acc[0..3] and acc[4..7]
		f.VSHUFF64X2(0b1, acc, acc, accShuffled)
		add(acc.Y(), accShuffled, qd.Y(), t5.Y(), acc.Y())

		f.VINSERTI64X4(1, acc.Y(), acc.Z(), acc.Z())

		add(b1.Y(), acc.Y(), qd.Y(), t3.Y(), b1.Y())
		add(b0, acc.Z(), qd, t5, b0)
	}, true)
	if !width24 {
		matMulExternalInPlace = f.Define("mat_mul_external_16", 0, func(args ...any) {
			matMulM4(b0, t0, t1, t2, qd, t5)
			f.VEXTRACTI64X4(1, b0, acc)
			add(acc, b0.Y(), qd.Y(), t5.Y(), acc)
			f.VSHUFF64X2(0b1, acc, acc, accShuffled)
			add(acc.Y(), accShuffled, qd.Y(), t5.Y(), acc.Y())
			f.VINSERTI64X4(1, acc.Y(), acc.Z(), acc.Z())
			add(b0, acc.Z(), qd, t5, b0)
		}, true)
	}

	// computes c = a * b mod q
	// a and b can be in [0, 2q)
	_mulD := f.Define("mulD", 11, func(args ...any) {
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
		c := args[10]

		f.VMOVSHDUP(a, aOdd) // keep high 32 bits (port 5 instead of port 0/1)
		f.VMOVSHDUP(b, bOdd) // keep high 32 bits (port 5 instead of port 0/1)

		// VPMULUDQ conveniently ignores the high 32 bits of each QWORD lane
		f.VPMULUDQ(a, b, b0)
		f.VPMULUDQ(aOdd, bOdd, b1)
		f.VPMULUDQ(b0, qInvNeg, PL0)
		f.VPMULUDQ(b1, qInvNeg, PL1)

		f.VPMULUDQ(PL0, q, PL0)
		f.VPADDQ(b0, PL0, b0)

		f.VPMULUDQ(PL1, q, PL1)
		f.VPADDQ(b1, PL1, c)

		f.VMOVSHDUPk(b0, amd64.K3, c)
	}, true)

	reduce1Q := f.Define("reduce1Q", 3, func(args ...any) {
		qd := args[0]
		c := args[1]
		r0 := args[2]

		f.VPSUBD(qd, c, r0)
		f.VPMINUD(c, r0, c)
	}, true)

	mulY := func(a, b, c amd64.VectorRegister, reduce bool) {
		mulInput := []any{a, b, aOdd, bOdd, t0, t1, PL0, PL1, qd, qInvNeg, c}
		for i := range mulInput {
			mulInput[i] = (mulInput[i]).(amd64.VectorRegister).Y()
		}
		_mulD(mulInput...)
		if reduce {
			reduce1Q(mulInput[8], mulInput[10], mulInput[7])
		}
	}

	mul := func(a, b, c amd64.VectorRegister, reduce bool) {
		mulInput := []any{a, b, aOdd, bOdd, t0, t1, PL0, PL1, qd, qInvNeg, c}
		for i := range mulInput {
			mulInput[i] = (mulInput[i]).(amd64.VectorRegister).Z()
		}
		_mulD(mulInput...)
		if reduce {
			reduce1Q(mulInput[8], mulInput[10], mulInput[7])
		}
	}

	var sbox, sboxPartial defineFn
	switch params.SBoxDegree {
	case 3:
		sbox = f.Define("sbox_full", 0, func(args ...any) {
			mul(b0, b0, t2, false)
			mul(b0, t2, b0, true)

			mulY(b1, b1, t1, false)
			mulY(b1, t1, b1, true)

		}, true)
		if !width24 {
			sbox = f.Define("sbox_full_16", 0, func(args ...any) {
				// Fused sbox3: b0^3 with odd lanes cached via VMOVSHDUP (port 5),
				// reused across both multiplications (saves 3 extractions → 1).
				f.VMOVSHDUP(b0, aOdd)

				// sq = b0 * b0 (no reduce)
				f.VPMULUDQ(b0, b0, t0)
				f.VPMULUDQ(aOdd, aOdd, t1)
				f.VPMULUDQ(t0, qInvNeg, PL0)
				f.VPMULUDQ(t1, qInvNeg, PL1)
				f.VPMULUDQ(PL0, qd, PL0)
				f.VPADDQ(t0, PL0, t0)
				f.VPMULUDQ(PL1, qd, PL1)
				f.VPADDQ(t1, PL1, t2)
				f.VMOVSHDUPk(t0, amd64.K3, t2)

				// b0 = b0 * sq (with reduce)
				f.VMOVSHDUP(t2, bOdd)
				f.VPMULUDQ(b0, t2, t0)
				f.VPMULUDQ(aOdd, bOdd, t1)
				f.VPMULUDQ(t0, qInvNeg, PL0)
				f.VPMULUDQ(t1, qInvNeg, PL1)
				f.VPMULUDQ(PL0, qd, PL0)
				f.VPADDQ(t0, PL0, t0)
				f.VPMULUDQ(PL1, qd, PL1)
				f.VPADDQ(t1, PL1, b0)
				f.VMOVSHDUPk(t0, amd64.K3, b0)

				reduce1Q(qd, b0, PL1)
			}, true)
		}

		sboxPartial = f.Define("sbox_partial", 0, func(args ...any) {
			// t2.X() = b0 * b0
			// this is similar to the mulD macro
			// but since we only care about the mul result at [0],
			// we unroll and remove unnecessary code.
			f.VPMULUDQ(v1.X(), v1.X(), t0.X())
			f.VPMULUDQ(t0.X(), qInvNeg.X(), PL0.X())
			f.VPMULUDQ(PL0.X(), qd.X(), PL0.X())
			f.VPADDQ(t0.X(), PL0.X(), t0.X())
			f.VPSRLQ("$32", t0.X(), t2.X())

			// b0 = b0 * t2.X()
			f.VPMULUDQ(v1.X(), t2.X(), t0.X())
			f.VPMULUDQ(t0.X(), qInvNeg.X(), PL0.X())
			f.VPMULUDQ(PL0.X(), qd.X(), PL0.X())
			f.VPADDQ(t0.X(), PL0.X(), t0.X())
			f.VPSRLQ("$32", t0.X(), v1.X())
			f.VPSUBD(qd.X(), v1.X(), PL0.X())
			f.VPMINUD(v1.X(), PL0.X(), v1.X())
		}, true)
	case 7:
		sbox = f.Define("sbox_full", 0, func(args ...any) {
			mul(b0, b0, t2, true)
			mul(t2, t2, t3, false)
			mul(b0, t2, b0, false)
			mul(b0, t3, b0, true)

			mulY(b1, b1, t2, true)
			mulY(t2, t2, t3, false)
			mulY(b1, t2, b1, false)
			mulY(b1, t3, b1, true)
		}, true)
		if !width24 {
			sbox = f.Define("sbox_full_16", 0, func(args ...any) {
				mul(b0, b0, t2, true)
				mul(t2, t2, t3, false)
				mul(b0, t2, b0, false)
				mul(b0, t3, b0, true)
			}, true)
		}

		sboxPartial = f.Define("sbox_partial", 0, func(args ...any) {
			mulY(v1, v1, t2, true)
			mulY(t2, t2, t3, false)
			mulY(v1, t2, v1, false)
			mulY(v1, t3, v1, true)
		}, true)
	default:
		panic("only SBox degree 3 and 7 are supported")
	}

	sumState := f.Define("sum_state", 0, func(args ...any) {
		// first we compute the sum
		f.VEXTRACTI64X4(1, b0, acc) // TODO @gbotrel here
		add(acc, b1.Y(), qd.Y(), t5.Y(), acc)
		add(acc, t4.Y(), qd.Y(), t5.Y(), acc)

		// now we can work with acc.Y()
		f.VSHUFF64X2(0b1, acc, acc, accShuffled)
		add(acc, accShuffled, qd.Y(), t5.Y(), acc)

		f.VPSHUFD(uint64(0x4e), acc, accShuffled)
		add(acc, accShuffled, qd.Y(), t5.Y(), acc)
		f.VPSHUFD(uint64(0xb1), acc, accShuffled)
		add(acc, accShuffled, qd.Y(), t5.Y(), acc)

		f.VINSERTI64X4(1, acc, acc.Z(), acc.Z())
	}, true)
	if !width24 {
		sumState = f.Define("sum_state_16", 0, func(args ...any) {
			// first we compute the sum
			f.VEXTRACTI64X4(1, b0, acc)
			add(acc, t4.Y(), qd.Y(), t5.Y(), acc)

			// now we can work with acc.Y()
			f.VSHUFF64X2(0b1, acc, acc, accShuffled)
			add(acc, accShuffled, qd.Y(), t5.Y(), acc)

			f.VPSHUFD(uint64(0x4e), acc, accShuffled)
			add(acc, accShuffled, qd.Y(), t5.Y(), acc)
			f.VPSHUFD(uint64(0xb1), acc, accShuffled)
			add(acc, accShuffled, qd.Y(), t5.Y(), acc)

			f.VINSERTI64X4(1, acc, acc.Z(), acc.Z())
		}, true)
	}

	fullRound := f.Define("full_round", 0, func(args ...any) {
		// load round keys
		f.VMOVDQU32(rKey.AtD(0), v0)
		f.VMOVDQU32(rKey.AtD(16), v1.Y())

		// add round keys
		add(b0, v0, qd, t5, b0)
		add(b1.Y(), v1.Y(), qd.Y(), t2.Y(), b1.Y())
		sbox()
		matMulExternalInPlace()
	}, true)
	if !width24 {
		fullRound = f.Define("full_round_16", 0, func(args ...any) {
			// load round keys
			f.VMOVDQU32(rKey.AtD(0), v0)

			// add round keys
			add(b0, v0, qd, t5, b0)
			sbox()
			matMulExternalInPlace()
		}, true)
	}

	partialRound := func() {
		// load round keys
		f.VMOVD(rKey.At(0), v0.X())
		// copy b0 to break the dependency chain
		f.VMOVDQA32(b0, t4)

		add(t4.X(), v0.X(), qd.X(), PL0.X(), v1.X())

		// do the sbox
		sboxPartial()

		// merge the sbox at the first index of b0
		f.VPBLENDMD(v1, t4, t4, amd64.K2)

		// multiply b1 by diagonal[1] (diag24)
		// this is equivalent to mulY(b1, d1, t3, true)
		// but we already have d1Odd that don't change so we unroll and modify the code
		if width24 {
			f.VPSRLQ("$32", b1.Y(), aOdd.Y())
			f.VPMULUDQ(b1.Y(), d1.Y(), t0.Y())
			f.VPMULUDQ(aOdd.Y(), d1odd.Y(), t1.Y())
			f.VPMULUDQ(t0.Y(), qInvNeg.Y(), PL0.Y())
			f.VPMULUDQ(t1.Y(), qInvNeg.Y(), PL1.Y())

			f.VPMULUDQ(PL0.Y(), qd.Y(), PL0.Y())
			f.VPADDQ(t0.Y(), PL0.Y(), t0.Y())

			f.VPMULUDQ(PL1.Y(), qd.Y(), PL1.Y())
			f.VPADDQ(t1.Y(), PL1.Y(), t3.Y())

			f.VMOVSHDUPk(t0.Y(), amd64.K3, t3.Y())
			f.VPSUBD(qd.Y(), t3.Y(), t5.Y())
			f.VPMINUD(t3.Y(), t5.Y(), t3.Y())
		}

		// multiply the part of b0 that don't depend on b[0] (i.e. round keys + sbox)
		f.VPSRLQ("$32", b0, aOdd)
		f.VPMULUDQ(aOdd, d0odd, t2)
		f.VPMULUDQ(t2, qInvNeg, PL1)
		f.VPMULUDQ(PL1, qd, PL1)
		f.VPADDQ(t2, PL1, t2)

		// compute the sum: this depends on applying the sbox to b[0]
		sumState()

		// multiply the part of b0 that depends on b[0]
		f.VPMULUDQ(t4, d0, t0)
		f.VPMULUDQ(t0, qInvNeg, PL0)
		f.VPMULUDQ(PL0, qd, PL0)
		f.VPADDQ(t0, PL0, t0)

		f.VMOVSHDUPk(t0, amd64.K3, t2)
		f.VPSUBD(qd, t2, t5)
		f.VPMINUD(t2, t5, b0)

		// now we add the sum
		add(b0, acc.Z(), qd, t5, b0)
		if width24 {
			add(t3.Y(), acc.Y(), qd.Y(), v1.Y(), b1.Y())
		}
	}

	if !width24 && params.SBoxDegree == 3 {
		// Optimized partial round for width=16, degree=3.
		//
		// Three independent chains run concurrently via OOO execution:
		//   A) sbox(s[0] + rc) — scalar X-register chain (~30 cycle latency)
		//   B) horizontal sum of all 16 elements — Y-register shuffles
		//   C) full diagonal multiply of all 16 elements — Z-register muls
		//
		// After all three complete, combine:
		//   sum_tail = sum_all_old - s0_old
		//   sum = sum_tail + s0_new
		//   result[i] = diag[i]*s[i] + sum  (for i > 0)
		//   result[0] = sum_tail - s0_new    (since diag[0] = -2)
		//
		// This hides the sum and diagonal multiply behind the sbox latency,
		// whereas the old code serialized: sbox → blend → sum → even-diag.

		// We reuse b1 (unused for width=16) to save s0_old,
		// and d1/d1odd (unused for width=16) as temps for even-lane diagonal.
		s0Old := b1    // Z3
		evnT0 := d1    // Z20
		evnT1 := d1odd // Z21

		partialRound = func() {
			f.Comment("optimized partial round (ILP)")

			f.VMOVD(rKey.At(0), v0.X())
			f.VMOVDQA32(b0.X(), s0Old.X()) // save s0_old
			add(b0.X(), v0.X(), qd.X(), PL0.X(), v1.X())

			// Chain A: sbox on v1.X() (uses t3.X() instead of t2.X() to avoid Chain C conflict)
			f.VPMULUDQ(v1.X(), v1.X(), t0.X())
			f.VPMULUDQ(t0.X(), qInvNeg.X(), PL0.X())
			f.VPMULUDQ(PL0.X(), qd.X(), PL0.X())
			f.VPADDQ(t0.X(), PL0.X(), t0.X())
			f.VPSRLQ("$32", t0.X(), t3.X())

			f.VPMULUDQ(v1.X(), t3.X(), t0.X())
			f.VPMULUDQ(t0.X(), qInvNeg.X(), PL0.X())
			f.VPMULUDQ(PL0.X(), qd.X(), PL0.X())
			f.VPADDQ(t0.X(), PL0.X(), t0.X())
			f.VPSRLQ("$32", t0.X(), v1.X())
			f.VPSUBD(qd.X(), v1.X(), PL0.X())
			f.VPMINUD(v1.X(), PL0.X(), v1.X())

			// Chain B: horizontal sum of original b0 (independent of Chain A)
			f.VEXTRACTI64X4(1, b0, acc)
			add(acc, b0.Y(), qd.Y(), t5.Y(), acc)
			f.VSHUFF64X2(0b1, acc, acc, accShuffled)
			add(acc, accShuffled, qd.Y(), t5.Y(), acc)
			f.VPSHUFD(uint64(0x4e), acc, accShuffled)
			add(acc, accShuffled, qd.Y(), t5.Y(), acc)
			f.VPSHUFD(uint64(0xb1), acc, accShuffled)
			add(acc, accShuffled, qd.Y(), t5.Y(), acc)

			// Chain C: full diagonal multiply on original b0
			f.VMOVSHDUP(b0, aOdd)

			f.VPMULUDQ(aOdd, d0odd, t2)
			f.VPMULUDQ(t2, qInvNeg, PL1)
			f.VPMULUDQ(PL1, qd, PL1)
			f.VPADDQ(t2, PL1, t2)

			f.VPMULUDQ(b0, d0, evnT0)
			f.VPMULUDQ(evnT0, qInvNeg, evnT1)
			f.VPMULUDQ(evnT1, qd, evnT1)
			f.VPADDQ(evnT0, evnT1, evnT0)

			f.VMOVSHDUPk(evnT0, amd64.K3, t2)
			f.VPSUBD(qd, t2, t5)
			f.VPMINUD(t2, t5, b0)

			// Combine: sum_tail = sum_all - s0_old, sum = sum_tail + s0_new
			f.VPSUBD(s0Old.X(), acc.X(), t4.X())
			f.VPADDD(qd.X(), t4.X(), PL0.X())
			f.VPMINUD(t4.X(), PL0.X(), t4.X())

			f.VPADDD(v1.X(), t4.X(), t3.X())
			f.VPSUBD(qd.X(), t3.X(), PL0.X())
			f.VPMINUD(t3.X(), PL0.X(), t3.X())

			// v0_new = sum_tail - s0_new (diag[0] = -2 identity)
			f.VPSUBD(v1.X(), t4.X(), v0.X())
			f.VPADDD(qd.X(), v0.X(), PL0.X())
			f.VPMINUD(v0.X(), PL0.X(), v0.X())

			f.VPBROADCASTD(t3.X(), t4)
			add(b0, t4, qd, t5, b0)
			f.VPBLENDMD(v0, b0, b0, amd64.K2)
		}
	}

	matMulExternalInPlace()

	for i := range rf {
		f.MOVQ(addrRoundKeys.At(i*3), rKey)
		fullRound()
	}

	f.Comment("loop over the partial rounds")
	{
		n := registers.Pop()
		addrRoundKeys2 := registers.Pop()
		f.MOVQ(partialRounds, n, fmt.Sprintf("nb partial rounds --> %d", partialRounds))
		f.MOVQ(addrRoundKeys, addrRoundKeys2)
		f.ADDQ(rf*24, addrRoundKeys2)

		f.Loop(n, func() {
			f.MOVQ(addrRoundKeys2.At(0), rKey)
			partialRound()
			f.ADDQ("$24", addrRoundKeys2)
		})
	}

	for i := rf + partialRounds; i < fullRounds+partialRounds; i++ {
		f.MOVQ(addrRoundKeys.At(i*3), rKey)
		fullRound()
	}

	f.VMOVDQU32(b0, addrInput.AtD(0))
	if width24 {
		f.VMOVDQU32(b1.Y(), addrInput.AtD(16))
	}

	f.RET()
}

// generatePoseidon2_F31_16x16xN generates the generalized version of
// permutation16x16x512 that accepts variable colSize via parameters
// (gatherIndices and nbSteps) instead of hardcoded 512-element stride.
//
// This is the SIMD/transposed form of Compressx16. In pure Go, the logical flow is:
//
//	var state [16][16]fr.Element
//	for step := 0; step < nbSteps; step++ {
//	    for lane := 0; lane < 16; lane++ {
//	        copy(state[lane][8:], matrix[lane*colSize+step*8:lane*colSize+step*8+8])
//	        Permutation(state[lane][:])
//	        for j := 0; j < 8; j++ {
//	            state[lane][j] = state[lane][8+j] + matrix[lane*colSize+step*8+j]
//	        }
//	    }
//	}
//	for lane := 0; lane < 16; lane++ {
//	    copy(result[lane][:], state[lane][:8])
//	}
//
// Here the 16 independent states are transposed into AVX-512 vectors:
// v[j][lane] == state[lane][j].
// generatePoseidon2_F31_16x16xN generates the compression kernel. When gather is
// true it emits permutation16x16xN_avx512, which reads a row-major matrix
// (leaf per row) and transposes the 16 lanes with VPGATHERDD. When gather is
// false it emits permutation16x16xN_columns_avx512, which reads a column-major
// matrix (matrix[col*16+lane]) so that each coordinate load is a single
// contiguous VMOVDQU32 — no gather. The column-major layout is what a Merkle
// commitment naturally has (16 contiguous leaves per column) and lets the kernel
// scale with cores instead of saturating the memory subsystem with scattered
// gather traffic.
func (_f *FFAmd64) generatePoseidon2_F31_16x16xN(params Poseidon2Parameters, gather bool) {
	f := &fieldHelper{FFAmd64: _f, twoAdicity: twoAdicityFromParams(params)}
	width := params.Width
	fullRounds := params.FullRounds
	partialRounds := params.PartialRounds
	rf := fullRounds / 2

	_ = rf
	_ = partialRounds

	if width != 16 {
		panic("only width 16 is supported")
	}
	fnName := "permutation16x16xN_avx512"
	// func permutation16x16xN_avx512(matrix *fr.Element, roundKeys [][]fr.Element, result *fr.Element, gatherIndices *uint32, nbSteps uint64)
	argSize := 7 * 8 // matrix(8) + roundKeys(24) + result(8) + gatherIndices(8) + nbSteps(8)
	nbStepsArg := "nbSteps+48(FP)"
	if !gather {
		fnName = "permutation16x16xN_columns_avx512"
		// func permutation16x16xN_columns_avx512(matrix *fr.Element, roundKeys [][]fr.Element, result *fr.Element, nbSteps uint64, state *fr.Element)
		argSize = 7 * 8 // matrix(8) + roundKeys(24) + result(8) + nbSteps(8) + state(8)
		nbStepsArg = "nbSteps+40(FP)"
	}
	stackSize := f.StackSize(f.NbWords*2+4, 2, 0)
	registers := f.FnHeader(fnName, stackSize, argSize, amd64.AX, amd64.DX)
	defer f.AssertCleanStack(stackSize, 0)
	f.registers = &registers

	// v[0..15] is the transposed Poseidon2 state:
	// v[j] holds coordinate j for 16 independent lanes.
	v := registers.PopVN(16)

	// constants
	f.loadQ()
	f.loadQInvNeg()

	addrMatrix := registers.Pop()
	addrResult := registers.Pop()
	addrRoundKeys := registers.Pop()
	rKey := registers.Pop()

	// prepare the mask used for the merging mul results
	f.MOVQ(uint64(0b01_01_01_01_01_01_01_01), amd64.AX)
	f.KMOVD(amd64.AX, amd64.K3)

	f.MOVQ("matrix+0(FP)", addrMatrix)
	f.MOVQ("roundKeys+8(FP)", addrRoundKeys)
	f.MOVQ("result+32(FP)", addrResult)

	const blockSize = 4
	const nbBlocks = 16 / blockSize

	if gather {
		for i := range 16 {
			f.VXORPS(v[i], v[i], v[i])
		}
	} else {
		// The column kernel optionally starts from 16 caller-provided 8-element
		// states in transposed layout. The nil check is outside the hot loop.
		addrState := registers.Pop()
		zeroState := f.NewLabel("zero_state")
		stateReady := f.NewLabel("state_ready")
		f.MOVQ("state+48(FP)", addrState)
		f.TESTQ(addrState, addrState)
		f.JEQ(zeroState)
		for i := range 8 {
			f.VMOVDQU32(addrState.AtD(i*16), v[i])
		}
		for i := 8; i < 16; i++ {
			f.VXORPS(v[i], v[i], v[i])
		}
		f.JMP(stateReady)
		f.LABEL(zeroState)
		for i := range 16 {
			f.VXORPS(v[i], v[i], v[i])
		}
		f.LABEL(stateReady)
		registers.Push(addrState)
	}

	// Load nbSteps from parameter (instead of hardcoded 64)
	N := registers.Pop()
	f.MOVQ(nbStepsArg, N)

	// maskFFFF is the all-ones mask reloaded before each gather/scatter.
	// addrIndexGather holds the gather-index address (gather mode) and is
	// reused for the scatter-index address at the tail.
	maskFFFF := registers.Pop()
	addrIndexGather := registers.Pop()
	f.MOVQ("$0xffffffffffffffff", maskFFFF)

	// vIndexGather is only needed for the gather-based transpose.
	var vIndexGather amd64.VectorRegister
	if gather {
		// Load gather indices from parameter (instead of global indexGather512)
		vIndexGather = registers.PopV()
		f.MOVQ("gatherIndices+40(FP)", addrIndexGather)
		f.VMOVDQU32(addrIndexGather.At(0), vIndexGather)
	}

	// addRoundKeySbox emits:
	//   v[index] = S(v[index] + rc[index])
	// where rc[index] is broadcast to all 16 SIMD lanes because all lanes are
	// at the same Poseidon2 round, only on different inputs.
	addRoundKeySbox := func(index int) {
		rc := registers.PopV()
		f.VPBROADCASTD(rKey.AtD(index), rc)
		f.add(v[index], rc, v[index])
		registers.PushV(rc)
		f.sbox(v[index], v[index], params.SBoxDegree)
	}

	fullRound := func() {
		// Full round:
		//   for j := 0; j < 16; j++ { x[j] = S(x[j] + rc[j]) }
		//   x = M_ext * x
		for j := range v {
			addRoundKeySbox(j)
		}
		f.matMulExternal(v, nbBlocks)
	}

	partialRound := func() {
		// Partial round:
		//   x[0] = S(x[0] + rc[0])
		//   x = M_int * x
		//
		// The internal linear layer is M_int(x) = sum(x)*1 + diag o x, where
		// diag[0] = -2 and the full sum = v[0] + v[1] + ... + v[15].
		//
		// We use the Plonky3 trick of computing sum_tail = v[1]+...+v[15] first
		// (independent of v[0]'s sbox result) and then deriving:
		//   sum = sum_tail + v[0]
		//   v[0] = sum - 2*v[0] = sum_tail - v[0]
		// This eliminates the double(v[0]) operation and enables better ILP:
		// sum_tail computation + diagonal multiply on v[1..15] can run concurrently
		// with the sbox dependency chain on v[0].
		addRoundKeySbox(0)

		sum := registers.PopV()
		t1 := registers.PopV()
		t2 := registers.PopV()
		t3 := registers.PopV()
		t4 := registers.PopV()

		{
			// sum_tail = v[1] + ... + v[15], computed as a tree (14 adds).
			// This is INDEPENDENT of v[0]'s sbox result.
			f.add(v[1], v[2], t2)
			f.add(v[3], v[4], t3)
			f.add(v[5], v[6], t4)
			f.add(v[7], v[8], sum)
			f.add(v[9], t2, t2)
			f.add(v[10], t3, t3)
			f.add(v[11], t4, t4)
			f.add(v[12], sum, sum)
			f.add(v[13], t2, t2)
			f.add(v[14], t3, t3)
			f.add(v[15], t4, t4)
			f.add(t2, t3, t2)
			f.add(t4, sum, t4)
			f.add(t2, t4, sum) // sum = sum_tail
		}

		// Apply diag16 coordinate-wise to v[1..15] (v[0] handled below via sum_tail trick).
		//
		// diag16 = [
		//   -2, 1, 2, 1/2, 3, 4, -1/2, -3,
		//   -4, 1/2^8, 1/2^3, 1/2^24, -1/2^8, -1/2^3, -1/2^4, -1/2^24,
		// ]
		f.double(v[2], v[2])
		f.halve(v[3], v[3])
		f.double(v[4], t2)
		f.add(v[4], t2, v[4])
		f.double(v[5], v[5])
		f.double(v[5], v[5])
		f.halve(v[6], v[6])
		f.double(v[7], t1)
		f.add(v[7], t1, v[7])
		f.double(v[8], v[8])
		f.double(v[8], v[8])

		registers.PushV(t1, t2, t3, t4)

		// Pre-load the odd factor r for the VPMADDUBSW-based N=8 specialisation.
		rConst := registers.PopV()
		oddFactor := (1 << (31 - f.twoAdicity)) - 1
		f.MOVD("$"+strconv.Itoa(oddFactor), amd64.AX)
		f.VPBROADCASTD(amd64.AX, rConst)

		ns := []int{8, 3, 24, 8, 3, 4, 24}
		for i := 9; i < len(v); i++ {
			if ns[i-9] == 8 {
				f.mul2ExpNeg8(v[i], rConst, v[i])
			} else {
				f.mul2ExpNegN(v[i], ns[i-9], v[i])
			}
		}

		registers.PushV(rConst)

		// Derive the full sum and apply it.
		// sum currently holds sum_tail = v[1]+...+v[15].
		// We need: full_sum = sum_tail + v[0] AND v[0]_new = sum_tail - v[0].
		// Compute full_sum FIRST (into t1), then overwrite v[0].
		t1 = registers.PopV()
		f.add(sum, v[0], t1)   // t1 = sum_tail + v[0] = full sum
		f.sub(sum, v[0], v[0]) // v[0] = sum_tail - v[0] = sum - 2*v[0]

		// Now t1 = full sum. Use it for elements 1..15.
		f.add(t1, v[1], v[1])
		f.add(v[2], t1, v[2])
		f.add(v[3], t1, v[3])
		f.add(v[4], t1, v[4])
		f.add(v[5], t1, v[5])
		f.sub(t1, v[6], v[6])
		f.sub(t1, v[7], v[7])
		f.sub(t1, v[8], v[8])
		for i := 9; i < len(v); i++ {
			if i <= 11 {
				f.add(v[i], t1, v[i])
			} else {
				f.sub(t1, v[i], v[i])
			}
		}

		registers.PushV(t1)
		registers.PushV(sum)
	}

	// loop emits a round loop over a contiguous range of round keys.
	loop := func(nbRounds int, fn func()) {
		n := registers.Pop()
		f.MOVQ(nbRounds, n)
		f.Loop(n, func() {
			f.MOVQ(addrRoundKeys.At(0), rKey)
			fn()
			f.ADDQ("$24", addrRoundKeys)
		})
		registers.Push(n)
	}

	// Main absorb/permutation loop over 8-column chunks.
	//
	// Each iteration does the SIMD/transposed equivalent of:
	//   copy(state[lane][8:], nextChunk)
	//   state[lane] = Permutation(state[lane])
	//   state[lane][0:8] = state[lane][8:16] + nextChunk
	if gather {
		registers.PushV(vIndexGather)
	}
	f.Loop(N, func() {
		vTmpInputs := registers.PopVN(8)
		if gather {
			vIndexGather = registers.PopV()
			f.VMOVDQU32(addrIndexGather.At(0), vIndexGather)
			// Gather matrix[*][step*8:(step+1)*8] into the rate coordinates v[8:16].
			// vTmpInputs keeps a copy for the feed-forward added after the permutation.
			for i := range 8 {
				f.KMOVD(maskFFFF, amd64.K1)
				f.VPGATHERDD(i*4, addrMatrix, vIndexGather, 4, amd64.K1, v[i+8])
				f.VMOVDQA32(v[i+8], vTmpInputs[i])
			}
			registers.PushV(vIndexGather)
		} else {
			// Column-major input: the 16 leaves of column (step*8+i) are 16
			// contiguous elements at matrix[(step*8+i)*16:]. Each rate
			// coordinate is therefore a single contiguous 512-bit load — no
			// gather. vTmpInputs keeps a copy for the feed-forward.
			for i := range 8 {
				f.VMOVDQU32(addrMatrix.AtD(i*16), v[i+8])
				f.VMOVDQA32(v[i+8], vTmpInputs[i])
			}
		}

		// Poseidon2 begins with the external matrix:
		//   x = M_ext * x
		f.matMulExternal(v, nbBlocks)

		f.Comment("loop over the first full rounds")
		loop(rf, fullRound)

		f.Comment("loop over the partial rounds")
		loop(partialRounds, partialRound)

		f.Comment("loop over the final full rounds")
		loop(rf, fullRound)

		// Feed-forward for the compression mode:
		//   state'[0:8] = state'[8:16] + absorbedChunk
		for i := range 8 {
			f.add(vTmpInputs[i], v[i+8], v[i])
		}
		registers.PushV(vTmpInputs...)

		if gather {
			// Row-major: move from chunk k to chunk k+1 inside each matrix row
			// (+8 elements); the gather indices provide the per-row stride.
			f.ADDQ(8*4, addrMatrix)
		} else {
			// Column-major: one step consumes 8 full columns = 8*16 elements.
			f.ADDQ(8*16*4, addrMatrix)
		}
		// Each chunk runs a fresh Poseidon2 permutation, so restart round keys at round 0.
		f.MOVQ("roundKeys+8(FP)", addrRoundKeys)
	})

	// Scatter the first 8 coordinates of the 16 transposed states back to the
	// row-major result buffer. This runs once per call (not per step), so the
	// scatter cost is negligible even for the column-major kernel.
	addrScatter8 := addrIndexGather
	vIndexScatter := vIndexGather
	if !gather {
		vIndexScatter = registers.PopV()
	}
	f.MOVQ("·indexScatter8+0(SB)", addrScatter8)
	f.VMOVDQU32(addrScatter8.At(0), vIndexScatter)

	for i := range 8 {
		f.KMOVD(maskFFFF, amd64.K1)
		f.VPSCATTERDD(i*4, addrResult, vIndexScatter, 4, amd64.K1, v[i])
	}

	f.RET()
}

func (_f *FFAmd64) generatePoseidon2_F31_16x24(params Poseidon2Parameters) {
	f := &fieldHelper{FFAmd64: _f, twoAdicity: twoAdicityFromParams(params)}
	width := params.Width
	fullRounds := params.FullRounds
	partialRounds := params.PartialRounds
	rf := fullRounds / 2

	_ = partialRounds
	_ = rf

	if width != 24 {
		panic("only width 24 is supported")
	}
	const fnName = "permutation16x24_avx512"
	// func permutation16x24_avx512(input *[24][16]fr.Element, roundKeys [][]fr.Element)
	const argSize = 4 * 8
	stackSize := f.StackSize(f.NbWords*2+4, 2, 0)
	registers := f.FnHeader(fnName, stackSize, argSize, amd64.AX, amd64.DX)
	defer f.AssertCleanStack(stackSize, 0)
	f.registers = &registers

	// input
	v := registers.PopVN(24)

	// constants
	f.loadQ()
	f.loadQInvNeg()

	addrInput := registers.Pop()
	addrRoundKeys := registers.Pop()
	rKey := registers.Pop()

	// prepare the mask used for the merging mul results
	f.MOVQ(uint64(0b01_01_01_01_01_01_01_01), amd64.AX)
	f.KMOVD(amd64.AX, amd64.K3)

	f.MOVQ("input+0(FP)", addrInput)
	f.MOVQ("roundKeys+8(FP)", addrRoundKeys)

	const blockSize = 4
	const nbBlocks = 24 / blockSize

	// load input
	for i := range v {
		f.VMOVDQU32(addrInput.AtD(i*16), v[i])
	}

	addRoundKeySbox := func(index int) {
		rc := registers.PopV()
		f.VPBROADCASTD(rKey.AtD(index), rc)
		f.add(v[index], rc, v[index])
		registers.PushV(rc)
		f.sbox(v[index], v[index], params.SBoxDegree)
	}

	fullRound := func() {
		for j := range v {
			addRoundKeySbox(j)
		}
		f.matMulExternal(v, nbBlocks)
	}

	partialRound := func() {
		addRoundKeySbox(0)

		// h.matMulInternalInPlace(input)
		// let's do it for koalabear only for now.
		sum := registers.PopV()
		t1 := registers.PopV()
		t2 := registers.PopV()
		t3 := registers.PopV()
		t4 := registers.PopV()

		{
			// compute the sum of all v[i]
			// we do it that way rather than accumulate to break some
			// dependencies chains
			f.add(v[0], v[1], t2)
			f.add(v[2], v[3], t3)
			f.add(v[4], v[5], t4)
			f.add(v[6], v[7], sum)
			for i := 8; i < len(v); i += 4 {
				f.add(v[i], t2, t2)
				f.add(v[i+1], t3, t3)
				f.add(v[i+2], t4, t4)
				f.add(v[i+3], sum, sum)
			}
			f.add(t2, t3, t2)
			f.add(t4, sum, t4)
			f.add(t2, t4, sum)

		}

		// mul by diag24:
		// koalabear:
		// [-2, 1, 2, 1/2, 3, 4, -1/2, -3, -4, 1/2^8, 1/4, 1/8, 1/16, 1/32, 1/64, 1/2^24, -1/2^8, -1/8, -1/16, -1/32, -1/64, -1/2^7, -1/2^9, -1/2^24]
		// babybear:
		// [-2, 1, 2, 1/2, 3, 4, -1/2, -3, -4, 1/2^8, 1/4, 1/8, 1/16, 1/2^7, 1/2^9, 1/2^27, -1/2^8, -1/4, -1/8, -1/16, -1/32, -1/64, -1/2^7, -1/2^27]
		// var temp fr.Element
		// input[0].Sub(&sum, temp.Double(&input[0]))
		f.double(v[0], v[0])

		// input[1].Add(&sum, &input[1])

		// input[2].Add(&sum, temp.Double(&input[2]))
		f.double(v[2], v[2])

		// temp.Set(&input[3]).Halve()
		// input[3].Add(&sum, &temp)
		f.halve(v[3], v[3])
		// temp.Set(&input[6]).Halve()
		// input[6].Sub(&sum, &temp)
		f.halve(v[6], v[6])

		// input[4].Add(&sum, temp.Double(&input[4]).Add(&temp, &input[4]))
		f.double(v[4], t2)
		f.add(v[4], t2, v[4])

		// input[5].Add(&sum, temp.Double(&input[5]).Double(&temp))
		f.double(v[5], v[5])
		f.double(v[5], v[5])

		// input[7].Sub(&sum, temp.Double(&input[7]).Add(&temp, &input[7]))
		f.double(v[7], t1)
		f.add(v[7], t1, v[7])

		// input[8].Sub(&sum, temp.Double(&input[8]).Double(&temp))
		f.double(v[8], v[8])
		f.double(v[8], v[8])

		registers.PushV(t1, t2, t3, t4)

		// Pre-load the odd factor r for the VPMADDUBSW-based N=8 specialisation.
		rConst := registers.PopV()
		oddFactor := (1 << (31 - f.twoAdicity)) - 1
		f.MOVD("$"+strconv.Itoa(oddFactor), amd64.AX)
		f.VPBROADCASTD(amd64.AX, rConst)

		var ns []int
		if params.SBoxDegree == 3 {
			// koalabear
			ns = []int{8, 2, 3, 4, 5, 6, 24, 8, 3, 4, 5, 6, 7, 9, 24}
		} else {
			// babybear
			ns = []int{8, 2, 3, 4, 7, 9, 27, 8, 2, 3, 4, 5, 6, 7, 27}
		}

		for i := 9; i < len(v); i++ {
			if ns[i-9] == 8 {
				f.mul2ExpNeg8(v[i], rConst, v[i])
			} else {
				f.mul2ExpNegN(v[i], ns[i-9], v[i])
			}
		}

		registers.PushV(rConst)

		// Sum part.
		f.sub(sum, v[0], v[0])
		f.add(sum, v[1], v[1])
		f.add(v[2], sum, v[2])
		f.add(v[3], sum, v[3])
		f.add(v[4], sum, v[4])
		f.add(v[5], sum, v[5])
		f.sub(sum, v[6], v[6])
		f.sub(sum, v[7], v[7])
		f.sub(sum, v[8], v[8])
		for i := 9; i < len(v); i++ {
			if i <= 15 {
				f.add(v[i], sum, v[i])
			} else {
				f.sub(sum, v[i], v[i])
			}
		}

		registers.PushV(sum)
	}

	// private function to help write for loops with known bounds
	// for the rounds
	loop := func(nbRounds int, fn func()) {

		n := registers.Pop()
		f.MOVQ(nbRounds, n)

		f.Loop(n, func() {
			// move the current round key address into rKey
			f.MOVQ(addrRoundKeys.At(0), rKey)

			fn()

			// move to the next round key
			f.ADDQ("$24", addrRoundKeys)
		})

		registers.Push(n)
	}

	f.matMulExternal(v, nbBlocks)

	f.Comment("loop over the first full rounds")
	loop(rf, fullRound)

	f.Comment("loop over the partial rounds")
	loop(partialRounds, partialRound)

	f.Comment("loop over the final full rounds")
	loop(rf, fullRound)

	// store the result back
	for i := range v {
		f.VMOVDQU32(v[i], addrInput.AtD(i*16))
	}

	f.RET()
}

type fieldHelper struct {
	*FFAmd64
	registers   *amd64.Registers
	qd, qInvNeg amd64.VectorRegister
	// twoAdicity is j where q = (2^k − 1)·2^j + 1 with k = 31−j.
	// koalabear: j = 24, babybear: j = 27.
	twoAdicity int
}

// twoAdicityFromParams returns j for the field's prime q = (2^k-1)·2^j + 1.
// Determined from SBoxDegree: degree 3 → koalabear (j=24), degree 7 → babybear (j=27).
func twoAdicityFromParams(p Poseidon2Parameters) int {
	switch p.SBoxDegree {
	case 3:
		return 24 // koalabear
	case 7:
		return 27 // babybear
	default:
		panic(fmt.Sprintf("unsupported SBoxDegree %d: expected 3 or 7", p.SBoxDegree))
	}
}

func (f *fieldHelper) loadQ() {
	f.qd = f.registers.PopV()
	f.MOVD("$const_q", amd64.AX)
	f.VPBROADCASTD(amd64.AX, f.qd)
}

func (f *fieldHelper) loadQInvNeg() {
	f.qInvNeg = f.registers.PopV()
	f.MOVD("$const_qInvNeg", amd64.AX)
	f.VPBROADCASTD(amd64.AX, f.qInvNeg)
}

type width int

const (
	fX width = iota
	fY
	fZ
)

// add a and b and store the result in into
func (f *fieldHelper) add(a, b, into amd64.VectorRegister, width ...width) {
	qd := f.qd
	r0 := f.registers.PopV()

	if len(width) > 0 {
		switch width[0] {
		case fX:
			qd = qd.X()
			r0 = r0.X()
			a = a.X()
			b = b.X()
			into = into.X()
		case fY:
			qd = qd.Y()
			r0 = r0.Y()
			a = a.Y()
			b = b.Y()
			into = into.Y()
		case fZ:
			qd = qd.Z()
			r0 = r0.Z()
			a = a.Z()
			b = b.Z()
			into = into.Z()
		default:
			panic("unknown field width")
		}
	}

	f.Define("add", 5, func(args ...any) {
		a := args[0]
		b := args[1]
		qd := args[2]
		r0 := args[3]
		into := args[4]

		f.VPADDD(b, a, into)
		f.VPSUBD(qd, into, r0)
		f.VPMINUD(into, r0, into)
	}, true)(a, b, qd, r0, into)
	f.registers.PushV(r0)
}

func (f *fieldHelper) addNoReduce(a, b, into amd64.VectorRegister) {
	f.VPADDD(b, a, into)
}

// double a and store the result in into
func (f *fieldHelper) double(a, into amd64.VectorRegister) {
	r0 := f.registers.PopV()
	f.Define("double", 4, func(args ...any) {
		a := args[0]
		qd := args[1]
		r0 := args[2]
		into := args[3]

		f.VPSLLD("$1", a, into)
		f.VPSUBD(qd, into, r0)
		f.VPMINUD(into, r0, into)
	}, true)(a, f.qd, r0, into)
	f.registers.PushV(r0)
}

// sub a and b and store the result in into
func (f *fieldHelper) sub(a, b, into amd64.VectorRegister) {
	r0 := f.registers.PopV()
	f.Define("sub", 5, func(args ...any) {
		a := args[0]
		b := args[1]
		qd := args[2]
		r0 := args[3]
		into := args[4]

		f.VPSUBD(b, a, into)
		f.VPADDD(qd, into, r0)
		f.VPMINUD(into, r0, into)
	}, true)(a, b, f.qd, r0, into)
	f.registers.PushV(r0)
}

// halve a and store the result in into
func (f *fieldHelper) halve(a, into amd64.VectorRegister) {
	ones := f.registers.PopV()

	f.Define("halve", 3, func(args ...any) {
		a := args[0]
		ones := args[1]
		qd := args[2]
		f.MOVD("$1", amd64.AX)
		f.VPBROADCASTD(amd64.AX, ones)

		f.VPTESTMD(a, ones, amd64.K4)
		// if a & 1 == 1 ; we add q;
		f.VPADDDk(a, qd, a, amd64.K4)
		// we shift right
		f.VPSRLD(1, a, a)
	}, true)(a, ones, f.qd)
	f.registers.PushV(ones)
}

func (f *fieldHelper) mul(a, b, into amd64.VectorRegister, reduce bool) {
	if f.registers.AvailableV() >= 5 {
		f.mul_5(a, b, into, reduce)
	} else {
		f.mul_4(a, b, into, reduce)
	}
}

// mul_4 a and b and store the result in into
// this version uses only 4 temporary registers
// see mul_5 for the version with 5 temporary registers
// see mul_6 for the version with 6 temporary registers
func (f *fieldHelper) mul_4(a, b, into amd64.VectorRegister, reduce bool) {
	t := f.registers.PopVN(4)

	f.Define("mul_4w", 9, func(args ...any) {
		a := args[0]
		b := args[1]
		aOdd := args[2]
		bOdd := args[3]
		t0 := args[4]
		t1 := args[5]
		c := args[6]
		qd := args[7]
		qInvNeg := args[8]

		PL0 := aOdd
		PL1 := bOdd

		f.VMOVSHDUP(a, aOdd) // keep high 32 bits (port 5 instead of port 0/1)
		f.VMOVSHDUP(b, bOdd) // keep high 32 bits (port 5 instead of port 0/1)

		// VPMULUDQ conveniently ignores the high 32 bits of each QWORD lane
		f.VPMULUDQ(a, b, t0)
		f.VPMULUDQ(aOdd, bOdd, t1)
		f.VPMULUDQ(t0, qInvNeg, PL0)
		f.VPMULUDQ(t1, qInvNeg, PL1)

		f.VPMULUDQ(PL0, qd, PL0)
		f.VPADDQ(t0, PL0, t0)

		f.VPMULUDQ(PL1, qd, PL1)
		f.VPADDQ(t1, PL1, c)

		f.VMOVSHDUPk(t0, amd64.K3, c)
	}, true)(a, b, t[0], t[1], t[2], t[3], into, f.qd, f.qInvNeg)
	f.registers.PushV(t...)

	if reduce {
		f.reduce1Q(into)
	}
}

func (f *fieldHelper) mul_5(a, b, into amd64.VectorRegister, reduce bool) {
	t := f.registers.PopVN(5)
	// same as mul_4, except we don't reuse aOdd for PL0

	f.Define("mul_5w", 10, func(args ...any) {
		a := args[0]
		b := args[1]
		aOdd := args[2]
		bOdd := args[3]
		t0 := args[4]
		t1 := args[5]
		PL0 := args[6]
		c := args[7]
		qd := args[8]
		qInvNeg := args[9]

		f.VMOVSHDUP(a, aOdd) // keep high 32 bits (port 5 instead of port 0/1)
		f.VMOVSHDUP(b, bOdd) // keep high 32 bits (port 5 instead of port 0/1)

		// VPMULUDQ conveniently ignores the high 32 bits of each QWORD lane
		f.VPMULUDQ(a, b, t0)
		f.VPMULUDQ(aOdd, bOdd, t1)
		f.VPMULUDQ(t0, qInvNeg, PL0)
		PL1 := bOdd
		f.VPMULUDQ(t1, qInvNeg, PL1)

		f.VPMULUDQ(PL0, qd, PL0)
		f.VPADDQ(t0, PL0, t0)

		f.VPMULUDQ(PL1, qd, PL1)
		f.VPADDQ(t1, PL1, c)

		f.VMOVSHDUPk(t0, amd64.K3, c)
	}, true)(a, b, t[0], t[1], t[2], t[3], t[4], into, f.qd, f.qInvNeg)
	f.registers.PushV(t...)

	if reduce {
		f.reduce1Q(into)
	}
}

// mul2ExpNegN multiplies a by 2^{-N} (i.e. divides by 2^N) modulo q.
// The "NegN" refers to the negative exponent, not a negation of the result.
//
// Uses a shift-based approach inspired by Plonky3 that avoids all VPMULUDQ
// multiplications.  For a prime of the form q = r·2^j + 1 with r = 2^k − 1
// (koalabear: r = 127, j = 24, k = 7; babybear: r = 15, j = 27, k = 4)
// the identity
//
//	x·2^{−N} ≡ x_hi − r·2^{j−N}·x_lo  (mod q)
//	         = x_hi + x_lo·2^{j−N} − x_lo·2^{k+j−N}
//
// where x_lo = x & ((1<<N)−1), x_hi = x >> N, and k+j = 31 for all
// supported F31 primes, lets us compute the result entirely with shifts,
// an addition, a subtraction, and a final sign correction, using only
// 2 temporary vector registers (down from 5) and 0 multiplications (down from 4).
func (f *fieldHelper) mul2ExpNegN(a amd64.VectorRegister, N int, into amd64.VectorRegister) {
	j := f.twoAdicity // e.g. 24 for koalabear, 27 for babybear
	k := 31 - j       // e.g. 7 for koalabear, 4 for babybear

	t := f.registers.PopVN(2)
	hi := t[0]
	lo := t[1]

	n := "$" + strconv.Itoa(N)
	compN := "$" + strconv.Itoa(32-N)
	if N == j {
		f.Define("mul_2_exp_neg_j", 8, func(args ...any) {
			a := args[0]
			into := args[1]
			hi := args[2]
			lo := args[3]
			qd := args[4]
			n := args[5]
			compN := args[6]
			k := args[7]

			f.VPSRLD(n, a, hi)
			f.VPSLLD(compN, a, lo)
			f.VPSRLD(compN, lo, lo)
			f.VPADDD(hi, lo, into)
			f.VPSLLD(k, lo, lo)
			f.VPSUBD(lo, into, into)
			f.VPADDD(qd, into, hi)
			f.VPMINUD(into, hi, into)
		}, true)(a, into, hi, lo, f.qd, n, compN, "$"+strconv.Itoa(k))
	} else {
		s1 := "$" + strconv.Itoa(31-N) // shift for the subtracted term (k+j−N)
		s2 := "$" + strconv.Itoa(j-N)  // shift for the added term (j−N)
		f.Define("mul_2_exp_neg_n", 9, func(args ...any) {
			a := args[0]
			into := args[1]
			hi := args[2]
			lo := args[3]
			qd := args[4]
			n := args[5]
			compN := args[6]
			s1 := args[7]
			s2 := args[8]

			f.VPSRLD(n, a, hi)
			f.VPSLLD(compN, a, lo)
			f.VPSRLD(compN, lo, lo)
			f.VPSLLD(s2, lo, into)
			f.VPADDD(hi, into, into)
			f.VPSLLD(s1, lo, lo)
			f.VPSUBD(lo, into, into)
			f.VPADDD(qd, into, hi)
			f.VPMINUD(into, hi, into)
		}, true)(a, into, hi, lo, f.qd, n, compN, s1, s2)
	}

	f.registers.PushV(t...)
}

// mul2ExpNeg8 is a specialised version of mul2ExpNegN for N = 8.
// It uses VPMADDUBSW to combine the byte extraction and multiplication by the
// odd factor r in a single instruction, saving 3 instructions over the generic
// shift-based path (6 instructions + 1 temp vs 9 instructions + 2 temps).
//
// rConst must be a pre-loaded vector of VPBROADCASTD(r) where r is the odd
// factor of q − 1 (127 for koalabear, 15 for babybear).
func (f *fieldHelper) mul2ExpNeg8(a, rConst, into amd64.VectorRegister) {
	j := f.twoAdicity
	hi := f.registers.PopV()
	f.Define("mul_2_exp_neg_8", 6, func(args ...any) {
		a := args[0]
		rConst := args[1]
		into := args[2]
		hi := args[3]
		qd := args[4]
		jMinus8 := args[5]

		f.VPSRLD("$8", a, hi)
		// In Go's x86 assembler syntax, `VPMADDUBSW rConst, a, into` uses `a` as
		// the unsigned-byte input and `rConst` as the signed-byte input.
		f.WriteLn(fmt.Sprintf("\tVPMADDUBSW %s, %s, %s", rConst, a, into))
		f.VPSLLD(jMinus8, into, into)
		f.VPSUBD(into, hi, into)
		f.VPADDD(qd, into, hi)
		f.VPMINUD(into, hi, into)
	}, true)(a, rConst, into, hi, f.qd, "$"+strconv.Itoa(j-8))

	f.registers.PushV(hi)
}

// reduce1Q reduces a by q and stores the result in into
func (f *fieldHelper) reduce1Q(a amd64.VectorRegister) {
	r0 := f.registers.PopV()
	f.Define("reduce1Q", 3, func(args ...any) {
		qd := args[0]
		c := args[1]
		r0 := args[2]

		f.VPSUBD(qd, c, r0)
		f.VPMINUD(c, r0, c)
	}, true)(f.qd, a, r0)
	f.registers.PushV(r0)
}

func (f *fieldHelper) matMul4(v []amd64.VectorRegister, nbBlocks int) {
	t01 := f.registers.PopV()
	t23 := f.registers.PopV()
	t0123 := f.registers.PopV()
	t01123 := f.registers.PopV()
	t01233 := f.registers.PopV()

	// for each block in v
	for i := range nbBlocks {
		// t01.Add(&s[4*i], &s[4*i+1])
		f.add(v[4*i], v[4*i+1], t01)
		// t23.Add(&s[4*i+2], &s[4*i+3])
		f.add(v[4*i+2], v[4*i+3], t23)
		// t0123.Add(&t01, &t23)
		f.add(t01, t23, t0123)
		// t01123.Add(&t0123, &s[4*i+1])
		f.add(t0123, v[4*i+1], t01123)
		// t01233.Add(&t0123, &s[4*i+3])
		f.add(t0123, v[4*i+3], t01233)

		// s[4*i+3].Double(&s[4*i]).Add(&s[4*i+3], &t01233)
		f.double(v[4*i], v[4*i+3])
		f.add(v[4*i+3], t01233, v[4*i+3])

		// s[4*i+1].Double(&s[4*i+2]).Add(&s[4*i+1], &t01123)
		f.double(v[4*i+2], v[4*i+1])
		f.add(v[4*i+1], t01123, v[4*i+1])

		// s[4*i].Add(&t01, &t01123)
		f.add(t01, t01123, v[4*i])

		// s[4*i+2].Add(&t23, &t01233)
		f.add(t23, t01233, v[4*i+2])
	}

	f.registers.PushV(t01, t23, t0123, t01123, t01233)
}

func (f *fieldHelper) matMulExternal(v []amd64.VectorRegister, nbBlocks int) {
	f.matMul4(v, nbBlocks)

	tmp0 := f.registers.PopV()
	tmp1 := f.registers.PopV()
	tmp2 := f.registers.PopV()
	tmp3 := f.registers.PopV()

	f.add(v[0], v[4], tmp0)
	f.add(v[1], v[5], tmp1)
	f.add(v[2], v[6], tmp2)
	f.add(v[3], v[7], tmp3)

	for i := 2; i < nbBlocks; i++ {
		f.add(tmp0, v[4*i], tmp0)
		f.add(tmp1, v[4*i+1], tmp1)
		f.add(tmp2, v[4*i+2], tmp2)
		f.add(tmp3, v[4*i+3], tmp3)
	}

	for i := range nbBlocks {
		f.add(v[4*i], tmp0, v[4*i])
		f.add(v[4*i+1], tmp1, v[4*i+1])
		f.add(v[4*i+2], tmp2, v[4*i+2])
		f.add(v[4*i+3], tmp3, v[4*i+3])
	}

	f.registers.PushV(tmp0, tmp1, tmp2, tmp3)
}

func (f *fieldHelper) sbox(a, into amd64.VectorRegister, degree int) {
	switch degree {
	case 3:
		f.sbox3(a, into)
	case 7:
		t0 := f.registers.PopV()
		t1 := f.registers.PopV()
		f.mul(a, a, t0, true)    // a^2 (reduce to prevent overflow in a^6 = a^3 * a^3)
		f.mul(t0, a, t0, false)  // a^3
		f.mul(t0, t0, t1, false) // a^6
		f.mul(t1, a, into, true) // a^7
		f.registers.PushV(t0, t1)
	default:
		panic(fmt.Sprintf("unsupported sbox degree %d: expected 3 or 7", degree))
	}
}

// sbox3 computes a^3 using a fused two-multiplication sequence that saves
// 2 instructions compared to two independent mul calls:
//  1. Caches aOdd (= VMOVSHDUP(a)) across both multiplies, eliminating
//     the redundant extraction in the square and reusing it in the cube.
//  2. Uses VMOVSHDUP (port 5) instead of VPSRLQ (port 0/1) for odd-lane
//     extraction, improving execution port balance.
func (f *fieldHelper) sbox3(a, into amd64.VectorRegister) {
	t := f.registers.PopVN(5)
	f.Define("sbox3", 9, func(args ...any) {
		a := args[0]
		into := args[1]
		aOdd := args[2]
		t0 := args[3]
		t1 := args[4]
		PL0 := args[5]
		PL1 := args[6]
		qd := args[7]
		qInvNeg := args[8]

		f.VMOVSHDUP(a, aOdd)
		f.VPMULUDQ(a, a, t0)
		f.VPMULUDQ(aOdd, aOdd, t1)
		f.VPMULUDQ(t0, qInvNeg, PL0)
		f.VPMULUDQ(t1, qInvNeg, PL1)
		f.VPMULUDQ(PL0, qd, PL0)
		f.VPADDQ(t0, PL0, t0)
		f.VPMULUDQ(PL1, qd, PL1)
		f.VPADDQ(t1, PL1, t1)
		f.VMOVSHDUPk(t0, amd64.K3, t1)
		f.VMOVSHDUP(t1, PL1)
		f.VPMULUDQ(a, t1, t0)
		f.VPMULUDQ(aOdd, PL1, PL0)
		f.VPMULUDQ(t0, qInvNeg, PL1)
		f.VPMULUDQ(PL0, qInvNeg, t1)
		f.VPMULUDQ(PL1, qd, PL1)
		f.VPADDQ(t0, PL1, t0)
		f.VPMULUDQ(t1, qd, t1)
		f.VPADDQ(PL0, t1, into)
		f.VMOVSHDUPk(t0, amd64.K3, into)
		f.VPSUBD(qd, into, PL0)
		f.VPMINUD(into, PL0, into)
	}, true)(a, into, t[0], t[1], t[2], t[3], t[4], f.qd, f.qInvNeg)

	f.registers.PushV(t...)
}
