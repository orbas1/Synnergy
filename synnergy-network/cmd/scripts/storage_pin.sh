#!/usr/bin/env bash

set -euo pipefail

CLI=./synnergy

usage() {
    echo "Usage: $0 FILE" >&2
    echo "Pins FILE to the storage network" >&2
}

if [[ $# -lt 1 ]]; then
    usage
    exit 1
fi

FILE=$1

if [[ ! -x $CLI ]]; then
    echo "Error: synnergy binary not found at $CLI" >&2
    exit 1
fi

if [[ ! -f $FILE ]]; then
    echo "Error: file '$FILE' not found" >&2
    exit 1
fi

"$CLI" storage pin "$FILE"
