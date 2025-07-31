# Synnergy Network

Synnergy is a modular blockchain framework written in Go. This repository contains
multiple command line utilities as well as the core libraries that power the
network. The primary entrypoint is the `synnergy` binary which exposes a large
set of sub‑commands through the [Cobra](https://github.com/spf13/cobra)
framework.

## Building

Go 1.20 or newer is recommended. Clone the repository and run the build command
from the repository root:

```bash
cd synnergy-network
go build ./cmd/synnergy
```

The resulting binary `synnergy` can then be executed with any of the available
sub‑commands.

## Command Groups

Each file in `cmd/cli` registers its own group of commands with the root
`cobra.Command`. The main program consolidates them so the CLI surface mirrors
all modules from the core library. Highlights include:

- `ai` – publish models and run inference jobs
- `amm` – swap tokens and manage liquidity pools
- `authority_node` – validator registration and voting
- `charity_pool` – query and disburse community funds
- `coin` – mint and transfer the native SYNN coin
- `compliance` – perform KYC/AML checks
- `consensus` – control the consensus engine
- `contracts` – deploy and invoke smart contracts
- `cross_chain` – configure asset bridges
- `data` – inspect and manage raw data storage
- `fault_tolerance` – simulate faults and backups
- `governance` – DAO style governance commands
- `green_technology` – sustainability features
- `ledger` – low level ledger inspection
- `network` – libp2p networking helpers
- `replication` – snapshot and replicate data
- `rollups` – manage rollup batches
- `security` – cryptographic utilities
- `sharding` – shard management
- `sidechain` – launch and interact with sidechains
- `state_channel` – open and settle payment channels
- `storage` – interact with on‑chain storage providers
- `tokens` – ERC‑20 style token commands
- `transactions` – build and sign transactions
- `utility_functions` – assorted helpers
- `virtual_machine` – run the on‑chain VM service
- `wallet` – mnemonic generation and signing

More details for each command can be found in `cmd/cli/cli_guide.md`.

## Testing

Unit tests are located in the `tests` directory. Run them using:

```bash
go test ./...
```

Some tests rely on running services such as the network or security daemon. They
may require additional environment variables or mock implementations.
