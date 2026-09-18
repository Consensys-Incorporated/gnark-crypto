// Copyright 2020-2026 Consensys Software Inc.
// Licensed under the Apache License, Version 2.0. See the LICENSE file for details.

package extensions_test

import (
	"math/big"
	"testing"

	fr "github.com/consensys/gnark-crypto/field/mamabear"
	"github.com/consensys/gnark-crypto/field/mamabear/extensions"
)

func randE3(t *testing.T) extensions.E3 {
	t.Helper()
	var a extensions.E3
	a.MustSetRandom()
	return a
}

func TestE3Add(t *testing.T) {
	// commutativity: a + b == b + a
	for range 100 {
		a, b := randE3(t), randE3(t)
		var ab, ba extensions.E3
		ab.Add(&a, &b)
		ba.Add(&b, &a)
		if !ab.Equal(&ba) {
			t.Fatal("Add not commutative")
		}
	}
	// associativity: (a+b)+c == a+(b+c)
	for range 100 {
		a, b, c := randE3(t), randE3(t), randE3(t)
		var lhs, rhs, tmp extensions.E3
		tmp.Add(&a, &b)
		lhs.Add(&tmp, &c)
		tmp.Add(&b, &c)
		rhs.Add(&a, &tmp)
		if !lhs.Equal(&rhs) {
			t.Fatal("Add not associative")
		}
	}
}

func TestE3Sub(t *testing.T) {
	// a - a == 0
	for range 100 {
		a := randE3(t)
		var z extensions.E3
		z.Sub(&a, &a)
		if !z.IsZero() {
			t.Fatal("a - a != 0")
		}
	}
	// a - b == a + (-b)
	for range 100 {
		a, b := randE3(t), randE3(t)
		var lhs, rhs, negb extensions.E3
		lhs.Sub(&a, &b)
		negb.Neg(&b)
		rhs.Add(&a, &negb)
		if !lhs.Equal(&rhs) {
			t.Fatal("Sub inconsistent with Neg+Add")
		}
	}
}

func TestE3MulCommutativity(t *testing.T) {
	for range 100 {
		a, b := randE3(t), randE3(t)
		var ab, ba extensions.E3
		ab.Mul(&a, &b)
		ba.Mul(&b, &a)
		if !ab.Equal(&ba) {
			t.Fatal("Mul not commutative")
		}
	}
}

func TestE3MulAssociativity(t *testing.T) {
	for range 50 {
		a, b, c := randE3(t), randE3(t), randE3(t)
		var lhs, rhs, tmp extensions.E3
		tmp.Mul(&a, &b)
		lhs.Mul(&tmp, &c)
		tmp.Mul(&b, &c)
		rhs.Mul(&a, &tmp)
		if !lhs.Equal(&rhs) {
			t.Fatal("Mul not associative")
		}
	}
}

func TestE3MulDistributivity(t *testing.T) {
	for range 50 {
		a, b, c := randE3(t), randE3(t), randE3(t)
		var lhs, rhs, bc, tmp extensions.E3
		bc.Add(&b, &c)
		lhs.Mul(&a, &bc)
		tmp.Mul(&a, &b)
		rhs.Mul(&a, &c)
		rhs.Add(&rhs, &tmp)
		if !lhs.Equal(&rhs) {
			t.Fatal("Mul not distributive over Add")
		}
	}
}

func TestE3Square(t *testing.T) {
	for range 100 {
		a := randE3(t)
		var sq, mul extensions.E3
		sq.Square(&a)
		mul.Mul(&a, &a)
		if !sq.Equal(&mul) {
			t.Fatal("Square inconsistent with Mul")
		}
	}
}

func TestE3Inverse(t *testing.T) {
	// a * a^{-1} == 1
	for range 100 {
		a := randE3(t)
		var inv, prod extensions.E3
		inv.Inverse(&a)
		prod.Mul(&a, &inv)
		if !prod.IsOne() {
			t.Fatal("a * Inverse(a) != 1")
		}
	}
	// zero inverse is zero
	var z extensions.E3
	z.Inverse(&z)
	if !z.IsZero() {
		t.Fatal("Inverse(0) != 0")
	}
}

func TestE3MulByNonResidue(t *testing.T) {
	// MulByNonResidue is the same as multiplying by (0,1,0) in E3
	for range 100 {
		a := randE3(t)
		var nr, lhs, rhs extensions.E3
		nr.A1.SetOne() // t = (0, 1, 0)
		lhs.MulByNonResidue(&a)
		rhs.Mul(&a, &nr)
		if !lhs.Equal(&rhs) {
			t.Fatal("MulByNonResidue inconsistent with Mul by t")
		}
	}
}

func TestE3FrobeniusOrder(t *testing.T) {
	// φ³ == identity: Frobenius has order 3 over F_p
	for range 20 {
		a := randE3(t)
		var b, c extensions.E3
		b.Frobenius(&a)
		c.Frobenius(&b)
		b.Frobenius(&c)
		if !b.Equal(&a) {
			t.Fatal("Frobenius does not have order 3")
		}
	}
}

func TestE3FrobeniusLinearity(t *testing.T) {
	// φ(a+b) == φ(a) + φ(b)
	for range 50 {
		a, b := randE3(t), randE3(t)
		var lhs, rhs, tmp extensions.E3
		tmp.Add(&a, &b)
		lhs.Frobenius(&tmp)
		var fa, fb extensions.E3
		fa.Frobenius(&a)
		fb.Frobenius(&b)
		rhs.Add(&fa, &fb)
		if !lhs.Equal(&rhs) {
			t.Fatal("Frobenius not F_p-linear over Add")
		}
	}
}

func TestE3Exp(t *testing.T) {
	// a^1 == a
	for range 20 {
		a := randE3(t)
		var b extensions.E3
		b.Exp(a, big.NewInt(1))
		if !b.Equal(&a) {
			t.Fatal("Exp(a,1) != a")
		}
	}
	// a^0 == 1
	for range 20 {
		a := randE3(t)
		var b extensions.E3
		b.Exp(a, big.NewInt(0))
		if !b.IsOne() {
			t.Fatal("Exp(a,0) != 1")
		}
	}
	// a^2 == Square(a)
	for range 50 {
		a := randE3(t)
		var sq, exp extensions.E3
		sq.Square(&a)
		exp.Exp(a, big.NewInt(2))
		if !sq.Equal(&exp) {
			t.Fatal("Exp(a,2) != Square(a)")
		}
	}
	// a^{|F_{p³}|} == a (little Fermat for E3)
	// |F_{p³}| = p³; check a^p³ == a for random elements
	for range 5 {
		a := randE3(t)
		var b extensions.E3
		p := new(big.Int).SetUint64(562932773552129)
		p3 := new(big.Int).Exp(p, big.NewInt(3), nil)
		b.Exp(a, p3)
		if !b.Equal(&a) {
			t.Fatal("a^{p³} != a (Fermat's little theorem for E3 failed)")
		}
	}
}

func TestBatchInvertE3(t *testing.T) {
	const n = 64
	a := make([]extensions.E3, n)
	for i := range a {
		a[i].MustSetRandom()
	}
	// Mix in a zero element
	a[n/2].SetZero()

	inv := extensions.BatchInvertE3(a)

	for i, ai := range a {
		if ai.IsZero() {
			if !inv[i].IsZero() {
				t.Fatalf("BatchInvertE3: inv[%d] should be zero for zero input", i)
			}
			continue
		}
		var prod extensions.E3
		prod.Mul(&ai, &inv[i])
		if !prod.IsOne() {
			t.Fatalf("BatchInvertE3: a[%d] * inv[%d] != 1", i, i)
		}
	}
}

// ---- Vector tests ------------------------------------------------------------

func TestE3VectorButterfly(t *testing.T) {
	const n = 64
	a := make(extensions.Vector, n)
	b := make(extensions.Vector, n)
	for i := range n {
		a[i].MustSetRandom()
		b[i].MustSetRandom()
	}
	aOrig := make(extensions.Vector, n)
	bOrig := make(extensions.Vector, n)
	copy(aOrig, a)
	copy(bOrig, b)

	a.Butterfly(b)

	for i := range n {
		var expectedA, expectedB extensions.E3
		expectedA.Add(&aOrig[i], &bOrig[i])
		expectedB.Sub(&aOrig[i], &bOrig[i])
		if !a[i].Equal(&expectedA) {
			t.Fatalf("Butterfly[%d]: a wrong", i)
		}
		if !b[i].Equal(&expectedB) {
			t.Fatalf("Butterfly[%d]: b wrong", i)
		}
	}
}

func TestE3VectorButterflyPair(t *testing.T) {
	const n = 64
	v := make(extensions.Vector, n)
	for i := range n {
		v[i].MustSetRandom()
	}
	orig := make(extensions.Vector, n)
	copy(orig, v)

	v.ButterflyPair()

	for i := 0; i < n; i += 2 {
		var ea, eb extensions.E3
		ea.Add(&orig[i], &orig[i+1])
		eb.Sub(&orig[i], &orig[i+1])
		if !v[i].Equal(&ea) {
			t.Fatalf("ButterflyPair[%d]: wrong", i)
		}
		if !v[i+1].Equal(&eb) {
			t.Fatalf("ButterflyPair[%d]: wrong", i+1)
		}
	}
}

func TestE3VectorMulByElement(t *testing.T) {
	const n = 32
	a := make(extensions.Vector, n)
	var scalars [n]fr.Element
	for i := range n {
		a[i].MustSetRandom()
		scalars[i].MustSetRandom()
	}

	res := make(extensions.Vector, n)
	res.MulByElement(a, scalars[:])

	for i := range n {
		var expected extensions.E3
		expected.MulByElement(&a[i], &scalars[i])
		if !res[i].Equal(&expected) {
			t.Fatalf("MulByElement[%d] wrong", i)
		}
	}
}

func TestE3VectorScalarMulByElement(t *testing.T) {
	const n = 32
	a := make(extensions.Vector, n)
	for i := range n {
		a[i].MustSetRandom()
	}
	var s fr.Element
	s.MustSetRandom()

	res := make(extensions.Vector, n)
	res.ScalarMulByElement(a, &s)

	for i := range n {
		var expected extensions.E3
		expected.MulByElement(&a[i], &s)
		if !res[i].Equal(&expected) {
			t.Fatalf("ScalarMulByElement[%d] wrong", i)
		}
	}
}

// ---- Benchmarks ---------------------------------------------------------------

var sinkE3 extensions.E3

func BenchmarkE3Mul(b *testing.B) {
	var x, y extensions.E3
	x.MustSetRandom()
	y.MustSetRandom()
	b.ResetTimer()
	for range b.N {
		sinkE3.Mul(&x, &y)
	}
}

func BenchmarkE3Square(b *testing.B) {
	var x extensions.E3
	x.MustSetRandom()
	b.ResetTimer()
	for range b.N {
		sinkE3.Square(&x)
	}
}

func BenchmarkE3Inverse(b *testing.B) {
	var x extensions.E3
	x.MustSetRandom()
	b.ResetTimer()
	for range b.N {
		sinkE3.Inverse(&x)
	}
}

func BenchmarkE3Frobenius(b *testing.B) {
	var x extensions.E3
	x.MustSetRandom()
	b.ResetTimer()
	for range b.N {
		sinkE3.Frobenius(&x)
	}
}

func BenchmarkE3Exp(b *testing.B) {
	var x extensions.E3
	x.MustSetRandom()
	e := new(big.Int).SetUint64(562932773552129) // p
	b.ResetTimer()
	for range b.N {
		sinkE3.Exp(x, e)
	}
}
