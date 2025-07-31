# Synnergy Command Line Guide

This short guide summarises the CLI entry points found in `cmd/cli`.  Each Go file wires a set of commands using the [Cobra](https://github.com/spf13/cobra) framework.  Commands are grouped by module and can be imported individually into the root program.

Most commands require environment variables or a configuration file to be present.  Refer to inline comments for a full list of options.

## Available Command Groups

The following command groups expose the same functionality available in the core modules. Each can be mounted on a root [`cobra.Command`](https://github.com/spf13/cobra).

- **ai** – Tools for publishing ML models and running anomaly detection jobs via gRPC to the AI service. Useful for training pipelines and on‑chain inference.
- **amm** – Swap tokens and manage liquidity pools. Includes helpers to quote routes and add/remove liquidity.
- **authority_node** – Register new validators, vote on authority proposals and list the active electorate.
- **charity_pool** – Query the community charity fund and trigger payouts for the current cycle.
- **coin** – Mint the base coin, transfer balances and inspect supply metrics.
- **compliance** – Run KYC/AML checks on addresses and export audit reports.
- **consensus** – Start, stop or inspect the node's consensus service. Provides status metrics for debugging.
- **contracts** – Deploy, upgrade and invoke smart contracts stored on chain.
- **cross_chain** – Bridge assets to or from other chains using lock and release commands.
- **data** – Inspect raw key/value pairs in the underlying data store for debugging.
- **fault_tolerance** – Inject faults, simulate network partitions and test recovery procedures.
- **governance** – Create proposals, cast votes and check DAO parameters.
- **green_technology** – View energy metrics and toggle any experimental sustainability features.
- **ledger** – Inspect blocks, query balances and perform administrative token operations via the ledger daemon.
- **network** – Manage peer connections and print networking statistics.
- **replication** – Trigger snapshot creation and replicate the ledger to new nodes.
- **rollups** – Create rollup batches or inspect existing ones.
- **security** – Key generation, signing utilities and password helpers.
- **sharding** – Migrate data between shards and check shard status.
- **sidechain** – Launch side chains or interact with remote side‑chain nodes.
- **state_channel** – Open, close and settle payment channels.
- **storage** – Configure the backing key/value store and inspect content.
- **tokens** – Register new token types and move balances between accounts.
- **transactions** – Build raw transactions, sign them and broadcast to the network.
- **utility_functions** – Miscellaneous helpers shared by other command groups.
- **virtual_machine** – Execute scripts in the built‑in VM for testing.
- **wallet** – Generate mnemonics, derive addresses and sign transactions.


To use these groups, import the corresponding command constructor (e.g. `ledger.NewLedgerCommand()`) in your main program and attach it to the root `cobra.Command`.

If you want to enable **all** CLI modules with a single call, use `cli.RegisterRoutes(rootCmd)` from the `cli` package. This helper mounts every exported command group so routes can be invoked like:

```bash
$ synnergy ~network ~start
```
