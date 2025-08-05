#!/usr/bin/env bash

set -euo pipefail

CLI=./synnergy

usage() {
    echo "Usage: $0 FILE [PROVIDER [PRICE [CAPACITY]]]" >&2
    echo "Pins FILE to storage and creates a marketplace listing" >&2
}

if [[ $# -lt 1 ]]; then
    usage
    exit 1
fi

FILE=$1
PROVIDER=${2:-"0000000000000000000000000000000000000000"}
PRICE=${3:-1}
CAPACITY=${4:-1}

if [[ ! -x $CLI ]]; then
    echo "Error: synnergy binary not found at $CLI" >&2
    exit 1
fi

if [[ ! -f $FILE ]]; then
    echo "Error: file '$FILE' not found" >&2
    exit 1
fi

"$CLI" storage pin --file "$FILE" --payer "$PROVIDER"
"$CLI" storage listing:create --provider "$PROVIDER" --price "$PRICE" --capacity "$CAPACITY"
