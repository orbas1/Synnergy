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
- **plasma** – Deposit into and withdraw from the Plasma chain.
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

## Command Reference

The sections below list each root command and its available sub‑commands. Every
command maps directly to logic in `synnergy-network/core` and can be composed as
needed in custom tooling.

### ai

| Sub-command | Description |
|-------------|-------------|
| `predict <tx.json>` | Predict fraud probability for a transaction. |
| `optimise <stats.json>` | Suggest an optimal base fee for the next block. |
| `volume <stats.json>` | Forecast upcoming transaction volume. |
| `publish <cid>` | Publish a model hash with optional royalty basis points. |
| `fetch <model-hash>` | Fetch metadata for a published model. |
| `list <price> <cid>` | Create a marketplace listing for a model. |
| `buy <listing-id> <buyer-addr>` | Buy a listed model with escrow. |
| `rent <listing-id> <renter-addr> <hours>` | Rent a model for a period of time. |
| `release <escrow-id>` | Release funds from escrow to the seller. |

### amm

| Sub-command | Description |
|-------------|-------------|
| `init <fixture-file>` | Initialise pools from a JSON fixture. |
| `swap <tokenIn> <amtIn> <tokenOut> <minOut> [trader]` | Swap tokens via the router. |
| `add <poolID> <provider> <amtA> <amtB>` | Add liquidity to a pool. |
| `remove <poolID> <provider> <lpTokens>` | Remove liquidity from a pool. |
| `quote <tokenIn> <amtIn> <tokenOut>` | Estimate output amount without executing. |
| `pairs` | List all tradable token pairs. |

### authority_node

| Sub-command | Description |
|-------------|-------------|
| `register <addr> <role>` | Submit a new authority-node candidate. |
| `vote <voterAddr> <candidateAddr>` | Cast a vote for a candidate. |
| `electorate <size>` | Sample a weighted electorate of active nodes. |
| `is <addr>` | Check if an address is an active authority node. |
| `info <addr>` | Display details for an authority node. |
| `list` | List authority nodes. |
| `deregister <addr>` | Remove an authority node and its votes. |

### charity_pool

| Sub-command | Description |
|-------------|-------------|
| `register <addr> <category> <name>` | Register a charity with the pool. |
| `vote <voterAddr> <charityAddr>` | Vote for a charity during the cycle. |
| `tick [timestamp]` | Manually trigger pool cron tasks. |
| `registration <addr> [cycle]` | Show registration info for a charity. |
| `winners [cycle]` | List winning charities for a cycle. |

### coin

| Sub-command | Description |
|-------------|-------------|
| `mint <addr> <amt>` | Mint the base SYNN coin. |
| `supply` | Display total supply. |
| `balance <addr>` | Query balance for an address. |
| `transfer <from> <to> <amt>` | Transfer SYNN between accounts. |
| `burn <addr> <amt>` | Burn SYNN from an address. |

### compliance

| Sub-command | Description |
|-------------|-------------|
| `validate <kyc.json>` | Validate and store a KYC document commitment. |
| `erase <address>` | Remove a user's KYC data. |
| `fraud <address> <severity>` | Record a fraud signal. |
| `risk <address>` | Retrieve accumulated fraud risk score. |
| `audit <address>` | Display the audit trail for an address. |
| `monitor <tx.json> <threshold>` | Run anomaly detection on a transaction. |
| `verifyzkp <blob.bin> <commitmentHex> <proofHex>` | Verify a zero‑knowledge proof. |

### consensus

| Sub-command | Description |
|-------------|-------------|
| `start` | Launch the consensus engine. |
| `stop` | Gracefully stop the consensus service. |
| `info` | Show consensus height and running status. |
| `weights <demand> <stake>` | Calculate dynamic consensus weights. |
| `threshold <demand> <stake>` | Compute the consensus switch threshold. |
| `set-weight-config <alpha> <beta> <gamma> <dmax> <smax>` | Update weight coefficients. |
| `get-weight-config` | Display current weight configuration. |

### contracts

| Sub-command | Description |
|-------------|-------------|
| `compile <src.wat|src.wasm>` | Compile WAT or WASM to deterministic bytecode. |
| `deploy --wasm <path> [--ric <file>] [--gas <limit>]` | Deploy compiled WASM. |
| `invoke <address>` | Invoke a contract method. |
| `list` | List deployed contracts. |
| `info <address>` | Show Ricardian manifest for a contract. |

### cross_chain

| Sub-command | Description |
|-------------|-------------|
| `register <source_chain> <target_chain> <relayer_addr>` | Register a bridge. |
| `list` | List registered bridges. |
| `get <bridge_id>` | Retrieve a bridge configuration. |
| `authorize <relayer_addr>` | Whitelist a relayer address. |
| `revoke <relayer_addr>` | Remove a relayer from the whitelist. |

### data

**Node operations**

| Sub-command | Description |
|-------------|-------------|
| `node register <address> <host:port> <capacityMB>` | Register a CDN node. |
| `node list` | List CDN nodes. |

**Asset operations**

| Sub-command | Description |
|-------------|-------------|
| `asset upload <filePath>` | Upload and pin an asset. |
| `asset retrieve <cid> [output]` | Retrieve an asset by CID. |

**Oracle feeds**

| Sub-command | Description |
|-------------|-------------|
| `oracle register <source>` | Register a new oracle feed. |
| `oracle push <oracleID> <value>` | Push a value to an oracle feed. |
| `oracle query <oracleID>` | Query the latest oracle value. |
| `oracle list` | List registered oracles. |

### fault_tolerance

| Sub-command | Description |
|-------------|-------------|
| `snapshot` | Dump current peer statistics. |
| `add-peer <addr>` | Add a peer to the health-checker set. |
| `rm-peer <addr|id>` | Remove a peer from the set. |
| `view-change` | Force a leader rotation. |
| `backup` | Create a ledger backup snapshot. |
| `restore <file>` | Restore ledger state from a snapshot. |
| `failover <addr>` | Force failover of a node. |
| `predict <addr>` | Predict failure probability for a node. |

### governance

| Sub-command | Description |
|-------------|-------------|
| `propose` | Submit a new governance proposal. |
| `vote <proposal-id>` | Cast a vote on a proposal. |
| `execute <proposal-id>` | Execute a proposal after the deadline. |
| `get <proposal-id>` | Display a single proposal. |
| `list` | List all proposals. |

### green_technology

| Sub-command | Description |
|-------------|-------------|
| `usage <validator-addr>` | Record energy and carbon usage for a validator. |
| `offset <validator-addr>` | Record carbon offset credits. |
| `certify` | Recompute certificates immediately. |
| `cert <validator-addr>` | Show the sustainability certificate. |
| `throttle <validator-addr>` | Check if a validator should be throttled. |
| `list` | List certificates for all validators. |

### ledger

| Sub-command | Description |
|-------------|-------------|
| `head` | Show chain height and latest block hash. |
| `block <height>` | Fetch a block by height. |
| `balance <addr>` | Display token balances of an address. |
| `utxo <addr>` | List UTXOs for an address. |
| `pool` | List mem-pool transactions. |
| `mint <addr>` | Mint tokens to an address. |
| `transfer <from> <to>` | Transfer tokens between addresses. |

### liquidity_pools

| Sub-command | Description |
|-------------|-------------|
| `create <tokenA> <tokenB> [feeBps]` | Create a new liquidity pool. |
| `add <poolID> <provider> <amtA> <amtB>` | Add liquidity to a pool. |
| `swap <poolID> <trader> <tokenIn> <amtIn> <minOut>` | Swap tokens within a pool. |
| `remove <poolID> <provider> <lpTokens>` | Remove liquidity from a pool. |
| `info <poolID>` | Show pool state. |
| `list` | List all pools. |

### loanpool

| Sub-command | Description |
|-------------|-------------|
| `submit <creator> <recipient> <type> <amount> <desc>` | Submit a loan proposal. |
| `vote <voter> <id>` | Vote on a proposal. |
| `disburse <id>` | Disburse an approved loan. |
| `tick` | Process proposals and update cycles. |
| `get <id>` | Display a single proposal. |
| `list` | List proposals in the pool. |

### network

| Sub-command | Description |
|-------------|-------------|
| `start` | Start the networking stack. |
| `stop` | Stop network services. |
| `peers` | List connected peers. |
| `broadcast <topic> <data>` | Publish data on the network. |
| `subscribe <topic>` | Subscribe to a topic. |

### replication

| Sub-command | Description |
|-------------|-------------|
| `start` | Launch replication goroutines. |
| `stop` | Stop the replication subsystem. |
| `status` | Show replication status. |
| `replicate <block-hash>` | Gossip a known block. |
| `request <block-hash>` | Request a block from peers. |
| `sync` | Synchronize blocks from peers. |

### rollups

| Sub-command | Description |
|-------------|-------------|
| `submit` | Submit a new rollup batch. |
| `challenge <batchID> <txIdx> <proof...>` | Submit a fraud proof for a batch. |
| `finalize <batchID>` | Finalize or revert a batch. |
| `info <batchID>` | Display batch header and state. |
| `list` | List recent batches. |
| `txs <batchID>` | List transactions in a batch. |

### security

| Sub-command | Description |
|-------------|-------------|
| `sign` | Sign a message with a private key. |
| `verify` | Verify a signature. |
| `aggregate <sig1,sig2,...>` | Aggregate BLS signatures. |
| `encrypt` | Encrypt data using XChacha20‑Poly1305. |
| `decrypt` | Decrypt an encrypted blob. |
| `merkle <leaf1,leaf2,...>` | Compute a double-SHA256 Merkle root. |
| `dilithium-gen` | Generate a Dilithium3 key pair. |
| `dilithium-sign` | Sign a message with a Dilithium key. |
| `dilithium-verify` | Verify a Dilithium signature. |
| `anomaly-score` | Compute an anomaly z-score from data. |

### sharding

| Sub-command | Description |
|-------------|-------------|
| `leader get <shardID>` | Show the leader for a shard. |
| `leader set <shardID> <addr>` | Set the leader address for a shard. |
| `map` | List shard-to-leader mappings. |
| `submit <fromShard> <toShard> <txHash>` | Submit a cross-shard transaction header. |
| `pull <shardID>` | Pull receipts for a shard. |
| `reshard <newBits>` | Increase the shard count. |
| `rebalance <threshold>` | List shards exceeding the load threshold. |

### sidechain

| Sub-command | Description |
|-------------|-------------|
| `register` | Register a new side-chain. |
| `header` | Submit a side-chain header. |
| `deposit` | Deposit tokens to a side-chain escrow. |
| `withdraw <proofHex>` | Verify a withdrawal proof. |
| `get-header` | Fetch a submitted side-chain header. |
| `meta <chainID>` | Display side-chain metadata. |
| `list` | List registered side-chains. |

### plasma

| Sub-command | Description |
|-------------|-------------|
| `deposit` | Deposit funds into the Plasma chain. |
| `withdraw <nonce>` | Finalise a Plasma exit. |

### state_channel

| Sub-command | Description |
|-------------|-------------|
| `open` | Open a new payment/state channel. |
| `close` | Submit a signed state to start closing. |
| `challenge` | Challenge a closing state with a newer one. |
| `finalize` | Finalize and settle an expired channel. |
| `status` | Show the current channel state. |
| `list` | List all open channels. |

### storage

| Sub-command | Description |
|-------------|-------------|
| `pin` | Pin a file or data blob to the gateway. |
| `get` | Retrieve data by CID. |
| `listing:create` | Create a storage listing. |
| `listing:get` | Get a storage listing by ID. |
| `listing:list` | List storage listings. |
| `deal:open` | Open a storage deal backed by escrow. |
| `deal:close` | Close a storage deal and release funds. |
| `deal:get` | Get details for a storage deal. |
| `deal:list` | List storage deals. |

### tokens

| Sub-command | Description |
|-------------|-------------|
| `list` | List registered tokens. |
| `info <id|symbol>` | Display token metadata. |
| `balance <tok> <addr>` | Query token balance of an address. |
| `transfer <tok>` | Transfer tokens between addresses. |
| `mint <tok>` | Mint new tokens. |
| `burn <tok>` | Burn tokens from an address. |
| `approve <tok>` | Approve a spender allowance. |
| `allowance <tok> <owner> <spender>` | Show current allowance. |

### transactions

| Sub-command | Description |
|-------------|-------------|
| `create` | Craft an unsigned transaction JSON. |
| `sign` | Sign a transaction JSON with a keystore key. |
| `verify` | Verify a signed transaction JSON. |
| `submit` | Submit a signed transaction to the network. |
| `pool` | List pending pool transaction hashes. |

### utility_functions

| Sub-command | Description |
|-------------|-------------|
| `hash` | Compute a cryptographic hash. |
| `short-hash` | Shorten a 32-byte hash to first4..last4 format. |
| `bytes2addr` | Convert big-endian bytes to an address. |

### virtual_machine

| Sub-command | Description |
|-------------|-------------|
| `start` | Start the VM HTTP daemon. |
| `stop` | Stop the VM daemon. |
| `status` | Show daemon status. |

### wallet

| Sub-command | Description |
|-------------|-------------|
| `create` | Generate a new wallet and mnemonic. |
| `import` | Import an existing mnemonic. |
| `address` | Derive an address from a wallet. |
| `sign` | Sign a transaction JSON using the wallet. |
