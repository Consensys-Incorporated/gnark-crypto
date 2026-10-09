// Copyright 2020-2026 Consensys Software Inc.
// Licensed under the Apache License, Version 2.0. See the LICENSE file for details.

package fptower

import (
	"testing"

	"github.com/consensys/gnark-crypto/ecc/bw6-761/fp"
	"github.com/leanovate/gopter"
	"github.com/leanovate/gopter/prop"
)

// GenE6D generates an E6D elmt
func GenE6D() gopter.Gen {
	return gopter.CombineGens(
		GenFp(),
		GenFp(),
		GenFp(),
		GenFp(),
		GenFp(),
		GenFp(),
	).Map(func(values []any) *E6D {
		return &E6D{A0: values[0].(fp.Element), A1: values[1].(fp.Element), A2: values[2].(fp.Element), A3: values[3].(fp.Element), A4: values[4].(fp.Element), A5: values[5].(fp.Element)}
	})
}

// ------------------------------------------------------------
// tests
func TestE6DReceiverIsOperand(t *testing.T) {

	parameters := gopter.DefaultTestParameters()
	parameters.MinSuccessfulTests = 100

	properties := gopter.NewProperties(parameters)

	genA := GenE6D()
	genB := GenE6D()

	properties.Property("[(direct) BW6-761] conversion between direct and tower extensions", prop.ForAll(
		func(a *E6D) bool {
			b := ToTower(a)
			c := FromTower(b)
			return a.Equal(c)
		},
		genA,
	))

	properties.Property("[(direct) BW6-761] Having the receiver as operand (addition) should output the same result", prop.ForAll(
		func(a, b *E6D) bool {
			var c, d E6D
			d.Set(a)
			c.Add(a, b)
			a.Add(a, b)
			b.Add(&d, b)
			return a.Equal(b) && a.Equal(&c) && b.Equal(&c)
		},
		genA,
		genB,
	))

	properties.Property("[(direct) BW6-761] Having the receiver as operand (sub) should output the same result", prop.ForAll(
		func(a, b *E6D) bool {
			var c, d E6D
			d.Set(a)
			c.Sub(a, b)
			a.Sub(a, b)
			b.Sub(&d, b)
			return a.Equal(b) && a.Equal(&c) && b.Equal(&c)
		},
		genA,
		genB,
	))

	properties.Property("[(direct) BW6-761] Having the receiver as operand (mul) should output the same result", prop.ForAll(
		func(a, b *E6D) bool {
			var c, d E6D
			d.Set(a)
			c.Mul(a, b)
			a.Mul(a, b)
			b.Mul(&d, b)
			return a.Equal(b) && a.Equal(&c) && b.Equal(&c)
		},
		genA,
		genB,
	))

	properties.Property("[(direct) BW6-761] Having the receiver as operand (square) should output the same result", prop.ForAll(
		func(a *E6D) bool {
			var b E6D
			b.Square(a)
			a.Square(a)
			return a.Equal(&b)
		},
		genA,
	))

	properties.Property("[(direct) BW6-761] Having the receiver as operand (double) should output the same result", prop.ForAll(
		func(a *E6D) bool {
			var b E6D
			b.Double(a)
			a.Double(a)
			return a.Equal(&b)
		},
		genA,
	))

	properties.Property("[(direct) BW6-761] Having the receiver as operand (Inverse) should output the same result", prop.ForAll(
		func(a *E6D) bool {
			var b E6D
			b.Inverse(a)
			a.Inverse(a)
			return a.Equal(&b)
		},
		genA,
	))

	properties.Property("[(direct) BW6-761] Having the receiver as operand (Conjugate) should output the same result", prop.ForAll(
		func(a *E6D) bool {
			var b E6D
			b.Conjugate(a)
			a.Conjugate(a)
			return a.Equal(&b)
		},
		genA,
	))

	properties.TestingRun(t, gopter.ConsoleReporter(false))
}

func TestE6DOps(t *testing.T) {

	parameters := gopter.DefaultTestParameters()
	parameters.MinSuccessfulTests = 100

	properties := gopter.NewProperties(parameters)

	genA := GenE6D()
	genB := GenE6D()

	properties.Property("[(direct) BW6-761] sub & add should leave an element invariant", prop.ForAll(
		func(a, b *E6D) bool {
			var c E6D
			c.Set(a)
			c.Add(&c, b).Sub(&c, b)
			return c.Equal(a)
		},
		genA,
		genB,
	))

	properties.Property("[(direct) BW6-761] tower mul and direct mul are the same", prop.ForAll(
		func(a, b *E6D) bool {
			var c E6D
			c.Mul(a, b)
			var _c E6
			_a := ToTower(a)
			_b := ToTower(b)
			_c.Mul(_a, _b)
			return c.Equal(FromTower(&_c))
		},
		genA,
		genB,
	))

	properties.Property("[(direct) BW6-761] tower square and direct square are the same", prop.ForAll(
		func(a *E6D) bool {
			var c E6D
			c.Square(a)
			var _c E6
			_a := ToTower(a)
			_c.Square(_a)
			return c.Equal(FromTower(&_c))
		},
		genA,
	))

	properties.Property("[(direct) BW6-761] mul & inverse should leave an element invariant", prop.ForAll(
		func(a, b *E6D) bool {
			var c, d E6D
			d.Inverse(b)
			c.Set(a)
			c.Mul(&c, b).Mul(&c, &d)
			return c.Equal(a)
		},
		genA,
		genB,
	))

	properties.Property("[(direct) BW6-761] inverse twice should leave an element invariant", prop.ForAll(
		func(a *E6D) bool {
			var b E6D
			b.Inverse(a).Inverse(&b)
			return a.Equal(&b)
		},
		genA,
	))

	properties.Property("[(direct) BW6-761] square and mul should output the same result", prop.ForAll(
		func(a *E6D) bool {
			var b, c E6D
			b.Mul(a, a)
			c.Square(a)
			return b.Equal(&c)
		},
		genA,
	))

	properties.TestingRun(t, gopter.ConsoleReporter(false))
}

// bench
func BenchmarkE6DMulTower(b *testing.B) {
	var a, c E6D
	a.MustSetRandom()
	c.MustSetRandom()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		a.mulTower(&a, &c)
	}
}

func BenchmarkE6DMulMontgomery6(b *testing.B) {
	var a, c E6D
	a.MustSetRandom()
	c.MustSetRandom()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		a.mulMontgomery6(&a, &c)
	}
}
func BenchmarkE6DSquareTower(b *testing.B) {
	var a E6D
	a.MustSetRandom()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		a.squareTower(&a)
	}
}

func BenchmarkE6DSquareMontgomery6(b *testing.B) {
	var a E6D
	a.MustSetRandom()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		a.squareMontgomery6(&a)
	}
}

func TestE6DSetString(t *testing.T) {
	var a E6D
	if _, err := a.SetString("1", "2", "3", "4", "5", "6"); err != nil {
		t.Fatal(err)
	}
	if a.A5.String() != "6" {
		t.Fatalf("last component: expected 6, got %s", a.A5.String())
	}

	before := a
	res, err := a.SetString("1", "2", "3", "4", "5", "not-a-number")
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
