#!/usr/bin/env bash
set -euo pipefail

ROOT_DIR=$(cd "$(dirname "${BASH_SOURCE[0]}")/../.." && pwd)
CI="$ROOT_DIR/.github/workflows/backend-ci.yml"
SECURITY="$ROOT_DIR/.github/workflows/security-scan.yml"
TRIVY_IGNORE="$ROOT_DIR/.trivyignore.yaml"
BACKEND_DOCKERFILE="$ROOT_DIR/backend/Dockerfile"
RELEASE="$ROOT_DIR/.github/workflows/release.yml"
VERSION_VALIDATOR="$ROOT_DIR/backend/scripts/validate-version.sh"
VERSION_RESOLVER="$ROOT_DIR/backend/scripts/resolve-version.sh"
GORELEASER_SIMPLE="$ROOT_DIR/.goreleaser.simple.yaml"
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
assert_contains "$CI" 'release-contract:'
assert_contains "$CI" 'make test-ci-contract'
assert_contains "$CI" '      - release-contract'
assert_contains "$CI" "github.event_name == 'push'"
assert_contains "$CI" 'type=raw,value=dev,enable='
assert_contains "$CI" 'type=raw,value=latest,enable='
assert_contains "$CI" 'type=sha,prefix=dev-,format=short,enable='
assert_contains "$CI" 'type=sha,prefix=main-,format=short,enable='
assert_contains "$CI" './backend/scripts/validate-version.sh'
assert_not_contains "$CI" 'if ! [[ "$VERSION" =~'
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
assert_count "$RELEASE" 'ref: ${{ needs.validate-version.outputs.release_sha }}' 2
assert_contains "$RELEASE" 'git rev-parse "refs/tags/$RELEASE_TAG^{}"'
assert_not_contains "$RELEASE" 'ref: ${{ github.event.inputs.tag || github.ref }}'
assert_contains "$RELEASE" "      - 'v*.*.*-*'"
assert_not_contains "$RELEASE" "      - 'v*'"
assert_not_contains "$RELEASE" 'docker.aiwelink.cc/sub2api-aiwelink-dev:latest'
assert_not_contains "$RELEASE" "node-version: '20'"
assert_contains "$RELEASE" "node-version: '24'"
assert_not_contains "$GORELEASER_SIMPLE" 'ghcr.io/{{ .Env.GITHUB_REPO_OWNER_LOWER }}/sub2api:latest'
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

"$VERSION_VALIDATOR" >/dev/null
if "$VERSION_VALIDATOR" "refs/tags/v$(tr -d '\r\n' < "$ROOT_DIR/backend/cmd/server/VERSION")" >/dev/null 2>&1; then
  printf 'fully-qualified release ref unexpectedly passed version validation\n' >&2
  exit 1
fi

repo_version=$(tr -d '\r\n' < "$ROOT_DIR/backend/cmd/server/VERSION")
repo_upstream_version=$(tr -d '\r\n' < "$ROOT_DIR/backend/cmd/server/UPSTREAM_VERSION")
version_fixture=$(mktemp -d)
trap 'rm -rf "$version_fixture"' EXIT
mkdir -p "$version_fixture/backend/scripts" "$version_fixture/backend/cmd/server"
cp "$VERSION_VALIDATOR" "$version_fixture/backend/scripts/validate-version.sh"
cp "$VERSION_RESOLVER" "$version_fixture/backend/scripts/resolve-version.sh"
printf '%s\n' "$repo_version" > "$version_fixture/backend/cmd/server/VERSION"
printf '%s\n' "$repo_upstream_version" > "$version_fixture/backend/cmd/server/UPSTREAM_VERSION"
"$version_fixture/backend/scripts/validate-version.sh" "v$repo_version" >/dev/null
if "$version_fixture/backend/scripts/validate-version.sh" "v$repo_upstream_version" >/dev/null 2>&1; then
  printf 'official upstream tag v%s unexpectedly passed AIWeLink validation\n' "$repo_upstream_version" >&2
  exit 1
fi
git -C "$version_fixture" init -q
git -C "$version_fixture" config user.name test
git -C "$version_fixture" config user.email test@example.com
git -C "$version_fixture" add backend
git -C "$version_fixture" commit -qm fixture
git -C "$version_fixture" tag "v$repo_upstream_version"
[ "$("$version_fixture/backend/scripts/resolve-version.sh")" = "$repo_version" ]
git -C "$version_fixture" tag -d "v$repo_upstream_version" >/dev/null
git -C "$version_fixture" tag "v$repo_version"
[ "$("$version_fixture/backend/scripts/resolve-version.sh")" = "$repo_version" ]

printf 'CI workflow contract checks passed\n'
