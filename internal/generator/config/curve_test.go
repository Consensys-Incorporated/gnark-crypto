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
			if !g.SubgroupCheck.IsExact(c.FpInfo.Modulus(), c.FrInfo.Modulus()) {
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

			// f = 1: N = 1 is not divisible by r
			check := *g.SubgroupCheck
			check.Coeffs = []string{"1"}
			if check.IsExact(p, r) {
				t.Errorf("%s/%s: expected f = 1 to be rejected", c.Name, g.PointName)
			}

			// f·h: h | N
			var h, ci big.Int
			mustSetString(&h, g.SubgroupCheck.Cofactor)
			check.Coeffs = make([]string, len(g.SubgroupCheck.Coeffs))
			for i, s := range g.SubgroupCheck.Coeffs {
				mustSetString(&ci, s)
				check.Coeffs[i] = ci.Mul(&ci, &h).String()
			}
			if check.IsExact(p, r) {
				t.Errorf("%s/%s: expected h·f to be rejected", c.Name, g.PointName)
			}
		}
	}
}
