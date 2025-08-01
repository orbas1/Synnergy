#!/usr/bin/env bash
set -euo pipefail

cd synnergy-network

./synnergy network start &
NET_PID=$!

./synnergy consensus start &
CONS_PID=$!

./synnergy replication start &
REPL_PID=$!

./synnergy vm start &
VM_PID=$!

echo "Synnergy network running. Press Ctrl+C to stop."
trap 'kill $NET_PID $CONS_PID $REPL_PID $VM_PID' INT TERM
wait $NET_PID $CONS_PID $REPL_PID $VM_PID
