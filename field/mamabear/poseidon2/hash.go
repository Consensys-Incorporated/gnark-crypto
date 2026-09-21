// Copyright 2020-2026 Consensys Software Inc.
// Licensed under the Apache License, Version 2.0. See the LICENSE file for details.

package poseidon2

import (
	"hash"
	"sync"

	fr "github.com/consensys/gnark-crypto/field/mamabear"
	gnarkHash "github.com/consensys/gnark-crypto/hash"
)

// NewMerkleDamgardHasher returns a Poseidon2 hasher using the Merkle-Damgard
// construction with the default parameters.
func NewMerkleDamgardHasher() gnarkHash.StateStorer {
	p := NewDefaultPermutation()
	return gnarkHash.NewMerkleDamgardHasher(
		p,
		make([]byte, p.params.Width/2*fr.Bytes),
	)
}

// NewDefaultPermutation returns a Poseidon2 permutation with the default parameters.
func NewDefaultPermutation() *Permutation {
	return &Permutation{params: GetDefaultParameters()}
}

// GetDefaultParameters returns the default Poseidon2 parameters (width=16, rF=6, rP=21).
var GetDefaultParameters = sync.OnceValue(func() *Parameters {
	return NewParameters(16, 6, 21)
})

var diag16 []fr.Element = make([]fr.Element, 16)
var diag24 []fr.Element = make([]fr.Element, 24)

func init() {
	// Diagonal values for matMulInternalInPlace, expressed via field operations so
	// the correct Montgomery encoding for MamaBear (R=2^52) is computed automatically.
	// Semantic values (same as KoalaBear):
	// diag16: [-2, 1, 2, 1/2, 3, 4, -1/2, -3, -4, 1/2^8, 1/8, 1/2^24, -1/2^8, -1/8, -1/16, -1/2^24]
	// diag24: [-2, 1, 2, 1/2, 3, 4, -1/2, -3, -4, 1/2^8, 1/4, 1/8, 1/16, 1/32, 1/64, 1/2^24,
	//           -1/2^8, -1/8, -1/16, -1/32, -1/64, -1/2^7, -1/2^9, -1/2^24]

	var one, two, three, four fr.Element
	one.SetOne()
	two.SetUint64(2)
	three.SetUint64(3)
	four.SetUint64(4)

	// diag16
	diag16[0].Neg(&two)
	diag16[1].SetOne()
	diag16[2].SetUint64(2)
	diag16[3].SetOne()
	diag16[3].Halve()
	diag16[4].SetUint64(3)
	diag16[5].SetUint64(4)
	diag16[6].SetOne()
	diag16[6].Halve()
	diag16[6].Neg(&diag16[6])
	diag16[7].Neg(&three)
	diag16[8].Neg(&four)
	diag16[9].Mul2ExpNegN(&one, 8)
	diag16[10].Mul2ExpNegN(&one, 3)
	diag16[11].Mul2ExpNegN(&one, 24)
	diag16[12].Mul2ExpNegN(&one, 8)
	diag16[12].Neg(&diag16[12])
	diag16[13].Mul2ExpNegN(&one, 3)
	diag16[13].Neg(&diag16[13])
	diag16[14].Mul2ExpNegN(&one, 4)
	diag16[14].Neg(&diag16[14])
	diag16[15].Mul2ExpNegN(&one, 24)
	diag16[15].Neg(&diag16[15])

	// diag24 (first 9 entries same as diag16)
	diag24[0].Neg(&two)
	diag24[1].SetOne()
	diag24[2].SetUint64(2)
	diag24[3].SetOne()
	diag24[3].Halve()
	diag24[4].SetUint64(3)
	diag24[5].SetUint64(4)
	diag24[6].SetOne()
	diag24[6].Halve()
	diag24[6].Neg(&diag24[6])
	diag24[7].Neg(&three)
	diag24[8].Neg(&four)
	diag24[9].Mul2ExpNegN(&one, 8)
	diag24[10].Mul2ExpNegN(&one, 2)
	diag24[11].Mul2ExpNegN(&one, 3)
	diag24[12].Mul2ExpNegN(&one, 4)
	diag24[13].Mul2ExpNegN(&one, 5)
	diag24[14].Mul2ExpNegN(&one, 6)
	diag24[15].Mul2ExpNegN(&one, 24)
	diag24[16].Mul2ExpNegN(&one, 8)
	diag24[16].Neg(&diag24[16])
	diag24[17].Mul2ExpNegN(&one, 3)
	diag24[17].Neg(&diag24[17])
	diag24[18].Mul2ExpNegN(&one, 4)
	diag24[18].Neg(&diag24[18])
	diag24[19].Mul2ExpNegN(&one, 5)
	diag24[19].Neg(&diag24[19])
	diag24[20].Mul2ExpNegN(&one, 6)
	diag24[20].Neg(&diag24[20])
	diag24[21].Mul2ExpNegN(&one, 7)
	diag24[21].Neg(&diag24[21])
	diag24[22].Mul2ExpNegN(&one, 9)
	diag24[22].Neg(&diag24[22])
	diag24[23].Mul2ExpNegN(&one, 24)
	diag24[23].Neg(&diag24[23])

	gnarkHash.RegisterHash(gnarkHash.POSEIDON2_MAMABEAR, func() hash.Hash {
		return NewMerkleDamgardHasher()
	})
}
