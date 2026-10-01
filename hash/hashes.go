// Copyright 2020-2025 Consensys Software Inc.
// Licensed under the Apache License, Version 2.0. See the LICENSE file for details.

package hash

import (
	"fmt"
	"hash"
	"strings"
)

var (
	hashes       = make([]func() hash.Hash, maxHash)
	hashesByName = make(map[string]func() hash.Hash)
)

func findStdHashByName(name string) Hash {
	for h := range maxHash {
		if h.String() == name {
			return h
		}
	}
	return maxHash
}

// NewHash returns a new hash.Hash object for the given hash function name.
// It can be a standard hash function (e.g. "MIMC_BN254"),
// or a custom hash function defined by the user through [RegisterCustomHash].
func NewHash(name string) hash.Hash {
	// first see if it's a standard hash function
	if h := findStdHashByName(name); h < maxHash {
		return h.New()
	}

	// see if it's a custom hash function - registered by the user
	if f, ok := hashesByName[name]; ok {
		return f()
	}

	panic(fmt.Errorf("hash function \"%s\" not registered", name))
}

// RegisterHash registers a new hash function constructor. Should be called in
// the init function of the hash package.
//
// To register all known hash functions in gnark-crypto, import the
// [github.com/consensys/gnark-crypto/hash/all] package in your code.
func RegisterHash(h Hash, new func() hash.Hash) {
	hashes[h] = new
}

// RegisterCustomHash registers a new hash function constructor, retrievable by name
// using NewHash. It does not allow overwriting standard hash functions.
func RegisterCustomHash(name string, new func() hash.Hash) {
	if h := findStdHashByName(name); h < maxHash {
		panic(fmt.Errorf("cannot overwrite standard hash function \"%s\"", name))
	}
	hashesByName[name] = new
}

// Hash defines a unique identifier for a hash function.
type Hash uint

const (
	// MIMC_BN254 is the MiMC hash function for the BN254 curve.
	MIMC_BN254 Hash = iota
	// MIMC_BLS12_381 is the MiMC hash function for the BLS12-381 curve.
	MIMC_BLS12_381
	// MIMC_BLS12_377 is the MiMC hash function for the BLS12-377 curve.
	MIMC_BLS12_377
	// MIMC_BW6_761 is the MiMC hash function for the BW6-761 curve.
	MIMC_BW6_761
	// MIMC_BLS24_315 is the MiMC hash function for the BLS24-315 curve.
	MIMC_BLS24_315
	// MIMC_BLS24_317 is the MiMC hash function for the BLS24-317 curve.
	MIMC_BLS24_317
	// MIMC_BW6_633 is the MiMC hash function for the BW6-633 curve.
	MIMC_BW6_633
	// MIMC_GRUMPKIN is the MiMC hash function for the Grumpkin curve.
	MIMC_GRUMPKIN

	// POSEIDON2_BLS12_377 is the Poseidon2 hash function for the BLS12-377 curve.
	POSEIDON2_BLS12_377
	// POSEIDON2_BLS12_381 is the Poseidon2 hash function for the BLS12-381 curve.
	POSEIDON2_BLS12_381
	// POSEIDON2_BN254 is the Poseidon2 hash function for the BN254 curve.
	POSEIDON2_BN254
	// POSEIDON2_GRUMPKIN is the Poseidon2 hash function for the Grumpkin curve.
	POSEIDON2_GRUMPKIN
	// POSEIDON2_BW6_761 is the Poseidon2 hash function for the BW6-761 curve.
	POSEIDON2_BW6_761
	// POSEIDON2_BW6_633 is the Poseidon2 hash function for the BW6-633 curve.
	POSEIDON2_BW6_633
	// POSEIDON2_BLS24_315 is the Poseidon2 hash function for the BLS21-315 curve.
	POSEIDON2_BLS24_315
	// POSEIDON2_BLS24_317 is the Poseidon2 hash function for the BLS21-317 curve.
	POSEIDON2_BLS24_317
	// POSEIDON2_KOALABEAR is the Poseidon2 hash function for the KoalaBear field.
	POSEIDON2_KOALABEAR
	// POSEIDON2_BABYBEAR is the Poseidon2 hash function for the BabyBear field.
	POSEIDON2_BABYBEAR
	// POSEIDON2_GOLDILOCKS is the Poseidon2 hash function for the Goldilocks field.
	POSEIDON2_GOLDILOCKS
	// POSEIDON2_MAMABEAR is the Poseidon2 hash function for the MamaBear field.
	POSEIDON2_MAMABEAR

	maxHash
)

// size of digests in bytes
var digestSize = []uint8{
	MIMC_BN254:     32,
	MIMC_BLS12_381: 48,
	MIMC_BLS12_377: 48,
	MIMC_BW6_761:   96,
	MIMC_BLS24_315: 48,
	MIMC_BLS24_317: 48,
	MIMC_BW6_633:   80,
	MIMC_GRUMPKIN:  32,

	POSEIDON2_BN254:     32,
	POSEIDON2_BLS12_381: 48,
	POSEIDON2_BLS12_377: 48,
	POSEIDON2_BW6_761:   96,
	POSEIDON2_BLS24_315: 48,
	POSEIDON2_BLS24_317: 48,
	POSEIDON2_BW6_633:   80,
	POSEIDON2_GRUMPKIN:  32,

	POSEIDON2_KOALABEAR:  4,
	POSEIDON2_BABYBEAR:   4,
	POSEIDON2_GOLDILOCKS: 8,
	POSEIDON2_MAMABEAR:   8,
}

// New initializes the hash function. This is a convenience function which does
// not allow setting hash-specific options.
func (m Hash) New() hash.Hash {
	if m < maxHash {
		f := hashes[m]
		if f != nil {
			return f()
		}
	}
	pkgname, _ := strings.CutPrefix(m.String(), "MIMC_")
	pkgname = strings.ToLower(pkgname)
	pkgname = strings.ReplaceAll(pkgname, "_", "-")
	msg := fmt.Sprintf(`requested hash function #%s not registered. Import the corresponding package to register it:
	import _ "github.com/consensys/gnark-crypto/ecc/%s/fr/mimc"`, m.String(), pkgname)
	panic(msg)
}

// String returns the unique identifier of the hash function.
func (m Hash) String() string {
	switch m {
	case MIMC_BN254:
		return "MIMC_BN254"
	case MIMC_BLS12_381:
		return "MIMC_BLS12_381"
	case MIMC_BLS12_377:
		return "MIMC_BLS12_377"
	case MIMC_BW6_761:
		return "MIMC_BW6_761"
	case MIMC_BLS24_315:
		return "MIMC_BLS24_315"
	case MIMC_BLS24_317:
		return "MIMC_BLS24_317"
	case MIMC_BW6_633:
		return "MIMC_BW6_633"
	case MIMC_GRUMPKIN:
		return "MIMC_GRUMPKIN"

	case POSEIDON2_BN254:
		return "POSEIDON2_BN254"
	case POSEIDON2_BLS12_381:
		return "POSEIDON2_BLS12_381"
	case POSEIDON2_BLS12_377:
		return "POSEIDON2_BLS12_377"
	case POSEIDON2_BW6_761:
		return "POSEIDON2_BW6_761"
	case POSEIDON2_BLS24_315:
		return "POSEIDON2_BLS24_315"
	case POSEIDON2_BLS24_317:
		return "POSEIDON2_BLS24_317"
	case POSEIDON2_BW6_633:
		return "POSEIDON2_BW6_633"
	case POSEIDON2_GRUMPKIN:
		return "POSEIDON2_GRUMPKIN"
	case POSEIDON2_KOALABEAR:
		return "POSEIDON2_KOALABEAR"
	case POSEIDON2_BABYBEAR:
		return "POSEIDON2_BABYBEAR"
	case POSEIDON2_GOLDILOCKS:
		return "POSEIDON2_GOLDILOCKS"
	case POSEIDON2_MAMABEAR:
		return "POSEIDON2_MAMABEAR"
	default:
		return "unknown hash function"
	}
}

// Available returns true if the hash function is available.
func (m Hash) Available() bool {
	return m < maxHash && hashes[m] != nil
}

// Size returns the size of the digest of the corresponding hash function
func (m Hash) Size() int {
	return int(digestSize[m])
}
