package mamabear

import (
	"fmt"
	"math/bits"
	"testing"
)

func montMulTrace(a, b uint64) uint64 {
	hi, lo := bits.Mul64(a, b)
	c0 := lo & rMask
	t := (lo >> rBits) | (hi << (64 - rBits))
	m0 := (c0 * qInvNeg) & rMask
	mHi, mLo := bits.Mul64(m0, q)
	qVal := (mLo >> rBits) | (mHi << (64 - rBits))
	carry := (c0 + (mLo & rMask)) >> rBits
	result := q + t + qVal + carry
	fmt.Printf("  montMul(%d, %d):\n", a, b)
	fmt.Printf("    hi=%d lo=%d c0=%d t=%d\n", hi, lo, c0, t)
	fmt.Printf("    m0=%d qVal=%d carry=%d\n", m0, qVal, carry)
	fmt.Printf("    q+t+qVal+carry = %d + %d + %d + %d = %d (q=%d)\n", q, t, qVal, carry, result, q)
	return result
}

func TestMontMulTrace(t *testing.T) {
	a_raw := uint64(441918283826716)
	b_raw := uint64(425375270931891)
	c_raw := uint64(96862663551149)

	a_mont := montMul(a_raw, rSquareRaw)
	if a_mont >= q {
		a_mont -= q
	}
	b_mont := montMul(b_raw, rSquareRaw)
	if b_mont >= q {
		b_mont -= q
	}
	c_mont := montMul(c_raw, rSquareRaw)
	if c_mont >= q {
		c_mont -= q
	}

	fmt.Printf("a_mont=%d b_mont=%d c_mont=%d\n", a_mont, b_mont, c_mont)

	// Compute bc = b * c
	fmt.Println("\n=== Computing b * c ===")
	bc_raw := montMulTrace(b_mont, c_mont)
	bc := bc_raw
	if bc >= q {
		bc -= q
	}
	fmt.Printf("bc_raw=%d bc=%d (>= q? %v)\n", bc_raw, bc, bc_raw >= q)

	// Compute a * bc
	fmt.Println("\n=== Computing a * (b*c) ===")
	a_bc_raw := montMulTrace(a_mont, bc)
	a_bc := a_bc_raw
	if a_bc >= q {
		a_bc -= q
	}
	fmt.Printf("a_bc_raw=%d a_bc=%d (>= q? %v)\n", a_bc_raw, a_bc, a_bc_raw >= q)

	fmt.Printf("\nFinal: a*(b*c) = %d\n", a_bc)
	fmt.Printf("q = %d, 2q-1 = %d\n", q, 2*q-1)
	fmt.Printf("a_bc_raw >= q: %v, a_bc_raw >= 2q: %v\n", a_bc_raw >= q, a_bc_raw >= 2*q)
}
