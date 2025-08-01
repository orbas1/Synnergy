# AI Marketplace GUI

Allows users to buy and sell AI services through a small Express backend.
The server invokes the `AIServiceMarketplace` smart contract using the Synnergy
CLI rather than Ethereum tooling.

## Usage

Set the following environment variables or provide a `.env` file:

- `CLI_PATH` – path to the built `synnergy` binary (defaults to `synnergy`)
- `MARKETPLACE_ADDRESS` – deployed contract address

Install dependencies and start the service:

```bash
npm install express dotenv
node server.js
```
