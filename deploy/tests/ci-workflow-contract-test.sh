#!/usr/bin/env bash
set -euo pipefail

ROOT_DIR=$(cd "$(dirname "${BASH_SOURCE[0]}")/../.." && pwd)
CI="$ROOT_DIR/.github/workflows/backend-ci.yml"
SECURITY="$ROOT_DIR/.github/workflows/security-scan.yml"
RELEASE="$ROOT_DIR/.github/workflows/release.yml"
GROWTH_CANARY="$ROOT_DIR/.github/workflows/growth-public-canary.yml"
GROWTH_CANARY_SCRIPT="$ROOT_DIR/deploy/tests/growth-public-canary.sh"
MAKEFILE="$ROOT_DIR/Makefile"

assert_contains() {
  local file=$1 text=$2
  grep -Fq -- "$text" "$file" || {
    printf 'missing %s in %s\n' "$text" "$file" >&2
    exit 1
  }
}

assert_not_contains() {
  local file=$1 text=$2
  if grep -Fq -- "$text" "$file"; then
    printf 'unexpected %s in %s\n' "$text" "$file" >&2
    exit 1
  fi
}

assert_contains "$CI" 'branches: [aiwelink-dev, main]'
assert_contains "$CI" 'cancel-in-progress: true'
assert_contains "$CI" 'ci-gate:'
assert_contains "$CI" 'publish-image:'
assert_contains "$CI" 'needs: ci-gate'
assert_contains "$CI" "github.event_name == 'push'"
assert_contains "$CI" 'type=raw,value=dev,enable='
assert_contains "$CI" 'type=raw,value=latest,enable='
assert_contains "$CI" 'type=sha,prefix=dev-,format=short,enable='
assert_contains "$CI" 'type=sha,prefix=main-,format=short,enable='
assert_not_contains "$CI" 'type=raw,value=${{ steps.version.outputs.version }}'
assert_not_contains "$SECURITY" "node-version: '20'"
assert_contains "$SECURITY" "node-version: '24'"
assert_contains "$RELEASE" 'Verify release commit belongs to main'
assert_contains "$RELEASE" 'Verify successful ci-gate for release commit'
assert_contains "$RELEASE" 'Scan AIWeLink release image'
assert_not_contains "$RELEASE" 'docker.aiwelink.cc/sub2api-aiwelink-dev:latest'
assert_contains "$GROWTH_CANARY" "cron: '*/30 * * * *'"
assert_contains "$GROWTH_CANARY" 'workflow_dispatch:'
assert_contains "$GROWTH_CANARY" 'contents: read'
assert_contains "$GROWTH_CANARY" 'timeout-minutes: 5'
assert_contains "$GROWTH_CANARY" 'cancel-in-progress: true'
assert_contains "$GROWTH_CANARY" 'GROWTH_CANARY_BASE_URL: https://api.aiwelink.cc'
assert_contains "$GROWTH_CANARY" 'GROWTH_CANARY_REFERRAL_CODE: ${{ secrets.GROWTH_CANARY_REFERRAL_CODE }}'
assert_contains "$GROWTH_CANARY" 'bash deploy/tests/growth-public-canary.sh'
assert_contains "$GROWTH_CANARY_SCRIPT" '--self-test'
assert_contains "$GROWTH_CANARY_SCRIPT" '^[a-hj-km-np-z2-9]{8}$'
assert_contains "$MAKEFILE" 'pnpm --dir frontend run test:run'
assert_contains "$MAKEFILE" 'pnpm --dir frontend run build'
assert_not_contains "$MAKEFILE" 'FRONTEND_CRITICAL_VITEST'
assert_contains "$ROOT_DIR/Dockerfile" 'ARG NODE_IMAGE=node:24-alpine'
test ! -e "$ROOT_DIR/.github/workflows/publish-aiwelink-dev-image.yml"
test ! -e "$ROOT_DIR/.github/workflows/cla.yml"

printf 'CI workflow contract checks passed\n'
