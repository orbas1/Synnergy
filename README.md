# Synnergy Network Repository

Synnergy is an experimental blockchain written in Go. This repository contains the command line applications, core packages, GUI front‑ends and example smart contracts used to simulate a full network. The code is primarily intended for research and learning. For the vision and background see [`synnergy-network/WHITEPAPER.md`](synnergy-network/WHITEPAPER.md).

## Repository Layout

```
./setup_synn.sh         # minimal bootstrap script for the CLI
./Synnergy.env.sh       # full environment setup with optional tooling
./Dockerfile            # container build for running Synnergy
./scripts/              # helper scripts for devnets and testnets
./synnergy-network/     # Go sources, GUIs and smart contracts
```

Inside `synnergy-network` you will find:

| Path | Description |
|------|-------------|
| `cmd/` | CLI source code organised into many command groups. See [`synnergy-network/README.md`](synnergy-network/README.md) for a detailed list. |
| `core/` | Core blockchain modules implementing consensus, ledger management, networking and the virtual machine. Each file is summarised in [`core/module_guide.md`](synnergy-network/core/module_guide.md). |
| `GUI/` | Web interfaces such as the wallet and explorers. Subprojects include `wallet`, `explorer`, `ai-marketplace`, `storage-marketplace`, `nft_marketplace` and more. |
| `walletserver/` | Go HTTP backend powering the wallet GUI. |
| `tests/` | Unit tests for the core packages. |
| `smart_contract_guide.md` | Guide to writing and deploying smart contracts on Synnergy. |
| `internal/` | Shared utilities for the CLI and services. |

## Building

The CLI requires Go 1.20 or newer. After cloning the repository run:

```bash
./setup_synn.sh        # installs Go and builds synnergy
```

For a complete environment with additional tools run `./Synnergy.env.sh` instead which also loads variables from `.env` if present. Both scripts build the CLI binary in `synnergy-network`.

To build manually:

```bash
cd synnergy-network
go mod tidy
GOFLAGS="-trimpath" go build -o synnergy ./cmd/synnergy
```

To compile binaries for multiple operating systems and architectures and validate the Docker image, run:

```bash
./scripts/build_matrix.sh
```

Built binaries are placed under `dist/` and a `synnergy:latest` Docker image is produced.

## Running a Local Node

Initialise a ledger and start the services:

```bash
cd synnergy-network
./synnergy ledger init --path ./ledger.db
./synnergy network start &
```

You can then open another terminal to create wallets or deploy contracts. The [`cmd/synnergy`](synnergy-network/cmd/synnergy) directory contains additional guides.

Two helper scripts simplify network startup:

- `scripts/devnet_start.sh` spins up multiple local nodes for development.
- `scripts/testnet_start.sh` launches a configurable testnet using a YAML file.

The Dockerfile can build a containerised node which runs the networking, consensus, replication and VM services automatically via `docker-entrypoint.sh`.

## CLI Overview

The CLI exposes dozens of commands grouped by module: AI management, token operations, governance tools, cross‑chain utilities and many more. Each file under `cmd/cli` registers a command group. Refer to [`synnergy-network/README.md`](synnergy-network/README.md) and `cmd/cli/cli_guide.md` for the full catalogue and examples.

## Core Modules

The Go packages in `core/` implement the blockchain runtime. Important modules include consensus, ledger storage, networking layers, data replication, sharding and the virtual machine. Development helpers in `core/helpers.go` allow the CLI to run without a full node. A summary of every file lives in [`core/module_guide.md`](synnergy-network/core/module_guide.md).

## AI Integration

Synnergy's AI engine supports fraud detection, fee optimisation and on‑chain
model marketplaces. Sensitive model parameters and training datasets are
encrypted at rest using a 32‑byte symmetric key supplied via the
`AI_STORAGE_KEY` environment variable. Set this variable before starting any
services that rely on the AI subsystem:

```bash
export AI_STORAGE_KEY="$(openssl rand -hex 16)"
```

The engine exposes gRPC endpoints defined in `ai.proto` for model inference,
training job management and performance drift monitoring. Training jobs are
started through `StartTraining`, their progress is queried with
`TrainingStatus` and completed models can be uploaded via `UploadModel`.

## GUI Projects

Web front‑ends are provided under `GUI/`. Each directory contains a standalone project with its own README. Highlights include:

- `wallet` – manage accounts and sign transactions using the wallet server.
- `explorer` – query balances and transactions via a simple interface.
- `ai-marketplace` – browse and purchase AI models.
- `smart-contract-marketplace` – deploy and trade contracts.
- `storage-marketplace` – pay for decentralized storage services.
- `nft_marketplace` – create and trade NFTs.
- `dao-explorer` – interact with on‑chain DAOs.
- `token-creation-tool` – generate and manage new tokens.
- `dex-screener` – view decentralized exchange listings.
- `authority-node-index` and `cross-chain-management` – administrative dashboards.

## Smart Contracts

Example contracts demonstrating Synnergy's opcode catalogue are located throughout `smart_contract_guide.md` and under various GUI directories. They illustrate token faucets, storage markets, DAO governance and more. Contracts are compiled to WebAssembly and deployed via the CLI. See [`synnergy-network/smart_contract_guide.md`](synnergy-network/smart_contract_guide.md) for a step‑by‑step tutorial.

## Tests

Unit tests reside in `synnergy-network/tests`. Execute them with:

```bash
go test ./...
```

Some tests expect running services or mock implementations. `go vet` and `go build` can be run in the same way to lint and compile the modules.

## Security Scan

Run static analysis with [gosec](https://github.com/securego/gosec) to detect common vulnerabilities:

```bash
./scripts/security_scan.sh
```

High severity findings must be addressed before merging changes.

## Contributing

Development follows the staged workflow described in [`AGENTS.md`](AGENTS.md). Work through the stages sequentially and modify no more than three files per pull request. Run `go fmt`, `go vet`, `go build` and `go test` on the packages you touch. Mark progress in `AGENTS.md` so others know which files are complete.

## License

Synnergy is provided for research and educational purposes. Third‑party dependencies in `third_party/` retain their original licenses. See those directories for details.
