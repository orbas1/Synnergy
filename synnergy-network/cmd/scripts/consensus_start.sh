#!/usr/bin/env bash
# Starts the Synnergy consensus engine. Builds the CLI binary if needed.

set -euo pipefail

# Always run relative to this script's directory
SCRIPT_DIR="$(cd -- "$(dirname -- "${BASH_SOURCE[0]}")" &>/dev/null && pwd)"
cd "$SCRIPT_DIR"

# Build the synnergy CLI if it doesn't exist
if [[ ! -x synnergy ]]; then
    GOFLAGS="-trimpath" go build -o synnergy ../synnergy
fi

./synnergy consensus start

