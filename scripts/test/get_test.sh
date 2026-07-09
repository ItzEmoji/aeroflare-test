#!/usr/bin/env bash
set -uo pipefail
HERE=$(cd "$(dirname "$0")" && pwd)
REPO_ROOT=$(cd "$HERE/../.." && pwd)
# shellcheck source=scripts/test/assert.sh
source "$HERE/assert.sh"

echo "get_test.sh"

ORIG_PATH=$PATH
TOOLS=(bash cat chmod cp dirname env grep jq ln mkdir rm sed tar tr uname zstd)

setup() {
  PATH=$ORIG_PATH
  WORK=$(mktemp -d)
  mkdir -p "$WORK/bin" "$WORK/assets" "$WORK/out"

  local t
  for t in "${TOOLS[@]}"; do
    ln -s "$(command -v "$t")" "$WORK/bin/$t"
  done

  printf '#!/bin/sh\necho stub "$@"\n' > "$WORK/assets/aeroflare"
  chmod +x "$WORK/assets/aeroflare"
  tar --zstd -cf "$WORK/assets/aeroflare-x86_64.tar.zst" -C "$WORK/assets" aeroflare

  cat > "$WORK/bin/gh" <<'EOF'
#!/bin/sh
case "$1 $2" in
  "release download")
    [ "${FAKE_GH_DOWNLOAD_FAIL:-0}" = 1 ] && exit 1
    dest=
    while [ $# -gt 0 ]; do
      case "$1" in
        --dir|-D) dest=$2; shift ;;
      esac
      shift
    done
    [ -n "$dest" ] || exit 1
    cp "$FAKE_ASSETS/aeroflare-x86_64.tar.zst" "$dest/" || exit 1
    exit 0 ;;
  "attestation verify")
    [ "${FAKE_GH_VERIFY_FAIL:-0}" = 1 ] && exit 1
    exit 0 ;;
esac
exit 0
EOF
  chmod +x "$WORK/bin/gh"

  export FAKE_ASSETS="$WORK/assets"
  export AEROFLARE_BIN_DIR="$WORK/out"
  export GH_TOKEN=fake
  unset FAKE_GH_DOWNLOAD_FAIL FAKE_GH_VERIFY_FAIL
  export PATH="$WORK/bin"
}
teardown() { PATH=$ORIG_PATH; rm -rf "$WORK"; }

run_get() { bash "$REPO_ROOT/scripts/get.sh" "$@" 2>&1; }

# --- happy path -------------------------------------------------------------
setup
out=$(run_get); rc=$?
assert_eq "get.sh defaults to aeroflare" "0" "$rc"
assert_eq "binary is executable" "yes" \
  "$([ -x "$AEROFLARE_BIN_DIR/aeroflare" ] && echo yes || echo no)"
assert_contains "reports what it installed" "aeroflare" "$out"
teardown

# --- argument validation ----------------------------------------------------
setup
out=$(run_get bogus); rc=$?
assert_eq "unknown binary exits 1" "1" "$rc"
assert_contains "unknown binary is named" "unknown binary 'bogus'" "$out"
assert_eq "unknown binary downloads nothing" "no" \
  "$([ -e "$AEROFLARE_BIN_DIR/bogus" ] && echo yes || echo no)"
teardown

# --- failure modes ----------------------------------------------------------
setup; export FAKE_GH_DOWNLOAD_FAIL=1
out=$(run_get); rc=$?
assert_eq "missing asset exits 1" "1" "$rc"
assert_contains "missing asset names the release" "ships no" "$out"
assert_eq "missing asset does not annotate" "no" \
  "$(case "$out" in *"::error::"*) echo yes ;; *) echo no ;; esac)"
teardown

setup; export FAKE_GH_VERIFY_FAIL=1
out=$(run_get); rc=$?
assert_eq "failed attestation exits 1" "1" "$rc"
assert_contains "failed attestation is explained" "provenance" "$out"
assert_eq "no binary left behind on verify failure" "no" \
  "$([ -e "$AEROFLARE_BIN_DIR/aeroflare" ] && echo yes || echo no)"
teardown

report
