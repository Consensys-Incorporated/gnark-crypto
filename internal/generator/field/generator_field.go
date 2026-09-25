package field

import (
	"path/filepath"

	"github.com/consensys/bavard"
	"github.com/consensys/gnark-crypto/internal/generator/addchain"
	"github.com/consensys/gnark-crypto/internal/generator/common"
	"github.com/consensys/gnark-crypto/internal/generator/field/asm/amd64"
	"github.com/consensys/gnark-crypto/internal/generator/field/asm/arm64"
	"github.com/consensys/gnark-crypto/internal/generator/field/config"
	"github.com/consensys/gnark-crypto/internal/generator/field/template"
	"golang.org/x/sync/errgroup"
)

func generateField(F *config.Field, outputDir, asmDirIncludePath, hashArm64, hashAMD64 string) error {

	// source file templates
	sourceFiles := []string{
		"element/base.go.tmpl",
		"element/reduce.go.tmpl",
		"element/exp.go.tmpl",
		"element/conv.go.tmpl",
		"element/muldoc.go.tmpl",
		"element/mulcios.go.tmpl",
		"element/mulnocarry.go.tmpl",
		"element/sqrt.go.tmpl",
		"element/cbrt.go.tmpl",
		"element/inverse.go.tmpl",
		"element/bignum.go.tmpl",
	}

	// test file templates
	testFiles := []string{
		"element/mulcios.go.tmpl",
		"element/mulnocarry.go.tmpl",
		"element/reduce.go.tmpl",
		"element/test.go.tmpl",
		"element/inversetests.go.tmpl",
	}
	funcs := common.Funcs()
	if F.UseAddChain {
		for _, f := range addchain.Functions {
			funcs[f.Name] = f.Func
		}
	}

	gen := NewGenerator(template.FS)

	generate := func(suffix string, templateNames []string, opts ...fieldOption) func() error {
		opt := fieldOptions(opts...)
		if opt.skip {
			return func() error { return nil }
		}
		return func() error {
			bavardOpts := []func(*bavard.Bavard) error{
				bavard.Funcs(funcs),
			}

			tmplData := any(F)
			if opt.tmplData != nil {
				tmplData = opt.tmplData
			}

			entry := bavard.Entry{
				File:      suffix,
				Templates: templateNames,
				BuildTag:  opt.buildTag,
			}

			return gen.GenerateWithOptions(tmplData, F.PackageName, outputDir, "", bavardOpts, entry)
		}
	}

	// generate asm files;
	// couple of cases;
	// 1. we generate arm64 and amd64
	// 2. we generate only amd64
	// 3. we generate only purego

	// sanity check
	if (F.GenerateOpsARM64 && !F.GenerateOpsAMD64) ||
		(F.GenerateVectorOpsAMD64 && !F.GenerateOpsAMD64) ||
		(F.GenerateVectorOpsARM64 && !F.GenerateOpsARM64) {
		panic("not implemented.")
	}

	// get hash of the common asm files to force compiler to recompile in case of changes.
	var amd64d, arm64d ASMWrapperData
	var err error

	if F.GenerateOpsAMD64 {
		amd64d, err = newASMWrapperData(hashAMD64, asmDirIncludePath, amd64.ElementASMFileName(F.NbWords, F.NbBits))
		if err != nil {
			return err
		}
	}

	if F.GenerateOpsARM64 {
		arm64d, err = newASMWrapperData(hashArm64, asmDirIncludePath, arm64.ElementASMFileName(F.NbWords, F.NbBits))
		if err != nil {
			return err
		}
	}

	if F.GenerateOpsAMD64 || F.GenerateOpsARM64 {
		F.ASMPackagePath, err = getImportPath(filepath.Join(outputDir, asmDirIncludePath, amd64.ElementASMBaseDir(F.NbWords, F.NbBits)))
		if err != nil {
			return err
		}
	}

	// purego files have no build tags if we don't generate asm
	pureGoBuildTag := "purego || (!amd64 && !arm64)"
	if (!F.GenerateOpsAMD64 || hashAMD64 == "") && (!F.GenerateOpsARM64 || hashArm64 == "") {
		pureGoBuildTag = ""
	} else if !F.GenerateOpsARM64 || hashArm64 == "" {
		pureGoBuildTag = "purego || (!amd64)"
	}

	pureGoVectorBuildTag := "purego || (!amd64 && !arm64)"
	if (!F.GenerateVectorOpsAMD64 || hashAMD64 == "") && (!F.GenerateVectorOpsARM64 || hashArm64 == "") {
		pureGoVectorBuildTag = ""
	} else if !F.GenerateVectorOpsARM64 || hashArm64 == "" {
		pureGoVectorBuildTag = "purego || (!amd64)"
	}

	if F.F31 || F.RadixSubWord {
		// element.go calls into the generic Mul/Square unconditionally for these
		// fields, so the "purego" file must always be compiled.
		pureGoBuildTag = ""
	}

	var g errgroup.Group

	g.Go(generate("element.go", sourceFiles))
	g.Go(generate("doc.go", []string{"element/doc.go.tmpl"}))
	g.Go(generate("vector.go", []string{"element/vector.go.tmpl"}))
	g.Go(generate("arith.go", []string{"element/arith.go.tmpl"}, only(!F.F31 && !F.RadixSubWord)))
	g.Go(generate("element_test.go", testFiles))
	g.Go(generate("vector_test.go", []string{"element/testvector.go.tmpl"}))

	g.Go(generate("element_amd64.s", []string{"element/asm_include.s.tmpl"}, only(F.GenerateOpsAMD64 && hashAMD64 != ""), withBuildTag("!purego"), withData(amd64d)))
	g.Go(generate("element_arm64.s", []string{"element/asm_include.s.tmpl"}, only(F.GenerateOpsARM64 && hashArm64 != ""), withBuildTag("!purego"), withData(arm64d)))

	g.Go(generate("element_amd64.go", []string{"element/opsamd64.go.tmpl", "element/muldoc.go.tmpl"}, only(F.GenerateOpsAMD64 && !F.F31 && !F.RadixSubWord && hashAMD64 != ""), withBuildTag("!purego")))
	g.Go(generate("element_arm64.go", []string{"element/opsarm64.go.tmpl", "element/mulnocarry.go.tmpl", "element/reduce.go.tmpl"}, only(F.GenerateOpsARM64 && !F.F31 && !F.RadixSubWord && hashArm64 != ""), withBuildTag("!purego")))

	g.Go(generate("element_purego.go", []string{"element/opsnoasm.go.tmpl", "element/mulradix52.go.tmpl", "element/mulcios.go.tmpl", "element/mulnocarry.go.tmpl", "element/reduce.go.tmpl", "element/muldoc.go.tmpl"}, withBuildTag(pureGoBuildTag)))

	g.Go(generate("vector_amd64.go", []string{"element/vectoropsamd64.go.tmpl"}, only(F.GenerateVectorOpsAMD64 && !F.F31 && hashAMD64 != ""), withBuildTag("!purego")))
	g.Go(generate("vector_amd64.go", []string{"element/vectoropsamd64f31.go.tmpl"}, only(F.GenerateVectorOpsAMD64 && F.F31 && hashAMD64 != ""), withBuildTag("!purego")))
	g.Go(generate("vector_arm64.go", []string{"element/vectoropspurego.go.tmpl"}, only(F.GenerateVectorOpsARM64 && !F.F31 && hashArm64 != ""), withBuildTag("!purego")))
	g.Go(generate("vector_arm64.go", []string{"element/vectoropsarm64f31.go.tmpl"}, only(F.GenerateVectorOpsARM64 && F.F31 && hashArm64 != ""), withBuildTag("!purego")))

	g.Go(generate("vector_purego.go", []string{"element/vectoropspurego.go.tmpl"}, withBuildTag(pureGoVectorBuildTag)))

	if F.UseAddChain {
		g.Go(generate("element_exp.go", []string{"element/fixedexp.go.tmpl"}))
	}

	return g.Wait()
}

type fieldOption func(*fieldConfig)
type fieldConfig struct {
	buildTag string
	skip     bool
	tmplData any
}

func withBuildTag(buildTag string) fieldOption {
	return func(opt *fieldConfig) {
		opt.buildTag = buildTag
	}
}

func only(condition bool) fieldOption {
	return func(opt *fieldConfig) {
		opt.skip = !condition
	}
}

func withData(data any) fieldOption {
	return func(opt *fieldConfig) {
		opt.tmplData = data
	}
}

// default options
func fieldOptions(opts ...fieldOption) fieldConfig {
	// apply options
	opt := fieldConfig{}
	for _, option := range opts {
		option(&opt)
	}
	return opt
}
