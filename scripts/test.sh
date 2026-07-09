#!/usr/bin/env bash
# Run the Go tests and the shell script tests.
set -euo pipefail
REPO_ROOT=$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)
cd "$REPO_ROOT"

echo "==> go test"
go test ./...

echo "==> shell tests"
bash scripts/test/all.sh
