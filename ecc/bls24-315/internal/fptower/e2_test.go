// Copyright 2020-2025 Consensys Software Inc.
// Licensed under the Apache License, Version 2.0. See the LICENSE file for details.

package fptower

import (
	"crypto/rand"
	"testing"

	"github.com/consensys/gnark-crypto/ecc/bls24-315/fp"
	"github.com/leanovate/gopter"
	"github.com/leanovate/gopter/prop"
)

// ------------------------------------------------------------
// tests

func TestE2ReceiverIsOperand(t *testing.T) {
	t.Parallel()
	parameters := gopter.DefaultTestParameters()
	parameters.MinSuccessfulTests = 100

	properties := gopter.NewProperties(parameters)

	genA := GenE2()
	genB := GenE2()
	genfp := GenFp()

	properties.Property("[BLS24-315] Having the receiver as operand (addition) should output the same result", prop.ForAll(
		func(a, b *E2) bool {
			var c, d E2
			d.Set(a)
			c.Add(a, b)
			a.Add(a, b)
			b.Add(&d, b)
			return a.Equal(b) && a.Equal(&c) && b.Equal(&c)
		},
		genA,
		genB,
	))

	properties.Property("[BLS24-315] Having the receiver as operand (sub) should output the same result", prop.ForAll(
		func(a, b *E2) bool {
			var c, d E2
			d.Set(a)
			c.Sub(a, b)
			a.Sub(a, b)
			b.Sub(&d, b)
			return a.Equal(b) && a.Equal(&c) && b.Equal(&c)
		},
		genA,
		genB,
	))

	properties.Property("[BLS24-315] Having the receiver as operand (mul) should output the same result", prop.ForAll(
		func(a, b *E2) bool {
			var c, d E2
			d.Set(a)
			c.Mul(a, b)
			a.Mul(a, b)
			b.Mul(&d, b)
			return a.Equal(b) && a.Equal(&c) && b.Equal(&c)
		},
		genA,
		genB,
	))

	properties.Property("[BLS24-315] Having the receiver as operand (square) should output the same result", prop.ForAll(
		func(a *E2) bool {
			var b E2
			b.Square(a)
			a.Square(a)
			return a.Equal(&b)
		},
		genA,
	))

	properties.Property("[BLS24-315] Having the receiver as operand (neg) should output the same result", prop.ForAll(
		func(a *E2) bool {
			var b E2
			b.Neg(a)
			a.Neg(a)
			return a.Equal(&b)
		},
		genA,
	))

	properties.Property("[BLS24-315] Having the receiver as operand (double) should output the same result", prop.ForAll(
		func(a *E2) bool {
			var b E2
			b.Double(a)
			a.Double(a)
			return a.Equal(&b)
		},
		genA,
	))

	properties.Property("[BLS24-315] Having the receiver as operand (mul by non residue) should output the same result", prop.ForAll(
		func(a *E2) bool {
			var b E2
			b.MulByNonResidue(a)
			a.MulByNonResidue(a)
			return a.Equal(&b)
		},
		genA,
	))

	properties.Property("[BLS24-315] Having the receiver as operand (mul by non residue inverse) should output the same result", prop.ForAll(
		func(a *E2) bool {
			var b E2
			b.MulByNonResidueInv(a)
			a.MulByNonResidueInv(a)
			return a.Equal(&b)
		},
		genA,
	))

	properties.Property("[BLS24-315] Having the receiver as operand (Inverse) should output the same result", prop.ForAll(
		func(a *E2) bool {
			var b E2
			b.Inverse(a)
			a.Inverse(a)
			return a.Equal(&b)
		},
		genA,
	))

	properties.Property("[BLS24-315] Having the receiver as operand (Conjugate) should output the same result", prop.ForAll(
		func(a *E2) bool {
			var b E2
			b.Conjugate(a)
			a.Conjugate(a)
			return a.Equal(&b)
		},
		genA,
	))

	properties.Property("[BLS24-315] Having the receiver as operand (mul by element) should output the same result", prop.ForAll(
		func(a *E2, b fp.Element) bool {
			var c E2
			c.MulByElement(a, &b)
			a.MulByElement(a, &b)
			return a.Equal(&c)
		},
		genA,
		genfp,
	))

	properties.Property("[BLS24-315] Having the receiver as operand (Sqrt) should output the same result", prop.ForAll(
		func(a *E2) bool {
			var b, c, d, s E2

			s.Square(a)
			a.Set(&s)
			b.Set(&s)

			a.Sqrt(a)
			b.Sqrt(&b)

			c.Square(a)
			d.Square(&b)
			return c.Equal(&d)
		},
		genA,
	))

	properties.TestingRun(t, gopter.ConsoleReporter(false))

}

func TestE2MulMaxed(t *testing.T) {
	t.Parallel()

	// let's pick a and b, with maxed A0 and A1
	var a, b E2
	fpMaxValue := fp.Element{
		8063698428123676673,
		4764498181658371330,
		16051339359738796768,
		15273757526516850351,
		342900304943437392,
	}
	fpMaxValue[0]--

	a.A0 = fpMaxValue
	a.A1 = fpMaxValue
	b.A0 = fpMaxValue
	b.A1 = fpMaxValue

	var c, d E2
	d.Inverse(&b)
	c.Set(&a)
	c.Mul(&c, &b).Mul(&c, &d)
	if !c.Equal(&a) {
		t.Fatal("mul with max fp failed")
	}
}

func TestE2Ops(t *testing.T) {
	t.Parallel()

	parameters := gopter.DefaultTestParameters()
	parameters.MinSuccessfulTests = 100

	properties := gopter.NewProperties(parameters)

	genA := GenE2()
	genB := GenE2()
	genfp := GenFp()

	properties.Property("[BLS24-315] sub & add should leave an element invariant", prop.ForAll(
		func(a, b *E2) bool {
			var c E2
			c.Set(a)
			c.Add(&c, b).Sub(&c, b)
			return c.Equal(a)
		},
		genA,
		genB,
	))

	properties.Property("[BLS24-315] mul & inverse should leave an element invariant", prop.ForAll(
		func(a, b *E2) bool {
			var c, d E2
			d.Inverse(b)
			c.Set(a)
			c.Mul(&c, b).Mul(&c, &d)
			return c.Equal(a)
		},
		genA,
		genB,
	))

	properties.Property("[BLS24-315] inverse twice should leave an element invariant", prop.ForAll(
		func(a *E2) bool {
			var b E2
			b.Inverse(a).Inverse(&b)
			return a.Equal(&b)
		},
		genA,
	))

	properties.Property("[BLS24-315] neg twice should leave an element invariant", prop.ForAll(
		func(a *E2) bool {
			var b E2
			b.Neg(a).Neg(&b)
			return a.Equal(&b)
		},
		genA,
	))

	properties.Property("[BLS24-315] mul and mulGeneric should output the same result", prop.ForAll(
		func(a, b *E2) bool {
			var c, d E2
			mulGenericE2(&c, a, b)
			d.Mul(a, b)
			return d.Equal(&c)
		},
		genA,
		genB,
	))

	properties.Property("[BLS24-315] square and mul should output the same result", prop.ForAll(
		func(a *E2) bool {
			var b, c E2
			b.Mul(a, a)
			c.Square(a)
			return b.Equal(&c)
		},
		genA,
	))

	properties.Property("[BLS24-315] MulByElement MulByElement inverse should leave an element invariant", prop.ForAll(
		func(a *E2, b fp.Element) bool {
			var c E2
			var d fp.Element
			d.Inverse(&b)
			c.MulByElement(a, &b).MulByElement(&c, &d)
			return c.Equal(a)
		},
		genA,
		genfp,
	))

	properties.Property("[BLS24-315] Double and mul by 2 should output the same result", prop.ForAll(
		func(a *E2) bool {
			var b E2
			var c fp.Element
			c.SetUint64(2)
			b.Double(a)
			a.MulByElement(a, &c)
			return a.Equal(&b)
		},
		genA,
	))

	properties.Property("[BLS24-315] Mulbynonres mulbynonresinv should leave the element invariant", prop.ForAll(
		func(a *E2) bool {
			var b E2
			b.MulByNonResidue(a).MulByNonResidueInv(&b)
			return a.Equal(&b)
		},
		genA,
	))

	properties.Property("[BLS24-315] a + pi(a), a-pi(a) should be real", prop.ForAll(
		func(a *E2) bool {
			var b, c, d E2
			var e, f fp.Element
			b.Conjugate(a)
			c.Add(a, &b)
			d.Sub(a, &b)
			e.Double(&a.A0)
			f.Double(&a.A1)
			return c.A1.IsZero() && d.A0.IsZero() && e.Equal(&c.A0) && f.Equal(&d.A1)
		},
		genA,
	))

	properties.Property("[BLS24-315] Legendre on square should output 1", prop.ForAll(
		func(a *E2) bool {
			var b E2
			b.Square(a)
			c := b.Legendre()
			return c == 1
		},
		genA,
	))

	properties.Property("[BLS24-315] square(sqrt) should leave an element invariant", prop.ForAll(
		func(a *E2) bool {
			var b, c, d, e E2
			b.Square(a)
			c.Sqrt(&b)
			d.Square(&c)
			e.Neg(a)
			return (c.Equal(a) || c.Equal(&e)) && d.Equal(&b)
		},
		genA,
	))

	// Regression test for the silent failure of E2.Sqrt on purely-real
	// inputs whose Fp coordinate is a non-residue. The Scott §6.3 formula
	// has an unstated precondition x.A1 != 0; for x = (a, 0) with a non-QR
	// in Fp the inner Fp.Sqrt failed silently and the function returned
	// a wrong root. Since this curve's E2 is hand-written (not generated
	// from the tower template), the template-level regression does not
	// cover it; we add a curve-local property here.
	properties.Property("[BLS24-315] square(sqrt) should be invariant for purely-real inputs", prop.ForAll(
		func(a fp.Element) bool {
			var x, root, sq E2
			x.A0.Set(&a)
			// (a, 0) is always an E2-square for a != 0: either a is QR in
			// Fp, or a/β is (since β is itself a non-residue).
			root.Sqrt(&x)
			sq.Square(&root)
			return sq.Equal(&x)
		},
		genfp,
	))

	properties.Property("[BLS24-315] neg(E2) == neg(E2.A0, E2.A1)", prop.ForAll(
		func(a *E2) bool {
			var b, c E2
			b.Neg(a)
			c.A0.Neg(&a.A0)
			c.A1.Neg(&a.A1)
			return c.Equal(&b)
		},
		genA,
	))

	properties.Property("[BLS24-315] Cmp and LexicographicallyLargest should be consistent", prop.ForAll(
		func(a *E2) bool {
			var negA E2
			negA.Neg(a)
			cmpResult := a.Cmp(&negA)
			lResult := a.LexicographicallyLargest()
			if lResult && cmpResult == 1 {
				return true
			}
			if !lResult && cmpResult != 1 {
				return true
			}
			return false
		},
		genA,
	))

	properties.TestingRun(t, gopter.ConsoleReporter(false))

}

// ------------------------------------------------------------
// benches

func BenchmarkE2Add(b *testing.B) {
	var a, c E2
	a.MustSetRandom()
	c.MustSetRandom()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		a.Add(&a, &c)
	}
}

func BenchmarkE2Sub(b *testing.B) {
	var a, c E2
	a.MustSetRandom()
	c.MustSetRandom()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		a.Sub(&a, &c)
	}
}

func BenchmarkE2Mul(b *testing.B) {
	var a, c E2
	a.MustSetRandom()
	c.MustSetRandom()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		a.Mul(&a, &c)
	}
}

func BenchmarkE2MulByElement(b *testing.B) {
	var a E2
	var c fp.Element
	c.MustSetRandom()
	a.MustSetRandom()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		a.MulByElement(&a, &c)
	}
}

func BenchmarkE2Square(b *testing.B) {
	var a E2
	a.MustSetRandom()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		a.Square(&a)
	}
}

func BenchmarkE2Sqrt(b *testing.B) {
	var a E2
	a.MustSetRandom()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		a.Sqrt(&a)
	}
}

func BenchmarkE2Exp(b *testing.B) {
	var x E2
	x.MustSetRandom()
	b1, _ := rand.Int(rand.Reader, fp.Modulus())
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		x.Exp(x, b1)
	}
}

func BenchmarkE2Inverse(b *testing.B) {
	var a E2
	a.MustSetRandom()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		a.Inverse(&a)
	}
}

func BenchmarkE2MulNonRes(b *testing.B) {
	var a E2
	a.MustSetRandom()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		a.MulByNonResidue(&a)
	}
}

func BenchmarkE2MulNonResInv(b *testing.B) {
	var a E2
	a.MustSetRandom()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		a.MulByNonResidueInv(&a)
	}
}

func BenchmarkE2Conjugate(b *testing.B) {
	var a E2
	a.MustSetRandom()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		a.Conjugate(&a)
	}
}

func TestE2Div(t *testing.T) {

	parameters := gopter.DefaultTestParameters()
	properties := gopter.NewProperties(parameters)

	genA := GenE2()
	genB := GenE2()

	properties.Property("[BLS24-317] dividing then multiplying by the same element does nothing", prop.ForAll(
		func(a, b *E2) bool {
			var c E2
			c.Div(a, b)
			c.Mul(&c, b)
			return c.Equal(a)
		},
		genA,
		genB,
	))

	properties.TestingRun(t, gopter.ConsoleReporter(false))
}

// TestE2AsmVsGeneric checks the assembly implementations against the generic
// ones on boundary values (0, 1, p-1 components) and aliased arguments.
func TestE2AsmVsGeneric(t *testing.T) {
	one := fp.One()
	var pMinus1 fp.Element
	pMinus1.Neg(&one)
	var zero fp.Element
	specials := []fp.Element{zero, one, pMinus1}
	var inputs []E2
	for _, a0 := range specials {
		for _, a1 := range specials {
			inputs = append(inputs, E2{A0: a0, A1: a1})
		}
	}
	for i := range inputs {
		for j := range inputs {
			var got, want E2
			mulGenericE2(&want, &inputs[i], &inputs[j])
			got.Mul(&inputs[i], &inputs[j])
			if !got.Equal(&want) {
				t.Fatalf("mul mismatch on boundary values (%d, %d)", i, j)
			}
		}
	}

	// aliasing
	for i := 0; i < 100; i++ {
		var x, y E2
		x.A0.MustSetRandom()
		x.A1.MustSetRandom()
		y.A0.MustSetRandom()
		y.A1.MustSetRandom()

		xCopy, yCopy := x, y
		var want E2
		mulGenericE2(&want, &x, &y)

		x.Mul(&x, &y)
		if !x.Equal(&want) {
			t.Fatal("aliasing z==x mismatch")
		}
		x = xCopy
		y.Mul(&x, &y)
		if !y.Equal(&want) {
			t.Fatal("aliasing z==y mismatch")
		}
		y = yCopy
	}
}
