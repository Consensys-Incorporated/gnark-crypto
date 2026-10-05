// Copyright 2020-2025 Consensys Software Inc.
// Licensed under the Apache License, Version 2.0. See the LICENSE file for details.

package fptower

import (
	"math/big"

	"github.com/consensys/gnark-crypto/ecc/bls24-317/fp"
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

// SetString sets a E2 element from strings
func (z *E2) SetString(s1, s2 string) (*E2, error) {
	if _, err := z.A0.SetString(s1); err != nil {
		return nil, err
	}
	if _, err := z.A1.SetString(s2); err != nil {
		return nil, err
	}
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
// It panics if reading from crypto/rand fails.
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

// Sub subtracts two elements of E2
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

func init() {
	q := fp.Modulus()
	tmp := big.NewInt(3)
	sqrtExp1.Set(q).Sub(&sqrtExp1, tmp).Rsh(&sqrtExp1, 2)

	tmp.SetUint64(1)
	sqrtExp2.Set(q).Sub(&sqrtExp2, tmp).Rsh(&sqrtExp2, 1)
}

var sqrtExp1, sqrtExp2 big.Int

// Sqrt sets z to the square root of x and returns z
// The function does not test whether the square root
// exists or not, it's up to the caller to call
// Legendre beforehand.
//
// "Optimized One-Dimensional SQIsign Verification
// on Intel and Cortex-M4" by Aardal et al.
// https://eprint.iacr.org/2024/1563.pdf (algo 3)
func (z *E2) Sqrt(a *E2) *E2 {
	// Aardal Algorithm 3 has an unstated precondition a.A1 != 0: for
	// a.A1 == 0 with a.A0 a non-residue in Fp the inner ExpBySqrtPp1o4
	// yields δ = -a.A0, then x0 = a.A0 + δ = 0 cascades to (0, 0). But
	// (a.A0, 0) is always a square in Fp² for nonzero a.A0 (Euler). For
	// bls24-317 we have u² = -1 so β = -1, and the true sqrt is
	// (0, sqrt(-a.A0)).
	if a.A1.IsZero() {
		if a.A0.Legendre() >= 0 {
			z.A0.Sqrt(&a.A0)
			z.A1.SetZero()
			return z
		}
		var aOverBeta fp.Element
		aOverBeta.Neg(&a.A0)
		z.A0.SetZero()
		z.A1.Sqrt(&aOverBeta)
		return z
	}

	var delta, x0, x1, t0, t1 fp.Element

	x0.Square(&a.A0)
	delta.Square(&a.A1).
		Add(&delta, &x0).
		ExpBySqrtPp1o4(delta)

	x0.Add(&a.A0, &delta)

	t0.Double(&x0)

	x1.ExpBySqrtPm3o4(t0)

	x0.Mul(&x0, &x1)

	x1.Mul(&a.A1, &x1)

	t1.Double(&x0).
		Square(&t1)

	if t1.Equal(&t0) {
		z.A0.Set(&x0)
		z.A1.Set(&x1)
		return z
	} else {
		z.A0.Set(&x1)
		z.A1.Neg(&x0)
		return z
	}
}

// Select is conditional move.
// If cond = 0, it sets z to caseZ and returns it. otherwise caseNz.
func (z *E2) Select(cond int, caseZ *E2, caseNz *E2) *E2 {
	//Might be able to save a nanosecond or two by an aggregate implementation

	z.A0.Select(cond, &caseZ.A0, &caseNz.A0)
	z.A1.Select(cond, &caseZ.A1, &caseNz.A1)

	return z
}

// Div divides an element in E2 by an element in E2
func (z *E2) Div(x *E2, y *E2) *E2 {
	var r E2
	r.Inverse(y).Mul(x, &r)
	return z.Set(&r)
}
