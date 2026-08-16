#!/usr/bin/env bash
set -eu

ROOT_DIR="$(CDPATH= cd -- "$(dirname -- "$0")/../.." && pwd)"
VERSION_FILE="$ROOT_DIR/backend/cmd/server/VERSION"
UPSTREAM_VERSION_FILE="$ROOT_DIR/backend/cmd/server/UPSTREAM_VERSION"
VALIDATE_VERSION="$ROOT_DIR/backend/scripts/validate-version.sh"
LAYOUT_FILE="$ROOT_DIR/frontend/src/components/layout/AppLayout.vue"
FRONTEND_SRC="$ROOT_DIR/frontend/src"

fail() {
  printf 'upstream sync contract failed: %s\n' "$1" >&2
  exit 1
}

assert_single_line() {
  file="$1"
  label="$2"
  [ -f "$file" ] || fail "$label file is missing: $file"
  awk 'END { exit (NR == 1 ? 0 : 1) }' "$file" \
    || fail "$label must contain exactly one line"
}

assert_single_line "$VERSION_FILE" "VERSION"
assert_single_line "$UPSTREAM_VERSION_FILE" "UPSTREAM_VERSION"

VERSION="$(tr -d '\r' < "$VERSION_FILE")"
UPSTREAM_VERSION="$(tr -d '\r' < "$UPSTREAM_VERSION_FILE")"
[ "$VERSION" = "0.1.176-1" ] \
  || fail "expected VERSION 0.1.176-1, found $VERSION"
[ "$UPSTREAM_VERSION" = "0.1.176" ] \
  || fail "expected UPSTREAM_VERSION 0.1.176, found $UPSTREAM_VERSION"

[ -x "$VALIDATE_VERSION" ] || fail "version validator is not executable: $VALIDATE_VERSION"
"$VALIDATE_VERSION" "v0.1.176-1" >/dev/null 2>&1 \
  || fail "validate-version.sh rejected exact tag v0.1.176-1"

for INVALID_TAG in \
  "0.1.176-1" \
  "refs/tags/v0.1.176-1" \
  "v0.1.176" \
  "v0.1.176-2"
do
  if "$VALIDATE_VERSION" "$INVALID_TAG" >/dev/null 2>&1; then
    fail "validate-version.sh accepted non-exact tag $INVALID_TAG"
  fi
done

grep -Fq 'bg-mesh-gradient' "$LAYOUT_FILE" \
  || fail "official v0.1.176 layout marker is missing"
grep -Fq 'lg:ml-64' "$LAYOUT_FILE" \
  || fail "official v0.1.176 sidebar width is missing"

for MARKER in workbench-shell workbench-content workbench-header; do
  if grep -R -n -F -- "$MARKER" "$FRONTEND_SRC" >&2; then
    fail "rejected marker $MARKER remains in frontend/src"
  fi
done

WORKBENCH_TEST="$(find "$FRONTEND_SRC" -type f -name 'workbenchVisualContract.spec.ts' -print -quit)"
if [ -n "$WORKBENCH_TEST" ]; then
  fail "rejected workbench visual contract test remains: $WORKBENCH_TEST"
fi

printf 'upstream sync contract passed: AIWeLink %s (Sub2API %s)\n' \
  "$VERSION" "$UPSTREAM_VERSION"
