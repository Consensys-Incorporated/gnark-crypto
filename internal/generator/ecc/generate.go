package ecc

import (
	"path/filepath"
	"slices"
	"sort"
	"strings"

	"github.com/consensys/bavard"
	"github.com/consensys/gnark-crypto/internal/generator/common"
	"github.com/consensys/gnark-crypto/internal/generator/config"
	"github.com/consensys/gnark-crypto/internal/generator/ecc/template"
)

func Generate(conf config.Curve, baseDir string, gen *common.Generator) error {
	packageName := strings.ReplaceAll(conf.Name, "-", "")

	// we need to use a new generator with the correct FS
	eccGen := common.NewDefaultGenerator(template.FS)

	var entries []bavard.Entry

	// G1
	entries = []bavard.Entry{
		{File: filepath.Join(baseDir, "g1.go"), Templates: []string{"point.go.tmpl"}},
		{File: filepath.Join(baseDir, "g1_test.go"), Templates: []string{"tests/point.go.tmpl"}},
	}

	g1 := pconf{conf, conf.G1}
	if err := eccGen.Generate(g1, packageName, "", "", entries...); err != nil {
		return err
	}

	// MSM
	entries = []bavard.Entry{
		{File: filepath.Join(baseDir, "multiexp.go"), Templates: []string{"multiexp.go.tmpl"}},
		{File: filepath.Join(baseDir, "multiexp_affine.go"), Templates: []string{"multiexp_affine.go.tmpl"}},
		{File: filepath.Join(baseDir, "multiexp_jacobian.go"), Templates: []string{"multiexp_jacobian.go.tmpl"}},
		{File: filepath.Join(baseDir, "multiexp_test.go"), Templates: []string{"tests/multiexp.go.tmpl"}},
	}
	conf.Package = packageName

	// return the last window size for a scalar;
	// this last window should accommodate a carry (from the NAF decomposition)
	// it can be == c if we have 1 available bit
	// it can be > c if we have 0 available bit
	// it can be < c if we have 2+ available bits
	lastC := func(c int) int {
		nbChunks := (conf.Fr.NbBits + c - 1) / c
		nbAvailableBits := (nbChunks * c) - conf.Fr.NbBits
		lc := c + 1 - nbAvailableBits
		if lc > 16 {
			panic("we have a problem since we are using uint16 to store digits")
		}
		return lc
	}
	// batchAffineC is the smallest window size whose chunk processor accumulates
	// buckets with the batch affine method; every smaller window uses extended
	// jacobian formulas. The msm cost model prices the two differently, so the
	// boundary has to come from one place.
	const batchAffineC = 10

	batchSize := func(c int) int {
		switch c {
		case 10:
			return 80
		case 11:
			return 150
		case 12:
			return 200
		case 13:
			return 350
		case 14:
			return 400
		case 15:
			return 500
		default:
			return 640
		}
	}

	funcs := common.Funcs()
	funcs["lastC"] = lastC
	funcs["batchSize"] = batchSize
	funcs["batchAffineC"] = func() int { return batchAffineC }

	funcs["nbBuckets"] = func(c int) int {
		return 1 << (c - 1)
	}

	funcs["contains"] = func(v int, s []int) bool {
		return slices.Contains(s, v)
	}
	lastCG1 := make([]int, 0)
	for {
		for i := 0; i < len(conf.G1.CRange); i++ {
			lc := lastC(conf.G1.CRange[i])
			if !contains(conf.G1.CRange, lc) && !contains(lastCG1, lc) {
				lastCG1 = append(lastCG1, lc)
			}
		}
		if len(lastCG1) == 0 {
			break
		}
		conf.G1.CRange = append(conf.G1.CRange, lastCG1...)
		sort.Ints(conf.G1.CRange)
		lastCG1 = lastCG1[:0]
	}

	lastCG2 := make([]int, 0)
	for {
		for i := 0; i < len(conf.G2.CRange); i++ {
			lc := lastC(conf.G2.CRange[i])
			if !contains(conf.G2.CRange, lc) && !contains(lastCG2, lc) {
				lastCG2 = append(lastCG2, lc)
			}
		}
		if len(lastCG2) == 0 {
			break
		}
		conf.G2.CRange = append(conf.G2.CRange, lastCG2...)
		sort.Ints(conf.G2.CRange)
		lastCG2 = lastCG2[:0]
	}

	// the window sizes whose chunk processor is the batch affine one, across both
	// groups; the msm cost model needs their batch sizes to price a bucket
	// accumulation on that path
	funcs["batchAffineCs"] = func() []int {
		cs := make([]int, 0, len(conf.G1.CRange)+len(conf.G2.CRange))
		for _, c := range append(append([]int{}, conf.G1.CRange...), conf.G2.CRange...) {
			if c >= batchAffineC && !slices.Contains(cs, c) {
				cs = append(cs, c)
			}
		}
		sort.Ints(cs)
		return cs
	}

	bavardOpts := []func(*bavard.Bavard) error{bavard.Funcs(funcs)}
	if err := eccGen.GenerateWithOptions(conf, packageName, "", "", bavardOpts, entries...); err != nil {
		return err
	}

	// marshal
	if conf.GenerateMarshal() {
		entries = []bavard.Entry{
			{File: filepath.Join(baseDir, "marshal.go"), Templates: []string{"marshal.go.tmpl"}},
			{File: filepath.Join(baseDir, "marshal_test.go"), Templates: []string{"tests/marshal.go.tmpl"}},
		}

		if err := eccGen.GenerateWithOptions(conf, packageName, "", "", bavardOpts, entries...); err != nil {
			return err
		}
	}

	// G2
	if conf.HasG2() {
		entries = []bavard.Entry{
			{File: filepath.Join(baseDir, "g2.go"), Templates: []string{"point.go.tmpl"}},
			{File: filepath.Join(baseDir, "g2_test.go"), Templates: []string{"tests/point.go.tmpl"}},
		}
		g2 := pconf{conf, conf.G2}
		if err := eccGen.Generate(g2, packageName, "", "", entries...); err != nil {
			return err
		}
	}

	return nil
}

type pconf struct {
	config.Curve
	config.Point
}

func contains(slice []int, v int) bool {
	return slices.Contains(slice, v)
}
