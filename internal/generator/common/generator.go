package common

import (
	"embed"
	"fmt"
	"os/exec"
	"path"
	"path/filepath"
	"strings"
	"sync/atomic"

	"github.com/consensys/bavard"
)

var (
	filesCount uint64
)

// Generator is a wrapper around bavard.BatchGenerator that handles embed.FS and goimports/asmfmt
type Generator struct {
	fs              embed.FS
	copyrightHolder string
	copyrightYear   int
	generatedBy     string
}

func (g *Generator) FilesCount() uint64 {
	return atomic.LoadUint64(&filesCount)
}

// NewGenerator returns a new Generator
func NewGenerator(fs embed.FS, copyrightHolder string, copyrightYear int, generatedBy string) *Generator {
	return &Generator{
		fs:              fs,
		copyrightHolder: copyrightHolder,
		copyrightYear:   copyrightYear,
		generatedBy:     generatedBy,
	}
}

// NewDefaultGenerator returns a new Generator with default gnark-crypto settings
func NewDefaultGenerator(fs embed.FS) *Generator {
	return NewGenerator(fs, "Consensys Software Inc.", 2020, "consensys/gnark-crypto")
}

// Generate generates files using the provided data and entries
func (g *Generator) Generate(data any, packageName string, outputDir string, templateDir string, entries ...bavard.Entry) error {
	return g.GenerateWithOptions(data, packageName, outputDir, templateDir, nil, entries...)
}

// GenerateWithOptions generates files using the provided data, entries and options
func (g *Generator) GenerateWithOptions(data any, packageName string, outputDir string, templateDir string, opts []func(*bavard.Bavard) error, entries ...bavard.Entry) error {
	for _, entry := range entries {
		var tmpls []string
		for _, t := range entry.Templates {
			name := t
			if templateDir != "" {
				// embed.FS paths always use forward slashes, so use path rather
				// than filepath, which joins with a backslash on Windows.
				name = path.Join(templateDir, t)
			}
			b, err := g.fs.ReadFile(name)
			if err != nil {
				return fmt.Errorf("error reading template %s: %w", name, err)
			}
			tmpls = append(tmpls, string(b)+"\n")
		}

		bavardOpts := []func(*bavard.Bavard) error{
			bavard.Apache2(g.copyrightHolder, g.copyrightYear),
			bavard.GeneratedBy(g.generatedBy),
			bavard.Funcs(Funcs()),
			bavard.Verbose(false),
			bavard.Import(false),
			bavard.Format(false), // we run formatters on the whole directory at the end
		}
		if !strings.HasSuffix(entry.File, ".s") {
			bavardOpts = append(bavardOpts, bavard.Package(packageName))
		}
		bavardOpts = append(bavardOpts, opts...)
		if entry.BuildTag != "" {
			bavardOpts = append(bavardOpts, bavard.BuildTag(entry.BuildTag))
		}

		outputFile := entry.File
		if outputDir != "" && !filepath.IsAbs(outputFile) {
			outputFile = filepath.Join(outputDir, outputFile)
		}

		if err := bavard.GenerateFromString(outputFile, tmpls, data, bavardOpts...); err != nil {
			return err
		}
		atomic.AddUint64(&filesCount, 1)
		// we format the whole directory when done, no need to do it file by file
		// if strings.HasSuffix(outputFile, ".go") {
		// 	_ = exec.Command("goimports", "-w", outputFile).Run()
		// } else if strings.HasSuffix(outputFile, ".s") {
		// 	_ = exec.Command("asmfmt", "-w", outputFile).Run()
		// }
	}
	return nil
}

// RunFormatters runs gofmt and asmfmt on the provided directory
func RunFormatters(outputDir string) error {
	var out strings.Builder
	{
		// run go fmt on whole directory
		cmd := exec.Command("gofmt", "-s", "-w", outputDir)
		cmd.Stdout = &out
		cmd.Stderr = &out
		if err := cmd.Run(); err != nil {
			return fmt.Errorf("gofmt failed: %v\n%s", err, out.String())
		}
	}
	{
		// run asmfmt on whole directory
		cmd := exec.Command("go", "tool", "asmfmt", "-w", outputDir)
		cmd.Stdout = &out
		cmd.Stderr = &out
		if err := cmd.Run(); err != nil {
			return fmt.Errorf("asmfmt failed: %v\n%s", err, out.String())
		}
	}
	{
		// run goimports on whole directory
		cmd := exec.Command("go", "tool", "goimports", "-w", outputDir)
		cmd.Stdout = &out
		cmd.Stderr = &out
		if err := cmd.Run(); err != nil {
			return fmt.Errorf("goimports failed: %v\n%s", err, out.String())
		}
	}
	return nil
}
