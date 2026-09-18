// Copyright 2020-2026 Consensys Software Inc.
// Licensed under the Apache License, Version 2.0. See the LICENSE file for details.

package mamabear

import "math/big"

// Sqrt z = √x (mod p)
//
// Uses Tonelli-Shanks with p − 1 = 2^34 · Q, Q = 32767 (odd).
// If x is not a quadratic residue, Sqrt returns nil and leaves z unchanged.
func (z *Element) Sqrt(x *Element) *Element {
	// p ≡ 1 (mod 4), 2-adicity = 34, odd part Q = 32767 = 2^15 − 1.
	// Tonelli-Shanks:
	//   w = x^{(Q-1)/2}  → w = x^16383
	//   y = x · w        → y = x^{(Q+1)/2}
	//   b = w · y        → b = x^Q
	//   g = nonResidue^Q (precomputed, order 2^34)
	//   r = 34

	var y, b, t, w Element

	// w = x^16383 = x^{(Q-1)/2}
	w.expByLegendreExp(*x)

	y.Mul(x, &w)  // y = x^{(Q+1)/2}
	b.Mul(&w, &y) // b = x^Q

	// g = 3^Q mod p in Montgomery form (nonResidue^Q, order 2^34)
	var g = Element{393730615033094} // precomputed: 3^32767 mod p, in Montgomery form
	r := uint64(34)

	// Legendre check: t = b^{2^{r-1}} should be 1 for x to be a QR.
	t = b
	for i := uint64(0); i < r-1; i++ {
		t.Square(&t)
	}
	if t.IsZero() {
		return z.SetZero()
	}
	if !t.IsOne() {
		return nil
	}

	for {
		var m uint64
		t = b
		for !t.IsOne() {
			t.Square(&t)
			m++
		}
		if m == 0 {
			return z.Set(&y)
		}
		// t = g^{2^{r-m-1}}
		ge := int(r - m - 1)
		t = g
		for ge > 0 {
			t.Square(&t)
			ge--
		}
		g.Square(&t)
		y.Mul(&y, &t)
		b.Mul(&b, &g)
		r = m
	}
}

// expByLegendreExp computes z = x^16383 mod p  (= x^{(Q-1)/2} where Q = 32767).
//
// 16383 = 2^14 − 1.  Add chain: compute x^{2^k − 1} for k = 1, 2, 3, 4, 7, 14.
func (z *Element) expByLegendreExp(x Element) *Element {
	var a1, a2, a3, a4, a7, a14 Element
	a1.Set(&x)

	// a2 = x^3 = x^{2^2 - 1}: a1^2 * a1
	a2.Square(&a1)
	a2.Mul(&a2, &a1)

	// a3 = x^7 = x^{2^3 - 1}: a2^2 * a1
	a3.Square(&a2)
	a3.Mul(&a3, &a1)

	// a4 = x^15 = x^{2^4 - 1}: a2^{2^2} * a2
	a4.Square(&a2)
	a4.Square(&a4)
	a4.Mul(&a4, &a2)

	// a7 = x^127 = x^{2^7 - 1}: a4^{2^3} * a3
	a7 = a4
	for range 3 {
		a7.Square(&a7)
	}
	a7.Mul(&a7, &a3)

	// a14 = x^16383 = x^{2^14 - 1}: a7^{2^7} * a7
	a14 = a7
	for range 7 {
		a14.Square(&a14)
	}
	a14.Mul(&a14, &a7)

	z.Set(&a14)
	return z
}

// Cbrt z = ∛x (mod p)
//
// p − 1 = 2^34 · 32767.  32767 = 3 · 10922 + 1, so 3 | (p−1) iff 3 | 32767.
// 32767 / 3 = 10922.333..., so 3 ∤ 32767.  Hence gcd(3, p−1) = gcd(3, 2^34 · 32767).
// Since 32767 is odd and 3 ∤ 32767 and 3 ∤ 2, we have gcd(3, p−1) = 1... wait:
// p ≡ 1 (mod 3) iff 3 | p−1.  p−1 = 2^34 · 32767.  32767 mod 3 = ?
// 32767 = 10922 * 3 + 1, so 32767 ≡ 1 (mod 3).  2^34 ≡ 1 (mod 3).
// So p−1 ≡ 1 * 1 = 1 (mod 3), meaning p ≡ 2 (mod 3).
// For p ≡ 2 (mod 3): every element has a unique cube root; cbrt(x) = x^{(2p-1)/3}.
func (z *Element) Cbrt(x *Element) *Element {
	// (2p - 1) / 3: since p ≡ 2 (mod 3), 2p ≡ 1 (mod 3), so (2p-1) ≡ 0 (mod 3). ✓
	// exp = (2*562932773552129 - 1) / 3 = 1125865547104257 / 3 = 375288515701419
	exp := big.NewInt(0)
	exp.SetString("375288515701419", 10)

	var y Element
	y.Exp(*x, exp)

	// Verify y³ = x
	var check Element
	check.Cube(&y)
	if !check.Equal(x) {
		// p ≡ 2 (mod 3) guarantees a unique cube root, so this should not happen.
		return nil
	}
	return z.Set(&y)
}
