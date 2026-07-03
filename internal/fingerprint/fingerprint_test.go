package fingerprint

import "testing"

func TestSHA256IsStable(t *testing.T) {
	a := SHA256("select * from t")
	b := SHA256("select * from t")
	if a != b {
		t.Fatalf("fingerprint not stable: %q vs %q", a, b)
	}
	if len(a) != len("sha256:")+64 {
		t.Fatalf("unexpected fingerprint length: %q", a)
	}
}
