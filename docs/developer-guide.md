# Synnergy Developer Guide

This guide provides developers with the information needed to build, test and extend the Synnergy Network.

## Prerequisites

- Go 1.20 or newer
- Make and Bash on Unix-like systems
- Optional: Docker for containerized builds

## Repository Layout

The repository is organized around the `synnergy-network` directory which contains Go sources, GUIs and smart contracts. Key paths include:

- `cmd/` – command line applications
- `core/` – consensus, ledger and networking modules
- `GUI/` – web interfaces and dashboards
- `tests/` – unit tests for core packages

Additional tooling lives at the repository root:

- `setup_synn.sh` – minimal bootstrap script
- `Synnergy.env.sh` – full development environment setup
- `scripts/` – helpers for devnet and testnet setups

## Environment Setup

Clone the repository and run the bootstrap script:

```bash
./setup_synn.sh
```

For an expanded environment with additional tooling use:

```bash
./Synnergy.env.sh
```

Both scripts build the CLI binary in `synnergy-network`.

## Building Manually

```bash
cd synnergy-network
go mod tidy
GOFLAGS="-trimpath" go build -o synnergy ./cmd/synnergy
```

## Testing

Unit tests reside under `synnergy-network/tests`. Run the entire suite with:

```bash
cd synnergy-network
go test ./...
```

## Code Style

- Follow idiomatic Go formatting via `gofmt`
- Keep pull requests small: no more than three files
- Document exported functions and types

## Extending the Project

When adding new packages or features:

1. Place Go code under `synnergy-network`
2. Update or add documentation under `docs/`
3. Run `go fmt`, `go vet`, `go build` and `go test` for affected packages

## Branching and Release Strategy

- The `main` branch holds the latest stable code.
- Create feature branches named `feature/<short-description>` and rebase on `main` before opening a pull request.
- Releases are tagged with semantic versions and built from `main`.

## Commit Guidelines

- Write imperative commit messages such as `add wallet cache`.
- Reference issue numbers when relevant.
- Group logically related changes into separate commits.

## CI and Code Review

- All pull requests run `go fmt`, `go vet`, `go build` and `go test` in CI.
- At least one reviewer must approve before merging.
- New features require unit tests and updated documentation.

## Issue Reporting

Report bugs or feature requests through GitHub issues and include:

1. Environment details (`go version`, OS)
2. Steps to reproduce
3. Expected vs. actual behavior

