package arm64

import (
	"fmt"
)

// scalarVecOp selects the kernel emitted by generateScalarVecF31.
type scalarVecOp int

const (
	negVecOp           scalarVecOp = iota // res[i] = -a[i]
	scalarAddVecOp                        // res[i] = a[i] + c
	scalarSubFromVecOp                    // res[i] = c - a[i]
)

// generateScalarVecF31 emits one of the unary / broadcast-scalar kernels on
// base F31 vectors:
//
//	func negVec(res, a *Element, n uint64)              // res[i] = -a[i]
//	func scalarAddVec(res, a, c *Element, n uint64)     // res[i] = a[i] + c
//	func scalarSubFromVec(res, a, c *Element, n uint64) // res[i] = c - a[i]
//
// where n is the number of 4-element blocks. The scalar operand (*c, or zero
// for negVec) is broadcast once before the loop, so callers may pass a pointer
// into res; the loop then runs one addMod / subMod per block: for canonical a,
// -a = min(0 - a, q - a) (as uint32).
func (f *FFArm64) generateScalarVecF31(op scalarVecOp) {
	var name string
	argSize := 32
	switch op {
	case negVecOp:
		name = "negVec"
		argSize = 24
		f.Comment("negVec(res, a *Element, n uint64) res[0...n] = -a[0...n]")
	case scalarAddVecOp:
		name = "scalarAddVec"
		f.Comment("scalarAddVec(res, a, c *Element, n uint64) res[0...n] = a[0...n] + c")
	case scalarSubFromVecOp:
		name = "scalarSubFromVec"
		f.Comment("scalarSubFromVec(res, a, c *Element, n uint64) res[0...n] = c - a[0...n]")
	}
	f.Comment("n is the number of blocks of 4 uint32 to process")
	registers := f.FnHeader(name, 0, argSize)
	defer f.AssertCleanStack(0, 0)

	resPtr := registers.Pop()
	aPtr := registers.Pop()
	n := registers.Pop()

	f.LDP("res+0(FP)", resPtr, aPtr)

	a := registers.PopV()
	t := registers.PopV()
	c := registers.PopV()
	q := registers.PopV()

	f.VMOVS("$const_q", q)
	f.VDUP(q.SAt(0), q.S4(), "broadcast q into "+string(q))

	// c is the broadcast scalar operand: zero for negVec.
	if op == negVecOp {
		f.MOVD("n+16(FP)", n)
		f.VEOR(c.B16(), c.B16(), c.B16(), "c = 0")
	} else {
		cPtr := registers.Pop()
		tmp := registers.Pop()
		f.LDP("c+16(FP)", cPtr, n)
		f.MOVWU(fmt.Sprintf("(%s)", cPtr), tmp)
		f.VDUP(tmp, c.S4(), "broadcast c")
		registers.Push(cPtr, tmp)
	}

	const offset = 4 * 4 // 4 uint32 = 16 bytes per vector

	f.Loop(n, func() {
		f.VLD1_P(offset, aPtr, a.S4())
		switch op {
		case scalarAddVecOp:
			f.VADD(a.S4(), c.S4(), a.S4(), "a = a + c")
			f.VSUB(q.S4(), a.S4(), t.S4(), "t = a - q")
		default:
			f.VSUB(a.S4(), c.S4(), a.S4(), "a = c - a")
			f.VADD(a.S4(), q.S4(), t.S4(), "t = a + q")
		}
		f.VUMIN(t.S4(), a.S4(), a.S4(), "a = min(t, a)")
		f.VST1_P(a.S4(), resPtr, offset, "res = a")
	})

	registers.Push(resPtr, aPtr, n)
	registers.PushV(a, t, c, q)

	f.RET()
}
