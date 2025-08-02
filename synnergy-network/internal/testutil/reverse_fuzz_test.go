package testutil

import "testing"

// FuzzReverse ensures Reverse is its own inverse via randomized inputs.
func FuzzReverse(f *testing.F) {
    seeds := []string{"", "foo", "bar", "世界"}
    for _, s := range seeds {
        f.Add(s)
    }
    f.Fuzz(func(t *testing.T, s string) {
        if got := Reverse(Reverse(s)); got != s {
            t.Fatalf("reverse twice mismatch: got %q want %q", got, s)
        }
    })
}

