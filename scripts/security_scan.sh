#!/usr/bin/env bash
# Run Go security static analysis using gosec.
set -euo pipefail

if ! command -v gosec >/dev/null 2>&1; then
  echo "Installing gosec..." >&2
  go install github.com/securego/gosec/v2/cmd/gosec@latest
fi
GOPATH_BIN="$(go env GOPATH)/bin"
export PATH="$GOPATH_BIN:$PATH"

echo "Running gosec security scanner"
gosec -exclude-dir=third_party -severity high -confidence high ./...
