# Synnergy Core Module Guide

This document provides a very high level description of the packages that make up the `core` library.  The goal is to help new contributors navigate the code base and discover where certain functionality lives.

Each Go file in this directory represents a self‑contained module used by the network stack.  Most modules depend only on `common_structs.go` and may be used independently.

## Modules

- **ai.go** – Interfaces with external machine learning services. It exposes an `AIEngine` used for anomaly detection, fee optimisation and model publication.
- **amm.go** – Implementation of an automated market maker used for token swaps and liquidity pools.
- **authority_nodes.go** – Manages validator and authority node information including staking data.
- **charity_pool.go** – Utilities for directing a portion of network revenue to a community charity pool.
- **coin.go** – Basic coin primitives and helpers for minting and transferring the native asset.
- **common_structs.go** – Shared type definitions used across the rest of the modules.
- **compliance.go** – Hooks that integrate regulatory compliance checks with transactions processed by the ledger.
- **consensus.go** – Hybrid Proof‑of‑History / Proof‑of‑Stake consensus engine wiring.
- **contracts.go** – Simple smart contract execution model leveraged by the virtual machine.
- **cross_chain.go** – Facilities for bridging transactions or assets between different chains.
- **data.go** – Basic on‑chain data store helpers used by higher level modules.
- **fault_tolerance.go** – Code paths that deal with Byzantine fault tolerance and recovery.
- **gas_table.go** – Lookup of gas costs for virtual machine opcodes.
- **governance.go** – Voting and DAO related utilities.
- **green_technology.go** – Placeholder for energy efficient consensus tweaks and sustainability features.
- **ledger.go** – Core blockchain ledger implementation supporting blocks, transactions and snapshots.
- **liquidity_pools.go** – Pool management used by the AMM and loan modules.
- **loanpool.go** – Primitive collateralised loan pool logic.
- **network.go** – Peer to peer networking built on `libp2p` for gossiping blocks and transactions.
- **opcode_dispatcher.go** – Maps opcodes from transactions to VM handlers.
- **replication.go** – Utilities for block propagation and ledger replication.
- **rollups.go** – Experimental rollup support for batching transactions off chain.
- **security.go** – Cryptography helpers including key management and signatures.
- **sharding.go** – Helpers for distributing state across shards.
- **sidechains.go** – Lightweight side‑chain management for application specific chains.
- **state_channel.go** – State channel primitives for high throughput payment channels.
- **storage.go** – Key/value storage adapters used by the ledger and VM.
- **tokens.go** – Registry of built‑in tokens and a token factory used throughout the network.
- **transactions.go** – Transaction definitions and validation logic.
- **utility_functions.go** – Assorted small helpers shared by many modules.
- **virtual_machine.go** – Execution environment for contracts and transaction scripts.
- **wallet.go** – Simple wallet utilities for constructing and signing transactions.

This list is intentionally brief.  For detailed behaviour, refer to the inline documentation and unit tests located in `synnergy-network/tests`.
