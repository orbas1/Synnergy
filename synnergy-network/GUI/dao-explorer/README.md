# DAO Explorer GUI

An interface for browsing and interacting with DAO proposals on the Synnergy Network. The backend uses Express with simple services that would normally interface with Synnergy smart contracts.

## Development

1. Install dependencies in the `backend` folder:
   ```bash
   npm install express body-parser cors dotenv
   ```
2. Start the API server:
   ```bash
   node server.js
   ```
3. Open `index.html` in a browser. The frontend communicates with the API under `/backend/api`.

This example stores proposals in memory but demonstrates how opcodes from `opcode_dispatcher.go` could be wired into smart contracts and exposed through a REST API.
