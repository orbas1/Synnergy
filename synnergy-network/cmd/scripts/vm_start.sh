#!/usr/bin/env bash

set -euo pipefail

CLI=./synnergy

if [[ ! -x $CLI ]]; then
    echo "Error: synnergy binary not found at $CLI" >&2
    exit 1
fi

"$CLI" vm start
