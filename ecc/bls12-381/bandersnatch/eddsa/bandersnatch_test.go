package eddsa

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"testing"

	"github.com/consensys/gnark-crypto/ecc/bls12-381/bandersnatch"
)

// TestBandersnatchKAT pins the package to Bandersnatch: the key derived from a
// fixed seed is a point of Bandersnatch and the deterministic signature of a
// fixed message has a fixed value. Hand-written (not generated), kept by the
// generator. Before the fix of the eddsa generator this package imported the
// Jubjub curve: for this seed it produced the Jubjub public key
// daac52a82846152c171193a2117ae80036dd81b666a986ff8515463ae385ba9d, which
// is not on Bandersnatch.
func TestBandersnatchKAT(t *testing.T) {
	seed, _ := hex.DecodeString("0102030405060708090a0b0c0d0e0f101112131415161718191a1b1c1d1e1f20")
	msg, _ := hex.DecodeString("000102030405060708090a0b0c0d0e0f101112131415161718191a1b1c1d1e1f")
	const (
		wantPub = "fd528572df93716b013e6b9861310b387d57f95e686b61e613c4cd7837fa2a81"
		wantAX  = "71b5bd4031b1de4c412a22a5a8799faf3e9ac873798b149e1da6917b09565aaf"
		wantAY  = "012afa3778cdc413e6616b685ef9577d380b3161986b3e016b7193df728552fd"
		wantSig = "265c48151514c0e041a975523c0e43b6972b83f180712b70cce48213abac9fa2052ed9ef9b2504e0d93a4e5626e0c0f0b81ecffb065a184211baef45421c743b"
	)

	priv, err := GenerateKey(bytes.NewReader(seed))
	if err != nil {
		t.Fatal(err)
	}

	// The public key must satisfy the Bandersnatch equation (a = -5), whatever
	// the type of PublicKey.A: the coordinates are copied into a Bandersnatch point.
	var p bandersnatch.PointAffine
	p.X, p.Y = priv.PublicKey.A.X, priv.PublicKey.A.Y
	if !p.IsOnCurve() {
		t.Fatalf("public key %x is not a point of Bandersnatch", priv.PublicKey.Bytes())
	}
	if got := hex.EncodeToString(priv.PublicKey.Bytes()); got != wantPub {
		t.Fatalf("public key %s, want %s", got, wantPub)
	}
	ax, ay := priv.PublicKey.A.X.Bytes(), priv.PublicKey.A.Y.Bytes()
	if hex.EncodeToString(ax[:]) != wantAX || hex.EncodeToString(ay[:]) != wantAY {
		t.Fatalf("public key coordinates (%x, %x), want (%s, %s)", ax, ay, wantAX, wantAY)
	}

	sig, err := priv.Sign(msg, sha256.New())
	if err != nil {
		t.Fatal(err)
	}
	if got := hex.EncodeToString(sig); got != wantSig {
		t.Fatalf("signature %s, want %s", got, wantSig)
	}
	ok, err := priv.PublicKey.Verify(sig, msg, sha256.New())
	if err != nil || !ok {
		t.Fatalf("verification: %v %v", ok, err)
	}
}
