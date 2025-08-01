# dexserver

A small HTTP service exposing AMM liquidity pool data for the DEX Screener GUI.
It loads the node configuration, initialises the ledger and AMM modules and
serves `/api/pools` which returns a JSON list of all liquidity pools.
