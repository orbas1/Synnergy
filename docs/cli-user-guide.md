# Synnergy CLI User Manual

The Synnergy command line interface exposes tools for managing ledgers, nodes
and smart contracts. This manual covers common workflows and command groups.

## Installation

Build the CLI by running the repository setup script or by compiling manually:

```bash
./setup_synn.sh
# or
cd synnergy-network
go mod tidy
GOFLAGS="-trimpath" go build -o synnergy ./cmd/synnergy
```

The resulting `synnergy` binary resides in `synnergy-network`.

## Initializing a Ledger

```bash
cd synnergy-network
./synnergy ledger init --path ./ledger.db
```

This command creates the ledger database used by other services.

## Starting a Local Node

```bash
./synnergy network start
```

The node runs networking, consensus and VM services.

## Creating Wallets

```bash
./synnergy wallet create --name alice
./synnergy wallet list
```

Wallet commands manage keys and balances.

## Deploying Smart Contracts

```bash
./synnergy contract deploy --wasm path/to/contract.wasm
```

Contract management commands support deployment, invocation and inspection.

## Additional Command Groups

The CLI includes many modules such as:

- `ai` – manage AI models and inference jobs
- `token` – issue and transfer SYN tokens
- `dao` – interact with on-chain governance
- `crosschain` – bridge assets between networks

Run `./synnergy help` to list all available commands and
`./synnergy <group> --help` for detailed usage.

## Configuration File

By default the CLI reads optional settings from `~/.synnergy/config.yaml`:

```yaml
network: mainnet
node:
  rpc: http://localhost:8080
```

Command-line flags override values in the configuration file.

## Security Considerations

- Protect wallet mnemonic and keystore files with filesystem permissions.
- Use the `--offline` flag for air-gapped signing when possible.
- Review downloaded scripts before executing them.

## Troubleshooting

- Run `./synnergy --log-level debug` for verbose output.
- Ensure ledger database paths are writable when initializing.
- Remove `~/.synnergy` to reset local state if commands misbehave.
