#!/bin/sh
set -eu

repo_root=$(CDPATH= cd -- "$(dirname -- "$0")/../.." && pwd)
cd "$repo_root"

compose_file=deploy/docker-compose.aiwelink-dev.yml
env_file=deploy/.env.aiwelink-dev.example
docker_bin=${DOCKER_BIN:-docker}

fail() {
  printf 'AIWeLink growth deployment test failed: %s\n' "$1" >&2
  exit 1
}

assert_line() {
  file=$1
  line=$2
  awk -v expected="$line" '
    { sub(/\r$/, "") }
    $0 == expected { found = 1 }
    END { exit found ? 0 : 1 }
  ' "$file" || fail "$file is missing: $line"
}

test -f "$compose_file" || fail "$compose_file is missing"
test -f "$env_file" || fail "$env_file is missing"

assert_line "$env_file" 'SERVER_PORT=8080'
assert_line "$env_file" 'GROWTH_REGISTRATION_ENABLED=true'
assert_line "$env_file" 'GROWTH_REGISTRATION_ENDPOINT=https://aiwelink.cc/internal/growth/registrations/bind'
assert_line "$env_file" 'GROWTH_REGISTRATION_SITE_ID=aiwelink'
assert_line "$env_file" 'GROWTH_REGISTRATION_COOKIE_NAME=awl_growth_sid'
assert_line "$env_file" 'GROWTH_REGISTRATION_CONNECT_TIMEOUT_SECONDS=2'
assert_line "$env_file" 'GROWTH_REGISTRATION_READ_TIMEOUT_SECONDS=5'

grep -Eq '^GROWTH_REGISTRATION_SERVICE_CREDENTIAL=replace_' "$env_file" || \
  fail 'service credential placeholder is missing'
grep -Eq '^GROWTH_REGISTRATION_OUTBOX_ENCRYPTION_KEY=replace_' "$env_file" || \
  fail 'outbox key placeholder is missing'

if grep -Eq '^GROWTH_LOGIN_' "$env_file"; then
  fail 'legacy GROWTH_LOGIN_* variables must not appear in the registration template'
fi

render_dir=$(mktemp -d deploy/.aiwelink-growth-test.XXXXXX)
rendered=$render_dir/rendered.yml
cp "$compose_file" "$render_dir/docker-compose.yml"
cp "$env_file" "$render_dir/.env"
cleanup() {
  rm -f "$rendered" "$render_dir/docker-compose.yml" "$render_dir/.env"
  rmdir "$render_dir"
}
trap cleanup EXIT HUP INT TERM

"$docker_bin" compose \
  --env-file "$render_dir/.env" \
  -f "$render_dir/docker-compose.yml" \
  config >"$rendered"

grep -Fq 'target: 8080' "$rendered" || fail 'container port 8080 was not rendered'
grep -Fq 'published: "8080"' "$rendered" || fail 'development host port 8080 was not rendered'
grep -Fq 'http://localhost:8080/health' "$rendered" || fail 'health check does not use container port 8080'
grep -Fq 'GROWTH_REGISTRATION_ENABLED: "true"' "$rendered" || fail 'registration integration is not enabled'
grep -Fq 'GROWTH_REGISTRATION_ENDPOINT: https://aiwelink.cc/internal/growth/registrations/bind' "$rendered" || \
  fail 'Traffic HTTPS endpoint was not rendered'
grep -Fq 'name: 1panel-network' "$rendered" || fail 'external 1Panel network was not rendered'

printf 'AIWeLink growth deployment test passed\n'
