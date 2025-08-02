package testutil

import (
	"bytes"
	"os"
	"testing"
)

func TestSandboxReadWrite(t *testing.T) {
	sb, err := NewSandbox()
	if err != nil {
		t.Fatalf("NewSandbox failed: %v", err)
	}
	defer sb.Cleanup()

	data := []byte("hello world")
	if err := sb.WriteFile("file.txt", data, 0600); err != nil {
		t.Fatalf("WriteFile failed: %v", err)
	}
	got, err := sb.ReadFile("file.txt")
	if err != nil {
		t.Fatalf("ReadFile failed: %v", err)
	}
	if !bytes.Equal(got, data) {
		t.Fatalf("data mismatch: got %q want %q", got, data)
	}
}

func TestSandboxCleanup(t *testing.T) {
	sb, err := NewSandbox()
	if err != nil {
		t.Fatalf("NewSandbox failed: %v", err)
	}
	path := sb.Path("temp")
	if err := sb.WriteFile("temp", []byte("x"), 0600); err != nil {
		t.Fatalf("WriteFile failed: %v", err)
	}
	if err := sb.Cleanup(); err != nil {
		t.Fatalf("Cleanup failed: %v", err)
	}
	if _, err := os.Stat(path); !os.IsNotExist(err) {
		t.Fatalf("expected sandbox to be removed")
	}
}

func TestReverseSandboxIntegration(t *testing.T) {
	sb, err := NewSandbox()
	if err != nil {
		t.Fatalf("NewSandbox failed: %v", err)
	}
	defer sb.Cleanup()

	original := "integration"
	reversed := Reverse(original)
	if err := sb.WriteFile("data", []byte(reversed), 0600); err != nil {
		t.Fatalf("WriteFile failed: %v", err)
	}
	data, err := sb.ReadFile("data")
	if err != nil {
		t.Fatalf("ReadFile failed: %v", err)
	}
	if got := Reverse(string(data)); got != original {
		t.Fatalf("reverse integration mismatch: got %q want %q", got, original)
	}
}
