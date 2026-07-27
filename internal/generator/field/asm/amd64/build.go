// Copyright 2020-2025 Consensys Software Inc.
// Licensed under the Apache License, Version 2.0. See the LICENSE file for details.

// Package amd64 contains syntactic sugar to generate amd64 assembly code
package amd64

import (
	"fmt"
	"io"
	"path/filepath"
	"strings"

	"github.com/consensys/bavard/amd64"
)

const SmallModulus = 6

func NewFFAmd64(w io.Writer, nbWords int) *FFAmd64 {
	F := &FFAmd64{
		amd64.NewAmd64(w),
		0,
		0,
		nbWords,
		nbWords - 1,
		make([]int, nbWords),
		make([]int, nbWords-1),
		make(map[string]defineFn),
		nil,
	}

	// indexes (template helpers)
	for i := 0; i < F.NbWords; i++ {
		F.NbWordsIndexesFull[i] = i
		if i > 0 {
			F.NbWordsIndexesNoZero[i-1] = i
		}
	}

	return F
}

type FFAmd64 struct {
	// *config.FieldConfig
	*amd64.Amd64
	nbElementsOnStack    int
	maxOnStack           int
	NbWords              int
	NbWordsLastIndex     int
	NbWordsIndexesFull   []int
	NbWordsIndexesNoZero []int
	mDefines             map[string]defineFn

	qStack []amd64.Register // when set, contains the words of q, the modulus, on the stack.
}

type defineFn func(args ...any)

func (f *FFAmd64) SetQStack(qStack []amd64.Register) {
	f.qStack = qStack
}

func (f *FFAmd64) UnsetQStack() {
	f.qStack = nil
}

func (f *FFAmd64) StackSize(maxNbRegistersNeeded, nbRegistersReserved, minStackSize int) int {
	got := amd64.NbRegisters - nbRegistersReserved
	r := got - maxNbRegistersNeeded
	if r >= 0 {
		return minStackSize
	}
	r *= -8
	return max(r, minStackSize)
}

func (f *FFAmd64) DefineFn(name string) (fn defineFn, err error) {
	name = strings.ToUpper(name)
	fn, ok := f.mDefines[name]
	if !ok {
		return nil, fmt.Errorf("function %s not defined", name)
	}
	return fn, nil
}

func (f *FFAmd64) CallDefine(name string, args ...any) {
	name = strings.ToUpper(name)
	fn, ok := f.mDefines[name]
	if !ok {
		panic(fmt.Sprintf("function %s not defined", name))
	}
	fn(args...)
}

func (f *FFAmd64) Define(name string, nbInputs int, fn defineFn, reuse bool) defineFn {

	inputs := make([]string, nbInputs)
	for i := range nbInputs {
		inputs[i] = fmt.Sprintf("in%d", i)
	}
	name = strings.ToUpper(name)

	for fn, ok := f.mDefines[name]; ok; {
		if reuse {
			// in that case, we don't redefine the define;
			// user explicitly asked for it
			return fn
		}
		panic("WARNING: not used at the moment, but if we reach this point, it means the function name already exist, for code generation purpose we add a suffix")
		// name already exist, for code generation purpose we add a suffix
		// should happen only with e2 deprecated functions
		// fmt.Println("WARNING: function name already defined, adding suffix")
		// i := 0
		// for {
		// 	newName := fmt.Sprintf("%s_%d", name, i)
		// 	if _, ok := f.mDefines[newName]; !ok {
		// 		name = newName
		// 		goto startDefine
		// 	}
		// 	i++
		// }
	}
	// startDefine:

	f.StartDefine()
	f.WriteLn("#define " + name + "(" + strings.Join(inputs, ", ") + ")")
	inputsRegisters := make([]any, nbInputs)
	for i := range nbInputs {
		inputsRegisters[i] = amd64.Register(inputs[i])
	}
	fn(inputsRegisters...)
	f.EndDefine()
	f.WriteLn("")

	toReturn := func(args ...any) {
		if len(args) != nbInputs {
			panic("invalid number of arguments")
		}
		inputsStr := make([]string, len(args))
		for i := range args {
			switch t := args[i].(type) {
			case amd64.Register:
				inputsStr[i] = string(t)
			case amd64.VectorRegister:
				inputsStr[i] = string(t)
			case amd64.MaskRegister:
				inputsStr[i] = string(t)
			case string:
				inputsStr[i] = t
			default:
				panic("invalid argument type")
			}
		}
		f.WriteLn(name + "(" + strings.Join(inputsStr, ", ") + ")")
	}

	f.mDefines[name] = toReturn

	return toReturn
}

func (f *FFAmd64) AssertCleanStack(reservedStackSize, minStackSize int) {
	if f.qStack != nil {
		panic("qStack not empty, use f.UnsetQStack()")
	}
	if f.nbElementsOnStack != 0 {
		panic(fmt.Sprintf("missing f.Push stack elements (NbWords=%d)", f.NbWords))
	}
	if reservedStackSize < minStackSize {
		panic(fmt.Sprintf("invalid minStackSize or reservedStackSize (NbWords=%d, reserved=%d, min=%d)", f.NbWords, reservedStackSize, minStackSize))
	}
	usedStackSize := f.maxOnStack * 8
	if usedStackSize > reservedStackSize {
		panic(fmt.Sprintf("using more stack size than reserved (NbWords=%d, reserved=%d, used=%d)", f.NbWords, reservedStackSize, usedStackSize))
	} else if max(usedStackSize, minStackSize) < reservedStackSize {
		// this panic is for dev purposes as this may be by design for alignment
		panic(fmt.Sprintf("reserved more stack size than needed (NbWords=%d, reserved=%d, used=%d)", f.NbWords, reservedStackSize, usedStackSize))
	}

	f.maxOnStack = 0
}

func (f *FFAmd64) Push(registers *amd64.Registers, rIn ...amd64.Register) {
	for _, r := range rIn {
		if strings.HasPrefix(string(r), "s") {
			// it's on the stack, decrease the offset
			f.nbElementsOnStack--
			continue
		}
		registers.Push(r)
	}
}

// UnsafePush behaves as Push, but doesn't check that the register is a valid register.
// This is useful when using R15 which is not included by default on the available registers.
func (f *FFAmd64) UnsafePush(registers *amd64.Registers, rIn ...amd64.Register) {
	for _, r := range rIn {
		if strings.HasPrefix(string(r), "s") {
			// it's on the stack, decrease the offset
			f.nbElementsOnStack--
			continue
		}
		registers.UnsafePush(r)
	}
}

func (f *FFAmd64) Pop(registers *amd64.Registers, forceStack ...bool) amd64.Register {
	if registers.Available() >= 1 && (len(forceStack) == 0 || !forceStack[0]) {
		return registers.Pop()
	}
	r := amd64.Register(fmt.Sprintf("s%d-%d(SP)", f.nbElementsOnStack, 8+f.nbElementsOnStack*8))
	f.nbElementsOnStack++
	if f.nbElementsOnStack > f.maxOnStack {
		f.maxOnStack = f.nbElementsOnStack
	}
	return r
}

func (f *FFAmd64) PopN(registers *amd64.Registers, forceStack ...bool) []amd64.Register {
	if len(forceStack) > 0 && forceStack[0] {
		nbStack := f.NbWords
		var u []amd64.Register

		for i := f.nbElementsOnStack; i < nbStack+f.nbElementsOnStack; i++ {
			u = append(u, amd64.Register(fmt.Sprintf("s%d-%d(SP)", i, 8+i*8)))
		}
		f.nbElementsOnStack += nbStack
		if f.nbElementsOnStack > f.maxOnStack {
			f.maxOnStack = f.nbElementsOnStack
		}
		return u
	}
	if registers.Available() >= f.NbWords {
		return registers.PopN(f.NbWords)
	}
	nbStack := f.NbWords - registers.Available()
	u := registers.PopN(registers.Available())

	for i := f.nbElementsOnStack; i < nbStack+f.nbElementsOnStack; i++ {
		u = append(u, amd64.Register(fmt.Sprintf("s%d-%d(SP)", i, 8+i*8)))
	}
	f.nbElementsOnStack += nbStack
	if f.nbElementsOnStack > f.maxOnStack {
		f.maxOnStack = f.nbElementsOnStack
	}
	return u
}

func (f *FFAmd64) qAt(index int) string {
	if f.qStack != nil {
		return string(f.qStack[index])
	}
	return fmt.Sprintf("·qElement+%d(SB)", index*8)
}

func (f *FFAmd64) qAt_u32(index int) string {
	if f.qStack != nil && f.NbWords == 4 {
		// so we have q on the stack as 4 uint64
		// but we want the addresses for 8 uint32
		// this is a not-future proof hack but should work for only current use case..;
		// ensure we have these:
		if f.qStack[0] != "s1-16(SP)" ||
			f.qStack[1] != "s2-24(SP)" ||
			f.qStack[2] != "s3-32(SP)" ||
			f.qStack[3] != "s4-40(SP)" {
			panic("qStack not initialized properly for qAt_bcst")
		}
		switch index {
		case 0:
			return "s10-16(SP)"
		case 1:
			return "s11-12(SP)" // stack grows down, so this is 12
		case 2:
			return "s20-24(SP)"
		case 3:
			return "s21-20(SP)"
		case 4:
			return "s30-32(SP)"
		case 5:
			return "s31-28(SP)"
		case 6:
			return "s40-40(SP)"
		case 7:
			return "s41-36(SP)"
		default:
			panic(fmt.Sprintf("invalid index %d for qAt_bcst", index))
		}

	}
	return fmt.Sprintf("·qElement+%d(SB)", index*4)
}

func (f *FFAmd64) qInv0() string {
	return "$const_qInvNeg"
}

func (f *FFAmd64) mu() string {
	return "$const_mu"
}

// GenerateCommonASM generates assembly code for the base field provided to the generator
// see internal/templates/ops*
func GenerateCommonASM(w io.Writer, nbWords, nbBits int, hasVector bool) error {
	f := NewFFAmd64(w, nbWords)
	f.Comment("Code generated by gnark-crypto/generator. DO NOT EDIT.")

	f.WriteLn("#include \"textflag.h\"")
	f.WriteLn("#include \"funcdata.h\"")
	f.WriteLn("#include \"go_asm.h\"")
	f.WriteLn("")

	if nbWords == 1 {
		if nbBits <= 31 {
			return GenerateF31ASM(f, hasVector)
		} else {
			panic("not implemented")
		}
	}

	f.GenerateReduceDefine()

	// reduce
	f.generateReduce()

	// mul by constants
	f.generateMulBy3()
	f.generateMulBy5()
	f.generateMulBy13()

	// fft butterflies
	f.generateButterfly()

	// mul
	f.generateMul(false)

	// from mont
	f.generateFromMont(false)

	if hasVector {
		f.WriteLn("")
		f.Comment("Vector operations are partially derived from Dag Arne Osvik's work in github.com/a16z/vectorized-fields")
		f.WriteLn("")

		f.generateAddVecW4()
		f.generateSubVecW4()
		f.generateSumVecW4()
		f.generateInnerProductW4()

		// IFMA-based vector operations (requires Ice Lake+ or Zen4+)
		if nbWords == 4 {
			f.WriteLn("")
			f.Comment("AVX-512 IFMA vector operations (requires Ice Lake+ or Zen4+)")
			f.WriteLn("")
			f.generateMulVecIFMA()
			f.generateScalarMulVecIFMA()
			// Note: innerProdVecIFMA is not generated as benchmarks showed IFMA
			// is not faster for inner product operations
		}
	}

	return nil
}

func GenerateF31FFTKernels(w io.Writer, nbBits int, kernels []int) error {
	if nbBits != 31 {
		return fmt.Errorf("only 31 bits supported for now")
	}
	f := NewFFAmd64(w, 1)

	f.WriteLn("")
	f.Comment("Code generated by gnark-crypto/generator. DO NOT EDIT.")
	f.Comment("Refer to the generator for more documentation.")
	f.Comment("Some sub-functions are derived from Plonky3:")
	f.Comment("https://github.com/Plonky3/Plonky3/blob/36e619f3c6526ee86e2e5639a24b3224e1c1700f/monty-31/src/x86_64_avx512/packing.rs#L319")
	f.WriteLn("")
	f.WriteLn("#include \"textflag.h\"")
	f.WriteLn("#include \"funcdata.h\"")
	f.WriteLn("#include \"go_asm.h\"")
	f.WriteLn("")

	f.generateFFTDefinesF31()
	f.generateFFTInnerDITF31()
	f.generateFFTInnerDIFF31()

	// unroll kernels for size 256 for now.
	fftKernels := []int{8}
	for _, ksize := range fftKernels {
		f.generateFFTKernelF31(ksize, true)
		f.generateFFTKernelF31(ksize, false)
	}

	return nil
}

func GenerateF31E4(w io.Writer) error {
	f := NewFFAmd64(w, 1)

	f.WriteLn("")
	f.Comment("Code generated by gnark-crypto/generator. DO NOT EDIT.")
	f.Comment("Refer to the generator for more documentation.")
	f.WriteLn("")
	f.WriteLn("#include \"textflag.h\"")
	f.WriteLn("#include \"funcdata.h\"")
	f.WriteLn("#include \"go_asm.h\"")
	f.WriteLn("")

	f.generateMulAccByElement()
	f.generateAddVecE4()
	f.generateSubVecE4()
	f.generateMulVecE4(e4VecMul)
	f.generateMulVecE4(e4VecScalarMul)
	f.generateMulVecE4(e4VecInnerProd)
	f.generateSumVecE4()
	f.generateMulVecElementE4()
	f.generateMulVecElementE6()
	f.generateMulAccVecElementE6()
	f.generateButterflyVecE4()
	f.generateButterflyPairVecE4()
	f.generateButterflyPairVecE6()
	f.generateInnerProdByElementE4()
	f.generateDITWithTwiddlesVecE6()
	f.generateDIFWithTwiddlesVecE6()

	return nil

}

func GenerateF31SIS(w io.Writer, nbBits int) error {
	if nbBits != 31 {
		return fmt.Errorf("only 31 bits supported for now")
	}
	f := NewFFAmd64(w, 1)

	f.WriteLn("")
	f.Comment("Code generated by gnark-crypto/generator. DO NOT EDIT.")
	f.Comment("Refer to the generator for more documentation.")
	f.Comment("Some sub-functions are derived from Plonky3:")
	f.Comment("https://github.com/Plonky3/Plonky3/blob/36e619f3c6526ee86e2e5639a24b3224e1c1700f/monty-31/src/x86_64_avx512/packing.rs#L319")
	f.WriteLn("")
	f.WriteLn("#include \"textflag.h\"")
	f.WriteLn("#include \"funcdata.h\"")
	f.WriteLn("#include \"go_asm.h\"")
	f.WriteLn("")

	f.generateFFTDefinesF31()

	f.generateSIS512_16F31()
	f.generateSISShuffleF31()
	f.generateSISUnhuffleF31()

	return nil
}

type Poseidon2Parameters struct {
	Width, FullRounds, PartialRounds, SBoxDegree int
	DiagInternal                                 []uint64
	HasCompressx16                               bool
}

func GenerateF31Poseidon2(w io.Writer, nbBits int, params []Poseidon2Parameters) error {
	if nbBits != 31 {
		return fmt.Errorf("only 31 bits supported for now")
	}
	f := NewFFAmd64(w, 1)

	f.WriteLn("")
	f.Comment("Code generated by gnark-crypto/generator. DO NOT EDIT.")
	f.Comment("Refer to the generator for more documentation.")
	f.WriteLn("")
	f.WriteLn("#include \"textflag.h\"")
	f.WriteLn("#include \"funcdata.h\"")
	f.WriteLn("#include \"go_asm.h\"")
	f.WriteLn("")

	for _, p := range params {
		f.generatePoseidon2_F31(p)

		if p.Width == 24 {
			f.generatePoseidon2_F31_16x24(p)
		}
		if p.Width == 16 && p.HasCompressx16 {
			f.generatePoseidon2_F31_16x16xN(p, true)
			f.generatePoseidon2_F31_16x16xN(p, false)
		}
	}

	return nil
}

func GenerateF31ASM(f *FFAmd64, hasVector bool) error {
	if !hasVector {
		return nil // nothing for now.
	}

	f.generateAddVecF31()
	f.generateSubVecF31()
	f.generateSumVecF31()
	f.generateSumVecSmallF31(16)
	f.generateSumVecSmallF31(24)
	f.generateMulVecF31()
	f.generateScalarMulVecF31()
	f.generateInnerProdVecF31()

	return nil
}

func ElementASMFileName(nbWords, nbBits int) string {
	const nameW1 = "element_%db_amd64.s"
	const nameWN = "element_%dw_amd64.s"

	const fW1 = "element_%db"
	const fWN = "element_%dw"

	if nbWords == 1 {
		return filepath.Join(fmt.Sprintf(fW1, nbBits), fmt.Sprintf(nameW1, nbBits))
	}
	return filepath.Join(fmt.Sprintf(fWN, nbWords), fmt.Sprintf(nameWN, nbWords))
}

func ElementASMBaseDir(nbWords, nbBits int) string {
	const fW1 = "element_%db"
	const fWN = "element_%dw"

	if nbWords == 1 {
		return fmt.Sprintf(fW1, 31)
	}
	return fmt.Sprintf(fWN, nbWords)
}
