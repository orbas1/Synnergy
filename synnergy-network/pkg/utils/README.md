# utils

Shared utility helpers for the Synnergy project.

## Versioning

Current version: v0.1.0

## APIs

- `Wrap(err error, message string) error`: attach context to errors.
- `EnvOrDefault(key, fallback string) string`: get an environment variable or a default.
- `EnvOrDefaultInt(key string, fallback int) int`: parse an integer environment variable with a default.
- `EnvOrDefaultUint64(key string, fallback uint64) uint64`: parse a `uint64` environment variable with a default.
