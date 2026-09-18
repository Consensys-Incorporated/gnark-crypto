// Copyright 2020-2026 Consensys Software Inc.
// Licensed under the Apache License, Version 2.0. See the LICENSE file for details.

package fft

import (
	"math/big"
	"math/bits"

	"github.com/consensys/gnark-crypto/ecc"
	"github.com/consensys/gnark-crypto/parallel"
	"github.com/consensys/gnark-crypto/utils"

	"github.com/consensys/gnark-crypto/field/mamabear"
	fext "github.com/consensys/gnark-crypto/field/mamabear/extensions"
)

// FFTExt computes the discrete Fourier transform of a slice of E3 extension field elements.
// Coefficients and evaluations are E3 elements; twiddle factors are base-field elements
// drawn from the same domain as FFT.
func (domain *Domain) FFTExt(a []fext.E3, decimation Decimation, opts ...Option) {
	opt := fftOptions(opts)

	maxSplits := bits.TrailingZeros64(ecc.NextPowerOfTwo(uint64(opt.nbTasks)))
	if opt.nbTasks == 1 {
		maxSplits = -1
	}

	if opt.coset {
		var cosetTable []mamabear.Element
		if decimation == DIT {
			if domain.cosetTableBitReversed != nil {
				cosetTable = domain.cosetTableBitReversed
			} else {
				cosetTable = make([]mamabear.Element, len(a))
				if domain.cosetTable != nil {
					copy(cosetTable, domain.cosetTable)
				} else {
					BuildExpTable(domain.FrMultiplicativeGen, cosetTable)
				}
				utils.BitReverse(cosetTable)
			}
		} else {
			if domain.cosetTable != nil {
				cosetTable = domain.cosetTable
			} else {
				cosetTable = make([]mamabear.Element, len(a))
				BuildExpTable(domain.FrMultiplicativeGen, cosetTable)
			}
		}
		parallel.ExecuteAligned(len(a), 1, func(start, end int) {
			va := fext.Vector(a[start:end])
			va.MulByElement(va, cosetTable[start:end])
		}, opt.nbTasks)
	}

	twiddles := domain.twiddles
	twiddlesStartStage := 0
	if !domain.withPrecompute {
		twiddlesStartStage = 3
		nbStages := int(bits.TrailingZeros64(domain.Cardinality))
		if nbStages-twiddlesStartStage > 0 {
			twiddles = make([][]mamabear.Element, nbStages-twiddlesStartStage)
			w := domain.Generator
			w.Exp(w, big.NewInt(int64(1<<twiddlesStartStage)))
			buildTwiddles(twiddles, w, uint64(nbStages-twiddlesStartStage))
		}
	}

	switch decimation {
	case DIF:
		difFFTExt(a, domain.Generator, twiddles, twiddlesStartStage, 0, maxSplits, nil, opt.nbTasks)
	case DIT:
		ditFFTExt(a, domain.Generator, twiddles, twiddlesStartStage, 0, maxSplits, nil, opt.nbTasks)
	default:
		panic("not implemented")
	}
}

// FFTInverseExt computes the inverse discrete Fourier transform of a slice of E3 elements.
// if decimation == DIT, the input must be in bit-reversed order.
// if decimation == DIF, the output will be in bit-reversed order.
func (domain *Domain) FFTInverseExt(a []fext.E3, decimation Decimation, opts ...Option) {
	opt := fftOptions(opts)

	maxSplits := bits.TrailingZeros64(ecc.NextPowerOfTwo(uint64(opt.nbTasks)))
	if opt.nbTasks == 1 {
		maxSplits = -1
	}

	twiddlesInv := domain.twiddlesInv
	twiddlesStartStage := 0
	if !domain.withPrecompute {
		twiddlesStartStage = 3
		nbStages := int(bits.TrailingZeros64(domain.Cardinality))
		if nbStages-twiddlesStartStage > 0 {
			twiddlesInv = make([][]mamabear.Element, nbStages-twiddlesStartStage)
			w := domain.GeneratorInv
			w.Exp(w, big.NewInt(int64(1<<twiddlesStartStage)))
			buildTwiddles(twiddlesInv, w, uint64(nbStages-twiddlesStartStage))
		}
	}

	switch decimation {
	case DIF:
		difFFTExt(a, domain.GeneratorInv, twiddlesInv, twiddlesStartStage, 0, maxSplits, nil, opt.nbTasks)
	case DIT:
		ditFFTExt(a, domain.GeneratorInv, twiddlesInv, twiddlesStartStage, 0, maxSplits, nil, opt.nbTasks)
	default:
		panic("not implemented")
	}

	// scale by CardinalityInv
	if !opt.coset {
		parallel.ExecuteAligned(len(a), 1, func(start, end int) {
			va := fext.Vector(a[start:end])
			va.ScalarMulByElement(va, &domain.CardinalityInv)
		}, opt.nbTasks)
		return
	}

	var cosetTableInv []mamabear.Element
	if decimation == DIT {
		if domain.cosetTableInv != nil {
			cosetTableInv = domain.cosetTableInv
		} else {
			cosetTableInv = make([]mamabear.Element, len(a))
			BuildExpTable(domain.FrMultiplicativeGenInv, cosetTableInv)
		}
	} else {
		if domain.cosetTableInvBitReversed != nil {
			cosetTableInv = domain.cosetTableInvBitReversed
		} else {
			cosetTableInv = make([]mamabear.Element, len(a))
			if domain.cosetTableInv != nil {
				copy(cosetTableInv, domain.cosetTableInv)
			} else {
				BuildExpTable(domain.FrMultiplicativeGenInv, cosetTableInv)
			}
			utils.BitReverse(cosetTableInv)
		}
	}
	parallel.ExecuteAligned(len(a), 1, func(start, end int) {
		va := fext.Vector(a[start:end])
		va.MulByElement(va, cosetTableInv[start:end])
		va.ScalarMulByElement(va, &domain.CardinalityInv)
	}, opt.nbTasks)
}

func difFFTExt(a []fext.E3, w mamabear.Element, twiddles [][]mamabear.Element, twiddlesStartStage, stage, maxSplits int, chDone chan struct{}, nbTasks int) {
	if chDone != nil {
		defer close(chDone)
	}

	n := len(a)
	if n == 1 {
		return
	} else if stage >= twiddlesStartStage {
		if n == 1<<9 {
			kerDIFNP_512Ext(a, twiddles, stage-twiddlesStartStage)
			return
		}
	}
	m := n >> 1

	parallelButterfly := (m > butterflyThreshold) && (stage < maxSplits)

	if stage < twiddlesStartStage {
		if parallelButterfly {
			w := w
			parallel.Execute(m, func(start, end int) {
				if start == 0 {
					fext.Butterfly(&a[0], &a[m])
					start++
				}
				var at mamabear.Element
				at.Exp(w, big.NewInt(int64(start)))
				innerDIFWithoutTwiddlesExt(a, at, w, start, end, m)
			}, nbTasks/(1<<(stage)))
		} else {
			innerDIFWithoutTwiddlesExt(a, w, w, 0, m, m)
		}
		w.Square(&w)
	} else {
		if parallelButterfly {
			parallel.Execute(m, func(start, end int) {
				innerDIFWithTwiddlesExt(a, twiddles[stage-twiddlesStartStage], start, end, m)
			}, nbTasks/(1<<(stage)))
		} else {
			innerDIFWithTwiddlesExt(a, twiddles[stage-twiddlesStartStage], 0, m, m)
		}
	}

	if m == 1 {
		return
	}

	nextStage := stage + 1
	if stage < maxSplits {
		chDone := make(chan struct{}, 1)
		go difFFTExt(a[m:n], w, twiddles, twiddlesStartStage, nextStage, maxSplits, chDone, nbTasks)
		difFFTExt(a[0:m], w, twiddles, twiddlesStartStage, nextStage, maxSplits, nil, nbTasks)
		<-chDone
	} else {
		difFFTExt(a[0:m], w, twiddles, twiddlesStartStage, nextStage, maxSplits, nil, nbTasks)
		difFFTExt(a[m:n], w, twiddles, twiddlesStartStage, nextStage, maxSplits, nil, nbTasks)
	}
}

func innerDIFWithTwiddlesExt(a []fext.E3, twiddles []mamabear.Element, start, end, m int) {
	va0 := fext.Vector(a[start:end])
	va1 := fext.Vector(a[start+m : end+m])
	va0.Butterfly(va1)
	va1.MulByElement(va1, twiddles[start:end])
}

func innerDIFWithoutTwiddlesExt(a []fext.E3, at, w mamabear.Element, start, end, m int) {
	if start == 0 {
		fext.Butterfly(&a[0], &a[m])
		start++
	}
	for i := start; i < end; i++ {
		fext.Butterfly(&a[i], &a[i+m])
		a[i+m].MulByElement(&a[i+m], &at)
		at.Mul(&at, &w)
	}
}

func ditFFTExt(a []fext.E3, w mamabear.Element, twiddles [][]mamabear.Element, twiddlesStartStage, stage, maxSplits int, chDone chan struct{}, nbTasks int) {
	if chDone != nil {
		defer close(chDone)
	}
	n := len(a)
	if n == 1 {
		return
	} else if stage >= twiddlesStartStage {
		if n == 1<<9 {
			kerDITNP_512Ext(a, twiddles, stage-twiddlesStartStage)
			return
		}
	}

	m := n >> 1

	nextStage := stage + 1
	nextW := w
	nextW.Square(&nextW)

	if stage < maxSplits {
		chDone := make(chan struct{}, 1)
		go ditFFTExt(a[m:], nextW, twiddles, twiddlesStartStage, nextStage, maxSplits, chDone, nbTasks)
		ditFFTExt(a[0:m], nextW, twiddles, twiddlesStartStage, nextStage, maxSplits, nil, nbTasks)
		<-chDone
	} else {
		ditFFTExt(a[0:m], nextW, twiddles, twiddlesStartStage, nextStage, maxSplits, nil, nbTasks)
		ditFFTExt(a[m:n], nextW, twiddles, twiddlesStartStage, nextStage, maxSplits, nil, nbTasks)
	}

	parallelButterfly := (m > butterflyThreshold) && (stage < maxSplits)

	if stage < twiddlesStartStage {
		if parallelButterfly {
			w := w
			parallel.Execute(m, func(start, end int) {
				if start == 0 {
					fext.Butterfly(&a[0], &a[m])
					start++
				}
				var at mamabear.Element
				at.Exp(w, big.NewInt(int64(start)))
				innerDITWithoutTwiddlesExt(a, at, w, start, end, m)
			}, nbTasks/(1<<(stage)))
		} else {
			innerDITWithoutTwiddlesExt(a, w, w, 0, m, m)
		}
		return
	}
	if parallelButterfly {
		parallel.Execute(m, func(start, end int) {
			innerDITWithTwiddlesExt(a, twiddles[stage-twiddlesStartStage], start, end, m)
		}, nbTasks/(1<<(stage)))
	} else {
		innerDITWithTwiddlesExt(a, twiddles[stage-twiddlesStartStage], 0, m, m)
	}
}

func innerDITWithTwiddlesExt(a []fext.E3, twiddles []mamabear.Element, start, end, m int) {
	va0 := fext.Vector(a[start:end])
	va1 := fext.Vector(a[start+m : end+m])
	va1.MulByElement(va1, twiddles[start:end])
	va0.Butterfly(va1)
}

func innerDITWithoutTwiddlesExt(a []fext.E3, at, w mamabear.Element, start, end, m int) {
	if start == 0 {
		fext.Butterfly(&a[0], &a[m])
		start++
	}
	for i := start; i < end; i++ {
		a[i+m].MulByElement(&a[i+m], &at)
		fext.Butterfly(&a[i], &a[i+m])
		at.Mul(&at, &w)
	}
}

// kerDIFNP_512Ext is an unrolled 512-element DIF kernel for E3 elements.
func kerDIFNP_512Ext(a []fext.E3, twiddles [][]mamabear.Element, stage int) {
	innerDIFWithTwiddlesExt(a[:512], twiddles[stage+0], 0, 256, 256)
	for offset := 0; offset < 512; offset += 256 {
		innerDIFWithTwiddlesExt(a[offset:offset+256], twiddles[stage+1], 0, 128, 128)
	}
	for offset := 0; offset < 512; offset += 128 {
		innerDIFWithTwiddlesExt(a[offset:offset+128], twiddles[stage+2], 0, 64, 64)
	}
	for offset := 0; offset < 512; offset += 64 {
		innerDIFWithTwiddlesExt(a[offset:offset+64], twiddles[stage+3], 0, 32, 32)
	}
	for offset := 0; offset < 512; offset += 32 {
		innerDIFWithTwiddlesExt(a[offset:offset+32], twiddles[stage+4], 0, 16, 16)
	}
	for offset := 0; offset < 512; offset += 16 {
		innerDIFWithTwiddlesExt(a[offset:offset+16], twiddles[stage+5], 0, 8, 8)
	}
	for offset := 0; offset < 512; offset += 8 {
		innerDIFWithTwiddlesExt(a[offset:offset+8], twiddles[stage+6], 0, 4, 4)
	}
	for offset := 0; offset < 512; offset += 4 {
		innerDIFWithTwiddlesExt(a[offset:offset+4], twiddles[stage+7], 0, 2, 2)
	}
	fext.Vector(a[:512]).ButterflyPair()
}

// kerDITNP_512Ext is the DIT counterpart of kerDIFNP_512Ext.
func kerDITNP_512Ext(a []fext.E3, twiddles [][]mamabear.Element, stage int) {
	fext.Vector(a[:512]).ButterflyPair()
	for offset := 0; offset < 512; offset += 4 {
		innerDITWithTwiddlesExt(a[offset:offset+4], twiddles[stage+7], 0, 2, 2)
	}
	for offset := 0; offset < 512; offset += 8 {
		innerDITWithTwiddlesExt(a[offset:offset+8], twiddles[stage+6], 0, 4, 4)
	}
	for offset := 0; offset < 512; offset += 16 {
		innerDITWithTwiddlesExt(a[offset:offset+16], twiddles[stage+5], 0, 8, 8)
	}
	for offset := 0; offset < 512; offset += 32 {
		innerDITWithTwiddlesExt(a[offset:offset+32], twiddles[stage+4], 0, 16, 16)
	}
	for offset := 0; offset < 512; offset += 64 {
		innerDITWithTwiddlesExt(a[offset:offset+64], twiddles[stage+3], 0, 32, 32)
	}
	for offset := 0; offset < 512; offset += 128 {
		innerDITWithTwiddlesExt(a[offset:offset+128], twiddles[stage+2], 0, 64, 64)
	}
	for offset := 0; offset < 512; offset += 256 {
		innerDITWithTwiddlesExt(a[offset:offset+256], twiddles[stage+1], 0, 128, 128)
	}
	innerDITWithTwiddlesExt(a[:512], twiddles[stage+0], 0, 256, 256)
}
