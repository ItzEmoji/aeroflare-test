#!/usr/bin/env bash
# Download, provenance-verify, and extract a prebuilt aeroflare release binary
# into ./bin. The local counterpart to scripts/install.sh, which does the same
# for the composite action.
#
# Usage: scripts/get.sh [binary]   (default: aeroflare)
set -euo pipefail
REPO_ROOT=$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)
# shellcheck source=scripts/lib.sh
source "$REPO_ROOT/scripts/lib.sh"

# lib.sh's die emits a GitHub Actions annotation. This script runs in a terminal,
# so replace it. Safe because sourcing lib.sh has no side effects.
die() {
  printf 'error: %s\n' "$1" >&2
  exit 1
}

repo=${AEROFLARE_REPO:-ItzEmoji/aeroflare}
bin=${1:-aeroflare}

case "$bin" in
  aeroflare|aeroflare-ci) ;;
  *) die "unknown binary '$bin'; expected aeroflare or aeroflare-ci" ;;
esac

command -v gh >/dev/null 2>&1 \
  || die "gh not found on PATH; install the GitHub CLI (https://cli.github.com)"

version=$(read_version "$REPO_ROOT")
arch=$(host_arch_label "$(uname -m)")
dest=${AEROFLARE_BIN_DIR:-$REPO_ROOT/bin}

fetch_release_binary "$repo" "$version" "$bin" "$arch" "$dest" \
  "check available assets with: gh release list --repo $repo"

printf '%s %s (%s) verified and installed to %s\n' "$bin" "$version" "$arch" "$dest/$bin"
