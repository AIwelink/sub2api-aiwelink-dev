#!/usr/bin/env bash
set -euo pipefail

PROBE_TMP_DIR=''
SELF_TEST_FIXTURE_DIR=''
SELF_TEST_SERVER_PID=''

cleanup_probe() {
  if [[ -n $PROBE_TMP_DIR ]]; then
    rm -rf -- "$PROBE_TMP_DIR"
  fi
}

cleanup_self_test() {
  if [[ -n $SELF_TEST_SERVER_PID ]]; then
    kill "$SELF_TEST_SERVER_PID" 2>/dev/null || true
    wait "$SELF_TEST_SERVER_PID" 2>/dev/null || true
  fi
  if [[ -n $SELF_TEST_FIXTURE_DIR ]]; then
    rm -rf -- "$SELF_TEST_FIXTURE_DIR"
  fi
}

fail() {
  printf 'Growth public canary failed: %s\n' "$1" >&2
  exit 1
}

parse_url() {
  local url=$1
  if [[ ! $url =~ ^(https?)://([^/?#]+)(/[^?#]*)?([?#].*)?$ ]]; then
    return 1
  fi

  URL_SCHEME=${BASH_REMATCH[1]}
  URL_AUTHORITY=${BASH_REMATCH[2]}
  URL_PATH=${BASH_REMATCH[3]:-/}
  [[ $URL_AUTHORITY != *@* ]] || return 1
  URL_HOST=${URL_AUTHORITY%%:*}
  [[ -n $URL_HOST ]]
}

host_is_allowed() {
  local candidate=$1 allowed
  IFS=',' read -ra allowed_hosts <<<"$GROWTH_CANARY_ALLOWED_REDIRECT_HOSTS"
  for allowed in "${allowed_hosts[@]}"; do
    if [[ $candidate == "$allowed" ]]; then
      return 0
    fi
  done
  return 1
}

validate_request_url() {
  local url=$1
  parse_url "$url" || fail 'received an invalid redirect URL'
  if [[ $URL_SCHEME != https && ${GROWTH_CANARY_ALLOW_HTTP_FOR_SELF_TEST:-false} != true ]]; then
    fail 'received a non-HTTPS redirect URL'
  fi
  host_is_allowed "$URL_HOST" || fail "redirect host is not allowed: $URL_HOST"

  local normalized_path=${URL_PATH,,}
  if [[ $normalized_path =~ (^|/)(register|registration|registrations|bind|login|payment|payments)(/|$) ]]; then
    fail 'redirect targets a write-capable or authentication endpoint'
  fi
}

read_header() {
  local headers=$1 name=${2,,}
  awk -v wanted="$name" '
    {
      sub(/\r$/, "")
      split($0, parts, ":")
      if (tolower(parts[1]) == wanted) {
        sub(/^[^:]*:[[:space:]]*/, "")
        value = $0
      }
    }
    END { print value }
  ' "$headers"
}

request() {
  local url=$1 headers=$2
  local -a resolve_args=() no_proxy_args=() tls_args=()
  if [[ -n ${GROWTH_CANARY_RESOLVE:-} ]]; then
    resolve_args=(--resolve "$GROWTH_CANARY_RESOLVE")
  fi
  if [[ -n ${GROWTH_CANARY_CURL_NOPROXY:-} ]]; then
    no_proxy_args=(--noproxy "$GROWTH_CANARY_CURL_NOPROXY")
  fi
  if [[ ${GROWTH_CANARY_INSECURE_TLS_FOR_SELF_TEST:-false} == true ]]; then
    tls_args=(--insecure)
  fi

  curl --silent --show-error \
    --connect-timeout 5 \
    --max-time 15 \
    --request GET \
    --output /dev/null \
    --dump-header "$headers" \
    --cookie "$COOKIE_JAR" \
    --cookie-jar "$COOKIE_JAR" \
    "${resolve_args[@]}" \
    "${no_proxy_args[@]}" \
    "${tls_args[@]}" \
    --write-out '%{http_code}' \
    "$url"
}

assert_growth_cookie() {
  local expected_domain=$1
  awk -F '\t' -v domain="#HttpOnly_${expected_domain}" '
    $1 == domain && $4 == "TRUE" && $6 == "awl_growth_sid" { found = 1 }
    END { exit found ? 0 : 1 }
  ' "$COOKIE_JAR" || fail 'secure HttpOnly growth session cookie is missing'
}

run_probe() {
  local base_url=${GROWTH_CANARY_BASE_URL:-https://api.aiwelink.cc}
  local code=${GROWTH_CANARY_REFERRAL_CODE:-}
  GROWTH_CANARY_ALLOWED_REDIRECT_HOSTS=${GROWTH_CANARY_ALLOWED_REDIRECT_HOSTS:-aiwelink.cc,api.aiwelink.cc}
  local expected_final_host=${GROWTH_CANARY_EXPECTED_FINAL_HOST:-api.aiwelink.cc}
  local expected_cookie_domain=${GROWTH_CANARY_COOKIE_DOMAIN:-.aiwelink.cc}

  [[ $code =~ ^[a-hj-km-np-z2-9]{8}$ ]] || fail 'GROWTH_CANARY_REFERRAL_CODE must be an 8-character referral code'
  parse_url "$base_url" || fail 'GROWTH_CANARY_BASE_URL is invalid'
  if [[ $URL_SCHEME != https && ${GROWTH_CANARY_ALLOW_HTTP_FOR_SELF_TEST:-false} != true ]]; then
    fail 'GROWTH_CANARY_BASE_URL must use HTTPS'
  fi

  local tmp_dir
  tmp_dir=$(mktemp -d)
  PROBE_TMP_DIR=$tmp_dir
  COOKIE_JAR="$tmp_dir/cookies.txt"
  : >"$COOKIE_JAR"
  trap cleanup_probe EXIT HUP INT TERM

  local -a resolve_args=() no_proxy_args=() tls_args=()
  if [[ -n ${GROWTH_CANARY_RESOLVE:-} ]]; then
    resolve_args=(--resolve "$GROWTH_CANARY_RESOLVE")
  fi
  if [[ -n ${GROWTH_CANARY_CURL_NOPROXY:-} ]]; then
    no_proxy_args=(--noproxy "$GROWTH_CANARY_CURL_NOPROXY")
  fi
  if [[ ${GROWTH_CANARY_INSECURE_TLS_FOR_SELF_TEST:-false} == true ]]; then
    tls_args=(--insecure)
  fi
  curl --silent --show-error --fail \
    --connect-timeout 5 --max-time 15 \
    --request GET --output /dev/null \
    "${resolve_args[@]}" \
    "${no_proxy_args[@]}" \
    "${tls_args[@]}" \
    "${base_url%/}/health" || fail 'health endpoint is unavailable'

  local current_url="${base_url%/}/r/$code"
  local headers="$tmp_dir/headers.txt"
  local status location cache_control final_url=''

  status=$(request "$current_url" "$headers") || fail 'referral entry request failed'
  [[ $status == 302 ]] || fail "referral entry returned HTTP $status instead of 302"
  cache_control=$(read_header "$headers" 'cache-control')
  [[ ${cache_control,,} == *no-store* ]] || fail 'referral entry is missing Cache-Control: no-store'
  location=$(read_header "$headers" 'location')
  [[ -n $location ]] || fail 'referral entry is missing Location'

  local hop
  for hop in 1 2 3 4 5; do
    current_url=$location
    validate_request_url "$current_url"
    status=$(request "$current_url" "$headers") || fail 'redirect request failed'
    if [[ $status =~ ^2[0-9][0-9]$ ]]; then
      final_url=$current_url
      break
    fi
    [[ $status =~ ^30[12378]$ ]] || fail "redirect returned unexpected HTTP $status"
    location=$(read_header "$headers" 'location')
    [[ -n $location ]] || fail 'redirect response is missing Location'
  done

  [[ -n $final_url ]] || fail 'referral chain exceeded five redirects'
  parse_url "$final_url" || fail 'final URL is invalid'
  [[ $URL_HOST == "$expected_final_host" ]] || fail "unexpected final host: $URL_HOST"
  assert_growth_cookie "$expected_cookie_domain"

  printf 'Growth public canary passed\n'
}

run_self_test() {
  command -v python3 >/dev/null 2>&1 || fail 'python3 is required for --self-test'
  command -v openssl >/dev/null 2>&1 || fail 'openssl is required for --self-test'

  local fixture_dir port_file server_log server_pid fixture_port fixture_base cert_file key_file
  fixture_dir=$(mktemp -d)
  port_file="$fixture_dir/port"
  server_log="$fixture_dir/server.log"
  cert_file="$fixture_dir/cert.pem"
  key_file="$fixture_dir/key.pem"

  openssl req -x509 -newkey rsa:2048 -nodes \
    -keyout "$key_file" -out "$cert_file" -days 1 \
    -subj '/CN=api.aiwelink.test' >/dev/null 2>&1

  python3 - "$port_file" "$cert_file" "$key_file" >"$server_log" 2>&1 <<'PY' &
import http.server
import pathlib
import ssl
import sys
import urllib.parse

port_file = pathlib.Path(sys.argv[1])
cert_file = sys.argv[2]
key_file = sys.argv[3]


class Handler(http.server.BaseHTTPRequestHandler):
    def log_message(self, _format, *_args):
        pass

    def do_GET(self):
        parsed = urllib.parse.urlparse(self.path)
        port = self.server.server_port
        origin = f"https://api.aiwelink.test:{port}"

        if parsed.path == "/health":
            self.send_response(200)
            self.end_headers()
            return

        if parsed.path.startswith("/r/"):
            code = parsed.path.rsplit("/", 1)[-1]
            if code == "bcdefghm":
                self.send_response(200)
                self.end_headers()
                return
            mode = "insecure-cookie" if code == "bcdefghk" else "bad-host" if code == "bcdefghn" else "good"
            self.send_response(302)
            self.send_header("Cache-Control", "private, no-store")
            self.send_header("Location", f"{origin}/traffic?mode={mode}")
            self.end_headers()
            return

        if parsed.path == "/traffic":
            mode = urllib.parse.parse_qs(parsed.query).get("mode", ["good"])[0]
            self.send_response(302)
            if mode == "insecure-cookie":
                self.send_header(
                    "Set-Cookie",
                    "awl_growth_sid=fixture-secret; Domain=.aiwelink.test; Path=/; SameSite=Lax",
                )
            else:
                self.send_header(
                    "Set-Cookie",
                    "awl_growth_sid=fixture-secret; Domain=.aiwelink.test; Path=/; Secure; HttpOnly; SameSite=Lax",
                )
            target_host = "127.0.0.1" if mode == "bad-host" else "api.aiwelink.test"
            self.send_header("Location", f"https://{target_host}:{port}/final")
            self.end_headers()
            return

        if parsed.path == "/final":
            self.send_response(200)
            self.end_headers()
            return

        self.send_response(404)
        self.end_headers()


server = http.server.ThreadingHTTPServer(("127.0.0.1", 0), Handler)
tls_context = ssl.SSLContext(ssl.PROTOCOL_TLS_SERVER)
tls_context.load_cert_chain(cert_file, key_file)
server.socket = tls_context.wrap_socket(server.socket, server_side=True)
port_file.write_text(str(server.server_port), encoding="ascii")
server.serve_forever()
PY
  server_pid=$!
  SELF_TEST_SERVER_PID=$server_pid
  SELF_TEST_FIXTURE_DIR=$fixture_dir
  trap cleanup_self_test EXIT HUP INT TERM

  local attempt
  for attempt in {1..50}; do
    [[ -s $port_file ]] && break
    sleep 0.1
  done
  [[ -s $port_file ]] || fail 'self-test fixture did not start'
  fixture_port=$(<"$port_file")
  fixture_base="https://api.aiwelink.test:$fixture_port"

  fixture_probe() {
    GROWTH_CANARY_BASE_URL="$fixture_base" \
    GROWTH_CANARY_REFERRAL_CODE="$1" \
    GROWTH_CANARY_RESOLVE="api.aiwelink.test:$fixture_port:127.0.0.1" \
    GROWTH_CANARY_CURL_NOPROXY='*' \
    GROWTH_CANARY_INSECURE_TLS_FOR_SELF_TEST=true \
    GROWTH_CANARY_ALLOWED_REDIRECT_HOSTS='api.aiwelink.test' \
    GROWTH_CANARY_EXPECTED_FINAL_HOST='api.aiwelink.test' \
    GROWTH_CANARY_COOKIE_DOMAIN='.aiwelink.test' \
      "$0"
  }

  expect_failure() {
    local name=$1 code=$2
    if fixture_probe "$code" >"$fixture_dir/$name.out" 2>&1; then
      fail "self-test case unexpectedly passed: $name"
    fi
  }

  expect_failure invalid-code 'abcd0fgh'
  expect_failure missing-cookie 'bcdefghk'
  expect_failure non-redirect 'bcdefghm'
  expect_failure unexpected-host 'bcdefghn'
  fixture_probe 'bcdefghj' >/dev/null

  printf 'Growth public canary self-test passed\n'
}

case ${1:-} in
  --self-test)
    run_self_test
    ;;
  '')
    run_probe
    ;;
  *)
    fail "unknown argument: $1"
    ;;
esac
