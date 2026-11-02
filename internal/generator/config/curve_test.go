package config

import (
	"math/big"
	"testing"
)

func TestSubgroupChecksAreExact(t *testing.T) {
	nbChecks := 0
	for _, c := range Curves {
		for _, g := range []Point{c.G1, c.G2} {
			if g.SubgroupCheck == nil {
				continue
			}
			nbChecks++
			lambda := c.EndoEigenvalue(g.SubgroupCheck.Endomorphism)
			if !g.SubgroupCheck.IsExact(c.FpInfo.Modulus(), c.FrInfo.Modulus(), lambda) {
				t.Errorf("%s/%s: subgroup check is not exact", c.Name, g.PointName)
			}
		}
	}
	if nbChecks == 0 {
		t.Fatal("no subgroup checks configured")
	}
}

// IsExact must reject relations which do not vanish on the r-torsion, or which
// are not injective on the cofactor part of the group.
func TestSubgroupCheckNotExact(t *testing.T) {
	for _, c := range Curves {
		for _, g := range []Point{c.G1, c.G2} {
			if g.SubgroupCheck == nil {
				continue
			}
			p, r := c.FpInfo.Modulus(), c.FrInfo.Modulus()
			lambda := c.EndoEigenvalue(g.SubgroupCheck.Endomorphism)
			reject := func(desc string, check EndoSubgroupCheck, lambda *big.Int) {
				t.Helper()
				if check.IsExact(p, r, lambda) {
					t.Errorf("%s/%s: expected %s to be rejected", c.Name, g.PointName, desc)
				}
			}

			// f = 1: does not vanish on G
			check := *g.SubgroupCheck
			check.Coeffs = []string{"1"}
			reject("f = 1", check, lambda)

			// f·h: h | N
			var h, ci big.Int
			mustSetString(&h, g.SubgroupCheck.Cofactor)
			check.Coeffs = make([]string, len(g.SubgroupCheck.Coeffs))
			for i, s := range g.SubgroupCheck.Coeffs {
				mustSetString(&ci, s)
				check.Coeffs[i] = ci.Mul(&ci, &h).String()
			}
			reject("h·f", check, lambda)

			// conjugate of f: same norm, but vanishes at the other eigenvalue
			// of the endomorphism mod r instead of λ. With X̄ = t - X,
			// conj(a + bX) = (a + bt) - bX.
			tr, _ := g.SubgroupCheck.charPoly(p)
			a, b := g.SubgroupCheck.reduce(p)
			check.Coeffs = []string{
				new(big.Int).Add(a, new(big.Int).Mul(b, tr)).String(),
				new(big.Int).Neg(b).String(),
			}
			if check.Norm(p).Cmp(g.SubgroupCheck.Norm(p)) != 0 {
				t.Fatalf("%s/%s: conjugate relation should have the same norm", c.Name, g.PointName)
			}
			reject("the conjugate of f", check, lambda)

			// ψ(P) == P: N = #E(Fp) is divisible by r, but f vanishes at the
			// eigenvalue 1 of ψ, not at p.
			if g.SubgroupCheck.Endomorphism == EndoPsi {
				check.Coeffs = []string{"-1", "1"}
				reject("f = X - 1", check, lambda)
			}

			// λ must be an eigenvalue of the endomorphism mod r
			reject("λ+1", *g.SubgroupCheck, new(big.Int).Add(lambda, big.NewInt(1)))
		}
	}
}
