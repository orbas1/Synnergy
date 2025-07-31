#!/usr/bin/env bash
set -euo pipefail
FILE=${1:?"file"}
./synnergy storage pin "$FILE"
