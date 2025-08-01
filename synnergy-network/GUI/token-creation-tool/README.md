# Token Creation Tool GUI

This interface allows users to deploy and manage custom tokens on a running Synnergy node. It consists of an Express backend and a Bootstrap 5 frontend.

## Features

- Deploy the `TokenFactory` smart contract with inline assembly opcodes.
- Create new tokens and store metadata in `tokens.json`.
- List existing tokens in a table view.
- Example Solidity contracts located in `cmd/smart_contracts/`.

## Setup

1. Run `npm install` inside `server/` to install dependencies.
2. Configure `server/.env` with `RPC_URL` and `PRIVATE_KEY` for deployment.
3. Start the backend: `node server.js`.
4. Open `index.html` in a browser to use the GUI.

The backend uses `solc` and `ethers` to compile and deploy contracts. It does not interact with a live chain by default; update the environment variables to connect to your node.
