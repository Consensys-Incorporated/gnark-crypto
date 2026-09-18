// Copyright 2020-2026 Consensys Software Inc.
// Licensed under the Apache License, Version 2.0. See the LICENSE file for details.

// AVX-512IFMA vectorized arithmetic for MamaBear: p = 2^49 − 2^34 + 1, R = 2^52.
//
// Corrected Montgomery multiply (Algorithm 1, MamaBearZKP paper):
//   c₀   = a·b mod 2^52            (VPMADD52LUQ into zeroed acc)
//   t    = ⌊a·b / 2^52⌋            (VPMADD52HUQ into acc pre-loaded with p)
//   c₁   = p + t
//   m₀   = c₀·(-p⁻¹) mod 2^52     (VPMADD52LUQ into zeroed acc)
//   Z2  += low52(m₀·p)             (Z2 = c₀ + low52(m₀·p) ∈ {0, R})
//   c₁  += ⌊m₀·p / 2^52⌋          (VPMADD52HUQ into c₁)
//   carry= Z2 >> 52 ∈ {0,1}
//   r    = c₁ + carry - p          (lazy result ∈ [0, 9p/8))
//   r   -= p if r >= p              (canonical result ∈ [0, p))
//
// All vector functions process blockSize=8 elements per 512-bit ZMM register.
// Constants const_q and const_qInvNeg come from go_asm.h (auto-generated from
// the package constants q and qInvNeg in element.go).

#include "textflag.h"
#include "funcdata.h"
#include "go_asm.h"

// addVec(res, a, b *Element, n uint64)
// res[i] = (a[i] + b[i]) mod p for i in [0, 8*n).
// n is the number of 8-element ZMM blocks.
TEXT ·addVec(SB), NOSPLIT, $0-32
	MOVQ         $const_q, AX
	VPBROADCASTQ AX, Z3
	MOVQ         res+0(FP), CX
	MOVQ         a+8(FP), R14
	MOVQ         b+16(FP), DX
	MOVQ         n+24(FP), BX

loop_1:
	TESTQ     BX, BX
	JEQ       done_2
	DECQ      BX
	VMOVDQU64 0(R14), Z0
	VMOVDQU64 0(DX), Z1
	VPADDQ    Z1, Z0, Z0    // z = a + b
	VPSUBQ    Z3, Z0, Z2    // t = z - p (wraps if z < p)
	VPMINUQ   Z0, Z2, Z0   // z = min(z, t); unsigned min gives correct mod-p result
	VMOVDQU64 Z0, 0(CX)

	ADDQ $64, R14
	ADDQ $64, DX
	ADDQ $64, CX
	JMP  loop_1

done_2:
	RET

// subVec(res, a, b *Element, n uint64)
// res[i] = (a[i] - b[i]) mod p for i in [0, 8*n).
TEXT ·subVec(SB), NOSPLIT, $0-32
	MOVQ         $const_q, AX
	VPBROADCASTQ AX, Z3
	MOVQ         res+0(FP), CX
	MOVQ         a+8(FP), R14
	MOVQ         b+16(FP), DX
	MOVQ         n+24(FP), BX

loop_3:
	TESTQ     BX, BX
	JEQ       done_4
	DECQ      BX
	VMOVDQU64 0(R14), Z0
	VMOVDQU64 0(DX), Z1
	VPSUBQ    Z1, Z0, Z0    // z = a - b (wraps if a < b)
	VPADDQ    Z3, Z0, Z2    // t = z + p (wraps back to [0,p) on underflow)
	VPMINUQ   Z0, Z2, Z0   // z = min(z, t); picks corrected value on underflow
	VMOVDQU64 Z0, 0(CX)

	ADDQ $64, R14
	ADDQ $64, DX
	ADDQ $64, CX
	JMP  loop_3

done_4:
	RET

// mulVec(res, a, b *Element, n uint64)
// res[i] = a[i] * b[i] * R^{-1} mod p for i in [0, 8*n).
// Inputs must be in canonical form [0, p).  Output is canonical.
TEXT ·mulVec(SB), NOSPLIT, $0-32
	MOVQ         $const_q, AX
	VPBROADCASTQ AX, Z16           // Z16 = broadcast(p)
	MOVQ         $const_qInvNeg, AX
	VPBROADCASTQ AX, Z17           // Z17 = broadcast(-p^{-1} mod 2^52)
	MOVQ         res+0(FP), CX
	MOVQ         a+8(FP), R14
	MOVQ         b+16(FP), DX
	MOVQ         n+24(FP), BX

loop_5:
	TESTQ     BX, BX
	JEQ       done_6
	DECQ      BX

	VMOVDQU64 0(R14), Z0           // a[0..7]
	VMOVDQU64 0(DX), Z1            // b[0..7]

	// c₀ = a·b mod 2^52
	VPXORQ       Z2, Z2, Z2
	VPMADD52LUQ  Z1, Z0, Z2       // Z2 = c₀

	// c₁ = p + floor(a·b / 2^52)
	VMOVDQA64    Z16, Z3
	VPMADD52HUQ  Z1, Z0, Z3       // Z3 = p + t

	// m₀ = c₀ · (-p^{-1}) mod 2^52
	VPXORQ       Z4, Z4, Z4
	VPMADD52LUQ  Z17, Z2, Z4      // Z4 = m₀

	// Z2 = c₀ + low52(m₀·p) ∈ {0, R=2^52} (for carry extraction)
	VPMADD52LUQ  Z16, Z4, Z2      // Z2 += low52(m₀ · p)

	// c₁ += floor(m₀·p / 2^52)
	VPMADD52HUQ  Z16, Z4, Z3      // Z3 = p + t + qVal

	// carry = Z2 >> 52 ∈ {0, 1}
	VPSRLQ       $52, Z2, Z2
	VPADDQ       Z2, Z3, Z0       // Z0 = p + t + qVal + carry

	// Unconditional subtract (removes the initial p): Z0 ∈ [0, 9p/8)
	VPSUBQ       Z16, Z0, Z0
	// Conditional subtract: Z2 = Z0 - p (wraps if Z0 < p); pick min
	VPSUBQ       Z16, Z0, Z2
	VPMINUQ      Z0, Z2, Z0      // Z0 ∈ [0, p)

	VMOVDQU64 Z0, 0(CX)

	ADDQ $64, R14
	ADDQ $64, DX
	ADDQ $64, CX
	JMP  loop_5

done_6:
	RET

// scalarMulVec(res, a, b *Element, n uint64)
// b points to a single scalar; res[i] = a[i] * (*b) * R^{-1} mod p for i in [0, 8*n).
TEXT ·scalarMulVec(SB), NOSPLIT, $0-32
	MOVQ         $const_q, AX
	VPBROADCASTQ AX, Z16
	MOVQ         $const_qInvNeg, AX
	VPBROADCASTQ AX, Z17
	MOVQ         res+0(FP), CX
	MOVQ         a+8(FP), R14
	MOVQ         b+16(FP), DX
	MOVQ         n+24(FP), BX
	VPBROADCASTQ 0(DX), Z1         // Z1 = broadcast(*b)

loop_7:
	TESTQ     BX, BX
	JEQ       done_8
	DECQ      BX

	VMOVDQU64    0(R14), Z0

	VPXORQ       Z2, Z2, Z2
	VPMADD52LUQ  Z1, Z0, Z2       // c₀

	VMOVDQA64    Z16, Z3
	VPMADD52HUQ  Z1, Z0, Z3       // p + t

	VPXORQ       Z4, Z4, Z4
	VPMADD52LUQ  Z17, Z2, Z4      // m₀

	VPMADD52LUQ  Z16, Z4, Z2      // Z2 = c₀ + low52(m₀·p) ∈ {0, R}
	VPMADD52HUQ  Z16, Z4, Z3      // Z3 = p + t + qVal

	VPSRLQ       $52, Z2, Z2      // carry
	VPADDQ       Z2, Z3, Z0       // p + t + qVal + carry
	VPSUBQ       Z16, Z0, Z0      // t + qVal + carry ∈ [0, 9p/8)
	VPSUBQ       Z16, Z0, Z2      // tentative second subtract
	VPMINUQ      Z0, Z2, Z0      // conditional: Z0 ∈ [0, p)

	VMOVDQU64 Z0, 0(CX)

	ADDQ $64, R14
	ADDQ $64, CX
	JMP  loop_7

done_8:
	RET

// sumVec(t *uint64, a *Element, n uint64)
// Accumulates into t[0..7]: t[i] += sum of a[j] for j≡i (mod 8), j in [0, 8*n).
// Caller must ensure accumulated values do not exceed 2^64 (use chunk-and-reduce).
TEXT ·sumVec(SB), NOSPLIT, $0-24
	MOVQ      t+0(FP), R14
	MOVQ      a+8(FP), R13
	MOVQ      n+16(FP), CX
	VMOVDQU64 0(R14), Z0           // load existing accumulator

loop_9:
	TESTQ     CX, CX
	JEQ       done_10
	DECQ      CX
	VMOVDQU64 0(R13), Z1
	VPADDQ    Z1, Z0, Z0

	ADDQ $64, R13
	JMP  loop_9

done_10:
	VMOVDQU64 Z0, 0(R14)
	RET

// innerProdVec(t *uint64, a, b *Element, n uint64)
// Accumulates into t[0..7]: t[i] += sum_j a[j*8+i] * b[j*8+i] * R^{-1} mod p.
// Each Montgomery product is reduced to [0, p) before accumulation.
// Caller must ensure accumulated values do not exceed 2^64 (use chunk-and-reduce).
TEXT ·innerProdVec(SB), NOSPLIT, $0-32
	MOVQ         $const_q, AX
	VPBROADCASTQ AX, Z16
	MOVQ         $const_qInvNeg, AX
	VPBROADCASTQ AX, Z17
	MOVQ         t+0(FP), R14
	MOVQ         a+8(FP), R13
	MOVQ         b+16(FP), DX
	MOVQ         n+24(FP), BX
	VMOVDQU64    0(R14), Z0        // load existing accumulator

loop_11:
	TESTQ     BX, BX
	JEQ       done_12
	DECQ      BX

	VMOVDQU64    0(R13), Z1
	VMOVDQU64    0(DX), Z2

	VPXORQ       Z3, Z3, Z3
	VPMADD52LUQ  Z2, Z1, Z3       // c₀

	VMOVDQA64    Z16, Z4
	VPMADD52HUQ  Z2, Z1, Z4      // p + t

	VPXORQ       Z5, Z5, Z5
	VPMADD52LUQ  Z17, Z3, Z5     // m₀

	VPMADD52LUQ  Z16, Z5, Z3     // Z3 = c₀ + low52(m₀·p) ∈ {0, R}
	VPMADD52HUQ  Z16, Z5, Z4     // Z4 = p + t + qVal

	VPSRLQ       $52, Z3, Z3     // carry
	VPADDQ       Z3, Z4, Z3      // p + t + qVal + carry
	VPSUBQ       Z16, Z3, Z3     // t + qVal + carry ∈ [0, 9p/8)
	VPSUBQ       Z16, Z3, Z5     // tentative second subtract
	VPMINUQ      Z3, Z5, Z3     // conditional: product ∈ [0, p)

	VPADDQ    Z3, Z0, Z0          // accumulate

	ADDQ $64, R13
	ADDQ $64, DX
	JMP  loop_11

done_12:
	VMOVDQU64 Z0, 0(R14)
	RET
