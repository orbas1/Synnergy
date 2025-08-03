# Synnergy GUI User Manual

The graphical interface allows users to manage nodes, wallets and smart
contracts without the command line.

## Prerequisites

- Node.js 18 or newer
- Yarn or npm
- A running Synnergy node

## Installation and Startup

```bash
cd synnergy-network/GUI
npm install
npm start
```

The development server launches at `http://localhost:3000` and proxies requests
to the local node.

## Dashboard Overview

The landing page displays node status, peer connections and recent blocks. Use
the navigation menu to access other modules.

## Wallet Management

- **Create Wallet** – Generate a new key pair and save the mnemonic.
- **View Balances** – Check token holdings and transaction history.
- **Send Tokens** – Transfer SYN assets to another address.

## Smart Contracts

Deploy and interact with contracts through the Contracts section:

1. Upload a Wasm artifact.
2. Specify constructor parameters.
3. Invoke functions and view emitted events.

## Settings and Diagnostics

Configure network endpoints, toggle dark mode and view logs from the Settings
panel.

## Troubleshooting

If the GUI fails to connect to the node:

1. Ensure the node is running and listening on the expected port.
2. Check browser console logs for CORS or network errors.
3. Restart both the node and the GUI.

## Production Build

For optimized static assets suitable for deployment:

```bash
npm run build
```

The compiled files are emitted to `build/` and can be served by any web server.

## Configuration

Adjust the default node endpoint in `src/config.ts`. Environment variables
prefixed with `REACT_APP_` override configuration at runtime.

## Security

The GUI never uploads private keys. Verify the browser address bar before
entering sensitive information and prefer hardware wallets for high-value
accounts.

## Support

Report issues via the project issue tracker and include browser version,
operating system and reproduction steps.
