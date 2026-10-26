// Copyright 2020-2026 Consensys Incorporated
// Licensed under the Apache License, Version 2.0. See the LICENSE file for details.

package matrix

import "math/big"

// Ring is the arithmetic a field element must provide for the Poseidon2
// internal-matrix check below. Every generated fr.Element satisfies it.
type Ring[E any] interface {
	*E
	SetOne() *E
	SetUint64(uint64) *E
	Set(*E) *E
	Add(*E, *E) *E
	Sub(*E, *E) *E
	Mul(*E, *E) *E
	Inverse(*E) *E
	IsZero() bool
	Equal(*E) bool
}

// HasNoInvariantSubspace reports whether the matrix m satisfies the condition
// Poseidon2 imposes on its internal matrix M_I: for every power M_I^i with
// i = 1..2t, the characteristic polynomial is irreducible over the field.
//
// An irreducible characteristic polynomial forces the minimal polynomial to
// equal it, so M_I^i is non-derogatory and fixes no proper subspace. That is
// what rules out the invariant-subspace trails of Section 5.2 of the Poseidon2
// paper (https://eprint.iacr.org/2023/323.pdf), and it is the condition
// check_minpoly_condition enforces in the reference parameter scripts
// (https://github.com/HorizenLabs/poseidon2).
//
// The condition depends on the field: a diagonal that satisfies it over one
// prime says nothing about another, so it must be rechecked whenever a
// diagonal is reused for a new modulus.
//
// modulus is the field characteristic. The check is O(t^5) field operations and
// is meant for tests, not for use at runtime.
func HasNoInvariantSubspace[E any, PE Ring[E]](m [][]E, modulus *big.Int) bool {
	t := len(m)
	cur := cloneMat[E, PE](m)
	for i := 1; i <= 2*t; i++ {
		cp := charPoly[E, PE](cur)
		if len(cp)-1 != t || !isIrreducible[E, PE](cp, modulus) {
			return false
		}
		cur = matMul[E, PE](m, cur)
	}
	return true
}

func newMat[E any, PE Ring[E]](t int) [][]E {
	m := make([][]E, t)
	for i := range m {
		m[i] = make([]E, t)
	}
	return m
}

func cloneMat[E any, PE Ring[E]](a [][]E) [][]E {
	m := newMat[E, PE](len(a))
	for i := range a {
		copy(m[i], a[i])
	}
	return m
}

func matMul[E any, PE Ring[E]](a, b [][]E) [][]E {
	t := len(a)
	c := newMat[E, PE](t)
	var tmp E
	for i := range t {
		for j := range t {
			for k := range t {
				PE(&tmp).Mul(&a[i][k], &b[k][j])
				PE(&c[i][j]).Add(&c[i][j], &tmp)
			}
		}
	}
	return c
}

// charPoly returns the characteristic polynomial of a as coefficients in
// increasing degree, monic of degree t, by the Faddeev-LeVerrier recurrence
//
//	M_1 = I,  M_k = a·M_{k-1} + c_{t-k+1}·I,  c_{t-k} = -tr(a·M_k)/k.
//
// Division by k is legitimate: t is at most the state width, far below the
// characteristic of any field this is used with.
func charPoly[E any, PE Ring[E]](a [][]E) []E {
	t := len(a)
	c := make([]E, t+1)
	PE(&c[t]).SetOne()

	m := newMat[E, PE](t)
	for k := 1; k <= t; k++ {
		if k == 1 {
			for i := range t {
				PE(&m[i][i]).SetOne()
			}
		} else {
			m = matMul[E, PE](a, m)
			for i := range t {
				PE(&m[i][i]).Add(&m[i][i], &c[t-k+1])
			}
		}
		am := matMul[E, PE](a, m)
		var tr E
		for i := range t {
			PE(&tr).Add(&tr, &am[i][i])
		}
		var kInv E
		PE(&kInv).SetUint64(uint64(k))
		PE(&kInv).Inverse(&kInv)
		PE(&tr).Mul(&tr, &kInv)

		var zero E
		PE(&c[t-k]).Sub(&zero, &tr)
	}
	return c
}

// --- univariate polynomials over the field, coefficients in increasing degree ---

func polyTrim[E any, PE Ring[E]](f []E) []E {
	i := len(f) - 1
	for i > 0 && PE(&f[i]).IsZero() {
		i--
	}
	return f[:i+1]
}

// polyMod reduces f modulo the monic polynomial g.
func polyMod[E any, PE Ring[E]](f, g []E) []E {
	dg := len(g) - 1
	r := make([]E, len(f))
	copy(r, f)
	var tmp E
	for i := len(r) - 1; i >= dg; i-- {
		if PE(&r[i]).IsZero() {
			continue
		}
		var coef E
		PE(&coef).Set(&r[i])
		for j := 0; j <= dg; j++ {
			PE(&tmp).Mul(&coef, &g[j])
			PE(&r[i-dg+j]).Sub(&r[i-dg+j], &tmp)
		}
	}
	if len(r) > dg {
		r = r[:dg]
	}
	if len(r) == 0 {
		r = make([]E, 1)
	}
	return polyTrim[E, PE](r)
}

func polyMulMod[E any, PE Ring[E]](a, b, g []E) []E {
	r := make([]E, len(a)+len(b)-1)
	var tmp E
	for i := range a {
		if PE(&a[i]).IsZero() {
			continue
		}
		for j := range b {
			PE(&tmp).Mul(&a[i], &b[j])
			PE(&r[i+j]).Add(&r[i+j], &tmp)
		}
	}
	return polyMod[E, PE](polyTrim[E, PE](r), g)
}

func polySub[E any, PE Ring[E]](a, b []E) []E {
	n := max(len(a), len(b))
	r := make([]E, n)
	for i := range r {
		if i < len(a) {
			PE(&r[i]).Set(&a[i])
		}
		if i < len(b) {
			PE(&r[i]).Sub(&r[i], &b[i])
		}
	}
	return polyTrim[E, PE](r)
}

// polyGCD returns a monic greatest common divisor of a and b.
func polyGCD[E any, PE Ring[E]](a, b []E) []E {
	a, b = polyTrim[E, PE](a), polyTrim[E, PE](b)
	for !(len(b) == 1 && PE(&b[0]).IsZero()) {
		var lc E
		PE(&lc).Inverse(&b[len(b)-1])
		bm := make([]E, len(b))
		for i := range b {
			PE(&bm[i]).Mul(&b[i], &lc)
		}
		a, b = bm, polyMod[E, PE](a, bm)
	}
	return a
}

// xPowQ returns x^(q^k) mod g, with q the field characteristic.
func xPowQ[E any, PE Ring[E]](k int, g []E, modulus *big.Int) []E {
	cur := make([]E, 2)
	PE(&cur[1]).SetOne() // x
	for range k {
		base := cur
		res := make([]E, 1)
		PE(&res[0]).SetOne()
		for i := modulus.BitLen() - 1; i >= 0; i-- {
			res = polyMulMod[E, PE](res, res, g)
			if modulus.Bit(i) == 1 {
				res = polyMulMod[E, PE](res, base, g)
			}
		}
		cur = res
	}
	return cur
}

// isIrreducible applies the standard Rabin test: f of degree n is irreducible
// iff x^(q^n) == x mod f and gcd(x^(q^(n/r)) - x, f) == 1 for every prime r | n.
func isIrreducible[E any, PE Ring[E]](f []E, modulus *big.Int) bool {
	n := len(f) - 1
	if n < 1 {
		return false
	}
	x := make([]E, 2)
	PE(&x[1]).SetOne()
	x = polyMod[E, PE](x, f) // degree-1 f reduces x to a constant

	d := polySub[E, PE](xPowQ[E, PE](n, f, modulus), x)
	if !(len(d) == 1 && PE(&d[0]).IsZero()) {
		return false
	}
	for r := 2; r <= n; r++ {
		if n%r != 0 || !isPrime(r) {
			continue
		}
		g := polyGCD[E, PE](f, polySub[E, PE](xPowQ[E, PE](n/r, f, modulus), x))
		if len(g) != 1 {
			return false
		}
	}
	return true
}

func isPrime(n int) bool {
	if n < 2 {
		return false
	}
	for i := 2; i*i <= n; i++ {
		if n%i == 0 {
			return false
		}
	}
	return true
}
