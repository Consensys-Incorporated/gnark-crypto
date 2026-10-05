// Copyright 2020-2025 Consensys Software Inc.
// Licensed under the Apache License, Version 2.0. See the LICENSE file for details.

package fptower

import (
	"math/big"

	"github.com/consensys/gnark-crypto/ecc/bls24-315/fp"
)

// E2 is a degree two finite field extension of fp.Element
type E2 struct {
	A0, A1 fp.Element
}

// Equal returns true if z equals x, false otherwise
func (z *E2) Equal(x *E2) bool {
	return z.A0.Equal(&x.A0) && z.A1.Equal(&x.A1)
}

// Cmp compares (lexicographic order) z and x and returns:
//
//	-1 if z <  x
//	 0 if z == x
//	+1 if z >  x
func (z *E2) Cmp(x *E2) int {
	if a1 := z.A1.Cmp(&x.A1); a1 != 0 {
		return a1
	}
	return z.A0.Cmp(&x.A0)
}

// LexicographicallyLargest returns true if this element is strictly lexicographically
// larger than its negation, false otherwise
func (z *E2) LexicographicallyLargest() bool {
	// adapted from github.com/zkcrypto/bls12_381
	if z.A1.IsZero() {
		return z.A0.LexicographicallyLargest()
	}
	return z.A1.LexicographicallyLargest()
}

// SetString sets z from strings and returns z.
// If any component fails to parse, z is left unchanged and SetString returns nil, error.
func (z *E2) SetString(s1, s2 string) (*E2, error) {
	var t E2
	if _, err := t.A0.SetString(s1); err != nil {
		return nil, err
	}
	if _, err := t.A1.SetString(s2); err != nil {
		return nil, err
	}
	*z = t
	return z, nil
}

// SetZero sets an E2 elmt to zero
func (z *E2) SetZero() *E2 {
	z.A0.SetZero()
	z.A1.SetZero()
	return z
}

// Set sets an E2 from x
func (z *E2) Set(x *E2) *E2 {
	z.A0 = x.A0
	z.A1 = x.A1
	return z
}

// SetOne sets z to 1 in Montgomery form and returns z
func (z *E2) SetOne() *E2 {
	z.A0.SetOne()
	z.A1.SetZero()
	return z
}

// SetRandom sets a0 and a1 to random values
func (z *E2) SetRandom() (*E2, error) {
	if _, err := z.A0.SetRandom(); err != nil {
		return nil, err
	}
	if _, err := z.A1.SetRandom(); err != nil {
		return nil, err
	}
	return z, nil
}

// MustSetRandom sets z to a random value.
// It panics if reading from crypto/rand fails
func (z *E2) MustSetRandom() *E2 {
	if _, err := z.SetRandom(); err != nil {
		panic(err)
	}
	return z
}

// IsZero returns true if z is zero, false otherwise
func (z *E2) IsZero() bool {
	return z.A0.IsZero() && z.A1.IsZero()
}

// IsOne returns true if z is one, false otherwise
func (z *E2) IsOne() bool {
	return z.A0.IsOne() && z.A1.IsZero()
}

// Add adds two elements of E2
func (z *E2) Add(x, y *E2) *E2 {
	addE2(z, x, y)
	return z
}

// Sub two elements of E2
func (z *E2) Sub(x, y *E2) *E2 {
	subE2(z, x, y)
	return z
}

// Double doubles an E2 element
func (z *E2) Double(x *E2) *E2 {
	doubleE2(z, x)
	return z
}

// Neg negates an E2 element
func (z *E2) Neg(x *E2) *E2 {
	negE2(z, x)
	return z
}

// String implements Stringer interface for fancy printing
func (z *E2) String() string {
	return (z.A0.String() + "+" + z.A1.String() + "*u")
}

// MulByElement multiplies an element in E2 by an element in fp
func (z *E2) MulByElement(x *E2, y *fp.Element) *E2 {
	var yCopy fp.Element
	yCopy.Set(y)
	z.A0.Mul(&x.A0, &yCopy)
	z.A1.Mul(&x.A1, &yCopy)
	return z
}

// Conjugate conjugates an element in E2
func (z *E2) Conjugate(x *E2) *E2 {
	z.A0 = x.A0
	z.A1.Neg(&x.A1)
	return z
}

// Legendre returns the Legendre symbol of z
func (z *E2) Legendre() int {
	var n fp.Element
	z.norm(&n)
	return n.Legendre()
}

// Exp sets z=xᵏ (mod q²) and returns it
func (z *E2) Exp(x E2, k *big.Int) *E2 {
	if k.IsUint64() && k.Uint64() == 0 {
		return z.SetOne()
	}

	e := k
	if k.Sign() == -1 {
		// negative k, we invert
		// if k < 0: xᵏ (mod q²) == (x⁻¹)ᵏ (mod q²)
		x.Inverse(&x)

		// we negate k in a temp big.Int since
		// Int.Bit(_) of k and -k is different
		e = bigIntPool.Get().(*big.Int)
		defer bigIntPool.Put(e)
		e.Neg(k)
	}

	z.SetOne()
	b := e.Bytes()
	for i := range b {
		w := b[i]
		for j := range 8 {
			z.Square(z)
			if (w & (0b10000000 >> j)) != 0 {
				z.Mul(z, &x)
			}
		}
	}

	return z
}

// Sqrt sets z to the square root of and returns z
// The function does not test whether the square root
// exists or not, it's up to the caller to call
// Legendre beforehand.
//
// "A note on the calculation of some functions in
// finite fields: Tricks of the Trade" by Michael Scott
// https://eprint.iacr.org/2020/1497.pdf (Sec. 6.3)
func (z *E2) Sqrt(x *E2) *E2 {
	// Scott §6.3 has an unstated precondition x.A1 != 0: for x.A1 == 0 with
	// x.A0 a non-residue in Fp the inner Fp.Sqrt fails silently and the
	// final Div goes through zero, returning a wrong root. But (x.A0, 0) is
	// always a square in Fp² for nonzero x.A0 (Euler). For bls24-315 we have
	// u² = 13 so β = 13, and the true sqrt is (0, sqrt(x.A0 / 13)).
	if x.A1.IsZero() {
		if x.A0.Legendre() >= 0 {
			z.A0.Sqrt(&x.A0)
			z.A1.SetZero()
			return z
		}
		var beta, aOverBeta fp.Element
		beta.SetUint64(13)
		aOverBeta.Inverse(&beta)
		aOverBeta.Mul(&aOverBeta, &x.A0)
		z.A0.SetZero()
		z.A1.Sqrt(&aOverBeta)
		return z
	}

	var x0, x1 fp.Element
	x.norm(&x0)
	x0.Sqrt(&x0)
	x1.Add(&x.A0, &x0).Halve()
	if x1.Legendre() != 1 {
		x1.Sub(&x.A0, &x0).Halve()
	}
	x1.Sqrt(&x1)
	z.A0.Set(&x1)
	x1.Double(&x1)
	z.A1.Div(&x.A1, &x1)

	return z
}

func (z *E2) Div(x *E2, y *E2) *E2 {
	var r E2
	r.Inverse(y).Mul(x, &r)
	return z.Set(&r)
}
