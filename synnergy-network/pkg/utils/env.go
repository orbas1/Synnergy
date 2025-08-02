package utils

import (
	"os"
	"strconv"
	"sync"
)

// envCache stores previously fetched non-empty environment variable values so
// repeat lookups avoid the relatively expensive syscall interaction.
var envCache sync.Map // map[string]string

// getEnv retrieves the value for key from the cache or the environment.
// Only non-empty values are cached.
func getEnv(key string) (string, bool) {
	if v, ok := envCache.Load(key); ok {
		return v.(string), true
	}
	if v := os.Getenv(key); v != "" {
		envCache.Store(key, v)
		return v, true
	}
	return "", false
}

// clearEnvCache removes any cached value for key. It is primarily used in
// tests where environment variables are modified between calls.
func clearEnvCache(key string) {
	envCache.Delete(key)
}

// EnvOrDefault returns the value of the environment variable identified by key
// or the provided fallback if the variable is unset or empty.
func EnvOrDefault(key, fallback string) string {
	if value, ok := os.LookupEnv(key); ok && value != "" {
		return value
	}
	return fallback
}

// EnvOrDefaultInt returns the integer value of the environment variable
// identified by key or the provided fallback if the variable is unset,
// empty, or cannot be parsed as an integer.
func EnvOrDefaultInt(key string, fallback int) int {
	if value, ok := os.LookupEnv(key); ok && value != "" {
		if num, err := strconv.Atoi(value); err == nil {
			return num
		}
	}
	return fallback
}

// EnvOrDefaultUint64 returns the uint64 value of the environment variable
// identified by key or the provided fallback if the variable is unset,
// empty, or cannot be parsed as a uint64.
func EnvOrDefaultUint64(key string, fallback uint64) uint64 {
	if value, ok := os.LookupEnv(key); ok && value != "" {
		if num, err := strconv.ParseUint(value, 10, 64); err == nil {
			return num
		}
	}
	return fallback
}
