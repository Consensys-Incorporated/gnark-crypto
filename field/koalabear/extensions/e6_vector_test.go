// Copyright 2020-2026 Consensys Software Inc.
// Licensed under the Apache License, Version 2.0. See the LICENSE file for details.

package extensions

import (
	"strconv"
	"testing"

	fr "github.com/consensys/gnark-crypto/field/koalabear"
	"github.com/stretchr/testify/require"
)

func TestVectorE6ButterflyPair(t *testing.T) {
	for _, size := range []int{2, 4, 8, 16, 18} {
		t.Run(strconv.Itoa(size), func(t *testing.T) {
			a := make([]E6, size)
			for i := range a {
				setE6TestValue(&a[i], uint64(i*8+1))
			}
			expected := append([]E6(nil), a...)
			for i := 0; i < len(expected); i += 2 {
				ButterflyE6(&expected[i], &expected[i+1])
			}

			VectorE6(a).ButterflyPair()
			for i := range expected {
				require.Equal(t, expected[i], a[i], "element %d", i)
			}
		})
	}
}

func TestVectorE6FusedTwiddles(t *testing.T) {
	for _, size := range []int{2, 4, 8, 16, 18} {
		t.Run(strconv.Itoa(size), func(t *testing.T) {
			a0, a1, twiddles := makeE6FusedTestInputs(size)

			expected0 := append([]E6(nil), a0...)
			expected1 := append([]E6(nil), a1...)
			for i := range expected0 {
				expected1[i].MulByElement(&expected1[i], &twiddles[i])
				ButterflyE6(&expected0[i], &expected1[i])
			}

			VectorE6(a0).MulByElementThenButterfly(VectorE6(a1), twiddles)
			require.Equal(t, expected0, a0)
			require.Equal(t, expected1, a1)

			a0, a1, twiddles = makeE6FusedTestInputs(size)
			expected0 = append([]E6(nil), a0...)
			expected1 = append([]E6(nil), a1...)
			for i := range expected0 {
				ButterflyE6(&expected0[i], &expected1[i])
				expected1[i].MulByElement(&expected1[i], &twiddles[i])
			}

			VectorE6(a0).ButterflyThenMulByElement(VectorE6(a1), twiddles)
			require.Equal(t, expected0, a0)
			require.Equal(t, expected1, a1)
		})
	}
}

func TestVectorE6MulAccByElement(t *testing.T) {
	for _, size := range []int{0, 1, 7, 8, 16, 18, 100} {
		t.Run(strconv.Itoa(size), func(t *testing.T) {
			dst := make([]E6, size)
			a := make([]E6, size)
			b := make(fr.Vector, size)
			for i := range size {
				dst[i].MustSetRandom()
				a[i].MustSetRandom()
				b[i].MustSetRandom()
			}

			expected := append([]E6(nil), dst...)
			var tmp E6
			for i := range expected {
				tmp.MulByElement(&a[i], &b[i])
				expected[i].Add(&expected[i], &tmp)
			}

			VectorE6(dst).MulAccByElement(VectorE6(a), b)
			for i := range expected {
				require.Equal(t, expected[i], dst[i], "element %d", i)
			}
		})
	}
}

func BenchmarkVectorE6MulAccByElement(b *testing.B) {
	const size = 1 << 12
	dst := make([]E6, size)
	a := make([]E6, size)
	scale := make(fr.Vector, size)
	for i := range size {
		dst[i].MustSetRandom()
		a[i].MustSetRandom()
		scale[i].MustSetRandom()
	}
	b.ResetTimer()
	for range b.N {
		VectorE6(dst).MulAccByElement(VectorE6(a), scale)
	}
}

func BenchmarkVectorE6MulAccByElementGeneric(b *testing.B) {
	const size = 1 << 12
	dst := make([]E6, size)
	a := make([]E6, size)
	scale := make(fr.Vector, size)
	for i := range size {
		dst[i].MustSetRandom()
		a[i].MustSetRandom()
		scale[i].MustSetRandom()
	}
	b.ResetTimer()
	for range b.N {
		mulAccByElementE6Generic(dst, a, scale)
	}
}

func TestVectorE6ScalarMul(t *testing.T) {
	for _, size := range []int{0, 1, 8, 15, 16, 17, 31, 32, 33, 100} {
		t.Run(strconv.Itoa(size), func(t *testing.T) {
			a := make([]E6, size)
			res := make([]E6, size)
			var s E6
			s.MustSetRandom()
			for i := range size {
				a[i].MustSetRandom()
				res[i].MustSetRandom() // must be fully overwritten
			}

			expected := make([]E6, size)
			for i := range expected {
				expected[i].Mul(&a[i], &s)
			}

			VectorE6(res).ScalarMul(VectorE6(a), &s)
			for i := range expected {
				require.Equal(t, expected[i], res[i], "element %d", i)
			}
		})
	}
}

func TestVectorE6ScalarMulAcc(t *testing.T) {
	for _, size := range []int{0, 1, 8, 15, 16, 17, 31, 32, 33, 100} {
		t.Run(strconv.Itoa(size), func(t *testing.T) {
			dst := make([]E6, size)
			a := make([]E6, size)
			var s E6
			s.MustSetRandom()
			for i := range size {
				dst[i].MustSetRandom()
				a[i].MustSetRandom()
			}

			expected := append([]E6(nil), dst...)
			var tmp E6
			for i := range expected {
				tmp.Mul(&a[i], &s)
				expected[i].Add(&expected[i], &tmp)
			}

			VectorE6(dst).ScalarMulAcc(VectorE6(a), &s)
			for i := range expected {
				require.Equal(t, expected[i], dst[i], "element %d", i)
			}
		})
	}
}

func TestVectorE6ScalarMulAccByElement(t *testing.T) {
	for _, size := range []int{0, 1, 7, 8, 9, 16, 17, 100} {
		t.Run(strconv.Itoa(size), func(t *testing.T) {
			dst := make([]E6, size)
			a := make(fr.Vector, size)
			var s E6
			s.MustSetRandom()
			for i := range size {
				dst[i].MustSetRandom()
				a[i].MustSetRandom()
			}

			expected := append([]E6(nil), dst...)
			var tmp E6
			for i := range expected {
				tmp.MulByElement(&s, &a[i])
				expected[i].Add(&expected[i], &tmp)
			}

			VectorE6(dst).ScalarMulAccByElement(a, &s)
			for i := range expected {
				require.Equal(t, expected[i], dst[i], "element %d", i)
			}
		})
	}
}

func BenchmarkVectorE6ScalarMul(b *testing.B) {
	const size = 1 << 12
	res := make([]E6, size)
	a := make([]E6, size)
	var s E6
	s.MustSetRandom()
	for i := range size {
		a[i].MustSetRandom()
	}
	b.ResetTimer()
	for range b.N {
		VectorE6(res).ScalarMul(VectorE6(a), &s)
	}
}

func BenchmarkVectorE6ScalarMulAcc(b *testing.B) {
	const size = 1 << 12
	dst := make([]E6, size)
	a := make([]E6, size)
	var s E6
	s.MustSetRandom()
	for i := range size {
		dst[i].MustSetRandom()
		a[i].MustSetRandom()
	}
	b.ResetTimer()
	for range b.N {
		VectorE6(dst).ScalarMulAcc(VectorE6(a), &s)
	}
}

func BenchmarkVectorE6ScalarMulAccGeneric(b *testing.B) {
	const size = 1 << 12
	dst := make([]E6, size)
	a := make([]E6, size)
	var s E6
	s.MustSetRandom()
	for i := range size {
		dst[i].MustSetRandom()
		a[i].MustSetRandom()
	}
	b.ResetTimer()
	for range b.N {
		scalarMulAccE6Generic(dst, a, &s)
	}
}

func BenchmarkVectorE6ScalarMulAccByElement(b *testing.B) {
	const size = 1 << 12
	dst := make([]E6, size)
	a := make(fr.Vector, size)
	var s E6
	s.MustSetRandom()
	for i := range size {
		dst[i].MustSetRandom()
		a[i].MustSetRandom()
	}
	b.ResetTimer()
	for range b.N {
		VectorE6(dst).ScalarMulAccByElement(a, &s)
	}
}

func BenchmarkVectorE6ScalarMulAccByElementGeneric(b *testing.B) {
	const size = 1 << 12
	dst := make([]E6, size)
	a := make(fr.Vector, size)
	var s E6
	s.MustSetRandom()
	for i := range size {
		dst[i].MustSetRandom()
		a[i].MustSetRandom()
	}
	b.ResetTimer()
	for range b.N {
		scalarMulAccByElementE6Generic(dst, a, &s)
	}
}

func setE6TestValue(z *E6, base uint64) {
	z.B0.A0.SetUint64(base)
	z.B0.A1.SetUint64(base + 1)
	z.B1.A0.SetUint64(base + 2)
	z.B1.A1.SetUint64(base + 3)
	z.B2.A0.SetUint64(base + 4)
	z.B2.A1.SetUint64(base + 5)
}

func makeE6FusedTestInputs(size int) ([]E6, []E6, fr.Vector) {
	a0 := make([]E6, size)
	a1 := make([]E6, size)
	twiddles := make(fr.Vector, size)
	for i := range size {
		setE6TestValue(&a0[i], uint64(i*16+1))
		setE6TestValue(&a1[i], uint64(i*16+9))
		twiddles[i].SetUint64(uint64(i + 3))
	}
	return a0, a1, twiddles
}
