package field

import (
	"bytes"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
)

// TestAssemblyTrimpath builds the same field operations from independent source
// roots and caches. Assembly includes outside the package must not leak those
// roots into either ordinary or stripped binaries, including when vendored.
func TestAssemblyTrimpath(t *testing.T) {
	if runtime.GOARCH != "amd64" && runtime.GOARCH != "arm64" {
		t.Skip("field assembly is only generated for amd64 and arm64")
	}
	t.Setenv("GOWORK", "off")
	t.Setenv("GOFLAGS", "")
	t.Setenv("CGO_ENABLED", "0")

	repo, err := filepath.Abs(filepath.Join("..", "..", ".."))
	if err != nil {
		t.Fatal(err)
	}
	fixture := filepath.Join(t.TempDir(), "fixture")
	module := filepath.Join(fixture, "gnark-crypto")
	if err := os.MkdirAll(module, 0o700); err != nil {
		t.Fatal(err)
	}

	// Copy the actual field packages and their in-module dependencies, including
	// all architecture variants. Keep the replacement relative so both copies
	// describe exactly the same build, apart from their filesystem locations.
	dirs := assemblyGo(t, repo, nil, "list", "-deps", "-f",
		`{{if not .Standard}}{{if eq .Module.Path "github.com/consensys/gnark-crypto"}}{{.Dir}}{{end}}{{end}}`,
		"./ecc/bls12-381/fp", "./ecc/bls12-381/fr")
	for _, dir := range strings.Split(strings.TrimSpace(string(dirs)), "\n") {
		if dir == "" {
			continue
		}
		rel, err := filepath.Rel(repo, dir)
		if err != nil {
			t.Fatal(err)
		}
		copyAssemblyFiles(t, dir, filepath.Join(module, rel))
	}
	for _, name := range []string{"go.mod", "go.sum"} {
		data, err := os.ReadFile(filepath.Join(repo, name))
		if err != nil {
			t.Fatal(err)
		}
		writeAssemblyFile(t, filepath.Join(module, name), data)
	}
	writeAssemblyFile(t, filepath.Join(fixture, "go.mod"), []byte(`module example.com/assemblytrimpath

go 1.23.0

require github.com/consensys/gnark-crypto v0.0.0

replace github.com/consensys/gnark-crypto => ./gnark-crypto
`))
	writeAssemblyFile(t, filepath.Join(fixture, "main.go"), []byte(`package main

import (
	"github.com/consensys/gnark-crypto/ecc/bls12-381/fp"
	"github.com/consensys/gnark-crypto/ecc/bls12-381/fr"
)

func main() {
	var p fp.Element
	p.SetUint64(3)
	p.Square(&p)
	if p.Uint64() != 9 {
		panic("incorrect fp square")
	}
	var r fr.Element
	r.SetUint64(5)
	r.Square(&r)
	if r.Uint64() != 25 {
		panic("incorrect fr square")
	}
}
`))
	assemblyGo(t, fixture, nil, "mod", "tidy")
	assemblyGo(t, fixture, nil, "mod", "vendor")

	for _, mode := range []string{"mod", "vendor"} {
		t.Run(mode, func(t *testing.T) {
			var reference [2][]byte
			for i, name := range []string{"a", "longer source path with spaces"} {
				root := filepath.Join(t.TempDir(), name)
				if err := os.CopyFS(root, os.DirFS(fixture)); err != nil {
					t.Fatal(err)
				}
				// Each source root and module layout gets a fresh cache. Reusing
				// objects compiled in another root can conceal the path leak.
				env := []string{"GOCACHE=" + t.TempDir()}
				for j, flags := range []string{"", "-s -w"} {
					binary := filepath.Join(t.TempDir(), "repro.exe")
					assemblyGo(t, root, env, "build", "-trimpath", "-buildvcs=false",
						"-mod="+mode, "-ldflags="+flags, "-o", binary, ".")
					data, err := os.ReadFile(binary)
					if err != nil {
						t.Fatal(err)
					}
					if bytes.Contains(data, []byte(filepath.ToSlash(root))) {
						t.Errorf("binary contains source root %q (ldflags=%q)", root, flags)
					}
					for _, field := range []string{"fp", "fr"} {
						file := "/ecc/bls12-381/" + field + "/element_" + runtime.GOARCH + ".s"
						if !bytes.Contains(data, []byte(file)) {
							t.Errorf("binary is missing assembly source location %q (ldflags=%q)", file, flags)
						}
					}
					if i == 0 {
						reference[j] = data
					} else if !bytes.Equal(reference[j], data) {
						t.Errorf("binaries differ between source roots (ldflags=%q)", flags)
					}
					if output, err := exec.Command(binary).CombinedOutput(); err != nil {
						t.Fatalf("field operations failed: %v\n%s", err, output)
					}
				}
			}
		})
	}
}

func assemblyGo(t *testing.T, dir string, env []string, args ...string) []byte {
	t.Helper()
	cmd := exec.Command("go", args...)
	cmd.Dir = dir
	cmd.Env = append(os.Environ(), env...)
	output, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("go %s in %s: %v\n%s", strings.Join(args, " "), dir, err, output)
	}
	return output
}

func copyAssemblyFiles(t *testing.T, src, dst string) {
	t.Helper()
	if err := os.MkdirAll(dst, 0o700); err != nil {
		t.Fatal(err)
	}
	entries, err := os.ReadDir(src)
	if err != nil {
		t.Fatal(err)
	}
	for _, entry := range entries {
		if entry.IsDir() || strings.HasSuffix(entry.Name(), "_test.go") {
			continue
		}
		data, err := os.ReadFile(filepath.Join(src, entry.Name()))
		if err != nil {
			t.Fatal(err)
		}
		writeAssemblyFile(t, filepath.Join(dst, entry.Name()), data)
	}
}

func writeAssemblyFile(t *testing.T, path string, data []byte) {
	t.Helper()
	// #nosec G703 -- Callers construct paths in t.TempDir from repository file names.
	if err := os.WriteFile(path, data, 0o600); err != nil {
		t.Fatal(err)
	}
}
