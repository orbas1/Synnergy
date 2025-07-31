#!/usr/bin/env bash
set -euo pipefail
OUT=${1:-wallet.json}
PASS=${2:-secret}
./synnergy wallet create --out "$OUT" --password "$PASS"
