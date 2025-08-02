package utils

import (
    "os"
    "testing"
)

func BenchmarkEnvOrDefault(b *testing.B) {
    os.Unsetenv("BENCH_KEY")
    b.ReportAllocs()
    for i := 0; i < b.N; i++ {
        EnvOrDefault("BENCH_KEY", "fallback")
    }
}

func BenchmarkEnvOrDefaultInt(b *testing.B) {
    os.Setenv("BENCH_INT", "123")
    b.ReportAllocs()
    for i := 0; i < b.N; i++ {
        EnvOrDefaultInt("BENCH_INT", 0)
    }
}

func BenchmarkEnvOrDefaultUint64(b *testing.B) {
    os.Setenv("BENCH_UINT", "123")
    b.ReportAllocs()
    for i := 0; i < b.N; i++ {
        EnvOrDefaultUint64("BENCH_UINT", 0)
    }
}

