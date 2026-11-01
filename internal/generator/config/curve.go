package config

import (
	"math/big"

	"github.com/consensys/gnark-crypto/internal/generator/addchain"
	"github.com/consensys/gnark-crypto/internal/generator/field/config"
)

// Curve describes parameters of the curve useful for the template
type Curve struct {
	Name         string
	CurvePackage string
	Package      string // current package being generated
	EnumID       string
	FpModulus    string
	FrModulus    string

	Fp           *config.Field
	Fr           *config.Field
	FpUnusedBits int

	FpInfo, FrInfo FieldInfo
	G1             Point
	G2             Point

	HashE1 HashSuite
	HashE2 HashSuite

	// NoFieldSuite disables the field-based auxiliary packages generated from Fr/Fp
	// in internal/generator/main.go: MiMC, polynomial, Poseidon2, and hash_to_field.
	// Curves generate that suite by default unless it is explicitly disabled here.
	NoFieldSuite bool

	// NoECC disables the ECC package generation (G1, G2, multiexp, marshal).
	// Used for curves that only need field arithmetic and ECDSA (e.g., stark-curve, secp256r1).
	NoECC bool

	// ECDSAKeyRecovery enables ECDSA public key recovery (SignForRecover, RecoverPublicKey).
	ECDSAKeyRecovery bool

	// E2 Cbrt precomputes (for Fp² cube root)
	E2CbrtP2Mod9  uint64 // p² mod 9
	E2CbrtP2Mod27 uint64 // p² mod 27
	// Torus-based E2 Cbrt
	E2CbrtTorusEnabled       bool     // whether torus cbrt is available
	E2CbrtTorusBeta          int64    // beta from Fp2=Fp[u]/(u²-beta), e.g. -1 or -5
	E2CbrtTorusBetaAbs       int64    // |beta|
	E2CbrtTorusBetaInvNeg    []uint64 // 1/|beta| in Montgomery form (when beta != -1)
	E2CbrtTorusHelperSquared bool     // m = norm·t² (true) or norm·t (false)
	E2CbrtTorusUniqueRoot    bool     // q ≡ 7 mod 9: no ζ-adjustment needed
	E2CbrtTorusNormInvM      int      // power of m in normInv = m^a · t^b
	E2CbrtTorusNormInvT      int      // power of t in normInv = m^a · t^b
	E2CbrtTorusLucasExponent []uint64 // 3⁻¹ mod (p+1) little-endian
	E2CbrtTorusLucasNLimbs   int
	E2CbrtTorusLucasTopBit   int // bit length - 1
}

type TwistedEdwardsCurve struct {
	Name    string
	Package string
	EnumID  string

	A, D, Cofactor, Order, BaseX, BaseY string

	// subgroup-test generation metadata
	Split2Torsion       bool
	QuarticExponentData *addchain.AddChainData
	OcticExponentData   *addchain.AddChainData

	// Split-2-torsion subgroup-check constants. These are curve-specific
	// constants used by the existing Koshelev-style cofactor-4 check.
	Split2TorsionT0 string
	Split2TorsionT1 string
	Split2TorsionB  string

	// set if endomorphism
	HasEndomorphism bool
	Endo0, Endo1    string
	Lambda          string
}

type FieldInfo struct {
	Bits    int
	Bytes   int
	Modulus func() *big.Int
}

func (c Curve) Equal(other Curve) bool {
	return c.Name == other.Name
}

func (c Curve) HasG1() bool {
	return c.G1.PointName != ""
}

func (c Curve) HasG2() bool {
	return c.G2.PointName != ""
}

func (c Curve) MarshalMetadataBits() int {
	return c.FpInfo.Bytes*8 - c.FpInfo.Bits
}

func (c Curve) SupportsPointCompression() bool {
	return c.MarshalMetadataBits() > 0
}

func (c Curve) GenerateMarshal() bool {
	return c.HasG1()
}

func (c Curve) GenerateECC() bool {
	return c.HasG1() && !c.NoECC
}

func (c Curve) GenerateFieldSuite() bool {
	return !c.NoFieldSuite
}

func (c Curve) GenerateFFT() bool {
	return c.GenerateFieldSuite() && c.GeneratePairingPackages()
}

func (c Curve) GenerateHashToCurve1() bool {
	return c.HashE1 != nil
}

func (c Curve) GenerateHashToCurve2() bool {
	return c.HashE2 != nil
}

func (c Curve) GenerateHashToCurve() bool {
	return c.GenerateHashToCurve1() || c.GenerateHashToCurve2()
}

func (c Curve) GeneratePairingPackages() bool {
	return c.HasG2()
}

func (p Point) CMax() int {
	if len(p.CRange) == 0 {
		return 0
	}
	return p.CRange[len(p.CRange)-1]
}

type Point struct {
	CoordType        string
	CoordExtDegree   uint8 // value n, such that q = pⁿ
	CoordExtRoot     int64 // value a, such that the field is Fp[X]/(Xⁿ - a)
	PointName        string
	GLV              bool   // scalar multiplication using GLV
	CofactorCleaning bool   // flag telling if the Cofactor cleaning is available
	CRange           []int  // multiexp bucket method: generate inner methods (with const arrays) for each c
	Projective       bool   // generate projective coordinates
	A                string // A linear coefficient in Weierstrass form y²=x³+ax+b
	B                string // B constant term in Weierstrass form y²=x³+ax+b

	// SubgroupCheck describes the endomorphism relation used by IsInSubGroup,
	// if any. It is validated when the curve is registered (see addCurve).
	SubgroupCheck *EndoSubgroupCheck
}

// Endomorphisms used in subgroup membership tests.
const (
	EndoPhi = "phi" // ϕ: (x,y) → (ωx,y), ϕ² + ϕ + 1 = 0
	EndoPsi = "psi" // ψ = untwist∘π∘twist, ψ² - tψ + p = 0 with t the trace of E(Fp)
)

// EndoSubgroupCheck describes an IsInSubGroup implementation which checks
// f(ϕ)(P) == O, where f(X) = Σ Coeffs[i]·Xⁱ and ϕ is Endomorphism.
//
// It mirrors the hand-written relation in the point template and is only used
// to check that this relation is an exact membership test.
type EndoSubgroupCheck struct {
	Endomorphism string   // EndoPhi or EndoPsi
	Trace        string   // EndoPsi only: trace t of Frobenius of E(Fp)
	Cofactor     string   // h = #E(F_q) / r for the group the check applies to
	Coeffs       []string // f(X) = Σ Coeffs[i]·Xⁱ
}

// Norm returns N = Res(X² - tX + n, f) where X² - tX + n = 0 is the
// characteristic polynomial of the endomorphism. f(ϕ) composed with its
// conjugate is the multiplication by N, so ker f(ϕ) ⊆ E[N].
func (c *EndoSubgroupCheck) Norm(p *big.Int) *big.Int {
	var t, n big.Int
	switch c.Endomorphism {
	case EndoPhi:
		t.SetInt64(-1)
		n.SetInt64(1)
	case EndoPsi:
		mustSetString(&t, c.Trace)
		n.Set(p)
	default:
		panic("unknown endomorphism " + c.Endomorphism)
	}
	// f(X) mod (X² - tX + n) = a + bX, with Xⁱ = xa + xb·X
	a, b := new(big.Int), new(big.Int)
	xa, xb := big.NewInt(1), big.NewInt(0)
	var ci, tmp big.Int
	for i, s := range c.Coeffs {
		if i > 0 {
			// X·(xa + xb·X) = -n·xb + (xa + t·xb)·X
			tmp.Mul(&t, xb).Add(&tmp, xa)
			xa.Mul(&n, xb).Neg(xa)
			xb.Set(&tmp)
		}
		mustSetString(&ci, s)
		a.Add(a, tmp.Mul(&ci, xa))
		b.Add(b, tmp.Mul(&ci, xb))
	}
	// N(a + bX) = a² + t·ab + n·b²
	N := new(big.Int).Mul(a, a)
	N.Add(N, tmp.Mul(a, b).Mul(&tmp, &t))
	N.Add(N, tmp.Mul(b, b).Mul(&tmp, &n))
	return N.Abs(N)
}

// IsExact reports whether f(ϕ)(P) == O ⟺ P ∈ G on E(F_q), where G is the
// subgroup of order r. This holds iff
//   - r | N, i.e. f(λ) ≡ 0 mod r for an eigenvalue λ of the endomorphism (G ⊆ ker),
//   - r ∤ h, so that E(F_q)[r] = G,
//   - gcd(h, N) = 1, so that f(ϕ) is injective on the cofactor part of E(F_q).
func (c *EndoSubgroupCheck) IsExact(p, r *big.Int) bool {
	var h, tmp big.Int
	mustSetString(&h, c.Cofactor)
	N := c.Norm(p)
	return tmp.Mod(N, r).Sign() == 0 &&
		tmp.Mod(&h, r).Sign() != 0 &&
		tmp.GCD(nil, nil, &h, N).Cmp(big.NewInt(1)) == 0
}

// seedPoly returns (Σ c[i]·xⁱ) / den in decimal. It panics if the division is
// not exact.
func seedPoly(x string, den int64, c ...int64) string {
	var bx, res, xi, tmp big.Int
	mustSetString(&bx, x)
	xi.SetInt64(1)
	for _, ci := range c {
		res.Add(&res, tmp.Mul(&xi, big.NewInt(ci)))
		xi.Mul(&xi, &bx)
	}
	var rem big.Int
	res.QuoRem(&res, big.NewInt(den), &rem)
	if rem.Sign() != 0 {
		panic("seedPoly: non-integral polynomial value")
	}
	return res.String()
}

func mustSetString(z *big.Int, s string) {
	if _, ok := z.SetString(s, 10); !ok {
		panic("invalid integer " + s)
	}
}

func (p Point) IsZeroA() bool {
	return p.A == "" || p.A == "0"
}

func (p Point) IsNeg3A() bool {
	return p.A == "-3"
}

var Curves []Curve
var TwistedEdwardsCurves []TwistedEdwardsCurve

func (c TwistedEdwardsCurve) HasSplit2TorsionSubgroupCheck() bool {
	return c.Cofactor == "4" && c.Split2Torsion &&
		c.Split2TorsionT0 != "" &&
		c.Split2TorsionT1 != "" &&
		c.Split2TorsionB != ""
}

func (c TwistedEdwardsCurve) HasNonSplitSubgroupCheck() bool {
	return !c.Split2Torsion && (c.Cofactor == "4" || c.Cofactor == "8")
}

func (c TwistedEdwardsCurve) HasOcticSubgroupCheck() bool {
	return c.HasNonSplitSubgroupCheck() && c.Cofactor == "8"
}

func (c TwistedEdwardsCurve) HasSubgroupCheck() bool {
	return c.HasSplit2TorsionSubgroupCheck() || c.HasNonSplitSubgroupCheck()
}

func defaultCRange() []int {
	// default range for C values in the multiExp
	return []int{4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15, 16}
}

func addCurve(c *Curve) {
	// init FpInfo and FrInfo
	c.FpInfo = newFieldInfo(c.FpModulus)
	c.FrInfo = newFieldInfo(c.FrModulus)

	// Compute E2 Cbrt precomputes
	p := c.FpInfo.Modulus()

	// the endomorphism-based subgroup membership tests must be exact
	for _, g := range []Point{c.G1, c.G2} {
		if g.SubgroupCheck != nil && !g.SubgroupCheck.IsExact(p, c.FrInfo.Modulus()) {
			panic("IsInSubGroup endomorphism relation is not an exact membership test for " + c.Name + "/" + g.PointName)
		}
	}
	p2 := new(big.Int).Mul(p, p)
	c.E2CbrtP2Mod9 = new(big.Int).Mod(p2, big.NewInt(9)).Uint64()
	c.E2CbrtP2Mod27 = new(big.Int).Mod(p2, big.NewInt(27)).Uint64()

	// Torus-based E2 Cbrt parameters
	// Check q mod 3 == 1 (needed for torus cbrt helper)
	pMod3 := new(big.Int).Mod(p, big.NewInt(3)).Uint64()
	if pMod3 == 1 {
		qMod9Torus := new(big.Int).Mod(p, big.NewInt(9)).Uint64()
		qMod27Torus := new(big.Int).Mod(p, big.NewInt(27)).Uint64()

		// Check if one of the supported cases
		var hasTorus bool
		switch {
		case qMod9Torus == 7, qMod9Torus == 4:
			hasTorus = true
		case qMod27Torus == 10, qMod27Torus == 19:
			hasTorus = true
		}

		if hasTorus {
			c.E2CbrtTorusEnabled = true

			// beta from Fp2 = Fp[u]/(u² - beta)
			c.E2CbrtTorusBeta = c.G2.CoordExtRoot
			if c.E2CbrtTorusBeta == 0 {
				c.E2CbrtTorusBeta = -1 // BN254 default
			}
			c.E2CbrtTorusBetaAbs = -c.E2CbrtTorusBeta

			c.E2CbrtTorusHelperSquared = (qMod27Torus == 10)
			c.E2CbrtTorusUniqueRoot = (qMod9Torus == 7)

			// normInv = m^a · t^b where a = r-2, b = d - (r-2)*s
			var d, r uint64
			var s int
			switch {
			case qMod9Torus == 7:
				d, r, s = 9, 7, 1
			case qMod9Torus == 4:
				d, r, s = 9, 4, 1
			case qMod27Torus == 10:
				d, r, s = 27, 10, 2
			case qMod27Torus == 19:
				d, r, s = 27, 19, 1
			}
			c.E2CbrtTorusNormInvM = int(r - 2)
			c.E2CbrtTorusNormInvT = int(d) - c.E2CbrtTorusNormInvM*s

			// Lucas exponent: 3⁻¹ mod (p+1)
			pPlus1 := new(big.Int).Add(p, big.NewInt(1))
			three := big.NewInt(3)
			lucasExp := new(big.Int).ModInverse(three, pPlus1)
			nLucasLimbs := (lucasExp.BitLen() + 63) / 64
			c.E2CbrtTorusLucasExponent = bigIntToLimbs(lucasExp, nLucasLimbs)
			c.E2CbrtTorusLucasNLimbs = nLucasLimbs
			c.E2CbrtTorusLucasTopBit = lucasExp.BitLen() - 1
		}
	}

	Curves = append(Curves, *c)
}

func addTwistedEdwardCurve(c *TwistedEdwardsCurve) {
	var parent *Curve
	for i := range Curves {
		if Curves[i].Name == c.Name {
			parent = &Curves[i]
			break
		}
	}
	if parent == nil {
		panic("missing parent curve for twisted Edwards curve " + c.Name)
	}

	var modulus, a, d, cofactor big.Int
	if _, ok := modulus.SetString(parent.FrModulus, 10); !ok {
		panic("invalid fr modulus for twisted Edwards curve " + c.Name)
	}
	if _, ok := a.SetString(c.A, 10); !ok {
		panic("invalid A parameter for twisted Edwards curve " + c.Name)
	}
	if _, ok := d.SetString(c.D, 10); !ok {
		panic("invalid D parameter for twisted Edwards curve " + c.Name)
	}
	if _, ok := cofactor.SetString(c.Cofactor, 10); !ok {
		panic("invalid cofactor for twisted Edwards curve " + c.Name)
	}

	a.Mod(&a, &modulus)
	d.Mod(&d, &modulus)
	var ad big.Int
	ad.Mul(&a, &d).Mod(&ad, &modulus)
	c.Split2Torsion = legendreSymbol(&ad, &modulus) == 1

	switch c.Cofactor {
	case "4", "8":
		if !c.HasSubgroupCheck() {
			panic("unsupported twisted Edwards subgroup check configuration for " + c.Name + "/" + c.Package)
		}
	}

	if c.HasNonSplitSubgroupCheck() {
		var exp big.Int
		exp.Sub(&modulus, big.NewInt(1))
		switch c.Cofactor {
		case "4":
			exp.Rsh(&exp, 2)
			c.QuarticExponentData = addchain.GetAddChain(&exp)
		case "8":
			exp.Rsh(&exp, 3)
			c.OcticExponentData = addchain.GetAddChain(&exp)
		}
	}

	TwistedEdwardsCurves = append(TwistedEdwardsCurves, *c)
}

func legendreSymbol(x, p *big.Int) int {
	if x.Sign() == 0 {
		return 0
	}
	var e, z big.Int
	e.Sub(p, big.NewInt(1)).Rsh(&e, 1)
	z.Exp(x, &e, p)
	if z.Sign() == 0 {
		return 0
	}
	if z.Cmp(big.NewInt(1)) == 0 {
		return 1
	}
	return -1
}

// bigIntToLimbs converts a big.Int to a little-endian slice of uint64 limbs.
func bigIntToLimbs(n *big.Int, nLimbs int) []uint64 {
	limbs := make([]uint64, nLimbs)
	mask := new(big.Int).SetUint64(^uint64(0))
	tmp := new(big.Int).Set(n)
	for i := 0; i < nLimbs; i++ {
		limbs[i] = new(big.Int).And(tmp, mask).Uint64()
		tmp.Rsh(tmp, 64)
	}
	return limbs
}

func newFieldInfo(modulus string) FieldInfo {
	var F FieldInfo
	var bModulus big.Int
	if _, ok := bModulus.SetString(modulus, 10); !ok {
		panic("invalid modulus " + modulus)
	}

	F.Bits = bModulus.BitLen()
	F.Bytes = (F.Bits + 7) / 8
	F.Modulus = func() *big.Int { return new(big.Int).Set(&bModulus) }
	return F
}
