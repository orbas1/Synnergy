#!/usr/bin/env bash
set -euo pipefail
FILE=${1:?"file"}
PROVIDER=${2:-"0000000000000000000000000000000000000000"}
PRICE=${3:-1}
CAPACITY=${4:-1}
./synnergy storage pin --file "$FILE" --payer "$PROVIDER"
./synnergy storage listing:create --provider "$PROVIDER" --price "$PRICE" --capacity "$CAPACITY"
