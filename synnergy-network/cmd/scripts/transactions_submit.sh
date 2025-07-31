#!/usr/bin/env bash
set -euo pipefail
FILE=${1:?"tx.json"}
./synnergy tx submit --json "$FILE"
