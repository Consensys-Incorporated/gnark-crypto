// Copyright 2020-2025 Consensys Software Inc.
// Licensed under the Apache License, Version 2.0. See the LICENSE file for details.

package merkletree

import (
	"bytes"
	"crypto/sha256"
	"math"
	"testing"
)

// subtreeRoot builds a standalone tree out of leaves and returns its root.
func subtreeRoot(t *testing.T, leaves ...[]byte) []byte {
	t.Helper()
	sub := New(sha256.New())
	for _, l := range leaves {
		sub.Push(l)
	}
	return sub.Root()
}

// TestPushSubTreeInvalidHeight checks that PushSubTree rejects heights that are
// not valid Merkle subtree heights. Such heights used to be accepted: shifting
// a uint64 by an amount >= its width yields 0 (and uint64(-1) is huge), so the
// subtree size collapsed to 0, a bogus subtree was pushed and currentIndex did
// not advance.
func TestPushSubTreeInvalidHeight(t *testing.T) {
	for _, height := range []int{-1, -2, -64, -1024, 64, 65, 128, math.MaxInt32} {
		for _, preFilled := range []bool{false, true} {
			tree := New(sha256.New())
			if preFilled {
				tree.Push([]byte("leaf"))
			}
			head, index := tree.head, tree.currentIndex

			if err := tree.PushSubTree(height, []byte("sum")); err == nil {
				t.Errorf("PushSubTree(height=%d, preFilled=%v) returned no error, want an error", height, preFilled)
			}

			// the tree must be left untouched
			if tree.head != head || tree.currentIndex != index {
				t.Errorf("PushSubTree(height=%d, preFilled=%v) modified the tree state", height, preFilled)
			}
		}
	}
}

// TestPushSubTreeIndexOverflow checks that the leaf count stays representable.
func TestPushSubTreeIndexOverflow(t *testing.T) {
	// One leaf away from the maximum: a subtree of height 0 still fits,
	// anything bigger does not.
	tree := New(sha256.New())
	tree.currentIndex = math.MaxUint64 - 1
	if err := tree.PushSubTree(1, []byte("sum")); err == nil {
		t.Error("PushSubTree(height=1) on an almost full tree returned no error, want an overflow error")
	}
	if tree.head != nil || tree.currentIndex != math.MaxUint64-1 {
		t.Error("PushSubTree modified the tree state on overflow")
	}
	if err := tree.PushSubTree(0, []byte("sum")); err != nil {
		t.Errorf("PushSubTree(height=0) should still fit: %v", err)
	}
	if tree.currentIndex != math.MaxUint64 {
		t.Errorf("currentIndex = %d, want %d", tree.currentIndex, uint64(math.MaxUint64))
	}

	// A tree that already holds MaxUint64 leaves cannot take any subtree.
	full := New(sha256.New())
	full.currentIndex = math.MaxUint64
	if err := full.PushSubTree(0, []byte("sum")); err == nil {
		t.Error("PushSubTree on a full tree returned no error, want an overflow error")
	}

	// Boundary: the largest representable subtree is accepted by an empty tree.
	empty := New(sha256.New())
	if err := empty.PushSubTree(maxSubTreeHeight, []byte("sum")); err != nil {
		t.Errorf("PushSubTree(height=%d) returned an error: %v", maxSubTreeHeight, err)
	}
	if want := uint64(1) << maxSubTreeHeight; empty.currentIndex != want {
		t.Errorf("currentIndex = %d, want %d", empty.currentIndex, want)
	}
}

// TestPushSubTreeLargerThanSmallest covers the pre-existing check that a
// subtree cannot be larger than the smallest subtree of the tree.
func TestPushSubTreeLargerThanSmallest(t *testing.T) {
	tree := New(sha256.New())
	tree.Push([]byte("leaf"))
	if err := tree.PushSubTree(1, []byte("sum")); err == nil {
		t.Error("PushSubTree(height=1) on a tree whose smallest subtree has height 0 returned no error")
	}
}

// TestPushSubTreeMatchesPush is a regression test: valid subtrees must still
// produce exactly the same tree as pushing the leaves one by one.
func TestPushSubTreeMatchesPush(t *testing.T) {
	leaves := make([][]byte, 8)
	for i := range leaves {
		leaves[i] = []byte{byte(i)}
	}

	want := New(sha256.New())
	for _, l := range leaves {
		want.Push(l)
	}

	// two subtrees of height 2
	got := New(sha256.New())
	if err := got.PushSubTree(2, subtreeRoot(t, leaves[0:4]...)); err != nil {
		t.Fatal(err)
	}
	if err := got.PushSubTree(2, subtreeRoot(t, leaves[4:8]...)); err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(got.Root(), want.Root()) {
		t.Errorf("2 x height 2: root = %x, want %x", got.Root(), want.Root())
	}

	// a subtree of height 0 is a single leaf
	got = New(sha256.New())
	for _, l := range leaves {
		if err := got.PushSubTree(0, subtreeRoot(t, l)); err != nil {
			t.Fatal(err)
		}
	}
	if !bytes.Equal(got.Root(), want.Root()) {
		t.Errorf("8 x height 0: root = %x, want %x", got.Root(), want.Root())
	}

	// mixed: subtrees of decreasing height, then plain leaves
	got = New(sha256.New())
	if err := got.PushSubTree(2, subtreeRoot(t, leaves[0:4]...)); err != nil {
		t.Fatal(err)
	}
	if err := got.PushSubTree(1, subtreeRoot(t, leaves[4:6]...)); err != nil {
		t.Fatal(err)
	}
	for _, l := range leaves[6:] {
		got.Push(l)
	}
	if !bytes.Equal(got.Root(), want.Root()) {
		t.Errorf("mixed: root = %x, want %x", got.Root(), want.Root())
	}
}
