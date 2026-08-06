package arm64

import (
	"fmt"
	"io"

	"github.com/consensys/bavard/arm64"
	"github.com/consensys/gnark-crypto/internal/generator/field/asm/amd64"
)

// vRegNum extracts the numeric register ID from a VectorRegister (V0 -> 0, V31 -> 31)
func vRegNum(v arm64.VectorRegister) uint32 {
	s := string(v)
	// Remove any suffix like .S4, .D2, etc.
	for i := 0; i < len(s); i++ {
		if s[i] == '.' {
			s = s[:i]
			break
		}
	}
	// Parse "Vn" where n is 0-31
	if len(s) < 2 || s[0] != 'V' {
		panic("invalid vector register: " + string(v))
	}
	var n uint32
	for i := 1; i < len(s); i++ {
		n = n*10 + uint32(s[i]-'0')
	}
	return n
}

// baseReg returns the base register name (e.g., "V0" from "V0.S4")
func baseReg(v arm64.VectorRegister) string {
	s := string(v)
	for i := 0; i < len(s); i++ {
		if s[i] == '.' {
			return s[:i]
		}
	}
	return s
}

func GenerateF31Poseidon2(w io.Writer, nbBits int, q, qInvNeg uint64, params []amd64.Poseidon2Parameters) error {
	f := NewFFArm64(w, (nbBits+63)/64)
	for _, p := range params {
		if p.Width == 16 && p.HasCompressx16 {
			f.generatePoseidon2_F31_16x16x512(p, q, qInvNeg, false)
			f.generatePoseidon2_F31_16x16x512(p, q, qInvNeg, true)
		}
	}
	return nil
}

// generatePoseidon2_F31_16x16x512 generates ARM64 NEON assembly for Poseidon2 permutation
// on F31 fields with width=16, processing 4 permutations in parallel using NEON vectors.
//
// Memory layout:
//   - matrix: input data, 16 rows × 512 field elements each (total 16 × 512 × 4 = 32768 bytes)
//   - roundKeys: slice header pointing to [][]fr.Element round keys
//   - result: output buffer, 16 rows × 8 field elements each (total 16 × 8 × 4 = 512 bytes)
//
// With columns == true, it instead generates permutation16x16xN_columns_arm64: same
// algorithm, but the input is column-major (matrix[pos*16+lane]) with a runtime nbSteps
// argument (colSize = nbSteps*8). In that layout the 4 lanes of a batch are contiguous
// in memory, so each rate coordinate is a single VLD1 instead of 4 scalar loads.
//
// Algorithm:
//   - We process 4 rows in parallel (4 NEON lanes), so we need 4 batches to cover all 16 rows
//   - Each batch processes N=64 steps (512 elements / 8 elements per step)
//   - The state v[0..15] holds 16 field elements for each of 4 parallel permutations
//   - Feed-forward: after all steps, state[j] += input[j] for j in [0,8)
//
// Register allocation:
//   - V0: constant q (field modulus broadcast to all lanes)
//   - V1: constant mu (Montgomery constant for reduction)
//   - V2-V17: state vectors (16 vectors, each holds 4 field elements from parallel permutations)
//   - V18-V27: temporary vectors for arithmetic operations
//   - V28: constant 1s for AND operations
//   - V29: used by mul as private temp
//   - V30-V31: scratch for modular arithmetic
//   - R0-R12: general purpose (addresses, counters, etc.)
func (f *FFArm64) generatePoseidon2_F31_16x16x512(params amd64.Poseidon2Parameters, constQ, constQInvNeg uint64, columns bool) {
	fullRounds := params.FullRounds
	partialRounds := params.PartialRounds
	rf := fullRounds / 2 // half rounds before and after partial rounds

	if params.Width != 16 {
		panic("only width 16 is supported")
	}

	fnName := "permutation16x16x512_arm64"
	argSize := 8 + 24 + 8 // matrix ptr + roundKeys slice header + result ptr
	if columns {
		fnName = "permutation16x16xN_columns_arm64"
		argSize += 8 // + nbSteps
	}

	// Stack frame for temporary storage during each step (8 vectors × 16 bytes = 128 bytes)
	const stackSize = 128
	registers := f.FnHeader(fnName, stackSize, argSize)

	// =========================================================================
	// Register Allocation
	// =========================================================================

	// General purpose registers for addresses and counters
	addrMatrix := registers.Pop()    // base address of input matrix
	addrRoundKeys := registers.Pop() // address of roundKeys slice header
	addrResult := registers.Pop()    // base address of result buffer

	f.MOVD("matrix+0(FP)", addrMatrix)
	f.MOVD("roundKeys+8(FP)", addrRoundKeys)
	f.MOVD("result+32(FP)", addrResult)

	// Constants in scalar registers for VDUP
	qReg := registers.Pop()       // field modulus q
	qInvNegReg := registers.Pop() // qInvNeg = -q^{-1} mod 2^32

	f.MOVD(constQ, qReg)
	f.MOVD(constQInvNeg, qInvNegReg)

	// Vector constants (broadcast to all 4 lanes)
	vQ := arm64.V0  // q broadcast
	vMu := arm64.V1 // qInvNeg broadcast (misnamed vMu for historical reasons)
	f.VDUP(qReg, vQ.S4())
	f.VDUP(qInvNegReg, vMu.S4())

	// Load constant 1 for LSB extraction in halve operation
	tmpReg := registers.Pop()
	f.MOVD(1, tmpReg)
	vOneVec := arm64.V28 // V28 = {1, 1, 1, 1} for AND operations
	f.VDUP(tmpReg, vOneVec.S4())

	// State vectors: v[0..15] = V2..V17
	// Each vector holds 4 field elements from 4 parallel permutations
	v := []arm64.VectorRegister{
		arm64.V2, arm64.V3, arm64.V4, arm64.V5,
		arm64.V6, arm64.V7, arm64.V8, arm64.V9,
		arm64.V10, arm64.V11, arm64.V12, arm64.V13,
		arm64.V14, arm64.V15, arm64.V16, arm64.V17,
	}

	// Temporary vectors for arithmetic: t[0..9] = V18..V27
	// V28 is reserved for vOneVec (constant 1s)
	t := []arm64.VectorRegister{
		arm64.V18, arm64.V19, arm64.V20, arm64.V21,
		arm64.V22, arm64.V23, arm64.V24, arm64.V25,
		arm64.V26, arm64.V27,
	}

	// Scratch registers (used within macros, can be overwritten freely)
	scratch0 := arm64.V30
	scratch1 := arm64.V31

	// Additional GP registers for loop control and addresses
	rKeyPtr := registers.Pop()  // current round key pointer
	batchIdx := registers.Pop() // outer loop counter (0..3)
	stepIdx := registers.Pop()  // inner loop counter (0..N-1)
	// Pointers to 4 rows for current batch
	ptr0 := registers.Pop()    // data pointer for batch row 0
	ptr1 := registers.Pop()    // data pointer for batch row 1
	ptr2 := registers.Pop()    // data pointer for batch row 2
	ptr3 := registers.Pop()    // data pointer for batch row 3
	tmpCalc := registers.Pop() // temporary for address calculations

	var nbSteps arm64.Register // number of steps (columns variant only)
	if columns {
		nbSteps = registers.Pop()
		f.MOVD("nbSteps+40(FP)", nbSteps)
	}

	// =========================================================================
	// Modular Arithmetic Macros (using Define)
	// =========================================================================

	// Add: computes (a + b) mod q using conditional subtraction
	// Inputs: a, b, into (all vector registers)
	// Uses scratch registers V30, V31
	add := f.Define("ADD_MOD", 3, func(args ...arm64.Register) {
		a := arm64.VectorRegister(args[0])
		b := arm64.VectorRegister(args[1])
		into := arm64.VectorRegister(args[2])
		f.VADD(a.S4(), b.S4(), scratch0.S4())
		f.VSUB(vQ.S4(), scratch0.S4(), scratch1.S4())
		f.VUMIN(scratch0.S4(), scratch1.S4(), into.S4())
	})

	// Sub: computes (a - b) mod q using conditional addition
	// Inputs: a, b, into (all vector registers)
	sub := f.Define("SUB_MOD", 3, func(args ...arm64.Register) {
		a := arm64.VectorRegister(args[0])
		b := arm64.VectorRegister(args[1])
		into := arm64.VectorRegister(args[2])
		f.VSUB(b.S4(), a.S4(), scratch0.S4())
		f.VADD(vQ.S4(), scratch0.S4(), scratch1.S4())
		f.VUMIN(scratch0.S4(), scratch1.S4(), into.S4())
	})

	// Double: computes 2*a mod q
	// Inputs: a, into
	double := f.Define("DOUBLE_MOD", 2, func(args ...arm64.Register) {
		a := arm64.VectorRegister(args[0])
		into := arm64.VectorRegister(args[1])
		f.VSHL("$1", a.S4(), scratch0.S4())
		f.VSUB(vQ.S4(), scratch0.S4(), scratch1.S4())
		f.VUMIN(scratch0.S4(), scratch1.S4(), into.S4())
	})

	// Mul: Montgomery multiplication (a * b * R^-1) mod q
	// Inputs: a, b, into
	// Uses V29 as private temp, V30-V31 as scratch, and t[8], t[9] as temps
	// Cannot use Define because it needs register number extraction
	mulTmp := arm64.V29
	mul := func(a, b, into arm64.VectorRegister) {
		an := vRegNum(a)
		bn := vRegNum(b)
		qn := vRegNum(vQ)
		s0n := vRegNum(scratch0)
		s1n := vRegNum(scratch1)
		mn := vRegNum(mulTmp)

		// Step 1: ab = a * b (64-bit widening multiply)
		encUmullLo := uint32(0x2ea0c000) | (bn << 16) | (an << 5) | s0n
		f.WriteLn(fmt.Sprintf("    WORD $0x%08x // UMULL %s.2D, %s.2S, %s.2S", encUmullLo, baseReg(scratch0), baseReg(a), baseReg(b)))

		encUmullHi := uint32(0x6ea0c000) | (bn << 16) | (an << 5) | s1n
		f.WriteLn(fmt.Sprintf("    WORD $0x%08x // UMULL2 %s.2D, %s.4S, %s.4S", encUmullHi, baseReg(scratch1), baseReg(a), baseReg(b)))

		// Step 2: Extract ab_lo
		f.VUZP1(scratch0.S4(), scratch1.S4(), mulTmp.S4())

		// Step 3: m = (ab_lo * qInvNeg) mod 2^32
		f.VMUL_S4(mulTmp.S4(), vMu.S4(), mulTmp.S4())

		// Step 4: Compute m * q
		t8n := vRegNum(t[8])
		t9n := vRegNum(t[9])
		encMqLo := uint32(0x2ea0c000) | (qn << 16) | (mn << 5) | t8n
		f.WriteLn(fmt.Sprintf("    WORD $0x%08x // UMULL %s.2D, %s.2S, %s.2S", encMqLo, baseReg(t[8]), baseReg(mulTmp), baseReg(vQ)))

		encMqHi := uint32(0x6ea0c000) | (qn << 16) | (mn << 5) | t9n
		f.WriteLn(fmt.Sprintf("    WORD $0x%08x // UMULL2 %s.2D, %s.4S, %s.4S", encMqHi, baseReg(t[9]), baseReg(mulTmp), baseReg(vQ)))

		// Step 5: Add ab + m*q
		f.VADD(scratch0.D2(), t[8].D2(), scratch0.D2())
		f.VADD(scratch1.D2(), t[9].D2(), scratch1.D2())

		// Step 6: Extract high 32 bits
		f.VUZP2(scratch0.S4(), scratch1.S4(), into.S4())

		// Step 7: Reduce if result >= q
		f.VSUB(vQ.S4(), into.S4(), mulTmp.S4())
		f.VUMIN(into.S4(), mulTmp.S4(), into.S4())
	}

	// Halve: computes a/2 mod q
	// Cannot use Define because it needs register number extraction
	halve := func(a, into arm64.VectorRegister) {
		an := vRegNum(a)
		s0n := vRegNum(scratch0)
		s1n := vRegNum(scratch1)
		dn := vRegNum(into)

		// mask = (a & 1) << 31
		f.VAND(a.B16(), vOneVec.B16(), scratch0.B16())
		f.VSHL("$31", scratch0.S4(), scratch0.S4())

		// mask = mask >> 31 (arithmetic shift, creates all 1s or all 0s)
		encoding := uint32(0x4f210400) | (s0n << 5) | s0n
		f.WriteLn(fmt.Sprintf("    WORD $0x%08x // SSHR V%d.4S, V%d.4S, #31", encoding, s0n, s0n))

		// masked_q = mask & q
		f.VAND(vQ.B16(), scratch0.B16(), scratch1.B16())

		// result = UHADD(a, masked_q) = (a + masked_q) / 2
		uhaddEncoding := uint32(0x6ea00400) | (s1n << 16) | (an << 5) | dn
		f.WriteLn(fmt.Sprintf("    WORD $0x%08x // UHADD V%d.4S, V%d.4S, V%d.4S", uhaddEncoding, dn, an, s1n))
	}

	// Triple: computes 3*a mod q = 2*a + a
	// Inputs: a, into
	triple := f.Define("TRIPLE_MOD", 2, func(args ...arm64.Register) {
		a := arm64.VectorRegister(args[0])
		into := arm64.VectorRegister(args[1])
		double(arm64.Register(a), arm64.Register(scratch0))
		add(arm64.Register(scratch0), arm64.Register(a), arm64.Register(into))
	})

	// Quadruple: computes 4*a mod q = 2*(2*a)
	// Inputs: a, into
	quadruple := f.Define("QUAD_MOD", 2, func(args ...arm64.Register) {
		a := arm64.VectorRegister(args[0])
		into := arm64.VectorRegister(args[1])
		double(arm64.Register(a), arm64.Register(scratch0))
		double(arm64.Register(scratch0), arm64.Register(into))
	})

	// S-box: applies x^3 (cubic S-box)
	// Cannot use Define because it calls mul which needs register extraction
	sbox := func(state arm64.VectorRegister) {
		mul(state, state, t[0]) // t[0] = state^2
		mul(state, t[0], state) // state = state^3
	}

	// mul2ExpNegN computes a * 2^(-n) mod q
	// For small n (n <= 4), use repeated halving
	// For large n (n > 4), use Montgomery reduction
	mul2ExpNegN := func(a, into arm64.VectorRegister, n int) {
		if n <= 4 {
			halve(a, into)
			for i := 1; i < n; i++ {
				halve(into, into)
			}
			return
		}

		// For larger n, use Montgomery reduction
		shift := 32 - n

		an := vRegNum(a)
		s0n := vRegNum(scratch0)
		s1n := vRegNum(scratch1)
		mn := vRegNum(mulTmp)
		qn := vRegNum(vQ)
		_ = vRegNum(into)

		// Step 1: Widen and shift left: v = a << shift (64-bit result)
		encLo := uint32(0x2f00a400) | (uint32(32+shift) << 16) | (an << 5) | s0n
		f.WriteLn(fmt.Sprintf("    WORD $0x%08x // USHLL V%d.2D, V%d.2S, #%d", encLo, s0n, an, shift))

		encHi := uint32(0x6f00a400) | (uint32(32+shift) << 16) | (an << 5) | s1n
		f.WriteLn(fmt.Sprintf("    WORD $0x%08x // USHLL2 V%d.2D, V%d.4S, #%d", encHi, s1n, an, shift))

		// Step 2: Extract v_lo
		f.VUZP1(scratch0.S4(), scratch1.S4(), mulTmp.S4())

		// Step 3: m = v_lo * mu
		f.VMUL_S4(mulTmp.S4(), vMu.S4(), mulTmp.S4())

		// Step 4: Compute m * q
		t8n := vRegNum(t[8])
		t9n := vRegNum(t[9])
		umullLoEnc := uint32(0x2ea0c000) | (qn << 16) | (mn << 5) | t8n
		f.WriteLn(fmt.Sprintf("    WORD $0x%08x // UMULL V%d.2D, V%d.2S, V%d.2S", umullLoEnc, t8n, mn, qn))

		umullHiEnc := uint32(0x6ea0c000) | (qn << 16) | (mn << 5) | t9n
		f.WriteLn(fmt.Sprintf("    WORD $0x%08x // UMULL2 V%d.2D, V%d.4S, V%d.4S", umullHiEnc, t9n, mn, qn))

		// Step 5: Add v + m*q
		f.VADD(scratch0.D2(), t[8].D2(), scratch0.D2())
		f.VADD(scratch1.D2(), t[9].D2(), scratch1.D2())

		// Step 6: Extract high 32 bits
		f.VUZP2(scratch0.S4(), scratch1.S4(), into.S4())

		// Step 7: Final reduction
		f.VSUB(vQ.S4(), into.S4(), mulTmp.S4())
		f.VUMIN(into.S4(), mulTmp.S4(), into.S4())
	}

	// =========================================================================
	// Matrix Multiplication Macros (using Define)
	// =========================================================================

	// matMul4: computes 4x4 circulant matrix multiplication
	// Matrix: (2 3 1 1)
	//         (1 2 3 1)
	//         (1 1 2 3)
	//         (3 1 1 2)
	// Uses only add and double Defines, so can be a Define
	matMul4 := f.Define("MAT_MUL_4", 4, func(args ...arm64.Register) {
		s0 := arm64.VectorRegister(args[0])
		s1 := arm64.VectorRegister(args[1])
		s2 := arm64.VectorRegister(args[2])
		s3 := arm64.VectorRegister(args[3])

		add(arm64.Register(s0), arm64.Register(s1), arm64.Register(t[0]))
		add(arm64.Register(s2), arm64.Register(s3), arm64.Register(t[1]))
		add(arm64.Register(t[0]), arm64.Register(t[1]), arm64.Register(t[2]))
		add(arm64.Register(t[2]), arm64.Register(s1), arm64.Register(t[3]))
		add(arm64.Register(t[2]), arm64.Register(s3), arm64.Register(t[4]))
		double(arm64.Register(s0), arm64.Register(s3))
		add(arm64.Register(s3), arm64.Register(t[4]), arm64.Register(s3))
		double(arm64.Register(s2), arm64.Register(s1))
		add(arm64.Register(s1), arm64.Register(t[3]), arm64.Register(s1))
		add(arm64.Register(t[0]), arm64.Register(t[3]), arm64.Register(s0))
		add(arm64.Register(t[1]), arm64.Register(t[4]), arm64.Register(s2))
	})

	// matMulExternal: computes external matrix for full rounds
	// Uses only matMul4 and add Defines, so can be a Define
	matMulExternal := f.Define("MAT_MUL_EXT", 16, func(args ...arm64.Register) {
		// Apply M4 to each block
		for i := range 4 {
			matMul4(args[i*4], args[i*4+1], args[i*4+2], args[i*4+3])
		}

		// Compute cross-block sums
		vv := make([]arm64.VectorRegister, 16)
		for i := range 16 {
			vv[i] = arm64.VectorRegister(args[i])
		}

		add(arm64.Register(vv[0]), arm64.Register(vv[4]), arm64.Register(t[0]))
		add(arm64.Register(t[0]), arm64.Register(vv[8]), arm64.Register(t[0]))
		add(arm64.Register(t[0]), arm64.Register(vv[12]), arm64.Register(t[0]))

		add(arm64.Register(vv[1]), arm64.Register(vv[5]), arm64.Register(t[1]))
		add(arm64.Register(t[1]), arm64.Register(vv[9]), arm64.Register(t[1]))
		add(arm64.Register(t[1]), arm64.Register(vv[13]), arm64.Register(t[1]))

		add(arm64.Register(vv[2]), arm64.Register(vv[6]), arm64.Register(t[2]))
		add(arm64.Register(t[2]), arm64.Register(vv[10]), arm64.Register(t[2]))
		add(arm64.Register(t[2]), arm64.Register(vv[14]), arm64.Register(t[2]))

		add(arm64.Register(vv[3]), arm64.Register(vv[7]), arm64.Register(t[3]))
		add(arm64.Register(t[3]), arm64.Register(vv[11]), arm64.Register(t[3]))
		add(arm64.Register(t[3]), arm64.Register(vv[15]), arm64.Register(t[3]))

		// Add cross-block sums to each element
		for i := range 16 {
			add(arm64.Register(vv[i]), arm64.Register(t[i%4]), arm64.Register(vv[i]))
		}
	})

	// matMulInternal: computes internal matrix for partial rounds
	// Cannot use Define because it calls halve and mul2ExpNegN which need register extraction
	matMulInternal := func() {
		// Compute sum of all elements (tree reduction)
		add(arm64.Register(v[0]), arm64.Register(v[1]), arm64.Register(t[0]))
		add(arm64.Register(v[2]), arm64.Register(v[3]), arm64.Register(t[1]))
		add(arm64.Register(v[4]), arm64.Register(v[5]), arm64.Register(t[2]))
		add(arm64.Register(v[6]), arm64.Register(v[7]), arm64.Register(t[3]))
		add(arm64.Register(t[0]), arm64.Register(t[1]), arm64.Register(t[0]))
		add(arm64.Register(t[2]), arm64.Register(t[3]), arm64.Register(t[2]))
		add(arm64.Register(t[0]), arm64.Register(t[2]), arm64.Register(t[0]))

		add(arm64.Register(v[8]), arm64.Register(v[9]), arm64.Register(t[4]))
		add(arm64.Register(v[10]), arm64.Register(v[11]), arm64.Register(t[5]))
		add(arm64.Register(v[12]), arm64.Register(v[13]), arm64.Register(t[6]))
		add(arm64.Register(v[14]), arm64.Register(v[15]), arm64.Register(t[7]))
		add(arm64.Register(t[4]), arm64.Register(t[5]), arm64.Register(t[4]))
		add(arm64.Register(t[6]), arm64.Register(t[7]), arm64.Register(t[6]))
		add(arm64.Register(t[4]), arm64.Register(t[6]), arm64.Register(t[4]))

		add(arm64.Register(t[0]), arm64.Register(t[4]), arm64.Register(t[0]))

		// Apply diagonal multiplication
		double(arm64.Register(v[0]), arm64.Register(v[0]))
		sub(arm64.Register(t[0]), arm64.Register(v[0]), arm64.Register(v[0]))

		add(arm64.Register(t[0]), arm64.Register(v[1]), arm64.Register(v[1]))

		double(arm64.Register(v[2]), arm64.Register(v[2]))
		add(arm64.Register(t[0]), arm64.Register(v[2]), arm64.Register(v[2]))

		halve(v[3], v[3])
		add(arm64.Register(t[0]), arm64.Register(v[3]), arm64.Register(v[3]))

		triple(arm64.Register(v[4]), arm64.Register(v[4]))
		add(arm64.Register(t[0]), arm64.Register(v[4]), arm64.Register(v[4]))

		quadruple(arm64.Register(v[5]), arm64.Register(v[5]))
		add(arm64.Register(t[0]), arm64.Register(v[5]), arm64.Register(v[5]))

		halve(v[6], v[6])
		sub(arm64.Register(t[0]), arm64.Register(v[6]), arm64.Register(v[6]))

		triple(arm64.Register(v[7]), arm64.Register(v[7]))
		sub(arm64.Register(t[0]), arm64.Register(v[7]), arm64.Register(v[7]))

		quadruple(arm64.Register(v[8]), arm64.Register(v[8]))
		sub(arm64.Register(t[0]), arm64.Register(v[8]), arm64.Register(v[8]))

		mul2ExpNegN(v[9], v[9], 8)
		add(arm64.Register(t[0]), arm64.Register(v[9]), arm64.Register(v[9]))

		mul2ExpNegN(v[10], v[10], 3)
		add(arm64.Register(t[0]), arm64.Register(v[10]), arm64.Register(v[10]))

		mul2ExpNegN(v[11], v[11], 24)
		add(arm64.Register(t[0]), arm64.Register(v[11]), arm64.Register(v[11]))

		mul2ExpNegN(v[12], v[12], 8)
		sub(arm64.Register(t[0]), arm64.Register(v[12]), arm64.Register(v[12]))

		mul2ExpNegN(v[13], v[13], 3)
		sub(arm64.Register(t[0]), arm64.Register(v[13]), arm64.Register(v[13]))

		mul2ExpNegN(v[14], v[14], 4)
		sub(arm64.Register(t[0]), arm64.Register(v[14]), arm64.Register(v[14]))

		mul2ExpNegN(v[15], v[15], 24)
		sub(arm64.Register(t[0]), arm64.Register(v[15]), arm64.Register(v[15]))
	}

	// =========================================================================
	// Round Functions
	// =========================================================================

	roundIdx := 0

	// addRoundKeyFull loads round key j for current full round and adds to state
	addRoundKeyFull := func(j int) {
		f.MOVD(fmt.Sprintf("%d(%s)", roundIdx*24, addrRoundKeys), rKeyPtr)
		f.ADD(uint64(j*4), rKeyPtr, tmpCalc)
		f.WriteLn(fmt.Sprintf("    VLD1R (%s), [%s]", tmpCalc, scratch0.S4()))
		add(arm64.Register(v[j]), arm64.Register(scratch0), arm64.Register(v[j]))
	}

	// addRoundKeyPartial loads the single round key for current partial round
	addRoundKeyPartial := func() {
		f.MOVD(fmt.Sprintf("%d(%s)", roundIdx*24, addrRoundKeys), rKeyPtr)
		f.WriteLn(fmt.Sprintf("    VLD1R (%s), [%s]", rKeyPtr, scratch0.S4()))
		add(arm64.Register(v[0]), arm64.Register(scratch0), arm64.Register(v[0]))
	}

	// fullRound applies round key, S-box to all elements, then external matrix
	fullRound := func() {
		for j := range 16 {
			addRoundKeyFull(j)
			sbox(v[j])
		}
		matMulExternal(arm64.Register(v[0]), arm64.Register(v[1]), arm64.Register(v[2]), arm64.Register(v[3]),
			arm64.Register(v[4]), arm64.Register(v[5]), arm64.Register(v[6]), arm64.Register(v[7]),
			arm64.Register(v[8]), arm64.Register(v[9]), arm64.Register(v[10]), arm64.Register(v[11]),
			arm64.Register(v[12]), arm64.Register(v[13]), arm64.Register(v[14]), arm64.Register(v[15]))
		roundIdx++
	}

	// partialRound applies round key and S-box only to v[0], then internal matrix
	partialRound := func() {
		addRoundKeyPartial()
		sbox(v[0])
		roundIdx++
		matMulInternal()
	}

	// =========================================================================
	// Main Loop Structure
	// =========================================================================

	f.MOVD(0, batchIdx)
	f.LABEL("batch_loop")

	// Zero state vectors
	for i := range 16 {
		f.VEOR(v[i].B16(), v[i].B16(), v[i].B16())
	}

	// Initialize pointers for 4 parallel inputs
	const N = 512 / 8 // 64 steps per batch
	f.MOVD(0, stepIdx)

	if columns {
		// lanes 4b..4b+3 of position pos live at matrix + pos*64 + b*16
		f.WriteLn(fmt.Sprintf("    LSL $4, %s, %s", batchIdx, tmpCalc))
		f.ADD(addrMatrix, tmpCalc, ptr0)
	} else {
		f.WriteLn(fmt.Sprintf("    LSL $13, %s, %s", batchIdx, tmpCalc))
		f.ADD(addrMatrix, tmpCalc, ptr0)
		f.ADD(2048, ptr0, ptr1)
		f.ADD(2048, ptr1, ptr2)
		f.ADD(2048, ptr2, ptr3)
	}

	f.LABEL("step_loop")

	if columns {
		// The 4 lanes of the batch are contiguous: one VLD1 per rate coordinate,
		// consecutive positions are 64 bytes apart.
		for j := range 8 {
			f.VLD1_P(16, ptr0, t[j].S4())
			f.ADD(48, ptr0, ptr0)
		}
	} else {
		// Load 8 elements from each of 4 lanes
		for j := range 8 {
			f.MOVWU(fmt.Sprintf("(%s)", ptr0), tmpCalc)
			f.WriteLn(fmt.Sprintf("    VMOV %s, %s", tmpCalc, t[j].SAt(0)))
			f.MOVWU(fmt.Sprintf("(%s)", ptr1), tmpCalc)
			f.WriteLn(fmt.Sprintf("    VMOV %s, %s", tmpCalc, t[j].SAt(1)))
			f.MOVWU(fmt.Sprintf("(%s)", ptr2), tmpCalc)
			f.WriteLn(fmt.Sprintf("    VMOV %s, %s", tmpCalc, t[j].SAt(2)))
			f.MOVWU(fmt.Sprintf("(%s)", ptr3), tmpCalc)
			f.WriteLn(fmt.Sprintf("    VMOV %s, %s", tmpCalc, t[j].SAt(3)))
			f.ADD(4, ptr0, ptr0)
			f.ADD(4, ptr1, ptr1)
			f.ADD(4, ptr2, ptr2)
			f.ADD(4, ptr3, ptr3)
		}
	}

	// Copy input into state[8..15]
	for j := range 8 {
		f.VMOV(t[j].B16(), v[8+j].B16())
	}

	// Store t[0..7] on stack for feed-forward
	f.MOVD("RSP", tmpCalc)
	for j := range 8 {
		f.VST1_P(t[j].S4(), tmpCalc, 16)
	}

	// Reset round index
	roundIdx = 0

	// Initial external matrix
	matMulExternal(arm64.Register(v[0]), arm64.Register(v[1]), arm64.Register(v[2]), arm64.Register(v[3]),
		arm64.Register(v[4]), arm64.Register(v[5]), arm64.Register(v[6]), arm64.Register(v[7]),
		arm64.Register(v[8]), arm64.Register(v[9]), arm64.Register(v[10]), arm64.Register(v[11]),
		arm64.Register(v[12]), arm64.Register(v[13]), arm64.Register(v[14]), arm64.Register(v[15]))

	// Apply rf full rounds
	for range rf {
		fullRound()
	}

	// Apply partial rounds
	for range partialRounds {
		partialRound()
	}

	// Apply rf more full rounds
	for range rf {
		fullRound()
	}

	// Restore t[0..7] from stack
	f.MOVD("RSP", tmpCalc)
	for j := range 8 {
		f.VLD1_P(16, tmpCalc, t[j].S4())
	}

	// Feed-forward: state[j] = state[8+j] + original_input[j]
	for j := range 8 {
		add(arm64.Register(v[8+j]), arm64.Register(t[j]), arm64.Register(v[j]))
	}

	// Loop control
	f.ADD(1, stepIdx, stepIdx)
	if columns {
		f.CMP(nbSteps, stepIdx)
	} else {
		f.CMP(N, stepIdx)
	}
	f.WriteLn("    BNE step_loop")

	// Store results for this batch
	f.WriteLn(fmt.Sprintf("    LSL $7, %s, %s", batchIdx, tmpCalc))
	f.ADD(addrResult, tmpCalc, ptr0)
	f.ADD(32, ptr0, ptr1)
	f.ADD(32, ptr1, ptr2)
	f.ADD(32, ptr2, ptr3)

	for j := range 8 {
		f.WriteLn(fmt.Sprintf("    VMOV %s, %s", v[j].SAt(0), tmpCalc))
		f.MOVWU(tmpCalc, fmt.Sprintf("(%s)", ptr0))
		f.WriteLn(fmt.Sprintf("    VMOV %s, %s", v[j].SAt(1), tmpCalc))
		f.MOVWU(tmpCalc, fmt.Sprintf("(%s)", ptr1))
		f.WriteLn(fmt.Sprintf("    VMOV %s, %s", v[j].SAt(2), tmpCalc))
		f.MOVWU(tmpCalc, fmt.Sprintf("(%s)", ptr2))
		f.WriteLn(fmt.Sprintf("    VMOV %s, %s", v[j].SAt(3), tmpCalc))
		f.MOVWU(tmpCalc, fmt.Sprintf("(%s)", ptr3))
		f.ADD(4, ptr0, ptr0)
		f.ADD(4, ptr1, ptr1)
		f.ADD(4, ptr2, ptr2)
		f.ADD(4, ptr3, ptr3)
	}

	// Batch loop control
	f.ADD(1, batchIdx, batchIdx)
	f.CMP(4, batchIdx)
	f.WriteLn("    BNE batch_loop")

	f.RET()
}
