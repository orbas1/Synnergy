# Synnergy Command Line Guide

This short guide summarises the CLI entry points found in `cmd/cli`.  Each Go file wires a set of commands using the [Cobra](https://github.com/spf13/cobra) framework.  Commands are grouped by module and can be imported individually into the root program.

Most commands require environment variables or a configuration file to be present.  Refer to inline comments for a full list of options.

## Available Command Groups

The following command groups expose the same functionality available in the core modules. Each can be mounted on a root [`cobra.Command`](https://github.com/spf13/cobra).

- **ai** – Tools for publishing ML models and running anomaly detection jobs via gRPC to the AI service. Useful for training pipelines and on‑chain inference.
- **ai-train** – Manage on-chain AI model training jobs.
- **ai_mgmt** – Manage marketplace listings for AI models.
- **ai_infer** – Advanced inference and batch analysis utilities.
- **amm** – Swap tokens and manage liquidity pools. Includes helpers to quote routes and add/remove liquidity.
- **authority_node** – Register new validators, vote on authority proposals and list the active electorate.
- **authority_apply** – Submit and vote on authority node applications.
- **charity_pool** – Query the community charity fund and trigger payouts for the current cycle.
- **coin** – Mint the base coin, transfer balances and inspect supply metrics.
 - **compliance** – Run KYC/AML checks on addresses and export audit reports.
 - **compliance_management** – Manage suspensions and whitelists for addresses.
 - **consensus** – Start, stop or inspect the node's consensus service. Provides status metrics for debugging.
- **compliance** – Run KYC/AML checks on addresses and export audit reports.
- **audit** – Manage on-chain audit logs.
- **consensus** – Start, stop or inspect the node's consensus service. Provides status metrics for debugging.
- **adaptive** – Manage adaptive consensus weights.
- **stake** – Adjust validator stakes and record penalties.
- **contracts** – Deploy, upgrade and invoke smart contracts stored on chain.
- **cross_chain** – Bridge assets to or from other chains using lock and release commands.
- **data** – Inspect raw key/value pairs in the underlying data store for debugging.
- **anomaly_detection** – Run anomaly analysis on transactions and list flagged hashes.
- **resource** – Manage stored data and VM gas allocations.
- **immutability** – Verify the chain against the genesis block.
- **fault_tolerance** – Inject faults, simulate network partitions and test recovery procedures.
- **employment** – Manage on-chain employment contracts and salaries.
- **governance** – Create proposals, cast votes and check DAO parameters.
- **token_vote** – Cast token weighted governance votes.
- **qvote** – Submit quadratic votes and view weighted results.
- **polls_management** – Create and vote on community polls.
- **governance_management** – Manage governance contracts on chain.
- **reputation_voting** – Reputation weighted governance commands.
- **timelock** – Manage delayed proposal execution.
- **dao** – Manage DAO creation and membership.
- **green_technology** – View energy metrics and toggle any experimental sustainability features.
- **ledger** – Inspect blocks, query balances and perform administrative token operations via the ledger daemon.
- **account** – manage accounts and balances
- **network** – Manage peer connections and print networking statistics.
 - **replication** – Trigger snapshot creation and replicate the ledger to new nodes.
 - **high_availability** – Manage standby nodes and promote backups.
 - **rollups** – Create rollup batches or inspect existing ones.
- **plasma** – Deposit into and withdraw from the Plasma chain.
- **replication** – Trigger snapshot creation and replicate the ledger to new nodes.
- **rollups** – Create rollup batches or inspect existing ones.
- **compression** – Save and load compressed ledger snapshots.
- **security** – Key generation, signing utilities and password helpers.
- **biometrics** – Manage biometric authentication templates.
- **sharding** – Migrate data between shards and check shard status.
- **sidechain** – Launch side chains or interact with remote side‑chain nodes.
- **state_channel** – Open, close and settle payment channels.
- **zero_trust_data_channels** – Manage encrypted data channels with escrow.
- **swarm** – Manage groups of nodes running together.
- **storage** – Configure the backing key/value store and inspect content.
- **dao_access** – Manage DAO membership roles.
- **sensor** – Manage external sensor inputs and webhooks.
- **real_estate** – Manage tokenised real estate assets.
- **healthcare** – Manage healthcare records and permissions.
- **warehouse** – Manage on-chain inventory records.
- **tokens** – Register new token types and move balances between accounts.
- **event_management** – Emit and query custom events stored on chain.
- **gaming** – Manage simple on-chain games.
- **transactions** – Build raw transactions, sign them and broadcast to the network.
- **transactionreversal** – Reverse confirmed transactions with authority approval.
- **transaction_distribution** – Distribute transaction fees between stakeholders.
- **utility_functions** – Miscellaneous helpers shared by other command groups.
- **quorum** – Manage quorum trackers for proposals or validation.
- **virtual_machine** – Execute scripts in the built‑in VM for testing.
- **supply** – Manage supply chain records.
- **wallet** – Generate mnemonics, derive addresses and sign transactions.
- **regulator** – Manage on-chain regulators and rule checks.
- **feedback** – Submit and review on‑chain user feedback.
- **system_health** – Monitor node metrics and emit log entries.
- **idwallet** – Register ID-token wallets and verify status.
- **offwallet** – Offline wallet utilities.
- **recovery** – Manage account recovery registration and execution.
- **workflow** – Build on-chain workflows using triggers and webhooks.
- **wallet_mgmt** – Manage wallets and submit ledger transfers.
- **devnet** – Launch a local multi-node developer network.
- **testnet** – Start an ephemeral test network from a YAML config.
- **faucet** – Dispense test funds with rate limiting.


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

### ai-train

| Sub-command | Description |
|-------------|-------------|
| `start <datasetCID> <modelCID>` | Begin a new training job. |
| `status <jobID>` | Display status for a training job. |
| `list` | List all active training jobs. |
| `cancel <jobID>` | Cancel a running job. |
### ai_mgmt

| Sub-command | Description |
|-------------|-------------|
| `get <id>` | Fetch a marketplace listing. |
| `ls` | List all AI model listings. |
| `update <id> <price>` | Update the price of your listing. |
| `remove <id>` | Remove a listing you own. |


| Sub-command | Description |
|-------------|-------------|
| `ai_infer run <model-hash> <input-file>` | Execute model inference on input data. |
| `ai_infer analyse <txs.json>` | Analyse a batch of transactions for fraud risk. |

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

### authority_apply

| Sub-command | Description |
|-------------|-------------|
| `submit <candidate> <role> <desc>` | Submit an authority node application. |
| `vote <voter> <id>` | Vote on an application. Use `--approve=false` to reject. |
| `finalize <id>` | Finalize and register the node if the vote passed. |
| `tick` | Check all pending applications for expiry. |
| `get <id>` | Display an application by ID. |
| `list` | List all applications. |

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

### audit

| Sub-command | Description |
|-------------|-------------|
| `log <addr> <event> [meta.json]` | Record an audit event. |
| `list <addr>` | List audit events for an address. |
### compliance_management

| Sub-command | Description |
|-------------|-------------|
| `suspend <addr>` | Suspend an address from transfers. |
| `resume <addr>` | Lift an address suspension. |
| `whitelist <addr>` | Add an address to the whitelist. |
| `unwhitelist <addr>` | Remove an address from the whitelist. |
| `status <addr>` | Show suspension and whitelist status. |
| `review <tx.json>` | Check a transaction before broadcast. |
### anomaly_detection

| Sub-command | Description |
|-------------|-------------|
| `analyze <tx.json>` | Run anomaly detection on a transaction. |
| `list` | List flagged transactions. |

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

### adaptive

| Sub-command | Description |
|-------------|-------------|
| `metrics` | Show current demand and stake levels. |
| `adjust` | Recompute consensus weights. |
| `set-config <alpha> <beta> <gamma> <dmax> <smax>` | Update weighting coefficients. |

### stake

| Sub-command | Description |
|-------------|-------------|
| `adjust <addr> <delta>` | Increase or decrease stake for a validator. |
| `penalize <addr> <points> [reason]` | Record penalty points against a validator. |
| `info <addr>` | Display stake and penalty totals. |

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

### resource

| Sub-command | Description |
|-------------|-------------|
| `store <owner> <key> <file> <gas>` | Store data and set a gas limit. |
| `load <owner> <key> [out|-]` | Load data for a key. |
| `delete <owner> <key>` | Remove stored data and reset the limit. |

### fault_tolerance
- **employment** – Manage on-chain employment contracts and salaries.

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

### token_vote

| Sub-command | Description |
|-------------|-------------|
| `cast <proposal-id> <voter> <token-id> <amount> [approve]` | Cast a token weighted vote on a proposal. |
=======
### qvote

| Sub-command | Description |
|-------------|-------------|
| `cast` | Submit a quadratic vote on a proposal. |
| `results <proposal-id>` | Display aggregated quadratic weights. |
### dao_access

| Sub-command | Description |
|-------------|-------------|
| `add <addr> <role>` | Add a DAO member with role `member` or `admin`. |
| `remove <addr>` | Remove a DAO member. |
| `role <addr>` | Display the member role. |
| `list` | List all DAO members. |
### polls_management

| Sub-command | Description |
|-------------|-------------|
| `create` | Create a new poll. |
| `vote <id>` | Cast a vote on a poll. |
| `close <id>` | Close a poll immediately. |
| `get <id>` | Display a poll. |
| `list` | List existing polls. |
### governance_management

| Sub-command | Description |
|-------------|-------------|
| `contract:add <addr> <name>` | Register a governance contract. |
| `contract:enable <addr>` | Enable a contract for voting. |
| `contract:disable <addr>` | Disable a contract. |
| `contract:get <addr>` | Display contract information. |
| `contract:list` | List registered contracts. |
| `contract:rm <addr>` | Remove a contract from the registry. |
### reputation_voting

| Sub-command | Description |
|-------------|-------------|
| `propose` | Submit a new reputation proposal. |
| `vote <proposal-id>` | Cast a weighted vote using SYN-REP. |
| `execute <proposal-id>` | Execute a reputation proposal. |
| `get <proposal-id>` | Display a reputation proposal. |
| `list` | List all reputation proposals. |
| `balance <addr>` | Show reputation balance. |
### timelock

| Sub-command | Description |
|-------------|-------------|
| `queue <proposal-id>` | Queue a proposal with a delay. |
| `cancel <proposal-id>` | Remove a queued proposal. |
| `execute` | Execute all due proposals. |
| `list` | List queued proposals. |
### dao

| Sub-command | Description |
|-------------|-------------|
| `create <name> <creator>` | Create a new DAO. |
| `join <dao-id> <addr>` | Join an existing DAO. |
| `leave <dao-id> <addr>` | Leave a DAO. |
| `info <dao-id>` | Display DAO information. |
| `list` | List all DAOs. |

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

### account

| Sub-command | Description |
|-------------|-------------|
| `create <addr>` | Create a new account. |
| `delete <addr>` | Delete an account. |
| `balance <addr>` | Show account balance. |
| `transfer` | Transfer between accounts. |

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

### loanpool_apply

| Sub-command | Description |
|-------------|-------------|
| `submit <applicant> <amount> <termMonths> <purpose>` | Submit a loan application. |
| `vote <voter> <id>` | Vote on an application. |
| `process` | Finalise pending applications. |
| `disburse <id>` | Disburse an approved application. |
| `get <id>` | Display a single application. |
| `list` | List loan applications. |

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

### high_availability

| Sub-command | Description |
|-------------|-------------|
| `add <addr>` | Register a standby node. |
| `remove <addr>` | Remove a standby node. |
| `list` | List registered standby nodes. |
| `promote <addr>` | Promote a standby to leader via view change. |
| `snapshot [path]` | Write a ledger snapshot to disk. |

### rollups

| Sub-command | Description |
|-------------|-------------|
| `submit` | Submit a new rollup batch. |
| `challenge <batchID> <txIdx> <proof...>` | Submit a fraud proof for a batch. |
| `finalize <batchID>` | Finalize or revert a batch. |
| `info <batchID>` | Display batch header and state. |
| `list` | List recent batches. |
| `txs <batchID>` | List transactions in a batch. |

### compression

| Sub-command | Description |
|-------------|-------------|
| `save <file>` | Write a compressed ledger snapshot. |
| `load <file>` | Load a compressed snapshot and display the height. |

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

### biometrics

| Sub-command | Description |
|-------------|-------------|
| `enroll <file>` | Enroll biometric data for an address. |
| `verify <file>` | Verify biometric data against an address. |
| `delete <addr>` | Remove stored biometric data. |

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

### zero_trust_data_channels

| Sub-command | Description |
|-------------|-------------|
| `open` | Open a new zero trust data channel. |
| `send` | Send a hex encoded payload over the channel. |
| `close` | Close the channel and release escrow. |

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
### real_estate

| Sub-command | Description |
|-------------|-------------|
| `register` | Register a new property. |
| `transfer` | Transfer a property to another owner. |
| `get` | Get property details. |
| `list` | List properties, optionally by owner. |


### escrow

| Sub-command | Description |
|-------------|-------------|
| `create` | Create a new multi-party escrow |
| `deposit` | Deposit additional funds |
| `release` | Release funds to participants |
| `cancel` | Cancel an escrow and refund |
| `info` | Show escrow details |
| `list` | List all escrows |
### marketplace

| Sub-command | Description |
|-------------|-------------|
| `listing:create <price> <metaJSON>` | Create a marketplace listing. |
| `listing:get <id>` | Fetch a listing by ID. |
| `listing:list` | List marketplace listings. |
| `buy <id> <buyer>` | Purchase a listing via escrow. |
| `cancel <id>` | Cancel an unsold listing. |
| `release <escrow>` | Release escrow funds to seller. |
| `deal:get <id>` | Retrieve deal details. |
| `deal:list` | List marketplace deals. |

| Sub-command | Description |
|-------------|-------------|
| `register <addr>` | Register a patient address. |
| `grant <patient> <provider>` | Allow a provider to submit records. |
| `revoke <patient> <provider>` | Revoke provider access. |
| `add <patient> <provider> <cid>` | Add a record CID for a patient. |
| `list <patient>` | List stored record IDs for a patient. |
### warehouse

| Sub-command | Description |
|-------------|-------------|
| `add` | Add a new inventory item. |
| `remove` | Delete an existing item. |
| `move` | Transfer item ownership. |
| `list` | List all warehouse items. |

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

### event_management

| Sub-command | Description |
|-------------|-------------|
| `emit <type> <data>` | Emit a new event and broadcast it. |
| `list <type>` | List recent events of a given type. |
| `get <type> <id>` | Fetch a specific event by ID. |
### token_management

| Sub-command | Description |
|-------------|-------------|
| `create` | Create a new token. |
| `balance <id> <addr>` | Check balance for a token ID. |
| `transfer <id>` | Transfer tokens between addresses. |
### tangible

| Sub-command | Description |
|-------------|-------------|
| `register <id> <owner> <meta> <value>` | Register a new tangible asset. |
| `transfer <id> <owner>` | Transfer ownership of an asset. |
| `info <id>` | Display asset metadata. |
| `list` | List all tangible assets. |
### gaming

| Sub-command | Description |
|-------------|-------------|
| `create` | Create a new game. |
| `join <id>` | Join an existing game. |
| `finish <id>` | Finish a game and release funds. |
| `get <id>` | Display a game record. |
| `list` | List games. |

### transactions

| Sub-command | Description |
|-------------|-------------|
| `create` | Craft an unsigned transaction JSON. |
| `sign` | Sign a transaction JSON with a keystore key. |
| `verify` | Verify a signed transaction JSON. |
| `submit` | Submit a signed transaction to the network. |
| `pool` | List pending pool transaction hashes. |

### transactionreversal

| Sub-command | Description |
|-------------|-------------|
| `reversal` | Reverse a confirmed transaction. Requires authority signatures. |

### utility_functions

| Sub-command | Description |
|-------------|-------------|
| `hash` | Compute a cryptographic hash. |
| `short-hash` | Shorten a 32-byte hash to first4..last4 format. |
| `bytes2addr` | Convert big-endian bytes to an address. |

### quorum

| Sub-command | Description |
|-------------|-------------|
| `init <total> <threshold>` | Initialise a global quorum tracker. |
| `vote <address>` | Record a vote from an address. |
| `check` | Check if the configured quorum is reached. |
| `reset` | Clear all recorded votes. |
### supply

| Sub-command | Description |
|-------------|-------------|
| `register <id> <desc> <owner> <location>` | Register a new item on chain. |
| `update-location <id> <location>` | Update item location. |
| `status <id> <status>` | Update item status. |
| `get <id>` | Fetch item metadata. |

### virtual_machine

| Sub-command | Description |
|-------------|-------------|
| `start` | Start the VM HTTP daemon. |
| `stop` | Stop the VM daemon. |
| `status` | Show daemon status. |

### swarm

| Sub-command | Description |
|-------------|-------------|
| `add <id> <addr>` | Add a node to the swarm. |
| `remove <id>` | Remove a node from the swarm. |
| `broadcast <tx.json>` | Broadcast a transaction to all nodes. |
| `peers` | List nodes currently in the swarm. |
| `start` | Start consensus for the swarm. |
| `stop` | Stop all nodes and consensus. |

### wallet

| Sub-command | Description |
|-------------|-------------|
| `create` | Generate a new wallet and mnemonic. |
| `import` | Import an existing mnemonic. |
| `address` | Derive an address from a wallet. |
| `sign` | Sign a transaction JSON using the wallet. |

### system_health

| Sub-command | Description |
|-------------|-------------|
| `snapshot` | Display current system metrics. |
| `log <level> <msg>` | Append a message to the system log. |

### idwallet

| Sub-command | Description |
|-------------|-------------|
| `register <address> <info>` | Register wallet and mint a SYN-ID token. |
| `check <address>` | Verify registration status. |
### offwallet

| Sub-command | Description |
|-------------|-------------|
| `create` | Create an offline wallet file. |
| `sign` | Sign a transaction offline using the wallet. |
### recovery

| Sub-command | Description |
|-------------|-------------|
| `register` | Register recovery credentials for an address. |
| `recover` | Restore an address by proving three credentials. |
### workflow

| Sub-command | Description |
|-------------|-------------|
| `new` | Create a new workflow by ID. |
| `add` | Append an opcode name to the workflow. |
| `trigger` | Set a cron expression for execution. |
| `webhook` | Register a webhook called after completion. |
| `run` | Execute the workflow immediately. |

### wallet_mgmt

| Sub-command | Description |
|-------------|-------------|
| `create` | Create a wallet and print the mnemonic. |
| `balance` | Show the SYNN balance for an address. |
| `transfer` | Send SYNN from a mnemonic to a target address. |
### devnet

| Sub-command | Description |
|-------------|-------------|
| `start [nodes]` | Start a local developer network with the given number of nodes. |

### testnet

| Sub-command | Description |
|-------------|-------------|
| `start <config.yaml>` | Launch a testnet using the node definitions in the YAML file. |
### faucet

| Sub-command | Description |
|-------------|-------------|
| `request <addr>` | Request faucet funds for an address. |
| `balance` | Display remaining faucet balance. |
| `config --amount <n> --cooldown <d>` | Update faucet parameters. |
