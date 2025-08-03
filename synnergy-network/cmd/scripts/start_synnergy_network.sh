#!/usr/bin/env bash
# Start essential Synnergy daemons using the CLI. Requires Go.
set -euo pipefail

# Run relative to this script's directory.
SCRIPT_DIR="$(cd -- "$(dirname -- "${BASH_SOURCE[0]}")" &>/dev/null && pwd)"
cd "$SCRIPT_DIR"

# Build the CLI binary
GOFLAGS="-trimpath" go build -o synnergy ../synnergy

# Launch background services
./synnergy network start &
NET_PID=$!

./synnergy consensus start &
CONS_PID=$!

./synnergy replication start &
REPL_PID=$!

./synnergy vm start &
VM_PID=$!

# Example security action: compute Merkle root for demo data
MERKLE=$(./synnergy '~sec' merkle 68656c6c6f,776f726c64)
echo "Merkle root: $MERKLE"

echo "Synnergy network running. Press Ctrl+C to stop."
trap 'kill $NET_PID $CONS_PID $REPL_PID $VM_PID' INT TERM
wait $NET_PID $CONS_PID $REPL_PID $VM_PID
