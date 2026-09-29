package vortex

import (
	"math/rand/v2"
	"testing"

	"github.com/consensys/gnark-crypto/field/mamabear"
	fext "github.com/consensys/gnark-crypto/field/mamabear/extensions"
	"github.com/stretchr/testify/require"
)

func TestLagrangeSimple(t *testing.T) {
	assert := require.New(t)
	params, err := NewParams(4, 4, nil, 2, 2)
	assert.NoError(err)

	t.Run("0-1-2-3", func(t *testing.T) {

		v := []mamabear.Element{
			mamabear.NewElement(0),
			mamabear.NewElement(1),
			mamabear.NewElement(2),
			mamabear.NewElement(3),
		}

		codeword := make([]mamabear.Element, params.SizeCodeWord())
		params.EncodeReedSolomon(v, codeword)

		for i := 0; i < len(codeword); i += 2 {
			if codeword[i] != v[i/2] {
				t.Errorf("failure at position (%v %v)", i, i/2)
			}
		}
	})

	t.Run("shifting", func(t *testing.T) {

		v := []mamabear.Element{
			mamabear.NewElement(0),
			mamabear.NewElement(1),
			mamabear.NewElement(2),
			mamabear.NewElement(3),
		}

		vShifted := []mamabear.Element{
			mamabear.NewElement(1),
			mamabear.NewElement(2),
			mamabear.NewElement(3),
			mamabear.NewElement(0),
		}

		codeword := make([]mamabear.Element, params.SizeCodeWord())
		params.EncodeReedSolomon(v, codeword)

		codewordShifted := make([]mamabear.Element, params.SizeCodeWord())
		params.EncodeReedSolomon(vShifted, codewordShifted)

		for i := range codeword {

			iShifted := i - 2
			if iShifted < 0 {
				iShifted += 8
			}

			if codeword[i] != codewordShifted[iShifted] {
				t.Errorf("mismatch between codeword and shifted codeword")
			}
		}

	})
}

func TestReedSolomonProperty(t *testing.T) {
	assert := require.New(t)

	var (
		size         = 16
		invRate      = 2
		v            = make([]mamabear.Element, size)
		encodedVFext = make([]fext.E3, size*invRate)

		// #nosec G404 -- test case generation does not require a cryptographic PRNG
		rng   = rand.New(rand.NewChaCha8([32]byte{}))
		randX = randFext(rng)
	)
	params, err := NewParams(size, 4, nil, 2, 2)
	assert.NoError(err)

	for i := range v {
		v[i] = randElement(rng)
	}

	encodedV := make([]mamabear.Element, params.SizeCodeWord())
	params.EncodeReedSolomon(v, encodedV)

	for i := range encodedVFext {
		encodedVFext[i].A0.Set(&encodedV[i])
	}

	assert.True(params.IsReedSolomonCodewords(encodedVFext), "codeword does not pass rs check")

	y0, err := EvalBasePolyLagrange(v, randX)
	assert.NoError(err)

	y1, err := EvalBasePolyLagrange(encodedV, randX)
	assert.NoError(err)

	y2, err := EvalFextPolyLagrange(encodedVFext, randX)
	assert.NoError(err)

	assert.Equal(y0, y1)
	assert.Equal(y0, y2)

}

// TestReedSolomonAllRates covers the coset-decomposition encoder across every
// supported inverse rate. The other tests in this file all use rate 2, which is
// the one rate served by the in-place specialization.
func TestReedSolomonAllRates(t *testing.T) {
	assert := require.New(t)

	// #nosec G404 -- test case generation does not require a cryptographic PRNG
	rng := rand.New(rand.NewChaCha8([32]byte{}))

	for _, invRate := range []int{2, 4, 8} {
		for _, size := range []int{1, 2, 4, 16, 64, 256} {
			params, err := NewParams(size, 4, nil, invRate, 2)
			assert.NoError(err)

			v := make([]mamabear.Element, size)
			for i := range v {
				v[i] = randElement(rng)
			}

			encoded := make([]mamabear.Element, params.SizeCodeWord())
			params.EncodeReedSolomon(v, encoded)

			// The rate-th points of the codeword are the input itself: coset 0
			// of the codeword domain is the interpolation domain.
			for i := range v {
				assert.Equal(v[i], encoded[invRate*i],
					"rate %d size %d: input not embedded at index %d", invRate, size, invRate*i)
			}

			// The codeword must lie in the Reed-Solomon code.
			encodedFext := make([]fext.E3, params.SizeCodeWord())
			for i := range encodedFext {
				encodedFext[i].A0.Set(&encoded[i])
			}
			assert.True(params.IsReedSolomonCodewords(encodedFext),
				"rate %d size %d: codeword does not pass rs check", invRate, size)

			// Encoding must not depend on the state of the output buffer.
			dirty := make([]mamabear.Element, params.SizeCodeWord())
			for i := range dirty {
				dirty[i] = randElement(rng)
			}
			params.EncodeReedSolomon(v, dirty)
			assert.Equal(encoded, dirty,
				"rate %d size %d: encoding depends on stale output buffer", invRate, size)
		}
	}
}
