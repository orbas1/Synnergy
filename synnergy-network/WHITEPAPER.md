# Synnergy Network Whitepaper

## Executive Summary
Synnergy Network is a modular blockchain stack built in Go that targets production deployment. It combines decentralized finance, data storage, and AI-driven compliance within a single ecosystem. The network emphasizes extensibility—every component is developed as an independent module for ease of upgrades and custom deployments. This whitepaper outlines the key architecture, token economics, tooling, and long-term business model for Synnergy as it transitions from research into a production-grade platform.

## About Synnergy Network
Synnergy started as a research project exploring novel consensus models and VM design. Its goal is to create a secure and scalable base layer that integrates real-time compliance checks and AI analytics. While previous versions focused on basic CLI experimentation, the current roadmap aims to deliver a mainnet capable of hosting decentralized applications and complex financial workflows. Synnergy is open-source under the BUSL-1.1 license and maintained by a community of contributors.

## Synnergy Ecosystem
The Synnergy ecosystem brings together several services:
- **Core Ledger and Consensus** – The canonical ledger stores blocks and coordinates the validator set.
- **Virtual Machine** – A modular VM executes smart contracts compiled to WASM or EVM-compatible bytecode.
- **Data Layer** – Integrated IPFS-style storage allows assets and off-chain data to be referenced on-chain.
- **AI Compliance** – A built-in AI service scans transactions for fraud patterns, KYC signals, and anomalies.
- **DEX and AMM** – Native modules manage liquidity pools and cross-chain swaps.
- **Governance** – Token holders can create proposals and vote on protocol upgrades.
- **Developer Tooling** – CLI modules, RPC services, and SDKs make integration straightforward.
All services are optional and run as independent modules that plug into the core.

## Synnergy Network Architecture
At a high level the network consists of:
1. **Peer-to-Peer Network** – Validators communicate using libp2p with gossip-based transaction propagation.
2. **Consensus Engine** – A hybrid approach combines Proof of History (PoH) and Proof of Stake (PoS) with pluggable modules for alternative algorithms.
3. **Ledger** – Blocks contain sub-blocks that optimize for data availability. Smart contracts and token transfers are recorded here.
4. **Virtual Machine** – The dispatcher assigns a 24-bit opcode to every protocol function. Gas is charged before execution using a deterministic cost table.
5. **Storage Nodes** – Off-chain storage is coordinated through specialized nodes for cheap archiving and retrieval.
6. **Rollups and Sharding** – Sidechains and rollup batches scale the system horizontally while maintaining security guarantees.
Each layer is intentionally separated so enterprises can replace components as needed (e.g., swap the consensus engine or choose a different storage back end).

## Synthron Coin
The native asset powering the network is `SYNTHRON` (ticker: THRON). It has three main functions:
- **Payment and Transaction Fees** – Every on-chain action consumes gas priced in THRON.
- **Staking** – Validators must lock tokens to participate in consensus and receive block rewards.
- **Governance** – Token holders vote on protocol parameters, feature releases, and treasury expenditures.

### Token Distribution
Initial supply is minted at genesis with a gradual release schedule:
- 40% allocated to validators and node operators to bootstrap security.
- 25% reserved for ecosystem grants and partnerships.
- 20% distributed to the development treasury for ongoing work.
- 10% sold in public rounds to encourage early community involvement.
- 5% kept as a liquidity buffer for exchanges and market making.
The supply inflates annually by 2% to maintain incentives and fund new initiatives.

## Full CLI Guide and Index
Synnergy comes with a powerful CLI built using the Cobra framework. Commands are grouped into modules mirroring the codebase. Below is a concise index; see `cmd/cli/cli_guide.md` for the detailed usage of each command group:
- `ai` – Publish machine learning models and run inference jobs.
- `amm` – Swap tokens and manage liquidity pools.
- `authority_node` – Register validators and manage the authority set.
- `charity_pool` – Contribute to or distribute from community charity funds.
- `coin` – Mint, transfer, and burn the base asset.
- `compliance` – Perform KYC/AML verification and auditing.
- `consensus` – Start or inspect the consensus node.
- `contracts` – Deploy and invoke smart contracts.
- `cross_chain` – Bridge assets to and from external chains.
- `data` – Low-level debugging of key/value storage and oracles.
- `fault_tolerance` – Simulate network failures and snapshot recovery.
- `governance` – Create proposals and cast votes.
- `green_technology` – Manage energy tracking and carbon offsets.
- `energy_efficiency` – Measure transaction energy use and compute efficiency scores.
- `ledger` – Inspect blocks, accounts, and token metrics.
- `liquidity_pools` – Create pools and provide liquidity.
- `loanpool` – Submit loan requests and disburse funds.
- `network` – Connect peers and view network metrics.
- `replication` – Replicate and synchronize ledger data across nodes.
- `rollups` – Manage rollup batches and fraud proofs.
- `security` – Generate keys and sign payloads.
- `sharding` – Split the ledger into shards and coordinate cross-shard messages.
- `sidechain` – Launch or interact with auxiliary chains.
- `state_channel` – Open and settle payment channels.
- `storage` – Manage off-chain storage deals.
- `tokens` – Issue and manage token contracts.
- `transactions` – Build and broadcast transactions manually.
- `utility_functions` – Miscellaneous support utilities.
- `virtual_machine` – Execute VM-level operations for debugging.
- `wallet` – Create wallets and sign transfers.
Each command group supports a help flag to display the individual sub-commands and options.

## Full Opcode and Operand Code Guide
All high-level functions in the protocol are mapped to unique 24-bit opcodes of the form `0xCCNNNN` where `CC` denotes the module category and `NNNN` is a numeric sequence. The catalogue is automatically generated and enforced at compile time. Operands are defined per opcode and typically reference stack values or state variables within the VM.

### Opcode Categories
```
0x01  AI                     0x0F  Liquidity
0x02  AMM                    0x10  Loanpool
0x03  Authority              0x11  Network
0x04  Charity                0x12  Replication
0x05  Coin                   0x13  Rollups
0x06  Compliance             0x14  Security
0x07  Consensus              0x15  Sharding
0x08  Contracts              0x16  Sidechains
0x09  CrossChain             0x17  StateChannel
0x0A  Data                   0x18  Storage
0x0B  FaultTolerance         0x19  Tokens
0x0C  Governance             0x1A  Transactions
0x0D  GreenTech              0x1B  Utilities
0x0E  Ledger                 0x1C  VirtualMachine
                                 0x1D  Wallet
```
The complete list of opcodes along with their handlers can be inspected in `core/opcode_dispatcher.go`. Tools like `synnergy opcodes` dump the catalogue in `<FunctionName>=<Hex>` format to aid audits.

### Operand Format
Operands are encoded as stack inputs consumed by the VM. Most instructions follow an EVM-like calling convention with big-endian word sizes. Specialized opcodes reference named parameters defined in the contracts package. Unknown or unpriced opcodes fall back to a default gas amount to prevent DoS attacks.

## Function Gas Index
Every opcode is assigned a deterministic gas cost defined in `core/gas_table.go`. Gas is charged before execution and refunded if the operation releases resources (e.g., SELFDESTRUCT). Example entries:
```
SwapExactIn       = 4_500
AddLiquidity      = 5_000
RecordVote        = 3_000
RegisterBridge    = 20_000
NewLedger         = 50_000
opSHA256          = 60
```
For a full table of over 200 operations see the gas schedule file. This ensures deterministic transaction fees across the network and simplifies metering in light clients.

## Consensus Guide
Synnergy employs a hybrid consensus combining Proof of History for ordering and Proof of Stake for finality. Validators produce PoH hashes to create a verifiable sequence of events. At defined intervals a committee of stakers signs sub-blocks which are then sealed into main blocks using a lightweight Proof of Work puzzle for spam prevention. This design allows fast block times while providing strong security guarantees. Future versions may enable hot-swappable consensus modules so enterprises can adopt algorithms that suit their regulatory environment.

## Transaction Distribution Guide
Transactions are propagated through a gossip network. Nodes maintain a mempool and relay validated transactions to peers. When a validator proposes a sub-block, it selects transactions from its pool based on fee priority and time of arrival. After consensus, the finalized block is broadcast to all peers and applied to local state. Replication modules ensure ledger data remains consistent even under network partitions or DDoS attempts.

## Financial and Numerical Forecasts
The following projections outline potential adoption metrics and pricing scenarios. These figures are purely illustrative and not financial advice.

### Network Growth Model
- **Year 1**: Target 50 validator nodes and 100,000 daily transactions. Estimated 10 million THRON in circulation with modest staking rewards.
- **Year 2**: Expand to 200 validators and introduce sharding. Daily volume expected to exceed 500,000 transactions. Circulating supply projected at 12 million THRON.
- **Year 3**: Full ecosystem of sidechains and rollups. Goal of 1 million transactions per day and 15 million THRON in circulation. Increased staking and governance participation anticipated.

### Pricing Predictions
Assuming gradual adoption and comparable DeFi activity:
- Initial token sale priced around $0.10 per THRON.
- Year 1 market range $0.15–$0.30 depending on DEX liquidity.
- Year 2 range $0.50–$1.00 as staking rewards attract more validators.
- Year 3 range $1.50–$3.00 if rollups and sidechains capture significant usage.
These estimates rely on continued development, security audits, and ecosystem partnerships.

## Conclusion
Synnergy Network aims to deliver a modular, enterprise-ready blockchain platform that blends advanced compliance, scalable architecture, and developer-friendly tools. The project is moving from early research into production and welcomes community feedback. For source code, development guides, and further documentation visit the repository.

