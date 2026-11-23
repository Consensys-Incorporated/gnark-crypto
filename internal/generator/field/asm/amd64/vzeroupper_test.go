package amd64

import (
	"bufio"
	"io/fs"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"testing"
)

// TestVZEROUPPER checks every amd64 assembly function of the repository: one
// that uses YMM / ZMM registers must execute VZEROUPPER right before each RET
// (see FFAmd64.RETAVX), and one that does not must not, as it may run on CPUs
// without AVX.
func TestVZEROUPPER(t *testing.T) {
	root, err := filepath.Abs("../../../../..")
	if err != nil {
		t.Fatal(err)
	}
	if _, err := os.Stat(filepath.Join(root, "go.mod")); err != nil {
		t.Fatalf("repository root not found at %s: %v", root, err)
	}

	skip, err := ignoredDirs(root)
	if err != nil {
		t.Fatal(err)
	}

	var files []string
	for _, dir := range []string{"ecc", "field"} {
		err := filepath.WalkDir(filepath.Join(root, dir), func(path string, d fs.DirEntry, err error) error {
			if err != nil {
				return err
			}
			rel, err := filepath.Rel(root, path)
			if err != nil {
				return err
			}
			rel = filepath.ToSlash(rel)
			if d.IsDir() {
				if skip[rel] {
					return filepath.SkipDir
				}
				return nil
			}
			if strings.HasSuffix(rel, "_amd64.s") {
				files = append(files, rel)
			}
			return nil
		})
		if err != nil {
			t.Fatal(err)
		}
	}
	if len(files) == 0 {
		t.Fatal("no amd64 assembly file found")
	}

	vecReg := regexp.MustCompile(`[\s,(]([YZ])([0-9]|[12][0-9]|3[01])\b`)
	nbFuncs := 0
	for _, rel := range files {
		checkFile(t, filepath.Join(root, rel), rel, vecReg, &nbFuncs)
	}
	if nbFuncs == 0 {
		t.Fatal("no amd64 assembly function found")
	}
}

// ignoredDirs returns the directories (relative to root, slash-separated)
// that hold uncommitted outputs: the dev scratch package, and the
// integration-test outputs listed in field/asm/.gitignore, which may be stale.
func ignoredDirs(root string) (map[string]bool, error) {
	skip := map[string]bool{"field/internal/dev": true}
	data, err := os.ReadFile(filepath.Join(root, "field", "asm", ".gitignore"))
	if err != nil {
		return nil, err
	}
	for _, line := range strings.Split(string(data), "\n") {
		line = strings.TrimSpace(line)
		if dir, ok := strings.CutSuffix(line, "/**"); ok && !strings.HasPrefix(line, "#") {
			skip["field/asm/"+dir] = true
		}
	}
	return skip, nil
}

// checkFile runs the TestVZEROUPPER check on the assembly file at path.
func checkFile(t *testing.T, path, rel string, vecReg *regexp.Regexp, nbFuncs *int) {
	t.Helper()
	f, err := os.Open(path)
	if err != nil {
		t.Fatal(err)
	}
	defer f.Close()

	var name, prev string
	var avx, hasVzeroupper, inMacro bool
	var rets, cleared int
	check := func() {
		if name == "" {
			return
		}
		*nbFuncs++
		if avx && cleared < rets {
			t.Errorf("%s: %s uses YMM/ZMM registers but returns without VZEROUPPER", rel, name)
		}
		if !avx && hasVzeroupper {
			t.Errorf("%s: %s does not use YMM/ZMM registers but executes VZEROUPPER", rel, name)
		}
	}
	sc := bufio.NewScanner(f)
	sc.Buffer(make([]byte, 1<<20), 1<<20)
	for sc.Scan() {
		line := sc.Text()
		// macro bodies are checked where the macros are expanded
		if inMacro || strings.HasPrefix(strings.TrimSpace(line), "#define") {
			inMacro = strings.HasSuffix(strings.TrimSpace(line), "\\")
			continue
		}
		code := strings.TrimSpace(line)
		if i := strings.Index(code, "//"); i >= 0 {
			code = strings.TrimSpace(code[:i])
		}
		switch {
		case code == "":
			continue
		case strings.HasPrefix(code, "TEXT "):
			check()
			name, avx, hasVzeroupper, rets, cleared = strings.Fields(code)[1], false, false, 0, 0
		case code == "VZEROUPPER":
			hasVzeroupper = true
		case code == "RET":
			rets++
			if prev == "VZEROUPPER" {
				cleared++
			}
		default:
			if vecReg.MatchString(" " + code) {
				avx = true
			}
		}
		prev = code
	}
	check()
	if err := sc.Err(); err != nil {
		t.Fatal(err)
	}
}
