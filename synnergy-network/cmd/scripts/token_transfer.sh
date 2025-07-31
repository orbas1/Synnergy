#!/usr/bin/env bash
set -euo pipefail
TOK=${1:-"SYNN"}
FROM=${2:-"addr1"}
TO=${3:-"addr2"}
AMT=${4:-1}
./synnergy tokens transfer "$TOK" --from "$FROM" --to "$TO" --amt "$AMT"
