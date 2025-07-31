# Synnergy Script Guide

The `start_synnergy_network.sh` script builds the `synnergy` CLI and boots the
core daemons using the commands provided by the CLI packages.  It also runs a
sample security command.

```bash
./start_synnergy_network.sh
```

This will:

1. Compile the CLI from `cmd/synnergy/main.go`.
2. Start the networking, consensus, replication and VM services.
3. Run a demo `~sec merkle` command.
4. Wait until the services exit (Ctrl+C to terminate).

Ensure that the Go toolchain is available in your `PATH` before running the
script.

## Additional Example Scripts

The following bash scripts demonstrate how to invoke various CLI modules. Each assumes the `synnergy` binary has been built in this directory using `build_cli.sh`.

- `build_cli.sh` – compile the CLI binary.
- `network_start.sh` – start the networking daemon.
- `network_peers.sh` – list connected peers.
- `consensus_start.sh` – launch the consensus service.
- `replication_status.sh` – query the replication daemon status.
- `vm_start.sh` – run the WASM virtual machine daemon.
- `coin_mint.sh` – mint SYNN coins.
- `token_transfer.sh` – transfer a token between two addresses.
- `contracts_deploy.sh` – deploy a smart contract from a WASM file.
- `wallet_create.sh` – create a new HD wallet file.
- `transactions_submit.sh` – submit a signed transaction JSON blob.
- `security_merkle.sh` – compute a Merkle root for auditing.
- `governance_propose.sh` – create a governance proposal.
- `cross_chain_register.sh` – register a cross‑chain bridge relayer.
- `rollup_submit_batch.sh` – submit an optimistic roll‑up batch.
- `sharding_leader.sh` – query the current shard leader.
- `sidechain_sync.sh` – list registered side‑chains.
- `fault_check.sh` – capture a fault‑tolerance snapshot.
- `state_channel_open.sh` – open a payment channel.
- `storage_pin.sh` – pin a file in the storage subsystem.

