package vortex

import (
	"fmt"
	"math/big"

	"github.com/consensys/gnark-crypto/field/koalabear"
	fext "github.com/consensys/gnark-crypto/field/koalabear/extensions"
	"github.com/consensys/gnark-crypto/field/koalabear/fft"
	"github.com/consensys/gnark-crypto/parallel"
)

// BatchEvalFextPolyLagrange evaluates extension field polynomials in Lagrange basis at the same point x
//
// Key optimization: The Lagrange basis {L₀(x), L₁(x), ..., Lₙ₋₁(x)} is computed once and reused
// for all polynomials, making batch evaluation significantly more efficient than individual evaluations.
func BatchEvalFextPolyLagrange(polys [][]fext.E4, x fext.E4, oncoset ...bool) ([]fext.E4, error) {

	if len(polys) == 0 {
		return []fext.E4{}, nil
	}

	err := checkSizeConsistency(polys)
	if err != nil {
		return nil, err
	}

	n := len(polys[0])
	lagrangeBasis, err := ComputeLagrangeBasisAtX(n, x, oncoset...)
	if err != nil {
		return nil, err
	}

	// Compute results in parallel
	results := make([]fext.E4, len(polys))
	parallel.Execute(len(polys), func(start, stop int) {
		for k := start; k < stop; k++ {
			res := fext.VectorE4(polys[k]).InnerProduct(fext.VectorE4(lagrangeBasis))
			results[k] = res
		}
	})

	return results, nil
}

// BatchEvalBasePolyLagrange evaluates base field polynomials in Lagrange basis at the same point x, returning extension field results
//
// Key optimization: The Lagrange basis {L₀(x), L₁(x), ..., Lₙ₋₁(x)} is computed once and reused
// for all polynomials, making batch evaluation significantly more efficient than individual evaluations.
func BatchEvalBasePolyLagrange(polys [][]koalabear.Element, x fext.E4, oncoset ...bool) ([]fext.E4, error) {

	if len(polys) == 0 {
		return []fext.E4{}, nil
	}

	err := checkSizeConsistency(polys)
	if err != nil {
		return nil, err
	}

	n := len(polys[0])
	lagrangeBasis, err := ComputeLagrangeBasisAtX(n, x, oncoset...)
	if err != nil {
		return nil, err
	}

	results := make([]fext.E4, len(polys))
	parallel.Execute(len(polys), func(start, stop int) {
		for k := start; k < stop; k++ {
			res := fext.VectorE4(lagrangeBasis).InnerProductByElement(polys[k])
			results[k] = res
		}
	})

	return results, nil
}

// ComputeLagrangeBasisAtX computes (Lᵢ(x))_{i<n} and numerator for Lagrange basis evaluation
func ComputeLagrangeBasisAtX(n int, x fext.E4, oncoset ...bool) ([]fext.E4, error) {

	generator, _ := fft.Generator(uint64(n))
	generatorInv := new(koalabear.Element).Inverse(&generator)
	one := koalabear.One()

	// Handle coset evaluation
	if len(oncoset) > 0 && oncoset[0] {
		frMultiplicativeGen := fft.GeneratorFullMultiplicativeGroup()
		frMultiplicativeGenInv := new(koalabear.Element).Inverse(&frMultiplicativeGen)
		x.MulByElement(&x, frMultiplicativeGenInv)
	}

	// (xⁿ - 1) / n
	var numerator fext.E4
	numerator.Exp(x, big.NewInt(int64(n)))
	numerator.B0.A0.Sub(&numerator.B0.A0, &one)

	cardInv := koalabear.NewElement(uint64(n))
	cardInv.Inverse(&cardInv)
	numerator.MulByElement(&numerator, &cardInv)
	numerator.Inverse(&numerator)

	// compute x-1, x/ω-1, x/ω²-1, ...
	res := make(fext.VectorE4, n)
	res[0] = x
	for i := 1; i < n; i++ {
		res[i].MulByElement(&res[i-1], generatorInv)
	}
	isRootOfUnity := -1
	for i := range res {
		res[i].B0.A0.Sub(&res[i].B0.A0, &one)
		if res[i].IsZero() { // it means that x is a root of unity
			isRootOfUnity = i
			break
		}
	}
	if isRootOfUnity != -1 {
		res = make(fext.VectorE4, n)
		res[isRootOfUnity].SetOne()
		return res, nil
	}
	res.ScalarMul(res, &numerator)

	// 1/(x-1), 1/(x/ω-1), 1/(x/ω²-1), ...
	res = fext.BatchInvertE4(res)

	return res, nil
}

// checkSizeConsistencyBase check that the polynomial are of the same size, and that the size is a power of two
func checkSizeConsistency[T any](polys [][]T) error {
	n := len(polys[0])
	for i := range polys {
		if len(polys[i]) != n {
			return fmt.Errorf("all polys should have the same length, expected %d but poly[%d] has length %d",
				n, i, len(polys[i]))
		}
	}
	if !isPowerOfTwo(n) {
		return fmt.Errorf("only support powers of two but poly has length %v", n)
	}
	return nil
}
