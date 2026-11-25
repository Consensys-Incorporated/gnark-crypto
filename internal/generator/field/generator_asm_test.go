package field

import (
	"bytes"
	"go/build/constraint"
	"os"
	"path/filepath"
	"testing"

	curveconfig "github.com/consensys/gnark-crypto/internal/generator/config"
	"github.com/consensys/gnark-crypto/internal/generator/field/asm/amd64"
	"github.com/consensys/gnark-crypto/internal/generator/field/asm/arm64"
	"github.com/consensys/gnark-crypto/internal/generator/field/config"
)

// TestGenerateAssembly checks that every enabled architecture gets the complete
// shared source in its package, including when fields generate concurrently and
// share an assembly cache entry. Unsupported and disabled variants stay absent.
func TestGenerateAssembly(t *testing.T) {
	root := t.TempDir()
	writeAssemblyFile(t, filepath.Join(root, "go.mod"), []byte("module example.com/assembly\n"))
	asmDir := filepath.Join(root, "asm")
	for _, tc := range []struct {
		name    string
		modulus string
		asm     bool
	}{
		{"bls12_381_fr", curveconfig.BLS12_381.FrModulus, true},
		{"bn254_fr", curveconfig.BN254.FrModulus, true},
		{"bls12_381_fp", curveconfig.BLS12_381.FpModulus, true},
		{"bls24_315_fp", curveconfig.BLS24_315.FpModulus, true},
		{"babybear", "2013265921", true},
		{"purego", "47", true},
		{"disabled", curveconfig.BLS12_381.FrModulus, false},
	} {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			f, err := config.NewFieldConfig(tc.name, "Element", tc.modulus, false)
			if err != nil {
				t.Fatal(err)
			}
			out := filepath.Join(root, tc.name)
			var opts []Option
			if tc.asm {
				opts = append(opts, WithASM(&config.Assembly{BuildDir: asmDir, IncludeDir: "../asm"}))
			}
			if err := GenerateFF(f, out, opts...); err != nil {
				t.Fatal(err)
			}
			for _, arch := range []struct {
				name    string
				enabled bool
				file    string
			}{
				{"amd64", f.GenerateOpsAMD64, amd64.ElementASMFileName(f.NbWords, f.NbBits)},
				{"arm64", f.GenerateOpsARM64, arm64.ElementASMFileName(f.NbWords, f.NbBits)},
			} {
				data, err := os.ReadFile(filepath.Join(out, "element_"+arch.name+".s"))
				if !tc.asm || !arch.enabled {
					if !os.IsNotExist(err) {
						t.Errorf("unexpected %s assembly: %v", arch.name, err)
					}
					continue
				}
				if err != nil {
					t.Fatal(err)
				}
				shared, err := os.ReadFile(filepath.Join(asmDir, arch.file))
				if err != nil {
					t.Fatal(err)
				}
				if len(shared) == 0 || !bytes.Contains(data, shared) {
					t.Errorf("%s assembly does not contain the complete shared source", arch.name)
				}
				line, _, _ := bytes.Cut(data, []byte("\n"))
				tag, err := constraint.Parse(string(line))
				if err != nil || tag.String() != "!purego" {
					t.Errorf("%s assembly is missing its purego constraint", arch.name)
				}
				if bytes.Contains(data, []byte(`#include "..`)) {
					t.Errorf("%s assembly includes source outside its package", arch.name)
				}
			}
		})
	}
}
