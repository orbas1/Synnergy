#!/usr/bin/env bash
# Launch a local multi-node development network.
# Each node runs with in-memory defaults and listens on sequential ports.
set -euo pipefail

NODES=${1:-3}
if ! [[ $NODES =~ ^[0-9]+$ ]] || [ "$NODES" -le 0 ]; then
  echo "usage: $0 [num_nodes]" >&2
  exit 1
fi

BIN_DIR="$(dirname "$0")/../synnergy-network"
cd "$BIN_DIR"
GOFLAGS="-trimpath" go build -o synnergy ./cmd/synnergy

pids=()
for i in $(seq 0 $((NODES-1))); do
  port=$((4101 + i))
  SYNN_ENV="" ./synnergy devnet start $port &
  pids+=("$!")
  echo "started node $i on port $port"
done

trap 'kill ${pids[*]}; exit 0' INT TERM
wait
