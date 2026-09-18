// Copyright 2020-2026 Consensys Software Inc.
// Licensed under the Apache License, Version 2.0. See the LICENSE file for details.

package mamabear

import (
	"crypto/rand"
	"math/big"
	"testing"

	"github.com/leanovate/gopter"
	ggen "github.com/leanovate/gopter/gen"
	"github.com/leanovate/gopter/prop"
	"github.com/stretchr/testify/require"
)

// -------------------------------------------------------------------------------------------------
// Benchmarks

var benchResElement Element

func BenchmarkElementMul(b *testing.B) {
	var x Element
	x.MustSetRandom()
	benchResElement.SetOne()
	b.ResetTimer()
	for range b.N {
		benchResElement.Mul(&benchResElement, &x)
	}
}

func BenchmarkElementAdd(b *testing.B) {
	var x Element
	x.MustSetRandom()
	benchResElement.MustSetRandom()
	b.ResetTimer()
	for range b.N {
		benchResElement.Add(&x, &benchResElement)
	}
}

func BenchmarkElementSub(b *testing.B) {
	var x Element
	x.MustSetRandom()
	benchResElement.MustSetRandom()
	b.ResetTimer()
	for range b.N {
		benchResElement.Sub(&x, &benchResElement)
	}
}

func BenchmarkElementSquare(b *testing.B) {
	benchResElement.MustSetRandom()
	b.ResetTimer()
	for range b.N {
		benchResElement.Square(&benchResElement)
	}
}

func BenchmarkElementInverse(b *testing.B) {
	var x Element
	x.MustSetRandom()
	b.ResetTimer()
	for range b.N {
		benchResElement.Inverse(&x)
	}
}

func BenchmarkElementSqrt(b *testing.B) {
	var a Element
	a.MustSetRandom()
	a.Square(&a)
	b.ResetTimer()
	for range b.N {
		benchResElement.Sqrt(&a)
	}
}

func BenchmarkElementCbrt(b *testing.B) {
	var a Element
	a.SetUint64(8)
	b.ResetTimer()
	for range b.N {
		benchResElement.Cbrt(&a)
	}
}

func BenchmarkElementExp(b *testing.B) {
	var x Element
	x.MustSetRandom()
	e, _ := rand.Int(rand.Reader, Modulus())
	b.ResetTimer()
	for range b.N {
		benchResElement.Exp(x, e)
	}
}

func BenchmarkElementButterfly(b *testing.B) {
	var x Element
	x.MustSetRandom()
	benchResElement.MustSetRandom()
	b.ResetTimer()
	for range b.N {
		Butterfly(&x, &benchResElement)
	}
}

// -------------------------------------------------------------------------------------------------
// Helpers

const (
	nbFuzzShort = 200
	nbFuzz      = 1000
)

type testPairElement struct {
	element Element
	bigint  big.Int
}

func (t testPairElement) String() string { return t.element.String() }

func gen() gopter.Gen {
	return ggen.UInt64Range(0, q-1).Map(func(v uint64) testPairElement {
		var e Element
		e[0] = v
		e.toMont()
		var bInt big.Int
		e.BigInt(&bInt)
		return testPairElement{element: e, bigint: bInt}
	})
}

func genFull() gopter.Gen {
	return ggen.UInt64Range(0, q-1).Map(func(v uint64) Element {
		var e Element
		e[0] = v
		e.toMont()
		return e
	})
}

// bigIntMod returns (a op b) mod p using big.Int as the reference implementation.
func modAdd(a, b *big.Int) *big.Int {
	r := new(big.Int).Add(a, b)
	return r.Mod(r, Modulus())
}

func modSub(a, b *big.Int) *big.Int {
	r := new(big.Int).Sub(a, b)
	return r.Mod(r, Modulus())
}

func modMul(a, b *big.Int) *big.Int {
	r := new(big.Int).Mul(a, b)
	return r.Mod(r, Modulus())
}

func modNeg(a *big.Int) *big.Int {
	r := new(big.Int).Neg(a)
	return r.Mod(r, Modulus())
}

// -------------------------------------------------------------------------------------------------
// Properties

func TestElementReduce(t *testing.T) {
	t.Parallel()
	parameters := gopter.DefaultTestParameters()
	if testing.Short() {
		parameters.MinSuccessfulTests = nbFuzzShort
	} else {
		parameters.MinSuccessfulTests = nbFuzz
	}
	properties := gopter.NewProperties(parameters)
	genA := genFull()

	properties.Property("reduce should output a result smaller than modulus", prop.ForAll(
		func(a Element) bool {
			b := a
			reduce(&a)
			_reduceGeneric(&b)
			return a.smallerThanModulus() && a.Equal(&b)
		},
		genA,
	))
	properties.TestingRun(t, gopter.ConsoleReporter(false))
}

func TestElementAdd(t *testing.T) {
	t.Parallel()
	parameters := gopter.DefaultTestParameters()
	if testing.Short() {
		parameters.MinSuccessfulTests = nbFuzzShort
	} else {
		parameters.MinSuccessfulTests = nbFuzz
	}
	properties := gopter.NewProperties(parameters)
	genA := gen()
	genB := gen()

	properties.Property("add: output matches big.Int reference", prop.ForAll(
		func(a, b testPairElement) bool {
			var z Element
			z.Add(&a.element, &b.element)
			var bz big.Int
			z.BigInt(&bz)
			return modAdd(&a.bigint, &b.bigint).Cmp(&bz) == 0
		},
		genA, genB,
	))
	properties.Property("add: commutative", prop.ForAll(
		func(a, b testPairElement) bool {
			var z1, z2 Element
			z1.Add(&a.element, &b.element)
			z2.Add(&b.element, &a.element)
			return z1.Equal(&z2)
		},
		genA, genB,
	))
	properties.TestingRun(t, gopter.ConsoleReporter(false))
}

func TestElementSub(t *testing.T) {
	t.Parallel()
	parameters := gopter.DefaultTestParameters()
	if testing.Short() {
		parameters.MinSuccessfulTests = nbFuzzShort
	} else {
		parameters.MinSuccessfulTests = nbFuzz
	}
	properties := gopter.NewProperties(parameters)
	genA := gen()
	genB := gen()

	properties.Property("sub: output matches big.Int reference", prop.ForAll(
		func(a, b testPairElement) bool {
			var z Element
			z.Sub(&a.element, &b.element)
			var bz big.Int
			z.BigInt(&bz)
			return modSub(&a.bigint, &b.bigint).Cmp(&bz) == 0
		},
		genA, genB,
	))
	properties.TestingRun(t, gopter.ConsoleReporter(false))
}

func TestElementMul(t *testing.T) {
	t.Parallel()
	parameters := gopter.DefaultTestParameters()
	if testing.Short() {
		parameters.MinSuccessfulTests = nbFuzzShort
	} else {
		parameters.MinSuccessfulTests = nbFuzz
	}
	properties := gopter.NewProperties(parameters)
	genA := gen()
	genB := gen()

	properties.Property("mul: output matches big.Int reference", prop.ForAll(
		func(a, b testPairElement) bool {
			var z Element
			z.Mul(&a.element, &b.element)
			var bz big.Int
			z.BigInt(&bz)
			return modMul(&a.bigint, &b.bigint).Cmp(&bz) == 0
		},
		genA, genB,
	))
	properties.Property("mul: commutative", prop.ForAll(
		func(a, b testPairElement) bool {
			var z1, z2 Element
			z1.Mul(&a.element, &b.element)
			z2.Mul(&b.element, &a.element)
			return z1.Equal(&z2)
		},
		genA, genB,
	))
	properties.Property("mul: associative", prop.ForAll(
		func(a, b, c testPairElement) bool {
			var z1, z2 Element
			z1.Mul(&a.element, &b.element).Mul(&z1, &c.element)
			z2.Mul(&b.element, &c.element).Mul(&a.element, &z2)
			return z1.Equal(&z2)
		},
		genA, genB, gen(),
	))
	properties.Property("mul: Fermat's little theorem x^p == x", prop.ForAll(
		func(a testPairElement) bool {
			var b, one Element
			one.SetOne()
			b.Exp(a.element, Modulus())
			return b.Equal(&a.element)
		},
		genA,
	))
	properties.TestingRun(t, gopter.ConsoleReporter(false))
}

func TestElementNeg(t *testing.T) {
	t.Parallel()
	parameters := gopter.DefaultTestParameters()
	if testing.Short() {
		parameters.MinSuccessfulTests = nbFuzzShort
	} else {
		parameters.MinSuccessfulTests = nbFuzz
	}
	properties := gopter.NewProperties(parameters)
	genA := gen()

	properties.Property("neg(0) == 0", func(_ *gopter.GenParameters) *gopter.PropResult {
		var a Element
		a.SetZero()
		a.Neg(&a)
		if a.IsZero() {
			return &gopter.PropResult{Status: gopter.PropTrue}
		}
		return &gopter.PropResult{Status: gopter.PropFalse}
	})
	properties.Property("neg: output matches big.Int reference", prop.ForAll(
		func(a testPairElement) bool {
			var z Element
			z.Neg(&a.element)
			var bz big.Int
			z.BigInt(&bz)
			return modNeg(&a.bigint).Cmp(&bz) == 0
		},
		genA,
	))
	properties.Property("x + neg(x) == 0", prop.ForAll(
		func(a testPairElement) bool {
			var z Element
			z.Neg(&a.element)
			z.Add(&z, &a.element)
			return z.IsZero()
		},
		genA,
	))
	properties.TestingRun(t, gopter.ConsoleReporter(false))
}

func TestElementInverse(t *testing.T) {
	t.Parallel()
	parameters := gopter.DefaultTestParameters()
	if testing.Short() {
		parameters.MinSuccessfulTests = nbFuzzShort
	} else {
		parameters.MinSuccessfulTests = nbFuzz
	}
	properties := gopter.NewProperties(parameters)
	genA := gen()

	properties.Property("x * inv(x) == 1 for x != 0", prop.ForAll(
		func(a testPairElement) bool {
			if a.element.IsZero() {
				return true
			}
			var z Element
			z.Inverse(&a.element)
			z.Mul(&z, &a.element)
			var one Element
			one.SetOne()
			return z.Equal(&one)
		},
		genA,
	))
	properties.Property("inv == x^(p-2)", prop.ForAll(
		func(a testPairElement) bool {
			if a.element.IsZero() {
				return true
			}
			exp := new(big.Int).Sub(Modulus(), big.NewInt(2))
			var z1, z2 Element
			z1.Inverse(&a.element)
			z2.Exp(a.element, exp)
			return z1.Equal(&z2)
		},
		genA,
	))
	properties.TestingRun(t, gopter.ConsoleReporter(false))
}

func TestElementSqrt(t *testing.T) {
	t.Parallel()
	parameters := gopter.DefaultTestParameters()
	if testing.Short() {
		parameters.MinSuccessfulTests = nbFuzzShort
	} else {
		parameters.MinSuccessfulTests = nbFuzz
	}
	properties := gopter.NewProperties(parameters)
	genA := gen()

	properties.Property("sqrt(x^2) roundtrip: z^2 == x^2", prop.ForAll(
		func(a testPairElement) bool {
			// a^2 is always a QR
			var sq, root Element
			sq.Square(&a.element)
			root.Sqrt(&sq)
			var check Element
			check.Square(&root)
			return check.Equal(&sq)
		},
		genA,
	))
	properties.Property("Sqrt and Legendre consistent: Legendre(x^2) == 1", prop.ForAll(
		func(a testPairElement) bool {
			if a.element.IsZero() {
				return true
			}
			var sq Element
			sq.Square(&a.element)
			return sq.Legendre() == 1
		},
		genA,
	))
	properties.TestingRun(t, gopter.ConsoleReporter(false))
}

func TestElementCbrt(t *testing.T) {
	t.Parallel()
	parameters := gopter.DefaultTestParameters()
	if testing.Short() {
		parameters.MinSuccessfulTests = nbFuzzShort
	} else {
		parameters.MinSuccessfulTests = nbFuzz
	}
	properties := gopter.NewProperties(parameters)
	genA := gen()

	properties.Property("cbrt(x)^3 == x for p ≡ 2 (mod 3)", prop.ForAll(
		func(a testPairElement) bool {
			var root Element
			root.Cbrt(&a.element)
			root.Square(&root).Mul(&root, &a.element) // root^2 * a... wrong
			// Actually: compute root^3 properly
			var cube Element
			root.Cbrt(&a.element)
			cube.Square(&root)
			cube.Mul(&cube, &root)
			return cube.Equal(&a.element)
		},
		genA,
	))
	properties.TestingRun(t, gopter.ConsoleReporter(false))
}

func TestElementBytes(t *testing.T) {
	t.Parallel()
	parameters := gopter.DefaultTestParameters()
	if testing.Short() {
		parameters.MinSuccessfulTests = nbFuzzShort
	} else {
		parameters.MinSuccessfulTests = nbFuzz
	}
	properties := gopter.NewProperties(parameters)
	genA := gen()

	properties.Property("SetBytes(Bytes()) roundtrip", prop.ForAll(
		func(a testPairElement) bool {
			var b Element
			bytes := a.element.Bytes()
			b.SetBytes(bytes[:])
			return a.element.Equal(&b)
		},
		genA,
	))
	properties.Property("BigInt(SetBigInt()) roundtrip", prop.ForAll(
		func(a testPairElement) bool {
			var b Element
			b.SetBigInt(&a.bigint)
			var bOut big.Int
			b.BigInt(&bOut)
			return a.bigint.Cmp(&bOut) == 0
		},
		genA,
	))
	properties.TestingRun(t, gopter.ConsoleReporter(false))
}

func TestElementCmp(t *testing.T) {
	var x, y Element
	require.Equal(t, 0, x.Cmp(&y), "x == y for zero values")

	var one Element
	one.SetOne()
	y.Sub(&y, &one) // y = p-1

	require.Equal(t, -1, x.Cmp(&y), "0 < p-1")
	require.Equal(t, 1, y.Cmp(&x), "p-1 > 0")

	x = y
	require.Equal(t, 0, x.Cmp(&y), "x == y after copy")
}

func TestElementNegZero(t *testing.T) {
	var a, b Element
	b.SetZero()
	a.Neg(&b)
	require.True(t, a.IsZero(), "neg(0) == 0")
}

func TestElementOne(t *testing.T) {
	var one Element
	one.SetOne()
	// 1 * 1 == 1
	var z Element
	z.Mul(&one, &one)
	require.True(t, z.Equal(&one), "1 * 1 == 1")
}

func TestFermat(t *testing.T) {
	// x^p == x for all x ∈ F_p (Fermat's little theorem)
	for range 20 {
		var x Element
		x.MustSetRandom()
		var xp Element
		xp.Exp(x, Modulus())
		require.True(t, xp.Equal(&x), "x^p == x")
	}
}

func TestMontRoundtrip(t *testing.T) {
	// fromMont(toMont(x)) == x
	for v := uint64(0); v < 1000; v++ {
		var x Element
		x.SetUint64(v)
		// x is now in Montgomery form
		// Convert back to canonical
		x.fromMont()
		require.Equal(t, v%q, x[0], "fromMont(toMont(%d)) = %d, want %d", v, x[0], v%q)
	}
}

func TestReduceFast(t *testing.T) {
	// ReduceFast should reduce values in [0, 2^52) to [0, 2p)
	p := q
	for _, v := range []uint64{0, 1, p - 1, p, p + 1, 2*p - 1, (1 << 52) - 1} {
		if v >= 1<<52 {
			continue
		}
		got := ReduceFast(v)
		// The canonical check: got ≡ v (mod p) and got < 2*p
		want := v % p
		got2 := got
		if got2 >= p {
			got2 -= p
		}
		require.Equal(t, want, got2, "ReduceFast(%d) mod p mismatch", v)
		require.Less(t, got, 2*p, "ReduceFast(%d) = %d should be < 2p", v, got)
	}
}

func FuzzMul(f *testing.F) {
	f.Add(uint64(0), uint64(0))
	f.Add(uint64(1), uint64(1))
	f.Add(uint64(q-1), uint64(q-1))
	f.Add(uint64(0), uint64(q-1))

	f.Fuzz(func(t *testing.T, a0, b0 uint64) {
		a0 = a0 % q
		b0 = b0 % q
		var a, b, z Element
		a.SetUint64(a0)
		b.SetUint64(b0)
		z.Mul(&a, &b)

		// Reference: (a * b) mod p using big.Int
		var bA, bB big.Int
		a.BigInt(&bA)
		b.BigInt(&bB)
		ref := new(big.Int).Mul(&bA, &bB)
		ref.Mod(ref, Modulus())

		var bZ big.Int
		z.BigInt(&bZ)
		if ref.Cmp(&bZ) != 0 {
			t.Fatalf("Mul(%d, %d): got %s, want %s", a0, b0, &bZ, ref)
		}
	})
}
