package field

import (
	"fmt"
	"math/bits"
	"os"
	"path/filepath"
	"strings"

	"github.com/consensys/bavard"
	eccconfig "github.com/consensys/gnark-crypto/internal/generator/config"
	"github.com/consensys/gnark-crypto/internal/generator/field/asm/amd64"
	"github.com/consensys/gnark-crypto/internal/generator/field/config"
	"github.com/consensys/gnark-crypto/internal/generator/field/template"
)

func generateFFT(F *config.Field, fft *config.FFT, outputDir string) error {

	if fft.GeneratorFullMultiplicativeGroup == 0 || fft.GeneratorMaxTwoAdicSubgroup == "" {
		// try to populate ourselves
		// TODO @gbotrel right now hardcoded lookup tables, we should do better.
		data, ok := fftConfigs[F.Modulus]
		if !ok {
			return fmt.Errorf("no fft config for modulus %s", F.Modulus)
		}
		fft = &data
	}

	fieldImportPath, err := getImportPath(outputDir)
	if err != nil {
		return err
	}
	data := &fftTemplateData{
		FFT:              *fft,
		FieldPackagePath: fieldImportPath,
		FF:               F.PackageName,
		HasASMKernel:     F.F31,
		Kernels:          []int{5, 8},
		Package:          "fft",
		SmallField:       F.F31 || F.SingleWordSmall,
		ExtType:          "E4",
		ExtAlign:         4,
	}
	if F.Q[0] == 562932773552129 {
		// mamabear builds its FFTExt over the cubic extension.
		data.ExtType = "E3"
		data.ExtAlign = 1
	}
	outputDir = filepath.Join(outputDir, "fft")

	pureGoBuildTag := ""
	if data.HasASMKernel {
		pureGoBuildTag = "purego || (!amd64)"
		data.Kernels = []int{8}
	}

	entries := []bavard.Entry{
		{File: filepath.Join(outputDir, "doc.go"), Templates: []string{"doc.go.tmpl"}},
		{File: filepath.Join(outputDir, "domain_test.go"), Templates: []string{"tests/domain.go.tmpl"}},
		{File: filepath.Join(outputDir, "domain.go"), Templates: []string{"domain.go.tmpl"}},
		{File: filepath.Join(outputDir, "fft_test.go"), Templates: []string{"tests/fft.go.tmpl"}},
		{File: filepath.Join(outputDir, "fft.go"), Templates: []string{"fft.go.tmpl"}},
		{File: filepath.Join(outputDir, "kernel_purego.go"), Templates: []string{"kernel.purego.go.tmpl"}, BuildTag: pureGoBuildTag},
		{File: filepath.Join(outputDir, "options.go"), Templates: []string{"options.go.tmpl"}},
	}
	// every small field gets an FFT over its extension; only the F31 fields
	// additionally carry the degree-6 variant.
	if data.SmallField {
		entries = append(entries, bavard.Entry{File: filepath.Join(outputDir, "fftext_test.go"), Templates: []string{"tests/fftext.go.tmpl"}})
		entries = append(entries, bavard.Entry{File: filepath.Join(outputDir, "fftext.go"), Templates: []string{"fftext.go.tmpl"}})
	}
	if F.F31 {
		entries = append(entries, bavard.Entry{File: filepath.Join(outputDir, "fftext6_test.go"), Templates: []string{"tests/fftext6.go.tmpl"}})
		entries = append(entries, bavard.Entry{File: filepath.Join(outputDir, "fftext6.go"), Templates: []string{"fftext6.go.tmpl"}})
	}
	if data.HasASMKernel {
		data.Q = F.Q[0]
		data.QInvNeg = F.QInverse[0]
		entries = append(entries,
			bavard.Entry{
				File:      filepath.Join(outputDir, "kernel_amd64.go"),
				Templates: []string{"kernel.amd64.go.tmpl"},
				BuildTag:  "!purego"})

		// generate the assembly file;
		fftKernels, err := os.Create(filepath.Join(outputDir, "kernel_amd64.s"))
		if err != nil {
			return err
		}

		fftKernels.WriteString("//go:build !purego\n")

		if err := amd64.GenerateF31FFTKernels(fftKernels, F.NbBits, data.Kernels); err != nil {
			fftKernels.Close()
			return err
		}
		fftKernels.Close()
	}

	funcs := make(map[string]any)
	funcs["bitReverse"] = bitReverse
	funcs["reverseBits"] = func(x, n any) uint64 {
		return bits.Reverse64(anyToUint64(x)) >> anyToUint64(n)
	}
	funcs["shl"] = func(x, n any) uint64 {
		return anyToUint64(x) << anyToUint64(n)
	}
	funcs["logicalOr"] = func(x, y any) uint64 {
		return anyToUint64(x) | anyToUint64(y)
	}

	bavardOpts := []func(*bavard.Bavard) error{bavard.Funcs(funcs)}

	g := NewGenerator(template.FS)

	if err := g.GenerateWithOptions(data, "fft", "", "fft", bavardOpts, entries...); err != nil {
		return err
	}

	// put the generator in the parent dir (fr)
	entries = []bavard.Entry{
		{File: filepath.Join(outputDir, "../generator.go"), Templates: []string{"fr.generator.go.tmpl"}},
	}
	fieldNameSplitted := strings.Split(data.FieldPackagePath, "/")
	fieldName := fieldNameSplitted[len(fieldNameSplitted)-1]
	if err := g.GenerateWithOptions(data, fieldName, "", "fft", bavardOpts, entries...); err != nil {
		return err
	}

	return runFormatters(outputDir)
}

type fftTemplateData struct {
	config.FFT

	FieldPackagePath string // path to the finite field package
	FF               string // name of the package corresponding to the finite field
	HasASMKernel     bool   // indicates if the kernels have an assembly impl
	Kernels          []int  // indicates which kernels to generate
	Package          string // package name
	Q, QInvNeg       uint64
	// SmallField is set for fields whose element fits in one word with a spare
	// bit. Those FFTs keep bit-reversed coset tables and use aligned parallel
	// execution. This is what .F31 meant here before a non-31-bit field
	// qualified.
	SmallField bool
	// ExtType is the extension the FFTExt variant operates over ("E4" or "E3"),
	// and ExtAlign the parallel alignment that matches its element size.
	ExtType  string
	ExtAlign int
}

func anyToUint64(x any) uint64 {
	switch v := x.(type) {
	case int:
		return uint64(v)
	case int64:
		return uint64(v)
	case uint64:
		return v
	default:
		panic("unknown type")
	}
}

var fftConfigs map[string]config.FFT

func init() {
	fftConfigs = make(map[string]config.FFT)

	// TODO @gbotrel temporary
	// bls12-377
	fftConfigs[eccconfig.BLS12_377.FrModulus] = config.NewConfig(
		22,
		"8065159656716812877374967518403273466521432693661810619979959746626482506078",
		"47",
	)

	// bls12-381
	fftConfigs[eccconfig.BLS12_381.FrModulus] = (config.NewConfig(
		7,
		"10238227357739495823651030575849232062558860180284477541189508159991286009131",
		"32",
	))

	// bn254
	fftConfigs[eccconfig.BN254.FrModulus] = (config.NewConfig(
		5,
		"19103219067921713944291392827692070036145651957329286315305642004821462161904",
		"28",
	))

	// bw6-761
	fftConfigs[eccconfig.BW6_761.FrModulus] = (config.NewConfig(
		15,
		"32863578547254505029601261939868325669770508939375122462904745766352256812585773382134936404344547323199885654433",
		"46",
	))

	// bw6-633
	fftConfigs[eccconfig.BW6_633.FrModulus] = (config.NewConfig(
		13,
		"4991787701895089137426454739366935169846548798279261157172811661565882460884369603588700158257",
		"20",
	))

	// bls24-315
	fftConfigs[eccconfig.BLS24_315.FrModulus] = (config.NewConfig(
		7,
		"1792993287828780812362846131493071959406149719416102105453370749552622525216",
		"22",
	))

	// bls24-317
	fftConfigs[eccconfig.BLS24_317.FrModulus] = (config.NewConfig(
		7,
		"16532287748948254263922689505213135976137839535221842169193829039521719560631",
		"60",
	))

	// goldilocks
	fftConfigs["18446744069414584321"] = (config.NewConfig(
		7,
		"1753635133440165772",
		"32",
	))

	// koala bear
	fftConfigs["2130706433"] = (config.NewConfig(
		3,
		"1791270792",
		"24",
	))

	// mama bear: q - 1 = 2^34 * 32767, so the largest NTT is 2^34.
	// 3 is the smallest multiplicative generator, and 3^32767 mod q has order
	// exactly 2^34.
	fftConfigs["562932773552129"] = (config.NewConfig(
		3,
		"57971402726332",
		"34",
	))

	// baby bear
	fftConfigs["2013265921"] = (config.NewConfig(
		31,
		"440564289",
		"27",
	))

}

func bitReverse(n, i int64) uint64 {
	nn := uint64(64 - bits.TrailingZeros64(uint64(n)))
	r := make([]uint64, n)
	for i := range r {
		r[i] = uint64(i)
	}
	for i := range r {
		irev := bits.Reverse64(r[i]) >> nn
		if irev > uint64(i) {
			r[i], r[irev] = r[irev], r[i]
		}
	}
	return r[i]
}
