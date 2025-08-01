#!/usr/bin/env bash
# Start a configurable testnet defined by a YAML file.
set -euo pipefail

CONF=${1:-testnet.yaml}
BIN_DIR="$(dirname "$0")/../synnergy-network"
cd "$BIN_DIR"
GOFLAGS="-trimpath" go build -o synnergy ./cmd/synnergy

if [ ! -f "$CONF" ]; then
  echo "config file $CONF not found" >&2
  exit 1
fi

./synnergy testnet start "$CONF" &
PID=$!
trap 'kill $PID' INT TERM
wait $PID
