#!/bin/bash

set -euo pipefail

ROOT_DIR=$(cd "$(dirname "${BASH_SOURCE[0]}")/../.." && pwd)
TEMP_DIR=$(mktemp -d)
trap 'rm -rf "$TEMP_DIR"' EXIT

cat > "$TEMP_DIR/curl" <<'EOF'
#!/bin/bash
printf '%s\n' "$@" > "$CURL_ARGS_LOG"
env > "${CURL_ARGS_LOG}.env"
cat > "${CURL_ARGS_LOG}.stdin"
EOF
chmod +x "$TEMP_DIR/curl"

mkdir "$TEMP_DIR/home"
cat > "$TEMP_DIR/home/.curlrc" <<'EOF'
url = "https://example.com/collect"
header = "X-Leaked-From-Curlrc: yes"
EOF

run_api_curl() {
    CURL_ARGS_LOG="$1" HOME="$TEMP_DIR/home" PATH="$TEMP_DIR:$PATH" UPDATE_GITHUB_TOKEN="${2:-}" \
        GITHUB_TOKEN="github-fallback" GH_TOKEN="gh-fallback" \
        bash -c 'source <(head -n -1 "$1"); github_api_curl -s "$2"' bash \
        "$ROOT_DIR/deploy/install.sh" "https://api.github.com/repos/AIwelink/sub2api-aiwelink-dev/releases/latest"
}

run_api_curl "$TEMP_DIR/authenticated" "update-secret"
test "$(head -n 1 "$TEMP_DIR/authenticated")" = '-q'
grep -Fxq -- '--config' "$TEMP_DIR/authenticated"
grep -Fxq -- '-' "$TEMP_DIR/authenticated"
grep -Fxq -- '--globoff' "$TEMP_DIR/authenticated"
grep -Fxq 'header = "Authorization: Bearer update-secret"' "$TEMP_DIR/authenticated.stdin"
if grep -Fq 'update-secret' "$TEMP_DIR/authenticated"; then
    echo "installer exposed the update token in curl argv" >&2
    exit 1
fi
if grep -Eq 'update-secret|github-fallback|gh-fallback' "$TEMP_DIR/authenticated.env"; then
    echo "installer exposed a token in curl environment" >&2
    exit 1
fi
test "$(grep -Fxc 'https://api.github.com/repos/AIwelink/sub2api-aiwelink-dev/releases/latest' "$TEMP_DIR/authenticated")" -eq 1
if grep -Fq 'example.com/collect' "$TEMP_DIR/authenticated" || grep -Fq 'X-Leaked-From-Curlrc' "$TEMP_DIR/authenticated" ||
    grep -Fq 'example.com/collect' "$TEMP_DIR/authenticated.stdin" || grep -Fq 'X-Leaked-From-Curlrc' "$TEMP_DIR/authenticated.stdin"; then
    echo "installer allowed hostile curl config into authenticated invocation" >&2
    exit 1
fi

run_api_curl "$TEMP_DIR/anonymous"
test "$(head -n 1 "$TEMP_DIR/anonymous")" = '-q'
if grep -Eq 'github-fallback|gh-fallback' "$TEMP_DIR/anonymous.env"; then
    echo "installer exposed a fallback token in anonymous curl environment" >&2
    exit 1
fi
if grep -Fq 'Authorization:' "$TEMP_DIR/anonymous"; then
    echo "installer unexpectedly used a fallback token" >&2
    exit 1
fi
test ! -s "$TEMP_DIR/anonymous.stdin"
test "$(grep -Fxc 'https://api.github.com/repos/AIwelink/sub2api-aiwelink-dev/releases/latest' "$TEMP_DIR/anonymous")" -eq 1
if grep -Fq 'example.com/collect' "$TEMP_DIR/anonymous" || grep -Fq 'X-Leaked-From-Curlrc' "$TEMP_DIR/anonymous"; then
    echo "installer allowed hostile curl config into anonymous invocation" >&2
    exit 1
fi

assert_unsafe_invocation_rejected() {
    local name=$1
    shift
    rm -f "$TEMP_DIR/$name" "$TEMP_DIR/$name.stdin"
    if CURL_ARGS_LOG="$TEMP_DIR/$name" PATH="$TEMP_DIR:$PATH" UPDATE_GITHUB_TOKEN="update-secret" \
        bash -c 'source <(head -n -1 "$1"); shift; github_api_curl "$@"' bash \
        "$ROOT_DIR/deploy/install.sh" "$@" 2>/dev/null; then
        echo "installer accepted unsafe curl invocation: $name" >&2
        exit 1
    fi
    if [ -e "$TEMP_DIR/$name" ]; then
        echo "installer invoked curl for unsafe request: $name" >&2
        exit 1
    fi
}

assert_unsafe_invocation_rejected non-api -s \
    "https://github.com/Wei-Shaw/sub2api/releases/download/v1/asset"
assert_unsafe_invocation_rejected mixed-host -s \
    "https://api.github.com/repos/Wei-Shaw/sub2api/releases/latest" \
    "https://example.com/collect"
assert_unsafe_invocation_rejected multiple-api -s \
    "https://api.github.com/repos/Wei-Shaw/sub2api/releases/latest" \
    "https://api.github.com/repos/Wei-Shaw/sub2api/releases"
assert_unsafe_invocation_rejected url-option -s --url \
    "https://example.com/collect" \
    "https://api.github.com/repos/Wei-Shaw/sub2api/releases/latest"

# Every installer release API request must use the scoped helper.
test "$(grep -c 'github_api_curl .*https://api.github.com/' "$ROOT_DIR/deploy/install.sh")" -eq 3

bash -c '
    source <(head -n -1 "$1")
    test "$(normalize_aiwelink_version "0.1.170-2.4")" = "v0.1.170-2.4"
    test "$(normalize_aiwelink_version "v0.1.170-1")" = "v0.1.170-1"
    ! normalize_aiwelink_version "0.1.170"
    ! normalize_aiwelink_version "v0.1.170-0"
    ! normalize_aiwelink_version "v0.1.170-2.0"
' bash "$ROOT_DIR/deploy/install.sh"

bash -c '
    source <(head -n -1 "$1")
    stable_release='"'"'{
      "tag_name": "v0.1.170-2.4",
      "draft": false,
      "prerelease": false
    }'"'"'
    draft_release='"'"'{
      "tag_name": "v0.1.170-2.4",
      "draft": true,
      "prerelease": false
    }'"'"'
    prerelease='"'"'{
      "tag_name": "v0.1.170-2.4",
      "draft": false,
      "prerelease": true
    }'"'"'

    test "$(stable_release_version_from_json "$stable_release")" = "v0.1.170-2.4"
    ! stable_release_version_from_json "$draft_release"
    ! stable_release_version_from_json "$prerelease"
' bash "$ROOT_DIR/deploy/install.sh"

grep -A 8 '^release:' "$ROOT_DIR/.goreleaser.simple.yaml" | grep -Fxq '  prerelease: true'

release_job_header=$(sed -n '/^  release:/,/^    steps:/p' "$ROOT_DIR/.github/workflows/release.yml")
if grep -Fq 'AIWELINK_REGISTRY_PASSWORD' <<<"$release_job_header"; then
    echo "private registry password is exposed at release-job scope" >&2
    exit 1
fi

goreleaser_line=$(grep -nF -- '- name: Run GoReleaser' "$ROOT_DIR/.github/workflows/release.yml" | cut -d: -f1)
registry_login_line=$(grep -nF -- '- name: Login to AIWeLink Container Registry' "$ROOT_DIR/.github/workflows/release.yml" | cut -d: -f1)
private_publish_line=$(grep -nF -- '- name: Publish AIWeLink private registry image' "$ROOT_DIR/.github/workflows/release.yml" | cut -d: -f1)
test "$registry_login_line" -gt "$goreleaser_line"
test "$registry_login_line" -lt "$private_publish_line"

# Asset and checksum downloads must continue to call curl directly.
grep -Fq 'curl -sL "$download_url"' "$ROOT_DIR/deploy/install.sh"
grep -Fq 'curl -sL "$checksum_url"' "$ROOT_DIR/deploy/install.sh"

echo "install GitHub token checks passed"
