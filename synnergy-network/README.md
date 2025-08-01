# Synnergy Network

Synnergy is a research blockchain exploring modular components for
decentralised finance, data storage and AI‑powered compliance.  It is written in
Go and ships a collection of command line utilities alongside the core
libraries.  The project focuses on extensibility rather than production
readiness.  Further background can be found in [`WHITEPAPER.md`](WHITEPAPER.md).

The primary entrypoint is the `synnergy` binary which exposes a large set of
sub‑commands through the
[Cobra](https://github.com/spf13/cobra) framework. Development helpers in
`core/helpers.go` allow the CLI to operate without a full node while modules are
implemented incrementally.

## Directory Layout

The repository is organised as follows:

| Path | Description |
|------|-------------|
| `core/` | Core blockchain modules such as consensus, storage and smart contract logic. A detailed list is available in [`core/module_guide.md`](core/module_guide.md). |
| `cmd/` | Command line sources, configuration files and helper scripts. CLI modules live under `cmd/cli`. |
| `tests/` | Go unit tests covering each module. Run with `go test ./...`. |
| `third_party/` | Vendored dependencies such as a libp2p fork used during early development. |
| `setup_synn.sh` | Convenience script that installs Go and builds the CLI. |
| `Synnergy.env.sh` | Optional environment bootstrap script that downloads tools and loads variables from `.env`. |

## Building

Go 1.20 or newer is recommended. Clone the repository and run the build command
from the repository root:

```bash
cd synnergy-network
go build ./cmd/synnergy
```

The resulting binary `synnergy` can then be executed with any of the available
sub‑commands. Development stubs in `core/helpers.go` expose `InitLedger` and
`NewTFStubClient` so the CLI can run without a full node during testing.

The provided script `./setup_synn.sh` automates dependency installation and
builds the CLI.  For a fully configured environment with additional tooling and
variables loaded from `.env`, run `./Synnergy.env.sh` after cloning the
repository.

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
- `resource` – manage stored data and VM gas allocations
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

## Core Modules

The Go packages under `core/` implement the blockchain runtime. Key modules
include consensus, storage, networking and the virtual machine.  A summary of
every file is maintained in [`core/module_guide.md`](core/module_guide.md). New
contributors should review that document to understand dependencies between
packages.

## Configuration

Runtime settings are defined using YAML files in `cmd/config/`.  The CLI loads
`default.yaml` by default and merges any environment specific file if the
`SYNN_ENV` environment variable is set (for example `SYNN_ENV=prod`).
`bootstrap.yaml` provides a template for running a dedicated bootstrap node.
The configuration schema is documented in [`cmd/config/config_guide.md`](cmd/config/config_guide.md).

## Running a Local Network

Once the CLI has been built you can initialise a test ledger and start the core
services locally.  A detailed walk‑through is provided in
[`cmd/synnergy/synnergy_set_up.md`](cmd/synnergy/synnergy_set_up.md), but the
basic steps are:

```bash
synnergy ledger init --path ./ledger.db
synnergy network start &
```

Additional helper scripts live under `cmd/scripts`.  Running
`start_synnergy_network.sh` will build the CLI, launch networking, consensus and
other daemons, then run a demo security command.


## Docker

A `Dockerfile` at the repository root allows running Synnergy without a local Go installation.
Build the image and start a node with:

```bash
docker build -t synnergy ..
docker run --rm -it synnergy
```

The container launches networking, consensus, replication and VM services automatically.

## Testing

Unit tests are located in the `tests` directory. Run them using:

```bash
go test ./...
```

Some tests rely on running services such as the network or security daemon. They
may require additional environment variables or mock implementations.

## Contributing

Development tasks are organised in stages described in [`AGENTS.md`](../AGENTS.md).
When contributing code, work through the stages sequentially and modify no more
than three files per commit.  Run `go fmt`, `go vet` and `go build` on the
packages you touch, then execute the relevant unit tests.  Mark completed files
in `AGENTS.md` so others know which tasks are in progress.

The `setup_synn.sh` script should be used when preparing a new environment.

## License

Synnergy is provided for research and educational purposes.  Third‑party
dependencies located under `third_party/` retain their original licenses.  Refer
to the respective `LICENSE` files in those directories for details.
