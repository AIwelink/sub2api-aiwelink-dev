#!/usr/bin/env bash
set -euo pipefail

ROOT_DIR=$(cd "$(dirname "${BASH_SOURCE[0]}")/../.." && pwd)
SUMMARY_SCRIPT="$ROOT_DIR/deploy/image-publish-summary.sh"
TMP_DIR=$(mktemp -d)
trap 'rm -rf -- "$TMP_DIR"' EXIT

assert_contains() {
  local file=$1 text=$2
  grep -Fq -- "$text" "$file" || {
    printf 'missing %s in %s\n' "$text" "$file" >&2
    exit 1
  }
}

run_summary() {
  local output=$1
  shift
  env \
    GITHUB_STEP_SUMMARY="$output" \
    IMAGE_NAME=docker.aiwelink.cc/sub2api-aiwelink-dev \
    GITHUB_SHA=0123456789abcdef0123456789abcdef01234567 \
    CHECKOUT_OUTCOME=success \
    BUILDX_OUTCOME=success \
    VERSION_OUTCOME=success \
    META_OUTCOME=success \
    IMAGE_OUTCOME=success \
    BUILD_OUTCOME=success \
    SCAN_OUTCOME=success \
    CREDENTIALS_OUTCOME=success \
    LOGIN_OUTCOME=success \
    "$@" \
    bash "$SUMMARY_SCRIPT"
}

SUCCESS_SUMMARY="$TMP_DIR/success.md"
run_summary "$SUCCESS_SUMMARY" \
  GITHUB_REF_NAME=aiwelink-dev \
  APPLICATION_VERSION=0.1.170-1 \
  UPSTREAM_VERSION=0.1.170 \
  PUSH_OUTCOME=success \
  DIGEST_OUTCOME=success \
  IMAGE_DIGEST=sha256:aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa

assert_contains "$SUCCESS_SUMMARY" '## AIWeLink image publication'
assert_contains "$SUCCESS_SUMMARY" '| Publish status | `success` |'
assert_contains "$SUCCESS_SUMMARY" '| Failure stage | `none` |'
assert_contains "$SUCCESS_SUMMARY" '| Application version | `0.1.170-1` |'
assert_contains "$SUCCESS_SUMMARY" '| Upstream version | `0.1.170` |'
assert_contains "$SUCCESS_SUMMARY" '| Branch | `aiwelink-dev` |'
assert_contains "$SUCCESS_SUMMARY" '| Commit | `0123456789abcdef0123456789abcdef01234567` |'
assert_contains "$SUCCESS_SUMMARY" 'docker.aiwelink.cc/sub2api-aiwelink-dev:dev'
assert_contains "$SUCCESS_SUMMARY" 'docker.aiwelink.cc/sub2api-aiwelink-dev:dev-0123456789ab'
assert_contains "$SUCCESS_SUMMARY" 'sha256:aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa'
assert_contains "$SUCCESS_SUMMARY" 'docker pull docker.aiwelink.cc/sub2api-aiwelink-dev:dev-0123456789ab'
assert_contains "$SUCCESS_SUMMARY" 'docker pull docker.aiwelink.cc/sub2api-aiwelink-dev@sha256:aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa'

FAILURE_SUMMARY="$TMP_DIR/failure.md"
run_summary "$FAILURE_SUMMARY" \
  GITHUB_REF_NAME=main \
  APPLICATION_VERSION=0.1.170-1 \
  UPSTREAM_VERSION=0.1.170 \
  BUILD_OUTCOME=failure \
  SCAN_OUTCOME=skipped \
  CREDENTIALS_OUTCOME=skipped \
  LOGIN_OUTCOME=skipped \
  PUSH_OUTCOME=skipped \
  DIGEST_OUTCOME=skipped

assert_contains "$FAILURE_SUMMARY" '| Publish status | `failure` |'
assert_contains "$FAILURE_SUMMARY" '| Failure stage | `build` |'
assert_contains "$FAILURE_SUMMARY" 'docker.aiwelink.cc/sub2api-aiwelink-dev:latest'
assert_contains "$FAILURE_SUMMARY" 'docker.aiwelink.cc/sub2api-aiwelink-dev:main-0123456789ab'
assert_contains "$FAILURE_SUMMARY" '| Digest | `unavailable` |'
assert_contains "$FAILURE_SUMMARY" 'Image was not published; pull commands are unavailable.'

CHECKOUT_FAILURE_SUMMARY="$TMP_DIR/checkout-failure.md"
run_summary "$CHECKOUT_FAILURE_SUMMARY" \
  GITHUB_REF_NAME=aiwelink-dev \
  CHECKOUT_OUTCOME=failure \
  BUILDX_OUTCOME=skipped \
  VERSION_OUTCOME=skipped \
  META_OUTCOME=skipped \
  IMAGE_OUTCOME=skipped \
  BUILD_OUTCOME=skipped \
  SCAN_OUTCOME=skipped \
  CREDENTIALS_OUTCOME=skipped \
  LOGIN_OUTCOME=skipped \
  PUSH_OUTCOME=skipped \
  DIGEST_OUTCOME=skipped \
  IMAGE_DIGEST=not-a-digest

assert_contains "$CHECKOUT_FAILURE_SUMMARY" '| Publish status | `failure` |'
assert_contains "$CHECKOUT_FAILURE_SUMMARY" '| Failure stage | `checkout` |'
assert_contains "$CHECKOUT_FAILURE_SUMMARY" '| Digest | `unavailable` |'

printf 'Image publish summary tests passed\n'
