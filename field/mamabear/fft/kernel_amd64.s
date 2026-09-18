//go:build !purego

// Copyright 2020-2026 Consensys Software Inc.
// Licensed under the Apache License, Version 2.0. See the LICENSE file for details.

// AVX-512IFMA FFT butterfly kernel for MamaBear: p = 2^49 − 2^34 + 1, R = 2^52.
// Each ZMM register holds 8 uint64 elements (8 × 64-bit = 512 bits).
//
// BUTTERFLYQ: canonical DIF/DIT butterfly over 64-bit field elements.
//   Given A, B ∈ [0, p):
//     new_A = (A + B) mod p  (VPADDQ + conditional subtract via VPMINUQ)
//     new_B = (A − B) mod p  (VPSUBQ + conditional add via VPMINUQ)
//
// MULMB: vectorised Montgomery multiply, Algorithm 1 from the MamaBearZKP paper.
//   Computes A = A · B · R^{-1} mod p using VPMADD52LUQ/HUQ.
//   13 instructions; result is canonical ∈ [0, p).

#include "textflag.h"
#include "funcdata.h"
#include "go_asm.h"

// BUTTERFLYQ(A, B, Q, T1, T2)
// A, B : ZMM in/out (canonical [0,p)); Q : ZMM broadcast(p) read-only;
// T1, T2 : ZMM temporaries (clobbered).
// After: A = (A+B) mod p, B = (A_old−B) mod p.
#define BUTTERFLYQ(A, B, Q, T1, T2) \
	VPADDQ  B, A, T1  \
	VPSUBQ  B, A, B   \
	VPSUBQ  Q, T1, T2 \
	VPMINUQ T1, T2, A \
	VPADDQ  Q, B, T1  \
	VPMINUQ B, T1, B  \

// MULMB(A, B, Q, QINV, C0, C1, M0)
// A : ZMM in/out (result = A·B·R^{-1} mod p);
// B : ZMM read-only multiplier;
// Q : ZMM broadcast(p) read-only; QINV : ZMM broadcast(qInvNeg) read-only;
// C0, C1, M0 : ZMM temporaries (clobbered).
#define MULMB(A, B, Q, QINV, C0, C1, M0)  \
	VPXORQ       C0, C0, C0         \
	VPMADD52LUQ  B, A, C0           \
	VMOVDQA64    Q, C1              \
	VPMADD52HUQ  B, A, C1           \
	VPXORQ       M0, M0, M0         \
	VPMADD52LUQ  QINV, C0, M0      \
	VPMADD52LUQ  Q, M0, C0         \
	VPMADD52HUQ  Q, M0, C1         \
	VPSRLQ       $52, C0, C0       \
	VPADDQ       C0, C1, A         \
	VPSUBQ       Q, A, A           \
	VPSUBQ       Q, A, C0          \
	VPMINUQ      A, C0, A          \

// innerDIFWithTwiddles_avx512(a, twiddles *Element, start, end, m int)
// DIF butterfly: for each 8-element block i in [0, end/8):
//   butterfly(a[i], a[i+m]) then a[i+m] *= twiddles[i]   (Montgomery)
// start is ignored (always 0 in the SIMD dispatch path).
// Register map: Z0=a[i], Z1=a[i+m], Z2=twiddles[i]; Z3,Z4 BUTTERFLYQ temps;
//               Z5,Z6,Z7 MULMB temps; Z16=q, Z17=qInvNeg.
TEXT ·innerDIFWithTwiddles_avx512(SB), NOSPLIT, $0-40
	MOVQ         $const_q, AX
	VPBROADCASTQ AX, Z16
	MOVQ         $const_qInvNeg, AX
	VPBROADCASTQ AX, Z17
	MOVQ         a+0(FP), R14
	MOVQ         twiddles+8(FP), DX
	MOVQ         end+24(FP), CX
	MOVQ         m+32(FP), BX
	SHRQ         $3, CX             // CX = end / 8 (blocks of 8 elements)
	SHLQ         $3, BX             // BX = m * 8 bytes
	MOVQ         R14, SI
	ADDQ         BX, SI             // SI = &a[m]

loop_dif:
	TESTQ     CX, CX
	JEQ       done_dif
	DECQ      CX
	VMOVDQU64 0(R14), Z0
	VMOVDQU64 0(SI), Z1
	VMOVDQU64 0(DX), Z2
	BUTTERFLYQ(Z0, Z1, Z16, Z3, Z4)
	MULMB(Z1, Z2, Z16, Z17, Z5, Z6, Z7)
	VMOVDQU64 Z0, 0(R14)
	VMOVDQU64 Z1, 0(SI)
	ADDQ      $64, R14
	ADDQ      $64, SI
	ADDQ      $64, DX
	JMP       loop_dif

done_dif:
	RET

// innerDITWithTwiddles_avx512(a, twiddles *Element, start, end, m int)
// DIT butterfly: for each 8-element block i in [0, end/8):
//   a[i+m] *= twiddles[i]   (Montgomery) then butterfly(a[i], a[i+m])
// start is ignored (always 0 in the SIMD dispatch path).
// Same register map as DIF above.
TEXT ·innerDITWithTwiddles_avx512(SB), NOSPLIT, $0-40
	MOVQ         $const_q, AX
	VPBROADCASTQ AX, Z16
	MOVQ         $const_qInvNeg, AX
	VPBROADCASTQ AX, Z17
	MOVQ         a+0(FP), R14
	MOVQ         twiddles+8(FP), DX
	MOVQ         end+24(FP), CX
	MOVQ         m+32(FP), BX
	SHRQ         $3, CX
	SHLQ         $3, BX
	MOVQ         R14, SI
	ADDQ         BX, SI

loop_dit:
	TESTQ     CX, CX
	JEQ       done_dit
	DECQ      CX
	VMOVDQU64 0(R14), Z0
	VMOVDQU64 0(SI), Z1
	VMOVDQU64 0(DX), Z2
	MULMB(Z1, Z2, Z16, Z17, Z5, Z6, Z7)
	BUTTERFLYQ(Z0, Z1, Z16, Z3, Z4)
	VMOVDQU64 Z0, 0(R14)
	VMOVDQU64 Z1, 0(SI)
	ADDQ      $64, R14
	ADDQ      $64, SI
	ADDQ      $64, DX
	JMP       loop_dit

done_dit:
	RET
