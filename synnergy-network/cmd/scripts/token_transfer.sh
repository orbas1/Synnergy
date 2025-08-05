#!/usr/bin/env bash

set -euo pipefail

CLI=./synnergy

usage() {
    echo "Usage: $0 TOKEN FROM TO AMOUNT" >&2
    echo "Transfers AMOUNT of TOKEN from FROM to TO" >&2
}

if [[ $# -ne 4 ]]; then
    usage
    exit 1
fi

if [[ ! -x $CLI ]]; then
    echo "Error: synnergy binary not found at $CLI" >&2
    exit 1
fi

TOK=$1
FROM=$2
TO=$3
AMT=$4

"$CLI" tokens transfer "$TOK" --from "$FROM" --to "$TO" --amt "$AMT"
