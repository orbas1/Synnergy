#!/usr/bin/env bash
# Enterprise security audit: compute Merkle root of provided hex leaves
set -euo pipefail
LEAVES=${1:-"deadbeef,baadf00d"}
./synnergy ~sec merkle "$LEAVES"
