//go:build purego || !amd64

// Copyright 2020-2026 Consensys Software Inc.
// Licensed under the Apache License, Version 2.0. See the LICENSE file for details.

package mamabear

func (vector *Vector) Add(a, b Vector) {
	addVecGeneric(*vector, a, b)
}

func (vector *Vector) Sub(a, b Vector) {
	subVecGeneric(*vector, a, b)
}

func (vector *Vector) ScalarMul(a Vector, b *Element) {
	scalarMulVecGeneric(*vector, a, b)
}

func (vector *Vector) Sum() (res Element) {
	sumVecGeneric(&res, *vector)
	return
}

func (vector *Vector) InnerProduct(other Vector) (res Element) {
	innerProductVecGeneric(&res, *vector, other)
	return
}

func (vector *Vector) Mul(a, b Vector) {
	mulVecGeneric(*vector, a, b)
}
