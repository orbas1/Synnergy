#!/usr/bin/env bash

set -euo pipefail

CLI=./synnergy

usage() {
    echo "Usage: $0 OUT_FILE PASSWORD" >&2
    echo "Creates a new wallet encrypted with PASSWORD at OUT_FILE" >&2
}

if [[ $# -ne 2 ]]; then
    usage
    exit 1
fi

OUT=$1
PASS=$2

if [[ ! -x $CLI ]]; then
    echo "Error: synnergy binary not found at $CLI" >&2
    exit 1
fi

"$CLI" wallet create --out "$OUT" --password "$PASS"
