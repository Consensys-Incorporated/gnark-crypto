// Copyright 2020-2026 Consensys Software Inc.
// Licensed under the Apache License, Version 2.0. See the LICENSE file for details.

package vortex

import (
	"github.com/consensys/gnark-crypto/field/mamabear"
	"github.com/consensys/gnark-crypto/field/mamabear/poseidon2"
)

var (
	compressPerm = poseidon2.NewPermutation(16, 6, 21)
	spongePerm   = poseidon2.NewPermutation(24, 6, 21)
)

// CompressPoseidon2 runs the Poseidon2 permutation over two hashes.
// The two 4-element inputs occupy positions 0..7 of a width-16 state
// (positions 8..15 are zero). The output is positions 0..3 with
// feed-forward from the left input.
func CompressPoseidon2(a, b Hash) Hash {
	var x [16]mamabear.Element
	copy(x[:4], a[:])
	copy(x[4:8], b[:])
	// x[8:16] are zero

	feedforward := a // save left input for feed-forward

	if err := compressPerm.Permutation(x[:]); err != nil {
		panic(err)
	}

	var res Hash
	for i := range 4 {
		res[i].Add(&feedforward[i], &x[i])
	}
	return res
}

// HashPoseidon2 returns a Poseidon2 sponge hash of a slice of field elements.
// The input is absorbed in blocks of 20 (stateSize - len(Hash)) using the duplex construction.
func HashPoseidon2(input []mamabear.Element) Hash {
	const (
		stateSize = 24
		rate      = stateSize - 4 // 4 elements reserved for capacity (= len(Hash))
	)
	var state [stateSize]mamabear.Element
	for i := 0; i < len(input); i += rate {
		end := i + rate
		if end > len(input) {
			end = len(input)
		}
		copy(state[4:], input[i:end])
		spongePerm.Permutation(state[:]) //nolint: errcheck
	}
	var res Hash
	copy(res[:], state[:4])
	return res
}
