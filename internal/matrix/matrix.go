// Copyright 2020-2026 Consensys Software Inc.
// Licensed under the Apache License, Version 2.0. See the LICENSE file for details.

// Package matrix holds helpers on matrices over generic ring elements.
package matrix

// FromLinearMap returns the dense width×width matrix of the linear map mul, which
// overwrites its argument with its image. Column j of the matrix is the image of
// the j-th basis vector. The linearity of mul is not checked.
func FromLinearMap[E any, PE interface {
	*E
	SetOne() *E
}](width int, mul func([]E)) [][]E {
	m := make([][]E, width)
	for i := range m {
		m[i] = make([]E, width)
	}

	col := make([]E, width)
	for j := range width {
		clear(col)
		PE(&col[j]).SetOne()
		mul(col)
		for i := range col {
			m[i][j] = col[i]
		}
	}
	return m
}
