#!/usr/bin/env bash

set -euo pipefail

CLI=./synnergy

usage() {
    echo "Usage: $0 FROM TO AMOUNT" >&2
    echo "Opens a state channel between FROM and TO for AMOUNT tokens" >&2
}

if [[ $# -ne 3 ]]; then
    usage
    exit 1
fi

if [[ ! -x $CLI ]]; then
    echo "Error: synnergy binary not found at $CLI" >&2
    exit 1
fi

FROM=$1
TO=$2
AMT=$3

"$CLI" channel open --from "$FROM" --to "$TO" --amt "$AMT"
