# Synnergy Enterprise Development Playbook

Synnergy Network is an experimental blockchain written in Go. This repository hosts the command line utilities, core modules and example smart contracts used to simulate a full network.  This document serves as the primary guide for contributors.  It explains how to set up the environment, the workflow for staged development and the current status of every package.  It also records future work needed to transform the prototype into a production‑ready platform.

## Repository Overview

```
./setup_synn.sh          # Minimal bootstrap script for building the CLI
./Synnergy.env.sh        # Full environment setup and dependency installer
synnergy-network/        # Go modules, commands and smart contracts
  cmd/                   # CLI source code and convenience scripts
  core/                  # Core blockchain modules
  tests/                 # Unit tests for the core packages
  WHITEPAPER.md          # Project vision and design goals
  README.md              # Build and usage instructions
```

Use `module_guide.md` under `synnergy-network/core` for a summary of each module.

## Environment Setup

1. Run `./setup_synn.sh` to install Go and fetch module dependencies.
2. Optionally execute `./Synnergy.env.sh` to install additional tools and load variables from `.env`.  This script also downloads Go modules and sets up `$GOPATH`.
3. After either script completes, enter the `synnergy-network` directory and run `go mod tidy` to ensure all dependencies are present.
4. Always start from a clean git state before beginning work on a new stage.

## Standard Workflow

1. Pick the next unchecked file from the stage lists below.  Work on **no more than three files at a time** to keep pull requests focused.
2. Verify no other open PR covers the same stage.  If it does, move to the next available stage.
3. Fix compile or logic errors in the selected files.
4. Run `go fmt` on all modified Go files.
5. Run `go vet` and `go build` only on the packages containing your changes, for example `go vet ./cmd/...`.
6. Execute `go test` for those same packages.  Unit tests live in `synnergy-network/tests` and can be invoked with `go test ./synnergy-network/core/<module>`.
7. If `go vet`, `go build` or `go test` fail due to missing dependencies, run `go mod tidy` and retry.
8. When all checks pass, update the checkboxes in this document and commit your changes.

### Using the CLI

Once the CLI builds successfully you can start a local node:

```bash
cd synnergy-network
go build -o synnergy ./cmd/synnergy
./synnergy ledger init --path ./ledger.db
./synnergy network start
```

In another terminal you may create a wallet and mint coins for testing:

```bash
./synnergy wallet create --out wallet.json
./synnergy coin mint $(jq -r .address wallet.json) 1000
```

Additional commands such as contract deployment or token transfers are documented in `cmd/cli/cli_guide.md`.

## Staged Implementation Plan

Development is divided into twenty‑five stages to keep the effort manageable.  Complete each stage in order, marking the checkboxes when the associated files compile and their tests pass.

1. **Stage 1** – CLI: `ai.go`, `amm.go`, `authority_node.go`
2. **Stage 2** – CLI: `charity_pool.go`, `coin.go`, `compliance.go`
3. **Stage 3** – CLI: `consensus.go`, `contracts.go`, `cross_chain.go`
4. **Stage 4** – CLI: `data.go`, `fault_tolerance.go`, `governance.go`
5. **Stage 5** – CLI: `green_technology.go`, `index.go`, `ledger.go`
6. **Stage 6** – CLI: `liquidity_pools.go`, `loanpool.go`, `network.go` ✅
7. **Stage 7** – CLI: `replication.go`, `rollups.go`, `security.go` ✅
8. **Stage 8** – CLI: `sharding.go`, `sidechain.go`, `state_channel.go` ✅
9. **Stage 9** – CLI: `storage.go`, `tokens.go`, `transactions.go` ✅
10. **Stage 10** – CLI: `utility_functions.go`, `virtual_machine.go`, `wallet.go` ✅
11. **Stage 11** – Core: `ai.go`, `amm.go`, `authority_nodes.go`
12. **Stage 12** – Core: `charity_pool.go`, `coin.go`, `common_structs.go`
13. **Stage 13** – Core: `compliance.go`, `consensus.go`, `contracts.go`
14. **Stage 14** – Core: `contracts_opcodes.go`, `cross_chain.go`, `data.go` ✅
15. **Stage 15** – Core: `fault_tolerance.go`, `gas_table.go`, `governance.go` ✅
16. **Stage 16** – Core: `green_technology.go`, `ledger.go`, `ledger_test.go` ✅
17. **Stage 17** – Core: `liquidity_pools.go`, `loanpool.go`, `network.go` ✅
18. **Stage 18** – Core: `opcode_dispatcher.go`, `replication.go`, `rollups.go` ✅
19. **Stage 19** – Core: `security.go`, `sharding.go`, `sidechains.go` ✅
20. **Stage 20** – Core: `state_channel.go`, `storage.go`, `tokens.go` ✅
21. **Stage 21** – Core: `transactions.go`, `utility_functions.go`, `virtual_machine.go` ✅
22. **Stage 22** – Core: `wallet.go` and review prior fixes ✅
23. **Stage 23** – Run integration tests across CLI packages
24. **Stage 24** – Launch a local network and verify node startup
25. **Stage 25** – Final pass through documentation and ensure all tests pass

## CLI Files Checklist

- [x] ai.go
- [x] amm.go
- [x] authority_node.go
- [x] charity_pool.go
- [x] coin.go
- [x] compliance.go
- [x] consensus.go
- [x] contracts.go
- [x] cross_chain.go
- [x] data.go
- [x] fault_tolerance.go
- [x] governance.go
- [x] green_technology.go
- [x] index.go
- [x] ledger.go
- [x] liquidity_pools.go
- [x] loanpool.go
- [x] network.go
- [x] replication.go
- [x] rollups.go
- [x] security.go
- [x] sharding.go
- [x] sidechain.go
- [x] state_channel.go
- [x] storage.go
- [x] tokens.go
- [x] transactions.go
- [x] utility_functions.go
- [x] virtual_machine.go
- [x] wallet.go

## Core Module Checklist

All modules reside under `synnergy-network/core`.  The `module_guide.md` file in that directory explains their responsibilities in detail.

- [x] ai.go
- [x] amm.go
- [x] authority_nodes.go
- [x] charity_pool.go
- [x] coin.go
- [x] common_structs.go
- [x] compliance.go
- [x] consensus.go
- [x] contracts.go
- [x] contracts_opcodes.go
- [x] cross_chain.go
- [x] data.go
- [x] fault_tolerance.go
- [x] gas_table.go
- [x] governance.go
- [x] green_technology.go
- [x] ledger.go
- [x] ledger_test.go
- [x] liquidity_pools.go
- [x] loanpool.go
- [x] network.go
- [x] opcode_dispatcher.go
- [x] replication.go
- [x] rollups.go
- [x] security.go
- [x] sharding.go
- [x] sidechains.go
- [x] state_channel.go
- [x] storage.go
- [x] tokens.go
- [x] transactions.go
- [x] utility_functions.go
- [x] virtual_machine.go
- [x] wallet.go

## Guidance and Tips

- Always work through the stages sequentially and keep your pull requests small.
- Modify no more than three files at once to reduce merge conflicts.
- Run formatting and tests as described in the workflow section before committing.
- Refer to `setup_synn.sh` for a quick environment bootstrap and `Synnergy.env.sh` for a more complete tool chain.
- Stub helpers are provided in `core/helpers.go`:
  - `core.InitLedger` and `core.CurrentLedger` manage a shared ledger instance.
  - `core.InitAuthoritySet` and `core.CurrentAuthoritySet` manage validator info.
  - `core.NewTFStubClient` returns a no‑op AI gRPC client for development.
  - `core.NewFlatGasCalculator` provides a simple gas model for tests and CLI prototypes.

## Upgrade Roadmap

The current code base is a functional prototype.  The following additions would help move the project toward a production‑ready blockchain:

### Missing Bash Scripts

- `faucet_fund.sh` – automate funding test accounts from a faucet service
- `dao_vote.sh` – submit DAO proposals and cast votes
- `marketplace_list.sh` – list items for sale in a decentralized marketplace
- `storage_marketplace_pin.sh` – interface with the storage marketplace module
- `loanpool_apply.sh` – submit a loan application transaction
- `authority_apply.sh` – request validator or authority node status

### Desired Smart Contracts

- **Faucet Contract** – simple token dispenser with rate limiting
- **DAO Governance Contract** – on‑chain voting and proposal execution
- **Marketplace Contract** – buy and sell digital goods using SYNN tokens
- **Storage Marketplace Contract** – pay for decentralized storage
- **LoanPool Application Contract** – terms negotiation and repayment logic
- **Authority Node Application Contract** – stake tokens to join the validator set

### Additional Directories and Tools

- `cmd/faucet/` – CLI commands for interacting with the faucet service
- `cmd/dao/` – DAO management commands
- `cmd/marketplace/` – marketplace operations
- `cmd/storage_marketplace/` – storage trading commands
- `cmd/loanpool_apply/` – loan application utilities
- `cmd/authority_apply/` – validator application workflow
- `internal/` packages for shared utilities and cross‑package helpers
- `GUI/` – web interfaces for wallet management, explorers and marketplaces
-   - `wallet/`, `explorer/`, `smart-contract-marketplace/`, `ai-marketplace/`,
    `storage-marketplace/`, `dao-explorer/`, `token-creation-tool/`,
    `dex-screener/`, `authority-node-index/`, `cross-chain-management/`
- `scripts/devnet_start.sh` for spinning up a multi‑node local network
- `scripts/benchmark.sh` for load and performance testing

### GUI Tasks

Frontend development lives under `synnergy-network/GUI`. Each subdirectory
contains a basic HTML page, a JavaScript stub and a README to bootstrap
interfaces such as the wallet, explorer and various marketplaces. These
projects are designed for web hosting and will evolve alongside the core
modules.
Recent work added an Express-based backend and modular Bootstrap frontend for the token-creation-tool GUI.

These upgrades will require corresponding tests and documentation.  Contributors are encouraged to propose additional improvements as they work through the stages.

---

This playbook should be kept up to date as the project evolves.  Check off files as they are completed and add new tasks or modules to the roadmap so that future developers have a clear picture of the current status.
