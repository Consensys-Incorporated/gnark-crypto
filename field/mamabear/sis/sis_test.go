// Copyright 2020-2026 Consensys Software Inc.
// Licensed under the Apache License, Version 2.0. See the LICENSE file for details.

package sis

import (
	"encoding/binary"
	"fmt"
	"math/big"
	"math/bits"
	"math/rand/v2"
	"testing"

	"github.com/consensys/gnark-crypto/field/mamabear"
	"github.com/consensys/gnark-crypto/field/mamabear/fft"
	"github.com/stretchr/testify/require"
)

type sisParams struct {
	logTwoBound, logTwoDegree int
}

// valid logTwoBound for MamaBear: multiple of 8, ≤ 49, and mamabear.Bytes (8) divisible by nbBytesPerLimb
// valid values: 8 (1B), 16 (2B), 32 (4B)
var params128Bits []sisParams = []sisParams{
	{logTwoBound: 8, logTwoDegree: 5},
	{logTwoBound: 8, logTwoDegree: 6},
	{logTwoBound: 16, logTwoDegree: 6},
	{logTwoBound: 16, logTwoDegree: 9},
}

func TestSISConsistency(t *testing.T) {
	if bits.UintSize == 32 {
		t.Skip("skipping this test in 32bit.")
	}
	assert := require.New(t)

	const maxNbElementsToHash = 1 << 10

	for _, param := range params128Bits {
		t.Logf("logTwoBound = %d, logTwoDegree = %d", param.logTwoBound, param.logTwoDegree)

		instance, err := NewRSis(42, param.logTwoDegree, param.logTwoBound, maxNbElementsToHash)
		assert.NoError(err)

		degree := 1 << param.logTwoDegree
		inputs := make(mamabear.Vector, degree)
		for i := range inputs {
			inputs[i].MustSetRandom()
		}

		res0 := make([]mamabear.Element, degree)
		res1 := make([]mamabear.Element, degree)

		err = instance.Hash(inputs, res0)
		assert.NoError(err)

		// Hash again — should be deterministic
		err = instance.Hash(inputs, res1)
		assert.NoError(err)

		for i := range res0 {
			assert.True(res0[i].Equal(&res1[i]), "results differ at index %d", i)
		}
	}
}

func TestLimbDecomposeBytes(t *testing.T) {
	assert := require.New(t)

	// MamaBear uses R = 2^52 (IFMA Montgomery radix), not 2^{Bytes*8} = 2^64.
	// Setting m[i][0] = l stores l in Montgomery form, representing l*R^{-1}.
	// Multiplying by montConstant = R restores l.
	var montConstant mamabear.Element
	var bMontConstant big.Int
	bMontConstant.SetUint64(1)
	bMontConstant.Lsh(&bMontConstant, 52) // R = 2^52
	montConstant.SetBigInt(&bMontConstant)

	nbElmts := 10
	a := make([]mamabear.Element, nbElmts)
	for i := range nbElmts {
		a[i].MustSetRandom()
	}

	for _, logTwoBound := range []int{8, 16, 32} {
		vr := NewLimbIterator(&VectorIterator{v: a}, logTwoBound/8)
		m := make(mamabear.Vector, nbElmts*mamabear.Bytes*8/logTwoBound)
		var ok bool
		for i := range len(m) {
			var l uint32
			l, ok = vr.NextLimb()
			assert.True(ok)
			m[i][0] = uint64(l)
		}

		for i := range len(m) {
			m[i].Mul(&m[i], &montConstant)
		}

		var x mamabear.Element
		x.SetUint64(1 << logTwoBound)

		coeffsPerFieldElmt := mamabear.Bytes * 8 / logTwoBound
		for i := range nbElmts {
			r := eval(m[i*coeffsPerFieldElmt:(i+1)*coeffsPerFieldElmt], x)
			assert.True(r.Equal(&a[i]), "limbDecomposeBytes failed for logTwoBound=%d at element %d", logTwoBound, i)
		}
	}
}

func eval(p []mamabear.Element, x mamabear.Element) mamabear.Element {
	var res mamabear.Element
	for i := len(p) - 1; i >= 0; i-- {
		res.Mul(&res, &x).Add(&res, &p[i])
	}
	return res
}

func makeKeyDeterministic(t *testing.T, sis *RSis, _seed int64) {
	t.Helper()
	// generate the key deterministically, the same way
	// we do in sage to generate the test vectors.

	polyRand := func(seed mamabear.Element, deg int) []mamabear.Element {
		res := make([]mamabear.Element, deg)
		for i := range deg {
			res[i].Square(&seed)
			seed.Set(&res[i])
		}
		return res
	}

	var seed, one mamabear.Element
	one.SetOne()
	seed.SetInt64(_seed)
	for i := range len(sis.A) {
		sis.A[i] = polyRand(seed, sis.Degree)
		copy(sis.Ag[i], sis.A[i])
		sis.Domain.FFT(sis.Ag[i], fft.DIF, fft.OnCoset())
		seed.Add(&seed, &one)
	}
}

func BenchmarkSIS(b *testing.B) {

	// max nb field elements to hash
	const nbInputs = 1 << 16

	inputs := make(mamabear.Vector, nbInputs)
	for i := range len(inputs) {
		inputs[i].MustSetRandom()
	}

	for _, param := range params128Bits {
		for n := 1 << 10; n <= nbInputs; n <<= 1 {
			in := inputs[:n]
			benchmarkSIS(b, in, false, param.logTwoBound, param.logTwoDegree)
		}

	}
}

func benchmarkSIS(b *testing.B, input []mamabear.Element, sparse bool, logTwoBound, logTwoDegree int) {
	b.Helper()

	n := len(input)

	benchName := "ring-sis/"
	if sparse {
		benchName += "sparse/"
	}
	benchName += fmt.Sprintf("inputs=%v/log2-bound=%v/log2-degree=%v", n, logTwoBound, logTwoDegree)

	b.Run(benchName, func(b *testing.B) {
		// report the throughput in MB/s
		b.SetBytes(int64(len(input)) * mamabear.Bytes)

		instance, err := NewRSis(0, logTwoDegree, logTwoBound, n)
		if err != nil {
			b.Fatal(err)
		}

		res := make([]mamabear.Element, 1<<logTwoDegree)

		b.ResetTimer()
		for range b.N {
			_ = instance.Hash(input, res)
		}
	})
}

const q = uint64(562932773552129)

func randElement(rng *rand.Rand) mamabear.Element {
	return mamabear.Element{rng.Uint64N(q)}
}

func FuzzSIS(f *testing.F) {

	f.Fuzz(func(t *testing.T, rngSeed, sisSeed int64, logTwoDegree uint16, logTwoBoundSwitch bool) {
		assert := require.New(t)

		if logTwoDegree > 10 || logTwoDegree < 2 {
			t.Skip("logTwoDegree out of range")
		}
		degree := int(1 << logTwoDegree)

		logTwoBound := 16
		if logTwoBoundSwitch {
			logTwoBound = 8
		}

		var seed [32]byte
		binary.PutVarint(seed[:], rngSeed)
		// #nosec G404 -- fuzz does not require a cryptographic PRNG
		rng := rand.New(rand.NewChaCha8(seed))

		// max elements to hash will be in [degree: 4*degree]
		maxElementsToHash := int(rng.IntN(3*degree)) + degree

		// size of input will be in [1: maxElementsToHash]
		size := int(rng.IntN(maxElementsToHash)) + 1

		// Create a new RSIS instance
		instance, err := NewRSis(sisSeed, int(logTwoDegree), logTwoBound, maxElementsToHash)
		assert.NoError(err, "failed to create SIS params")

		a0 := make([]mamabear.Element, size)
		a1 := make([]mamabear.Element, size)

		for i := range a0 {
			a0[i] = randElement(rng)
		}

		copy(a1[:], a0[:])

		res0 := make([]mamabear.Element, degree)
		res1 := make([]mamabear.Element, degree)
		err = instance.Hash(a0, res0)
		assert.NoError(err, "hashing failed")

		err = instance.Hash(a1, res1)
		assert.NoError(err, "hashing failed")

		// compare the results
		for i := range res0 {
			assert.True(res0[i].Equal(&res1[i]), "results differ at index %d", i)
		}

	})
}
