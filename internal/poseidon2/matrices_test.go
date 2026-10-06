// Copyright 2020-2026 Consensys Software Inc.
// Licensed under the Apache License, Version 2.0. See the LICENSE file for details.

package poseidon2

import (
	"testing"

	fr "github.com/consensys/gnark-crypto/field/koalabear"
)

func fromRows(rows [][]int64) [][]fr.Element {
	m := newMatrix[fr.Element](len(rows))
	for i := range rows {
		for j := range rows[i] {
			m[i][j].SetInt64(rows[i][j])
		}
	}
	return m
}

func requireEqual(t *testing.T, name string, got, want [][]fr.Element) {
	t.Helper()
	if len(got) != len(want) {
		t.Fatalf("%s: got %d rows, want %d", name, len(got), len(want))
	}
	for i := range want {
		if len(got[i]) != len(want[i]) {
			t.Fatalf("%s: row %d has %d entries, want %d", name, i, len(got[i]), len(want[i]))
		}
		for j := range want[i] {
			if !got[i][j].Equal(&want[i][j]) {
				t.Fatalf("%s: entry (%d, %d) is %s, want %s", name, i, j, got[i][j].String(), want[i][j].String())
			}
		}
	}
}

func TestExternalMatrixSmallWidths(t *testing.T) {
	for _, kind := range []ExternalMatrixKind{Plonky3, Paper} {
		requireEqual(t, "width 2", ExternalMatrix[fr.Element](2, kind), fromRows([][]int64{
			{2, 1},
			{1, 2},
		}))
		requireEqual(t, "width 3", ExternalMatrix[fr.Element](3, kind), fromRows([][]int64{
			{2, 1, 1},
			{1, 2, 1},
			{1, 1, 2},
		}))
	}
}

func TestExternalMatrixWidth4IsM4(t *testing.T) {
	requireEqual(t, "plonky3", ExternalMatrix[fr.Element](4, Plonky3), fromRows([][]int64{
		{2, 3, 1, 1},
		{1, 2, 3, 1},
		{1, 1, 2, 3},
		{3, 1, 1, 2},
	}))
	requireEqual(t, "paper", ExternalMatrix[fr.Element](4, Paper), fromRows([][]int64{
		{5, 7, 1, 3},
		{4, 6, 1, 1},
		{1, 3, 5, 7},
		{1, 1, 4, 6},
	}))
}

// TestExternalMatrixCirculant checks circ(2·M4, M4, ..., M4) for the widths 4k, k > 1.
func TestExternalMatrixCirculant(t *testing.T) {
	for _, kind := range []ExternalMatrixKind{Plonky3, Paper} {
		for _, width := range []int{8, 12, 16, 24} {
			got := ExternalMatrix[fr.Element](width, kind)
			want := newMatrix[fr.Element](width)
			for bi := range width / 4 {
				for bj := range width / 4 {
					for r := range 4 {
						for c := range 4 {
							v := m4Blocks[kind][r][c]
							if bi == bj {
								v *= 2
							}
							want[4*bi+r][4*bj+c].SetInt64(v)
						}
					}
				}
			}
			requireEqual(t, "circulant", got, want)
		}
	}
}

func TestExternalMatrixPanics(t *testing.T) {
	for _, width := range []int{-4, 0, 1, 5, 6, 7, 10} {
		func() {
			defer func() {
				if recover() == nil {
					t.Fatalf("width %d: expected a panic", width)
				}
			}()
			ExternalMatrix[fr.Element](width, Plonky3)
		}()
	}
	func() {
		defer func() {
			if recover() == nil {
				t.Fatal("expected a panic for an unknown kind")
			}
		}()
		ExternalMatrix[fr.Element](4, ExternalMatrixKind(7))
	}()
}

func TestInternalMatrix(t *testing.T) {
	d := make([]fr.Element, 4)
	for i, v := range []int64{3, -1, 0, 7} {
		d[i].SetInt64(v)
	}
	requireEqual(t, "internal", InternalMatrix(d), fromRows([][]int64{
		{4, 1, 1, 1},
		{1, 0, 1, 1},
		{1, 1, 1, 1},
		{1, 1, 1, 8},
	}))
	// d is not modified
	var want fr.Element
	want.SetInt64(3)
	if !d[0].Equal(&want) {
		t.Fatal("InternalMatrix modified its argument")
	}
}
