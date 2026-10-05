package fptower

import "testing"

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

	_, wantErr := want.A0.SetString("nope")
	if wantErr == nil {
		t.Fatal("subfield accepted an invalid coefficient")
	}
	for _, coeffs := range [][2]string{{"nope", "1"}, {"1", "nope"}} {
		got, err = z.SetString(coeffs[0], coeffs[1])
		if got != nil {
			t.Fatalf("SetString(%q, %q) returned a receiver", coeffs[0], coeffs[1])
		}
		if err == nil || err.Error() != wantErr.Error() {
			t.Fatalf("SetString(%q, %q) error = %v, want %v", coeffs[0], coeffs[1], err, wantErr)
		}
	}
}
