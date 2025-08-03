#!/usr/bin/env bash
# Submit an optimistic rollup batch from a file.
set -euo pipefail

BATCH=${1:?"batch.json"}

# Always run relative to this script's directory.
SCRIPT_DIR="$(cd -- "$(dirname -- "${BASH_SOURCE[0]}")" &>/dev/null && pwd)"
cd "$SCRIPT_DIR"

# Build the CLI if it doesn't exist.
if [[ ! -x synnergy ]]; then
    GOFLAGS="-trimpath" go build -o synnergy ../synnergy
fi

./synnergy '~rollup' submit --file "$BATCH"
