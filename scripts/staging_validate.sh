#!/usr/bin/env bash
# Deploy a staging network and execute a sample transaction to validate the release candidate.
set -euo pipefail

CONF=${1:-cmd/config/staging.yaml}
BIN_DIR="$(dirname "$0")/../synnergy-network"

cd "$BIN_DIR"
GOFLAGS="-trimpath" go build -o synnergy ./cmd/synnergy

SYNN_ENV=staging ./synnergy testnet start "$CONF" &
NET_PID=$!

cleanup() {
  kill "$NET_PID" 2>/dev/null || true
}
trap cleanup EXIT INT TERM

# Allow network time to start
sleep 5

# Execute a token transfer transaction as a basic validation
./cmd/scripts/token_transfer.sh "SYNN" "addr1" "addr2" 1

wait "$NET_PID"
