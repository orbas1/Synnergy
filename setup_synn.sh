#!/usr/bin/env bash
# Setup script for Synnergy network
set -euo pipefail

sudo apt-get update
sudo apt-get install -y build-essential curl git golang-go

export GO111MODULE=on
cd synnergy-network

go mod download
GOFLAGS="-trimpath" go build -o synnergy ./cmd/synnergy

echo "Synnergy CLI built successfully."
