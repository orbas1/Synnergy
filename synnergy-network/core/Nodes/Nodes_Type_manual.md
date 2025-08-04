[# Node Types Manual

This guide surveys every node variant available in the Synnergy network.  Each
node type couples the common peer‐to‐peer layer with additional services or
responsibilities.  The sections below outline the purpose and main
capabilities of each role.

## Foundation and Consensus

### Base Node
- Minimal peer in the network providing dialling, broadcasting and
  subscription APIs.
- Used as an embedded component by most higher level nodes.

### Full Node
- Maintains the entire ledger and block history.
- Validates blocks, propagates transactions and exposes RPC utilities.

### Light Node
- Stores only block headers for lightweight clients.
- Relies on connected peers for full block data while still verifying headers.

### Master Node
- High availability service node capable of acting as a routing hub for
  subordinate peers.
- Designed for enhanced reliability and uptime guarantees.

### Super Node
- Extends the basic interface with smart‑contract execution and persistent
  storage helpers.
- Acts as a heavy duty execution environment for decentralised applications.

### Validator Node
- Participates in consensus using Proof‑of‑History, Proof‑of‑Stake and
  Proof‑of‑Work components.
- Manages staking, validator registration and penalty enforcement.

### Bootstrap Node
- Seed node that new peers contact to discover the rest of the network.
- Maintains a rolling list of reachable peers for healthy network growth.

### Consensus Specific Node
- Specialised node tuned for pluggable consensus mechanisms.
- Provides hooks for adapting consensus parameters at runtime.

## Governance and Authority

### Authority Nodes
- Umbrella term for nodes with elevated governance powers.
- Roles include Government, CentralBank, Regulator, CreditorBank and Standard
  authority nodes, each with bespoke permissions over token issuance and
  protocol upgrades.

### Elected Authority Node
- Gains authority status once votes exceed five percent of all nodes and loses
  it when reports pass two and a half percent.
- Offers hooks for validating transactions, creating blocks and reversing
  operations under community oversight.

### Government Authority Node
- Integrates the compliance engine to check transactions against regulatory
  policy.
- Can freeze transactions, update legal frameworks and provide audit trails.

### Central Banking Node
- The only role authorised to deploy SYN‑10/11/12 monetary token standards.
- Manages issuance, redemption and macro‑level monetary policy actions.

### Regulatory Node
- Proposes security upgrades and enforces network regulations.
- Interfaces with external regulators and stores policy metadata on‑chain.

### Bank Institutional Node
- Tailored for large financial institutions.
- Validates and queues transactions, performs compliance analytics and connects
  to legacy banking networks.

### Custodial Node
- Offers asset custody services.
- Handles account registration, deposits, withdrawals, transfers and audit
  reporting with on‑ledger tracking.

### Staking Node
- Specialised interface for staking operations.
- Enables stake management and reward distribution for delegators.

### Time Locked Node
- Manages contracts or accounts that enforce time based restrictions.
- Useful for escrows and delayed release mechanisms.

## Infrastructure and Integration

### API Node
- HTTP gateway exposing balance queries, transaction submission and block data.
- Utilises a standard node for peer communication while serving REST requests.

### Gateway Node
- Provides cross‑chain connectivity and external data ingestion.
- Manages connections to other chains and publishes external data to the
  network.

### Content Node
- Registers storage capacity and replicates encrypted content across the CDN.
- Handles content pinning, retrieval and node registry maintenance.

### Indexing Node
- Builds and serves searchable indexes over blockchain data.
- Optimised for fast lookups and analytical queries.

### Integration Node
- Connects off‑chain services and oracles into the network.
- Maintains registries for third‑party integrations.

### Lightning Node
- Facilitates off‑chain payment channels for instant settlement.
- Monitors channel states and propagates updates.

### Mining Node
- Performs proof‑of‑work hashing to propose new blocks.
- Can operate as a standalone miner or in pool configuration.

### Mobile Node
- Lightweight node designed for mobile devices.
- Optimises bandwidth and storage while retaining key wallet functions.

### Mobile Mining Node
- Mobile‑optimised miner capable of contributing hash power from handheld
  devices.
- Includes power management and simplified control APIs.

### Orphan Node
- Maintains awareness of orphaned blocks for chain reorganisation analysis.
- Useful for research and network diagnostics.

### Indexing and Archival Nodes
- **Historical Node** keeps an archive of past blocks and serves them to
  clients.
- **Archival Witness Node** notarises transactions and blocks with witness
  signatures, storing certifiable records on the ledger.

## Analytics, Monitoring and Recovery

### Audit Node
- Performs compliance audits and records findings for governance review.

### Autonomous Agent Node
- Executes user‑defined rules automatically based on on‑chain data.
- Supports programmable triggers for automation and DAO style agents.

### Disaster Recovery Node
- Stores backup metadata and assists in restoring the network after failures.

### Environmental Monitoring Node
- Streams sensor data for environmental applications and green tracking.

### Geospatial Node
- Provides location‑aware services and spatial indexing.

### Watchtower Node
- Observes transactions and channel updates, emitting alerts on contract
  violations or suspicious activity.

### Forensic Node
- Runs deep transaction analysis with AI anomaly detection and compliance
  monitoring.

### Energy Efficient Node
- Records energy usage statistics for validators and computes transactions per
  kilowatt‑hour metrics for sustainability.

### AI Enhanced Node
- Couples a standard node with the AI engine for predictive load analysis and
  transaction anomaly scoring.

### Experimental Node
- Sandboxed environment for deploying experimental features or testing
  contracts before mainnet release.

### Holographic Node
- Distributes data holographically for redundancy and can execute stored code on
  demand.

### Molecular Node
- Interfaces with nano‑scale sensors and actuators, enabling atomic
  transactions and molecular data storage.

### Syn845 Debt Node
- Manages SYN‑845 debt instruments.
- Issues debt records, tracks payments and adjusts interest rates.

## Security and Privacy

### Biometric Security Node
- Incorporates biometric authentication for access control to critical
  functions.

### Quantum‑Resistant Node
- Employs post‑quantum cryptography, encrypts and signs messages with
  Dilithium keys and supports secure broadcasts.

### ZKP Node
- Processes transactions accompanied by zero‑knowledge proofs and stores the
  proofs for later verification.

### Regulatory and Compliance Nodes
- **Government Authority Node**, **Regulatory Node** and **Bank Institutional
  Node** collectively enforce policy, monitor transactions and manage reporting
  requirements.

### Warfare Node
- Simulates defence scenarios and coordinates secure communication in hostile
  environments.

### Custodial and Watch Nodes
- **Custodial Node** safeguards assets for users.
- **Watchtower Node** monitors network activity to flag potential fraud.

## Other Specialised Nodes

### Gateway and External Integration Nodes
- Handle cross‑chain bridges and push or pull data from external HTTP services.

### Time Locked Node
- Enforces timelocks on transactions or contract execution.

### Quantum and ZKP Nodes
- Provide advanced cryptography for privacy and future‑proof security.

### Warfare and Orphan Nodes
- Address niche scenarios such as conflict‑zone communication and orphaned block
  tracking.

The Synnergy ecosystem encourages experimentation, therefore new node types can
be composed by embedding the base `NodeInterface` and layering domain specific
logic on top.  This manual provides an overview of the current catalogue and can
serve as a reference when selecting the appropriate node for a given task.
