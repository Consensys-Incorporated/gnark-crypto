package amd64

import (
	"path/filepath"
	"testing"
)

// TestElementASMPathsAgree pins the two path helpers together.
//
// ElementASMFileName derives the directory from nbBits, while
// ElementASMBaseDir used to hardcode 31 for every single-word field. A 49-bit
// field therefore wrote its assembly to element_49b/ but reported
// element_31b/ as its package path, so the blank import that keeps `go mod
// vendor` from dropping the .s files pointed at the wrong directory. That
// compiles and links fine, which is exactly why it needs a test.
func TestElementASMPathsAgree(t *testing.T) {
	t.Parallel()

	for _, tc := range []struct {
		nbWords, nbBits int
		wantDir         string
	}{
		{1, 31, "element_31b"},
		{1, 49, "element_49b"},
		{4, 254, "element_4w"},
		{6, 377, "element_6w"},
	} {
		dir := ElementASMBaseDir(tc.nbWords, tc.nbBits)
		if dir != tc.wantDir {
			t.Errorf("ElementASMBaseDir(%d, %d) = %q, want %q", tc.nbWords, tc.nbBits, dir, tc.wantDir)
		}
		fileDir := filepath.Dir(ElementASMFileName(tc.nbWords, tc.nbBits))
		if fileDir != dir {
			t.Errorf("ElementASMFileName(%d, %d) lives in %q but ElementASMBaseDir says %q",
				tc.nbWords, tc.nbBits, fileDir, dir)
		}
	}
}
