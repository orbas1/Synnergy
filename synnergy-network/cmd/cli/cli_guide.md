# Synnergy Command Line Guide

This short guide summarises the CLI entry points found in `cmd/cli`.  Each Go file wires a set of commands using the [Cobra](https://github.com/spf13/cobra) framework.  Commands are grouped by module and can be imported individually into the root program.

Most commands require environment variables or a configuration file to be present.  Refer to inline comments for a full list of options.

## Available Command Groups

- **ai** – Manage on‑chain AI models and run anomaly predictions.
- **amm** – Swap tokens and manage liquidity pools through the automated market maker.
- **authority_node** – Inspect or update validator node status.
- **charity_pool** – Query and distribute funds from the community charity pool.
- **coin** – Mint, transfer and query the native coin supply.
- **compliance** – Perform compliance checks or export audit logs.
- **consensus** – Start or stop the consensus service and monitor its status.
- **contracts** – Deploy and invoke smart contracts.
- **cross_chain** – Operations related to bridging assets across chains.
- **data** – Low level data store inspection utilities.
- **fault_tolerance** – Tools for recovery and fault injection testing.
- **governance** – Vote on proposals and manage DAO parameters.
- **green_technology** – Energy saving options and environmental metrics.
- **ledger** – Inspect blocks, balances and perform administrative token actions.
- **network** – Control the libp2p networking layer and monitor peers.
- **replication** – Manage node replication and backup tasks.
- **rollups** – Create or inspect rollup batches.
- **security** – Key management and signature utilities.
- **sharding** – Commands for inspecting shards and migrating data.
- **sidechain** – Manage side‑chains and interact with side‑chain nodes.
- **state_channel** – Open or close payment channels and check channel status.
- **storage** – Interact with the underlying key/value stores.
- **tokens** – List registered tokens and move balances between accounts.
- **transactions** – Build or submit raw transactions to the network.
- **utility_functions** – Miscellaneous helper commands used by other groups.
- **virtual_machine** – Execute scripts directly in the VM sandbox.
- **wallet** – Simple wallet operations such as key generation and signing.

To use these groups, import the corresponding command constructor (e.g. `ledger.NewLedgerCommand()`) in your main program and attach it to the root `cobra.Command`.
