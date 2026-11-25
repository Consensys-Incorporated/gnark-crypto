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
	"strconv"

	"github.com/consensys/bavard/amd64"
)

// generateRadix8VecE6 emits vectorRadix8DIT_E6_avx512 / vectorRadix8DIF_E6_avx512:
//
//	func name(a *E6, twiddles *fr.Element, N uint64)
//
// Each runs the m=4, m=2 and m=1 FFT stages on every block of 8 consecutive
// E6 elements. DIT runs m=1 -> m=2 -> m=4 (multiply the second element of each
// pair by its twiddle, then butterfly); DIF runs m=4 -> m=2 -> m=1 (butterfly,
// then multiply).
//
// One iteration handles 8 blocks (64 E6) with no wasted lane. A block is 24
// qwords: qword q = 3k + c holds coordinates (A0, A1) of B_c of element e_k.
// The 8 blocks form an 8x24 qword matrix; three 8x8 qword transposes turn it
// into 24 zmm R[q], each holding qword q of the 8 blocks (lane pair b = block
// b). Every butterfly is then between two whole registers:
//
//	m=4: (R[q], R[q+12]), m=2: (R[q], R[q+6]), m=1: (R[q], R[q+3])
//
// and every twiddle is one scalar for the whole register.
//
// On the way back, a single unpack round leaves in each 128-bit chunk two
// consecutive qwords of one block, so the chunks are stored directly
// (VEXTRACTI64X2 to memory is a pure store) instead of running the full
// transpose.
//
// The DIT kernel prefetches about two iterations ahead (see the loop).
//
// twiddles points to 4 zmm built by the Go caller: t4[1], t4[2], t4[3] and
// t2[1], each broadcast to 16 lanes.
//
// Precondition: N % 64 == 0.
func (_f *FFAmd64) generateRadix8VecE6(dit bool) {
	name := "vectorRadix8DIF_E6_avx512"
	if dit {
		name = "vectorRadix8DIT_E6_avx512"
	}
	const argSize = 3 * 8
	stackSize := _f.StackSize(3, 0, 0)
	registers := _f.FnHeader(name, stackSize, argSize, amd64.DX, amd64.AX)
	defer _f.AssertCleanStack(stackSize, 0)
	f := &fieldHelper{FFAmd64: _f, registers: &registers}

	addrA := registers.Pop()
	addrTw := registers.Pop()
	N := registers.Pop()

	f.loadQ()
	f.loadQInvNeg()
	// K3: low dword of each qword, used by mul.
	f.MOVQ(uint64(0x5555), amd64.AX)
	f.KMOVD(amd64.AX, amd64.K3)

	f.MOVQ("a+0(FP)", addrA)
	f.MOVQ("twiddles+8(FP)", addrTw)
	f.MOVQ("N+16(FP)", N)
	f.SHRQ("$6", N)

	tw := func(i int) string { return addrTw.At(8 * i) }
	const (
		tw41 = iota // t4[1]
		tw42        // t4[2]
		tw43        // t4[3]
		tw21        // t2[1]
	)

	var R [24]amd64.VectorRegister
	for i := range R {
		R[i] = registers.PopV()
	}
	// spare register: each two-output step writes its first result here and
	// frees one of its inputs, so the register names rotate.
	sp := registers.PopV()

	// transpose8 transposes the 8x8 qword matrix held in x (row i in x[i]),
	// returning the registers holding its columns in order.
	transpose8 := func(x [8]amd64.VectorRegister) [8]amd64.VectorRegister {
		// (x[i], x[j]) <- (op1(x[i], x[j]), op2(x[i], x[j])), via sp
		step := func(i, j int, op1, op2 func(lo, hi, dst amd64.VectorRegister)) {
			op1(x[i], x[j], sp)
			op2(x[i], x[j], x[j])
			x[i], sp = sp, x[i]
		}
		unpackLo := func(lo, hi, dst amd64.VectorRegister) { f.VPUNPCKLQDQ(hi, lo, dst) }
		unpackHi := func(lo, hi, dst amd64.VectorRegister) { f.VPUNPCKHQDQ(hi, lo, dst) }
		// 128-bit chunks: [lo.c0, lo.c2, hi.c0, hi.c2] (even) or c1, c3 (odd)
		even := func(lo, hi, dst amd64.VectorRegister) { f.VSHUFI64X2("$0x88", hi, lo, dst) }
		odd := func(lo, hi, dst amd64.VectorRegister) { f.VSHUFI64X2("$0xdd", hi, lo, dst) }

		// x[2p] = cols {0,2,4,6}, x[2p+1] = cols {1,3,5,7} of rows 2p, 2p+1
		for p := 0; p < 8; p += 2 {
			step(p, p+1, unpackLo, unpackHi)
		}
		// x[h], x[h+2] = cols {h, 4+h}, {2+h, 6+h} of rows 0..3 (h = 0, 1),
		// and of rows 4..7 (h = 4, 5)
		for _, h := range []int{0, 1, 4, 5} {
			step(h, h+2, even, odd)
		}
		// x[h] = col h, x[h+4] = col h+4, of all 8 rows (h = 0..3)
		for _, h := range []int{0, 1, 2, 3} {
			step(h, h+4, even, odd)
		}
		return x
	}

	// (R[i], R[j]) <- (R[i]+R[j], R[i]-R[j]). If lazy, R[j] is left in
	// [0, 2q): it only feeds a twiddle multiplication, which is correct for
	// an input below 2q and reduces it.
	butterfly := func(i, j int, lazy bool) {
		f.add(R[i], R[j], sp)
		if lazy {
			f.VPSUBD(R[j], R[i], R[j])
			f.VPADDD(f.qd, R[j], R[j])
		} else {
			f.sub(R[i], R[j], R[j])
		}
		R[i], sp = sp, R[i]
	}
	// twiddled(q) is true for the registers multiplied right after the stage.
	stage := func(d int, twiddled func(q int) bool) {
		for q := 0; q < 24; q++ {
			if (q/d)%2 == 0 {
				butterfly(q, q+d, twiddled(q+d))
			}
		}
	}
	mul := func(from, to, w int) {
		for q := from; q < to; q++ {
			f.mulTwiddle(R[q], tw(w), R[q])
		}
	}
	// m=4: e5, e6, e7 (e4 has twiddle one); m=2: e3, e7 (e2, e6 have one).
	mul4 := func() { mul(15, 18, tw41); mul(18, 21, tw42); mul(21, 24, tw43) }
	mul2 := func() { mul(9, 12, tw21); mul(21, 24, tw21) }
	isMul4 := func(q int) bool { return q >= 15 }
	isMul2 := func(q int) bool { return (q >= 9 && q < 12) || q >= 21 }
	none := func(int) bool { return false }

	f.Loop(N, func() {
		if dit {
			// DIT runs this kernel first on each chunk, so at large sizes the
			// data comes from DRAM and the hardware prefetcher falls behind:
			// fetch 24 lines (one iteration's worth) 3584 bytes ahead.
			for i := range 24 {
				f.PREFETCHT0(strconv.Itoa(3584+64*i) + "(" + string(addrA) + ")")
			}
		}
		for j := range 3 {
			var x [8]amd64.VectorRegister
			for b := range 8 {
				x[b] = R[8*j+b]
				f.VMOVDQU32(addrA.At(24*b+8*j), x[b])
			}
			c := transpose8(x)
			copy(R[8*j:8*j+8], c[:])
		}

		if dit {
			stage(3, isMul2)
			mul2()
			stage(6, isMul4)
			mul4()
			stage(12, none)
		} else {
			stage(12, isMul4)
			mul4()
			stage(6, isMul2)
			mul2()
			stage(3, none)
		}

		// chunk i of unpackLo(R[q], R[q+1]) is qwords q, q+1 of block 2i;
		// unpackHi gives block 2i+1.
		for q := 0; q < 24; q += 2 {
			f.VPUNPCKLQDQ(R[q+1], R[q], sp)
			f.VPUNPCKHQDQ(R[q+1], R[q], R[q+1])
			R[q], sp = sp, R[q]
			for h := range 2 {
				for i := range 4 {
					dst := addrA.At(24*(2*i+h) + q)
					if i == 0 {
						f.VMOVDQU64(R[q+h].X(), dst)
					} else {
						f.VEXTRACTI64X2("$"+strconv.Itoa(i), R[q+h], dst)
					}
				}
			}
		}
		f.ADDQ("$1536", addrA)
	})

	f.RETAVX()
}

// mulTwiddle sets into = a*w (Montgomery, reduced), for a twiddle w whose two
// dwords in each qword are equal (here, a broadcast): the odd-lane copy of w is
// w itself, which saves the VMOVSHDUP of mul. w may be a register or a 64-byte
// memory operand. a may be in [0, 2q); into may be a.
func (f *fieldHelper) mulTwiddle(a amd64.VectorRegister, w any, into amd64.VectorRegister) {
	t := f.registers.PopVN(4)
	f.Define("mul_tw", 9, func(args ...any) {
		a := args[0]
		w := args[1]
		aOdd := args[2]
		t0 := args[3]
		t1 := args[4]
		pl0 := args[5]
		c := args[6]
		qd := args[7]
		qInvNeg := args[8]

		f.VMOVSHDUP(a, aOdd)
		f.VPMULUDQ(w, a, t0)
		f.VPMULUDQ(w, aOdd, t1)
		f.VPMULUDQ(t0, qInvNeg, pl0)
		f.VPMULUDQ(t1, qInvNeg, aOdd)
		f.VPMULUDQ(pl0, qd, pl0)
		f.VPADDQ(t0, pl0, t0)
		f.VPMULUDQ(aOdd, qd, aOdd)
		f.VPADDQ(t1, aOdd, c)
		f.VMOVSHDUPk(t0, amd64.K3, c)
	}, true)(a, w, t[0], t[1], t[2], t[3], into, f.qd, f.qInvNeg)
	f.registers.PushV(t...)
	f.reduce1Q(into)
}
