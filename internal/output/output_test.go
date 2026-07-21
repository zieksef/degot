package output

import (
	"testing"
)

func TestSprintf(t *testing.T) {
	got := Sprintf("created %s\n", "orderservice")

	want := "[degot]: created orderservice\n"
	if got != want {
		t.Errorf("Sprintf = %q, want %q", got, want)
	}
}
