# Token Creation Tool GUI

This interface allows users to deploy custom tokens on a running Synnergy node.
It consists of a small Express backend exposing a `/api/tokens` endpoint and a
Bootstrap 5 frontend.

## Setup

1. `npm install` inside `server/`
2. `node server.js`
3. Open `index.html` in a browser.

The backend stores created tokens in `tokens.json` as a stand‑in for blockchain
integration. Smart contract examples live in `cmd/smart_contracts/token_creator.sol`.
