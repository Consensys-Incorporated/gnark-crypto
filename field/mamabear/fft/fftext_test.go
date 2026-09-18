// Copyright 2020-2026 Consensys Software Inc.
// Licensed under the Apache License, Version 2.0. See the LICENSE file for details.

package fft

import (
	"math/big"
	"testing"

	"github.com/consensys/gnark-crypto/utils"

	"github.com/consensys/gnark-crypto/field/mamabear"
	fext "github.com/consensys/gnark-crypto/field/mamabear/extensions"

	"github.com/leanovate/gopter"
	"github.com/leanovate/gopter/gen"
	"github.com/leanovate/gopter/prop"
)

func TestFFTExt(t *testing.T) {
	parameters := gopter.DefaultTestParameters()
	parameters.MinSuccessfulTests = 6
	properties := gopter.NewProperties(parameters)

	for maxSize := 2; maxSize <= 1<<10; maxSize <<= 1 {

		domainWithPrecompute := NewDomain(uint64(maxSize))
		domainWithoutPrecompute := NewDomain(uint64(maxSize), WithoutPrecompute())

		for domainName, domain := range map[string]*Domain{
			"with precompute":    domainWithPrecompute,
			"without precompute": domainWithoutPrecompute,
		} {
			t.Logf("domain: %s", domainName)

			properties.Property("DIF FFTExt should be consistent with dual basis", prop.ForAll(
				func(ithpower int) bool {
					pol := make([]fext.E3, maxSize)
					backupPol := make([]fext.E3, maxSize)
					for i := range maxSize {
						pol[i].MustSetRandom()
					}
					copy(backupPol, pol)

					domain.FFTExt(pol, DIF)
					utils.BitReverse(pol)

					sample := domain.Generator
					sample.Exp(sample, big.NewInt(int64(ithpower)))

					eval := evaluatePolynomialExt(backupPol, sample)
					return eval.Equal(&pol[ithpower])
				},
				gen.IntRange(0, maxSize-1),
			))

			properties.Property("DIT FFTExt should be consistent with dual basis", prop.ForAll(
				func(ithpower int) bool {
					pol := make([]fext.E3, maxSize)
					backupPol := make([]fext.E3, maxSize)
					for i := range maxSize {
						pol[i].MustSetRandom()
					}
					copy(backupPol, pol)

					utils.BitReverse(pol)
					domain.FFTExt(pol, DIT)

					sample := domain.Generator
					sample.Exp(sample, big.NewInt(int64(ithpower)))

					eval := evaluatePolynomialExt(backupPol, sample)
					return eval.Equal(&pol[ithpower])
				},
				gen.IntRange(0, maxSize-1),
			))

			properties.Property("DIT FFTExt(DIF FFTExt)==id", prop.ForAll(
				func() bool {
					pol := make([]fext.E3, maxSize)
					backupPol := make([]fext.E3, maxSize)
					for i := range maxSize {
						pol[i].MustSetRandom()
					}
					copy(backupPol, pol)

					domain.FFTInverseExt(pol, DIF)
					domain.FFTExt(pol, DIT)

					for i := range pol {
						if !pol[i].Equal(&backupPol[i]) {
							return false
						}
					}
					return true
				},
			))

			properties.Property("BitReverse(DIF FFTExt(DIT FFTExt(BitReverse)))==id", prop.ForAll(
				func() bool {
					pol := make([]fext.E3, maxSize)
					backupPol := make([]fext.E3, maxSize)
					for i := range maxSize {
						pol[i].MustSetRandom()
					}
					copy(backupPol, pol)

					utils.BitReverse(pol)
					domain.FFTExt(pol, DIT)
					domain.FFTInverseExt(pol, DIF)
					utils.BitReverse(pol)

					for i := range pol {
						if !pol[i].Equal(&backupPol[i]) {
							return false
						}
					}
					return true
				},
			))

			properties.Property("DIF FFTExt on cosets should be consistent with dual basis", prop.ForAll(
				func(ithpower int) bool {
					pol := make([]fext.E3, maxSize)
					backupPol := make([]fext.E3, maxSize)
					for i := range maxSize {
						pol[i].MustSetRandom()
					}
					copy(backupPol, pol)

					domain.FFTExt(pol, DIF, OnCoset())
					utils.BitReverse(pol)

					sample := domain.Generator
					sample.Exp(sample, big.NewInt(int64(ithpower))).
						Mul(&sample, &domain.FrMultiplicativeGen)

					eval := evaluatePolynomialExt(backupPol, sample)
					return eval.Equal(&pol[ithpower])
				},
				gen.IntRange(0, maxSize-1),
			))

			properties.Property("DIT FFTExt(DIF FFTExt)==id on cosets", prop.ForAll(
				func() bool {
					pol := make([]fext.E3, maxSize)
					backupPol := make([]fext.E3, maxSize)
					for i := range maxSize {
						pol[i].MustSetRandom()
					}
					copy(backupPol, pol)

					domain.FFTInverseExt(pol, DIF, OnCoset())
					domain.FFTExt(pol, DIT, OnCoset())

					for i := range pol {
						if !pol[i].Equal(&backupPol[i]) {
							return false
						}
					}
					return true
				},
			))
		}
		properties.TestingRun(t, gopter.ConsoleReporter(false))
	}
}

// evaluatePolynomialExt evaluates sum_i pol[i] * val^i where pol[i] ∈ F_{p³} and val ∈ F_p.
func evaluatePolynomialExt(pol []fext.E3, val mamabear.Element) fext.E3 {
	var res, tmp fext.E3
	var acc mamabear.Element
	res.Set(&pol[0])
	acc.SetOne()
	for i := 1; i < len(pol); i++ {
		acc.Mul(&acc, &val)
		tmp.MulByElement(&pol[i], &acc)
		res.Add(&res, &tmp)
	}
	return res
}

// ---- Benchmarks ---------------------------------------------------------------

func BenchmarkFFTExt(b *testing.B) {
	const maxSize = 1 << 16
	pol := make([]fext.E3, maxSize)
	for i := range pol {
		pol[i].MustSetRandom()
	}
	domain := NewDomain(maxSize)

	b.ResetTimer()
	for range b.N {
		domain.FFTExt(pol, DIF)
	}
}

func BenchmarkFFTInverseExt(b *testing.B) {
	const maxSize = 1 << 16
	pol := make([]fext.E3, maxSize)
	for i := range pol {
		pol[i].MustSetRandom()
	}
	domain := NewDomain(maxSize)

	b.ResetTimer()
	for range b.N {
		domain.FFTInverseExt(pol, DIF)
	}
}
