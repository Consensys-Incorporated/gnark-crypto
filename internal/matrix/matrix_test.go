// Copyright 2020-2026 Consensys Software Inc.
// Licensed under the Apache License, Version 2.0. See the LICENSE file for details.

package matrix

import (
	"testing"

	fr "github.com/consensys/gnark-crypto/field/koalabear"
)

// TestFromLinearMap checks that FromLinearMap recovers a non-symmetric matrix, and so the
// orientation of its columns.
func TestFromLinearMap(t *testing.T) {
	rows := [][]uint64{
		{1, 2, 3},
		{4, 5, 6},
		{7, 8, 10},
	}
	var a [3][3]fr.Element
	for i := range rows {
		for j := range rows[i] {
			a[i][j].SetUint64(rows[i][j])
		}
	}
	mul := func(x []fr.Element) {
		var res [3]fr.Element
		for i := range a {
			var tmp fr.Element
			for j := range a[i] {
				res[i].Add(&res[i], tmp.Mul(&a[i][j], &x[j]))
			}
		}
		copy(x, res[:])
	}

	m := FromLinearMap(3, mul)
	for i := range a {
		for j := range a[i] {
			if !m[i][j].Equal(&a[i][j]) {
				t.Fatalf("entry (%d, %d) is %s, want %s", i, j, m[i][j].String(), a[i][j].String())
			}
		}
	}
}
