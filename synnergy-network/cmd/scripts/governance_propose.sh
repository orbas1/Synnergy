#!/usr/bin/env bash
# Enterprise governance proposal creation
set -euo pipefail
TITLE=${1:-"Upgrade"}
BODY=${2:-proposal.md}
./synnergy ~gov propose --title "$TITLE" --body "$BODY"
