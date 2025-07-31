# Synnergy Core Module Guide

This document provides a very high level description of the packages that make up the `core` library.  The goal is to help new contributors navigate the code base and discover where certain functionality lives.

Each Go file in this directory represents a self‑contained module used by the network stack.  Most modules depend only on `common_structs.go` and may be used independently.

## Modules

The following files make up the core runtime.  Each summary touches on the key
responsibilities and dependencies of that module.

- **ai.go** – Connects to a TensorFlow Serving cluster over gRPC and exposes an
  `AIEngine` for anomaly detection and model publishing.  It stores metadata on
  chain and uses a prediction cache to keep fees low.
- **amm.go** – Implements a constant‑product automated market maker with gas‑aware
  routing helpers.  Liquidity pools are stored in `liquidity_pools.go` and this
  file provides user facing helpers such as `AddLiquidity` and `SwapExactIn`.
- **authority_nodes.go** – Manages staking records and validator roles.  Voting
  and random validator selection for consensus are handled here.
- **charity_pool.go** – Collects 5% of every gas fee into a community fund.  It
  schedules 90‑day cycles where registered charities receive payouts based on
  token‑holder votes.
- **coin.go** – Defines the base coin constants and basic mint/transfer logic.
  Used throughout the ledger and wallet packages.
- **common_structs.go** – Shared type definitions that appear in multiple modules
  without creating circular dependencies.
- **compliance.go** – Wraps transaction processing with optional KYC/AML checks
  and audit logging hooks.
- **consensus.go** – Hybrid Proof‑of‑History / Proof‑of‑Stake engine that
  aggregates sub‑blocks under a Proof‑of‑Work main block.  Relies on `ledger`,
  `network` and `security`.
- **contracts.go** – Provides a minimal contract interface executed by the
  virtual machine.  Contracts are stored on chain and invoked by transactions.
- **cross_chain.go** – Utilities for bridging assets between other chains.  This
  includes proof verification and lock/release semantics.
- **data.go** – Low level helpers for reading and writing on‑chain data.  Other
  modules depend on this for simple key/value access.
- **fault_tolerance.go** – Structures for Byzantine fault detection and recovery
  procedures used by consensus and replication.
- **gas_table.go** – Maps VM opcodes to gas costs.  Updated as new opcodes are
  introduced.
- **governance.go** – Handles protocol parameter proposals, voting and enactment
  of changes.
- **green_technology.go** – Experiments around low energy consensus tweaks and
  sustainability metrics.
- **ledger.go** – Core blockchain state including blocks, transactions and
  snapshots.  Supports a write‑ahead log and periodic state snapshots.
- **liquidity_pools.go** – Stores pool reserves and pricing functions used by the
  AMM and loan modules.
- **loanpool.go** – Simple collateralised loan mechanics built on top of the
  token module.
- **network.go** – libp2p networking stack with pub‑sub gossip and peer discovery.
- **opcode_dispatcher.go** – Maps transaction opcodes to virtual machine handlers
  at runtime.
- **replication.go** – Block propagation utilities and snapshot distribution to
  new nodes.
- **rollups.go** – Support for batching transactions into rollups that settle on
  the main chain.
- **security.go** – Ed25519 key management, signature helpers and basic crypto
  utilities.
- **sharding.go** – Splits state into independent shards and provides cross shard
  migration helpers.
- **sidechains.go** – Tools for launching and interacting with application
  specific side chains.
- **state_channel.go** – Manages off‑chain payment channels with on‑chain dispute
  resolution.
- **storage.go** – Backend‑agnostic key/value storage adapters used by the ledger
  and VM.
- **tokens.go** – Registry of built‑in tokens and factory functions for issuing
  new assets.
- **transactions.go** – Core transaction structures, signature validation and
  fee calculation logic.
- **utility_functions.go** – Grab bag of small helpers shared across the code
  base.
- **virtual_machine.go** – Lightweight WebAssembly runtime that executes
  contracts and transaction scripts.
- **wallet.go** – HD wallet implementation with mnemonic generation and signing
  utilities.

This list is intentionally brief.  For detailed behaviour, refer to the inline documentation and unit tests located in `synnergy-network/tests`.
