package eddsa

import (
	"path/filepath"

	"github.com/consensys/bavard"
	"github.com/consensys/gnark-crypto/internal/generator/common"
	"github.com/consensys/gnark-crypto/internal/generator/config"
	"github.com/consensys/gnark-crypto/internal/generator/edwards/eddsa/template"
)

func Generate(conf config.TwistedEdwardsCurve, baseDir string, gen *common.Generator) error {
	// eddsa; conf.Package stays the curve package ("twistededwards" or
	// "bandersnatch") so that the templates import the curve they sign on.
	baseDir = filepath.Join(baseDir, "eddsa")

	entries := []bavard.Entry{
		{File: filepath.Join(baseDir, "doc.go"), Templates: []string{"doc.go.tmpl"}},
		{File: filepath.Join(baseDir, "eddsa.go"), Templates: []string{"eddsa.go.tmpl"}},
		{File: filepath.Join(baseDir, "eddsa_test.go"), Templates: []string{"eddsa.test.go.tmpl"}},
		{File: filepath.Join(baseDir, "marshal.go"), Templates: []string{"marshal.go.tmpl"}},
	}
	eddsaGen := common.NewDefaultGenerator(template.FS)
	return eddsaGen.Generate(conf, "eddsa", "", "", entries...)

}
