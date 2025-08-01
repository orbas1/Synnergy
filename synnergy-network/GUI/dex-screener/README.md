# DEX Screener GUI

Shows token charts and provides liquidity pool creation without trading features.
It relies on `cmd/dexserver` which exposes liquidity pool data from the
Synnergy node via `/api/pools`. The JavaScript frontend fetches this endpoint and
renders a simple table of pools using Tailwind CSS.
