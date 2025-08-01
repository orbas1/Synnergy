#!/usr/bin/env bash
set -euo pipefail
ID=${1:?"proposal"}
APPROVE=${2:-true}
./synnergy ~gov vote "$ID" --approve="$APPROVE"
