// Copyright 2020-2025 Consensys Incorporated
// Licensed under the Apache License, Version 2.0. See the LICENSE file for details.

package amd64

import (
	"fmt"

	"github.com/consensys/bavard/amd64"
)

// ifmaHelper wraps FFAmd64 and caches Define macros for IFMA operations
type ifmaHelper struct {
	*FFAmd64

	// Cached define functions for reuse
	carryProp  defineFn // carry propagation (extract carry, mask, add to next)
	borrowProp defineFn // borrow propagation (extract sign, add to next, mask)
	condSelect defineFn // conditional select using VPTERNLOGQ
	zeroReg    defineFn // zero a ZMM register

	// Frequently used vector registers for IFMA operations
	mask52    amd64.VectorRegister // Z31 - 52-bit mask
	qInvNeg52 amd64.VectorRegister // Z30 - qInvNeg52
	// q in radix-52: Z25=ql0, Z26=ql1, Z27=ql2, Z28=ql3, Z29=ql4
	qRadix52 [5]amd64.VectorRegister

	// Scratch registers for transpose operations
	scratch [4]amd64.VectorRegister // Z18-Z21
}

// newIfmaHelper creates a new IFMA helper with cached defines
func (f *FFAmd64) newIfmaHelper() *ifmaHelper {
	h := &ifmaHelper{
		FFAmd64:   f,
		mask52:    amd64.Z31,
		qInvNeg52: amd64.Z30,
		qRadix52:  [5]amd64.VectorRegister{amd64.Z25, amd64.Z26, amd64.Z27, amd64.Z28, amd64.Z29},
		scratch:   [4]amd64.VectorRegister{amd64.Z18, amd64.Z19, amd64.Z20, amd64.Z21},
	}
	h.initDefines()
	return h
}

// initDefines initializes all the reusable Define macros
func (h *ifmaHelper) initDefines() {
	// CARRY_PROP: Extract carry from limb, mask it, add carry to next limb
	// args: src, dst_mask, dst_next, mask52, tmp
	h.carryProp = h.Define("CARRY_PROP", 5, func(args ...any) {
		src := args[0]
		dstMask := args[1]
		dstNext := args[2]
		mask52 := args[3]
		tmp := args[4]
		h.VPSRLQ("$52", src, tmp)
		h.VPANDQ(mask52, src, dstMask)
		h.VPADDQ(tmp, dstNext, dstNext)
	}, true)

	// BORROW_PROP: Extract borrow (sign bit), add to next, mask current
	// args: src, dst_next, mask52, tmp
	h.borrowProp = h.Define("BORROW_PROP", 4, func(args ...any) {
		src := args[0]
		dstNext := args[1]
		mask52 := args[2]
		tmp := args[3]
		h.VPSRAQ("$63", src, tmp)
		h.VPANDQ(mask52, src, src)
		h.VPADDQ(tmp, dstNext, dstNext)
	}, true)

	// COND_SELECT: Conditional select using VPTERNLOGQ
	// if mask == all 1s, keep dst; else use src
	// args: src, mask, dst
	h.condSelect = h.Define("COND_SELECT", 3, func(args ...any) {
		src := args[0]
		mask := args[1]
		dst := args[2]
		h.VPTERNLOGQ("$0xE2", src, mask, dst)
	}, true)

	// ZERO_REG: Zero a ZMM register
	h.zeroReg = h.Define("ZERO_REG", 1, func(args ...any) {
		reg := args[0]
		h.VPXORQ(reg, reg, reg)
	}, true)
}

// generateMulVecIFMA generates AVX-512 IFMA based vector multiplication.
func (f *FFAmd64) generateMulVecIFMA() {
	h := f.newIfmaHelper()
	h.generateMulVecIFMABody("mulVec", false)
}

// generateScalarMulVecIFMA generates IFMA-based scalar multiplication
func (f *FFAmd64) generateScalarMulVecIFMA() {
	h := f.newIfmaHelper()
	h.generateMulVecIFMABody("scalarMulVec", true)
}

// generateMulVecIFMABody is the common implementation for mul and scalarMul
func (h *ifmaHelper) generateMulVecIFMABody(funcName string, scalarMul bool) {
	if scalarMul {
		h.Comment(fmt.Sprintf("%s(res, a, b *Element, n uint64)", funcName))
		h.Comment("Performs n scalar multiplications using AVX-512 IFMA instructions")
	} else {
		h.Comment(fmt.Sprintf("%s(res, a, b *Element, n uint64)", funcName))
		h.Comment("Performs n multiplications using AVX-512 IFMA instructions")
	}
	h.Comment("Processes 8 elements in parallel using radix-52 representation")

	const argSize = 4 * 8
	stackSize := 0
	registers := h.FnHeader(funcName, stackSize, argSize, amd64.AX, amd64.DX)
	defer h.AssertCleanStack(stackSize, 0)

	addrRes := h.Pop(&registers)
	addrA := h.Pop(&registers)
	addrB := h.Pop(&registers)
	n := h.Pop(&registers)

	h.MOVQ("res+0(FP)", addrRes)
	h.MOVQ("a+8(FP)", addrA)
	h.MOVQ("b+16(FP)", addrB)
	h.MOVQ("n+24(FP)", n)

	// Initialize constants
	h.Comment("Load constants for radix-52 conversion and reduction")
	h.initIFMAConstants()

	// A in radix-52: Z0-Z4
	// B in radix-52: Z5-Z9
	aRadix52 := [5]amd64.VectorRegister{amd64.Z0, amd64.Z1, amd64.Z2, amd64.Z3, amd64.Z4}
	bRadix52 := [5]amd64.VectorRegister{amd64.Z5, amd64.Z6, amd64.Z7, amd64.Z8, amd64.Z9}

	// Result in radix-64 after conversion: Z14-Z17
	resultRadix64 := [4]amd64.VectorRegister{amd64.Z14, amd64.Z15, amd64.Z16, amd64.Z17}

	h.Loop(n, func() {
		h.Comment("Process 8 elements in parallel")

		h.Comment("Load and convert 8 elements from a[] to radix-52")
		h.loadAndConvertToRadix52(addrA, aRadix52)

		if scalarMul {
			// Reload scalar each iteration since Barrett reduction clobbers Z5-Z9
			h.Comment("Load scalar and convert to radix-52 (broadcast)")
			h.loadScalarToRadix52(addrB, bRadix52)
		} else {
			h.Comment("Load and convert 8 elements from b[] to radix-52")
			h.loadAndConvertToRadix52(addrB, bRadix52)
		}

		h.Comment("Montgomery multiplication using IFMA (CIOS variant)")
		h.montgomeryMulIFMA(aRadix52, bRadix52)

		h.Comment("Barrett reduction from [0, 32q) to [0, q)")
		h.barrettReduction(aRadix52)

		h.Comment("Convert result from radix-52 back to radix-64")
		h.convertFromRadix52(aRadix52, resultRadix64)

		h.Comment("Transpose back to AoS format and store")
		h.transposeAndStore(addrRes, resultRadix64)

		h.Comment("Advance pointers")
		h.ADDQ("$256", addrA)
		if !scalarMul {
			h.ADDQ("$256", addrB)
		}
		h.ADDQ("$256", addrRes)
	})

	h.RETAVX()

	h.Push(&registers, addrRes, addrA, addrB, n)
}

// initIFMAConstants initializes the constants needed for IFMA operations
// Uses precomputed constants from Go code (qInvNeg52, qRadix52_0, etc.)
func (h *ifmaHelper) initIFMAConstants() {
	h.MOVQ("$0xFFFFFFFFFFFFF", amd64.R15, "52-bit mask in R15")
	h.VPBROADCASTQ(amd64.R15, h.mask52, "Z31 = mask52")

	// Use precomputed qInvNeg52 constant instead of runtime masking
	h.MOVQ("$const_qInvNeg52", amd64.AX)
	h.VPBROADCASTQ(amd64.AX, h.qInvNeg52, "Z30 = qInvNeg52")

	h.Comment("Load modulus in radix-52 form (precomputed)")
	h.loadModulusRadix52()
}

// loadModulusRadix52 loads the precomputed radix-52 form of q
func (h *ifmaHelper) loadModulusRadix52() {
	h.Comment("q in radix-52: Z25=ql0, Z26=ql1, Z27=ql2, Z28=ql3, Z29=ql4")
	// Use precomputed qRadix52_i constants
	for i := range 5 {
		h.MOVQ(fmt.Sprintf("$const_qRadix52_%d", i), amd64.AX)
		h.VPBROADCASTQ(amd64.AX, h.qRadix52[i])
	}
}

// loadScalarToRadix52 loads a scalar and broadcasts it to all lanes in radix-52
func (h *ifmaHelper) loadScalarToRadix52(addr amd64.Register, out [5]amd64.VectorRegister) {
	// Load the 4 words of the scalar
	h.MOVQ(fmt.Sprintf("0(%s)", addr), amd64.R9)
	h.MOVQ(fmt.Sprintf("8(%s)", addr), amd64.R10)
	h.MOVQ(fmt.Sprintf("16(%s)", addr), amd64.R11)
	h.MOVQ(fmt.Sprintf("24(%s)", addr), amd64.R12)

	// Convert to radix-52 and broadcast
	// l0 = a0 & mask52
	h.MOVQ(amd64.R9, amd64.R8)
	h.ANDQ(amd64.R15, amd64.R8)
	h.VPBROADCASTQ(amd64.R8, out[0])

	// l1 = (a0 >> 52) | (a1 << 12) & mask52
	h.SHRQ("$52", amd64.R9)
	h.MOVQ(amd64.R10, amd64.R8)
	h.SHLQ("$12", amd64.R8)
	h.ORQ(amd64.R9, amd64.R8)
	h.ANDQ(amd64.R15, amd64.R8)
	h.VPBROADCASTQ(amd64.R8, out[1])

	// l2 = (a1 >> 40) | (a2 << 24) & mask52
	h.SHRQ("$40", amd64.R10)
	h.MOVQ(amd64.R11, amd64.R8)
	h.SHLQ("$24", amd64.R8)
	h.ORQ(amd64.R10, amd64.R8)
	h.ANDQ(amd64.R15, amd64.R8)
	h.VPBROADCASTQ(amd64.R8, out[2])

	// l3 = (a2 >> 28) | (a3 << 36) & mask52
	h.SHRQ("$28", amd64.R11)
	h.MOVQ(amd64.R12, amd64.R8)
	h.SHLQ("$36", amd64.R8)
	h.ORQ(amd64.R11, amd64.R8)
	h.ANDQ(amd64.R15, amd64.R8)
	h.VPBROADCASTQ(amd64.R8, out[3])

	// l4 = a3 >> 16
	h.SHRQ("$16", amd64.R12)
	h.VPBROADCASTQ(amd64.R12, out[4])
}

// montgomeryMulIFMA performs Montgomery multiplication using IFMA
func (h *ifmaHelper) montgomeryMulIFMA(a, b [5]amd64.VectorRegister) {
	h.Comment("A = [Z0-Z4], B = [Z5-Z9], result in [Z0-Z4]")

	// Accumulators: T[0..5] in Z10-Z15
	acc := [6]amd64.VectorRegister{amd64.Z10, amd64.Z11, amd64.Z12, amd64.Z13, amd64.Z14, amd64.Z15}

	// Initialize accumulators to zero
	for _, reg := range acc {
		h.zeroReg(reg)
	}

	// Temporary register for m computation
	tmp := amd64.Z20

	// Process each limb of B (CIOS rounds)
	for i := range 5 {
		h.Comment(fmt.Sprintf("CIOS Round %d", i))
		h.ciosRound(i, a, b, acc, tmp)
	}

	// Fused x16 shift + normalization.
	// This carry chain must be propagated sequentially: each carry changes the
	// next limb before that limb's carry is known.
	h.Comment("Fused x16 shift + normalization")
	for i := range 5 {
		h.VPSLLQ("$4", acc[i], a[i])
	}

	// Propagate carries sequentially after the x16 shift.
	tmp = amd64.Z20
	for i := range 4 {
		h.carryProp(a[i], a[i], a[i+1], h.mask52, tmp)
	}
	h.VPANDQ(h.mask52, a[4], a[4])
}

// ciosRound generates one round of the CIOS Montgomery multiplication
func (h *ifmaHelper) ciosRound(i int, a, b [5]amd64.VectorRegister, acc [6]amd64.VectorRegister, tmp amd64.VectorRegister) {
	bi := b[i]

	// T += A * B[i]
	for j := range 5 {
		h.VPMADD52LUQ(bi, a[j], acc[j])
		h.VPMADD52HUQ(bi, a[j], acc[j+1])
	}

	// Normalize T[0]
	h.VPSRLQ("$52", acc[0], tmp)
	h.VPANDQ(h.mask52, acc[0], acc[0])
	h.VPADDQ(tmp, acc[1], acc[1])

	// m = T[0] * qInvNeg52 mod 2^52
	h.VPXORQ(tmp, tmp, tmp)
	h.VPMADD52LUQ(h.qInvNeg52, acc[0], tmp)
	h.VPANDQ(h.mask52, tmp, tmp)

	// T += m * q
	for j := range 5 {
		h.VPMADD52LUQ(h.qRadix52[j], tmp, acc[j])
		h.VPMADD52HUQ(h.qRadix52[j], tmp, acc[j+1])
	}

	// Shift: T[j] = T[j+1]
	h.VPSRLQ("$52", acc[0], tmp)
	h.VPADDQ(tmp, acc[1], acc[0])
	h.VMOVDQA64(acc[2], acc[1])
	h.VMOVDQA64(acc[3], acc[2])
	h.VMOVDQA64(acc[4], acc[3])
	h.VMOVDQA64(acc[5], acc[4])
	h.VPXORQ(acc[5], acc[5], acc[5])
}

// barrettReduction performs Barrett reduction from [0, 32q) to [0, q)
func (h *ifmaHelper) barrettReduction(a [5]amd64.VectorRegister) {
	h.Comment("k = (l4 * mu) >> 58, subtract k*q, then conditional subtract q")

	// k register
	k := amd64.Z5

	// Load Barrett constant and compute k
	h.MOVQ("$const_muBarrett52", amd64.AX)
	h.VPBROADCASTQ(amd64.AX, k)
	h.VPSRLQ("$20", a[4], amd64.Z6)
	h.VPMULUDQ(k, amd64.Z6, k)
	h.VPSRLQ("$38", k, k)

	// k*q low parts: Z6-Z10
	kqLow := [5]amd64.VectorRegister{amd64.Z6, amd64.Z7, amd64.Z8, amd64.Z9, amd64.Z10}
	// k*q high parts: Z15-Z19
	kqHigh := [5]amd64.VectorRegister{amd64.Z15, amd64.Z16, amd64.Z17, amd64.Z18, amd64.Z19}

	h.Comment("k*q using VPMADD52")
	for i := range kqLow {
		h.zeroReg(kqLow[i])
	}
	for i := range kqHigh {
		h.zeroReg(kqHigh[i])
	}

	// Low and high parts
	for i := range 5 {
		h.VPMADD52LUQ(h.qRadix52[i], k, kqLow[i])
		h.VPMADD52HUQ(h.qRadix52[i], k, kqHigh[i])
	}

	// Subtract k*q from result
	h.Comment("Subtract k*q")
	for i := range 5 {
		h.VPSUBQ(kqLow[i], a[i], a[i])
	}

	// Subtract high parts (carries)
	for i := range 4 {
		h.VPSUBQ(kqHigh[i], a[i+1], a[i+1])
	}

	// Propagate borrows
	h.Comment("Propagate borrows")
	tmp := amd64.Z15
	for i := range 4 {
		h.borrowProp(a[i], a[i+1], h.mask52, tmp)
	}
	h.VPANDQ(h.mask52, a[4], a[4])

	// Final conditional subtraction
	h.Comment("Final conditional subtraction of q")
	h.conditionalSubtractQ(a)
}

// conditionalSubtractQ performs conditional subtraction of q
func (h *ifmaHelper) conditionalSubtractQ(a [5]amd64.VectorRegister) {
	// Result - q in Z10-Z14
	sub := [5]amd64.VectorRegister{amd64.Z10, amd64.Z11, amd64.Z12, amd64.Z13, amd64.Z14}

	// Compute result - q
	for i := range 5 {
		h.VPSUBQ(h.qRadix52[i], a[i], sub[i])
	}

	// Propagate borrows
	tmp := amd64.Z20
	for i := range 4 {
		h.VPSRAQ("$63", sub[i], tmp)
		h.VPADDQ(tmp, sub[i+1], sub[i+1])
	}

	// Get final borrow mask
	h.VPSRAQ("$63", sub[4], tmp)

	// Mask subtracted limbs
	for i := range 5 {
		h.VPANDQ(h.mask52, sub[i], sub[i])
	}

	// Conditional select: if borrow, keep a; else use sub
	for i := range 5 {
		h.condSelect(sub[i], tmp, a[i])
	}
}

// loadAndConvertToRadix52 loads 8 elements and converts to radix-52
func (h *ifmaHelper) loadAndConvertToRadix52(addr amd64.Register, out [5]amd64.VectorRegister) {
	h.Comment(fmt.Sprintf("Load 8 elements from %s", addr))

	// Load registers for raw data
	raw := [4]amd64.VectorRegister{amd64.Z10, amd64.Z11, amd64.Z12, amd64.Z13}
	// After transpose: Z14-Z17 hold a0, a1, a2, a3
	transposed := [4]amd64.VectorRegister{amd64.Z14, amd64.Z15, amd64.Z16, amd64.Z17}

	h.Comment("Load element words using gather pattern")
	h.VMOVDQU64(fmt.Sprintf("0(%s)", addr), raw[0])
	h.VMOVDQU64(fmt.Sprintf("64(%s)", addr), raw[1])
	h.VMOVDQU64(fmt.Sprintf("128(%s)", addr), raw[2])
	h.VMOVDQU64(fmt.Sprintf("192(%s)", addr), raw[3])

	h.Comment("Transpose 8 elements for vertical SIMD processing")
	h.transposeForIFMA(raw, transposed)

	h.Comment("Convert to radix-52")
	// l0 = a0 & mask52
	h.VPANDQ(h.mask52, transposed[0], out[0])

	// l1 = (a0 >> 52) | ((a1 << 12) & mask52)
	h.VPSRLQ("$52", transposed[0], h.scratch[0])
	h.VPSLLQ("$12", transposed[1], h.scratch[1])
	h.VPORQ(h.scratch[0], h.scratch[1], h.scratch[0])
	h.VPANDQ(h.mask52, h.scratch[0], out[1])

	// l2 = (a1 >> 40) | ((a2 << 24) & mask52)
	h.VPSRLQ("$40", transposed[1], h.scratch[0])
	h.VPSLLQ("$24", transposed[2], h.scratch[1])
	h.VPORQ(h.scratch[0], h.scratch[1], h.scratch[0])
	h.VPANDQ(h.mask52, h.scratch[0], out[2])

	// l3 = (a2 >> 28) | ((a3 << 36) & mask52)
	h.VPSRLQ("$28", transposed[2], h.scratch[0])
	h.VPSLLQ("$36", transposed[3], h.scratch[1])
	h.VPORQ(h.scratch[0], h.scratch[1], h.scratch[0])
	h.VPANDQ(h.mask52, h.scratch[0], out[3])

	// l4 = a3 >> 16
	h.VPSRLQ("$16", transposed[3], out[4])
}

// transposeForIFMA transposes 8 elements from AoS to SoA format
func (h *ifmaHelper) transposeForIFMA(in, out [4]amd64.VectorRegister) {
	h.Comment("8x4 transpose using AVX-512 shuffles")
	h.Comment("Keep lanes in the native [e0,e2,e1,e3,e4,e6,e5,e7] order")
	h.Comment("throughout the IFMA pipeline to avoid full-lane permutations")

	// Step 1: Interleave low/high qwords between pairs
	h.VPUNPCKLQDQ(in[1], in[0], h.scratch[0], "[e0.a0, e2.a0, e0.a2, e2.a2, e1.a0, e3.a0, e1.a2, e3.a2]")
	h.VPUNPCKHQDQ(in[1], in[0], h.scratch[1], "[e0.a1, e2.a1, e0.a3, e2.a3, e1.a1, e3.a1, e1.a3, e3.a3]")
	h.VPUNPCKLQDQ(in[3], in[2], h.scratch[2], "[e4.a0, e6.a0, e4.a2, e6.a2, e5.a0, e7.a0, e5.a2, e7.a2]")
	h.VPUNPCKHQDQ(in[3], in[2], h.scratch[3], "[e4.a1, e6.a1, e4.a3, e6.a3, e5.a1, e7.a1, e5.a3, e7.a3]")

	// Step 2: Interleave across the 4 intermediate registers
	h.VSHUFI64X2("$0x88", h.scratch[2], h.scratch[0], out[0], "a0: lanes 0,2 from Z18 and Z20")
	h.VSHUFI64X2("$0xDD", h.scratch[2], h.scratch[0], out[2], "a2: lanes 1,3 from Z18 and Z20")
	h.VSHUFI64X2("$0x88", h.scratch[3], h.scratch[1], out[1], "a1: lanes 0,2 from Z19 and Z21")
	h.VSHUFI64X2("$0xDD", h.scratch[3], h.scratch[1], out[3], "a3: lanes 1,3 from Z19 and Z21")
}

// convertFromRadix52 converts from radix-52 to radix-64
func (h *ifmaHelper) convertFromRadix52(in [5]amd64.VectorRegister, out [4]amd64.VectorRegister) {
	h.Comment("Convert from radix-52 to radix-64")

	// a0 = l0 | (l1 << 52)
	h.VPSLLQ("$52", in[1], h.scratch[0])
	h.VPORQ(h.scratch[0], in[0], out[0])

	// a1 = (l1 >> 12) | (l2 << 40)
	h.VPSRLQ("$12", in[1], h.scratch[0])
	h.VPSLLQ("$40", in[2], h.scratch[1])
	h.VPORQ(h.scratch[1], h.scratch[0], out[1])

	// a2 = (l2 >> 24) | (l3 << 28)
	h.VPSRLQ("$24", in[2], h.scratch[0])
	h.VPSLLQ("$28", in[3], h.scratch[1])
	h.VPORQ(h.scratch[1], h.scratch[0], out[2])

	// a3 = (l3 >> 36) | (l4 << 16)
	h.VPSRLQ("$36", in[3], h.scratch[0])
	h.VPSLLQ("$16", in[4], h.scratch[1])
	h.VPORQ(h.scratch[1], h.scratch[0], out[3])
}

// transposeAndStore transposes from SoA to AoS format and stores
func (h *ifmaHelper) transposeAndStore(addr amd64.Register, in [4]amd64.VectorRegister) {
	// Output registers
	out := [4]amd64.VectorRegister{amd64.Z10, amd64.Z11, amd64.Z12, amd64.Z13}

	h.transposeFromIFMA(in, out)

	h.VMOVDQU64(out[0], fmt.Sprintf("0(%s)", addr))
	h.VMOVDQU64(out[1], fmt.Sprintf("64(%s)", addr))
	h.VMOVDQU64(out[2], fmt.Sprintf("128(%s)", addr))
	h.VMOVDQU64(out[3], fmt.Sprintf("192(%s)", addr))
}

// transposeFromIFMA reverses the transpose from SoA back to AoS
func (h *ifmaHelper) transposeFromIFMA(in, out [4]amd64.VectorRegister) {
	h.WriteLn("// 4x8 reverse transpose (SoA to AoS)")

	// The input already uses the native [e0,e2,e1,e3,e4,e6,e5,e7] lane order.
	// Pair a0 with a1 and a2 with a3 directly.
	h.VPUNPCKLQDQ(in[1], in[0], h.scratch[0], "pairs (a0,a1) for elements 0,1,4,5")
	h.VPUNPCKHQDQ(in[1], in[0], h.scratch[1], "pairs (a0,a1) for elements 2,3,6,7")
	h.VPUNPCKLQDQ(in[3], in[2], h.scratch[2], "pairs (a2,a3) for elements 0,1,4,5")
	h.VPUNPCKHQDQ(in[3], in[2], h.scratch[3], "pairs (a2,a3) for elements 2,3,6,7")

	// Step 2: Combine (a0,a1) with (a2,a3)
	h.VSHUFI64X2("$0x44", h.scratch[2], h.scratch[0], out[0])
	h.VSHUFI64X2("$0x44", h.scratch[3], h.scratch[1], out[1])
	h.VSHUFI64X2("$0xEE", h.scratch[2], h.scratch[0], out[2])
	h.VSHUFI64X2("$0xEE", h.scratch[3], h.scratch[1], out[3])

	// Step 3: Fix lane ordering inside each 256-bit half
	for i := range out {
		h.VSHUFI64X2("$0xD8", out[i], out[i], out[i])
	}
}

// generateInnerProdVecIFMA generates AVX-512 IFMA based inner product
func (f *FFAmd64) generateInnerProdVecIFMA() {
	f.Comment("innerProdVecIFMA(t *[8]uint64, a, b *Element, n uint64)")
	f.Comment("Computes inner product using IFMA multiplication and lazy accumulation")
	f.Comment("Processes 8 element pairs in parallel, accumulates into 8 qword accumulators")

	const argSize = 4 * 8
	stackSize := 0
	registers := f.FnHeader("innerProdVecIFMA", stackSize, argSize, amd64.AX, amd64.DX)
	defer f.AssertCleanStack(stackSize, 0)

	addrT := f.Pop(&registers)
	addrA := f.Pop(&registers)
	addrB := f.Pop(&registers)
	n := f.Pop(&registers)

	f.MOVQ("t+0(FP)", addrT)
	f.MOVQ("a+8(FP)", addrA)
	f.MOVQ("b+16(FP)", addrB)
	f.MOVQ("n+24(FP)", n)

	f.Comment("Initialize 8 qword accumulators in output buffer to zero")
	f.Comment("acc[0]=a0_low, acc[1]=a0_high, acc[2]=a1_low, ..., acc[7]=a3_high")
	f.VPXORQ(amd64.Z0, amd64.Z0, amd64.Z0)
	f.VMOVDQU64(amd64.Z0, fmt.Sprintf("0(%s)", addrT))

	f.Comment("Load constants for radix-52 conversion and reduction")
	h := f.newIfmaHelper()
	h.initIFMAConstants()

	// A in radix-52: Z0-Z4
	// B in radix-52: Z5-Z9
	aRadix52 := [5]amd64.VectorRegister{amd64.Z0, amd64.Z1, amd64.Z2, amd64.Z3, amd64.Z4}
	bRadix52 := [5]amd64.VectorRegister{amd64.Z5, amd64.Z6, amd64.Z7, amd64.Z8, amd64.Z9}

	// Result in radix-64 after conversion: Z14-Z17
	resultRadix64 := [4]amd64.VectorRegister{amd64.Z14, amd64.Z15, amd64.Z16, amd64.Z17}

	// Mask for low 32 bits
	mask32 := amd64.Z20

	f.Loop(n, func() {
		f.Comment("Process 8 element pairs in parallel")

		f.Comment("Load and convert 8 elements from a[] to radix-52")
		h.loadAndConvertToRadix52(addrA, aRadix52)

		f.Comment("Load and convert 8 elements from b[] to radix-52")
		h.loadAndConvertToRadix52(addrB, bRadix52)

		f.Comment("Montgomery multiplication using IFMA (CIOS variant)")
		h.montgomeryMulIFMA(aRadix52, bRadix52)

		f.Comment("Barrett reduction from [0, 32q) to [0, q)")
		h.barrettReduction(aRadix52)

		f.Comment("Convert result from radix-52 back to radix-64 (SoA format)")
		f.Comment("Z14=all a0, Z15=all a1, Z16=all a2, Z17=all a3")
		h.convertFromRadix52(aRadix52, resultRadix64)

		f.Comment("Split dwords and horizontal sum each accumulator position")
		f.Comment("Split before horizontal sum to avoid overflow (8 * 2^32 < 2^64)")

		// Create mask for low 32 bits
		f.MOVQ("$0xFFFFFFFF", amd64.AX)
		f.VPBROADCASTQ(amd64.AX, mask32, "mask for low 32 bits")

		// Process each limb: split into low/high, then horizontal sum
		for i, limb := range resultRadix64 {
			accLowIdx := i * 2
			accHighIdx := i*2 + 1

			// Extract low 32 bits and horizontal sum
			f.Comment(fmt.Sprintf("Limb %d low 32 bits -> acc[%d]", i, accLowIdx))
			f.VPANDQ(mask32, limb, h.scratch[0])
			// Horizontal sum: 8 qwords -> 1 qword
			f.VEXTRACTI64X4("$1", h.scratch[0], h.scratch[1].Y())
			f.VPADDQ(h.scratch[1].Y(), h.scratch[0].Y(), h.scratch[0].Y())
			f.VEXTRACTI64X2("$1", h.scratch[0].Y(), h.scratch[1].X())
			f.VPADDQ(h.scratch[1].X(), h.scratch[0].X(), h.scratch[0].X())
			f.VPSHUFD("$0xEE", h.scratch[0].X(), h.scratch[1].X())
			f.VPADDQ(h.scratch[1].X(), h.scratch[0].X(), h.scratch[0].X())
			f.VMOVQ(h.scratch[0].X(), amd64.R8)
			f.ADDQ(amd64.R8, fmt.Sprintf("%d(%s)", accLowIdx*8, addrT))

			// Extract high 32 bits and horizontal sum
			f.Comment(fmt.Sprintf("Limb %d high 32 bits -> acc[%d]", i, accHighIdx))
			f.VPSRLQ("$32", limb, h.scratch[0])
			// Horizontal sum: 8 qwords -> 1 qword
			f.VEXTRACTI64X4("$1", h.scratch[0], h.scratch[1].Y())
			f.VPADDQ(h.scratch[1].Y(), h.scratch[0].Y(), h.scratch[0].Y())
			f.VEXTRACTI64X2("$1", h.scratch[0].Y(), h.scratch[1].X())
			f.VPADDQ(h.scratch[1].X(), h.scratch[0].X(), h.scratch[0].X())
			f.VPSHUFD("$0xEE", h.scratch[0].X(), h.scratch[1].X())
			f.VPADDQ(h.scratch[1].X(), h.scratch[0].X(), h.scratch[0].X())
			f.VMOVQ(h.scratch[0].X(), amd64.R8)
			f.ADDQ(amd64.R8, fmt.Sprintf("%d(%s)", accHighIdx*8, addrT))
		}

		f.Comment("Advance pointers")
		f.ADDQ("$256", addrA)
		f.ADDQ("$256", addrB)
	})

	f.RETAVX()

	f.Push(&registers, addrT, addrA, addrB, n)
}
