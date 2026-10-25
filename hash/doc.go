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
// # Field-native hashing
//
// [FieldHasher] is the common interface for hashing field elements without byte
// encoding, and [FieldCompressor] the corresponding 2-to-1 compression function.
// Construct field hashers with poseidon2.NewHasher or mimc.NewHasher in the
// corresponding field package; see those packages for their block and digest
// sizes.
//
// [MerkleDamgardElement] and [MerkleDamgardArray] provide reusable field cores.
// The former absorbs one element per block into an element digest. The latter
// absorbs len(D) elements into an array digest, for the lengths in [ArrayBlock].
// Both take a compressor and initial chaining state; Reset restores that state.
// Array construction needs an explicit element type, for example
// NewMerkleDamgardArray[fr.Element](compressor, initialDigest), since Go cannot
// infer it from a union-of-arrays constraint. Hashing after construction does
// not allocate when the compressor itself does not allocate.
//
// WriteElement avoids the temporary variadic slice when called through an
// interface or generic type parameter. WriteElements accepts an existing slice.
// Both absorb a continuous element stream. SumDigest leaves the state unchanged;
// PadBlock commits a partial block with zeros on the left. Transcript callers
// should commit the block before continuing after a squeeze. BlockLen counts
// input elements, whereas hash.Hash.BlockSize counts bytes.
//
// [NewByteHasher] gives a field core a byte-oriented [StateStorer] view, with
// each Write ending on a block boundary.
//
// Extension-field input encoding and challenge derivation belong to the caller:
// absorb coordinates in a protocol-defined order, retain the full array digest
// for commitments, and convert to an extension element explicitly for challenges.
// Converting only part of a digest reduces its collision security; it is not an
// equivalent commitment digest.
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
// MiMC byte input is interpreted as canonical field elements, strictly smaller
// than the field modulus. Every Write left-pads a short final element with zeros.
// Poseidon2 byte input is similarly left-padded to its full input block. These
// encodings do not bind the length; callers hashing arbitrary bytes must provide
// explicit padding or length encoding. Prefer [FieldHasher] for field inputs.
//
// See open issues:
//   - https://github.com/Consensys/gnark-crypto/issues/504
//   - https://github.com/Consensys/gnark-crypto/issues/485
package hash
