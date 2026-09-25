package vortex

import (
	"fmt"
	"math/big"

	"github.com/consensys/gnark-crypto/field/mamabear"
	fext "github.com/consensys/gnark-crypto/field/mamabear/extensions"
	"github.com/consensys/gnark-crypto/field/mamabear/fft"
	"github.com/consensys/gnark-crypto/parallel"
)

// BatchEvalFextPolyLagrange evaluates extension field polynomials in Lagrange basis at the same point x.
func BatchEvalFextPolyLagrange(polys [][]fext.E3, x fext.E3, oncoset ...bool) ([]fext.E3, error) {

	if len(polys) == 0 {
		return []fext.E3{}, nil
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

	results := make([]fext.E3, len(polys))
	parallel.Execute(len(polys), func(start, stop int) {
		for k := start; k < stop; k++ {
			res := fext.Vector(polys[k]).InnerProduct(fext.Vector(lagrangeBasis))
			results[k] = res
		}
	})

	return results, nil
}

// BatchEvalBasePolyLagrange evaluates base field polynomials in Lagrange basis at the same point x.
func BatchEvalBasePolyLagrange(polys [][]mamabear.Element, x fext.E3, oncoset ...bool) ([]fext.E3, error) {

	if len(polys) == 0 {
		return []fext.E3{}, nil
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

	results := make([]fext.E3, len(polys))
	parallel.Execute(len(polys), func(start, stop int) {
		for k := start; k < stop; k++ {
			res := fext.Vector(lagrangeBasis).InnerProductByElement(polys[k])
			results[k] = res
		}
	})

	return results, nil
}

// ComputeLagrangeBasisAtX computes (Lᵢ(x))_{i<n} for Lagrange basis evaluation.
func ComputeLagrangeBasisAtX(n int, x fext.E3, oncoset ...bool) ([]fext.E3, error) {

	generator, _ := fft.Generator(uint64(n))
	generatorInv := new(mamabear.Element).Inverse(&generator)
	one := mamabear.One()

	if len(oncoset) > 0 && oncoset[0] {
		frMultiplicativeGen := fft.GeneratorFullMultiplicativeGroup()
		frMultiplicativeGenInv := new(mamabear.Element).Inverse(&frMultiplicativeGen)
		x.MulByElement(&x, frMultiplicativeGenInv)
	}

	// (xⁿ - 1) / n
	var numerator fext.E3
	numerator.Exp(x, big.NewInt(int64(n)))
	numerator.A0.Sub(&numerator.A0, &one)

	cardInv := mamabear.NewElement(uint64(n))
	cardInv.Inverse(&cardInv)
	numerator.MulByElement(&numerator, &cardInv)
	numerator.Inverse(&numerator)

	res := make(fext.Vector, n)
	res[0] = x
	for i := 1; i < n; i++ {
		res[i].MulByElement(&res[i-1], generatorInv)
	}
	isRootOfUnity := -1
	for i := range res {
		res[i].A0.Sub(&res[i].A0, &one)
		if res[i].IsZero() {
			isRootOfUnity = i
			break
		}
	}
	if isRootOfUnity != -1 {
		res = make(fext.Vector, n)
		res[isRootOfUnity].SetOne()
		return res, nil
	}
	res.ScalarMul(res, &numerator)

	res = fext.BatchInvertE3(res)

	return res, nil
}

// checkSizeConsistency checks that all polynomials have the same size (a power of two).
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
