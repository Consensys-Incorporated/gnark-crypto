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

	if p.ReedSolomonInvRate == 2 {
		// In-place specialization of the coset decomposition below: with a
		// single extra coset the interpolant's evaluations can be produced in
		// res[:NbColumns] and scattered in descending order without a scratch
		// buffer, since the target index 2j+1 is always above the source j.
		copy(res, input)
		inputCoeffs := mamabear.Vector(res[:p.NbColumns])

		p.Domains[0].FFTInverse(inputCoeffs, fft.DIF, fft.WithNbTasks(1))
		inputCoeffs.Mul(inputCoeffs, p.CosetTableBitReverse)

		p.Domains[0].FFT(inputCoeffs, fft.DIT, fft.WithNbTasks(1))
		for j := p.NbColumns - 1; j >= 0; j-- {
			res[2*j+1] = res[j]
			res[2*j] = input[j]
		}
		return
	}

	p.encodeReedSolomonCosets(input, res)
}

// encodeReedSolomonCosets encodes by evaluating the interpolant on each coset of
// the codeword domain separately, rather than running one zero-padded FFT over
// the whole domain.
//
// With shift a primitive (rho*n)-th root of unity and w = shift^rho a primitive
// n-th root, the codeword domain <shift> is the disjoint union of the rho cosets
// shift^q * <w>. Codeword index rho*j+q is the evaluation at shift^(rho*j+q) =
// w^j * shift^q, so each coset is an ordinary size-n FFT of the interpolant
// pre-scaled by shift^(q*j), and coset 0 is the input itself.
//
// This is the chirp-block decomposition of MamaBearZKP §6.2. It replaces one
// size-(rho*n) FFT with rho-1 size-n FFTs plus (rho-1)*n scalings, skipping the
// first log2(rho) FFT layers, which on zero-padded input only copy coefficients
// into more branches and accumulate deterministic twiddle phases.
func (p *Params) encodeReedSolomonCosets(input, res []mamabear.Element) {
	n, rho := p.NbColumns, p.ReedSolomonInvRate

	// coeffs holds the interpolant in bit-reversed order, rescaled in place by
	// shift^j on each pass so that pass q sees shift^(q*j); evals receives the
	// FFT, which consumes it destructively.
	scratch := make([]mamabear.Element, 2*n)
	coeffs := mamabear.Vector(scratch[:n])
	evals := scratch[n:]

	copy(coeffs, input)
	p.Domains[0].FFTInverse(coeffs, fft.DIF, fft.WithNbTasks(1))

	// Coset 0 is <w> itself, on which the input already is the evaluations.
	for j := range n {
		res[rho*j] = input[j]
	}

	for q := 1; q < rho; q++ {
		coeffs.Mul(coeffs, p.CosetTableBitReverse)
		copy(evals, coeffs)
		p.Domains[0].FFT(evals, fft.DIT, fft.WithNbTasks(1))
		for j := range n {
			res[rho*j+q] = evals[j]
		}
	}
}

// IsReedSolomonCodewords returns true iff the argument is a correct codeword.
func (p *Params) IsReedSolomonCodewords(codeword []fext.E3) bool {
	if len(codeword) != p.SizeCodeWord() {
		return false
	}
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
