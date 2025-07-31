#!/usr/bin/env bash
set -euo pipefail
WASM=${1:?"wasm file"}
./synnergy contracts deploy --wasm "$WASM"
