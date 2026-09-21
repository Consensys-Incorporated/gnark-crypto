package vortex

import (
	"fmt"

	"github.com/consensys/gnark-crypto/field/mamabear"
	fext "github.com/consensys/gnark-crypto/field/mamabear/extensions"
	"github.com/consensys/gnark-crypto/field/mamabear/fft"
	"github.com/consensys/gnark-crypto/utils"
)

// EncodeReedSolomon encodes a vector of field elements into a reed-solomon codeword.
func (p *Params) EncodeReedSolomon(input, res []mamabear.Element) {
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

	// fast path for rho=2
	inputCoeffs := mamabear.Vector(res[:p.NbColumns])

	p.Domains[0].FFTInverse(inputCoeffs, fft.DIF, fft.WithNbTasks(1))
	inputCoeffs.Mul(inputCoeffs, p.CosetTableBitReverse)

	p.Domains[0].FFT(inputCoeffs, fft.DIT, fft.WithNbTasks(1))
	for j := p.NbColumns - 1; j >= 0; j-- {
		res[rho*j+1] = res[j]
		res[rho*j] = input[j]
	}
}

// IsReedSolomonCodewords returns true iff the argument is a correct codeword.
func (p *Params) IsReedSolomonCodewords(codeword []fext.E3) bool {

	coeffs := make([]mamabear.Element, p.SizeCodeWord())

	for i := range coeffs {
		coeffs[i] = codeword[i].A0
	}

	p.Domains[1].FFTInverse(coeffs, fft.DIF)
	utils.BitReverse(coeffs)
	for i := p.NbColumns; i < p.SizeCodeWord(); i++ {
		if !coeffs[i].IsZero() {
			return false
		}
	}

	for i := range coeffs {
		coeffs[i] = codeword[i].A1
	}

	p.Domains[1].FFTInverse(coeffs, fft.DIF)
	utils.BitReverse(coeffs)
	for i := p.NbColumns; i < p.SizeCodeWord(); i++ {
		if !coeffs[i].IsZero() {
			return false
		}
	}

	for i := range coeffs {
		coeffs[i] = codeword[i].A2
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
