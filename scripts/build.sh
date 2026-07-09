#!/usr/bin/env bash
# Build both binaries for the host architecture into ./bin.
set -euo pipefail
REPO_ROOT=$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)
cd "$REPO_ROOT"

mkdir -p bin
for bin in aeroflare aeroflare-ci; do
  case "$bin" in
    aeroflare)    pkg=. ;;
    aeroflare-ci) pkg=./cmd/aeroflare-ci ;;
  esac
  echo "==> $bin"
  CGO_ENABLED=0 go build -trimpath -ldflags="-s -w" -o "bin/$bin" "$pkg"
done
