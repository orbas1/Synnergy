#!/usr/bin/env bash
# check_circular_imports.sh - detects Go import cycles within the module
set -euo pipefail

REPO_ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
MOD_ROOT="$REPO_ROOT/synnergy-network"

cd "$MOD_ROOT"

# Capture output of go list to analyse potential import cycles
if OUTPUT=$(go list ./... 2>&1); then
    echo "$OUTPUT" >/dev/null
    echo "No circular imports detected"
else
    echo "$OUTPUT"
    if echo "$OUTPUT" | grep -qi 'import cycle'; then
        echo "Circular imports detected"
    fi
    exit 1
fi
