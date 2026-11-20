package amd64

import (
	"bufio"
	"os"
	"os/exec"
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

	// only the committed files: generated directories may hold stale,
	// git-ignored outputs.
	out, err := exec.Command("git", "-C", root, "ls-files", "*_amd64.s").Output()
	if err != nil {
		t.Skipf("git ls-files: %v", err)
	}
	files := strings.Fields(string(out))
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

// checkFile runs the TestVZEROUPPER check on the assembly file at path.
func checkFile(t *testing.T, path, rel string, vecReg *regexp.Regexp, nbFuncs *int) {
	t.Helper()
	{
		f, err := os.Open(path)
		if err != nil {
			t.Fatal(err)
		}
		defer f.Close()

		var name, prev string
		var avx, inMacro bool
		var rets, cleared int
		check := func() {
			if name == "" {
				return
			}
			*nbFuncs++
			if avx && cleared < rets {
				t.Errorf("%s: %s uses YMM/ZMM registers but returns without VZEROUPPER", rel, name)
			}
			if !avx && cleared > 0 {
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
				name, avx, rets, cleared = strings.Fields(code)[1], false, 0, 0
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
}
