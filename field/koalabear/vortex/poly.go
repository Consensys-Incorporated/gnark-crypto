package vortex

import (
	"fmt"
	"math/big"

	"github.com/consensys/gnark-crypto/field/koalabear"
	fext "github.com/consensys/gnark-crypto/field/koalabear/extensions"
	"github.com/consensys/gnark-crypto/field/koalabear/fft"
)

// EvalBasePolyLagrange evaluates a polynomial in Lagrange basis over the base field
// at a given point in the field extension basis.
func EvalBasePolyLagrange(poly []koalabear.Element, x fext.E4) (fext.E4, error) {

	if !isPowerOfTwo(len(poly)) {
		return fext.E4{}, fmt.Errorf("only support powers of two but poly has length %v", len(poly))
	}

	var (
		n            = len(poly)
		denominators = make([]fext.E4, n)
		one          = koalabear.One()
		generator, _ = fft.Generator(uint64(n))
		generatorInv = new(koalabear.Element).Inverse(&generator)
		cardInv      fext.E4
	)

	cardInv.B0.A0 = koalabear.NewElement(uint64(n))
	cardInv.Inverse(&cardInv)

	// The denominator is constructed as:
	// 		D_x = \frac{X}{x} - g for x \in H
	// 	where H is the subgroup of the roots of unity (not the coset)
	// 	and g a field element such that gH is the coset.
	denominators[0] = x
	for i := 1; i < n; i++ {
		denominators[i].MulByElement(&denominators[i-1], generatorInv)
	}

	for i := range n {
		// This subtracts a field extension by a base field element.
		denominators[i].B0.A0.Sub(&denominators[i].B0.A0, &one)
		if denominators[i].IsZero() {
			// edge-case : x is a root of unity of the domain. In this case, we can just return
			// the associated value for poly
			res := fext.E4{}
			res.B0.A0.Set(&poly[i])
			return res, nil
		}
	}

	denominators = fext.VectorE4(denominators).BatchInvert()
	res, tmp := fext.E4{}, fext.E4{}
	for i := range denominators {
		tmp.MulByElement(&denominators[i], &poly[i])
		res.Add(&res, &tmp)
	}

	tmp.Exp(x, big.NewInt(int64(n)))
	tmp.B0.A0.Sub(&tmp.B0.A0, &one)
	tmp.Mul(&tmp, &cardInv)
	res.Mul(&res, &tmp)

	return res, nil
}

// EvalFextPolyLagrange evaluates a polynomial in Lagrange basis over the field extension
// at a given point in the field extension.
func EvalFextPolyLagrange(poly []fext.E4, x fext.E4) (fext.E4, error) {

	if !isPowerOfTwo(len(poly)) {
		return fext.E4{}, fmt.Errorf("only support powers of two but poly has length %v", len(poly))
	}

	var (
		n            = len(poly)
		denominators = make([]fext.E4, n)
		one          = koalabear.One()
		generator, _ = fft.Generator(uint64(n))
		generatorInv = new(koalabear.Element).Inverse(&generator)
		cardInv      fext.E4
	)

	cardInv.B0.A0 = koalabear.NewElement(uint64(n))
	cardInv.Inverse(&cardInv)

	// The denominator is constructed as:
	// 		D_x = \frac{X}{x} - g for x \in H
	// 	where H is the subgroup of the roots of unity (not the coset)
	// 	and g a field element such that gH is the coset.
	denominators[0] = x
	for i := 1; i < n; i++ {
		denominators[i].MulByElement(&denominators[i-1], generatorInv)
	}

	for i := range n {
		// This subtracts a field extension by a base field element.
		denominators[i].B0.A0.Sub(&denominators[i].B0.A0, &one)
		if denominators[i].IsZero() {
			// edge-case : x is a root of unity of the domain. In this case, we can just return
			// the associated value for poly
			return poly[i], nil
		}
	}

	denominators = fext.VectorE4(denominators).BatchInvert()
	res, tmp := fext.E4{}, fext.E4{}
	for i := range denominators {
		tmp.Mul(&denominators[i], &poly[i])
		res.Add(&res, &tmp)
	}

	tmp.Exp(x, big.NewInt(int64(n)))
	tmp.B0.A0.Sub(&tmp.B0.A0, &one)
	tmp.Mul(&tmp, &cardInv)
	res.Mul(&res, &tmp)

	return res, nil
}

// EvalFextPolyHorner evaluates a polynomial in coefficient basis over the field
// extension at a given point in the field extension.
func EvalFextPolyHorner(poly []fext.E4, x fext.E4) fext.E4 {
	res := fext.E4{}
	for i := len(poly) - 1; i >= 0; i-- {
		res.Mul(&res, &x)
		res.Add(&res, &poly[i])
	}
	return res
}

// EvalBasePolyHorner evaluates a polynomial in coefficient basis over the field
// extension at a given point in the field extension.
func EvalBasePolyHorner(poly []koalabear.Element, x fext.E4) fext.E4 {
	res := fext.E4{}
	for i := len(poly) - 1; i >= 0; i-- {
		res.Mul(&res, &x)
		res.B0.A0.Add(&res.B0.A0, &poly[i])
	}
	return res
}
