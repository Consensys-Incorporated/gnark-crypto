//go:build !purego

// Copyright 2020-2026 Consensys Software Inc.
// Licensed under the Apache License, Version 2.0. See the LICENSE file for details.

package mamabear

import (
	"github.com/consensys/gnark-crypto/utils/cpu"
)

//go:noescape
func addVec(res, a, b *Element, n uint64)

//go:noescape
func subVec(res, a, b *Element, n uint64)

//go:noescape
func sumVec(t *uint64, a *Element, n uint64)

//go:noescape
func mulVec(res, a, b *Element, n uint64)

//go:noescape
func scalarMulVec(res, a, b *Element, n uint64)

//go:noescape
func innerProdVec(t *uint64, a, b *Element, n uint64)

// blockSize is the number of elements processed per AVX-512IFMA ZMM register.
// 8 lanes × 64-bit = 512-bit ZMM register.
const blockSize = 8

// maxSumBlocks is the largest number of 8-element blocks that can be accumulated
// per lane before 64-bit overflow. Each element < p < 2^49; 2^14 × 2^49 = 2^63 < 2^64.
const maxSumBlocks = 1 << 13

// Add adds two vectors element-wise and stores the result in self.
// It panics if the vectors don't have the same length.
func (vector *Vector) Add(a, b Vector) {
	if len(a) != len(b) || len(a) != len(*vector) {
		panic("vector.Add: vectors don't have the same length")
	}
	n := uint64(len(a))
	if n == 0 {
		return
	}
	if !cpu.SupportAVX512IFMA {
		addVecGeneric(*vector, a, b)
		return
	}
	addVec(&(*vector)[0], &a[0], &b[0], n/blockSize)
	if n%blockSize != 0 {
		start := n - n%blockSize
		addVecGeneric((*vector)[start:], a[start:], b[start:])
	}
}

// Sub subtracts two vectors element-wise and stores the result in self.
// It panics if the vectors don't have the same length.
func (vector *Vector) Sub(a, b Vector) {
	if len(a) != len(b) || len(a) != len(*vector) {
		panic("vector.Sub: vectors don't have the same length")
	}
	n := uint64(len(a))
	if n == 0 {
		return
	}
	if !cpu.SupportAVX512IFMA {
		subVecGeneric(*vector, a, b)
		return
	}
	subVec(&(*vector)[0], &a[0], &b[0], n/blockSize)
	if n%blockSize != 0 {
		start := n - n%blockSize
		subVecGeneric((*vector)[start:], a[start:], b[start:])
	}
}

// ScalarMul multiplies a vector by a scalar element-wise and stores the result in self.
// It panics if the vectors don't have the same length.
func (vector *Vector) ScalarMul(a Vector, b *Element) {
	if len(a) != len(*vector) {
		panic("vector.ScalarMul: vectors don't have the same length")
	}
	n := uint64(len(a))
	if n == 0 {
		return
	}
	if !cpu.SupportAVX512IFMA {
		scalarMulVecGeneric(*vector, a, b)
		return
	}
	scalarMulVec(&(*vector)[0], &a[0], b, n/blockSize)
	if n%blockSize != 0 {
		start := n - n%blockSize
		scalarMulVecGeneric((*vector)[start:], a[start:], b)
	}
}

// Sum computes the sum of all elements in the vector.
func (vector *Vector) Sum() (res Element) {
	n := uint64(len(*vector))
	if n == 0 {
		return
	}
	if !cpu.SupportAVX512IFMA {
		sumVecGeneric(&res, *vector)
		return
	}
	// Process in chunks to prevent 64-bit accumulator overflow.
	// Each element < p < 2^49; per lane, maxSumBlocks × 2^49 < 2^63 < 2^64.
	var t [blockSize]uint64
	var v Element
	fullBlocks := n / blockSize
	for start := uint64(0); start < fullBlocks; {
		chunk := fullBlocks - start
		if chunk > maxSumBlocks {
			chunk = maxSumBlocks
		}
		// Zero accumulator between chunks (sumVec adds into t).
		for i := range t {
			t[i] = 0
		}
		sumVec(&t[0], &(*vector)[start*blockSize], chunk)
		for i := range blockSize {
			v[0] = t[i] % q
			res.Add(&res, &v)
		}
		start += chunk
	}
	if n%blockSize != 0 {
		startElem := fullBlocks * blockSize
		sumVecGeneric(&res, (*vector)[startElem:])
	}
	return
}

// InnerProduct computes the inner product of two vectors.
// It panics if the vectors don't have the same length.
func (vector *Vector) InnerProduct(other Vector) (res Element) {
	n := uint64(len(*vector))
	if n != uint64(len(other)) {
		panic("vector.InnerProduct: vectors don't have the same length")
	}
	if n == 0 {
		return
	}
	if !cpu.SupportAVX512IFMA {
		innerProductVecGeneric(&res, *vector, other)
		return
	}
	// Chunk-and-reduce to prevent 64-bit accumulator overflow.
	var t [blockSize]uint64
	var v Element
	fullBlocks := n / blockSize
	for start := uint64(0); start < fullBlocks; {
		chunk := fullBlocks - start
		if chunk > maxSumBlocks {
			chunk = maxSumBlocks
		}
		for i := range t {
			t[i] = 0
		}
		innerProdVec(&t[0], &(*vector)[start*blockSize], &other[start*blockSize], chunk)
		for i := range blockSize {
			v[0] = t[i] % q
			res.Add(&res, &v)
		}
		start += chunk
	}
	if n%blockSize != 0 {
		startElem := fullBlocks * blockSize
		innerProductVecGeneric(&res, (*vector)[startElem:], other[startElem:])
	}
	return
}

// Mul multiplies two vectors element-wise and stores the result in self.
// It panics if the vectors don't have the same length.
func (vector *Vector) Mul(a, b Vector) {
	if len(a) != len(b) || len(a) != len(*vector) {
		panic("vector.Mul: vectors don't have the same length")
	}
	n := uint64(len(a))
	if n == 0 {
		return
	}
	if !cpu.SupportAVX512IFMA {
		mulVecGeneric(*vector, a, b)
		return
	}
	mulVec(&(*vector)[0], &a[0], &b[0], n/blockSize)
	if n%blockSize != 0 {
		start := n - n%blockSize
		mulVecGeneric((*vector)[start:], a[start:], b[start:])
	}
}
