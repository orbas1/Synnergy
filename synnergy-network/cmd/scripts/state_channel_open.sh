#!/usr/bin/env bash
set -euo pipefail
FROM=${1:-"addr1"}
TO=${2:-"addr2"}
AMT=${3:-1}
./synnergy channel open --from "$FROM" --to "$TO" --amt "$AMT"
