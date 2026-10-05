package fptower

import "testing"

func TestE6DSetString(t *testing.T) {
	t.Parallel()

	coeffs := [6]string{"1", "2", "3", "4", "5", "0x6"}
	var want E6D
	if _, err := want.A0.SetString(coeffs[0]); err != nil {
		t.Fatal(err)
	}
	if _, err := want.A1.SetString(coeffs[1]); err != nil {
		t.Fatal(err)
	}
	if _, err := want.A2.SetString(coeffs[2]); err != nil {
		t.Fatal(err)
	}
	if _, err := want.A3.SetString(coeffs[3]); err != nil {
		t.Fatal(err)
	}
	if _, err := want.A4.SetString(coeffs[4]); err != nil {
		t.Fatal(err)
	}
	if _, err := want.A5.SetString(coeffs[5]); err != nil {
		t.Fatal(err)
	}

	var z E6D
	got, err := z.SetString(coeffs[0], coeffs[1], coeffs[2], coeffs[3], coeffs[4], coeffs[5])
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
	for _, badAt := range []int{0, 3, 5} {
		bad := coeffs
		bad[badAt] = "nope"
		got, err = z.SetString(bad[0], bad[1], bad[2], bad[3], bad[4], bad[5])
		if got != nil {
			t.Fatalf("SetString returned a receiver with invalid coefficient %d", badAt)
		}
		if err == nil || err.Error() != wantErr.Error() {
			t.Fatalf("coefficient %d error = %v, want %v", badAt, err, wantErr)
		}
	}
}
