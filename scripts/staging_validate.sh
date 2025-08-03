#!/usr/bin/env bash
# Deploy a staging network and execute a sample transaction to validate the release candidate.
set -euo pipefail

CONF=${1:-cmd/config/staging.yaml}
BIN_DIR="$(dirname "$0")/../synnergy-network"
LOG_DIR="$(mktemp -d)"

cd "$BIN_DIR"
GOFLAGS="-trimpath" go build -o synnergy ./cmd/synnergy

SYNN_ENV=staging ./synnergy testnet start "$CONF" >"$LOG_DIR/network.log" 2>&1 &
NET_PID=$!

cleanup() {
  kill "$NET_PID" 2>/dev/null || true
  rm -rf "$LOG_DIR"
}
trap cleanup EXIT INT TERM

# Allow network time to start
sleep 5

# Execute a token transfer transaction and validate output
TX_OUT=$(./cmd/scripts/token_transfer.sh "SYNN" "addr1" "addr2" 1 2>&1)
echo "$TX_OUT" >"$LOG_DIR/token_transfer.log"
if ! grep -q "transfer SYNN" "$LOG_DIR/token_transfer.log"; then
  echo "token transfer failed" >&2
  cat "$LOG_DIR/token_transfer.log" >&2
  exit 1
fi

wait "$NET_PID"
echo "staging validation successful"
