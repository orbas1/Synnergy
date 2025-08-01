# Smart-Contract Marketplace GUI

This interface allows users to deploy and browse smart contracts on the Synnergy network.
It ships a lightweight Express backend that stores deployed contracts and
exposes a JSON API consumed by the HTML frontend.

## Getting Started

1. Install Node.js and run `npm install` in the project root.
2. Configure `.env` to point at your built `synnergy` CLI.
3. Start the backend with `node server/server.js` and open `views/listings.html`.

Deployed contracts are stored under `server/data`. The service calls the Synnergy
CLI to deploy WASM code using opcode `Deploy` and transfers payments with
`Tokens_Transfer`.
4. Visit `detail.html?id=<contractId>` to download the compiled WASM.
