#!/usr/bin/env bash
set -euo pipefail
BATCH=${1:?"batch.json"}
./synnergy ~rollup submit --file "$BATCH"
