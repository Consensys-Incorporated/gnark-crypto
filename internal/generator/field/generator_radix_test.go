package field

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	"github.com/consensys/gnark-crypto/internal/generator/field/config"
)

// TestRadixSubWordIntegration generates the element layer for a field whose
// Montgomery radix is narrower than its storage word, then runs the generated
// property tests against it.
//
// The generated element_test.go checks every operation against math/big, so
// this exercises the radix-52 montMul, the conditional subtracts layered on top
// of its lazy [0, 9q/8) output, Inverse, Sqrt (Sarkar, since the 2-adicity is
// 34) and Cbrt. Without this, nothing compiles the sub-word radix path until
// the field is registered for real.
func TestRadixSubWordIntegration(t *testing.T) {
	if testing.Short() {
		t.Skip("generates and compiles a field package")
	}

	const rootDir = "radix_integration_test"
	// p = 2^49 - 2^34 + 1, the mamabear modulus, with R = 2^52.
	const modulus = "0x1FFFC00000001"

	if err := os.RemoveAll(rootDir); err != nil {
		t.Fatal(err)
	}
	if err := os.MkdirAll(rootDir, 0o700); err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(rootDir)

	f, err := config.NewFieldConfig("radixsubword", "Element", modulus, true, config.WithMontgomeryRadixBits(52))
	if err != nil {
		t.Fatal(err)
	}
	if !f.RadixSubWord || f.RBits != 52 {
		t.Fatalf("expected a sub-word radix field, got RadixSubWord=%v RBits=%d", f.RadixSubWord, f.RBits)
	}

	childDir := filepath.Join(rootDir, "radixsubword")
	if err := GenerateFF(f, childDir); err != nil {
		t.Fatal(err)
	}

	cmd := exec.Command("go", "test")
	cmd.Dir = childDir
	var out strings.Builder
	cmd.Stdout = &out
	cmd.Stderr = &out
	if err := cmd.Run(); err != nil {
		t.Fatalf("go test on the generated sub-word radix field failed:\n%s\n%s", out.String(), err)
	}
}
