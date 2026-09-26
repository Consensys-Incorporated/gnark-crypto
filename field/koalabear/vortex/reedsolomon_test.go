package vortex

import (
	"math/rand/v2"
	"testing"

	"github.com/consensys/gnark-crypto/field/koalabear"
	fext "github.com/consensys/gnark-crypto/field/koalabear/extensions"
	"github.com/stretchr/testify/require"
)

func TestLagrangeSimple(t *testing.T) {
	assert := require.New(t)
	params, err := NewParams(4, 4, nil, 2, 2)
	assert.NoError(err)

	t.Run("0-1-2-3", func(t *testing.T) {

		v := []koalabear.Element{
			koalabear.NewElement(0),
			koalabear.NewElement(1),
			koalabear.NewElement(2),
			koalabear.NewElement(3),
		}

		codeword := make([]koalabear.Element, params.SizeCodeWord())
		params.EncodeReedSolomon(v, codeword)

		for i := 0; i < len(codeword); i += 2 {
			if codeword[i] != v[i/2] {
				t.Errorf("failure at position (%v %v)", i, i/2)
			}
		}
	})

	t.Run("shifting", func(t *testing.T) {

		v := []koalabear.Element{
			koalabear.NewElement(0),
			koalabear.NewElement(1),
			koalabear.NewElement(2),
			koalabear.NewElement(3),
		}

		vShifted := []koalabear.Element{
			koalabear.NewElement(1),
			koalabear.NewElement(2),
			koalabear.NewElement(3),
			koalabear.NewElement(0),
		}

		codeword := make([]koalabear.Element, params.SizeCodeWord())
		params.EncodeReedSolomon(v, codeword)

		codewordShifted := make([]koalabear.Element, params.SizeCodeWord())
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
		v            = make([]koalabear.Element, size)
		encodedVFext = make([]fext.E4, size*invRate)

		// #nosec G404 -- test case generation does not require a cryptographic PRNG
		rng   = rand.New(rand.NewChaCha8([32]byte{}))
		randX = randFext(rng)
	)
	params, err := NewParams(size, 4, nil, 2, 2)
	assert.NoError(err)

	for i := range v {
		v[i] = randElement(rng)
	}

	encodedV := make([]koalabear.Element, params.SizeCodeWord())
	params.EncodeReedSolomon(v, encodedV)

	for i := range encodedVFext {
		encodedVFext[i].B0.A0.Set(&encodedV[i])
	}

	assert.True(params.IsReedSolomonCodewords(encodedVFext), "codeword does not pass rs check")
	assert.False(params.IsReedSolomonCodewords(encodedVFext[:len(encodedVFext)-1]), "short codeword must be rejected")
	assert.False(params.IsReedSolomonCodewords(append(encodedVFext, fext.E4{})), "oversized codeword must be rejected")

	y0, err := EvalBasePolyLagrange(v, randX)
	assert.NoError(err)

	y1, err := EvalBasePolyLagrange(encodedV, randX)
	assert.NoError(err)

	y2, err := EvalFextPolyLagrange(encodedVFext, randX)
	assert.NoError(err)

	assert.Equal(y0, y1)
	assert.Equal(y0, y2)

}
