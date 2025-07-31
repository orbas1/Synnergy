# Synnergy CLI Setup

This document describes how to build the `synnergy` command line interface and
run a small local network for development.

## Requirements

- Go 1.20 or newer
- Git
- Optional: Docker for running auxiliary services

## Building the CLI

```bash
git clone <repo-url>
cd synnergy-network
# fetch Go modules
go mod tidy
# build the binary
go build ./cmd/synnergy
```

The resulting executable `synnergy` will be placed in the current directory.
You can move it anywhere in your `$PATH` for convenience.

## Environment Variables

Many commands rely on services running in the background. The most common
environment variables are:

- `LEDGER_PATH` – path to the ledger database
- `P2P_LISTEN_ADDR` – multiaddr for the libp2p node
- `P2P_BOOTSTRAP` – comma‑separated list of bootstrap peers
- `KEYSTORE_PATH` – directory containing node and validator keys
- `SECURITY_API_ADDR` – host:port of the security daemon

Create a `.env` file with these values or export them before running commands.

## Starting a Local Node

1. Initialise the ledger:
   ```bash
   synnergy ledger init --path ./ledger.db
   ```
2. Start the network stack:
   ```bash
   synnergy network start
   ```
3. In a new terminal, create a wallet and fund it:
   ```bash
   synnergy wallet create --out wallet.json
   synnergy coin mint $(cat wallet.json | jq -r .address) 1000
   ```

From here you can experiment with the various command groups such as
`contracts` for deploying Wasm code or `tokens` for managing ERC‑20 assets.

Consult `cmd/cli/cli_guide.md` for a description of every available command.
