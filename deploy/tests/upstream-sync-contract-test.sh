#!/usr/bin/env bash
set -eu

ROOT_DIR="$(CDPATH= cd -- "$(dirname -- "$0")/../.." && pwd)"
VERSION_FILE="$ROOT_DIR/backend/cmd/server/VERSION"
UPSTREAM_VERSION_FILE="$ROOT_DIR/backend/cmd/server/UPSTREAM_VERSION"
VALIDATE_VERSION="$ROOT_DIR/backend/scripts/validate-version.sh"
LAYOUT_FILE="$ROOT_DIR/frontend/src/components/layout/AppLayout.vue"
FRONTEND_SRC="$ROOT_DIR/frontend/src"
HOME_VIEW="$ROOT_DIR/frontend/src/views/HomeView.vue"
HOME_NAVIGATION="$ROOT_DIR/frontend/src/components/home/HomepageNavigation.vue"
HOME_HERO="$ROOT_DIR/frontend/src/components/home/HomepageHero.vue"
HOME_FINAL_CTA="$ROOT_DIR/frontend/src/components/home/HomepageFinalCta.vue"
THEME_FILE="$ROOT_DIR/frontend/src/styles/theme.css"
FRONTEND_ENTRY="$ROOT_DIR/frontend/src/main.ts"

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
[ "$VERSION" = "0.1.177-1" ] \
  || fail "expected VERSION 0.1.177-1, found $VERSION"
[ "$UPSTREAM_VERSION" = "0.1.177" ] \
  || fail "expected UPSTREAM_VERSION 0.1.177, found $UPSTREAM_VERSION"

[ -x "$VALIDATE_VERSION" ] || fail "version validator is not executable: $VALIDATE_VERSION"
"$VALIDATE_VERSION" "v0.1.177-1" >/dev/null 2>&1 \
  || fail "validate-version.sh rejected exact tag v0.1.177-1"

for INVALID_TAG in \
  "0.1.177-1" \
  "refs/tags/v0.1.177-1" \
  "v0.1.177" \
  "v0.1.177-2"
do
  if "$VALIDATE_VERSION" "$INVALID_TAG" >/dev/null 2>&1; then
    fail "validate-version.sh accepted non-exact tag $INVALID_TAG"
  fi
done

grep -Fq 'bg-mesh-gradient' "$LAYOUT_FILE" \
  || fail "official v0.1.177 layout marker is missing"
grep -Fq 'lg:ml-64' "$LAYOUT_FILE" \
  || fail "official v0.1.177 sidebar width is missing"

for MARKER in workbench-shell workbench-content workbench-header workbench-sidebar; do
  if grep -R -n -F -- "$MARKER" "$FRONTEND_SRC" >&2; then
    fail "rejected marker $MARKER remains in frontend/src"
  fi
done

WORKBENCH_TEST="$(find "$FRONTEND_SRC" -type f -name 'workbenchVisualContract.spec.ts' -print -quit)"
if [ -n "$WORKBENCH_TEST" ]; then
  fail "rejected workbench visual contract test remains: $WORKBENCH_TEST"
fi

grep -Fq '@/components/home/AIWeLinkHome.vue' "$HOME_VIEW" \
  || fail "HomeView does not use @/components/home/AIWeLinkHome.vue"

for HOME_ENTRY in "$HOME_NAVIGATION" "$HOME_HERO" "$HOME_FINAL_CTA"; do
  grep -Fq ": '/login'" "$HOME_ENTRY" \
    || fail "homepage entry does not target /login: $HOME_ENTRY"
  if grep -Fq ": '/register'" "$HOME_ENTRY"; then
    fail "homepage entry still targets /register: $HOME_ENTRY"
  fi
done

grep -Fq "t('home.login')" "$HOME_NAVIGATION" \
  || fail "homepage navigation does not use the login label"

[ -f "$THEME_FILE" ] || fail "theme file is missing: $THEME_FILE"
grep -Fq -- '--color-primary-500: 210 31 75;' "$THEME_FILE" \
  || fail "theme.css is missing AIWeLink primary color #D21F4B"
grep -Fq -- '--color-primary-500: 255 194 71;' "$THEME_FILE" \
  || fail "theme.css is missing AIWeLink accent color #FFC247"

grep -Fq './styles/theme.css' "$FRONTEND_ENTRY" \
  || fail "frontend entry does not import ./styles/theme.css"

printf 'upstream sync contract passed: AIWeLink %s (Sub2API %s)\n' \
  "$VERSION" "$UPSTREAM_VERSION"
