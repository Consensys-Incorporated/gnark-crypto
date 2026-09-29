//go:build !purego

// Copyright 2020-2026 Consensys Software Inc.
// Licensed under the Apache License, Version 2.0. See the LICENSE file for details.

// Asm-kernel vs generic-path equivalence tests: edge values in every ZMM
// lane, ragged tails, aliasing, and the sumVec/innerProdVec chunk boundary.
// Requires an AVX-512IFMA host to exercise the asm path (build-gated !purego
// because it calls the kernels directly).

package mamabear

import (
	"math/rand/v2"
	"testing"

	"github.com/consensys/gnark-crypto/utils/cpu"
)

func ifmaEdgeElements() []Element {
	var out []Element
	for _, v := range []uint64{
		0, 1, 2, 3,
		q - 1, q - 2, q - 3,
		q / 2, q/2 + 1,
		1 << 48, (1 << 48) + 1,
		1 << 34, (1 << 34) - 1,
		0x1fffbffffffff, // q - 2^34
		12345, 6789,
	} {
		var e Element
		e[0] = v
		out = append(out, e)
	}
	return out
}

func randCanonical(rng *rand.Rand) Element {
	var e Element
	e[0] = rng.Uint64() % q
	return e
}

func randVector(rng *rand.Rand, n int) Vector {
	v := make(Vector, n)
	for i := range v {
		v[i] = randCanonical(rng)
	}
	return v
}

// TestReviewVectorKernelsVsGeneric feeds edge-case and random canonical inputs
// through every dispatched vector op and compares against the generic path.
// Covers tails (n%8 != 0), self-aliasing, and all-pairs edge combinations.
func TestReviewVectorKernelsVsGeneric(t *testing.T) {
	if !cpu.SupportAVX512IFMA {
		t.Skip("AVX-512IFMA absent: asm kernels not exercised")
	}
	rng := rand.New(rand.NewPCG(0xdeadbeef, 0xcafef00d))
	edges := ifmaEdgeElements()

	sizes := []int{1, 2, 7, 8, 9, 15, 16, 17, 31, 64, 65, 100, 255, 256, 1000}

	check := func(name string, got, want Vector) {
		t.Helper()
		if len(got) != len(want) {
			t.Fatalf("%s: len %d != %d", name, len(got), len(want))
		}
		for i := range got {
			if got[i] != want[i] {
				t.Fatalf("%s: index %d: got %d, want %d", name, i, got[i][0], want[i][0])
			}
		}
	}

	// All-pairs edge coverage at exactly one block (8 lanes) so every lane
	// sees every edge value.
	a8 := make(Vector, 8)
	b8 := make(Vector, 8)
	copy(a8, edges[:8])
	copy(b8, edges[8:16])
	// rotate edges through lanes
	for rot := 0; rot < len(edges)-8; rot++ {
		copy(a8, edges[rot:rot+8])
		for _, op := range []string{"add", "sub", "mul"} {
			gotAsm := make(Vector, 8)
			gotGen := make(Vector, 8)
			switch op {
			case "add":
				addVec(&gotAsm[0], &a8[0], &b8[0], 1)
				addVecGeneric(gotGen, a8, b8)
			case "sub":
				subVec(&gotAsm[0], &a8[0], &b8[0], 1)
				subVecGeneric(gotGen, a8, b8)
			case "mul":
				mulVec(&gotAsm[0], &a8[0], &b8[0], 1)
				mulVecGeneric(gotGen, a8, b8)
			}
			check(op+"-edges", gotAsm, gotGen)
		}
		var bScalar Element = b8[0]
		gotAsm := make(Vector, 8)
		gotGen := make(Vector, 8)
		scalarMulVec(&gotAsm[0], &a8[0], &bScalar, 1)
		scalarMulVecGeneric(gotGen, a8, &bScalar)
		check("scalarmul-edges", gotAsm, gotGen)
	}

	// Random sizes incl. tails, all dispatched ops vs generic.
	for _, n := range sizes {
		a := randVector(rng, n)
		b := randVector(rng, n)
		var scalar Element
		scalar = randCanonical(rng)

		got := make(Vector, n)
		want := make(Vector, n)

		got.Add(a, b)
		addVecGeneric(want, a, b)
		check("Add", got, want)

		got.Sub(a, b)
		subVecGeneric(want, a, b)
		check("Sub", got, want)

		got.Mul(a, b)
		mulVecGeneric(want, a, b)
		check("Mul", got, want)

		got.ScalarMul(a, &scalar)
		scalarMulVecGeneric(want, a, &scalar)
		check("ScalarMul", got, want)

		// Self-aliasing: res == a, res == b.
		alias := randVector(rng, n)
		backup := make(Vector, n)
		copy(backup, alias)
		alias.Add(alias, b)
		addVecGeneric(backup, backup, b)
		check("Add-alias", alias, backup)

		copy(backup, alias)
		alias.Mul(alias, b)
		mulVecGeneric(backup, backup, b)
		check("Mul-alias", alias, backup)

		// Sum / InnerProduct oracles.
		var sumGot, sumWant Element
		sumGot = a.Sum()
		sumVecGeneric(&sumWant, a)
		if sumGot != sumWant {
			t.Fatalf("Sum n=%d: got %d want %d", n, sumGot[0], sumWant[0])
		}
		var ipGot, ipWant Element
		ipGot = a.InnerProduct(b)
		innerProductVecGeneric(&ipWant, a, b)
		if ipGot != ipWant {
			t.Fatalf("InnerProduct n=%d: got %d want %d", n, ipGot[0], ipWant[0])
		}
	}
}

// TestReviewSumChunkBoundary crosses the maxSumBlocks chunk boundary
// (2^13 blocks = 65536 elements) for Sum and InnerProduct.
func TestReviewSumChunkBoundary(t *testing.T) {
	if !cpu.SupportAVX512IFMA {
		t.Skip("AVX-512IFMA absent: asm kernels not exercised")
	}
	rng := rand.New(rand.NewPCG(1, 2))
	base := maxSumBlocks * blockSize // 65536
	for _, n := range []int{base - 1, base, base + 1, 2*base + 7, 2*base + 8, 3*base + 5} {
		a := randVector(rng, n)
		b := randVector(rng, n)
		var sumGot, sumWant Element
		sumGot = a.Sum()
		sumVecGeneric(&sumWant, a)
		if sumGot != sumWant {
			t.Fatalf("Sum n=%d: got %d want %d", n, sumGot[0], sumWant[0])
		}
		var ipGot, ipWant Element
		ipGot = a.InnerProduct(b)
		innerProductVecGeneric(&ipWant, a, b)
		if ipGot != ipWant {
			t.Fatalf("InnerProduct n=%d: got %d want %d", n, ipGot[0], ipWant[0])
		}
	}
}

// FuzzReviewVectorOps fuzzes the dispatched vector ops against the generic
// path with random lengths and contents.
func FuzzReviewVectorOps(f *testing.F) {
	f.Add(uint64(1000), uint64(37))
	f.Add(uint64(0), uint64(9))
	f.Add(uint64(q-1), uint64(17))
	f.Fuzz(func(t *testing.T, seed, nRaw uint64) {
		if !cpu.SupportAVX512IFMA {
			t.Skip("AVX-512IFMA absent")
		}
		n := int(nRaw % 3000)
		rng := rand.New(rand.NewPCG(seed, seed^0x9e3779b9))
		a := randVector(rng, n)
		b := randVector(rng, n)
		got := make(Vector, n)
		want := make(Vector, n)
		got.Mul(a, b)
		mulVecGeneric(want, a, b)
		for i := range got {
			if got[i] != want[i] {
				t.Fatalf("Mul: n=%d index %d: got %d want %d", n, i, got[i][0], want[i][0])
			}
		}
		got.Add(a, b)
		addVecGeneric(want, a, b)
		for i := range got {
			if got[i] != want[i] {
				t.Fatalf("Add: n=%d index %d: got %d want %d", n, i, got[i][0], want[i][0])
			}
		}
		got.Sub(a, b)
		subVecGeneric(want, a, b)
		for i := range got {
			if got[i] != want[i] {
				t.Fatalf("Sub: n=%d index %d: got %d want %d", n, i, got[i][0], want[i][0])
			}
		}
		scalar := randCanonical(rng)
		got.ScalarMul(a, &scalar)
		scalarMulVecGeneric(want, a, &scalar)
		for i := range got {
			if got[i] != want[i] {
				t.Fatalf("ScalarMul: n=%d index %d: got %d want %d", n, i, got[i][0], want[i][0])
			}
		}
		var sumGot, sumWant Element
		sumGot = a.Sum()
		sumVecGeneric(&sumWant, a)
		if sumGot != sumWant {
			t.Fatalf("Sum: n=%d got %d want %d", n, sumGot[0], sumWant[0])
		}
		var ipGot, ipWant Element
		ipGot = a.InnerProduct(b)
		innerProductVecGeneric(&ipWant, a, b)
		if ipGot != ipWant {
			t.Fatalf("InnerProduct: n=%d got %d want %d", n, ipGot[0], ipWant[0])
		}
	})
}
