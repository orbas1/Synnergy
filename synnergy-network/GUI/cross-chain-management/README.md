# Cross-Chain Management GUI

This interface provides a dashboard for managing bridge configurations and testing cross-chain opcodes. It communicates with the HTTP server under `cmd/xchainserver`.

## Features

- Register new bridge connections between chains
- View all registered bridges
- Authorize or revoke relayer addresses
- Trigger `LockAndMint` and `BurnAndRelease` operations for testing

## Development

1. Start the cross-chain server:
   ```bash
   go run ./cmd/xchainserver
   ```
2. Open `index.html` in your browser.

The frontend uses Bootstrap 5 and vanilla ES modules.
