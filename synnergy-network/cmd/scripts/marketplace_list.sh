#!/usr/bin/env bash
set -euo pipefail
PRICE=${1:?"price"}
CID=${2:?"cid"}
./synnergy ai list "$PRICE" "$CID"
