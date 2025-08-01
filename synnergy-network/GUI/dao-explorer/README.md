# DAO Explorer GUI

An interface for browsing and interacting with DAO proposals on the Synnergy Network. The backend uses Express with simple services that would normally interface with Synnergy smart contracts.

## Development

1. Install dependencies in the `backend` folder:
   ```bash
   npm install
   ```
2. Start the API server (which also serves the static frontend from `views/`):
   ```bash
   node server.js
   ```
3. Open `views/index.html` in a browser. The frontend communicates with the API under `/backend/api`.

If `CONTRACT_ADDRESS` is set in `.env`, the backend uses `ethers.js` to send
transactions to the `DAOExplorer` contract compiled from Solidity. Without a
configured contract, an in-memory store is used for demo purposes. This setup
demonstrates how opcodes from `opcode_dispatcher.go` can power smart-contract
interactions exposed through a REST API.
