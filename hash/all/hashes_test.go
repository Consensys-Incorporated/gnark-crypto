// Copyright 2020-2026 Consensys Incorporated
// Licensed under the Apache License, Version 2.0. See the LICENSE file for details.

package all

import (
	"testing"

	bls12377 "github.com/consensys/gnark-crypto/ecc/bls12-377/fr"
	bls12381 "github.com/consensys/gnark-crypto/ecc/bls12-381/fr"
	bls24315 "github.com/consensys/gnark-crypto/ecc/bls24-315/fr"
	bls24317 "github.com/consensys/gnark-crypto/ecc/bls24-317/fr"
	bn254 "github.com/consensys/gnark-crypto/ecc/bn254/fr"
	bw6633 "github.com/consensys/gnark-crypto/ecc/bw6-633/fr"
	bw6761 "github.com/consensys/gnark-crypto/ecc/bw6-761/fr"
	grumpkin "github.com/consensys/gnark-crypto/ecc/grumpkin/fr"
	"github.com/consensys/gnark-crypto/field/babybear"
	babybearPoseidon2 "github.com/consensys/gnark-crypto/field/babybear/poseidon2"
	"github.com/consensys/gnark-crypto/field/goldilocks"
	goldilocksPoseidon2 "github.com/consensys/gnark-crypto/field/goldilocks/poseidon2"
	"github.com/consensys/gnark-crypto/field/koalabear"
	koalabearPoseidon2 "github.com/consensys/gnark-crypto/field/koalabear/poseidon2"
	"github.com/consensys/gnark-crypto/field/mamabear"
	mamabearPoseidon2 "github.com/consensys/gnark-crypto/field/mamabear/poseidon2"
	"github.com/consensys/gnark-crypto/hash"
)

// TestDigestSizeMatchesHasher checks the static digestSize table against the
// hashers it describes. The two drifted apart for every small-field Poseidon2
// hash: the table recorded the size of one field element while the hasher
// returns Width/2 of them, so a caller sizing a buffer from hash.Hash.Size()
// under-allocated by a factor of 4 to 8.
func TestDigestSizeMatchesHasher(t *testing.T) {
	for h := hash.Hash(0); h.Available(); h++ {
		hsh := h.New()
		if got, want := h.Size(), hsh.Size(); got != want {
			t.Errorf("%s: Hash.Size() = %d, but New().Size() = %d", h, got, want)
		}
		if got, want := len(hsh.Sum(nil)), hsh.Size(); got != want {
			t.Errorf("%s: len(Sum(nil)) = %d, but Size() = %d", h, got, want)
		}
	}
}

// checkFieldHasher checks that a field hasher is registered under name over F and D.
func checkFieldHasher[F, D any](t *testing.T, name string) {
	t.Helper()
	if _, err := hash.NewFieldHasher[F, D](name); err != nil {
		t.Error(err)
	}
}

// TestFieldHashersRegistered checks that this package provides every field hasher: MiMC and
// Poseidon2 over each curve's scalar field, and Poseidon2 over each small field with both digest
// lengths.
func TestFieldHashersRegistered(t *testing.T) {
	for _, name := range []string{hash.FieldHasherMIMC, hash.FieldHasherPOSEIDON2} {
		checkFieldHasher[bls12377.Element, bls12377.Element](t, name)
		checkFieldHasher[bls12381.Element, bls12381.Element](t, name)
		checkFieldHasher[bls24315.Element, bls24315.Element](t, name)
		checkFieldHasher[bls24317.Element, bls24317.Element](t, name)
		checkFieldHasher[bn254.Element, bn254.Element](t, name)
		checkFieldHasher[bw6633.Element, bw6633.Element](t, name)
		checkFieldHasher[bw6761.Element, bw6761.Element](t, name)
		checkFieldHasher[grumpkin.Element, grumpkin.Element](t, name)
	}

	checkFieldHasher[babybear.Element, [babybearPoseidon2.BlockLen]babybear.Element](t, hash.FieldHasherPOSEIDON2)
	checkFieldHasher[babybear.Element, [babybearPoseidon2.WideBlockLen]babybear.Element](t, hash.FieldHasherPOSEIDON2)
	checkFieldHasher[goldilocks.Element, [goldilocksPoseidon2.BlockLen]goldilocks.Element](t, hash.FieldHasherPOSEIDON2)
	checkFieldHasher[goldilocks.Element, [goldilocksPoseidon2.WideBlockLen]goldilocks.Element](t, hash.FieldHasherPOSEIDON2)
	checkFieldHasher[koalabear.Element, [koalabearPoseidon2.BlockLen]koalabear.Element](t, hash.FieldHasherPOSEIDON2)
	checkFieldHasher[koalabear.Element, [koalabearPoseidon2.WideBlockLen]koalabear.Element](t, hash.FieldHasherPOSEIDON2)
	checkFieldHasher[mamabear.Element, [mamabearPoseidon2.BlockLen]mamabear.Element](t, hash.FieldHasherPOSEIDON2)
	checkFieldHasher[mamabear.Element, [mamabearPoseidon2.WideBlockLen]mamabear.Element](t, hash.FieldHasherPOSEIDON2)
}
