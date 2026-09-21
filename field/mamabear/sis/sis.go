// Copyright 2020-2026 Consensys Software Inc.
// Licensed under the Apache License, Version 2.0. See the LICENSE file for details.

package sis

import (
	"encoding/binary"
	"errors"
	"fmt"
	"math/bits"

	"github.com/consensys/gnark-crypto/field/mamabear"
	"github.com/consensys/gnark-crypto/field/mamabear/fft"
	"github.com/consensys/gnark-crypto/parallel"
	"golang.org/x/crypto/blake2b"
)

// RSis is the Ring-SIS instance
type RSis struct {
	// Vectors in ℤ_{p}/Xⁿ+1
	// A[i] is the i-th polynomial.
	// Ag the evaluation form of the polynomials in A on the coset √(g) * <g>
	A  [][]mamabear.Element
	Ag [][]mamabear.Element

	// LogTwoBound (Infinity norm) of the vector to hash. It means that each component in m
	// is < 2^B, where m is the vector to hash (the hash being A*m).
	// cf https://hackmd.io/7OODKWQZRRW9RxM5BaXtIw , B >= 3.
	LogTwoBound int

	// d, the degree of X^{d}+1
	Degree int

	// domain for the polynomial multiplication
	Domain *fft.Domain

	maxNbElementsToHash int

	kz mamabear.Vector // zeroes used to zeroize the limbs buffer faster.
}

// NewRSis creates an instance of RSis.
// seed: seed for the randomness for generating A.
// logTwoDegree: if d := logTwoDegree, the ring will be ℤ_{p}[X]/Xᵈ-1, where X^{2ᵈ} is the 2ᵈ⁺¹-th cyclotomic polynomial
// logTwoBound: the bound of the vector to hash (using the infinity norm).
// maxNbElementsToHash: maximum number of field elements the instance handles
// used to derived n, the number of polynomials in A, and max size of instance's internal buffer.
func NewRSis(seed int64, logTwoDegree, logTwoBound, maxNbElementsToHash int) (*RSis, error) {

	if logTwoBound > 64 || logTwoBound > mamabear.Bits {
		return nil, errors.New("logTwoBound too large")
	}
	if logTwoBound%8 != 0 {
		return nil, errors.New("logTwoBound must be a multiple of 8")
	}
	if bits.UintSize == 32 {
		return nil, errors.New("unsupported architecture; need 64bit target")
	}

	degree := 1 << logTwoDegree

	// n: number of polynomials in A
	// len(m) == degree * n
	// with each element in m being logTwoBounds bits from the instance buffer.
	// that is, to fill m, we need [degree * n * logTwoBound] bits of data

	// First n <- #limbs to represent a single field element
	nbBytesPerLimb := logTwoBound / 8
	if mamabear.Bytes%nbBytesPerLimb != 0 {
		return nil, errors.New("nbBytesPerLimb must divide field size")
	}
	n := mamabear.Bytes / nbBytesPerLimb

	// Then multiply by the number of field elements
	n *= maxNbElementsToHash

	// And divide (+ ceil) to get the number of polynomials
	if n%degree == 0 {
		n /= degree
	} else {
		n /= degree // number of polynomials
		n++
	}

	// domains (shift is √{gen} )
	shift, err := mamabear.Generator(uint64(2 * degree))
	if err != nil {
		return nil, err
	}

	r := &RSis{
		LogTwoBound:         logTwoBound,
		Degree:              degree,
		Domain:              fft.NewDomain(uint64(degree), fft.WithShift(shift)),
		A:                   make([][]mamabear.Element, n),
		Ag:                  make([][]mamabear.Element, n),
		kz:                  make(mamabear.Vector, degree),
		maxNbElementsToHash: maxNbElementsToHash,
	}

	// filling A
	a := make([]mamabear.Element, n*r.Degree)
	ag := make([]mamabear.Element, n*r.Degree)

	parallel.Execute(n, func(start, end int) {
		for i := start; i < end; i++ {
			rstart, rend := i*r.Degree, (i+1)*r.Degree
			r.A[i] = a[rstart:rend:rend]
			r.Ag[i] = ag[rstart:rend:rend]
			for j := range r.Degree {
				r.A[i][j] = deriveRandomElementFromSeed(seed, int64(i), int64(j))
			}

			// fill Ag the evaluation form of the polynomials in A on the coset √(g) * <g>
			copy(r.Ag[i], r.A[i])
			r.Domain.FFT(r.Ag[i], fft.DIF, fft.OnCoset(), fft.WithNbTasks(1))
		}
	})

	return r, nil
}

// Hash interprets the input vector as a sequence of coefficients of size r.LogTwoBound bits long,
// and return the hash of the polynomial corresponding to the sum sum_i A[i]*m Mod X^{d}+1
func (r *RSis) Hash(v, res []mamabear.Element) error {
	if len(res) != r.Degree {
		return fmt.Errorf("output vector must have length %d", r.Degree)
	}

	if len(v) > r.maxNbElementsToHash {
		return fmt.Errorf("can't hash more than %d elements with params provided in constructor", r.maxNbElementsToHash)
	}

	// zeroing res
	for i := range res {
		res[i].SetZero()
	}

	// inner hash
	k := make([]mamabear.Element, r.Degree)
	it := NewLimbIterator(&VectorIterator{v: v}, r.LogTwoBound/8)
	mask := ^uint64(0) // mask is unused in MamaBear (no unrolled FFT path)
	for i := range len(r.Ag) {
		r.InnerHash(&it, res, k, r.kz, i, mask)
	}

	// reduces mod Xᵈ+1
	r.Domain.FFTInverse(res, fft.DIT, fft.OnCoset(), fft.WithNbTasks(1))

	return nil
}

// InnerHash computes the inner hash of the polynomial corresponding to the i-th polynomial in A.
// It accumulates the result in res.
// It does not reduce mod Xᵈ+1.
// res, k, kz must have size r.Degree.
// kz is a buffer of zeroes used to zeroize the limbs buffer faster.
// mask is unused in MamaBear.
func (r *RSis) InnerHash(it *LimbIterator, res, k, kz mamabear.Vector, polId int, mask uint64) {
	copy(k, kz)
	zero := uint32(0)

	for j := range r.Degree {
		l, ok := it.NextLimb()
		if !ok {
			break
		}
		zero |= l
		k[j][0] = uint64(l)
	}
	if zero == 0 {
		// means m[i*r.Degree : (i+1)*r.Degree] == [0...0]
		// we can skip this, FFT(0) = 0
		return
	}
	r.Domain.FFT(k, fft.DIF, fft.OnCoset(), fft.WithNbTasks(1))

	// we compute k * r.Ag[polId] in ℤ_{p}[X]/Xᵈ+1.
	// k and r.Ag[polId] are in evaluation form on √(g) * <g>
	// we accumulate the result in res; the FFT inverse is done once every multiplications are done.
	k.Mul(k, mamabear.Vector(r.Ag[polId]))
	res.Add(res, k)
}

func deriveRandomElementFromSeed(seed, i, j int64) mamabear.Element {
	var buf [3 + 3*8]byte
	copy(buf[:3], "SIS")
	binary.BigEndian.PutUint64(buf[3:], uint64(seed))
	binary.BigEndian.PutUint64(buf[11:], uint64(i))
	binary.BigEndian.PutUint64(buf[19:], uint64(j))

	digest := blake2b.Sum256(buf[:])

	var res mamabear.Element
	res.SetBytes(digest[:])

	return res
}

// ElementIterator is an iterator over a stream of field elements.
type ElementIterator interface {
	Next() (mamabear.Element, bool)
}

// VectorIterator iterates over a vector of field element.
type VectorIterator struct {
	v mamabear.Vector
	i int
}

// NewVectorIterator creates a new VectorIterator
func NewVectorIterator(v mamabear.Vector) *VectorIterator {
	return &VectorIterator{v: v}
}

// Next returns the next element of the vector.
func (vi *VectorIterator) Next() (mamabear.Element, bool) {
	if vi.i == len(vi.v) {
		return mamabear.Element{}, false
	}
	vi.i++
	return vi.v[vi.i-1], true
}

// LimbIterator iterates over a stream of field elements, limb by limb.
type LimbIterator struct {
	v        mamabear.Vector
	vi       int
	buf      [mamabear.Bytes]byte
	j        int // position in buf
	limbSize int
}

// NewLimbIterator creates a new LimbIterator
// it is an iterator over a stream of field elements
// The elements are interpreted in little endian.
// The limb is also in little endian.
func NewLimbIterator(it ElementIterator, limbSize int) LimbIterator {
	switch limbSize {
	case 1, 2, 4:

	default:
		panic("unsupported limb size")
	}

	// Keep the hot iterator state concrete and return by value: storing the
	// ElementIterator interface here makes VectorIterator escape to the heap.
	vi, ok := it.(*VectorIterator)
	if !ok {
		panic("unsupported element iterator")
	}

	return LimbIterator{
		v:        vi.v,
		vi:       vi.i,
		j:        mamabear.Bytes,
		limbSize: limbSize,
	}
}

// NextLimb returns the next limb of the vector.
func (vr *LimbIterator) NextLimb() (uint32, bool) {
	if vr.j == mamabear.Bytes {
		if vr.vi == len(vr.v) {
			return 0, false
		}
		vr.j = 0
		mamabear.LittleEndian.PutElement(&vr.buf, vr.v[vr.vi])
		vr.vi++
	}

	var r uint32
	switch vr.limbSize {
	case 1:
		r = uint32(vr.buf[vr.j])
	case 2:
		r = uint32(binary.LittleEndian.Uint16(vr.buf[vr.j:]))
	case 4:
		r = binary.LittleEndian.Uint32(vr.buf[vr.j:])
	default:
		panic("unsupported limb size")
	}
	vr.j += vr.limbSize
	return r, true
}

// Reset resets the iterator with a new ElementIterator.
func (vr *LimbIterator) Reset(it ElementIterator) {
	vi, ok := it.(*VectorIterator)
	if !ok {
		panic("unsupported element iterator")
	}
	vr.v = vi.v
	vr.vi = vi.i
	vr.j = mamabear.Bytes
}
