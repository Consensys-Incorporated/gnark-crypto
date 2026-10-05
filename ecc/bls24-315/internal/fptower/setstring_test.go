package fptower

import (
	"testing"

	"github.com/consensys/gnark-crypto/ecc/bls24-315/fp"
)

func TestE2SetString(t *testing.T) {
	t.Parallel()

	var want E2
	if _, err := want.A0.SetString("0x10"); err != nil {
		t.Fatal(err)
	}
	if _, err := want.A1.SetString("2"); err != nil {
		t.Fatal(err)
	}

	var z E2
	got, err := z.SetString("0x10", "2")
	if err != nil {
		t.Fatalf("valid coefficients: %v", err)
	}
	if got != &z {
		t.Fatal("SetString did not return the receiver")
	}
	if !z.Equal(&want) {
		t.Fatalf("got %s, want %s", z.String(), want.String())
	}

	wantErr := subfieldParseError(t)
	for _, coeffs := range [][2]string{{"nope", "1"}, {"1", "nope"}} {
		got, err = z.SetString(coeffs[0], coeffs[1])
		requireParseError(t, got, err, wantErr)
	}
}

func TestE4SetString(t *testing.T) {
	t.Parallel()

	coeffs := [4]string{"1", "2", "3", "0x4"}
	var want E4
	leaves := [4]*fp.Element{&want.B0.A0, &want.B0.A1, &want.B1.A0, &want.B1.A1}
	for i, s := range coeffs {
		if _, err := leaves[i].SetString(s); err != nil {
			t.Fatal(err)
		}
	}

	var z E4
	got, err := z.SetString(coeffs[0], coeffs[1], coeffs[2], coeffs[3])
	if err != nil {
		t.Fatalf("valid coefficients: %v", err)
	}
	if got != &z {
		t.Fatal("SetString did not return the receiver")
	}
	if !z.Equal(&want) {
		t.Fatalf("got %s, want %s", z.String(), want.String())
	}

	wantErr := subfieldParseError(t)
	for _, badAt := range []int{0, 3} {
		bad := coeffs
		bad[badAt] = "nope"
		got, err = z.SetString(bad[0], bad[1], bad[2], bad[3])
		requireParseError(t, got, err, wantErr)
	}
}

func TestE12SetString(t *testing.T) {
	t.Parallel()

	var coeffs [12]string
	for i := range coeffs {
		coeffs[i] = "1"
	}
	coeffs[11] = "0x2"

	var want E12
	leaves := e12Leaves(&want)
	for i, s := range coeffs {
		if _, err := leaves[i].SetString(s); err != nil {
			t.Fatal(err)
		}
	}

	var z E12
	got, err := z.SetString(
		coeffs[0], coeffs[1], coeffs[2], coeffs[3],
		coeffs[4], coeffs[5], coeffs[6], coeffs[7],
		coeffs[8], coeffs[9], coeffs[10], coeffs[11],
	)
	if err != nil {
		t.Fatalf("valid coefficients: %v", err)
	}
	if got != &z {
		t.Fatal("SetString did not return the receiver")
	}
	if !z.Equal(&want) {
		t.Fatalf("got %s, want %s", z.String(), want.String())
	}

	wantErr := subfieldParseError(t)
	for _, badAt := range []int{0, 5, 11} {
		bad := coeffs
		bad[badAt] = "nope"
		got, err = z.SetString(
			bad[0], bad[1], bad[2], bad[3],
			bad[4], bad[5], bad[6], bad[7],
			bad[8], bad[9], bad[10], bad[11],
		)
		requireParseError(t, got, err, wantErr)
	}
}

func TestE24SetString(t *testing.T) {
	t.Parallel()

	var coeffs [24]string
	for i := range coeffs {
		coeffs[i] = "1"
	}
	coeffs[23] = "0x2"

	var want E24
	leaves := e24Leaves(&want)
	for i, s := range coeffs {
		if _, err := leaves[i].SetString(s); err != nil {
			t.Fatal(err)
		}
	}

	var z E24
	got, err := z.SetString(
		coeffs[0], coeffs[1], coeffs[2], coeffs[3], coeffs[4], coeffs[5],
		coeffs[6], coeffs[7], coeffs[8], coeffs[9], coeffs[10], coeffs[11],
		coeffs[12], coeffs[13], coeffs[14], coeffs[15], coeffs[16], coeffs[17],
		coeffs[18], coeffs[19], coeffs[20], coeffs[21], coeffs[22], coeffs[23],
	)
	if err != nil {
		t.Fatalf("valid coefficients: %v", err)
	}
	if got != &z {
		t.Fatal("SetString did not return the receiver")
	}
	if !z.Equal(&want) {
		t.Fatalf("got %s, want %s", z.String(), want.String())
	}

	wantErr := subfieldParseError(t)
	for _, badAt := range []int{0, 12, 23} {
		bad := coeffs
		bad[badAt] = "nope"
		got, err = z.SetString(
			bad[0], bad[1], bad[2], bad[3], bad[4], bad[5],
			bad[6], bad[7], bad[8], bad[9], bad[10], bad[11],
			bad[12], bad[13], bad[14], bad[15], bad[16], bad[17],
			bad[18], bad[19], bad[20], bad[21], bad[22], bad[23],
		)
		requireParseError(t, got, err, wantErr)
	}
}

func e12Leaves(z *E12) [12]*fp.Element {
	return [12]*fp.Element{
		&z.C0.B0.A0, &z.C0.B0.A1, &z.C0.B1.A0, &z.C0.B1.A1,
		&z.C1.B0.A0, &z.C1.B0.A1, &z.C1.B1.A0, &z.C1.B1.A1,
		&z.C2.B0.A0, &z.C2.B0.A1, &z.C2.B1.A0, &z.C2.B1.A1,
	}
}

func e24Leaves(z *E24) [24]*fp.Element {
	d0 := e12Leaves(&z.D0)
	d1 := e12Leaves(&z.D1)
	var leaves [24]*fp.Element
	copy(leaves[:12], d0[:])
	copy(leaves[12:], d1[:])
	return leaves
}

func subfieldParseError(t *testing.T) error {
	t.Helper()
	var coeff fp.Element
	_, err := coeff.SetString("nope")
	if err == nil {
		t.Fatal("subfield accepted an invalid coefficient")
	}
	return err
}

func requireParseError[T any](t *testing.T, got *T, err, want error) {
	t.Helper()
	if got != nil {
		t.Fatal("SetString returned a receiver for an invalid coefficient")
	}
	if err == nil || err.Error() != want.Error() {
		t.Fatalf("error = %v, want %v", err, want)
	}
}
