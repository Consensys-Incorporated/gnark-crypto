package config

import (
	"fmt"
	"math/big"

	field "github.com/consensys/gnark-crypto/internal/generator/field/config"
)

type FieldElementToCurvePoint string

const (
	SSWU FieldElementToCurvePoint = "SSWU"
	SVDW FieldElementToCurvePoint = "SVDW"
)

type Isogeny struct {

	//Isogeny to original curve
	XMap RationalPolynomial
	YMap RationalPolynomial // The y map is also evaluated on x. The result is multiplied by y.

	// Chains holds optional preprocessed multiplication chains for the four maps
	// (see PolyChain). Maps without a chain are evaluated with Horner's rule.
	Chains *IsogenyChains
}

// PolyChain is a straight-line multiplication chain evaluating one isogeny polynomial
// with floor(n/2)+1 multiplications (+1 when the leading coefficient is not 1) instead
// of Horner's n-1 (+1). It is produced offline from the polynomial's coefficients by
// internal/generator/hash_to_curve/gen_isogeny_chains.py, the decoder of
//
//	T. D. Ahle, "Fast Evaluation of Polynomials with Rational Preprocessing",
//	https://arxiv.org/abs/2609.06022
//
// Wire 0 is the input x and the output of gate i (0-based) is wire i+1. Every gate
// multiplies two linear forms in the wires; the polynomial is the output linear form,
// times Leading when the polynomial is not monic.
type PolyChain struct {
	Degree    int
	Leading   []string   // leading coefficient (one string per coordinate); nil when monic
	Constants [][]string // field constants referenced by LinearForm.Const
	Gates     []ChainGate
	Output    LinearForm
}

// ChainGate is the product of two linear forms.
type ChainGate struct {
	Left, Right LinearForm
}

// LinearForm is Σ coeff·wire + Constants[Const] (no constant when Const < 0).
// The integer coefficients are small (|coeff| ≤ 64) and are realised by doublings
// and additions in the generated code.
type LinearForm struct {
	Terms [][2]int // {wire, coefficient}
	Const int
}

// IsogenyChains are the chains of the four isogeny maps; a nil entry means Horner.
type IsogenyChains struct {
	XNum, XDen, YNum, YDen *PolyChain
}

type RationalPolynomial struct {
	Num [][]string //Num is stored
	Den [][]string //Den is stored. It is also monic. The leading coefficient (1) is omitted.
}

type HashSuite interface {
	GetInfo(baseField *field.Field, g *Point, name string) HashSuiteInfo
}

type HashSuiteSswu struct {
	//TODO: Move into Isogeny
	A []string // A is the Weierstrass curve coefficient of x in the isogenous curve over which the SSWU map is evaluated.
	B []string // B is the Weierstrass curve constant term in the isogenous curve over which the SSWU map is evaluated.

	Z       []int // z (or zeta) is a quadratic non-residue with //TODO: some extra nice properties, refer to WB19
	Isogeny *Isogeny
}

func toBigIntSlice(z []int) []big.Int {
	res := make([]big.Int, len(z))
	for i := range z {
		res[i].SetInt64(int64(z[i]))
	}
	return res
}

type HashSuiteSvdw struct {
	z  []string
	c1 []string
	c2 []string
	c3 []string
	c4 []string
}

func (parameters *HashSuiteSvdw) GetInfo(baseField *field.Field, g *Point, name string) HashSuiteInfo {
	f := field.NewTower(baseField, g.CoordExtDegree, g.CoordExtRoot)
	c := []field.Element{
		field.NewElement(parameters.z),
		field.NewElement(parameters.c1),
		field.NewElement(parameters.c2),
		field.NewElement(parameters.c3),
		field.NewElement(parameters.c4),
	}
	return HashSuiteInfo{
		PrecomputedParams: c,
		CofactorClearing:  g.CofactorCleaning,
		Point:             g,
		MappingAlgorithm:  SVDW,
		Name:              name,
		FieldCoordName:    field.CoordNameForExtensionDegree(g.CoordExtDegree),
		Field:             &f,
	}
}

func (suite *HashSuiteSswu) GetInfo(baseField *field.Field, g *Point, name string) HashSuiteInfo {

	f := field.NewTower(baseField, g.CoordExtDegree, g.CoordExtRoot)
	fieldSizeMod256 := uint8(f.Size.Bits()[0])

	Z := toBigIntSlice(suite.Z)
	var c []field.Element

	// Sarkar parameters
	var sarkarEnabled bool
	var sarkarN, sarkarK int
	var sarkarL []int
	var sarkarG field.Element

	if fieldSizeMod256%4 == 3 {
		c = make([]field.Element, 2)
		c[0] = make([]big.Int, 1)
		c[0][0].Rsh(&f.Size, 2)

		c[1] = f.Neg(Z)
		c[1] = f.Sqrt(c[1])

	} else if fieldSizeMod256%8 == 5 {
		c = make([]field.Element, 3)
		c[0] = make([]big.Int, 1)
		c[0][0].Rsh(&f.Size, 3)

		c[1] = make([]big.Int, f.Degree)
		c[1][0].SetInt64(-1)
		c[1] = f.Sqrt(c[1])

		c[2] = f.Inverse(c[1])
		c[2] = f.Mul(Z, c[2])
		c[2] = f.Sqrt(c[2])

	} else if fieldSizeMod256%8 == 1 {
		ONE := big.NewInt(1)
		c = make([]field.Element, 3)

		c[0] = make([]big.Int, 5)
		// c1 .. c5 stored as c[0][0] .. c[0][4]
		c[0][0].Sub(&f.Size, big.NewInt(1))
		c1 := c[0][0].TrailingZeroBits()
		c[0][0].SetUint64(uint64(c1))

		var twoPowC1 big.Int
		twoPowC1.Lsh(ONE, c1)
		c[0][1].Rsh(&f.Size, c1)
		c[0][2].Rsh(&c[0][1], 1)
		c[0][3].Sub(&twoPowC1, ONE)
		c[0][4].Rsh(&twoPowC1, 1)

		// c6, c7 stored as c[1], c[2] respectively
		c[1] = f.Exp(Z, &c[0][1])

		var c7Pow big.Int
		c7Pow.Add(&c[0][1], ONE)
		c7Pow.Rsh(&c7Pow, 1)
		c[2] = f.Exp(Z, &c7Pow)

		// Enable Sarkar when 2-adicity e >= 10 and field is Fp (not extension fields)
		// Extension fields (E2, E4, etc.) don't have ExpBySqrtExp method
		e := int(c1)
		if e >= 10 && g.CoordExtDegree == 1 {
			sarkarEnabled = true
			sarkarN = e
			sarkarK, sarkarL = computeSarkarBlockSizes(e)
			// sarkarG = Z^m where m = (q-1)/2^e (odd part)
			sarkarG = c[1] // c[1] = Z^m = c6
		}

	} else {
		panic("this is logically impossible")
	}

	return HashSuiteInfo{
		A:                 field.NewElement(suite.A),
		B:                 field.NewElement(suite.B),
		Z:                 Z,
		Point:             g,
		CofactorClearing:  g.CofactorCleaning,
		Name:              name,
		Isogeny:           newIsogenousCurveInfoOptional(suite.Isogeny),
		FieldSizeMod256:   fieldSizeMod256,
		PrecomputedParams: c,
		Field:             &f,
		FieldCoordName:    field.CoordNameForExtensionDegree(g.CoordExtDegree),
		MappingAlgorithm:  SSWU,
		SarkarEnabled:     sarkarEnabled,
		SarkarN:           sarkarN,
		SarkarK:           sarkarK,
		SarkarL:           sarkarL,
		SarkarG:           sarkarG,
	}
}

func stringMatrixToIntMatrix(s [][]string) []field.Element {
	res := make([]field.Element, len(s))
	for i, S := range s {
		res[i] = field.NewElement(S)
	}
	return res
}

func newIsogenousCurveInfoOptional(isogenousCurve *Isogeny) *IsogenyInfo {
	if isogenousCurve == nil {
		return nil
	}
	info := &IsogenyInfo{
		XMap: RationalPolynomialInfo{
			stringMatrixToIntMatrix(isogenousCurve.XMap.Num),
			stringMatrixToIntMatrix(isogenousCurve.XMap.Den),
		},
		YMap: RationalPolynomialInfo{
			stringMatrixToIntMatrix(isogenousCurve.YMap.Num),
			stringMatrixToIntMatrix(isogenousCurve.YMap.Den),
		},
		Chains: &IsogenyChainsInfo{},
	}
	if c := isogenousCurve.Chains; c != nil {
		info.Chains.XNum = newPolyChainInfo(c.XNum, len(isogenousCurve.XMap.Num)-1)
		info.Chains.XDen = newPolyChainInfo(c.XDen, len(isogenousCurve.XMap.Den))
		info.Chains.YNum = newPolyChainInfo(c.YNum, len(isogenousCurve.YMap.Num)-1)
		info.Chains.YDen = newPolyChainInfo(c.YDen, len(isogenousCurve.YMap.Den))
	}
	return info
}

func newPolyChainInfo(c *PolyChain, degree int) *PolyChainInfo {
	if c == nil {
		return nil
	}
	if c.Degree != degree {
		panic(fmt.Sprintf("isogeny chain of degree %d for a polynomial of degree %d", c.Degree, degree))
	}
	info := &PolyChainInfo{
		Degree:    c.Degree,
		Constants: stringMatrixToIntMatrix(c.Constants),
		Gates:     c.Gates,
		Output:    c.Output,
	}
	if c.Leading != nil {
		info.Leading = field.NewElement(c.Leading)
	}
	return info
}

// PolyChainInfo is PolyChain with the constants parsed into field elements.
type PolyChainInfo struct {
	Degree    int
	Leading   field.Element // nil when monic
	Constants []field.Element
	Gates     []ChainGate
	Output    LinearForm
}

// Multiplications is the number of field multiplications of the chain.
func (c *PolyChainInfo) Multiplications() int {
	if c.Leading != nil {
		return len(c.Gates) + 1
	}
	return len(c.Gates)
}

// HornerMultiplications is the number of field multiplications Horner's rule uses
// for the same polynomial.
func (c *PolyChainInfo) HornerMultiplications() int {
	if c.Leading != nil {
		return c.Degree
	}
	return c.Degree - 1
}

// IsogenyChainsInfo mirrors IsogenyChains; a nil entry means Horner.
type IsogenyChainsInfo struct {
	XNum, XDen, YNum, YDen *PolyChainInfo
}

// computeSarkarBlockSizes computes optimal block sizes for Sarkar's algorithm.
// Given 2-adicity e, we split e-1 into k blocks of sizes L[i], targeting block size ~7.
// Returns k (number of blocks) and L (block sizes).
func computeSarkarBlockSizes(e int) (int, []int) {
	total := e - 1
	k := max(
		// ceiling division, target block size ~7
		(total+6)/7, 2)
	base := total / k
	extra := total % k
	L := make([]int, k)
	// Distribute sizes: first (k-extra) blocks get 'base', remaining get 'base+1'
	for i := 0; i < k-extra; i++ {
		L[i] = base
	}
	for i := k - extra; i < k; i++ {
		L[i] = base + 1
	}
	return k, L
}

type IsogenyInfo struct {
	XMap   RationalPolynomialInfo
	YMap   RationalPolynomialInfo // The y map is also evaluated on x. The result is multiplied by y.
	Chains *IsogenyChainsInfo     // never nil; its entries are nil for maps evaluated with Horner
}

type RationalPolynomialInfo struct {
	Num []field.Element
	Den []field.Element //denominator is monic. The leading coefficient (1) is omitted.
}

type HashSuiteInfo struct {
	//Isogeny to original curve
	Isogeny *IsogenyInfo //pointer so it's nullable.

	A []big.Int //TODO: Move inside IsogenyInfo
	B []big.Int

	Point             *Point
	Field             *field.Extension
	FieldCoordName    string
	Name              string
	FieldSizeMod256   uint8
	PrecomputedParams []field.Element // PrecomputedParams[0][n] correspond to integer cₙ₋₁ in std doc
	// PrecomputedParams[n≥1] correspond to field element c_( len(PrecomputedParams[0]) + n - 1 ) in std doc
	Z                []big.Int // z (or zeta) is a quadratic non-residue with //TODO: some extra nice properties, refer to WB19
	CofactorClearing bool
	MappingAlgorithm FieldElementToCurvePoint

	// Sarkar parameters for optimized SqrtRatio when 2-adicity e >= 10
	SarkarEnabled bool
	SarkarN       int           // 2-adicity (e)
	SarkarK       int           // number of blocks
	SarkarL       []int         // block sizes, sum(L) = SarkarN - 1
	SarkarG       field.Element // g = Z^m, primitive 2^e-th root of unity
}
