// Copyright 2020 Consensys Software Inc.
//
// Licensed under the Apache License, Version 2.0 (the "License"); you may not
// use this file except in compliance with the License. You may obtain a copy of
// the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS, WITHOUT
// WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied. See the
// License for the specific language governing permissions and limitations under
// the License.

package merkletree

import (
	"crypto/sha256"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
)

// TestReadAllRejectsNonPositiveSegmentSize ensures ReadAll fails fast instead of
// looping forever (segmentSize == 0) or panicking in make([]byte, segmentSize)
// (segmentSize < 0). The error must reach callers such as ReaderRoot and
// BuildReaderProof, which both delegate to Tree.ReadAll.
func TestReadAllRejectsNonPositiveSegmentSize(t *testing.T) {
	tree := New(sha256.New())
	require.Error(t, tree.ReadAll(strings.NewReader("data"), 0))
	require.Error(t, tree.ReadAll(strings.NewReader("data"), -1))

	// The public helpers that route through ReadAll must also surface the error.
	_, err := ReaderRoot(strings.NewReader("data"), sha256.New(), 0)
	require.Error(t, err)
	_, _, _, err = BuildReaderProof(strings.NewReader("data"), sha256.New(), -1, 0)
	require.Error(t, err)
}

// TestReadAllValidSegmentSize is a regression guard proving the success path is
// unchanged: a positive segmentSize still builds a deterministic root.
func TestReadAllValidSegmentSize(t *testing.T) {
	data := []byte("abcdef") // 3 leaves of 2 bytes each

	tree := New(sha256.New())
	require.NoError(t, tree.ReadAll(strings.NewReader(string(data)), 2))
	root := tree.Root()
	require.NotEmpty(t, root)

	// ReaderRoot must produce the same root for the same input.
	root2, err := ReaderRoot(strings.NewReader(string(data)), sha256.New(), 2)
	require.NoError(t, err)
	require.Equal(t, root, root2)
}
