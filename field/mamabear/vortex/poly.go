package vortex

import (
	"fmt"
	"math/big"

	"github.com/consensys/gnark-crypto/field/mamabear"
	fext "github.com/consensys/gnark-crypto/field/mamabear/extensions"
	"github.com/consensys/gnark-crypto/field/mamabear/fft"
)

// EvalBasePolyLagrange evaluates a polynomial in Lagrange basis over the base field
// at a given point in the field extension.
func EvalBasePolyLagrange(poly []mamabear.Element, x fext.E3) (fext.E3, error) {

	if !isPowerOfTwo(len(poly)) {
		return fext.E3{}, fmt.Errorf("only support powers of two but poly has length %v", len(poly))
	}

	var (
		n            = len(poly)
		denominators = make([]fext.E3, n)
		one          = mamabear.One()
		generator, _ = fft.Generator(uint64(n))
		generatorInv = new(mamabear.Element).Inverse(&generator)
		cardInv      fext.E3
	)

	cardInv.A0 = mamabear.NewElement(uint64(n))
	cardInv.Inverse(&cardInv)

	denominators[0] = x
	for i := 1; i < n; i++ {
		denominators[i].MulByElement(&denominators[i-1], generatorInv)
	}

	for i := range n {
		denominators[i].A0.Sub(&denominators[i].A0, &one)
		if denominators[i].IsZero() {
			res := fext.E3{}
			res.A0.Set(&poly[i])
			return res, nil
		}
	}

	denominators = fext.BatchInvertE3(denominators)
	res, tmp := fext.E3{}, fext.E3{}
	for i := range denominators {
		tmp.MulByElement(&denominators[i], &poly[i])
		res.Add(&res, &tmp)
	}

	tmp.Exp(x, big.NewInt(int64(n)))
	tmp.A0.Sub(&tmp.A0, &one)
	tmp.Mul(&tmp, &cardInv)
	res.Mul(&res, &tmp)

	return res, nil
}

// EvalFextPolyLagrange evaluates a polynomial in Lagrange basis over the field extension
// at a given point in the field extension.
func EvalFextPolyLagrange(poly []fext.E3, x fext.E3) (fext.E3, error) {

	if !isPowerOfTwo(len(poly)) {
		return fext.E3{}, fmt.Errorf("only support powers of two but poly has length %v", len(poly))
	}

	var (
		n            = len(poly)
		denominators = make([]fext.E3, n)
		one          = mamabear.One()
		generator, _ = fft.Generator(uint64(n))
		generatorInv = new(mamabear.Element).Inverse(&generator)
		cardInv      fext.E3
	)

	cardInv.A0 = mamabear.NewElement(uint64(n))
	cardInv.Inverse(&cardInv)

	denominators[0] = x
	for i := 1; i < n; i++ {
		denominators[i].MulByElement(&denominators[i-1], generatorInv)
	}

	for i := range n {
		denominators[i].A0.Sub(&denominators[i].A0, &one)
		if denominators[i].IsZero() {
			return poly[i], nil
		}
	}

	denominators = fext.BatchInvertE3(denominators)
	res, tmp := fext.E3{}, fext.E3{}
	for i := range denominators {
		tmp.Mul(&denominators[i], &poly[i])
		res.Add(&res, &tmp)
	}

	tmp.Exp(x, big.NewInt(int64(n)))
	tmp.A0.Sub(&tmp.A0, &one)
	tmp.Mul(&tmp, &cardInv)
	res.Mul(&res, &tmp)

	return res, nil
}

// EvalFextPolyHorner evaluates a polynomial in coefficient basis over the field
// extension at a given point in the field extension.
func EvalFextPolyHorner(poly []fext.E3, x fext.E3) fext.E3 {
	res := fext.E3{}
	for i := len(poly) - 1; i >= 0; i-- {
		res.Mul(&res, &x)
		res.Add(&res, &poly[i])
	}
	return res
}

// EvalBasePolyHorner evaluates a polynomial in coefficient basis over the field
// extension at a given point in the field extension.
func EvalBasePolyHorner(poly []mamabear.Element, x fext.E3) fext.E3 {
	res := fext.E3{}
	for i := len(poly) - 1; i >= 0; i-- {
		res.Mul(&res, &x)
		res.A0.Add(&res.A0, &poly[i])
	}
	return res
}
