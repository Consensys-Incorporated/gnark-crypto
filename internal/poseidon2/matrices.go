// Copyright 2020-2026 Consensys Software Inc.
// Licensed under the Apache License, Version 2.0. See the LICENSE file for details.

// Package poseidon2 holds the definitions of the linear layers of the Poseidon2
// permutation that are shared by the generated field and curve packages.
package poseidon2

// ExternalMatrixKind names the source of the 4×4 block M4 from which the
// external matrix is built for the widths that are multiples of 4.
type ExternalMatrixKind int

const (
	// Plonky3 is the block used by Plonky3, whose rows are
	//
	//	2 3 1 1
	//	1 2 3 1
	//	1 1 2 3
	//	3 1 1 2
	Plonky3 ExternalMatrixKind = iota
	// Paper is the block of the Poseidon2 paper (https://eprint.iacr.org/2023/323.pdf,
	// appendix B), whose rows are
	//
	//	5 7 1 3
	//	4 6 1 1
	//	1 3 5 7
	//	1 1 4 6
	Paper
)

// m4Blocks holds the block M4 of each ExternalMatrixKind.
var m4Blocks = [...][4][4]int64{
	Plonky3: {
		{2, 3, 1, 1},
		{1, 2, 3, 1},
		{1, 1, 2, 3},
		{3, 1, 1, 2},
	},
	Paper: {
		{5, 7, 1, 3},
		{4, 6, 1, 1},
		{1, 3, 5, 7},
		{1, 1, 4, 6},
	},
}

// ring is the set of operations on a pointer PE to a ring element E that the
// matrix constructions need.
type ring[E any] interface {
	*E
	SetInt64(int64) *E
	Add(*E, *E) *E
}

// ExternalMatrix returns the dense width×width external matrix M_E, which the
// permutation applies before the first round and after each full round. With I
// and J the identity and the all-ones matrices, it is
//
//   - I + J for width 2 and 3, whatever the kind;
//   - M4 for width 4;
//   - circ(2·M4, M4, ..., M4) = (I + J) ⊗ M4 for width 4k with k > 1, where
//     I and J have size k.
//
// M4 is the block selected by kind. It panics for any other width, and for an
// unknown kind when a block is needed.
func ExternalMatrix[E any, PE ring[E]](width int, kind ExternalMatrixKind) [][]E {
	m := newMatrix[E](width)

	switch {
	case width == 2 || width == 3:
		for i := range m {
			for j := range m[i] {
				v := int64(1)
				if i == j {
					v = 2
				}
				PE(&m[i][j]).SetInt64(v)
			}
		}
		return m
	case width%4 != 0 || width <= 0:
		panic("poseidon2: only widths 2, 3 and multiples of 4 are supported")
	}

	if kind < 0 || int(kind) >= len(m4Blocks) {
		panic("poseidon2: unknown external matrix kind")
	}
	m4 := &m4Blocks[kind]
	for i := range m {
		for j := range m[i] {
			v := m4[i%4][j%4]
			if width > 4 && i/4 == j/4 {
				v *= 2
			}
			PE(&m[i][j]).SetInt64(v)
		}
	}
	return m
}

// InternalMatrix returns the dense internal matrix M_I = J + diag(d) of width
// len(d), where J is the all-ones matrix: entry (i, j) is 1, plus d[i] when
// i == j. For a state x, (M_I·x)_i = Σ_j x_j + d[i]·x_i. The permutation applies
// M_I after each partial round.
func InternalMatrix[E any, PE ring[E]](d []E) [][]E {
	m := newMatrix[E](len(d))
	for i := range m {
		for j := range m[i] {
			PE(&m[i][j]).SetInt64(1)
		}
		PE(&m[i][i]).Add(PE(&m[i][i]), PE(&d[i]))
	}
	return m
}

func newMatrix[E any](n int) [][]E {
	m := make([][]E, n)
	for i := range m {
		m[i] = make([]E, n)
	}
	return m
}
