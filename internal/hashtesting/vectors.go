// Copyright 2020-2026 Consensys Incorporated
// Licensed under the Apache License, Version 2.0. See the LICENSE file for details.

package hashtesting

import (
	"encoding/hex"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"hash"
	"io/fs"
	"math/rand/v2"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"
)

var update = flag.Bool("update", false, "rewrite the test vectors from the current implementation")

// Vector is a test vector of a byte hasher: the digest of the writes In.
type Vector struct {
	In  []string `json:"in"`  // the hex encodings of the inputs of successive calls to Write
	Out string   `json:"out"` // the hex encoding of the digest
}

// CheckVectors checks the hasher against the vectors in the JSON file at path.
// With -update, it rewrites the file from the current implementation instead:
// with the outputs for its inputs if the file exists, and otherwise with
// standard inputs of elements of elemSize bytes.
func CheckVectors(t *testing.T, newHasher func() hash.Hash, elemSize int, path string) {
	t.Helper()

	var vectors []Vector
	data, err := os.ReadFile(path)
	switch {
	case err == nil:
		require.NoError(t, json.Unmarshal(data, &vectors))
	case *update && errors.Is(err, fs.ErrNotExist):
		vectors = standardVectors(newHasher, elemSize)
	default:
		require.NoError(t, err)
	}

	if *update {
		for i := range vectors {
			vectors[i].Out = digest(t, newHasher, vectors[i].In)
		}
		data, err = json.MarshalIndent(vectors, "", "  ")
		require.NoError(t, err)
		require.NoError(t, os.MkdirAll(filepath.Dir(path), 0o755))
		require.NoError(t, os.WriteFile(path, append(data, '\n'), 0o600))
		t.Log("updated", path)
		return
	}

	for i, v := range vectors {
		t.Run(fmt.Sprint(i), func(t *testing.T) {
			require.Equal(t, v.Out, digest(t, newHasher, v.In))
		})
	}
}

// digest returns the hex encoding of the digest of the writes in.
func digest(t *testing.T, newHasher func() hash.Hash, in []string) string {
	t.Helper()
	h := newHasher()
	for _, s := range in {
		b, err := hex.DecodeString(s)
		require.NoError(t, err)
		_, err = h.Write(b)
		require.NoError(t, err)
	}
	return hex.EncodeToString(h.Sum(nil))
}

// standardVectors returns the inputs of a new file of vectors, without outputs.
// Random inputs are drawn from a fixed seed until newHasher accepts them, so that
// they are canonical at any alignment on elements.
func standardVectors(newHasher func() hash.Hash, elemSize int) []Vector {
	rng := rand.New(rand.NewPCG(1, 2)) //nolint:gosec // a fixed seed makes the standard inputs reproducible
	random := func(n int) []byte {
		b := make([]byte, n)
		for i := range b {
			b[i] = byte(rng.Uint32()) & 0x3f
		}
		return b
	}
	accepted := func(draw func() [][]byte) [][]byte {
		for range 1 << 20 {
			in := draw()
			h := newHasher()
			ok := true
			for _, b := range in {
				if _, err := h.Write(b); err != nil {
					ok = false
					break
				}
			}
			if ok {
				return in
			}
		}
		panic("hashtesting: no canonical standard input found")
	}
	// elements returns n random elements, the first byte of each being 0.
	elements := func(n int) []byte {
		var b []byte
		for range n {
			b = append(b, accepted(func() [][]byte {
				e := random(elemSize)
				e[0] = 0
				return [][]byte{e}
			})[0]...)
		}
		return b
	}

	var small []byte
	for i := 1; i <= 3; i++ {
		e := make([]byte, elemSize)
		e[elemSize-1] = byte(i)
		small = append(small, e...)
	}

	inputs := [][][]byte{
		{},
		{small},
		{elements(1)},
		{elements(1), elements(7), elements(20)},
		accepted(func() [][]byte { return [][]byte{random(5)} }),
		accepted(func() [][]byte {
			b := append(random(elemSize), random(5)...)
			b[0] = 0
			return [][]byte{b}
		}),
	}
	vectors := make([]Vector, len(inputs))
	for i, in := range inputs {
		vectors[i].In = make([]string, len(in))
		for j, b := range in {
			vectors[i].In[j] = hex.EncodeToString(b)
		}
	}
	return vectors
}
