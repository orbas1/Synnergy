#!/usr/bin/env bash
set -euo pipefail
ADDR=${1:?"address"}
HOST=${FAUCET_HOST:-"http://localhost:8080"}
curl -s -X POST -H "Content-Type: application/json" -d "{\"address\":\"$ADDR\"}" "$HOST/fund"
