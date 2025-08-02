package utils

import (
	"os"
	"testing"
)

func BenchmarkEnvOrDefault(b *testing.B) {
	const key = "BENCH_KEY"
	os.Setenv(key, "value")
	clearEnvCache(key)
	// warm cache
	EnvOrDefault(key, "fallback")
	b.ResetTimer()
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		EnvOrDefault(key, "fallback")
	}
}

func BenchmarkEnvOrDefaultInt(b *testing.B) {
	const key = "BENCH_INT"
	os.Setenv(key, "123")
	clearEnvCache(key)
	EnvOrDefaultInt(key, 0)
	b.ResetTimer()
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		EnvOrDefaultInt(key, 0)
	}
}

func BenchmarkEnvOrDefaultUint64(b *testing.B) {
	const key = "BENCH_UINT"
	os.Setenv(key, "123")
	clearEnvCache(key)
	EnvOrDefaultUint64(key, 0)
	b.ResetTimer()
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		EnvOrDefaultUint64(key, 0)
	}
}
