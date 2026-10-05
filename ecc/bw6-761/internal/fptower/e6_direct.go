// Copyright 2020-2026 Consensys Software Inc.
// Licensed under the Apache License, Version 2.0. See the LICENSE file for details.

package fptower

import (
	"github.com/consensys/gnark-crypto/ecc/bw6-761/fp"
)

// E6D is a degree 6 finite field extension of fp
type E6D struct {
	A0, A1, A2, A3, A4, A5 fp.Element
}

// Equal returns true if z equals x, false otherwise
func (z *E6D) Equal(x *E6D) bool {
	return z.A0.Equal(&x.A0) &&
		z.A1.Equal(&x.A1) &&
		z.A2.Equal(&x.A2) &&
		z.A3.Equal(&x.A3) &&
		z.A4.Equal(&x.A4) &&
		z.A5.Equal(&x.A5)
}

// String puts E6D elmt in string form
func (z *E6D) String() string {
	return (z.A0.String() + "+(" + z.A1.String() + ")*u+(" + z.A2.String() + ")*u**2+(" + z.A3.String() + ")*u**3+(" + z.A4.String() + ")*u**4+(" + z.A5.String() + ")*u**5")
}

// SetString sets z from coefficient strings and returns z.
// It returns nil and the subfield SetString error when a coefficient cannot be parsed.
func (z *E6D) SetString(s1, s2, s3, s4, s5, s6 string) (*E6D, error) {
	if _, err := z.A0.SetString(s1); err != nil {
		return nil, err
	}
	if _, err := z.A1.SetString(s2); err != nil {
		return nil, err
	}
	if _, err := z.A2.SetString(s3); err != nil {
		return nil, err
	}
	if _, err := z.A3.SetString(s4); err != nil {
		return nil, err
	}
	if _, err := z.A4.SetString(s5); err != nil {
		return nil, err
	}
	if _, err := z.A5.SetString(s6); err != nil {
		return nil, err
	}
	return z, nil
}

// Set copies x into z and returns z
func (z *E6D) Set(x *E6D) *E6D {
	*z = *x
	return z
}

// SetOne sets z to 1 form and returns z
func (z *E6D) SetOne() *E6D {
	z.A0.SetOne()
	z.A1.SetZero()
	z.A2.SetZero()
	z.A3.SetZero()
	z.A4.SetZero()
	z.A5.SetZero()
	return z
}

// Add sets z=x+y in E6D and returns z
func (z *E6D) Add(x, y *E6D) *E6D {
	z.A0.Add(&x.A0, &y.A0)
	z.A1.Add(&x.A1, &y.A1)
	z.A2.Add(&x.A2, &y.A2)
	z.A3.Add(&x.A3, &y.A3)
	z.A4.Add(&x.A4, &y.A4)
	z.A5.Add(&x.A5, &y.A5)
	return z
}

// Sub sets z to x-y and returns z
func (z *E6D) Sub(x, y *E6D) *E6D {
	z.A0.Sub(&x.A0, &y.A0)
	z.A1.Sub(&x.A1, &y.A1)
	z.A2.Sub(&x.A2, &y.A2)
	z.A3.Sub(&x.A3, &y.A3)
	z.A4.Sub(&x.A4, &y.A4)
	z.A5.Sub(&x.A5, &y.A5)
	return z
}

// Double sets z=2*x and returns z
func (z *E6D) Double(x *E6D) *E6D {
	z.A0.Double(&x.A0)
	z.A1.Double(&x.A1)
	z.A2.Double(&x.A2)
	z.A3.Double(&x.A3)
	z.A4.Double(&x.A4)
	z.A5.Double(&x.A5)
	return z
}

// SetRandom used only in tests
func (z *E6D) SetRandom() (*E6D, error) {
	if _, err := z.A0.SetRandom(); err != nil {
		return nil, err
	}
	if _, err := z.A1.SetRandom(); err != nil {
		return nil, err
	}
	if _, err := z.A2.SetRandom(); err != nil {
		return nil, err
	}
	if _, err := z.A3.SetRandom(); err != nil {
		return nil, err
	}
	if _, err := z.A4.SetRandom(); err != nil {
		return nil, err
	}
	if _, err := z.A5.SetRandom(); err != nil {
		return nil, err
	}
	return z, nil
}

// MustSetRandom sets z to a random value.
// It panics if reading from crypto/rand fails.
func (z *E6D) MustSetRandom() *E6D {
	if _, err := z.SetRandom(); err != nil {
		panic(err)
	}
	return z
}

// IsZero returns true if z is zero, false otherwise
func (z *E6D) IsZero() bool {
	return z.A0.IsZero() && z.A1.IsZero() && z.A2.IsZero() && z.A3.IsZero() && z.A4.IsZero() && z.A5.IsZero()
}

// IsOne returns true if z is one, false otherwise
func (z *E6D) IsOne() bool {
	return z.A0.IsOne() && z.A1.IsZero() && z.A2.IsZero() && z.A3.IsZero() && z.A4.IsZero() && z.A5.IsZero()
}

// Mul sets z=x*y in E6D and returns z
func (z *E6D) Mul(x, y *E6D) *E6D {
	return z.mulMontgomery6(x, y)
}

func (z *E6D) mulTower(x, y *E6D) *E6D {
	_x := ToTower(x)
	_y := ToTower(y)
	_x.Mul(_x, _y)
	_z := FromTower(_x)
	return z.Set(_z)
}

func (z *E6D) mulMontgomery6(a, b *E6D) *E6D {
	// Ref.: Peter L. Montgomery. Five, six, and seven-term Karatsuba-like formulae. IEEE
	// Transactions on Computers, 54(3):362–369, 2005.
	//
	// The product of two degree-5 polynomials a(X) and b(X):
	// a(X) = a0 + a1*X + a2*X^2 + a3*X^3 + a4*X^4 + a5*X^5
	// b(X) = b0 + b1*X + b2*X^2 + b3*X^3 + b4*X^4 + b5*X^5
	//
	// The result c(X) = a(X) * b(X) according to the ref. is:
	//
	// c(X) =
	//   	  (a0 + a1 + a2 + a3 + a4 + a5)(b0 + b1 + b2 + b3 + b4 + b5) * C
	// 		+ (a1 + a2 + a4 + a5)(b1 + b2 + b4 + b5) * (-C + X^6)
	// 		+ (a0 + a1 + a3 + a4)(b0 + b1 + b3 + b4) * (-C + X^4)
	// 		+ (a0 - a2 - a3 + a5)(b0 - b2 - b3 + b5) * (C - X^7 + X^6 - X^5 + X^4 - X^3)
	// 		+ (a0 - a2 - a5)(b0 - b2 - b5) * (C - X^5 + X^4 - X^3)
	// 		+ (a0 + a3 - a5)(b0 + b3 - b5) * (C - X^7 + X^6 - X^5)
	// 		+ (a0 + a1 + a2)(b0 + b1 + b2) * (C - X^7 + 6*X^6 - 2*X^5 + 2*X^4 - 2*X^3 + X^2)
	// 		+ (a3 + a4 + a5)(b3 + b4 + b5) * (C + X^8 - 2*X^7 + 2*X^6 - 2*X^5 + X^4 - X^3)
	// 		+ (a2 + a3)(b2 + b3) * (-2*C + X^7 - X^6 + 2*X^5 - X^4 + X^3)
	// 		+ (a1 - a4)(b1 - b4) * (-C + X^4 - X^5 + X^6)
	// 		+ (a1 + a2)(b1 + b2) * (-C + X^7 - 2*X^6 + 2*X^5 - 2*X^4 + 3*X^3 - X^2)
	// 		+ (a3 + a4)(b3 + b4) * (-C - X^8 + 3*X^7 - 2*X^6 + 2*X^5 - 2*X^4 + X^3)
	// 		+ (a0 + a1)(b0 + b1) * (-C + X^7 - X^6 + 2*X^5 - 3*X^4 + 2*X^3 - X^2 + X)
	// 		+ (a4 + a5)(b4 + b5) * (-C + X^9 - X^8 + 2*X^7 - 3*X^6 + 2*X^5 - X^4 + X^3)
	// 		+ a0*b0 * (3*C + 2*X^7 - 2*X^6 - 3*X^5 + 2*X^4 + 2*X^3 - X + 1)
	// 		+ a1*b1 * (3*C - X^7 - X^5 - X^4 - 3*X^3 + 2*X^2 - X)
	// 		+ a4*b4 * (3*C - X^9 + 2*X^8 - 3*X^7 + X^6 - X^5 - X^3)
	// 		+ a5*b5 * (-3*C + X^10 - X^9 + 2*X^7 - 2*X^6 + 3*X^5 - 2*X^4 + 2*X^3)
	//
	// We fix the parameter C to X^6 so that the second term disappears. We then compute the interpolation points
	// vi = a(Xi)*b(Xi) at Xi={0, ±1, ±2, ±3, ±4, 5,∞}:
	//
	//		v0 = (a0 + a1 + a2 + a3 + a4 + a5)(b0 + b1 + b2 + b3 + b4 + b5)
	//		v2 = (a0 + a1 + a3 + a4)(b0 + b1 + b3 + b4)
	//		v3 = (a0 − a2 − a3 + a5)(b0 − b2 − b3 + b5)
	//		v4 = (a0 − a2 − a5)(b0 − b2 − b5)
	//		v5 = (a0 + a3 − a5)(b0 + b3 − b5)
	//		v6 = (a0 + a1 + a2)(b0 + b1 + b2)
	//		v7 = (a3 + a4 + a5)(b3 + b4 + b5)
	//		v8 = (a2 + a3)(b2 + b3)
	//		v9 = (a1 − a4)(b1 − b4)
	//		v10 = (a1 + a2)(b1 + b2)
	//		v11 = (a3 + a4)(b3 + b4)
	//		v12 = (a0 + a1)(b0 + b1)
	//		v13 = (a4 + a5)(b4 + b5)
	//		v14 = a0b0
	//		v15 = a1b1
	//		v16 = a4b4
	//		v17 = a5b5
	//
	// 		We do this optimally in 17 multiplications and 30 additions/subtractions in Fp.

	// v[0..17] stores the interpolation points v0..v17
	// t[0..13] stores the intermediate sums for a and b, and later helper vars
	var v [18]fp.Element
	var t [14]fp.Element

	// -------------------------------------------------------------------------
	// Phase 1: Evaluation of A (Store results in v)
	// -------------------------------------------------------------------------

	// Level 1: Simple Sums (store directly in v to serve as 't' for A)
	v[12].Add(&a.A0, &a.A1) // v12 factor
	v[13].Add(&a.A4, &a.A5) // v13 factor
	v[10].Add(&a.A1, &a.A2) // v10 factor
	v[11].Add(&a.A3, &a.A4) // v11 factor
	v[8].Add(&a.A2, &a.A3)  // v8 factor
	v[9].Sub(&a.A1, &a.A4)  // v9 factor

	// Level 2: Derived Sums (reuse v entries)
	v[6].Add(&v[12], &a.A2)  // v6 factor
	v[7].Add(&v[11], &a.A5)  // v7 factor
	v[2].Add(&v[12], &v[11]) // v2 factor
	v[0].Add(&v[6], &v[7])   // v0 factor

	// Level 3: Difference Block (reuse t[0] as temp)
	t[0].Sub(&a.A0, &a.A5)
	v[4].Sub(&t[0], &a.A2) // v4 factor
	v[5].Add(&v[4], &v[8]) // v5 factor
	t[0].Add(&a.A0, &a.A5)
	v[3].Sub(&t[0], &v[8]) // v3 factor

	// -------------------------------------------------------------------------
	// Phase 2: Evaluation of B (Store results in t)
	// -------------------------------------------------------------------------
	// We mirror the exact logic of A, but store in t[0..13]

	// Level 1
	t[12].Add(&b.A0, &b.A1)
	t[13].Add(&b.A4, &b.A5)
	t[10].Add(&b.A1, &b.A2)
	t[11].Add(&b.A3, &b.A4)
	t[8].Add(&b.A2, &b.A3)
	t[9].Sub(&b.A1, &b.A4)

	// Level 2
	t[6].Add(&t[12], &b.A2)
	t[7].Add(&t[11], &b.A5)
	t[2].Add(&t[12], &t[11])
	t[0].Add(&t[6], &t[7])

	// Level 3
	t[1].Sub(&b.A0, &b.A5) // reuse t[1] temporarily
	t[4].Sub(&t[1], &b.A2)
	t[5].Add(&t[4], &t[8])
	t[1].Add(&b.A0, &b.A5)
	t[3].Sub(&t[1], &t[8])

	// -------------------------------------------------------------------------
	// Phase 3: Pointwise Multiplication
	// -------------------------------------------------------------------------
	// v[i] = A_factor[i] * B_factor[i]
	v[0].Mul(&v[0], &t[0])
	v[2].Mul(&v[2], &t[2])
	v[3].Mul(&v[3], &t[3])
	v[4].Mul(&v[4], &t[4])
	v[5].Mul(&v[5], &t[5])
	v[6].Mul(&v[6], &t[6])
	v[7].Mul(&v[7], &t[7])
	v[8].Mul(&v[8], &t[8])
	v[9].Mul(&v[9], &t[9])
	v[10].Mul(&v[10], &t[10])
	v[11].Mul(&v[11], &t[11])
	v[12].Mul(&v[12], &t[12])
	v[13].Mul(&v[13], &t[13])

	// Direct multiplications (v14..v17)
	v[14].Mul(&a.A0, &b.A0)
	v[15].Mul(&a.A1, &b.A1)
	v[16].Mul(&a.A4, &b.A4)
	v[17].Mul(&a.A5, &b.A5)

	// We then we re-arrange the terms in function of the degree of X and use
	// the fact that X^6=-β, because we construct 𝔽p⁶[v] as 𝔽p/v⁶-β. The
	// resulting coefficients c0,c1,c3,c4 and c5 are:
	//
	// 	c0 = v14 + β(v0 − v2 + v4 + 2(v3+v5+v6-v12) + 3(v7+v15-v8-v10-v11) +
	// 	4(v16-v13) − 5(v14+v17))
	//
	//  c1 = v12 − (v14 + v15) + β(v8 + v10 + v12 − (v3 + v5 + v6 + v15) +
	//  2(v14 + v17 + v13 - v7) + 3(v11 - v16))
	//
	// 	c2 = 2v15 + v6 − (v10 + v12) + β(2v16 + v7 − (v11 + v13))
	//
	// 	c3 = v8 + v11 + v13 − (v3 + v4 + v7 + v16) + 3(v10 - v15) + 2(v12 + v14
	// 	+ v17 - v6) + β(v13 − (v16 + v17))
	//
	// 	c4 = v2 + v3 + v4 + v7 + v15 + v9 − (v8 + v13) − 3v12 + 2(v6 − (v17 +
	// 	v10 + v11 + v14)) + βv17
	//
	//  c5 = −(v3 + v4 + v5 + v9 + v15 + v16) + 2(v8 + v10 + v11 + v12 + v13 −
	//  (v6 + v7)) + 3(v14 + v17)
	//
	// Now given than β=-4, we compute the coefficients in 100
	// additions/subtractions/doublings.

	// -------------------------------------------------------------------------
	// Phase 4: Reconstruction / Interpolation
	// -------------------------------------------------------------------------
	// Mapping helpers to t array:
	// t[0] = s14_17 (v14 + v17)
	// t[1] = s10_11 (v10 + v11)
	// t[2] = s8_10_11 (v8 + v10 + v11)
	// t[3] = s3_5_6 (v3 + v5 + v6)
	// t[4] = s11_13 (v11 + v13)
	// t[5], t[6] = scratch

	t[0].Add(&v[14], &v[17])
	t[1].Add(&v[10], &v[11])
	t[2].Add(&v[8], &t[1])
	t[3].Add(&v[3], &v[5])
	t[3].Add(&t[3], &v[6])
	t[4].Add(&v[11], &v[13])

	// --- Compute c0 ---
	// Inner: (v0 - v2 + v4) + 2(t3 - v12) + 3(v7 + v15 - t2) + 4(v16 - v13) - 5(t0)
	// Use z.A0 as accumulator
	z.A0.Sub(&t[3], &v[12])
	z.A0.Add(&z.A0, &z.A0) // 2*...

	t[5].Add(&v[7], &v[15])
	t[5].Sub(&t[5], &t[2])
	t[6].Add(&t[5], &t[5]) // reuse t[6] as scratch
	t[5].Add(&t[5], &t[6]) // 3*...
	z.A0.Add(&z.A0, &t[5])

	t[5].Sub(&v[16], &v[13])
	t[5].Add(&t[5], &t[5])
	t[5].Add(&t[5], &t[5]) // 4*...
	z.A0.Add(&z.A0, &t[5])

	t[5].Add(&t[0], &t[0])
	t[5].Add(&t[5], &t[5])
	t[5].Add(&t[5], &t[0]) // 5*...
	z.A0.Sub(&z.A0, &t[5])

	t[5].Sub(&v[0], &v[2])
	t[5].Add(&t[5], &v[4])
	z.A0.Add(&z.A0, &t[5])

	// Apply beta=-4: c0 = v14 - 4 * Inner
	z.A0.Add(&z.A0, &z.A0)
	z.A0.Add(&z.A0, &z.A0)
	z.A0.Sub(&v[14], &z.A0)

	// --- Compute c1 ---
	// Inner: (v8 + v10 + v12) - (t3 + v15) + 2(t0 + v13 - v7) + 3(v11 - v16)
	// Note: t[2] is (v8 + v10 + v11). We must subtract v11 to get (v8 + v10).
	z.A1.Sub(&t[2], &v[11]) // z.A1 = v8 + v10
	z.A1.Add(&z.A1, &v[12]) // z.A1 = v8 + v10 + v12

	t[5].Add(&t[3], &v[15])
	z.A1.Sub(&z.A1, &t[5])

	t[5].Add(&t[0], &v[13])
	t[5].Sub(&t[5], &v[7])
	t[5].Add(&t[5], &t[5]) // 2*...
	z.A1.Add(&z.A1, &t[5])

	t[5].Sub(&v[11], &v[16])
	t[6].Add(&t[5], &t[5])
	t[5].Add(&t[5], &t[6]) // 3*...
	z.A1.Add(&z.A1, &t[5])

	// c1 = (v12 - v14 - v15) - 4 * Inner
	z.A1.Add(&z.A1, &z.A1)
	z.A1.Add(&z.A1, &z.A1)
	t[5].Sub(&v[12], &v[14])
	t[5].Sub(&t[5], &v[15])
	z.A1.Sub(&t[5], &z.A1)

	// --- Compute c2 ---
	// Inner (for beta): 2v16 + v7 - t4
	t[5].Add(&v[16], &v[16])
	t[5].Add(&t[5], &v[7])
	t[5].Sub(&t[5], &t[4])
	t[5].Add(&t[5], &t[5])
	t[5].Add(&t[5], &t[5]) // 4 * Inner

	z.A2.Add(&v[15], &v[15])
	z.A2.Add(&z.A2, &v[6])
	z.A2.Sub(&z.A2, &v[10])
	z.A2.Sub(&z.A2, &v[12])
	z.A2.Sub(&z.A2, &t[5])

	// --- Compute c3 ---
	// Inner (for beta): v13 - (v16 + v17)
	t[5].Add(&v[16], &v[17])
	t[5].Sub(&v[13], &t[5])
	t[5].Add(&t[5], &t[5])
	t[5].Add(&t[5], &t[5]) // 4 * Inner

	// Parts
	t[6].Sub(&v[10], &v[15]) // 3(v10 - v15)
	z.A3.Add(&t[6], &t[6])
	z.A3.Add(&z.A3, &t[6])

	t[6].Add(&v[12], &t[0]) // 2(v12 + t0 - v6)
	t[6].Sub(&t[6], &v[6])
	t[6].Add(&t[6], &t[6])
	z.A3.Add(&z.A3, &t[6])

	z.A3.Add(&z.A3, &v[8])
	z.A3.Add(&z.A3, &t[4]) // v8 + s11_13
	z.A3.Sub(&z.A3, &v[3])
	z.A3.Sub(&z.A3, &v[4])
	z.A3.Sub(&z.A3, &v[7])
	z.A3.Sub(&z.A3, &v[16])
	z.A3.Sub(&z.A3, &t[5]) // - beta term

	// --- Compute c5 ---
	// Calculate shared group t[6] = v3+v4+v9+v15
	t[6].Add(&v[3], &v[4])
	t[6].Add(&t[6], &v[9])
	t[6].Add(&t[6], &v[15])

	// NegGroup: t[6] + v5 + v16
	t[5].Add(&t[6], &v[5])
	t[5].Add(&t[5], &v[16])

	// 2 * Group: 2(t2 + v12 + v13 - v6 - v7)
	// Note: t2 is (v8 + v10 + v11), which is correct for c5.
	z.A5.Add(&t[2], &v[12])
	z.A5.Add(&z.A5, &v[13])
	z.A5.Sub(&z.A5, &v[6])
	z.A5.Sub(&z.A5, &v[7])
	z.A5.Add(&z.A5, &z.A5)

	// 3 * t0
	z.A4.Add(&t[0], &t[0]) // reuse z.A4 as scratch
	z.A4.Add(&z.A4, &t[0])

	z.A5.Sub(&z.A5, &t[5])
	z.A5.Add(&z.A5, &z.A4)

	// --- Compute c4 ---
	// c4 = (v2 + v7 + t6) - (v8 + v13) - 3v12 + 2(v6 - (t0 + t1)) - 4v17
	z.A4.Add(&v[2], &v[7])
	z.A4.Add(&z.A4, &t[6]) // Add shared group
	z.A4.Sub(&z.A4, &v[8])
	z.A4.Sub(&z.A4, &v[13])

	t[5].Add(&v[12], &v[12])
	t[5].Add(&t[5], &v[12])
	z.A4.Sub(&z.A4, &t[5]) // -3v12

	t[5].Add(&t[0], &t[1])
	t[5].Sub(&v[6], &t[5])
	t[5].Add(&t[5], &t[5])
	z.A4.Add(&z.A4, &t[5]) // +2(...)

	t[5].Add(&v[17], &v[17])
	t[5].Add(&t[5], &t[5])
	z.A4.Sub(&z.A4, &t[5]) // -4v17

	return z
}

// MulBy023 multiplies z by an E6 sparse element 023
func (z *E6D) MulBy023(c0, c1 *fp.Element) *E6D {
	var a, b, tmp, a0, a1, a2, b0, b1, b2, d, zC10, zC11, zC12, one, t0, t1, t2 fp.Element
	a.Mul(&z.A0, c0)
	b.Mul(&z.A2, c1)
	tmp.Add(&z.A2, &z.A4)
	a0.Mul(c1, &tmp)
	a0.Sub(&b, &a0)
	a0.Double(&a0).Double(&a0)
	a0.Add(&a0, &a)
	a2.Mul(&z.A4, c0)
	a2.Add(&a2, &b)
	a1.Add(c0, c1)
	tmp.Add(&z.A0, &z.A2)
	a1.Mul(&a1, &tmp)
	a1.Sub(&a1, &a)
	a1.Sub(&a1, &b)

	b0.Double(&z.A5).Double(&b0)
	b2.Neg(&z.A3)
	b1.Neg(&z.A1)

	one.SetOne()
	d.Add(c1, &one)

	zC10.Add(&z.A1, &z.A0)
	zC11.Add(&z.A3, &z.A2)
	zC12.Add(&z.A5, &z.A4)

	a.Mul(&zC10, c0)
	b.Mul(&zC11, &d)
	tmp.Add(&zC11, &zC12)
	t0.Mul(&d, &tmp)
	t0.Sub(&b, &t0)
	t0.Double(&t0).Double(&t0)
	t0.Add(&t0, &a)
	t2.Mul(&zC12, c0)
	t2.Add(&t2, &b)
	t1.Add(c0, &d)
	tmp.Add(&zC10, &zC11)
	t1.Mul(&t1, &tmp)
	t1.Sub(&t1, &a)
	t1.Sub(&t1, &b)

	zC10.Sub(&t0, &a0)
	zC11.Sub(&t1, &a1)
	zC12.Sub(&t2, &a2)

	zC10.Add(&zC10, &b0)
	zC11.Add(&zC11, &b1)
	zC12.Add(&zC12, &b2)

	var zC00, zC01, zC02 fp.Element
	zC00.Double(&b2).Double(&zC00)
	zC00.Add(&zC00, &a0)
	zC01.Sub(&a1, &b0)
	zC02.Sub(&a2, &b1)

	z.A0.Set(&zC00)
	z.A1.Set(&zC10)
	z.A2.Set(&zC01)
	z.A3.Set(&zC11)
	z.A4.Set(&zC02)
	z.A5.Set(&zC12)

	return z
}

// Square sets z=x*x in E6D and returns z
func (z *E6D) Square(x *E6D) *E6D {
	return z.squareTower(x)
}
func (z *E6D) squareTower(x *E6D) *E6D {
	_x := ToTower(x)
	_x.Square(_x)
	_z := FromTower(_x)
	return z.Set(_z)
}

func (z *E6D) squareMontgomery6(a *E6D) *E6D {
	// We proceed similarly to the multiplication.
	//
	// The square of a degree-5 polynomials a(X):
	// a(X) = a0 + a1*X + a2*X^2 + a3*X^3 + a4*X^4 + a5*X^5
	//
	// Interpolation points Xi={0, ±1, ±2, ±3, ±4, 5,∞}:
	//		v0 = (a0 + a1 + a2 + a3 + a4 + a5)^2
	//		v2 = (a0 + a1 + a3 + a4)^2
	//		v3 = (a0 − a2 − a3 + a5)^2
	//		v4 = (a0 − a2 − a5)^2
	//		v5 = (a0 + a3 − a5)^2
	//		v6 = (a0 + a1 + a2)^2
	//		v7 = (a3 + a4 + a5)^2
	//		v8 = (a2 + a3)^2
	//		v9 = (a1 − a4)^2
	//		v10 = (a1 + a2)^2
	//		v11 = (a3 + a4)^2
	//		v12 = (a0 + a1)^2
	//		v13 = (a4 + a5)^2
	//		v14 = a0^2
	//		v15 = a1^2
	//		v16 = a4^2
	//		v17 = a5^2
	//
	// 		We do this optimally in 17 squares and 15 additions/subtractions in Fp.

	// v[0..17] stores the interpolation points v0..v17
	// t[0..6] stores the intermediate sums for a and b, and later helper vars
	var v [18]fp.Element
	var t [7]fp.Element

	// -------------------------------------------------------------------------
	// Phase 1: Evaluation of A (Store results in v)
	// -------------------------------------------------------------------------

	// Level 1: Simple Sums
	v[12].Add(&a.A0, &a.A1) // v12 factor
	v[13].Add(&a.A4, &a.A5) // v13 factor
	v[10].Add(&a.A1, &a.A2) // v10 factor
	v[11].Add(&a.A3, &a.A4) // v11 factor
	v[8].Add(&a.A2, &a.A3)  // v8 factor
	v[9].Sub(&a.A1, &a.A4)  // v9 factor

	// Level 2: Derived Sums (reuse v entries)
	v[6].Add(&v[12], &a.A2)  // v6 factor
	v[7].Add(&v[11], &a.A5)  // v7 factor
	v[2].Add(&v[12], &v[11]) // v2 factor
	v[0].Add(&v[6], &v[7])   // v0 factor

	// Level 3: Difference Block (reuse t[0] as temp)
	t[0].Sub(&a.A0, &a.A5)
	v[4].Sub(&t[0], &a.A2) // v4 factor
	v[5].Add(&v[4], &v[8]) // v5 factor
	t[0].Add(&a.A0, &a.A5)
	v[3].Sub(&t[0], &v[8]) // v3 factor

	// -------------------------------------------------------------------------
	// Phase 3: Squares
	// -------------------------------------------------------------------------
	// v[i] = A_factor[i]^2
	v[0].Square(&v[0])
	v[2].Square(&v[2])
	v[3].Square(&v[3])
	v[4].Square(&v[4])
	v[5].Square(&v[5])
	v[6].Square(&v[6])
	v[7].Square(&v[7])
	v[8].Square(&v[8])
	v[9].Square(&v[9])
	v[10].Square(&v[10])
	v[11].Square(&v[11])
	v[12].Square(&v[12])
	v[13].Square(&v[13])

	// Direct squares
	v[14].Square(&a.A0)
	v[15].Square(&a.A1)
	v[16].Square(&a.A4)
	v[17].Square(&a.A5)

	// We then we re-arrange the terms in function of the degree of X and use
	// the fact that X^6=-β, because we construct 𝔽p⁶[v] as 𝔽p/v⁶-β. The
	// resulting coefficients c0,c1,c3,c4 and c5 are:
	//
	// 	c0 = v14 + β(v0 − v2 + v4 + 2(v3+v5+v6-v12) + 3(v7+v15-v8-v10-v11) +
	// 	4(v16-v13) − 5(v14+v17))
	//
	//  c1 = v12 − (v14 + v15) + β(v8 + v10 + v12 − (v3 + v5 + v6 + v15) +
	//  2(v14 + v17 + v13 - v7) + 3(v11 - v16))
	//
	// 	c2 = 2v15 + v6 − (v10 + v12) + β(2v16 + v7 − (v11 + v13))
	//
	// 	c3 = v8 + v11 + v13 − (v3 + v4 + v7 + v16) + 3(v10 - v15) + 2(v12 + v14
	// 	+ v17 - v6) + β(v13 − (v16 + v17))
	//
	// 	c4 = v2 + v3 + v4 + v7 + v15 + v9 − (v8 + v13) − 3v12 + 2(v6 − (v17 +
	// 	v10 + v11 + v14)) + βv17
	//
	//  c5 = −(v3 + v4 + v5 + v9 + v15 + v16) + 2(v8 + v10 + v11 + v12 + v13 −
	//  (v6 + v7)) + 3(v14 + v17)
	//
	// Now given than β=-4, we compute the coefficients in 100
	// additions/subtractions/doublings.

	// -------------------------------------------------------------------------
	// Phase 4: Reconstruction / Interpolation
	// -------------------------------------------------------------------------
	// Mapping helpers to t array:
	// t[0] = s14_17 (v14 + v17)
	// t[1] = s10_11 (v10 + v11)
	// t[2] = s8_10_11 (v8 + v10 + v11)
	// t[3] = s3_5_6 (v3 + v5 + v6)
	// t[4] = s11_13 (v11 + v13)
	// t[5], t[6] = scratch

	t[0].Add(&v[14], &v[17])
	t[1].Add(&v[10], &v[11])
	t[2].Add(&v[8], &t[1])
	t[3].Add(&v[3], &v[5])
	t[3].Add(&t[3], &v[6])
	t[4].Add(&v[11], &v[13])

	// --- Compute c0 ---
	// Inner: (v0 - v2 + v4) + 2(t3 - v12) + 3(v7 + v15 - t2) + 4(v16 - v13) - 5(t0)
	// Use z.A0 as accumulator
	z.A0.Sub(&t[3], &v[12])
	z.A0.Add(&z.A0, &z.A0) // 2*...

	t[5].Add(&v[7], &v[15])
	t[5].Sub(&t[5], &t[2])
	t[6].Add(&t[5], &t[5]) // reuse t[6] as scratch
	t[5].Add(&t[5], &t[6]) // 3*...
	z.A0.Add(&z.A0, &t[5])

	t[5].Sub(&v[16], &v[13])
	t[5].Add(&t[5], &t[5])
	t[5].Add(&t[5], &t[5]) // 4*...
	z.A0.Add(&z.A0, &t[5])

	t[5].Add(&t[0], &t[0])
	t[5].Add(&t[5], &t[5])
	t[5].Add(&t[5], &t[0]) // 5*...
	z.A0.Sub(&z.A0, &t[5])

	t[5].Sub(&v[0], &v[2])
	t[5].Add(&t[5], &v[4])
	z.A0.Add(&z.A0, &t[5])

	// Apply beta=-4: c0 = v14 - 4 * Inner
	z.A0.Add(&z.A0, &z.A0)
	z.A0.Add(&z.A0, &z.A0)
	z.A0.Sub(&v[14], &z.A0)

	// --- Compute c1 ---
	// Inner: (v8 + v10 + v12) - (t3 + v15) + 2(t0 + v13 - v7) + 3(v11 - v16)
	// Note: t[2] is (v8 + v10 + v11). We must subtract v11 to get (v8 + v10).
	z.A1.Sub(&t[2], &v[11]) // z.A1 = v8 + v10
	z.A1.Add(&z.A1, &v[12]) // z.A1 = v8 + v10 + v12

	t[5].Add(&t[3], &v[15])
	z.A1.Sub(&z.A1, &t[5])

	t[5].Add(&t[0], &v[13])
	t[5].Sub(&t[5], &v[7])
	t[5].Add(&t[5], &t[5]) // 2*...
	z.A1.Add(&z.A1, &t[5])

	t[5].Sub(&v[11], &v[16])
	t[6].Add(&t[5], &t[5])
	t[5].Add(&t[5], &t[6]) // 3*...
	z.A1.Add(&z.A1, &t[5])

	// c1 = (v12 - v14 - v15) - 4 * Inner
	z.A1.Add(&z.A1, &z.A1)
	z.A1.Add(&z.A1, &z.A1)
	t[5].Sub(&v[12], &v[14])
	t[5].Sub(&t[5], &v[15])
	z.A1.Sub(&t[5], &z.A1)

	// --- Compute c2 ---
	// Inner (for beta): 2v16 + v7 - t4
	t[5].Add(&v[16], &v[16])
	t[5].Add(&t[5], &v[7])
	t[5].Sub(&t[5], &t[4])
	t[5].Add(&t[5], &t[5])
	t[5].Add(&t[5], &t[5]) // 4 * Inner

	z.A2.Add(&v[15], &v[15])
	z.A2.Add(&z.A2, &v[6])
	z.A2.Sub(&z.A2, &v[10])
	z.A2.Sub(&z.A2, &v[12])
	z.A2.Sub(&z.A2, &t[5])

	// --- Compute c3 ---
	// Inner (for beta): v13 - (v16 + v17)
	t[5].Add(&v[16], &v[17])
	t[5].Sub(&v[13], &t[5])
	t[5].Add(&t[5], &t[5])
	t[5].Add(&t[5], &t[5]) // 4 * Inner

	// Parts
	t[6].Sub(&v[10], &v[15]) // 3(v10 - v15)
	z.A3.Add(&t[6], &t[6])
	z.A3.Add(&z.A3, &t[6])

	t[6].Add(&v[12], &t[0]) // 2(v12 + t0 - v6)
	t[6].Sub(&t[6], &v[6])
	t[6].Add(&t[6], &t[6])
	z.A3.Add(&z.A3, &t[6])

	z.A3.Add(&z.A3, &v[8])
	z.A3.Add(&z.A3, &t[4]) // v8 + s11_13
	z.A3.Sub(&z.A3, &v[3])
	z.A3.Sub(&z.A3, &v[4])
	z.A3.Sub(&z.A3, &v[7])
	z.A3.Sub(&z.A3, &v[16])
	z.A3.Sub(&z.A3, &t[5]) // - beta term

	// --- Compute c5 ---
	// Calculate shared group t[6] = v3+v4+v9+v15
	t[6].Add(&v[3], &v[4])
	t[6].Add(&t[6], &v[9])
	t[6].Add(&t[6], &v[15])

	// NegGroup: t[6] + v5 + v16
	t[5].Add(&t[6], &v[5])
	t[5].Add(&t[5], &v[16])

	// 2 * Group: 2(t2 + v12 + v13 - v6 - v7)
	// Note: t2 is (v8 + v10 + v11), which is correct for c5.
	z.A5.Add(&t[2], &v[12])
	z.A5.Add(&z.A5, &v[13])
	z.A5.Sub(&z.A5, &v[6])
	z.A5.Sub(&z.A5, &v[7])
	z.A5.Add(&z.A5, &z.A5)

	// 3 * t0
	z.A4.Add(&t[0], &t[0]) // reuse z.A4 as scratch
	z.A4.Add(&z.A4, &t[0])

	z.A5.Sub(&z.A5, &t[5])
	z.A5.Add(&z.A5, &z.A4)

	// --- Compute c4 ---
	// c4 = (v2 + v7 + t6) - (v8 + v13) - 3v12 + 2(v6 - (t0 + t1)) - 4v17
	z.A4.Add(&v[2], &v[7])
	z.A4.Add(&z.A4, &t[6]) // Add shared group
	z.A4.Sub(&z.A4, &v[8])
	z.A4.Sub(&z.A4, &v[13])

	t[5].Add(&v[12], &v[12])
	t[5].Add(&t[5], &v[12])
	z.A4.Sub(&z.A4, &t[5]) // -3v12

	t[5].Add(&t[0], &t[1])
	t[5].Sub(&v[6], &t[5])
	t[5].Add(&t[5], &t[5])
	z.A4.Add(&z.A4, &t[5]) // +2(...)

	t[5].Add(&v[17], &v[17])
	t[5].Add(&t[5], &t[5])
	z.A4.Sub(&z.A4, &t[5]) // -4v17

	return z
}

// Inverse sets z to the inverse of x in E6D and returns z
//
// if x == 0, sets and returns z = x
func (z *E6D) Inverse(x *E6D) *E6D {
	_x := ToTower(x)
	_x.Inverse(_x)
	_z := FromTower(_x)
	return z.Set(_z)
}

// InverseUnitary inverses a unitary element
func (z *E6D) InverseUnitary(x *E6D) *E6D {
	return z.Conjugate(x)
}

// Conjugate sets z to x conjugated and returns z
func (z *E6D) Conjugate(x *E6D) *E6D {
	z.A0.Set(&x.A0)
	z.A1.Neg(&x.A1)
	z.A2.Set(&x.A2)
	z.A3.Neg(&x.A3)
	z.A4.Set(&x.A4)
	z.A5.Neg(&x.A5)
	return z
}

// FromTower
func FromTower(x *E6) *E6D {
	// gnark-crypto uses a quadratic over cubic sextic extension of Fp.
	// The two towers are isomorphic and the coefficients are permuted as follows:
	// 		a00 a01 a02 a10 a11 a12
	// 		A0  A2  A4  A1  A3  A5
	var z E6D
	z.A0.Set(&x.B0.A0)
	z.A1.Set(&x.B1.A0)
	z.A2.Set(&x.B0.A1)
	z.A3.Set(&x.B1.A1)
	z.A4.Set(&x.B0.A2)
	z.A5.Set(&x.B1.A2)
	return &z
}

// ToTower
func ToTower(x *E6D) *E6 {
	// gnark-crypto uses a quadratic over cubic sextic extension of Fp.
	// The two towers are isomorphic and the coefficients are permuted as follows:
	// 		a00 a01 a02 a10 a11 a12
	// 		A0  A2  A4  A1  A3  A5
	var z E6
	z.B0.A0.Set(&x.A0)
	z.B1.A0.Set(&x.A1)
	z.B0.A1.Set(&x.A2)
	z.B1.A1.Set(&x.A3)
	z.B0.A2.Set(&x.A4)
	z.B1.A2.Set(&x.A5)
	return &z
}
