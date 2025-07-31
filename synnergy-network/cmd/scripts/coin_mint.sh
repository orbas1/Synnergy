#!/usr/bin/env bash
set -euo pipefail
ADDR=${1:-"0x0"}
AMT=${2:-1}
./synnergy coin mint "$ADDR" "$AMT"
