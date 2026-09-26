package vortex

import (
	"fmt"

	"github.com/consensys/gnark-crypto/field/koalabear"
	fext "github.com/consensys/gnark-crypto/field/koalabear/extensions"
	"github.com/consensys/gnark-crypto/field/koalabear/fft"
	"github.com/consensys/gnark-crypto/utils"
)

// EncodeReedSolomon encodes a vector of field elements into a reed-solomon codewords.
// The function checks that:
//   - the input argument has the right size
func (p *Params) EncodeReedSolomon(input, res []koalabear.Element) {
	if len(input) != p.NbColumns {
		panic(fmt.Sprintf("expected %d input values, got %d", p.NbColumns, len(input)))
	}

	copy(res, input)

	const rho = 2
	if rho != p.ReedSolomonInvRate {
		// slow path
		p.Domains[0].FFTInverse(res[:p.NbColumns], fft.DIF, fft.WithNbTasks(1))
		utils.BitReverse(res[:p.NbColumns])
		p.Domains[1].FFT(res, fft.DIF, fft.WithNbTasks(1))
		utils.BitReverse(res)
		return
	}

	// fast path; we avoid the bit reverse operations and work on the smaller domain.
	inputCoeffs := koalabear.Vector(res[:p.NbColumns])

	p.Domains[0].FFTInverse(inputCoeffs, fft.DIF, fft.WithNbTasks(1))
	inputCoeffs.Mul(inputCoeffs, p.CosetTableBitReverse)

	p.Domains[0].FFT(inputCoeffs, fft.DIT, fft.WithNbTasks(1))
	for j := p.NbColumns - 1; j >= 0; j-- {
		res[rho*j+1] = res[j]
		res[rho*j] = input[j]
	}
}

// IsCodeword returns nil iff the argument `v` is a correct codeword and an
// error is returned otherwise.
func (p *Params) IsReedSolomonCodewords(codeword []fext.E4) bool {
	if len(codeword) != p.SizeCodeWord() {
		return false
	}

	// As we don't have a dedicated FFT for field extensions, we apply
	// the FFT algorithm coordinates-by-coordinates. This might be
	// improvable by a direct AVX implementation but this only matters
	// for the verifier and not for the prover.

	coeffs := make([]koalabear.Element, p.SizeCodeWord())

	for i := range coeffs {
		coeffs[i] = codeword[i].B0.A0
	}

	p.Domains[1].FFTInverse(coeffs, fft.DIF)
	utils.BitReverse(coeffs)
	for i := p.NbColumns; i < p.SizeCodeWord(); i++ {
		if !coeffs[i].IsZero() {
			return false
		}
	}

	for i := range coeffs {
		coeffs[i] = codeword[i].B0.A1
	}

	p.Domains[1].FFTInverse(coeffs, fft.DIF)
	utils.BitReverse(coeffs)
	for i := p.NbColumns; i < p.SizeCodeWord(); i++ {
		if !coeffs[i].IsZero() {
			return false
		}
	}

	for i := range coeffs {
		coeffs[i] = codeword[i].B1.A0
	}

	p.Domains[1].FFTInverse(coeffs, fft.DIF)
	utils.BitReverse(coeffs)
	for i := p.NbColumns; i < p.SizeCodeWord(); i++ {
		if !coeffs[i].IsZero() {
			return false
		}
	}

	for i := range coeffs {
		coeffs[i] = codeword[i].B1.A1
	}

	p.Domains[1].FFTInverse(coeffs, fft.DIF)
	utils.BitReverse(coeffs)
	for i := p.NbColumns; i < p.SizeCodeWord(); i++ {
		if !coeffs[i].IsZero() {
			return false
		}
	}

	return true
}
