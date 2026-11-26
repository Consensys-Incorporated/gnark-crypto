package config

import (
	"fmt"
	"math/big"

	"github.com/consensys/gnark-crypto/internal/generator/addchain"
)

// CubicSymbol holds the precomputed values for the cubic residue symbol
// (Element.CubicResidueSymbol), generated for multi-word fields with q ≡ 1 (mod 3).
//
// q splits in the Eisenstein integers ℤ[ω] as q = β·β̄ with β = a + bω primary, and
// ℤ[ω]/β ≅ F_q with ω ↦ -a·b⁻¹ (mod q). The symbol is computed by (1-ω)-ary division steps
// on Eisenstein integers whose coordinates are two's complement integers of NbLimbs words.
type CubicSymbol struct {
	NbLimbs      int   // H: words of the coordinates
	LimbsIndexes []int // 0 … H-1, for template ranges
	GNbLimbs     int   // words of G1, G2

	BetaA, BetaB, BetaAMinusB []uint64 // a, b and a-b, two's complement on H words (b ≡ 0 mod 3, b > 0)
	A, B                      string   // a, b in base 10 (documentation)
	JNegBeta                  bool     // a ≡ 1 (mod 3): the algorithm starts from f = -β ≡ -1 (mod 3)

	// x mod β by Babai rounding: x/β ≈ x·conj(β)/q with conj(β) = (a-b) - bω.
	// G1 = ⌊2^(64·NbWords)·|a-b|/q⌋, G2 = ⌊2^(64·NbWords)·|b|/q⌋.
	G1, G2     []uint64
	AMinusBNeg bool // a - b < 0
	BNeg       bool // b < 0

	JPw []uint64 // 2^(64(i+1)) mod 3³⁸, i < H: Montgomery factors of the residues mod 3³⁸

	Omega        []uint64               // ω = -a·b⁻¹ mod q, montgomery form: symbol k ⇔ x^((q-1)/3) = ωᵏ
	OmegaDecimal string                 // ω in base 10 (documentation)
	ExponentHex  string                 // (q-1)/3
	ExponentData *addchain.AddChainData // addition chain for (q-1)/3, if the field uses them
}

// initCubicSymbol sets F.CubicSymbol for the fields where it applies. q must be prime.
func (F *Field) initCubicSymbol(q *big.Int) error {
	three := big.NewInt(3)
	if new(big.Int).Mod(q, three).Int64() != 1 || !F.UsingP20Inverse {
		return nil
	}

	// r: a primitive cube root of unity mod q
	e := new(big.Int).Sub(q, big.NewInt(1))
	e.Div(e, three)
	r := new(big.Int)
	for c := int64(2); ; c++ {
		r.Exp(big.NewInt(c), e, q)
		if r.Cmp(big.NewInt(1)) != 0 {
			break
		}
	}

	// β = gcd(q, ω - r) in ℤ[ω]: β | q and ω ≡ r (mod β), so N(β) = q.
	a, b := eisGCD(new(big.Int).Set(q), big.NewInt(0), new(big.Int).Neg(r), big.NewInt(1))
	// primary associate (b ≡ 0 mod 3), sign chosen so that b > 0
	for i := 0; i < 3 && new(big.Int).Mod(b, three).Sign() != 0; i++ {
		a, b = new(big.Int).Neg(b), new(big.Int).Sub(a, b) // β ← ω·β
	}
	if b.Sign() < 0 {
		a.Neg(a)
		b.Neg(b)
	}
	if new(big.Int).Mod(b, three).Sign() != 0 {
		return fmt.Errorf("cubic symbol: no primary associate")
	}
	norm := new(big.Int).Mul(a, a)
	norm.Add(norm, new(big.Int).Mul(b, b)).Sub(norm, new(big.Int).Mul(a, b))
	if norm.Cmp(q) != 0 {
		return fmt.Errorf("cubic symbol: N(β) ≠ q")
	}
	// ω ↦ -a/b must be r
	omega := new(big.Int).ModInverse(b, q)
	omega.Mul(omega, a).Neg(omega).Mod(omega, q)
	if omega.Cmp(r) != 0 {
		return fmt.Errorf("cubic symbol: -a/b ≠ r")
	}

	aMinusB := new(big.Int).Sub(a, b)

	// coordinates stay below a small multiple of √q
	H := (q.BitLen()/2 + 8 + 63) / 64
	if H < 2 || H > F.NbWords {
		return nil
	}

	g := func(c *big.Int) []uint64 {
		t := new(big.Int).Abs(c)
		t.Lsh(t, uint(64*F.NbWords)).Quo(t, q)
		return toUint64Slice(t, (t.BitLen()+63)/64)
	}
	G1, G2 := g(aMinusB), g(b)
	GH := max(len(G1), len(G2))
	G1 = append(G1, make([]uint64, GH-len(G1))...)
	G2 = append(G2, make([]uint64, GH-len(G2))...)

	twos := func(c *big.Int) []uint64 {
		mod := new(big.Int).Lsh(big.NewInt(1), uint(64*H))
		t := new(big.Int).Mod(c, mod)
		return toUint64Slice(t, H)
	}

	jm := new(big.Int).Exp(three, big.NewInt(38), nil)
	jPw := make([]uint64, H)
	for i := range jPw {
		jPw[i] = new(big.Int).Exp(big.NewInt(2), big.NewInt(int64(64*(i+1))), jm).Uint64()
	}

	cs := &CubicSymbol{
		JPw:          jPw,
		JNegBeta:     new(big.Int).Mod(a, three).Int64() == 1,
		NbLimbs:      H,
		GNbLimbs:     GH,
		BetaA:        twos(a),
		BetaB:        twos(b),
		BetaAMinusB:  twos(aMinusB),
		A:            a.String(),
		B:            b.String(),
		G1:           G1,
		G2:           G2,
		AMinusBNeg:   aMinusB.Sign() < 0,
		BNeg:         b.Sign() < 0,
		OmegaDecimal: omega.String(),
		ExponentHex:  e.Text(16),
	}
	for i := 0; i < H; i++ {
		cs.LimbsIndexes = append(cs.LimbsIndexes, i)
	}
	omegaMont := new(big.Int).Lsh(omega, F.RBits)
	omegaMont.Mod(omegaMont, q)
	cs.Omega = toUint64Slice(omegaMont, F.NbWords)

	if F.UseAddChain {
		cs.ExponentData = addchain.GetAddChain(e)
	}

	F.CubicSymbol = cs
	return nil
}

// eisGCD returns a gcd of x = x0 + x1ω and y = y0 + y1ω in ℤ[ω] (Euclid with
// nearest-coordinate rounding; the remainder norm drops by at least a factor 3/4).
func eisGCD(x0, x1, y0, y1 *big.Int) (*big.Int, *big.Int) {
	for y0.Sign() != 0 || y1.Sign() != 0 {
		// x·conj(y) with conj(y0 + y1ω) = (y0 - y1) - y1ω
		c0, c1 := new(big.Int).Sub(y0, y1), new(big.Int).Neg(y1)
		n0, n1 := eisMul(x0, x1, c0, c1)
		n := new(big.Int).Mul(y0, y0)
		n.Add(n, new(big.Int).Mul(y1, y1)).Sub(n, new(big.Int).Mul(y0, y1))
		q0, q1 := roundDiv(n0, n), roundDiv(n1, n)
		t0, t1 := eisMul(q0, q1, y0, y1)
		r0, r1 := new(big.Int).Sub(x0, t0), new(big.Int).Sub(x1, t1)
		x0, x1, y0, y1 = y0, y1, r0, r1
	}
	return x0, x1
}

// eisMul returns (x0 + x1ω)(y0 + y1ω) = (x0y0 - x1y1) + (x0y1 + x1y0 - x1y1)ω.
func eisMul(x0, x1, y0, y1 *big.Int) (*big.Int, *big.Int) {
	x1y1 := new(big.Int).Mul(x1, y1)
	z0 := new(big.Int).Mul(x0, y0)
	z0.Sub(z0, x1y1)
	z1 := new(big.Int).Mul(x0, y1)
	z1.Add(z1, new(big.Int).Mul(x1, y0)).Sub(z1, x1y1)
	return z0, z1
}

// roundDiv returns the integer nearest to a/d, d > 0.
func roundDiv(a, d *big.Int) *big.Int {
	t := new(big.Int).Lsh(a, 1)
	t.Add(t, d)
	d2 := new(big.Int).Lsh(d, 1)
	return t.Div(t, d2) // Euclidean division: ⌊(2a + d) / 2d⌋
}
