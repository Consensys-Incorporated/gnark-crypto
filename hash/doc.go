// Package hash provides algebraic hash function defined over implemented curves
//
// This package is kept for backwards compatibility for initializing hash
// functions directly. The recommended way to initialize hash function is to
// directly use the constructors in the corresponding packages (e.g.
// ecc/bn254/fr/mimc). Using the direct constructors allows to apply options for
// altering the hash function behavior (endianness, input splicing etc.) and
// returns more specific types with additional methods.
//
// See [Importing hash functions] below for more information.
//
// This package also provides a construction for a generic hash function from
// the compression primitive. See the interface [Compressor] and the
// corresponding initialization function [NewMerkleDamgardHasher].
//
// # Importing hash functions
//
// The package follows registration pattern for importing hash functions. To
// import all known hash functions in gnark-crypto, import the
// [github.com/consensys/gnark-crypto/hash/all] package in your code. To import
// only a specific hash, then import the corresponding package directly, e.g.
// [github.com/consensys/gnark-crypto/ecc/bn254/fr/mimc]. The import format should be:
//
//	import _ "github.com/consensys/gnark-crypto/ecc/bn254/fr/mimc"
//
// # Length extension attack
//
// The MiMC hash function is vulnerable to a length extension attack. For
// example when we have a hash
//
//	h = MiMC(k || m)
//
// and we want to hash a new message
//
//	m' = m || m2,
//
// we can compute
//
//	h' = MiMC(k || m || m2)
//
// without knowing k by computing
//
//	h' = MiMC(h || m2).
//
// This is because the MiMC hash function is a simple iterated cipher, and the
// hash value is the state of the cipher after encrypting the message.
//
// There are several ways to mitigate this attack:
//   - use a random key for each hash
//   - use a domain separation tag for different use cases:
//     h = MiMC(k || tag || m)
//   - use the secret input as last input:
//     h = MiMC(m || k)
//
// In general, inside a circuit the length-extension attack is not a concern as
// due to the circuit definition the attacker can not append messages to
// existing hash. But the user has to consider the cases when using a secret key
// and MiMC in different contexts.
//
// # Hash input format
//
// The MiMC hash function is defined over a field. The input to the hash
// function is a byte slice. The byte slice is interpreted as a sequence of
// field elements. Due to this interpretation, the input byte slice length must
// be multiple of the field modulus size. And every sequence of byte slice for a
// single field element must be strictly less than the field modulus.
//
// # Hashing extension field elements
//
// A hasher over a base field hashes the elements of an extension of it by
// writing their coordinates, in the order of AppendBase. Its digest is made of
// base field elements, and the caller converts it explicitly to an extension
// element, for example with the first coordinates of the digest and SetBase:
//
//	h := poseidon2.NewMerkleDamgardFieldHasher()
//	var coords []koalabear.Element
//	for _, e := range elements { // elements of extensions.E6
//		coords = e.AppendBase(coords[:0])
//		h.WriteElements(coords...)
//	}
//	d := h.Sum()
//	var digest extensions.E6
//	digest.SetBase(d[:6])
//
// See open issues:
//   - https://github.com/Consensys/gnark-crypto/issues/504
//   - https://github.com/Consensys/gnark-crypto/issues/485
package hash
