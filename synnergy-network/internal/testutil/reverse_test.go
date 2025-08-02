package testutil

import "testing"

// TestReverse ensures that reversing a string twice yields the original value.
func TestReverse(t *testing.T) {
	cases := []string{"", "a", "ab", "Hello, 世界"}
	for _, c := range cases {
		if got := Reverse(Reverse(c)); got != c {
			t.Fatalf("reverse twice mismatch: got %q want %q", got, c)
		}
	}
}
