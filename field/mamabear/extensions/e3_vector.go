// Copyright 2020-2026 Consensys Software Inc.
// Licensed under the Apache License, Version 2.0. See the LICENSE file for details.

package extensions

import (
	"unsafe"

	fr "github.com/consensys/gnark-crypto/field/mamabear"
)

// Butterfly sets a = a+b and b = a-b (mod p) for each component.
func Butterfly(a, b *E3) {
	t := *a
	a.Add(a, b)
	b.Sub(&t, b)
}

// Vector is a slice of E3 elements.
type Vector []E3

// Butterfly computes the in-place butterfly between two same-length vectors:
//
//	vector[i] = vector[i] + other[i]
//	other[i]  = old vector[i] - other[i]
func (vector Vector) Butterfly(other Vector) {
	if len(vector) != len(other) {
		panic("vector.Butterfly: length mismatch")
	}
	for i := range vector {
		Butterfly(&vector[i], &other[i])
	}
}

// ButterflyPair applies Butterfly to each adjacent pair in the vector:
// (vector[0], vector[1]), (vector[2], vector[3]), ...
// Length must be even.
func (vector Vector) ButterflyPair() {
	if len(vector)%2 != 0 {
		panic("vector.ButterflyPair: length must be even")
	}
	for i := 0; i < len(vector); i += 2 {
		Butterfly(&vector[i], &vector[i+1])
	}
}

// MulByElement sets vector[i] = a[i] * b[i] where b[i] ∈ F_p.
func (vector Vector) MulByElement(a Vector, b fr.Vector) {
	if len(vector) != len(a) || len(vector) != len(b) {
		panic("vector.MulByElement: length mismatch")
	}
	for i := range vector {
		vector[i].MulByElement(&a[i], &b[i])
	}
}

// ScalarMulByElement sets vector[i] = a[i] * b for all i, where b ∈ F_p.
//
// Reinterprets the E3 slice as a flat fr.Vector (3 contiguous fr.Elements per E3)
// to leverage the optimized fr.Vector.ScalarMul path.
func (vector Vector) ScalarMulByElement(a Vector, b *fr.Element) {
	if len(vector) != len(a) {
		panic("vector.ScalarMulByElement: length mismatch")
	}
	if len(vector) == 0 {
		return
	}
	// E3 = {A0, A1, A2} with no padding — safe to reinterpret as 3×fr.Element.
	M := len(a) * 3
	vBase := fr.Vector(unsafe.Slice((*fr.Element)(unsafe.Pointer(&a[0])), M))
	vRes := fr.Vector(unsafe.Slice((*fr.Element)(unsafe.Pointer(&vector[0])), M))
	vRes.ScalarMul(vBase, b)
}
