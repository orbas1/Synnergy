#!/usr/bin/env bash
# Synnergy.env.sh - Environment setup script for Synnergy Network
# This script installs dependencies and configures environment variables
# for development with full network capabilities.

set -euo pipefail

# Update package list and install common tools if missing
if command -v apt-get >/dev/null 2>&1; then
    apt-get update -y
    apt-get install -y curl git build-essential wget
fi

# Install Go if not already installed
if ! command -v go >/dev/null 2>&1; then
    GOVERSION="1.21.5"
    wget -q https://go.dev/dl/go${GOVERSION}.linux-amd64.tar.gz
    tar -C /usr/local -xzf go${GOVERSION}.linux-amd64.tar.gz
    rm go${GOVERSION}.linux-amd64.tar.gz
    export PATH="/usr/local/go/bin:$PATH"
fi

# Configure Go environment
export GO111MODULE=on
export GOPATH="$(go env GOPATH)"
export PATH="$GOPATH/bin:$PATH"

# Load environment variables from project .env file if it exists
if [ -f synnergy-network/.env ]; then
    set -o allexport
    source synnergy-network/.env
    set +o allexport
fi

# Download Go module dependencies
(cd synnergy-network && go mod download)

echo "Synnergy environment is ready."
