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
| `core/binary_tree_operations.go` | Ledger-backed binary search tree implementation with CLI and VM opcodes. |
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
- `ai_contract` – manage AI enhanced contracts with risk checks
- `ai-train` – manage on-chain AI model training
- `ai_mgmt` – manage AI model marketplace listings
- `ai_infer` – advanced inference and transaction analysis
- `amm` – swap tokens and manage liquidity pools
- `authority_node` – validator registration and voting
- `access` – manage role based access permissions
- `authority_apply` – submit and approve authority node applications
- `charity_pool` – query and disburse community funds
- `charity_mgmt` – manage donations and internal payouts
- `identity` – manage verified addresses
- `coin` – mint and transfer the native SYNN coin
- `compliance` – perform KYC/AML checks
- `audit` – manage on-chain audit logs
- `compliance_management` – suspend or whitelist addresses
- `consensus` – control the consensus engine
- `adaptive` – dynamically adjust consensus weights
- `stake` – adjust validator stake and record penalties
- `contracts` – deploy and invoke smart contracts
- `contractops` – administrative tasks such as pausing and upgrading contracts
- `cross_chain` – configure asset bridges
- `xcontract` – register cross-chain contract mappings
- `cross_tx` – execute cross-chain transfers
- `cross_chain_connection` – manage chain-to-chain links
- `cross_chain_agnostic_protocols` – register cross-chain protocols
- `cross_chain_bridge` – manage cross-chain transfers
- `data` – inspect and manage raw data storage
- `fork` – inspect and resolve chain forks
- `messages` – queue, process and broadcast network messages
- `partition` – partition data and apply compression
- `distribution` – publish datasets and handle paid access
- `oracle_management` – monitor oracle performance and sync feeds
- `data_ops` – advanced data feed operations
- `anomaly_detection` – analyse transactions for suspicious activity
- `resource` – manage stored data and VM gas allocations
- `immutability` – verify the canonical chain state
- `fault_tolerance` – simulate faults and backups
- `plasma` – deposit tokens and process exits on the plasma bridge
- `resource_allocation` – manage per-contract gas limits
- `failover` – manage ledger snapshots and trigger recovery
- `employment` – manage on-chain employment contracts and salaries
- `governance` – DAO style governance commands
- `token_vote` – cast token weighted governance votes
- `qvote` – quadratic voting on governance proposals
- `polls_management` – create and vote on community polls
- `governance_management` – register and control governance contracts
- `reputation_voting` – weighted voting using SYN-REP tokens
- `timelock` – schedule and execute delayed proposals
- `dao` – create and manage DAOs
- `green_technology` – sustainability features
- `resource_management` – track and charge node resources
- `carbon_credit_system` – manage carbon offset projects and credits
- `energy_efficiency` – track transaction energy usage and efficiency metrics
- `ledger` – low level ledger inspection
- `account` – manage basic accounts and balances
- `loanpool` – submit loan proposals and disburse funds
- `loanpool_apply` – manage loan applications with on-chain voting
- `network` – libp2p networking helpers
- `connpool` – manage reusable outbound connections
- `peer` – peer discovery and connection utilities
 - `replication` – snapshot and replicate data
 - `high_availability` – manage standby nodes and promote backups
 - `rollups` – manage rollup batches
- `plasma` – manage plasma deposits and exits
- `replication` – snapshot and replicate data
- `coordination` – coordinate distributed nodes and broadcast ledger state
 - `rollups` – manage rollup batches and aggregator state
- `initrep` – bootstrap a new node by synchronizing the ledger
- `synchronization` – manage blockchain catch-up and progress
- `rollups` – manage rollup batches
- `compression` – save and load compressed ledger snapshots
- `security` – cryptographic utilities
- `firewall` – manage block lists for addresses, tokens and peers
- `biometrics` – manage biometric authentication templates
- `sharding` – shard management
- `sidechain` – launch, manage and interact with sidechains
- `state_channel` – open and settle payment channels
- `plasma` – manage plasma deposits and block commitments
- `state_channel_mgmt` – pause, resume and force-close channels
- `zero_trust_data_channels` – encrypted data channels with ledger-backed escrows
- `swarm` – orchestrate groups of nodes for high availability
- `storage` – interact with on‑chain storage providers
- `legal` – manage Ricardian contracts and signatures
- `ipfs` – manage IPFS pins and retrieval through the gateway
- `resource` – rent computing resources via the marketplace
- `staking` – lock and release tokens for governance
- `dao_access` – manage DAO membership roles
- `sensor` – manage external sensor inputs and webhooks
- `real_estate` – manage tokenised real estate
- `escrow` – manage multi-party escrow accounts
- `marketplace` – buy and sell items using escrow
- `healthcare` – manage on‑chain healthcare records
- `tangible` – register and transfer tangible asset records
- `warehouse` – manage on‑chain inventory records
- `tokens` – ERC‑20 style token commands
- `defi` – insurance, betting and other DeFi operations
- `event_management` – record and query on-chain events
- `token_management` – high level token creation and administration
- `gaming` – create and join simple on-chain games
- `transactions` – build and sign transactions
- `private_tx` – encrypt and submit private transactions
- `transactionreversal` – request authority-backed reversals
- `transaction_distribution` – split transaction fees between miner and treasury
- `faucet` – dispense test tokens or coins with rate limits
- `utility_functions` – assorted helpers
- `geolocation` – record node location information
- `distribution` – bulk token distribution and airdrops
- `finalization_management` – finalize blocks, rollup batches and channels
- `quorum` – simple quorum tracker management
- `virtual_machine` – run the on‑chain VM service
- `sandbox` – manage VM sandboxes
- `workflow` – automate multi-step tasks with triggers and webhooks
- `supply` – manage supply chain assets on chain
- `wallet` – mnemonic generation and signing
- `execution` – orchestrate block creation and transaction execution
- `regulator` – manage approved regulators and enforce rules
- `feedback` – submit and query user feedback
- `system_health` – monitor runtime metrics and write logs
- `idwallet` – register ID-token wallets and verify status
- `offwallet` – offline wallet utilities
- `recovery` – register and invoke account recovery
- `wallet_mgmt` – manage wallets and send SYNN directly via the ledger

Quadratic voting allows token holders to weight their governance votes by the
square root of the staked amount. The `qvote` command submits these weighted
votes and queries results alongside standard governance commands.

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

By default the node attempts to open the listening port on your router using
UPnP or NAT‑PMP. You can inspect and manage these mappings with the new `nat`
command group.

Additional helper scripts live under `cmd/scripts`.  Running
`start_synnergy_network.sh` will build the CLI, launch networking, consensus and
other daemons, then run a demo security command.

Two top level scripts provide larger network setups:
`scripts/devnet_start.sh` spins up a local multi-node developer network, while
`scripts/testnet_start.sh` starts an ephemeral testnet defined by a YAML
configuration. Both build the CLI automatically and clean up all processes on
`Ctrl+C`.


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
