// Copyright 2020-2026 Consensys Software Inc.
// Licensed under the Apache License, Version 2.0. See the LICENSE file for details.

package mamabear

import "math/bits"

// montMul computes a · b · R⁻¹ mod p where R = 2^52.
//
// Implements Algorithm 1 of the MamaBearZKP paper in pure Go:
//
//	t    ← ⌊a·b / R⌋                    (high bits of the product)
//	c₀   ← a·b mod R                    (low 52 bits)
//	m₀   ← c₀ · qInvNeg mod R
//	q'   ← ⌊m₀ · p / R⌋
//	carry← 1 if c₀ > 0, else 0          (carry from c₀ + m₀·p mod R = R)
//	r    ← t + q' + carry               (lazy result ∈ [0, 9p/8))
//
// For canonical inputs a, b ∈ [0, p), the output is in [0, 9p/8) ⊂ [0, 2p).
// One conditional subtract brings it to [0, p).
func montMul(a, b uint64) uint64 {
	// Full 104-bit product a·b stored as hi:lo (lo = low 64 bits, hi ≤ 2^40).
	hi, lo := bits.Mul64(a, b)

	// c₀ = a·b mod 2^52 (low 52 bits)
	c0 := lo & rMask

	// t = floor(a·b / 2^52): bits 52..63 of lo in positions 0..11,
	// bits 0..39 of hi in positions 12..51; no overlap.
	t := (lo >> rBits) | (hi << (64 - rBits))

	// m₀ = c₀ · (-p⁻¹) mod 2^52
	m0 := (c0 * qInvNeg) & rMask

	// floor(m₀ · p / 2^52); the low 52 bits of m₀·p equal (R − c₀) by Montgomery
	mHi, mLo := bits.Mul64(m0, q)
	qVal := (mLo >> rBits) | (mHi << (64 - rBits))

	// carry = 1 when c₀ > 0: c₀ + (m₀·p mod R) = R exactly, so the addition
	// carries into bit 52.  When c₀ = 0, m₀ = 0 and the sum is 0 (no carry).
	carry := (c0 + (mLo & rMask)) >> rBits

	// Result = t + qVal + carry ≡ a·b·R⁻¹ (mod p), in [0, 9p/8).
	// One conditional subtract in the caller brings it to [0, p).
	return t + qVal + carry
}

func fromMont(z *Element) {
	_fromMontGeneric(z)
}

func reduce(z *Element) {
	_reduceGeneric(z)
}

// Mul z = x · y (mod p)
//
// Result is in canonical form [0, p).
func (z *Element) Mul(x, y *Element) *Element {
	r := montMul(x[0], y[0])
	// lazy result ∈ [0, 9p/8); one conditional subtract gives [0, p)
	if r >= q {
		r -= q
	}
	z[0] = r
	return z
}

// Square z = x · x (mod p)
func (z *Element) Square(x *Element) *Element {
	r := montMul(x[0], x[0])
	if r >= q {
		r -= q
	}
	z[0] = r
	return z
}

// Butterfly sets a, b = a + b (mod p), a − b (mod p)
func Butterfly(a, b *Element) {
	_butterflyGeneric(a, b)
}
