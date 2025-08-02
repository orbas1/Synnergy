package testutil

import "testing"

func FuzzSandboxReadWrite(f *testing.F) {
	f.Add([]byte("seed"))
	f.Fuzz(func(t *testing.T, data []byte) {
		sb, err := NewSandbox()
		if err != nil {
			t.Fatalf("NewSandbox failed: %v", err)
		}
		defer sb.Cleanup()
		if err := sb.WriteFile("fuzz", data, 0600); err != nil {
			t.Fatalf("WriteFile failed: %v", err)
		}
		out, err := sb.ReadFile("fuzz")
		if err != nil {
			t.Fatalf("ReadFile failed: %v", err)
		}
		if string(out) != string(data) {
			t.Fatalf("mismatch: got %q want %q", out, data)
		}
	})
}
