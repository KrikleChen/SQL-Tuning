package normalize

import "testing"

func TestNormalizeCollapsesWhitespace(t *testing.T) {
	got := SQL(" select  *\nfrom   t  ")
	want := "select * from t"
	if got != want {
		t.Fatalf("got %q, want %q", got, want)
	}
}
