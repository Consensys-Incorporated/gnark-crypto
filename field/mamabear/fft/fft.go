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
)

// Decimation is used in the FFT call to select decimation in time or in frequency
type Decimation uint8

const (
	DIT Decimation = iota
	DIF
)

// parallelize threshold for a single butterfly op, if the fft stage is not parallelized already
const butterflyThreshold = 16

// FFT computes (recursively) the discrete Fourier transform of a and stores the result in a
// if decimation == DIT (decimation in time), the input must be in bit-reversed order
// if decimation == DIF (decimation in frequency), the output will be in bit-reversed order
func (domain *Domain) FFT(a []mamabear.Element, decimation Decimation, opts ...Option) {
	opt := fftOptions(opts)

	maxSplits := bits.TrailingZeros64(ecc.NextPowerOfTwo(uint64(opt.nbTasks)))
	if opt.nbTasks == 1 {
		maxSplits = -1
	}

	// if coset != 0, scale by coset table
	if opt.coset {
		var cosetTable []mamabear.Element
		if decimation == DIT {
			// DIT needs bit-reversed coset table
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
			// DIF needs natural-order coset table
			if domain.cosetTable != nil {
				cosetTable = domain.cosetTable
			} else {
				cosetTable = make([]mamabear.Element, len(a))
				BuildExpTable(domain.FrMultiplicativeGen, cosetTable)
			}
		}
		if opt.nbTasks <= 1 {
			v1 := mamabear.Vector(a)
			v2 := mamabear.Vector(cosetTable[:len(a)])
			v1.Mul(v1, v2)
		} else {
			parallel.ExecuteAligned(len(a), 16, func(start, end int) {
				v1 := mamabear.Vector(a[start:end])
				v2 := mamabear.Vector(cosetTable[start:end])
				v1.Mul(v1, v2)
			}, opt.nbTasks)
		}
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
		difFFT(a, domain.Generator, twiddles, twiddlesStartStage, 0, maxSplits, nil, opt.nbTasks)
	case DIT:
		ditFFT(a, domain.Generator, twiddles, twiddlesStartStage, 0, maxSplits, nil, opt.nbTasks)
	default:
		panic("not implemented")
	}
}

// FFTInverse computes (recursively) the inverse discrete Fourier transform of a and stores the result in a
// if decimation == DIT (decimation in time), the input must be in bit-reversed order
// if decimation == DIF (decimation in frequency), the output will be in bit-reversed order
// len(a) must be a power of 2, and w must be a len(a)th root of unity in field F.
func (domain *Domain) FFTInverse(a []mamabear.Element, decimation Decimation, opts ...Option) {
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
		difFFT(a, domain.GeneratorInv, twiddlesInv, twiddlesStartStage, 0, maxSplits, nil, opt.nbTasks)
	case DIT:
		ditFFT(a, domain.GeneratorInv, twiddlesInv, twiddlesStartStage, 0, maxSplits, nil, opt.nbTasks)
	default:
		panic("not implemented")
	}

	// scale by CardinalityInv
	if !opt.coset {
		if opt.nbTasks <= 1 {
			v := mamabear.Vector(a)
			v.ScalarMul(v, &domain.CardinalityInv)
		} else {
			parallel.ExecuteAligned(len(a), 16, func(start, end int) {
				v := mamabear.Vector(a[start:end])
				v.ScalarMul(v, &domain.CardinalityInv)
			}, opt.nbTasks)
		}
		return
	}
	var cosetTableInv []mamabear.Element
	if decimation == DIT {
		// DIT inverse needs natural-order inverse coset table
		if domain.cosetTableInv != nil {
			cosetTableInv = domain.cosetTableInv
		} else {
			cosetTableInv = make([]mamabear.Element, len(a))
			BuildExpTable(domain.FrMultiplicativeGenInv, cosetTableInv)
		}
	} else {
		// DIF inverse needs bit-reversed inverse coset table
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
	if opt.nbTasks <= 1 {
		v := mamabear.Vector(a)
		v.Mul(v, mamabear.Vector(cosetTableInv[:len(a)]))
		v.ScalarMul(v, &domain.CardinalityInv)
	} else {
		parallel.ExecuteAligned(len(a), 16, func(start, end int) {
			v := mamabear.Vector(a[start:end])
			v.Mul(v, mamabear.Vector(cosetTableInv[start:end]))
			v.ScalarMul(v, &domain.CardinalityInv)
		}, opt.nbTasks)
	}
}

func difFFT(a []mamabear.Element, w mamabear.Element, twiddles [][]mamabear.Element, twiddlesStartStage, stage, maxSplits int, chDone chan struct{}, nbTasks int) {
	if chDone != nil {
		defer close(chDone)
	}

	n := len(a)
	if n == 1 {
		return
	} else if stage >= twiddlesStartStage {
		switch n {
		case 64:
			kerDIFNP_64(a, twiddles, stage-twiddlesStartStage)
			return
		case 128:
			kerDIFNP_128(a, twiddles, stage-twiddlesStartStage)
			return
		case 1 << 8:
			kerDIFNP_256(a, twiddles, stage-twiddlesStartStage)
			return
		case 512:
			kerDIFNP_512(a, twiddles, stage-twiddlesStartStage)
			return
		case 1024:
			kerDIFNP_1024(a, twiddles, stage-twiddlesStartStage)
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
					butterfly(&a[0], &a[m])
					start++
				}
				var at mamabear.Element
				at.Exp(w, big.NewInt(int64(start)))
				innerDIFWithoutTwiddles(a, at, w, start, end, m)
			}, nbTasks/(1<<(stage)))
		} else {
			innerDIFWithoutTwiddles(a, w, w, 0, m, m)
		}
		w.Square(&w)
	} else {
		innerDIFWithTwiddles(a, twiddles[stage-twiddlesStartStage], 0, m, m)
	}

	if m == 1 {
		return
	}

	nextStage := stage + 1
	if stage < maxSplits {
		chDone := make(chan struct{}, 1)
		go difFFT(a[m:n], w, twiddles, twiddlesStartStage, nextStage, maxSplits, chDone, nbTasks)
		difFFT(a[0:m], w, twiddles, twiddlesStartStage, nextStage, maxSplits, nil, nbTasks)
		<-chDone
	} else {
		difFFT(a[0:m], w, twiddles, twiddlesStartStage, nextStage, maxSplits, nil, nbTasks)
		difFFT(a[m:n], w, twiddles, twiddlesStartStage, nextStage, maxSplits, nil, nbTasks)
	}
}

func innerDIFWithTwiddlesGeneric(a []mamabear.Element, twiddles []mamabear.Element, start, end, m int) {
	if start == 0 {
		butterfly(&a[0], &a[m])
		start++
	}
	for i := start; i < end; i++ {
		butterfly(&a[i], &a[i+m])
	}
	v1 := mamabear.Vector(a[start+m : end+m])
	v2 := mamabear.Vector(twiddles[start:end])
	v1.Mul(v1, v2)
}

func innerDIFWithoutTwiddles(a []mamabear.Element, at, w mamabear.Element, start, end, m int) {
	if start == 0 {
		butterfly(&a[0], &a[m])
		start++
	}
	for i := start; i < end; i++ {
		butterfly(&a[i], &a[i+m])
		a[i+m].Mul(&a[i+m], &at)
		at.Mul(&at, &w)
	}
}

func ditFFT(a []mamabear.Element, w mamabear.Element, twiddles [][]mamabear.Element, twiddlesStartStage, stage, maxSplits int, chDone chan struct{}, nbTasks int) {
	if chDone != nil {
		defer close(chDone)
	}
	n := len(a)
	if n == 1 {
		return
	} else if stage >= twiddlesStartStage {
		switch n {
		case 64:
			kerDITNP_64(a, twiddles, stage-twiddlesStartStage)
			return
		case 128:
			kerDITNP_128(a, twiddles, stage-twiddlesStartStage)
			return
		case 1 << 8:
			kerDITNP_256(a, twiddles, stage-twiddlesStartStage)
			return
		case 512:
			kerDITNP_512(a, twiddles, stage-twiddlesStartStage)
			return
		case 1024:
			kerDITNP_1024(a, twiddles, stage-twiddlesStartStage)
			return
		}
	}

	m := n >> 1

	nextStage := stage + 1
	nextW := w
	nextW.Square(&nextW)

	if stage < maxSplits {
		chDone := make(chan struct{}, 1)
		go ditFFT(a[m:], nextW, twiddles, twiddlesStartStage, nextStage, maxSplits, chDone, nbTasks)
		ditFFT(a[0:m], nextW, twiddles, twiddlesStartStage, nextStage, maxSplits, nil, nbTasks)
		<-chDone
	} else {
		ditFFT(a[0:m], nextW, twiddles, twiddlesStartStage, nextStage, maxSplits, nil, nbTasks)
		ditFFT(a[m:n], nextW, twiddles, twiddlesStartStage, nextStage, maxSplits, nil, nbTasks)
	}

	parallelButterfly := (m > butterflyThreshold) && (stage < maxSplits)

	if stage < twiddlesStartStage {
		if parallelButterfly {
			w := w
			parallel.Execute(m, func(start, end int) {
				if start == 0 {
					butterfly(&a[0], &a[m])
					start++
				}
				var at mamabear.Element
				at.Exp(w, big.NewInt(int64(start)))
				innerDITWithoutTwiddles(a, at, w, start, end, m)
			}, nbTasks/(1<<(stage)))
		} else {
			innerDITWithoutTwiddles(a, w, w, 0, m, m)
		}
		return
	}
	innerDITWithTwiddles(a, twiddles[stage-twiddlesStartStage], 0, m, m)
}

func innerDITWithTwiddlesGeneric(a []mamabear.Element, twiddles []mamabear.Element, start, end, m int) {
	if start == 0 {
		butterfly(&a[0], &a[m])
		start++
	}
	v1 := mamabear.Vector(a[start+m : end+m])
	v2 := mamabear.Vector(twiddles[start:end])
	v1.Mul(v1, v2)
	for i := start; i < end; i++ {
		butterfly(&a[i], &a[i+m])
	}
}

func innerDITWithoutTwiddles(a []mamabear.Element, at, w mamabear.Element, start, end, m int) {
	if start == 0 {
		butterfly(&a[0], &a[m])
		start++
	}
	for i := start; i < end; i++ {
		a[i+m].Mul(&a[i+m], &at)
		butterfly(&a[i], &a[i+m])
		at.Mul(&at, &w)
	}
}

func kerDIFNP_256generic(a []mamabear.Element, twiddles [][]mamabear.Element, stage int) {
	innerDIFWithTwiddlesGeneric(a[:256], twiddles[stage+0], 0, 128, 128)
	for offset := 0; offset < 256; offset += 128 {
		innerDIFWithTwiddlesGeneric(a[offset:offset+128], twiddles[stage+1], 0, 64, 64)
	}
	for offset := 0; offset < 256; offset += 64 {
		innerDIFWithTwiddlesGeneric(a[offset:offset+64], twiddles[stage+2], 0, 32, 32)
	}
	for offset := 0; offset < 256; offset += 32 {
		innerDIFWithTwiddlesGeneric(a[offset:offset+32], twiddles[stage+3], 0, 16, 16)
	}
	for offset := 0; offset < 256; offset += 16 {
		innerDIFWithTwiddlesGeneric(a[offset:offset+16], twiddles[stage+4], 0, 8, 8)
	}
	for offset := 0; offset < 256; offset += 8 {
		innerDIFWithTwiddlesGeneric(a[offset:offset+8], twiddles[stage+5], 0, 4, 4)
	}
	for offset := 0; offset < 256; offset += 4 {
		innerDIFWithTwiddlesGeneric(a[offset:offset+4], twiddles[stage+6], 0, 2, 2)
	}
	for offset := 0; offset < 256; offset += 2 {
		butterfly(&a[offset], &a[offset+1])
	}
}

func kerDITNP_256generic(a []mamabear.Element, twiddles [][]mamabear.Element, stage int) {
	for offset := 0; offset < 256; offset += 2 {
		butterfly(&a[offset], &a[offset+1])
	}
	for offset := 0; offset < 256; offset += 4 {
		innerDITWithTwiddlesGeneric(a[offset:offset+4], twiddles[stage+6], 0, 2, 2)
	}
	for offset := 0; offset < 256; offset += 8 {
		innerDITWithTwiddlesGeneric(a[offset:offset+8], twiddles[stage+5], 0, 4, 4)
	}
	for offset := 0; offset < 256; offset += 16 {
		innerDITWithTwiddlesGeneric(a[offset:offset+16], twiddles[stage+4], 0, 8, 8)
	}
	for offset := 0; offset < 256; offset += 32 {
		innerDITWithTwiddlesGeneric(a[offset:offset+32], twiddles[stage+3], 0, 16, 16)
	}
	for offset := 0; offset < 256; offset += 64 {
		innerDITWithTwiddlesGeneric(a[offset:offset+64], twiddles[stage+2], 0, 32, 32)
	}
	for offset := 0; offset < 256; offset += 128 {
		innerDITWithTwiddlesGeneric(a[offset:offset+128], twiddles[stage+1], 0, 64, 64)
	}
	innerDITWithTwiddlesGeneric(a[:256], twiddles[stage+0], 0, 128, 128)
}

// kerDIFNP_64 is an unrolled 64-element DIF kernel.
func kerDIFNP_64(a []mamabear.Element, twiddles [][]mamabear.Element, stage int) {
	innerDIFWithTwiddles(a[:64], twiddles[stage+0], 0, 32, 32)
	for offset := 0; offset < 64; offset += 32 {
		innerDIFWithTwiddles(a[offset:offset+32], twiddles[stage+1], 0, 16, 16)
	}
	{
		tw := twiddles[stage+2]
		for offset := 0; offset < 64; offset += 32 {
			o1, o2 := offset, offset+16
			butterfly(&a[o1], &a[o1+8])
			butterfly(&a[o2], &a[o2+8])
			for i := 1; i < 8; i++ {
				butterfly(&a[o1+i], &a[o1+i+8])
				butterfly(&a[o2+i], &a[o2+i+8])
			}
			for i := 1; i < 8; i++ {
				a[o1+i+8].Mul(&a[o1+i+8], &tw[i])
				a[o2+i+8].Mul(&a[o2+i+8], &tw[i])
			}
		}
	}
	{
		tw := twiddles[stage+3]
		for offset := 0; offset < 64; offset += 16 {
			o1, o2 := offset, offset+8
			butterfly(&a[o1], &a[o1+4])
			butterfly(&a[o2], &a[o2+4])
			for i := 1; i < 4; i++ {
				butterfly(&a[o1+i], &a[o1+i+4])
				butterfly(&a[o2+i], &a[o2+i+4])
			}
			for i := 1; i < 4; i++ {
				a[o1+i+4].Mul(&a[o1+i+4], &tw[i])
				a[o2+i+4].Mul(&a[o2+i+4], &tw[i])
			}
		}
	}
	{
		tw := twiddles[stage+4]
		for offset := 0; offset < 64; offset += 4 {
			butterfly(&a[offset], &a[offset+2])
			butterfly(&a[offset+1], &a[offset+3])
			a[offset+3].Mul(&a[offset+3], &tw[1])
		}
	}
	for offset := 0; offset < 64; offset += 2 {
		butterfly(&a[offset], &a[offset+1])
	}
}

// kerDITNP_64 is the DIT counterpart of kerDIFNP_64.
func kerDITNP_64(a []mamabear.Element, twiddles [][]mamabear.Element, stage int) {
	for offset := 0; offset < 64; offset += 2 {
		butterfly(&a[offset], &a[offset+1])
	}
	{
		tw := twiddles[stage+4]
		for offset := 0; offset < 64; offset += 4 {
			a[offset+3].Mul(&a[offset+3], &tw[1])
			butterfly(&a[offset], &a[offset+2])
			butterfly(&a[offset+1], &a[offset+3])
		}
	}
	{
		tw := twiddles[stage+3]
		for offset := 0; offset < 64; offset += 16 {
			o1, o2 := offset, offset+8
			for i := 1; i < 4; i++ {
				a[o1+i+4].Mul(&a[o1+i+4], &tw[i])
				a[o2+i+4].Mul(&a[o2+i+4], &tw[i])
			}
			butterfly(&a[o1], &a[o1+4])
			butterfly(&a[o2], &a[o2+4])
			for i := 1; i < 4; i++ {
				butterfly(&a[o1+i], &a[o1+i+4])
				butterfly(&a[o2+i], &a[o2+i+4])
			}
		}
	}
	{
		tw := twiddles[stage+2]
		for offset := 0; offset < 64; offset += 32 {
			o1, o2 := offset, offset+16
			for i := 1; i < 8; i++ {
				a[o1+i+8].Mul(&a[o1+i+8], &tw[i])
				a[o2+i+8].Mul(&a[o2+i+8], &tw[i])
			}
			butterfly(&a[o1], &a[o1+8])
			butterfly(&a[o2], &a[o2+8])
			for i := 1; i < 8; i++ {
				butterfly(&a[o1+i], &a[o1+i+8])
				butterfly(&a[o2+i], &a[o2+i+8])
			}
		}
	}
	for offset := 0; offset < 64; offset += 32 {
		innerDITWithTwiddles(a[offset:offset+32], twiddles[stage+1], 0, 16, 16)
	}
	innerDITWithTwiddles(a[:64], twiddles[stage+0], 0, 32, 32)
}

// kerDIFNP_128 is an unrolled 128-element DIF kernel.
func kerDIFNP_128(a []mamabear.Element, twiddles [][]mamabear.Element, stage int) {
	innerDIFWithTwiddles(a[:128], twiddles[stage+0], 0, 64, 64)
	kerDIFNP_64(a[:64], twiddles, stage+1)
	kerDIFNP_64(a[64:], twiddles, stage+1)
}

// kerDITNP_128 is the DIT counterpart of kerDIFNP_128.
func kerDITNP_128(a []mamabear.Element, twiddles [][]mamabear.Element, stage int) {
	kerDITNP_64(a[:64], twiddles, stage+1)
	kerDITNP_64(a[64:], twiddles, stage+1)
	innerDITWithTwiddles(a[:128], twiddles[stage+0], 0, 64, 64)
}

// kerDIFNP_512 is an optimized 512-element DIF kernel.
func kerDIFNP_512(a []mamabear.Element, twiddles [][]mamabear.Element, stage int) {
	innerDIFWithTwiddles(a, twiddles[stage], 0, 256, 256)
	kerDIFNP_256(a[:256], twiddles, stage+1)
	kerDIFNP_256(a[256:], twiddles, stage+1)
}

// kerDITNP_512 is an optimized 512-element DIT kernel.
func kerDITNP_512(a []mamabear.Element, twiddles [][]mamabear.Element, stage int) {
	kerDITNP_256(a[:256], twiddles, stage+1)
	kerDITNP_256(a[256:], twiddles, stage+1)
	innerDITWithTwiddles(a, twiddles[stage], 0, 256, 256)
}

// kerDIFNP_1024 is an optimized 1024-element DIF kernel.
func kerDIFNP_1024(a []mamabear.Element, twiddles [][]mamabear.Element, stage int) {
	innerDIFWithTwiddles(a, twiddles[stage], 0, 512, 512)
	innerDIFWithTwiddles(a[:512], twiddles[stage+1], 0, 256, 256)
	innerDIFWithTwiddles(a[512:], twiddles[stage+1], 0, 256, 256)
	kerDIFNP_256(a[:256], twiddles, stage+2)
	kerDIFNP_256(a[256:512], twiddles, stage+2)
	kerDIFNP_256(a[512:768], twiddles, stage+2)
	kerDIFNP_256(a[768:], twiddles, stage+2)
}

// kerDITNP_1024 is an optimized 1024-element DIT kernel.
func kerDITNP_1024(a []mamabear.Element, twiddles [][]mamabear.Element, stage int) {
	kerDITNP_256(a[:256], twiddles, stage+2)
	kerDITNP_256(a[256:512], twiddles, stage+2)
	kerDITNP_256(a[512:768], twiddles, stage+2)
	kerDITNP_256(a[768:], twiddles, stage+2)
	innerDITWithTwiddles(a[:512], twiddles[stage+1], 0, 256, 256)
	innerDITWithTwiddles(a[512:], twiddles[stage+1], 0, 256, 256)
	innerDITWithTwiddles(a, twiddles[stage], 0, 512, 512)
}
