// Copyright 2020-2026 Consensys Software Inc.
// Licensed under the Apache License, Version 2.0. See the LICENSE file for details.

package poseidon2

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"slices"

	"golang.org/x/crypto/sha3"

	fr "github.com/consensys/gnark-crypto/field/mamabear"
)

var (
	ErrInvalidSizebuffer = errors.New("the size of the input should match the size of the hash buffer")
)

const (
	d = 3
)

// DegreeSBox returns the degree of the sBox function used in the Poseidon2 permutation.
func DegreeSBox() int {
	return d
}

// Parameters describes the Poseidon2 permutation instance.
type Parameters struct {
	Width           int
	NbFullRounds    int
	NbPartialRounds int
	RoundKeys       [][]fr.Element
}

// NewParameters returns parameters with round keys derived from a digest of the parameters.
func NewParameters(width, nbFullRounds, nbPartialRounds int) *Parameters {
	p := Parameters{Width: width, NbFullRounds: nbFullRounds, NbPartialRounds: nbPartialRounds}
	p.initRC(p.String())
	return &p
}

// NewParametersWithSeed returns parameters with round keys derived from the given seed.
func NewParametersWithSeed(width, nbFullRounds, nbPartialRounds int, seed string) *Parameters {
	p := Parameters{Width: width, NbFullRounds: nbFullRounds, NbPartialRounds: nbPartialRounds}
	p.initRC(seed)
	return &p
}

func (p *Parameters) String() string {
	return fmt.Sprintf("Poseidon2-mamabear[t=%d,rF=%d,rP=%d,d=%d]", p.Width, p.NbFullRounds, p.NbPartialRounds, d)
}

func (p *Parameters) initRC(seed string) {
	bseed := ([]byte)(seed)
	hash := sha3.NewLegacyKeccak256()
	_, _ = hash.Write(bseed)
	rnd := hash.Sum(nil)
	hash.Reset()
	_, _ = hash.Write(rnd)

	roundKeys := make([][]fr.Element, p.NbFullRounds+p.NbPartialRounds)
	for i := range p.NbFullRounds / 2 {
		roundKeys[i] = make([]fr.Element, p.Width)
		for j := range p.Width {
			rnd = hash.Sum(nil)
			roundKeys[i][j].SetBytes(rnd)
			hash.Reset()
			_, _ = hash.Write(rnd)
		}
	}
	for i := p.NbFullRounds / 2; i < p.NbPartialRounds+p.NbFullRounds/2; i++ {
		roundKeys[i] = make([]fr.Element, 1)
		rnd = hash.Sum(nil)
		roundKeys[i][0].SetBytes(rnd)
		hash.Reset()
		_, _ = hash.Write(rnd)
	}
	for i := p.NbPartialRounds + p.NbFullRounds/2; i < p.NbPartialRounds+p.NbFullRounds; i++ {
		roundKeys[i] = make([]fr.Element, p.Width)
		for j := range p.Width {
			rnd = hash.Sum(nil)
			roundKeys[i][j].SetBytes(rnd)
			hash.Reset()
			_, _ = hash.Write(rnd)
		}
	}
	p.RoundKeys = roundKeys
}

// Permutation stores the buffer of the Poseidon2 permutation.
type Permutation struct {
	params *Parameters
}

// NewPermutation returns a new Poseidon2 permutation instance.
func NewPermutation(t, rf, rp int) *Permutation {
	if t != 16 && t != 24 {
		panic("only Width=16,24 are supported")
	}
	return &Permutation{params: NewParameters(t, rf, rp)}
}

// NewPermutationWithSeed returns a new Poseidon2 permutation instance with a given seed.
func NewPermutationWithSeed(t, rf, rp int, seed string) *Permutation {
	if t != 16 && t != 24 {
		panic("only Width=16,24 are supported")
	}
	return &Permutation{params: NewParametersWithSeed(t, rf, rp, seed)}
}

func (h *Permutation) sBox(index int, input []fr.Element) {
	var tmp fr.Element
	tmp.Set(&input[index])
	input[index].Square(&input[index]).Mul(&input[index], &tmp)
}

// matMulM4InPlace computes s <- M4*s where M4 = circ(2,3,1,1).
func (h *Permutation) matMulM4InPlace(s []fr.Element) {
	c := len(s) / 4
	for i := range c {
		var t01, t23, t0123, t01123, t01233 fr.Element
		t01.Add(&s[4*i], &s[4*i+1])
		t23.Add(&s[4*i+2], &s[4*i+3])
		t0123.Add(&t01, &t23)
		t01123.Add(&t0123, &s[4*i+1])
		t01233.Add(&t0123, &s[4*i+3])
		s[4*i+3].Double(&s[4*i]).Add(&s[4*i+3], &t01233)
		s[4*i+1].Double(&s[4*i+2]).Add(&s[4*i+1], &t01123)
		s[4*i].Add(&t01, &t01123)
		s[4*i+2].Add(&t23, &t01233)
	}
}

func (h *Permutation) matMulExternalInPlace(input []fr.Element) {
	if h.params.Width%4 != 0 {
		panic("only Width = 0 mod 4 are supported")
	}
	h.matMulM4InPlace(input)
	tmp := make([]fr.Element, 4)
	for i := range h.params.Width / 4 {
		tmp[0].Add(&tmp[0], &input[4*i])
		tmp[1].Add(&tmp[1], &input[4*i+1])
		tmp[2].Add(&tmp[2], &input[4*i+2])
		tmp[3].Add(&tmp[3], &input[4*i+3])
	}
	for i := range h.params.Width / 4 {
		input[4*i].Add(&input[4*i], &tmp[0])
		input[4*i+1].Add(&input[4*i+1], &tmp[1])
		input[4*i+2].Add(&input[4*i+2], &tmp[2])
		input[4*i+3].Add(&input[4*i+3], &tmp[3])
	}
}

func (h *Permutation) matMulInternalInPlace(input []fr.Element) {
	switch h.params.Width {
	case 16:
		var sum fr.Element
		sum.Set(&input[0])
		for i := 1; i < h.params.Width; i++ {
			sum.Add(&sum, &input[i])
		}
		// diagonal: [-2, 1, 2, 1/2, 3, 4, -1/2, -3, -4, 1/2^8, 1/8, 1/2^24, -1/2^8, -1/8, -1/16, -1/2^24]
		var temp fr.Element
		input[0].Sub(&sum, temp.Double(&input[0]))
		input[1].Add(&sum, &input[1])
		input[2].Add(&sum, temp.Double(&input[2]))
		temp.Set(&input[3])
		temp.Halve()
		input[3].Add(&sum, &temp)
		input[4].Add(&sum, temp.Double(&input[4]).Add(&temp, &input[4]))
		input[5].Add(&sum, temp.Double(&input[5]).Double(&temp))
		temp.Set(&input[6])
		temp.Halve()
		input[6].Sub(&sum, &temp)
		input[7].Sub(&sum, temp.Double(&input[7]).Add(&temp, &input[7]))
		input[8].Sub(&sum, temp.Double(&input[8]).Double(&temp))
		input[9].Add(&sum, temp.Mul2ExpNegN(&input[9], 8))
		input[10].Add(&sum, temp.Mul2ExpNegN(&input[10], 3))
		input[11].Add(&sum, temp.Mul2ExpNegN(&input[11], 24))
		input[12].Sub(&sum, temp.Mul2ExpNegN(&input[12], 8))
		input[13].Sub(&sum, temp.Mul2ExpNegN(&input[13], 3))
		input[14].Sub(&sum, temp.Mul2ExpNegN(&input[14], 4))
		input[15].Sub(&sum, temp.Mul2ExpNegN(&input[15], 24))
	case 24:
		var sum fr.Element
		sum.Set(&input[0])
		for i := 1; i < h.params.Width; i++ {
			sum.Add(&sum, &input[i])
		}
		// diagonal: [-2, 1, 2, 1/2, 3, 4, -1/2, -3, -4, 1/2^8, 1/4, 1/8, 1/16, 1/32, 1/64, 1/2^24, -1/2^8, -1/8, -1/16, -1/32, -1/64, -1/2^7, -1/2^9, -1/2^24]
		var temp fr.Element
		input[0].Sub(&sum, temp.Double(&input[0]))
		input[1].Add(&sum, &input[1])
		input[2].Add(&sum, temp.Double(&input[2]))
		temp.Set(&input[3])
		temp.Halve()
		input[3].Add(&sum, &temp)
		input[4].Add(&sum, temp.Double(&input[4]).Add(&temp, &input[4]))
		input[5].Add(&sum, temp.Double(&input[5]).Double(&temp))
		temp.Set(&input[6])
		temp.Halve()
		input[6].Sub(&sum, &temp)
		input[7].Sub(&sum, temp.Double(&input[7]).Add(&temp, &input[7]))
		input[8].Sub(&sum, temp.Double(&input[8]).Double(&temp))
		input[9].Add(&sum, temp.Mul2ExpNegN(&input[9], 8))
		input[10].Add(&sum, temp.Mul2ExpNegN(&input[10], 2))
		input[11].Add(&sum, temp.Mul2ExpNegN(&input[11], 3))
		input[12].Add(&sum, temp.Mul2ExpNegN(&input[12], 4))
		input[13].Add(&sum, temp.Mul2ExpNegN(&input[13], 5))
		input[14].Add(&sum, temp.Mul2ExpNegN(&input[14], 6))
		input[15].Add(&sum, temp.Mul2ExpNegN(&input[15], 24))
		input[16].Sub(&sum, temp.Mul2ExpNegN(&input[16], 8))  //nolint: gosec
		input[17].Sub(&sum, temp.Mul2ExpNegN(&input[17], 3))  //nolint: gosec
		input[18].Sub(&sum, temp.Mul2ExpNegN(&input[18], 4))  //nolint: gosec
		input[19].Sub(&sum, temp.Mul2ExpNegN(&input[19], 5))  //nolint: gosec
		input[20].Sub(&sum, temp.Mul2ExpNegN(&input[20], 6))  //nolint: gosec
		input[21].Sub(&sum, temp.Mul2ExpNegN(&input[21], 7))  //nolint: gosec
		input[22].Sub(&sum, temp.Mul2ExpNegN(&input[22], 9))  //nolint: gosec
		input[23].Sub(&sum, temp.Mul2ExpNegN(&input[23], 24)) //nolint: gosec
	default:
		panic("only Width=16,24 are supported")
	}
}

func (h *Permutation) addRoundKeyInPlace(round int, input []fr.Element) {
	for i := range len(h.params.RoundKeys[round]) {
		input[i].Add(&input[i], &h.params.RoundKeys[round][i])
	}
}

func (h *Permutation) BlockSize() int {
	return h.params.Width / 2 * fr.Bytes
}

// Permutation applies the permutation on input in place.
func (h *Permutation) Permutation(input []fr.Element) error {
	if len(input) != h.params.Width {
		return ErrInvalidSizebuffer
	}

	h.matMulExternalInPlace(input)

	rf := h.params.NbFullRounds / 2
	for i := range rf {
		h.addRoundKeyInPlace(i, input)
		for j := range h.params.Width {
			h.sBox(j, input)
		}
		h.matMulExternalInPlace(input)
	}
	for i := rf; i < rf+h.params.NbPartialRounds; i++ {
		h.addRoundKeyInPlace(i, input)
		h.sBox(0, input)
		h.matMulInternalInPlace(input)
	}
	for i := rf + h.params.NbPartialRounds; i < h.params.NbFullRounds+h.params.NbPartialRounds; i++ {
		h.addRoundKeyInPlace(i, input)
		for j := range h.params.Width {
			h.sBox(j, input)
		}
		h.matMulExternalInPlace(input)
	}

	return nil
}

// Compress uses the permutation to compress two equal-length byte slices.
func (h *Permutation) Compress(left []byte, right []byte) ([]byte, error) {
	n := h.params.Width / 2
	if h.params.Width != 2*n {
		return nil, errors.New("need even width")
	}

	desiredLen := n * fr.Bytes
	if len(left) != desiredLen || len(right) != desiredLen {
		return nil, fmt.Errorf("left input should be %d bytes", desiredLen)
	}

	reader := io.MultiReader(bytes.NewReader(left), bytes.NewReader(right))
	x := make([]fr.Element, h.params.Width)
	var buf [fr.Bytes]byte

	for i := range x {
		if _, err := io.ReadFull(reader, buf[:]); err != nil {
			return nil, err
		}
		if err := x[i].SetBytesCanonical(buf[:]); err != nil {
			return nil, err
		}
	}

	res := slices.Clone(x[n:])
	if err := h.Permutation(x[:]); err != nil {
		return nil, err
	}

	outBytes := make([]byte, 0, n*fr.Bytes)
	for i := range res {
		outBytes = append(outBytes, res[i].Add(&res[i], &x[n+i]).Marshal()...)
	}

	return outBytes, nil
}
