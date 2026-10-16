package hashadapter

import (
	"path/filepath"

	"github.com/consensys/bavard"
	"github.com/consensys/gnark-crypto/internal/generator/common"
	"github.com/consensys/gnark-crypto/internal/generator/field/config"
	"github.com/consensys/gnark-crypto/internal/generator/hashadapter/template"
)

// Generate generates, in baseDir, the files adapters.go and adapters_test.go
// holding, for each of confs, an adapter that hashes the elements of the
// extension with a hasher over its base field. The confs must share the package
// that defines the extensions.
func Generate(baseDir string, confs ...config.FieldDependency) error {
	data := struct {
		FieldPackagePath     string
		BaseFieldPackagePath string
		Adapters             []config.FieldDependency
	}{confs[0].FieldPackagePath, confs[0].BaseFieldPackagePath, confs}

	return common.NewDefaultGenerator(template.FS).Generate(data, "hash", "", "",
		bavard.Entry{File: filepath.Join(baseDir, "adapters.go"), Templates: []string{"adapters.go.tmpl"}},
		bavard.Entry{File: filepath.Join(baseDir, "adapters_test.go"), Templates: []string{"adapters_test.go.tmpl"}},
	)
}
