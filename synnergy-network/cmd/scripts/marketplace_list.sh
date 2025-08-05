#!/usr/bin/env bash
# List an AI model on the marketplace.
set -euo pipefail

PRICE=${1:?"price"}
CID=${2:?"cid"}

# Run relative to this script's directory and build the CLI if missing.
SCRIPT_DIR="$(cd -- "$(dirname -- "${BASH_SOURCE[0]}")" &>/dev/null && pwd)"
cd "$SCRIPT_DIR"

if [[ ! -x synnergy ]]; then
    GOFLAGS="-trimpath" go build -o synnergy ../synnergy
fi

./synnergy ai list "$PRICE" "$CID"
