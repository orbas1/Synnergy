# Cross-Chain Management GUI

This interface provides a simple dashboard for managing bridge configurations.
It communicates with the cross-chain HTTP server located under `cmd/xchainserver`.

## Features

- Register new bridge connections between chains
- View all registered bridges
- Authorize or revoke relayer addresses (via API)

## Development

1. Start the cross-chain server:
   ```bash
   go run ./cmd/xchainserver
   ```
2. Open `index.html` in your browser.

The frontend uses Bootstrap 5 and vanilla ES modules.
