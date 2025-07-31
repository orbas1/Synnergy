#!/usr/bin/env bash
set -euo pipefail
GOFLAGS="-trimpath" go build -o synnergy ../synnergy
