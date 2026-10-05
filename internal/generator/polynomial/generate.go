package polynomial

import (
	"path/filepath"
	"strings"

	"github.com/consensys/gnark-crypto/internal/generator/common"
	"github.com/consensys/gnark-crypto/internal/generator/field/config"

	"github.com/consensys/bavard"
	"github.com/consensys/gnark-crypto/internal/generator/polynomial/template"
)

// Generate generates, in baseDir, the polynomial package over the field
// described by conf. For an extension, every free function and type is suffixed
// with the extension name (e.g. PolynomialE6) and the file names carry the
// lowercase name (e.g. polynomial_e6.go), so that several extensions can share
// a package. For a base field no suffix is added. doc.go is generated only when
// withDoc is set.
func Generate(conf config.FieldDependency, baseDir string, withDoc, generateTests bool, gen *common.Generator) error {
	ext := ""
	if name := conf.ExtensionName(); name != "" {
		ext = "_" + strings.ToLower(name)
	}

	var entries []bavard.Entry
	if withDoc {
		entries = append(entries, bavard.Entry{File: filepath.Join(baseDir, "doc.go"), Templates: []string{"doc.go.tmpl"}})
	}
	entries = append(entries,
		bavard.Entry{File: filepath.Join(baseDir, "polynomial"+ext+".go"), Templates: []string{"polynomial.go.tmpl"}},
		bavard.Entry{File: filepath.Join(baseDir, "multilin"+ext+".go"), Templates: []string{"multilin.go.tmpl"}},
		bavard.Entry{File: filepath.Join(baseDir, "pool"+ext+".go"), Templates: []string{"pool.go.tmpl"}},
	)

	if generateTests {
		entries = append(entries,
			bavard.Entry{File: filepath.Join(baseDir, "polynomial"+ext+"_test.go"), Templates: []string{"polynomial.test.go.tmpl"}},
			bavard.Entry{File: filepath.Join(baseDir, "multilin"+ext+"_test.go"), Templates: []string{"multilin.test.go.tmpl"}},
		)
	}

	polyGen := common.NewDefaultGenerator(template.FS)
	return polyGen.Generate(conf, "polynomial", "", "", entries...)
}
