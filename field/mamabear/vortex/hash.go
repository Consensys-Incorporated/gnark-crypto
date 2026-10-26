// Copyright 2020-2026 Consensys Incorporated
// Licensed under the Apache License, Version 2.0. See the LICENSE file for details.

package vortex

import (
	"github.com/consensys/gnark-crypto/field/mamabear"
	"github.com/consensys/gnark-crypto/field/mamabear/poseidon2"
)

// The permutations come from the field's own default parameters rather than
// from literals. Hardcoding them here is how this package ended up running
// mamabear with koalabear's round numbers (6, 21): the partial-round count has
// to grow with log2(p), and 21 rounds over a 49-bit field targets about 33 bits
// of security instead of 128. See the mamabear case in
// internal/generator/field/generator_poseidon2.go for the derivation.
var (
	compressPerm = poseidon2.NewDefaultPermutation()
	spongePerm   = poseidon2.NewPermutation(spongeStateSize, defaultFullRounds, defaultPartialRounds)
)

const (
	spongeStateSize      = 24
	defaultFullRounds    = 8
	defaultPartialRounds = 32
)

// CompressPoseidon2 runs the Poseidon2 permutation over two hashes.
// The two len(Hash)-element inputs fill the whole width-2*len(Hash) state. The
// output is the first len(Hash) positions with feed-forward from the left
// input.
func CompressPoseidon2(a, b Hash) Hash {
	var x [2 * HashSize]mamabear.Element
	copy(x[:HashSize], a[:])
	copy(x[HashSize:], b[:])

	feedforward := a // save left input for feed-forward

	if err := compressPerm.Permutation(x[:]); err != nil {
		panic(err)
	}

	var res Hash
	for i := range HashSize {
		res[i].Add(&feedforward[i], &x[i])
	}
	return res
}

// HashPoseidon2 returns a Poseidon2 sponge hash of a slice of field elements.
// The input is absorbed using the duplex construction, with len(Hash) elements
// of the state reserved as capacity. The capacity must match the digest width:
// a narrower capacity would cap the collision resistance of the sponge below
// that of the digest it produces.
func HashPoseidon2(input []mamabear.Element) Hash {
	const (
		stateSize = spongeStateSize
		rate      = stateSize - HashSize
	)
	var state [stateSize]mamabear.Element
	for i := 0; i < len(input); i += rate {
		end := i + rate
		if end > len(input) {
			end = len(input)
		}
		copy(state[HashSize:], input[i:end])
		spongePerm.Permutation(state[:]) //nolint: errcheck
	}
	var res Hash
	copy(res[:], state[:HashSize])
	return res
}
