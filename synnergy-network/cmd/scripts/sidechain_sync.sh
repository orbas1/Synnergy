#!/usr/bin/env bash
# List registered side-chains.
set -euo pipefail

# Run relative to this script's directory.
SCRIPT_DIR="$(cd -- "$(dirname -- "${BASH_SOURCE[0]}")" &>/dev/null && pwd)"
cd "$SCRIPT_DIR"

# Build the CLI if needed.
if [[ ! -x synnergy ]]; then
    GOFLAGS="-trimpath" go build -o synnergy ../synnergy
fi

./synnergy '~sc' list
