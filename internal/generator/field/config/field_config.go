// Copyright 2020-2025 Consensys Software Inc.
// Licensed under the Apache License, Version 2.0. See the LICENSE file for details.

// Package config provides Golang code generation for efficient field arithmetic operations.
package config

import (
	"errors"
	"fmt"
	"math"
	"math/big"
	"strconv"
	"strings"

	"github.com/consensys/bavard"
	"github.com/consensys/gnark-crypto/internal/generator/addchain"
)

var (
	errParseModulus     = errors.New("can't parse modulus")
	errInvalidModulus   = errors.New("modulus must be an odd integer greater than 2")
	errUnsupportedRadix = errors.New("a Montgomery radix narrower than the storage word is only supported for single-word fields with NbBits < RBits < 64")
)

// sparsePrimeExponents reports whether q == 2^hi - 2^lo + 1 with hi > lo > 0,
// the shape that admits a cheap reduction via 2^hi = 2^lo - 1 (mod q).
func sparsePrimeExponents(q *big.Int) (hi, lo uint, ok bool) {
	// q - 1 == 2^hi - 2^lo, i.e. lo trailing zeros followed by (hi-lo) ones.
	t := new(big.Int).Sub(q, big.NewInt(1))
	if t.Sign() <= 0 {
		return 0, 0, false
	}
	lo = uint(t.TrailingZeroBits())
	hi = uint(t.BitLen())
	// rebuild and compare, so we accept only the exact shape
	want := new(big.Int).Lsh(big.NewInt(1), hi)
	want.Sub(want, new(big.Int).Lsh(big.NewInt(1), lo))
	if hi <= lo || want.Cmp(t) != 0 {
		return 0, 0, false
	}
	return hi, lo, true
}

// Field precomputed values used in template for code generation of field element APIs
type Field struct {
	PackageName                string
	ElementName                string
	ModulusBig                 *big.Int
	Modulus                    string
	ModulusHex                 string
	NbWords                    int
	NbBits                     int
	NbBytes                    int
	NbWordsLastIndex           int
	NbWordsIndexesNoZeroNoLast []int
	NbWordsIndexesNoZero       []int
	NbWordsIndexesFull         []int
	P20InversionCorrectiveFac  []uint64
	P20InversionNbIterations   int
	UsingP20Inverse            bool
	IsMSWSaturated             bool // indicates if the most significant word is 0xFFFFF...FFFF
	Q                          []uint64
	QInverse                   []uint64
	QMinusOneHalvedP           []uint64 // ((q-1) / 2 ) + 1
	Mu                         uint64   // mu = 2^288 / q for 4.5 word barrett reduction
	RSquare                    []uint64
	One, Eleven, Thirteen      []uint64
	LegendreExponent           string // big.Int to base16 string
	NoCarry                    bool
	NoCarrySquare              bool // used if NoCarry is set, but some op may overflow in square optimization
	SqrtQ3Mod4                 bool
	SqrtAtkin                  bool
	SqrtTonelliShanks          bool
	SqrtSarkar                 bool
	SqrtE                      uint64
	SqrtS                      []uint64
	SqrtSarkarK                int
	SqrtSarkarL                []uint64
	SqrtAtkinExponent          string   // big.Int to base16 string
	SqrtSMinusOneOver2         string   // big.Int to base16 string
	SqrtQ3Mod4Exponent         string   // big.Int to base16 string
	SqrtQ3Mod4Exponent2        string   // big.Int to base16 string
	SqrtG                      []uint64 // NonResidue ^  SqrtR (montgomery form)
	NonResidue                 big.Int  // (montgomery form)
	LegendreExponentData       *addchain.AddChainData
	SqrtAtkinExponentData      *addchain.AddChainData
	SqrtSMinusOneOver2Data     *addchain.AddChainData
	SqrtQ3Mod4ExponentData     *addchain.AddChainData
	SqrtQ3Mod4ExponentData2    *addchain.AddChainData
	UseAddChain                bool

	// Cbrt pre computes
	CbrtQ2Mod3             bool     // q ≡ 2 (mod 3)
	CbrtQ1Mod3             bool     // q ≡ 1 (mod 3), need special handling
	CbrtQ7Mod9             bool     // q ≡ 7 (mod 9), use (q+2)/9 exponent
	CbrtQ4Mod9             bool     // q ≡ 4 (mod 9), use (2q+1)/9 exponent
	CbrtQ10Mod27           bool     // q ≡ 10 (mod 27), use (2q+7)/27 exponent + adjustment
	CbrtQ19Mod27           bool     // q ≡ 19 (mod 27), use (q+8)/27 exponent + adjustment
	CbrtE                  uint64   // such that q-1 = 3^CbrtE * CbrtS, with CbrtS not divisible by 3
	CbrtS                  []uint64 // CbrtS
	CbrtQ2Mod3Exponent     string   // (2q-1)/3 for q ≡ 2 (mod 3)
	CbrtQPlus2Div9         string   // (q+2)/9 for q ≡ 7 (mod 9)
	Cbrt2QPlus1Div9        string   // (2q+1)/9 for q ≡ 4 (mod 9)
	Cbrt2QPlus7Div27       string   // (2q+7)/27 for q ≡ 10 (mod 27)
	CbrtQPlus8Div27        string   // (q+8)/27 for q ≡ 19 (mod 27)
	CbrtSPlus1Div3         string   // (CbrtS+1)/3 for q ≡ 1 (mod 3) with CbrtS ≡ 2 (mod 3)
	CbrtSMinus1Div3        string   // (CbrtS-1)/3 for q ≡ 1 (mod 3) with CbrtS ≡ 1 (mod 3)
	CbrtG                  []uint64 // NonCubicResidue ^ CbrtS (montgomery form) -- primitive 3^CbrtE root of unity (ζ)
	CbrtG2                 []uint64 // CbrtG squared (montgomery form) -- ζ² for adjustment
	ThirdRootOne           []uint64 // CbrtG cubed (montgomery form) -- primitive 3rd root of unity (ω = ζ³)
	ThirdRootOneSquare     []uint64 // ThirdRootOne squared (montgomery form) -- ω² = ζ⁶
	NonCubicResidue        big.Int  // (montgomery form)
	CbrtQ2Mod3ExponentData *addchain.AddChainData
	CbrtQPlus2Div9Data     *addchain.AddChainData
	Cbrt2QPlus1Div9Data    *addchain.AddChainData
	Cbrt2QPlus7Div27Data   *addchain.AddChainData
	CbrtQPlus8Div27Data    *addchain.AddChainData
	CbrtSPlus1Div3Data     *addchain.AddChainData
	CbrtSMinus1Div3Data    *addchain.AddChainData

	// Torus cbrt helper (for E2 cbrt via algebraic torus)
	CbrtTorusHelperData *addchain.AddChainData
	CbrtTorusHelperName string // e.g. "CbrtHelperQMinus19Div27"

	Word Word // 32 iff Q < 2^32, else 64
	F31  bool // 31 bits field

	// Montgomery radix: R = 2^RBits.
	//
	// Defaults to NbWords*Word.BitSize, which is what every word-aligned field
	// uses (2^32 for F31, 2^64*NbWords otherwise). A field may override it via
	// WithMontgomeryRadixBits when the radix is deliberately narrower than the
	// storage word, as for AVX-512IFMA fields whose products go through
	// VPMADD52 and therefore want R = 2^52.
	//
	// Not to be confused with QRadix52 below, which is a radix-52 *limb
	// decomposition* of a 4-word modulus whose R is still 2^256.
	RBits uint
	RMask uint64 // (1<<RBits)-1; only meaningful when RadixSubWord

	// RadixSubWord is true when RBits is not a whole number of words, i.e. the
	// Montgomery shift does not line up with a word boundary, so the reduction
	// cannot be expressed as the usual word-aligned CIOS/no-carry loop.
	// Implies NbWords == 1 today.
	RadixSubWord bool

	// SingleWordSmall is true when the element fits in one word with at least
	// one spare bit (NbWords == 1 && NbBits < Word.BitSize). True for koalabear,
	// babybear and mamabear; false for goldilocks, whose modulus fills the word.
	// This is the property several templates were really testing when they
	// tested F31.
	SingleWordSmall bool

	// SparsePrime is set when q == 2^PHiBit - 2^PLoBit + 1, which admits a
	// cheap reduction using 2^PHiBit = 2^PLoBit - 1 (mod q).
	SparsePrime    bool
	PHiBit, PLoBit uint

	// asm code generation
	GenerateOpsAMD64       bool
	GenerateOpsARM64       bool
	GenerateVectorOpsAMD64 bool
	GenerateVectorOpsARM64 bool

	ASMPackagePath string

	// IFMA (AVX-512) specific constants for 4-word fields
	// These are used for radix-52 Montgomery multiplication
	QRadix52       []uint64 // q in radix-52 form (5 limbs for 4-word fields)
	QInvNeg52      uint64   // low 52 bits of qInvNeg (for IFMA Montgomery)
	MuBarrett52    uint64   // Barrett constant for reducing from [0, 32q) to [0, q)
	BarrettShift52 int      // Shift amount for Barrett reduction (typically 58)
}

type Word struct {
	BitSize   int    // 32 or 64
	ByteSize  int    // 4 or 8
	TypeLower string // uint32 or uint64
	TypeUpper string // Uint32 or Uint64
	Add       string // Add64 or Add32
	Sub       string // Sub64 or Sub32
	Len       string // Len64 or Len32
}

// FieldOption customizes a Field before its derived constants are computed.
type FieldOption func(*Field)

// WithMontgomeryRadixBits overrides the Montgomery radix, so that R = 2^n
// instead of the default 2^(NbWords*Word.BitSize).
//
// Only single-word fields with n < 64 are supported; NewFieldConfig returns an
// error otherwise.
func WithMontgomeryRadixBits(n uint) FieldOption {
	return func(f *Field) { f.RBits = n }
}

// NewFieldConfig returns a data structure with needed information to generate apis for field element
//
// See field/generator package
func NewFieldConfig(packageName, elementName, modulus string, useAddChain bool, opts ...FieldOption) (*Field, error) {
	// parse modulus
	var bModulus big.Int
	if _, ok := bModulus.SetString(modulus, 0); !ok {
		return nil, errParseModulus
	}
	if bModulus.Cmp(big.NewInt(3)) < 0 || bModulus.Bit(0) == 0 {
		return nil, errInvalidModulus
	}

	// field info
	F := &Field{
		PackageName: packageName,
		ElementName: elementName,
		Modulus:     bModulus.Text(10),
		ModulusHex:  bModulus.Text(16),
		ModulusBig:  new(big.Int).Set(&bModulus),
		UseAddChain: useAddChain,
	}
	// pre compute field constants
	F.NbBits = bModulus.BitLen()

	F.F31 = F.NbBits <= 31
	F.NbWords = len(bModulus.Bits())
	F.NbWordsLastIndex = F.NbWords - 1

	// set q from big int repr
	F.Q = toUint64Slice(&bModulus)
	F.IsMSWSaturated = F.Q[len(F.Q)-1] == math.MaxUint64
	_qHalved := big.NewInt(0)
	bOne := new(big.Int).SetUint64(1)
	_qHalved.Sub(&bModulus, bOne).Rsh(_qHalved, 1).Add(_qHalved, bOne)
	F.QMinusOneHalvedP = toUint64Slice(_qHalved, F.NbWords)

	// Word size; we pick uint32 only if the modulus is less than 2^32
	F.Word.BitSize = 64
	F.Word.ByteSize = 8
	F.Word.TypeLower = "uint64"
	F.Word.TypeUpper = "Uint64"
	F.Word.Add = "Add64"
	F.Word.Sub = "Sub64"
	F.Word.Len = "Len64"
	if F.F31 {
		F.Word.BitSize = 32
		F.Word.ByteSize = 4
		F.Word.TypeLower = "uint32"
		F.Word.TypeUpper = "Uint32"
		F.Word.Add = "Add32"
		F.Word.Sub = "Sub32"
		F.Word.Len = "Len32"
	}

	F.NbBytes = F.NbWords * F.Word.ByteSize

	// Montgomery radix. The default lines the radix up with the storage words,
	// which is what every field did before RBits existed; options may narrow it.
	wordAlignedRBits := uint(F.NbWords) * uint(F.Word.BitSize)
	F.RBits = wordAlignedRBits
	for _, o := range opts {
		o(F)
	}
	F.RadixSubWord = F.RBits != wordAlignedRBits
	if F.RadixSubWord {
		if F.NbWords != 1 || F.RBits >= 64 || F.RBits <= uint(F.NbBits) {
			return nil, errUnsupportedRadix
		}
		F.RMask = (uint64(1) << F.RBits) - 1
	}
	F.SingleWordSmall = F.NbWords == 1 && F.NbBits < F.Word.BitSize

	// Sparse prime detection: q == 2^hi - 2^lo + 1.
	if hi, lo, ok := sparsePrimeExponents(&bModulus); ok {
		F.SparsePrime, F.PHiBit, F.PLoBit = true, hi, lo
	}

	//  setting qInverse
	_r := big.NewInt(1)
	_r.Lsh(_r, F.RBits)
	_rInv := big.NewInt(1)
	_qInv := big.NewInt(0)
	extendedEuclideanAlgo(_r, &bModulus, _rInv, _qInv)
	_qInv.Mod(_qInv, _r)
	F.QInverse = toUint64Slice(_qInv, F.NbWords)

	// Pornin20 inversion correction factors
	k := 32 // Optimized for 64 bit machines, still works for 32

	p20InvInnerLoopNbIterations := 2*F.NbBits - 1
	// if constant time inversion then p20InvInnerLoopNbIterations-- (among other changes)
	F.P20InversionNbIterations = (p20InvInnerLoopNbIterations-1)/(k-1) + 1 // ⌈ (2 * field size - 1) / (k-1) ⌉
	F.P20InversionNbIterations += F.P20InversionNbIterations % 2           // "round up" to a multiple of 2

	kLimbs := k * F.NbWords
	p20InversionCorrectiveFacPower := kLimbs*6 + F.P20InversionNbIterations*(kLimbs-k+1)
	p20InversionCorrectiveFac := big.NewInt(1)
	p20InversionCorrectiveFac.Lsh(p20InversionCorrectiveFac, uint(p20InversionCorrectiveFacPower))
	p20InversionCorrectiveFac.Mod(p20InversionCorrectiveFac, &bModulus)
	F.P20InversionCorrectiveFac = toUint64Slice(p20InversionCorrectiveFac, F.NbWords)

	{
		c := F.NbWords * 64
		// TODO @gbotrel check inverse performance for 32 bits
		F.UsingP20Inverse = F.NbWords > 1 && F.NbBits < c && F.Word.BitSize == 64
	}

	// rsquare
	_rSquare := big.NewInt(1)
	_rSquare.Lsh(_rSquare, F.RBits*2).Mod(_rSquare, &bModulus)
	F.RSquare = toUint64Slice(_rSquare, F.NbWords)

	var one big.Int
	one.SetUint64(1)
	one.Lsh(&one, F.RBits).Mod(&one, &bModulus)
	F.One = toUint64Slice(&one, F.NbWords)

	{
		var n big.Int
		n.SetUint64(11)
		n.Lsh(&n, F.RBits).Mod(&n, &bModulus)
		F.Eleven = toUint64Slice(&n, F.NbWords)
	}

	{
		var n big.Int
		n.SetUint64(13)
		n.Lsh(&n, F.RBits).Mod(&n, &bModulus)
		F.Thirteen = toUint64Slice(&n, F.NbWords)
	}

	// indexes (template helpers)
	F.NbWordsIndexesFull = make([]int, F.NbWords)
	for i := range F.NbWords {
		F.NbWordsIndexesFull[i] = i
	}
	F.NbWordsIndexesNoZero = F.NbWordsIndexesFull[1:]
	if F.NbWords >= 2 {
		F.NbWordsIndexesNoZeroNoLast = F.NbWordsIndexesFull[1 : F.NbWords-1]
	}

	// See https://hackmd.io/@gnark/modular_multiplication
	// if the last word of the modulus is smaller or equal to B,
	// we can simplify the montgomery multiplication
	const B = (^uint64(0) >> 1) - 1
	F.NoCarry = (F.Q[len(F.Q)-1] <= B) && F.NbWords <= 12
	const BSquare = ^uint64(0) >> 2
	F.NoCarrySquare = F.Q[len(F.Q)-1] <= BSquare

	// Legendre exponent (p-1)/2
	var legendreExponent big.Int
	legendreExponent.SetUint64(1)
	legendreExponent.Sub(&bModulus, &legendreExponent)
	legendreExponent.Rsh(&legendreExponent, 1)
	F.LegendreExponent = legendreExponent.Text(16)
	if F.UseAddChain {
		F.LegendreExponentData = addchain.GetAddChain(&legendreExponent)
	}

	// Sqrt pre computes
	var qMod big.Int
	qMod.SetUint64(4)
	if qMod.Mod(&bModulus, &qMod).Cmp(new(big.Int).SetUint64(3)) == 0 {
		// q ≡ 3 (mod 4)
		// using  z ≡ ± x^((p+1)/4) (mod q)
		F.SqrtQ3Mod4 = true
		var sqrtExponent, sqrtExponent2 big.Int
		sqrtExponent.SetUint64(1)
		sqrtExponent.Add(&bModulus, &sqrtExponent)
		sqrtExponent.Rsh(&sqrtExponent, 2)
		sqrtExponent2.SetUint64(3)
		sqrtExponent2.Sub(&bModulus, &sqrtExponent2)
		sqrtExponent2.Rsh(&sqrtExponent2, 2)
		F.SqrtQ3Mod4Exponent = sqrtExponent.Text(16)
		F.SqrtQ3Mod4Exponent2 = sqrtExponent2.Text(16)

		// add chain stuff
		if F.UseAddChain {
			F.SqrtQ3Mod4ExponentData = addchain.GetAddChain(&sqrtExponent)
			F.SqrtQ3Mod4ExponentData2 = addchain.GetAddChain(&sqrtExponent2)
		}

	} else {
		// q ≡ 1 (mod 4)
		qMod.SetUint64(8)
		if qMod.Mod(&bModulus, &qMod).Cmp(new(big.Int).SetUint64(5)) == 0 {
			// q ≡ 5 (mod 8)
			// use Atkin's algorithm
			// see modSqrt5Mod8Prime in math/big/int.go
			F.SqrtAtkin = true
			e := new(big.Int).Rsh(&bModulus, 3) // e = (q - 5) / 8
			F.SqrtAtkinExponent = e.Text(16)
			if F.UseAddChain {
				F.SqrtAtkinExponentData = addchain.GetAddChain(e)
			}
		} else {
			// use Tonelli-Shanks (and optionally Sarkar for high 2-adicity)
			F.SqrtTonelliShanks = true

			// Write q-1 =2ᵉ * s , s odd
			var s big.Int
			one.SetUint64(1)
			s.Sub(&bModulus, &one)

			e := s.TrailingZeroBits()
			s.Rsh(&s, e)
			F.SqrtE = uint64(e)
			F.SqrtS = toUint64Slice(&s)

			// find non residue
			var nonResidue big.Int
			nonResidue.SetInt64(2)
			one.SetUint64(1)
			for big.Jacobi(&nonResidue, &bModulus) != -1 {
				nonResidue.Add(&nonResidue, &one)
			}

			// g = nonresidue ^ s
			var g big.Int
			g.Exp(&nonResidue, &s, &bModulus)
			// store g in montgomery form
			g.Lsh(&g, F.RBits).Mod(&g, &bModulus)
			F.SqrtG = toUint64Slice(&g, F.NbWords)

			// store non residue in montgomery form
			F.NonResidue = F.ToMont(nonResidue)

			// (s+1) /2
			s.Sub(&s, &one).Rsh(&s, 1)
			F.SqrtSMinusOneOver2 = s.Text(16)

			if F.UseAddChain {
				F.SqrtSMinusOneOver2Data = addchain.GetAddChain(&s)
			}

			// We use Sarkar when:
			// - the field has high 2-adicity. Otherwise, due to smaller constants Tonelli-Shanks is more efficient
			// - the field is not a small field (i.e. not F31 nor goldilocks-like 64-bit fields)
			//
			// Sub-word radix fields are the exception to "small": they are one
			// word wide but can carry a 2-adicity high enough that the
			// Tonelli-Shanks loop dominates (mamabear: 34).
			if e >= 10 && !F.F31 && (F.NbBits > 64 || F.RadixSubWord) {
				F.SqrtSarkar = true
				F.SqrtSarkarK, F.SqrtSarkarL = chooseSarkarParams(int(e))
			}
		}
	}

	// Cbrt pre computes
	// Check if q ≡ 1 (mod 3) or q ≡ 2 (mod 3)
	var qMod3 big.Int
	qMod3.SetUint64(3)
	qMod3.Mod(&bModulus, &qMod3)

	if qMod3.Cmp(new(big.Int).SetUint64(2)) == 0 {
		// q ≡ 2 (mod 3)
		// using z = x^((2q-1)/3) (mod q)
		F.CbrtQ2Mod3 = true
		var cbrtExponent big.Int
		cbrtExponent.Mul(&bModulus, big.NewInt(2))
		cbrtExponent.Sub(&cbrtExponent, big.NewInt(1))
		cbrtExponent.Div(&cbrtExponent, big.NewInt(3))
		F.CbrtQ2Mod3Exponent = cbrtExponent.Text(16)
		if F.UseAddChain {
			F.CbrtQ2Mod3ExponentData = addchain.GetAddChain(&cbrtExponent)
		}
	} else if qMod3.Cmp(big.NewInt(1)) == 0 {
		// q ≡ 1 (mod 3)
		// use Tonelli-Shanks variant for cube roots
		F.CbrtQ1Mod3 = true

		// Write q-1 = 3^e * s, where s is not divisible by 3
		var s big.Int
		one.SetUint64(1)
		s.Sub(&bModulus, &one)

		// Count the power of 3 in q-1
		three := big.NewInt(3)
		nine := big.NewInt(9)
		e := uint64(0)
		var remainder big.Int
		for {
			remainder.Mod(&s, three)
			if remainder.Sign() != 0 {
				break
			}
			s.Div(&s, three)
			e++
		}
		F.CbrtE = e
		F.CbrtS = toUint64Slice(&s)

		// Check q mod 9 and q mod 27 for optimized exponentiation
		// Reference: Lemma 3 of https://eprint.iacr.org/2021/1446.pdf
		var qMod9, qMod27 big.Int
		qMod9.Mod(&bModulus, nine)
		twentySeven := big.NewInt(27)
		qMod27.Mod(&bModulus, twentySeven)

		if e == 1 && qMod9.Cmp(big.NewInt(7)) == 0 {
			// q ≡ 7 (mod 9): use cbrt(x) = x^((q+2)/9)
			F.CbrtQ7Mod9 = true
			var exp big.Int
			exp.Add(&bModulus, big.NewInt(2))
			exp.Div(&exp, nine)
			F.CbrtQPlus2Div9 = exp.Text(16)
			if F.UseAddChain {
				F.CbrtQPlus2Div9Data = addchain.GetAddChain(&exp)
			}
		} else if e == 1 && qMod9.Cmp(big.NewInt(4)) == 0 {
			// q ≡ 4 (mod 9): use cbrt(x) = x^((2q+1)/9)
			F.CbrtQ4Mod9 = true
			var exp big.Int
			exp.Mul(&bModulus, big.NewInt(2))
			exp.Add(&exp, big.NewInt(1))
			exp.Div(&exp, nine)
			F.Cbrt2QPlus1Div9 = exp.Text(16)
			if F.UseAddChain {
				F.Cbrt2QPlus1Div9Data = addchain.GetAddChain(&exp)
			}
		} else if e == 2 && qMod27.Cmp(big.NewInt(10)) == 0 {
			// q ≡ 10 (mod 27): use cbrt(x) = x^((2q+7)/27) * ζ^k
			// where ζ is a primitive 9th root of unity and k ∈ {0, 1, 2}
			F.CbrtQ10Mod27 = true
			var exp big.Int
			exp.Mul(&bModulus, big.NewInt(2))
			exp.Add(&exp, big.NewInt(7))
			exp.Div(&exp, twentySeven)
			F.Cbrt2QPlus7Div27 = exp.Text(16)
			if F.UseAddChain {
				F.Cbrt2QPlus7Div27Data = addchain.GetAddChain(&exp)
			}
		} else if e == 2 && qMod27.Cmp(big.NewInt(19)) == 0 {
			// q ≡ 19 (mod 27): use cbrt(x) = x^((q+8)/27) * ζ^k
			// where ζ is a primitive 9th root of unity and k ∈ {0, 1, 2}
			F.CbrtQ19Mod27 = true
			var exp big.Int
			exp.Add(&bModulus, big.NewInt(8))
			exp.Div(&exp, twentySeven)
			F.CbrtQPlus8Div27 = exp.Text(16)
			if F.UseAddChain {
				F.CbrtQPlus8Div27Data = addchain.GetAddChain(&exp)
			}
		}

		// find non-cubic residue (element g such that g^((q-1)/3) ≠ 1)
		// Only needed for e >= 2 (Tonelli-Shanks adjustment)
		var nonCubicResidue, qMinus1Over3, test big.Int
		qMinus1Over3.Sub(&bModulus, big.NewInt(1))
		qMinus1Over3.Div(&qMinus1Over3, three)
		nonCubicResidue.SetInt64(2)
		for {
			test.Exp(&nonCubicResidue, &qMinus1Over3, &bModulus)
			if test.Cmp(big.NewInt(1)) != 0 {
				break
			}
			nonCubicResidue.Add(&nonCubicResidue, big.NewInt(1))
		}

		// g = nonCubicResidue ^ s (primitive 3^e root of unity, ζ)
		var g big.Int
		g.Exp(&nonCubicResidue, &s, &bModulus)

		// Precompute related constants (compute in standard form first, then convert to montgomery):
		// ζ² = g² (for adjustment y * ζ²)
		var g2 big.Int
		g2.Mul(&g, &g).Mod(&g2, &bModulus)

		// ζ³ = g³
		var g3 big.Int
		g3.Mul(&g2, &g).Mod(&g3, &bModulus)

		// ζ⁶ = (ζ³)²
		var g6 big.Int
		g6.Mul(&g3, &g3).Mod(&g6, &bModulus)

		// ω = ζ⁶ = primitive 3rd root of unity (matches thirdRootOneG1 convention)
		// ω² = ζ³ (matches thirdRootOneG2 = thirdRootOneG1²)
		// Note: ζ⁶ and ζ³ are the two primitive 3rd roots of unity (since ζ⁹ = 1)

		// Convert all to montgomery form and store
		g.Lsh(&g, F.RBits).Mod(&g, &bModulus)
		F.CbrtG = toUint64Slice(&g, F.NbWords)

		g2.Lsh(&g2, F.RBits).Mod(&g2, &bModulus)
		F.CbrtG2 = toUint64Slice(&g2, F.NbWords)

		// ThirdRootOne = ζ⁶ (matches thirdRootOneG1)
		g6.Lsh(&g6, F.RBits).Mod(&g6, &bModulus)
		F.ThirdRootOne = toUint64Slice(&g6, F.NbWords)

		// ThirdRootOneSquare = ζ³ (matches thirdRootOneG2 = thirdRootOneG1²)
		g3.Lsh(&g3, F.RBits).Mod(&g3, &bModulus)
		F.ThirdRootOneSquare = toUint64Slice(&g3, F.NbWords)

		// store non-cubic residue in montgomery form
		F.NonCubicResidue = F.ToMont(nonCubicResidue)

		// Compute exponent based on s mod 3 (for e >= 2 cases)
		// If s ≡ 1 (mod 3): use (s-1)/3
		// If s ≡ 2 (mod 3): use (s+1)/3
		var sMod3 big.Int
		sMod3.Mod(&s, three)
		if sMod3.Cmp(big.NewInt(1)) == 0 {
			// s ≡ 1 (mod 3)
			var exp big.Int
			exp.Sub(&s, big.NewInt(1))
			exp.Div(&exp, three)
			F.CbrtSMinus1Div3 = exp.Text(16)
			if F.UseAddChain {
				F.CbrtSMinus1Div3Data = addchain.GetAddChain(&exp)
			}
		} else {
			// s ≡ 2 (mod 3)
			var exp big.Int
			exp.Add(&s, big.NewInt(1))
			exp.Div(&exp, three)
			F.CbrtSPlus1Div3 = exp.Text(16)
			if F.UseAddChain {
				F.CbrtSPlus1Div3Data = addchain.GetAddChain(&exp)
			}
		}

		// Torus cbrt helper exponent: norm^((q-r)/d) for shared cbrt + inverse
		if useAddChain {
			qMod9Val := qMod9.Uint64()
			qMod27Val := qMod27.Uint64()
			var helperExp big.Int
			switch qMod9Val {
			case 7:
				helperExp.Sub(&bModulus, big.NewInt(7)).Div(&helperExp, big.NewInt(9))
				F.CbrtTorusHelperName = "CbrtHelperQMinus7Div9"
			case 4:
				helperExp.Sub(&bModulus, big.NewInt(4)).Div(&helperExp, big.NewInt(9))
				F.CbrtTorusHelperName = "CbrtHelperQMinus4Div9"
			default:
				switch qMod27Val {
				case 10:
					helperExp.Sub(&bModulus, big.NewInt(10)).Div(&helperExp, big.NewInt(27))
					F.CbrtTorusHelperName = "CbrtHelperQMinus10Div27"
				case 19:
					helperExp.Sub(&bModulus, big.NewInt(19)).Div(&helperExp, big.NewInt(27))
					F.CbrtTorusHelperName = "CbrtHelperQMinus19Div27"
				}
			}
			if F.CbrtTorusHelperName != "" {
				F.CbrtTorusHelperData = addchain.GetAddChain(&helperExp)
			}
		}
	}
	// The only prime with q ≡ 0 (mod 3) is q = 3. There is no generated Cbrt
	// specialization for that degenerate field, so leave the Cbrt flags unset.

	// note: to simplify output files generated, we generated ASM code only for
	// moduli that meet the condition F.NoCarry
	// asm code generation for moduli with more than 6 words can be optimized further
	f31ASM := F.F31 && F.NbBits == 31
	F.GenerateOpsAMD64 = f31ASM || (F.NoCarry && F.NbWords <= 12 && F.NbWords > 1)
	if F.NbWords == 4 && F.GenerateOpsAMD64 && F.NbBits <= 226 {
		// 4 words field with 226 bits or less have no vector ops
		// for now since we generate both in same file we disable
		// TODO @gbotrel
		F.GenerateOpsAMD64 = false
	}
	F.GenerateVectorOpsAMD64 = f31ASM || (F.GenerateOpsAMD64 && F.NbWords == 4 && F.NbBits > 226)
	F.GenerateOpsARM64 = f31ASM || (F.GenerateOpsAMD64 && (F.NbWords%2 == 0))
	F.GenerateVectorOpsARM64 = f31ASM

	// setting Mu 2^288 / q
	if F.GenerateVectorOpsAMD64 && F.NbWords == 4 {
		_mu := big.NewInt(1)
		_mu.Lsh(_mu, 288)
		_mu.Div(_mu, &bModulus)
		muSlice := toUint64Slice(_mu, F.NbWords)
		F.Mu = muSlice[0]

		// IFMA constants for AVX-512 radix-52 Montgomery multiplication
		// q in radix-52 form (5 limbs)
		// l0 = q & mask52, l1 = (q >> 52) & mask52, etc.
		const mask52 = (1 << 52) - 1
		F.QRadix52 = make([]uint64, 5)
		qCopy := new(big.Int).Set(&bModulus)
		for i := range 5 {
			F.QRadix52[i] = qCopy.Uint64() & mask52
			qCopy.Rsh(qCopy, 52)
		}

		// qInvNeg52 = qInvNeg & mask52 (low 52 bits of qInvNeg for IFMA)
		F.QInvNeg52 = F.QInverse[0] & mask52

		// Barrett constant for reducing from [0, 32q) to [0, q)
		// After IFMA Montgomery (R=2^260), we multiply by 16 to correct to R=2^256
		// This gives result in [0, 32q), which we reduce using Barrett reduction
		// mu = floor(2^58 / (q >> 208)), shift = 58
		// This allows: k = (l4 * mu) >> 58, then subtract k*q
		F.BarrettShift52 = 58
		qHigh := new(big.Int).Rsh(&bModulus, 208)
		muBig := new(big.Int).Lsh(big.NewInt(1), uint(F.BarrettShift52))
		muBig.Div(muBig, qHigh)
		F.MuBarrett52 = muBig.Uint64()
	}
	if f31ASM {
		F.Mu = -F.QInverse[0]
	}

	return F, nil
}

func chooseSarkarParams(n int) (int, []uint64) {
	bestCost := -1
	bestK := 1
	var bestL []uint64

	for k := 1; k < n; k++ {
		l := int(math.Ceil(float64(n-1) / float64(k)))
		k2 := (n - 1) - (l-1)*k
		if k2 < 1 || k2 > k {
			continue
		}
		k1 := k - k2
		lvals := make([]uint64, 0, k)
		for range k1 {
			lvals = append(lvals, uint64(l-1))
		}
		for range k2 {
			lvals = append(lvals, uint64(l))
		}

		sq := (n - int(lvals[0]))
		for _, li := range lvals {
			sq += int(li*(li-1)) / 2
		}
		mul := n + k + 1 + (k*(k-1))/2
		for i := 0; i < k; i++ {
			mul += (k - i) * (int(lvals[i]) - 1)
		}

		cost := sq + mul
		if bestCost == -1 || cost < bestCost {
			bestCost = cost
			bestK = k
			bestL = lvals
		}
	}

	return bestK, bestL
}

func toUint64Slice(b *big.Int, nbWords ...int) (s []uint64) {
	if len(nbWords) > 0 && nbWords[0] > len(b.Bits()) {
		s = make([]uint64, nbWords[0])
	} else {
		s = make([]uint64, len(b.Bits()))
	}

	for i, v := range b.Bits() {
		s[i] = (uint64)(v)
	}
	return
}

// https://en.wikipedia.org/wiki/Extended_Euclidean_algorithm
// r > q, modifies rinv and qinv such that rinv.r - qinv.q = 1
func extendedEuclideanAlgo(r, q, rInv, qInv *big.Int) {
	var s1, s2, t1, t2, qi, tmpMuls, riPlusOne, tmpMult, a, b big.Int
	t1.SetUint64(1)
	rInv.Set(big.NewInt(1))
	qInv.Set(big.NewInt(0))
	a.Set(r)
	b.Set(q)

	// r_i+1 = r_i-1 - q_i.r_i
	// s_i+1 = s_i-1 - q_i.s_i
	// t_i+1 = t_i-1 - q_i.s_i
	for b.Sign() > 0 {
		qi.Div(&a, &b)
		riPlusOne.Mod(&a, &b)

		tmpMuls.Mul(&s1, &qi)
		tmpMult.Mul(&t1, &qi)

		s2.Set(&s1)
		t2.Set(&t1)

		s1.Sub(rInv, &tmpMuls)
		t1.Sub(qInv, &tmpMult)
		rInv.Set(&s2)
		qInv.Set(&t2)

		a.Set(&b)
		b.Set(&riPlusOne)
	}
	qInv.Neg(qInv)
}

// StringToMont takes an element written in string form, and returns it in Montgomery form
// Useful for hard-coding in implementation field elements from standards documents
func (f *Field) StringToMont(str string) big.Int {

	var i big.Int
	i.SetString(str, 0)
	i = f.ToMont(i)

	return i
}

func (f *Field) ToMont(nonMont big.Int) big.Int {
	var mont big.Int
	mont.Lsh(&nonMont, f.RBits)
	mont.Mod(&mont, f.ModulusBig)
	return mont
}

func (f *Field) FromMont(nonMont *big.Int, mont *big.Int) *Field {

	if f.NbWords == 0 {
		nonMont.SetInt64(0)
		return f
	}
	f.halve(nonMont, mont)
	for i := 1; i < f.NbWords*f.Word.BitSize; i++ {
		f.halve(nonMont, nonMont)
	}

	return f
}

func (f *Field) Exp(res *big.Int, x *big.Int, pow *big.Int) *Field {
	res.SetInt64(1)

	for i := pow.BitLen() - 1; ; {

		if pow.Bit(i) == 1 {
			res.Mul(res, x)
		}

		if i == 0 {
			break
		}
		i--

		res.Mul(res, res).Mod(res, f.ModulusBig)
	}

	res.Mod(res, f.ModulusBig)
	return f
}

func (f *Field) halve(res *big.Int, x *big.Int) {
	var z big.Int
	if x.Bit(0) == 0 {
		z.Set(x)
	} else {
		z.Add(x, f.ModulusBig)
	}
	res.Rsh(&z, 1)
}

func (f *Field) Mul(z *big.Int, x *big.Int, y *big.Int) *Field {
	z.Mul(x, y).Mod(z, f.ModulusBig)
	return f
}

func (f *Field) Add(z *big.Int, x *big.Int, y *big.Int) *Field {
	z.Add(x, y).Mod(z, f.ModulusBig)
	return f
}

func (f *Field) ToMontSlice(x []big.Int) []big.Int {
	z := make(Element, len(x))
	for i := range x {
		z[i] = f.ToMont(x[i])
	}
	return z
}

// TODO: Spaghetti Alert: Okay to have codegen functions here?
func CoordNameForExtensionDegree(degree uint8) string {
	switch degree {
	case 1:
		return ""
	case 2:
		return "A"
	case 6:
		return "B"
	case 12:
		return "C"
	}
	panic(fmt.Sprint("unknown extension degree", degree))
}

func (f *Field) WriteElement(element Element) string {
	var builder strings.Builder

	builder.WriteString("{")
	length := len(element)
	var subElementNames string
	if length > 1 {
		builder.WriteString("\n")
		subElementNames = CoordNameForExtensionDegree(uint8(length))
	}
	for i, e := range element {
		if length > 1 {
			builder.WriteString(subElementNames)
			builder.WriteString(strconv.Itoa(i))
			builder.WriteString(": fp.Element{")
		}
		mont := f.ToMont(e)
		bavard.WriteBigIntAsUint64Slice(&builder, &mont)
		if length > 1 {
			builder.WriteString("},\n")
		}
	}
	builder.WriteString("}")
	return builder.String()
}

type FieldDependency struct {
	ElementType      string
	FieldPackagePath string
	FieldPackageName string
}
