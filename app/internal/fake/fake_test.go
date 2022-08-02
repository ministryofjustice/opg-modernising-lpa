package fake

import "testing"

func TestGoodBye(t *testing.T) {
	got := GoodBye()
	want := "See ya!"

	if got != want {
		t.Errorf("got %q want %q", got, want)
	}
}
