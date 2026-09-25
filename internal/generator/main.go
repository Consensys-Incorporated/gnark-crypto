package main

import (
	"embed"
	"fmt"
	"math/big"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/consensys/bavard"
	"github.com/consensys/gnark-crypto/internal/generator/common"
	"github.com/consensys/gnark-crypto/internal/generator/config"
	configTemplate "github.com/consensys/gnark-crypto/internal/generator/config/template"
	"github.com/consensys/gnark-crypto/internal/generator/crypto/hash/mimc"
	"github.com/consensys/gnark-crypto/internal/generator/crypto/hash/poseidon2"
	"github.com/consensys/gnark-crypto/internal/generator/ecc"
	"github.com/consensys/gnark-crypto/internal/generator/ecdsa"
	"github.com/consensys/gnark-crypto/internal/generator/edwards"
	"github.com/consensys/gnark-crypto/internal/generator/edwards/eddsa"
	"github.com/consensys/gnark-crypto/internal/generator/fflonk"
	"github.com/consensys/gnark-crypto/internal/generator/field"
	fieldConfig "github.com/consensys/gnark-crypto/internal/generator/field/config"
	"github.com/consensys/gnark-crypto/internal/generator/hash_to_curve"
	"github.com/consensys/gnark-crypto/internal/generator/hash_to_field"
	"github.com/consensys/gnark-crypto/internal/generator/kzg"
	"github.com/consensys/gnark-crypto/internal/generator/mpcsetup"
	"github.com/consensys/gnark-crypto/internal/generator/pairing"
	"github.com/consensys/gnark-crypto/internal/generator/pedersen"
	"github.com/consensys/gnark-crypto/internal/generator/permutation"
	"github.com/consensys/gnark-crypto/internal/generator/polynomial"
	"github.com/consensys/gnark-crypto/internal/generator/shplonk"
	"github.com/consensys/gnark-crypto/internal/generator/tower"
)

const (
	copyrightHolder = "Consensys Software Inc."
	copyrightYear   = 2020
)

var gen = common.NewDefaultGenerator(embed.FS{})

var (
	stopOnce    sync.Once
	stopSpinner = make(chan struct{})
)

//go:generate go run main.go
func main() {
	start := time.Now()
	go spinner(stopSpinner)

	baseDir, err := filepath.Abs(filepath.Join("..", ".."))
	assertNoError(err)
	// first we loop through the field arithmetic we must generate.
	// then, we create the common files (only once) for the assembly code.
	asmDirBuildPath := filepath.Join(baseDir, "field", "asm")

	// this enable the generation of fft functions;
	// the parameters are hard coded in a lookup table for now for the modulus we use.
	fftConfig := &fieldConfig.FFT{}

	var wg sync.WaitGroup

	for _, conf := range config.Fields {
		wg.Add(1)
		go func(f config.Field) {
			defer wg.Done()
			var fieldOpts []fieldConfig.FieldOption
			if f.RBits != 0 {
				fieldOpts = append(fieldOpts, fieldConfig.WithMontgomeryRadixBits(f.RBits))
			}
			fc, err := fieldConfig.NewFieldConfig(f.Name, "Element", f.Modulus, true, fieldOpts...)
			assertNoError(err)
			outputDir := filepath.Join(baseDir, "field", f.Name)
			relAsmDir, err := filepath.Rel(outputDir, asmDirBuildPath)
			assertNoError(err)
			asmConfig := &fieldConfig.Assembly{BuildDir: asmDirBuildPath, IncludeDir: relAsmDir}
			assertNoError(field.GenerateFF(fc, outputDir,
				field.WithASM(asmConfig),
				field.WithFFT(fftConfig),
				field.WithSIS(),
				field.WithPoseidon2(),
				field.WithExtensions(),
				field.WithIOP(),
			))
		}(conf)
	}

	// clean up previously generated files before regenerating.
	// files with the "DO NOT EDIT" header are removed; hand-written files
	// (without this header) are preserved.
	for _, conf := range config.Curves {
		curveDir := filepath.Join(baseDir, "ecc", conf.Name)
		cleanGeneratedFiles(curveDir)
	}

	for _, conf := range config.Curves {
		wg.Add(1)
		// for each curve, generate the needed files
		go func(conf config.Curve) {
			defer wg.Done()

			var err error

			conf.Fp, err = fieldConfig.NewFieldConfig("fp", "Element", conf.FpModulus, true)
			assertNoError(err)

			conf.Fr, err = fieldConfig.NewFieldConfig("fr", "Element", conf.FrModulus, !conf.Equal(config.STARK_CURVE))
			assertNoError(err)
			// The torus cbrt helper (ExpByCbrtHelper*) is only used by E2.cbrtTorus via Fp.
			// Clear it from Fr to avoid emitting dead code in scalar field packages.
			conf.Fr.CbrtTorusHelperData = nil
			conf.Fr.CbrtTorusHelperName = ""

			curveDir := filepath.Join(baseDir, "ecc", conf.Name)

			conf.FpUnusedBits = (64 - (conf.Fp.NbBits % 64)) % 64

			// Torus cbrt: compute betaInvNeg now that Fp config is available
			if conf.E2CbrtTorusEnabled && conf.E2CbrtTorusBeta != -1 {
				p := conf.FpInfo.Modulus()
				betaAbs := new(big.Int).SetInt64(conf.E2CbrtTorusBetaAbs)
				betaInv := new(big.Int).ModInverse(betaAbs, p)
				betaInvMont := new(big.Int).Lsh(betaInv, uint(conf.Fp.NbWords)*uint(conf.Fp.Word.BitSize))
				betaInvMont.Mod(betaInvMont, p)
				conf.E2CbrtTorusBetaInvNeg = make([]uint64, conf.Fp.NbWords)
				mask := new(big.Int).SetUint64(^uint64(0))
				tmp := new(big.Int).Set(betaInvMont)
				for i := 0; i < conf.Fp.NbWords; i++ {
					conf.E2CbrtTorusBetaInvNeg[i] = new(big.Int).And(tmp, mask).Uint64()
					tmp.Rsh(tmp, 64)
				}
			}

			// fp
			{
				outputDir := filepath.Join(curveDir, "fp")
				relAsmDir, err := filepath.Rel(outputDir, asmDirBuildPath)
				assertNoError(err)
				asmConfig := &fieldConfig.Assembly{BuildDir: asmDirBuildPath, IncludeDir: relAsmDir}
				assertNoError(field.GenerateFF(conf.Fp, outputDir,
					field.WithASM(asmConfig),
				))
			}

			// fr
			{
				outputDir := filepath.Join(curveDir, "fr")
				relAsmDir, err := filepath.Rel(outputDir, asmDirBuildPath)
				assertNoError(err)
				asmConfig := &fieldConfig.Assembly{BuildDir: asmDirBuildPath, IncludeDir: relAsmDir}

				frOpts := []field.Option{field.WithASM(asmConfig)}
				if conf.GenerateFFT() {
					frOpts = append(frOpts, field.WithFFT(fftConfig), field.WithIOP())
				}
				if conf.Equal(config.BLS12_377) {
					frOpts = append(frOpts, field.WithSIS())
				}

				assertNoError(field.GenerateFF(conf.Fr, outputDir, frOpts...))
			}

			// generate ecdsa
			assertNoError(ecdsa.Generate(conf, curveDir, gen))

			// generate G1, G2, multiExp, marshal, ...
			if conf.GenerateECC() {
				assertNoError(ecc.Generate(conf, curveDir, gen))
			}

			// field suite: mimc, polynomial, poseidon2, hash_to_field
			if conf.GenerateFieldSuite() {
				frInfo := fieldConfig.FieldDependency{
					FieldPackagePath: "github.com/consensys/gnark-crypto/ecc/" + conf.Name + "/fr",
					FieldPackageName: "fr",
					ElementType:      "fr.Element",
				}

				fpInfo := fieldConfig.FieldDependency{
					FieldPackagePath: "github.com/consensys/gnark-crypto/ecc/" + conf.Name + "/fp",
					FieldPackageName: "fp",
					ElementType:      "fp.Element",
				}

				assertNoError(mimc.Generate(conf, filepath.Join(curveDir, "fr", "mimc"), gen))
				assertNoError(polynomial.Generate(frInfo, filepath.Join(curveDir, "fr", "polynomial"), true, gen))
				assertNoError(poseidon2.Generate(conf, filepath.Join(curveDir, "fr", "poseidon2"), gen))
				assertNoError(hash_to_field.Generate(frInfo, filepath.Join(curveDir, "fr", "hash_to_field"), gen))
				assertNoError(hash_to_field.Generate(fpInfo, filepath.Join(curveDir, "fp", "hash_to_field"), gen))
			}

			// hash to curve (only if hash suite is configured and ECC is generated)
			if conf.GenerateHashToCurve() && conf.GenerateECC() {
				assertNoError(hash_to_curve.Generate(conf, curveDir, gen))
			}

			// pairing-dependent packages
			if conf.GeneratePairingPackages() {
				assertNoError(pedersen.Generate(conf, filepath.Join(curveDir, "fr", "pedersen"), gen))
				assertNoError(tower.Generate(conf, filepath.Join(curveDir, "internal", "fptower"), gen))
				assertNoError(pairing.Generate(conf, curveDir, gen))
				assertNoError(mpcsetup.Generate(conf, filepath.Join(curveDir, "mpcsetup"), gen))
				assertNoError(kzg.Generate(conf, filepath.Join(curveDir, "kzg"), gen))
				assertNoError(shplonk.Generate(conf, filepath.Join(curveDir, "shplonk"), gen))
				assertNoError(fflonk.Generate(conf, filepath.Join(curveDir, "fflonk"), gen))
				assertNoError(permutation.Generate(conf, filepath.Join(curveDir, "fr", "permutation"), gen))
			}

		}(conf)

	}

	for _, conf := range config.TwistedEdwardsCurves {
		wg.Add(1)

		go func(conf config.TwistedEdwardsCurve) {
			defer wg.Done()

			curveDir := filepath.Join(baseDir, "ecc", conf.Name, conf.Package)
			// generate twisted edwards companion curves
			assertNoError(edwards.Generate(conf, curveDir, gen))

			// generate eddsa on companion curves
			assertNoError(eddsa.Generate(conf, curveDir, gen))
		}(conf)

	}

	wg.Wait()

	// generate the ecc_field.go file
	{
		entries := []bavard.Entry{
			{File: filepath.Join(baseDir, "ecc", "ecc_field.go"), Templates: []string{"ecc_field.go.tmpl"}},
		}
		eccFieldGen := common.NewGenerator(configTemplate.FS, copyrightHolder, copyrightYear, "consensys/gnark-crypto")
		assertNoError(eccFieldGen.Generate(config.Curves, "ecc", "", "", entries...))
	}

	// format the whole directory

	cmd := exec.Command("gofmt", "-s", "-w", baseDir)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	assertNoError(cmd.Run())

	cmd = exec.Command("go", "tool", "asmfmt", "-w", baseDir)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	assertNoError(cmd.Run())

	cmd = exec.Command("go", "tool", "goimports", "-w", baseDir)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	assertNoError(cmd.Run())

	stopOnce.Do(func() { close(stopSpinner) })
	fmt.Fprintf(os.Stderr, "\r\033[Kgenerated %d files in %s\n", gen.FilesCount(), time.Since(start))
}

func spinner(stop chan struct{}) {
	ticker := time.NewTicker(100 * time.Millisecond)
	defer ticker.Stop()
	chars := []rune{'|', '/', '-', '\\'}
	i := 0
	for {
		select {
		case <-stop:
			return
		case <-ticker.C:
			fmt.Fprintf(os.Stderr, "\r\033[Kgenerating files ... %c", chars[i%len(chars)])
			i++
		}
	}
}

func assertNoError(err error) {
	if err != nil {
		stopOnce.Do(func() { close(stopSpinner) })
		fmt.Fprintf(os.Stderr, "\n%s\n", err.Error())
		os.Exit(-1)
	}
}

// cleanGeneratedFiles removes all previously generated files (those with the
// "Code generated by consensys/gnark-crypto DO NOT EDIT" header) from dir and
// its subdirectories. This prevents stale generated files from persisting when
// the generation logic changes.
func cleanGeneratedFiles(dir string) {
	const generatedHeader = "Code generated by consensys/gnark-crypto DO NOT EDIT"
	filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {
		if err != nil || info.IsDir() {
			return nil
		}
		if isGeneratedFile(path, generatedHeader) {
			os.Remove(path)
		}
		return nil
	})
}

func isGeneratedFile(path string, header string) bool {
	ext := filepath.Ext(path)
	if ext != ".go" && ext != ".s" {
		return false
	}
	f, err := os.Open(path)
	if err != nil {
		return false
	}
	buf := make([]byte, 4096)
	n, _ := f.Read(buf)
	f.Close()
	return n > 0 && strings.Contains(string(buf[:n]), header)
}
