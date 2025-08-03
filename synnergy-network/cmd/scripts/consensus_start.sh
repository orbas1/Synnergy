#!/usr/bin/env bash
# Starts the Synnergy consensus engine. Builds the CLI binary if needed.

set -euo pipefail

# Always run relative to this script's directory
SCRIPT_DIR="$(cd -- "$(dirname -- "${BASH_SOURCE[0]}")" &>/dev/null && pwd)"
cd "$SCRIPT_DIR"

# Ensure Go toolchain is available
if ! command -v go >/dev/null 2>&1; then
    echo "go command not found" >&2
    exit 1
fi

# Build the synnergy CLI if it doesn't exist
if [[ ! -x synnergy ]]; then
    GOFLAGS=${GOFLAGS:-} go build -trimpath -o synnergy ../synnergy
fi

./synnergy consensus start "$@"

