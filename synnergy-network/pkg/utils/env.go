package utils

import (
	"os"
	"strconv"
)

// EnvOrDefault returns the value of the environment variable identified by key
// or the provided fallback if the variable is unset or empty.
func EnvOrDefault(key, fallback string) string {
	if v, ok := os.LookupEnv(key); ok && v != "" {
		return v
	}
	return fallback
}

// EnvOrDefaultInt returns the integer value of the environment variable
// identified by key or the provided fallback if the variable is unset,
// empty, or cannot be parsed as an integer.
func EnvOrDefaultInt(key string, fallback int) int {
	if v, ok := os.LookupEnv(key); ok && v != "" {
		if n, err := strconv.Atoi(v); err == nil {
			return n
		}
	}
	return fallback
}

// EnvOrDefaultUint64 returns the uint64 value of the environment variable
// identified by key or the provided fallback if the variable is unset,
// empty, or cannot be parsed as a uint64.
func EnvOrDefaultUint64(key string, fallback uint64) uint64 {
	if v, ok := os.LookupEnv(key); ok && v != "" {
		if n, err := strconv.ParseUint(v, 10, 64); err == nil {
			return n
		}
	}
	return fallback
}
