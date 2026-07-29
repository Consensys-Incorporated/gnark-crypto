package field

import (
	"os"
	"path/filepath"

	"github.com/consensys/bavard"
	"github.com/consensys/gnark-crypto/internal/generator/field/asm/amd64"
	"github.com/consensys/gnark-crypto/internal/generator/field/config"
	"github.com/consensys/gnark-crypto/internal/generator/field/template"
)

func generateExtensions(F *config.Field, outputDir string) error {

	fieldImportPath, err := getImportPath(outputDir)
	if err != nil {
		return err
	}

	outputDir = filepath.Join(outputDir, "extensions")

	entries_ext2 := []bavard.Entry{
		{File: filepath.Join(outputDir, "doc.go"), Templates: []string{"doc.go.tmpl"}},
		{File: filepath.Join(outputDir, "utils.go"), Templates: []string{"utils.go.tmpl"}},
		{File: filepath.Join(outputDir, "e2.go"), Templates: []string{"e2.go.tmpl"}},
		{File: filepath.Join(outputDir, "e2_test.go"), Templates: []string{"e2_test.go.tmpl"}},
	}

	type extensionsTemplateData struct {
		FF               string
		FieldPackagePath string
		F31              bool
		Q, QInvNeg       uint64
		IsKoalaBear      bool
		IsBabyBear       bool
		// QuadraticNonResidue is α, where E2 = Fr[u]/(u²-α).
		QuadraticNonResidue uint64
		// E6ReplA / E6ReplB are the VPERMI2D index tables for the gather-free E6
		// scalar-mul kernels, split into rows of 16 for the generated Go literals.
		// See amd64.E6ReplicationTables.
		E6ReplA, E6ReplB [][]uint32
	}

	isKoalaBear := F.Q[0] == 2130706433
	isBabyBear := F.Q[0] == 2013265921
	var quadraticNonResidue uint64
	switch {
	case isKoalaBear:
		quadraticNonResidue = 3
	case isBabyBear:
		quadraticNonResidue = 11
	}
	data := &extensionsTemplateData{
		FF:                  F.PackageName,
		FieldPackagePath:    fieldImportPath,
		F31:                 F.F31,
		IsKoalaBear:         isKoalaBear,
		IsBabyBear:          isBabyBear,
		Q:                   F.Q[0],
		QInvNeg:             F.QInverse[0],
		QuadraticNonResidue: quadraticNonResidue,
	}

	if isKoalaBear {
		replA, replB := amd64.E6ReplicationTables()
		data.E6ReplA = chunkU32(replA, 16)
		data.E6ReplB = chunkU32(replB, 16)
	}

	g := NewGenerator(template.FS)

	if err := g.Generate(data, "extensions", "", "extensions", entries_ext2...); err != nil {
		return err
	}
	if F.F31 {
		entries_ext4 := []bavard.Entry{
			{File: filepath.Join(outputDir, "e4.go"), Templates: []string{"e4.go.tmpl"}},
			{File: filepath.Join(outputDir, "vector.go"), Templates: []string{"vector.go.tmpl"}},
			{File: filepath.Join(outputDir, "e4_test.go"), Templates: []string{"e4_test.go.tmpl"}},
			{File: filepath.Join(outputDir, "e6.go"), Templates: []string{"e6.go.tmpl"}},
			{File: filepath.Join(outputDir, "e6_test.go"), Templates: []string{"e6_test.go.tmpl"}},
		}

		if isKoalaBear {
			entries_ext4 = append(entries_ext4, bavard.Entry{File: filepath.Join(outputDir, "e4_amd64.go"), Templates: []string{"e4.amd64.go.tmpl"}, BuildTag: "!purego"})
			entries_ext4 = append(entries_ext4, bavard.Entry{File: filepath.Join(outputDir, "e4_purego.go"), Templates: []string{"e4.purego.go.tmpl"}, BuildTag: "purego || (!amd64)"})
		}

		if err := g.Generate(data, "extensions", "", "extensions", entries_ext4...); err != nil {
			return err
		}

		if isKoalaBear {
			// generate the assembly file;
			asmFile, err := os.Create(filepath.Join(outputDir, "e4_amd64.s"))
			if err != nil {
				return err
			}

			asmFile.WriteString("//go:build !purego\n")

			if err := amd64.GenerateF31E4(asmFile); err != nil {
				asmFile.Close()
				return err
			}
			asmFile.Close()
		}
	}

	return runFormatters(outputDir)
}

// chunkU32 splits s into consecutive rows of at most n elements, for laying out
// generated slice literals one logical table per line.
func chunkU32(s []uint32, n int) [][]uint32 {
	var rows [][]uint32
	for len(s) > 0 {
		k := min(n, len(s))
		rows = append(rows, s[:k])
		s = s[k:]
	}
	return rows
}
