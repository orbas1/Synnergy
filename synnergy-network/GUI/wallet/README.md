# Wallet GUI

This interface allows users to manage their Synnergy wallets through a small HTTP
service. The backend exposes REST endpoints used by the frontend components to
generate new wallets, import existing mnemonics and sign transactions.

## Development

1. Build and run the wallet server:

```bash
cd ../../walletserver
go build -o walletserver .
./walletserver
```

2. Open `views/index.html` in a browser. The page uses Bootstrap 5 and fetches
from `localhost:8081` by default (configurable via `.env`).

## Endpoints

- `GET /api/wallet/create` – generate a new wallet and mnemonic
- `POST /api/wallet/import` – import an existing mnemonic
- `POST /api/wallet/address` – derive an address from posted wallet data
- `POST /api/wallet/sign` – sign a transaction structure
