#!/usr/bin/env bash
set -euo pipefail
CREATOR=${1:?"creator"}
RECIPIENT=${2:?"recipient"}
TYPE=${3:-0}
AMOUNT=${4:-100}
DESC=${5:-"loan"}
./synnergy loanpool submit "$CREATOR" "$RECIPIENT" "$TYPE" "$AMOUNT" "$DESC"
