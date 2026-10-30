// Copyright 2020-2026 Consensys Incorporated
// Licensed under the Apache License, Version 2.0. See the LICENSE file for details.

package all

import (
	"testing"

	"github.com/consensys/gnark-crypto/hash"
)

// TestDigestSizeMatchesHasher checks the static digestSize table against the
// hashers it describes. The two drifted apart for every small-field Poseidon2
// hash: the table recorded the size of one field element while the hasher
// returns Width/2 of them, so a caller sizing a buffer from hash.Hash.Size()
// under-allocated by a factor of 4 to 8.
func TestDigestSizeMatchesHasher(t *testing.T) {
	for h := hash.Hash(0); h.Available(); h++ {
		hsh := h.New()
		if got, want := h.Size(), hsh.Size(); got != want {
			t.Errorf("%s: Hash.Size() = %d, but New().Size() = %d", h, got, want)
		}
		if got, want := len(hsh.Sum(nil)), hsh.Size(); got != want {
			t.Errorf("%s: len(Sum(nil)) = %d, but Size() = %d", h, got, want)
		}
	}
}
