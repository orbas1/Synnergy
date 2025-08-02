package utils

import (
	"os"
	"testing"
)

func TestEnvOrDefault(t *testing.T) {
	const key = "UTIL_TEST_STRING"
	_ = os.Unsetenv(key)
	if got := EnvOrDefault(key, "fallback"); got != "fallback" {
		t.Fatalf("expected fallback, got %q", got)
	}
	_ = os.Setenv(key, "value")
	if got := EnvOrDefault(key, "fallback"); got != "value" {
		t.Fatalf("expected value, got %q", got)
	}
}

func TestEnvOrDefaultInt(t *testing.T) {
	const key = "UTIL_TEST_INT"
	_ = os.Unsetenv(key)
	if got := EnvOrDefaultInt(key, 10); got != 10 {
		t.Fatalf("expected 10, got %d", got)
	}
	_ = os.Setenv(key, "5")
	if got := EnvOrDefaultInt(key, 10); got != 5 {
		t.Fatalf("expected 5, got %d", got)
	}
	_ = os.Setenv(key, "bad")
	if got := EnvOrDefaultInt(key, 7); got != 7 {
		t.Fatalf("expected fallback on parse error, got %d", got)
	}
}

func TestEnvOrDefaultUint64(t *testing.T) {
	const key = "UTIL_TEST_UINT64"
	_ = os.Unsetenv(key)
	if got := EnvOrDefaultUint64(key, 99); got != 99 {
		t.Fatalf("expected 99, got %d", got)
	}
	_ = os.Setenv(key, "42")
	if got := EnvOrDefaultUint64(key, 99); got != 42 {
		t.Fatalf("expected 42, got %d", got)
	}
	_ = os.Setenv(key, "bad")
	if got := EnvOrDefaultUint64(key, 77); got != 77 {
		t.Fatalf("expected fallback on parse error, got %d", got)
	}
}
