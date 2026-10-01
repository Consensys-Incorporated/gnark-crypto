package config

import "testing"

// TestRadixDefaultsToWordAligned guards the invariant that makes the RBits axis
// safe to introduce: unless a field explicitly asks for a narrower Montgomery
// radix, R stays exactly 2^(NbWords*Word.BitSize), which is what every field
// used before RBits existed. If this ever drifts, every generated field changes
// silently; failing here costs milliseconds instead of a full regeneration.
func TestRadixDefaultsToWordAligned(t *testing.T) {
	t.Parallel()

	moduli := []struct{ name, modulus string }{
		{"goldilocks", "0xFFFFFFFF00000001"},
		{"koalabear", "0x7f000001"},
		{"babybear", "0x78000001"},
		{"bn254Fr", "21888242871839275222246405745257275088548364400416034343698204186575808495617"},
		{"bn254Fp", "21888242871839275222246405745257275088696311157297823662689037894645226208583"},
		{"bls12-381Fr", "52435875175126190479447740508185965837690552500527637822603658699938581184513"},
		{"bw6-761Fp", "6891450384315732539396789682275657542479668912536150109513790160209623422243491736087683183289411687640864567753786613451161759120554247759349511699125301598951605099378508850372543631423596795951899700429969112842764913119068299"},
		{"secp256k1Fp", "0xFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFEFFFFFC2F"},
	}

	for _, m := range moduli {
		t.Run(m.name, func(t *testing.T) {
			f, err := NewFieldConfig(m.name, "Element", m.modulus, false)
			if err != nil {
				t.Fatal(err)
			}
			want := uint(f.NbWords) * uint(f.Word.BitSize)
			if f.RBits != want {
				t.Errorf("RBits = %d, want %d (NbWords=%d, Word.BitSize=%d)", f.RBits, want, f.NbWords, f.Word.BitSize)
			}
			if f.RadixSubWord {
				t.Errorf("RadixSubWord = true, want false for a word-aligned field")
			}
		})
	}
}

// TestMontgomeryRadixOverride pins the derived constants for a sub-word radix
// against the hand-written mamabear implementation (p = 2^49 - 2^34 + 1,
// R = 2^52). These values are only reproduced when R is genuinely 2^52; with
// the default R = 2^64 both One and RSquare differ, which is the whole reason
// the RBits axis exists.
func TestMontgomeryRadixOverride(t *testing.T) {
	t.Parallel()

	const mamabear = "0x1FFFC00000001"

	f, err := NewFieldConfig("mamabear", "Element", mamabear, false, WithMontgomeryRadixBits(52))
	if err != nil {
		t.Fatal(err)
	}

	if f.NbWords != 1 || f.NbBits != 49 {
		t.Fatalf("NbWords=%d NbBits=%d, want 1 and 49", f.NbWords, f.NbBits)
	}
	if !f.RadixSubWord || f.RBits != 52 {
		t.Errorf("RadixSubWord=%v RBits=%d, want true and 52", f.RadixSubWord, f.RBits)
	}
	if f.RMask != (1<<52)-1 {
		t.Errorf("RMask = %d, want %d", f.RMask, uint64(1<<52)-1)
	}
	if f.F31 {
		t.Error("F31 = true, want false for a 49-bit field")
	}
	if !f.SingleWordSmall {
		t.Error("SingleWordSmall = false, want true (49 bits in a 64-bit word)")
	}
	if !f.SparsePrime || f.PHiBit != 49 || f.PLoBit != 34 {
		t.Errorf("SparsePrime=%v PHiBit=%d PLoBit=%d, want true, 49, 34", f.SparsePrime, f.PHiBit, f.PLoBit)
	}

	// Constants transcribed from the hand-written field/mamabear/element.go,
	// tagged mamabear-handwritten-v0.
	for _, tc := range []struct {
		name string
		got  []uint64
		want uint64
	}{
		{"QInverse", f.QInverse, 562932773552127},
		{"One", f.One, 137438953464},
		{"RSquare", f.RSquare, 15393129233472},
		{"QMinusOneHalvedP", f.QMinusOneHalvedP, 281466386776065},
	} {
		if len(tc.got) == 0 || tc.got[0] != tc.want {
			t.Errorf("%s = %v, want [%d]", tc.name, tc.got, tc.want)
		}
	}

	// 2-adicity 34 with odd part 32767; the Sarkar sqrt gate keys off this.
	if f.SqrtE != 34 {
		t.Errorf("SqrtE = %d, want 34", f.SqrtE)
	}
	if len(f.SqrtS) == 0 || f.SqrtS[0] != 32767 {
		t.Errorf("SqrtS = %v, want [32767]", f.SqrtS)
	}
	// With a 2-adicity of 34 the Tonelli-Shanks loop dominates, so this field
	// must take the Sarkar path, as the hand-written implementation does.
	if !f.SqrtSarkar {
		t.Error("SqrtSarkar = false, want true at 2-adicity 34")
	}

	// Existing single-word fields must NOT be dragged onto the Sarkar path by
	// the widened gate.
	for _, m := range []string{"0x7f000001", "0xFFFFFFFF00000001"} {
		other, err := NewFieldConfig("x", "Element", m, false)
		if err != nil {
			t.Fatal(err)
		}
		if other.SqrtSarkar {
			t.Errorf("modulus %s: SqrtSarkar = true, want false", m)
		}
	}
}

// TestMontgomeryRadixRejected checks that unsupported radix overrides fail loudly
// rather than silently producing a field with wrong constants.
func TestMontgomeryRadixRejected(t *testing.T) {
	t.Parallel()

	for _, tc := range []struct {
		name    string
		modulus string
		bits    uint
	}{
		{"multiWord", "21888242871839275222246405745257275088548364400416034343698204186575808495617", 52},
		{"radixWiderThanWord", "0x1FFFC00000001", 70},
		{"radixBelowModulus", "0x1FFFC00000001", 48},
	} {
		t.Run(tc.name, func(t *testing.T) {
			if _, err := NewFieldConfig("x", "Element", tc.modulus, false, WithMontgomeryRadixBits(tc.bits)); err == nil {
				t.Fatal("expected an error, got nil")
			}
		})
	}
}

// TestSparsePrimeExponents covers the detector used to enable the sparse-prime
// fast reduction, including moduli that must NOT be classified as sparse.
func TestSparsePrimeExponents(t *testing.T) {
	t.Parallel()

	for _, tc := range []struct {
		name    string
		modulus string
		sparse  bool
		hi, lo  uint
	}{
		{"mamabear", "0x1FFFC00000001", true, 49, 34},
		{"koalabear", "0x7f000001", true, 31, 24},
		{"babybear", "0x78000001", true, 31, 27},
		{"goldilocks", "0xFFFFFFFF00000001", true, 64, 32},
		{"bn254Fr", "21888242871839275222246405745257275088548364400416034343698204186575808495617", false, 0, 0},
	} {
		t.Run(tc.name, func(t *testing.T) {
			f, err := NewFieldConfig(tc.name, "Element", tc.modulus, false)
			if err != nil {
				t.Fatal(err)
			}
			if f.SparsePrime != tc.sparse {
				t.Fatalf("SparsePrime = %v, want %v", f.SparsePrime, tc.sparse)
			}
			if tc.sparse && (f.PHiBit != tc.hi || f.PLoBit != tc.lo) {
				t.Errorf("got 2^%d - 2^%d + 1, want 2^%d - 2^%d + 1", f.PHiBit, f.PLoBit, tc.hi, tc.lo)
			}
		})
	}
}
