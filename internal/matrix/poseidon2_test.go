// Copyright 2020-2026 Consensys Incorporated
// Licensed under the Apache License, Version 2.0. See the LICENSE file for details.

package matrix

import (
	"testing"

	fr "github.com/consensys/gnark-crypto/field/koalabear"
)

func fe(v uint64) fr.Element {
	var e fr.Element
	e.SetUint64(v)
	return e
}

func feNeg(v uint64) fr.Element {
	var e, z fr.Element
	e.SetUint64(v)
	e.Sub(&z, &e)
	return e
}

// internalMatrix builds M_I = J + diag(d), the shape every Poseidon2 internal
// matrix in this repository has.
func internalMatrix(d []fr.Element) [][]fr.Element {
	t := len(d)
	m := make([][]fr.Element, t)
	for i := range m {
		m[i] = make([]fr.Element, t)
		for j := range m[i] {
			m[i][j].SetOne()
		}
		m[i][i].Add(&m[i][i], &d[i])
	}
	return m
}

// TestCharPolyCayleyHamilton checks charPoly against the Cayley-Hamilton
// theorem: the characteristic polynomial must annihilate its own matrix.
func TestCharPolyCayleyHamilton(t *testing.T) {
	d := []fr.Element{feNeg(2), fe(1), fe(2), fe(7), fe(3), fe(4), feNeg(5), feNeg(3)}
	m := internalMatrix(d)
	cp := charPoly(m)

	if len(cp)-1 != len(d) {
		t.Fatalf("charpoly degree: got %d, want %d", len(cp)-1, len(d))
	}
	if !cp[len(cp)-1].IsOne() {
		t.Fatal("charpoly is not monic")
	}

	// acc = sum_k cp[k] * m^k
	n := len(d)
	acc := newMat[fr.Element, *fr.Element](n)
	pow := newMat[fr.Element, *fr.Element](n)
	for i := range n {
		pow[i][i].SetOne()
	}
	var tmp fr.Element
	for k := range cp {
		for i := range n {
			for j := range n {
				tmp.Mul(&cp[k], &pow[i][j])
				acc[i][j].Add(&acc[i][j], &tmp)
			}
		}
		pow = matMul(m, pow)
	}
	for i := range n {
		for j := range n {
			if !acc[i][j].IsZero() {
				t.Fatalf("Cayley-Hamilton failed at (%d,%d)", i, j)
			}
		}
	}
}

// TestIsIrreducible pins the Rabin test against polynomials whose factorization
// is known, so a checker that answers "irreducible" to everything fails here.
func TestIsIrreducible(t *testing.T) {
	q := fr.Modulus()

	// x - 1: irreducible (degree 1).
	linear := []fr.Element{feNeg(1), fe(1)}
	if !isIrreducible(linear, q) {
		t.Error("x - 1 should be irreducible")
	}

	// (x - 1)(x - 2) = x^2 - 3x + 2: reducible.
	reducible := []fr.Element{fe(2), feNeg(3), fe(1)}
	if isIrreducible(reducible, q) {
		t.Error("(x-1)(x-2) should be reducible")
	}

	// x^2 - g for a non-residue g: irreducible. koalabear has p = 3 mod 4 only
	// if -1 is a non-residue; find a non-residue by trial instead of assuming.
	var nonResidue fr.Element
	for v := uint64(2); v < 200; v++ {
		nonResidue.SetUint64(v)
		if nonResidue.Legendre() == -1 {
			break
		}
	}
	var negNR fr.Element
	negNR.Neg(&nonResidue)
	quadratic := []fr.Element{negNR, {}, fe(1)} // x^2 - nonResidue
	if !isIrreducible(quadratic, q) {
		t.Error("x^2 - (non-residue) should be irreducible")
	}

	// x^2 - s for a residue s: reducible.
	var residue, negR fr.Element
	residue.SetUint64(4)
	negR.Neg(&residue)
	if isIrreducible([]fr.Element{negR, {}, fe(1)}, q) {
		t.Error("x^2 - 4 should be reducible")
	}
}

// TestHasNoInvariantSubspaceRejects checks the condition is not vacuous: the
// all-ones-plus-identity matrix has an obvious invariant subspace and must be
// rejected.
func TestHasNoInvariantSubspaceRejects(t *testing.T) {
	d := make([]fr.Element, 8)
	for i := range d {
		d[i].SetOne() // every diagonal entry equal => repeated eigenvalue
	}
	if HasNoInvariantSubspace(internalMatrix(d), fr.Modulus()) {
		t.Fatal("J + I should be rejected")
	}
}

// TestHasNoInvariantSubspaceAccepts checks it against a diagonal known to be
// good: the width-16 internal diagonal koalabear ships, which comes from
// Plonky3 and was generated under this very condition.
func TestHasNoInvariantSubspaceAccepts(t *testing.T) {
	raw := []uint64{2130706431, 1, 2, 1065353217, 3, 4, 1065353216, 2130706430, 2130706429, 2122383361, 1864368129, 2130706306, 8323072, 266338304, 133169152, 127}
	d := make([]fr.Element, len(raw))
	for i, v := range raw {
		d[i].SetUint64(v)
	}
	if !HasNoInvariantSubspace(internalMatrix(d), fr.Modulus()) {
		t.Fatal("koalabear's shipped diag16 should pass")
	}
}
