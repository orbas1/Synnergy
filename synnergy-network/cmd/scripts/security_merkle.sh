#!/usr/bin/env bash
# Enterprise security audit: compute Merkle root of provided hex leaves.
set -euo pipefail

LEAVES=${1:-"deadbeef,baadf00d"}

# Run relative to this script's directory.
SCRIPT_DIR="$(cd -- "$(dirname -- "${BASH_SOURCE[0]}")" &>/dev/null && pwd)"
cd "$SCRIPT_DIR"

# Build the CLI if needed.
if [[ ! -x synnergy ]]; then
    GOFLAGS="-trimpath" go build -o synnergy ../synnergy
fi

./synnergy '~sec' merkle "$LEAVES"
