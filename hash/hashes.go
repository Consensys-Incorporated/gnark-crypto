// Copyright 2020-2025 Consensys Incorporated
// Licensed under the Apache License, Version 2.0. See the LICENSE file for details.

package hash

import (
	"fmt"
	"hash"
	"reflect"
	"strings"
)

var (
	hashes       = make([]func() hash.Hash, maxHash)
	hashesByName = make(map[string]func() hash.Hash)
	fieldHashers = make(fieldHasherRegistry)
)

// The names of the algorithms of the registered field hashers, to be used with
// [RegisterFieldHasher] and [NewFieldHasher].
const (
	FieldHasherMiMC      = "MIMC"
	FieldHasherPoseidon2 = "POSEIDON2"
)

// fieldHasherRegistry maps the names of field hashers, and the types they are
// registered over, to their constructors.
type fieldHasherRegistry map[fieldHasherKey]any

// fieldHasherKey identifies a field hasher by its name and the type
// FieldHasher[F, D] it is registered over. The registered value is its
// constructor, a func() FieldHasher[F, D].
type fieldHasherKey struct {
	name       string
	hasherType reflect.Type
}

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

// RegisterFieldHasher registers newHasher as the constructor of the field hasher
// called name, over elements of type F with digests of type D. It should be
// called in the init function of the package defining the hasher, and panics if
// name is already registered over F and D. Like [RegisterHash], it is not safe
// for use after init.
//
// Unlike the names of [NewHash], those of field hashers are the names of their
// algorithm alone, such as [FieldHasherMiMC] or [FieldHasherPoseidon2], with no
// suffix for the field: the field and the digest are given by F and D. The same
// name may be registered over several pairs of types.
func RegisterFieldHasher[F, D any](name string, newHasher func() FieldHasher[F, D]) {
	registerFieldHasher(fieldHashers, name, newHasher)
}

// NewFieldHasher returns a new field hasher called name, over elements of type F
// with digests of type D, or an error if none is registered under these. The
// package defining the hasher must have been imported, by a blank import or
// through [github.com/consensys/gnark-crypto/hash/all], as for [NewHash].
//
// Unlike the names of [NewHash], those of field hashers are the names of their
// algorithm alone, such as [FieldHasherMiMC] or [FieldHasherPoseidon2], with no
// suffix for the field: the field and the digest are given by F and D. The same
// name over different digest types is a different function: [FieldHasherPoseidon2]
// over [8] and [12] elements of KoalaBear are Poseidon2 permutations of widths 16
// and 24.
func NewFieldHasher[F, D any](name string) (FieldHasher[F, D], error) {
	return newFieldHasher[F, D](fieldHashers, name)
}

func registerFieldHasher[F, D any](r fieldHasherRegistry, name string, newHasher func() FieldHasher[F, D]) {
	key := fieldHasherKey{name, reflect.TypeFor[FieldHasher[F, D]]()}
	if _, ok := r[key]; ok {
		panic(fmt.Errorf("field hasher \"%s\" over elements of type %v and digests of type %v already registered",
			name, reflect.TypeFor[F](), reflect.TypeFor[D]()))
	}
	r[key] = newHasher
}

func newFieldHasher[F, D any](r fieldHasherRegistry, name string) (FieldHasher[F, D], error) {
	newHasher, ok := r[fieldHasherKey{name, reflect.TypeFor[FieldHasher[F, D]]()}].(func() FieldHasher[F, D])
	if !ok {
		return nil, fmt.Errorf("field hasher \"%s\" over elements of type %v and digests of type %v not registered",
			name, reflect.TypeFor[F](), reflect.TypeFor[D]())
	}
	return newHasher(), nil
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

// size of digests in bytes.
//
// These must agree with the Size() of the hasher each one describes;
// TestDigestSizeMatchesHasher in hash/all enforces it. The curve entries are
// the size of the SCALAR field fr, which is what MiMC and Poseidon2 are
// instantiated over -- not the base field fp. Using fp here over-reported every
// non-BN254 curve digest by up to a factor of two.
var digestSize = []uint8{
	MIMC_BN254:     32,
	MIMC_BLS12_381: 32,
	MIMC_BLS12_377: 32,
	MIMC_BW6_761:   48,
	MIMC_BLS24_315: 32,
	MIMC_BLS24_317: 32,
	MIMC_BW6_633:   40,
	MIMC_GRUMPKIN:  32,

	POSEIDON2_BN254:     32,
	POSEIDON2_BLS12_381: 32,
	POSEIDON2_BLS12_377: 32,
	POSEIDON2_BW6_761:   48,
	POSEIDON2_BLS24_315: 32,
	POSEIDON2_BLS24_317: 32,
	POSEIDON2_BW6_633:   40,
	POSEIDON2_GRUMPKIN:  32,

	// The small-field Poseidon2 hashers are Merkle-Damgard over a width-t
	// compression function, so a digest is t/2 field elements, each stored
	// padded to fr.Bytes. These are NOT the size of a single element: entries
	// of 4, 4, 8, 8 here used to under-report the digest by a factor of 4 to 8.
	POSEIDON2_KOALABEAR:  32, // 8 x 4 bytes (t = 16)
	POSEIDON2_BABYBEAR:   32, // 8 x 4 bytes (t = 16)
	POSEIDON2_GOLDILOCKS: 32, // 4 x 8 bytes (t = 8)
	POSEIDON2_MAMABEAR:   48, // 6 x 8 bytes (t = 12)
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
