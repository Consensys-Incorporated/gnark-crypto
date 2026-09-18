// Copyright 2020-2026 Consensys Software Inc.
// Licensed under the Apache License, Version 2.0. See the LICENSE file for details.

package mamabear

import (
	"fmt"
	"math/big"
	"math/bits"

	"github.com/consensys/gnark-crypto/ecc"
)

// Generator returns a generator for Z/2^(log(m))Z,
// or an error if m is too big (the required root of unity doesn't exist).
//
// p − 1 = 2^34 · 32767, so the maximum supported NTT size is 2^34.
func Generator(m uint64) (Element, error) {
	x := ecc.NextPowerOfTwo(m)

	// rootOfUnity has order 2^34 (primitive root g=3, rootOfUnity = 3^32767 mod p)
	var rootOfUnity Element
	rootOfUnity[0] = 393730615033094 // 3^32767 mod p, in Montgomery form
	const maxOrderRoot uint64 = 34

	logx := uint64(bits.TrailingZeros64(x))
	if logx > maxOrderRoot {
		return Element{}, fmt.Errorf("m (%d) is too big: the required root of unity does not exist", m)
	}

	expo := uint64(1 << (maxOrderRoot - logx))
	var generator Element
	generator.Exp(rootOfUnity, big.NewInt(int64(expo))) // order x
	return generator, nil
}
