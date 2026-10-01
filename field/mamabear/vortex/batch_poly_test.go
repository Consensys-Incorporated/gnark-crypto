package vortex

import (
	"testing"

	"github.com/consensys/gnark-crypto/field/mamabear"
	fext "github.com/consensys/gnark-crypto/field/mamabear/extensions"
	"github.com/consensys/gnark-crypto/field/mamabear/fft"
	"github.com/consensys/gnark-crypto/utils"
	"github.com/stretchr/testify/require"
)

func randomPoly(size int) []mamabear.Element {
	res := make([]mamabear.Element, size)
	for i := range res {
		res[i].SetRandom()
	}
	return res
}

func TestComputeLagrangeBasisAtX(t *testing.T) {

	n := 8
	g, _ := fft.Generator(8)
	ge := fext.E3{A0: g}
	var gi fext.E3
	gi.SetOne()

	expected := make([]fext.E3, n)
	for i := range n {
		expected[i].SetOne()
		cc, _ := ComputeLagrangeBasisAtX(n, gi)
		for j := range n {
			if !cc[j].Equal(&expected[j]) {
				t.Fatal("error computeLagrangeBasisAtX")
			}
		}
		expected[i].SetZero()
		gi.Mul(&gi, &ge)
	}

}

func randomPolyExt(size int) []fext.E3 {
	res := make([]fext.E3, size)
	for i := range res {
		res[i].MustSetRandom()
	}
	return res
}

func TestBatchEvaluateLagrangeOnFext(t *testing.T) {
	const sizePoly = 16
	const nbPoly = 20

	polys := make([][]fext.E3, nbPoly)
	for i := range polys {
		polys[i] = randomPolyExt(sizePoly)
	}

	var x fext.E3
	x.MustSetRandom()

	expected := make([]fext.E3, nbPoly)
	for i := range expected {
		expected[i] = EvalFextPolyHorner(polys[i], x)
	}

	domain := fft.NewDomain(uint64(sizePoly))

	testCases := []struct {
		name    string
		onCoset bool
	}{
		{"without coset", false},
		{"with coset", true},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			lagrangePolys := make([][]fext.E3, nbPoly)
			for i := range lagrangePolys {
				lagrangePolys[i] = append([]fext.E3{}, polys[i]...)
				if tc.onCoset {
					domain.FFTExt(lagrangePolys[i], fft.DIF, fft.OnCoset())
				} else {
					domain.FFTExt(lagrangePolys[i], fft.DIF)
				}
				utils.BitReverse(lagrangePolys[i])
			}

			results, err := BatchEvalFextPolyLagrange(lagrangePolys, x, tc.onCoset)
			require.NoError(t, err)

			for i := range results {
				require.Equal(t, expected[i].String(), results[i].String(),
					"Mismatch at polynomial %d", i)
			}
		})
	}
}

func TestBatchEvalBasePolyLagrange(t *testing.T) {
	const sizePoly = 64
	const nbPoly = 20

	polys := make([][]mamabear.Element, nbPoly)
	for i := range polys {
		polys[i] = randomPoly(sizePoly)
	}

	var x fext.E3
	x.MustSetRandom()

	expected := make([]fext.E3, nbPoly)
	for i := range expected {
		expected[i] = EvalBasePolyHorner(polys[i], x)
	}

	domain := fft.NewDomain(uint64(sizePoly))

	testCases := []struct {
		name    string
		onCoset bool
	}{
		{"without coset", false},
		{"with coset", true},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			lagrangePolys := make([][]mamabear.Element, nbPoly)
			for i := range lagrangePolys {
				lagrangePolys[i] = append([]mamabear.Element{}, polys[i]...)
				if tc.onCoset {
					domain.FFT(lagrangePolys[i], fft.DIF, fft.OnCoset())
				} else {
					domain.FFT(lagrangePolys[i], fft.DIF)
				}
				utils.BitReverse(lagrangePolys[i])
			}

			results, err := BatchEvalBasePolyLagrange(lagrangePolys, x, tc.onCoset)
			require.NoError(t, err)

			for i := range results {
				require.Equal(t, expected[i].String(), results[i].String(),
					"Mismatch at polynomial %d", i)
			}
		})
	}
}
