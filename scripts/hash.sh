#!/usr/bin/env bash
# Recompute default.nix's vendorHash. Run this after changing go.mod or go.sum.
#
# Nix reports the correct hash only when the specified one is wrong, so set a
# known-wrong value, read `got:` from the failure, and write it back.
set -euo pipefail
REPO_ROOT=$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)
cd "$REPO_ROOT"

FAKE='sha256-AAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAA='
NIX_FLAGS=(--no-link --extra-experimental-features 'nix-command flakes')

die() { printf 'error: %s\n' "$1" >&2; exit 1; }

command -v nix >/dev/null 2>&1 \
  || die "nix not found on PATH; see https://nixos.org/download"

backup=$(mktemp)
cp default.nix "$backup"

# Restore on *any* early exit, not just ERR: `die` calls `exit`, which does not
# fire an ERR trap, and a botched run must never leave the fake hash on disk.
# Idempotent, so the success path can disarm the trap and drop the backup.
restore() {
  if [ -f "$backup" ]; then
    cp "$backup" default.nix
    rm -f "$backup"
  fi
}
trap restore EXIT INT TERM

sed -i "s|vendorHash = \"sha256-[^\"]*\";|vendorHash = \"$FAKE\";|" default.nix
grep -q "$FAKE" default.nix || die "could not find a vendorHash line in default.nix"

echo "==> probing for the real vendorHash"
log=$(nix build .#default "${NIX_FLAGS[@]}" 2>&1 || true)

# `|| true`: grep exits 1 when nix failed for a reason other than a hash
# mismatch. Under `set -o pipefail` that would abort here, past the guard below
# that explains what went wrong.
hash=$(printf '%s\n' "$log" \
  | grep -oE 'got:[[:space:]]+sha256-[A-Za-z0-9+/=]+' \
  | grep -oE 'sha256-[A-Za-z0-9+/=]+' \
  | tail -1 || true)

[ -n "$hash" ] || { printf '%s\n' "$log" >&2; die "nix did not report a hash; see the build log above"; }

sed -i "s|vendorHash = \"$FAKE\";|vendorHash = \"$hash\";|" default.nix

echo "==> verifying $hash"
nix build .#default "${NIX_FLAGS[@]}" || die "build still fails with the recomputed hash"

trap - EXIT INT TERM
rm -f "$backup"
echo "vendorHash = $hash"
