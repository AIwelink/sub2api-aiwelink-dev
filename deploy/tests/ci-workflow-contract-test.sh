#!/usr/bin/env bash
set -euo pipefail

ROOT_DIR=$(cd "$(dirname "${BASH_SOURCE[0]}")/../.." && pwd)
CI="$ROOT_DIR/.github/workflows/backend-ci.yml"
SECURITY="$ROOT_DIR/.github/workflows/security-scan.yml"
TRIVY_IGNORE="$ROOT_DIR/.trivyignore.yaml"
BACKEND_DOCKERFILE="$ROOT_DIR/backend/Dockerfile"
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

assert_count() {
  local file=$1 text=$2 expected=$3 actual
  actual=$(grep -Fc -- "$text" "$file" || true)
  [ "$actual" -eq "$expected" ] || {
    printf 'expected %s occurrences of %s in %s, found %s\n' "$expected" "$text" "$file" "$actual" >&2
    exit 1
  }
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
assert_contains "$CI" 'trivyignores: .trivyignore.yaml'
assert_contains "$SECURITY" 'trivyignores: .trivyignore.yaml'
assert_count "$TRIVY_IGNORE" 'CVE-2026-34040' 1
assert_count "$TRIVY_IGNORE" 'CVE-2023-30533' 1
assert_count "$TRIVY_IGNORE" 'CVE-2024-22363' 1
assert_count "$TRIVY_IGNORE" 'frontend/pnpm-lock.yaml' 2
assert_count "$TRIVY_IGNORE" 'backend/go.mod' 1
assert_count "$TRIVY_IGNORE" '- "Dockerfile"' 1
assert_count "$TRIVY_IGNORE" 'Dockerfile.goreleaser' 1
assert_count "$TRIVY_IGNORE" 'deploy/Dockerfile' 1
assert_count "$TRIVY_IGNORE" 'AVD-DS-0002' 1
assert_count "$TRIVY_IGNORE" 'expired_at: 2026-09-11' 1
assert_count "$TRIVY_IGNORE" 'expired_at: 2026-10-06' 2
assert_count "$TRIVY_IGNORE" 'expired_at: 2026-11-11' 1
assert_not_contains "$TRIVY_IGNORE" 'backend/Dockerfile'
assert_contains "$BACKEND_DOCKERFILE" 'USER sub2api'
assert_contains "$RELEASE" 'Verify release commit belongs to main'
assert_contains "$RELEASE" 'Verify successful ci-gate for release commit'
assert_contains "$RELEASE" 'Scan AIWeLink release image'
assert_not_contains "$RELEASE" 'docker.aiwelink.cc/sub2api-aiwelink-dev:latest'
assert_not_contains "$RELEASE" "node-version: '20'"
assert_contains "$RELEASE" "node-version: '24'"
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
