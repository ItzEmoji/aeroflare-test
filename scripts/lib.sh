#!/usr/bin/env bash
# Shared helpers for the aeroflare composite action. Sourced, not executed.
# Sourcing must have no side effects.

# die <message> — emit a GitHub error annotation and abort.
die() {
  printf '::error::%s\n' "$1" >&2
  exit 1
}

# arch_label <runner_arch> — map RUNNER_ARCH to the release asset's arch label.
arch_label() {
  case "$1" in
    X64)   printf 'x86_64\n' ;;
    ARM64) printf 'aarch64\n' ;;
    *)     die "unsupported RUNNER_ARCH '$1'; aeroflare ships linux x86_64 and aarch64 only" ;;
  esac
}

# token_env_var <registry> — mirrors TokenEnvVar in internal/ci/cachespec.go.
# ghcr.io -> AEROFLARE_TOKEN_GHCR_IO
token_env_var() {
  local host=$1
  host=${host//./_}
  host=${host//:/_}
  printf 'AEROFLARE_TOKEN_%s\n' "$(printf '%s' "$host" | tr '[:lower:]' '[:upper:]')"
}

# read_version <action_path> — the version release-please maintains, e.g. 1.8.0.
read_version() {
  local manifest=$1/version.json version
  [ -f "$manifest" ] || die "no version.json at $manifest"
  version=$(jq -er '.["."]' "$manifest") || die "version.json has no \".\" key"
  printf '%s\n' "$version"
}

# split_list <string> — split on newlines and commas, trim, drop empties,
# emit one entry per line. Mirrors splitEnvList in cmd/aeroflare-ci/main.go.
split_list() {
  printf '%s' "$1" \
    | tr ',' '\n' \
    | sed -e 's/^[[:space:]]*//' -e 's/[[:space:]]*$//' \
    | grep -v '^$' || true
}

# host_arch_label <machine> — map `uname -m` output to the release asset's arch
# label. The CI counterpart is arch_label, which maps RUNNER_ARCH.
host_arch_label() {
  case "$1" in
    x86_64)        printf 'x86_64\n' ;;
    aarch64|arm64) printf 'aarch64\n' ;;
    *)             die "unsupported architecture '$1'; aeroflare ships linux x86_64 and aarch64 only" ;;
  esac
}

# fetch_release_binary <repo> <version> <bin> <arch> <dest> <missing_hint>
# Download, provenance-verify, and extract a release binary into <dest>.
# The sole implementation of the verified fetch: install.sh and get.sh both
# call it, so attestation verification cannot drift between them.
#
# A verification failure must never fall through to extraction — `die` exits.
fetch_release_binary() {
  local repo=$1 version=$2 bin=$3 arch=$4 dest=$5 missing_hint=$6
  local tag="v$version" archive="$bin-$arch.tar.zst"

  mkdir -p "$dest"

  gh release download "$tag" --repo "$repo" --pattern "$archive" --dir "$dest" \
    || die "release $tag of $repo ships no $archive; $missing_hint"

  gh attestation verify "$dest/$archive" --repo "$repo" \
    || die "provenance verification failed for $archive from $tag"

  tar --zstd -xf "$dest/$archive" -C "$dest" \
    || die "could not extract $archive"
  chmod +x "$dest/$bin"
}
