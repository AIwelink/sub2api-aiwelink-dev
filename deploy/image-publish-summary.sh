#!/usr/bin/env bash
set -euo pipefail

: "${GITHUB_STEP_SUMMARY:?GITHUB_STEP_SUMMARY is required}"
: "${IMAGE_NAME:?IMAGE_NAME is required}"
: "${GITHUB_REF_NAME:?GITHUB_REF_NAME is required}"
: "${GITHUB_SHA:?GITHUB_SHA is required}"

ROOT_DIR=$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)
short_sha=${GITHUB_SHA:0:12}

case "$GITHUB_REF_NAME" in
  aiwelink-dev)
    mutable_tag="$IMAGE_NAME:dev"
    immutable_tag="$IMAGE_NAME:dev-$short_sha"
    ;;
  main)
    mutable_tag="$IMAGE_NAME:latest"
    immutable_tag="$IMAGE_NAME:main-$short_sha"
    ;;
  *)
    mutable_tag="$IMAGE_NAME:$GITHUB_REF_NAME"
    immutable_tag="$IMAGE_NAME:$GITHUB_REF_NAME-$short_sha"
    ;;
esac

read_version_file() {
  local file=$1
  if [ -r "$file" ]; then
    tr -d '\r\n' < "$file"
  fi
}

application_version=${APPLICATION_VERSION:-}
if [ -z "$application_version" ]; then
  application_version=$(read_version_file "$ROOT_DIR/backend/cmd/server/VERSION")
fi
application_version=${application_version:-unknown}

upstream_version=${UPSTREAM_VERSION:-}
if [ -z "$upstream_version" ]; then
  upstream_version=$(read_version_file "$ROOT_DIR/backend/cmd/server/UPSTREAM_VERSION")
fi
upstream_version=${upstream_version:-unknown}

escape_markdown() {
  printf '%s' "$1" | tr '\r\n' '  ' | sed 's/|/\\|/g; s/`/\\`/g'
}

step_outcome() {
  local variable=$1
  local value=${!variable:-skipped}
  printf '%s' "$value"
}

publish_status=skipped
failure_stage=none
push_outcome=$(step_outcome PUSH_OUTCOME)

if [ "$push_outcome" = success ]; then
  publish_status=success
else
  steps=(
    "checkout:$(step_outcome CHECKOUT_OUTCOME)"
    "buildx:$(step_outcome BUILDX_OUTCOME)"
    "version:$(step_outcome VERSION_OUTCOME)"
    "metadata:$(step_outcome META_OUTCOME)"
    "image-selection:$(step_outcome IMAGE_OUTCOME)"
    "build:$(step_outcome BUILD_OUTCOME)"
    "scan:$(step_outcome SCAN_OUTCOME)"
    "credentials:$(step_outcome CREDENTIALS_OUTCOME)"
    "login:$(step_outcome LOGIN_OUTCOME)"
    "push:$push_outcome"
  )

  for step in "${steps[@]}"; do
    stage=${step%%:*}
    outcome=${step#*:}
    if [ "$outcome" = failure ] || [ "$outcome" = cancelled ]; then
      publish_status=failure
      failure_stage="$stage"
      break
    fi
  done
fi

digest=${IMAGE_DIGEST:-}
if ! [[ "$digest" =~ ^sha256:[0-9a-f]{64}$ ]]; then
  digest=unavailable
fi

application_version=$(escape_markdown "$application_version")
upstream_version=$(escape_markdown "$upstream_version")
branch=$(escape_markdown "$GITHUB_REF_NAME")
commit=$(escape_markdown "$GITHUB_SHA")
mutable_tag=$(escape_markdown "$mutable_tag")
immutable_tag=$(escape_markdown "$immutable_tag")
digest=$(escape_markdown "$digest")
failure_stage=$(escape_markdown "$failure_stage")
digest_outcome=$(escape_markdown "${DIGEST_OUTCOME:-skipped}")

{
  printf '## AIWeLink image publication\n\n'
  printf '| Field | Value |\n'
  printf '| --- | --- |\n'
  printf '| Publish status | `%s` |\n' "$publish_status"
  printf '| Failure stage | `%s` |\n' "$failure_stage"
  printf '| Branch | `%s` |\n' "$branch"
  printf '| Commit | `%s` |\n' "$commit"
  printf '| Application version | `%s` |\n' "$application_version"
  printf '| Upstream version | `%s` |\n' "$upstream_version"
  printf '| Mutable tag | `%s` |\n' "$mutable_tag"
  printf '| Immutable tag | `%s` |\n' "$immutable_tag"
  printf '| Digest | `%s` |\n' "$digest"
  printf '| Digest lookup | `%s` |\n' "$digest_outcome"
  printf '\n'
  if [ "$publish_status" = success ]; then
    printf '### Pull commands\n\n```bash\n'
    printf 'docker pull %s\n' "$immutable_tag"
    if [ "$digest" != unavailable ]; then
      printf 'docker pull %s@%s\n' "$IMAGE_NAME" "$digest"
    fi
    printf '```\n'
  else
    printf 'Image was not published; pull commands are unavailable.\n'
  fi
} >> "$GITHUB_STEP_SUMMARY"
