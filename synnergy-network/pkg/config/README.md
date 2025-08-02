# config

Reusable configuration loader for Synnergy applications.

## Versioning

Current version: v0.1.0

## APIs

- `Load(env string) (*Config, error)`: load configuration for the given environment name.
- `LoadFromEnv() (*Config, error)`: load configuration using the `SYNN_ENV` environment variable.

The resulting configuration is stored in the package variable `AppConfig`.
