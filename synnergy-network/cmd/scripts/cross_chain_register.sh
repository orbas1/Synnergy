#!/usr/bin/env bash
# Enterprise-grade cross-chain bridge registration
set -euo pipefail
SRC=${1:-"chainA"}
DST=${2:-"chainB"}
RELAYER=${3:-"0xrelayer"}
./synnergy xchain register "$SRC" "$DST" "$RELAYER"
