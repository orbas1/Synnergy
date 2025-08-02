#!/usr/bin/env bash
# Build Synnergy binaries for multiple OS/architecture combinations
# and validate the Docker image builds successfully.
set -euo pipefail

# Determine repository root
ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
MODULE_DIR="$ROOT/synnergy-network"
DIST_DIR="$ROOT/dist"

# Target platforms
platforms=(
  "linux/amd64"
  "linux/arm64"
  "darwin/amd64"
  "darwin/arm64"
  "windows/amd64"
)

mkdir -p "$DIST_DIR"

for platform in "${platforms[@]}"; do
  IFS=/ read -r GOOS GOARCH <<<"$platform"
  out_dir="$DIST_DIR/${GOOS}_${GOARCH}"
  mkdir -p "$out_dir"
  bin_name="synnergy"
  if [ "$GOOS" = "windows" ]; then
    bin_name="${bin_name}.exe"
  fi
  echo "Building $GOOS/$GOARCH..."
  cgo=0
  if [ "$GOOS" = "linux" ] && [ "$GOARCH" = "amd64" ]; then
    cgo=1
  fi
  if (
    cd "$MODULE_DIR" &&
    CGO_ENABLED=$cgo GOOS="$GOOS" GOARCH="$GOARCH" \
      go build -trimpath -o "$out_dir/$bin_name" ./cmd/synnergy
  ); then
    echo "Built $out_dir/$bin_name"
  else
    echo "Skipping $GOOS/$GOARCH (build failed)"
  fi

done

# Build and validate Docker image when docker is available
if command -v docker >/dev/null 2>&1; then
  cd "$ROOT"
  echo "Building Docker image synnergy:latest..."
  docker build -t synnergy:latest .
  docker image inspect synnergy:latest >/dev/null
  echo "Docker image synnergy:latest built successfully"
else
  echo "Docker not found; skipping Docker image build"
fi
