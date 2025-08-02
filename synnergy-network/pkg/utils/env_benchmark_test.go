package utils

import (
	"os"
	"testing"
)

func BenchmarkEnvOrDefault(b *testing.B) {
	if err := os.Unsetenv("BENCH_KEY"); err != nil {
		b.Fatalf("Unsetenv: %v", err)
	}
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		EnvOrDefault("BENCH_KEY", "fallback")
}

func BenchmarkEnvOrDefaultInt(b *testing.B) {
	if err := os.Setenv("BENCH_INT", "123"); err != nil {
		b.Fatalf("Setenv: %v", err)
	}
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		EnvOrDefaultInt("BENCH_INT", 0)
	}
}

func BenchmarkEnvOrDefaultUint64(b *testing.B) {
	if err := os.Setenv("BENCH_UINT", "123"); err != nil {
		b.Fatalf("Setenv: %v", err)
	}
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		EnvOrDefaultUint64("BENCH_UINT", 0)
	}
}
