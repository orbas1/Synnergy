#!/usr/bin/env bash
set -euo pipefail
ADDR=${1:?"address"}
ROLE=${2:-"validator"}
./synnergy auth register "$ADDR" "$ROLE"
