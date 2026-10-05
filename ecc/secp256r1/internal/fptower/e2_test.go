package fptower

import (
	"testing"

	"github.com/consensys/gnark-crypto/ecc/secp256r1/fp"
	"github.com/leanovate/gopter"
	"github.com/leanovate/gopter/prop"
)

const (
	nbFuzzShort = 10
	nbFuzz      = 50
)

func TestE2ReceiverIsOperand(t *testing.T) {

	t.Parallel()
	parameters := gopter.DefaultTestParameters()
	if testing.Short() {
		parameters.MinSuccessfulTests = nbFuzzShort
	} else {
		parameters.MinSuccessfulTests = nbFuzz
	}

	properties := gopter.NewProperties(parameters)

	genA := GenE2()
	genB := GenE2()
	genfp := GenFp()

	properties.Property("[P256] Having the receiver as operand (addition) should output the same result", prop.ForAll(
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

	properties.Property("[P256] Having the receiver as operand (sub) should output the same result", prop.ForAll(
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

	properties.Property("[P256] Having the receiver as operand (mul) should output the same result", prop.ForAll(
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

	properties.Property("[P256] Having the receiver as operand (square) should output the same result", prop.ForAll(
		func(a *E2) bool {
			var b E2
			b.Square(a)
			a.Square(a)
			return a.Equal(&b)
		},
		genA,
	))

	properties.Property("[P256] Having the receiver as operand (neg) should output the same result", prop.ForAll(
		func(a *E2) bool {
			var b E2
			b.Neg(a)
			a.Neg(a)
			return a.Equal(&b)
		},
		genA,
	))

	properties.Property("[P256] Having the receiver as operand (double) should output the same result", prop.ForAll(
		func(a *E2) bool {
			var b E2
			b.Double(a)
			a.Double(a)
			return a.Equal(&b)
		},
		genA,
	))

	properties.Property("[P256] Having the receiver as operand (Inverse) should output the same result", prop.ForAll(
		func(a *E2) bool {
			var b E2
			b.Inverse(a)
			a.Inverse(a)
			return a.Equal(&b)
		},
		genA,
	))

	properties.Property("[P256] Having the receiver as operand (Conjugate) should output the same result", prop.ForAll(
		func(a *E2) bool {
			var b E2
			b.Conjugate(a)
			a.Conjugate(a)
			return a.Equal(&b)
		},
		genA,
	))

	properties.Property("[P256] Having the receiver as operand (mul by element) should output the same result", prop.ForAll(
		func(a *E2, b fp.Element) bool {
			var c E2
			c.MulByElement(a, &b)
			a.MulByElement(a, &b)
			return a.Equal(&c)
		},
		genA,
		genfp,
	))

	properties.Property("[P256] Having the receiver as operand (Sqrt) should output the same result", prop.ForAll(
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

func TestE2Ops(t *testing.T) {

	t.Parallel()
	parameters := gopter.DefaultTestParameters()
	if testing.Short() {
		parameters.MinSuccessfulTests = nbFuzzShort
	} else {
		parameters.MinSuccessfulTests = nbFuzz
	}

	properties := gopter.NewProperties(parameters)

	genA := GenE2()
	genB := GenE2()
	genfp := GenFp()

	properties.Property("[P256] sub & add should leave an element invariant", prop.ForAll(
		func(a, b *E2) bool {
			var c E2
			c.Set(a)
			c.Add(&c, b).Sub(&c, b)
			return c.Equal(a)
		},
		genA,
		genB,
	))

	properties.Property("[P256] mul & inverse should leave an element invariant", prop.ForAll(
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

	properties.Property("[P256] inverse twice should leave an element invariant", prop.ForAll(
		func(a *E2) bool {
			var b E2
			b.Inverse(a).Inverse(&b)
			return a.Equal(&b)
		},
		genA,
	))

	properties.Property("[P256] neg twice should leave an element invariant", prop.ForAll(
		func(a *E2) bool {
			var b E2
			b.Neg(a).Neg(&b)
			return a.Equal(&b)
		},
		genA,
	))

	properties.Property("[P256] square and mul should output the same result", prop.ForAll(
		func(a *E2) bool {
			var b, c E2
			b.Mul(a, a)
			c.Square(a)
			return b.Equal(&c)
		},
		genA,
	))

	properties.Property("[P256] MulByElement MulByElement inverse should leave an element invariant", prop.ForAll(
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

	properties.Property("[P256] Double and mul by 2 should output the same result", prop.ForAll(
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

	properties.Property("[P256] a + pi(a), a-pi(a) should be real", prop.ForAll(
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

	properties.Property("[P256] Legendre on square should output 1", prop.ForAll(
		func(a *E2) bool {
			var b E2
			b.Square(a)
			c := b.Legendre()
			return c == 1
		},
		genA,
	))

	properties.Property("[P256] square(sqrt) should leave an element invariant", prop.ForAll(
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

	properties.Property("[P256] cube(cbrt) should leave an element invariant", prop.ForAll(
		func(a *E2) bool {
			var b, c, d E2
			b.Square(a).Mul(&b, a) // b = a³
			result := c.Cbrt(&b)
			if result == nil {
				return false
			}
			d.Square(&c).Mul(&d, &c) // d = c³
			return d.Equal(&b)
		},
		genA,
	))

	properties.Property("[P256] neg(E2) == neg(E2.A0, E2.A1)", prop.ForAll(
		func(a *E2) bool {
			var b, c E2
			b.Neg(a)
			c.A0.Neg(&a.A0)
			c.A1.Neg(&a.A1)
			return c.Equal(&b)
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
	for range b.N {
		a.Add(&a, &c)
	}
}

func BenchmarkE2Sub(b *testing.B) {
	var a, c E2
	a.MustSetRandom()
	c.MustSetRandom()
	b.ResetTimer()
	for range b.N {
		a.Sub(&a, &c)
	}
}

func BenchmarkE2Mul(b *testing.B) {
	var a, c E2
	a.MustSetRandom()
	c.MustSetRandom()
	b.ResetTimer()
	for range b.N {
		a.Mul(&a, &c)
	}
}

func BenchmarkE2Square(b *testing.B) {
	var a E2
	a.MustSetRandom()
	b.ResetTimer()
	for range b.N {
		a.Square(&a)
	}
}

func BenchmarkE2Inverse(b *testing.B) {
	var a E2
	a.MustSetRandom()
	b.ResetTimer()
	for range b.N {
		a.Inverse(&a)
	}
}

func BenchmarkE2Sqrt(b *testing.B) {
	var a E2
	a.MustSetRandom()
	a.Square(&a)
	b.ResetTimer()
	for range b.N {
		a.Sqrt(&a)
	}
}

func BenchmarkE2Cbrt(b *testing.B) {
	var a, c E2
	a.MustSetRandom()
	c.Square(&a).Mul(&c, &a)
	b.ResetTimer()
	for range b.N {
		a.Cbrt(&c)
	}
}

func BenchmarkE2Exp(b *testing.B) {
	var x E2
	x.MustSetRandom()
	b.ResetTimer()
	for range b.N {
		x.Exp(x, fp.Modulus())
	}
}

func TestE2SetString(t *testing.T) {
	var a E2
	if _, err := a.SetString("1", "2"); err != nil {
		t.Fatal(err)
	}
	if a.A1.String() != "2" {
		t.Fatalf("last component: expected 2, got %s", a.A1.String())
	}

	before := a
	res, err := a.SetString("1", "not-a-number")
	if err == nil {
		t.Fatal("expected an error for an invalid component")
	}
	if res != nil {
		t.Fatal("expected a nil result on error")
	}
	if a != before {
		t.Fatal("receiver must be left unchanged on error")
	}
}
