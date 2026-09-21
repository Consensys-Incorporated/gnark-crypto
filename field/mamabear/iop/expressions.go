// Copyright 2020-2026 Consensys Software Inc.
// Licensed under the Apache License, Version 2.0. See the LICENSE file for details.

package iop

import (
	"errors"
	"math/bits"

	"github.com/consensys/gnark-crypto/field/mamabear"
	"github.com/consensys/gnark-crypto/parallel"
)

// Expression represents a multivariate polynomial.
type Expression func(i int, x ...mamabear.Element) mamabear.Element

// Evaluate evaluates f on each entry of x. The returned value is
// the vector of evaluations of e on x.
// The form of the result is form.
// if r is provided (not nil), it is used as the result vector,
// that is the memory space for the coefficients of the resulting polynomial.
// The Size field of the result is the same as the one of x[0].
// The blindedSize field of the result is the same as Size.
// The Shift field of the result is 0.
func Evaluate(f Expression, r []mamabear.Element, form Form, x ...*Polynomial) (*Polynomial, error) {
	if len(x) == 0 {
		return nil, errors.New("need at least one input")
	}

	// check that the sizes are consistent
	n := x[0].coefficients.Len()
	m := len(x)
	for i := 1; i < m; i++ {
		if n != x[i].coefficients.Len() {
			return nil, ErrInconsistentSize
		}
	}

	// check result len
	if r == nil {
		r = make([]mamabear.Element, n)
	} else if len(r) != n {
		return nil, ErrInconsistentSize
	}

	// result coefficients
	idx := func(i int) int {
		return i
	}
	if form.Layout != Regular {
		nn := uint64(64 - bits.TrailingZeros(uint(n)))
		idx = func(i int) int {
			return int(bits.Reverse64(uint64(i)) >> nn)
		}
	}

	parallel.Execute(n, func(start, end int) {
		vx := make([]mamabear.Element, m)
		for i := start; i < end; i++ {
			for j := range m {
				vx[j] = x[j].GetCoeff(i)
			}
			r[idx(i)] = f(i, vx...)
		}
	})

	res := NewPolynomial(&r, form)
	res.size = x[0].size
	res.shift = 0

	return res, nil
}
