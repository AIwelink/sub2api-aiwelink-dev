# AIWeLink Growth Registration Deployment Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Add a validated, secret-free AIWeLink development/gray Compose deployment example that enables registration attribution through the public Traffic endpoint.

**Architecture:** A dedicated Compose file keeps AIWeLink-specific external PostgreSQL, Redis, port, image, and network assumptions out of the generic upstream deployment files. A companion env template owns all environment-specific values, while a shell contract test renders the Compose model and rejects the old `GROWTH_LOGIN_*` names. A Chinese operations guide documents deployment, secret handling, shared-database constraints, and end-to-end verification.

**Tech Stack:** Docker Compose v2, POSIX shell, GitHub Actions, Markdown, Go 1.26.5 regression tests

---

## File Structure

- Create `deploy/docker-compose.aiwelink-dev.yml`: AIWeLink-only Sub2API service definition with external database, Redis, and Docker network.
- Create `deploy/.env.aiwelink-dev.example`: non-secret development/gray runtime template and the eight registration-attribution variables.
- Create `deploy/tests/aiwelink-growth-deployment-test.sh`: executable deployment contract test using static assertions plus `docker compose config`.
- Modify `.github/workflows/backend-ci.yml`: execute the new deployment contract test in a dedicated Ubuntu Compose job.
- Create `deploy/AIWELINK_GROWTH_REGISTRATION_CN.md`: Chinese deployment and diagnosis runbook.
- Modify `deploy/README.md`: link the new AIWeLink-specific files without changing generic deployment behavior.

### Task 1: Add The Failing Deployment Contract Test

**Files:**
- Create: `deploy/tests/aiwelink-growth-deployment-test.sh`
- Modify: `.github/workflows/backend-ci.yml`

- [x] **Step 1: Create the contract test before the deployment files**

Create a POSIX shell test with these checks:

```sh
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
```

- [x] **Step 2: Wire the test into a dedicated Ubuntu CI job**

Add a job that runs on an image with Docker Compose available, leaving the existing macOS shell job unchanged:

```yaml
  compose:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v6
      - name: Check AIWeLink growth deployment
        run: /bin/sh deploy/tests/aiwelink-growth-deployment-test.sh
```

- [x] **Step 3: Run the test and verify it fails for the intended reason**

Run:

```bash
sh deploy/tests/aiwelink-growth-deployment-test.sh
```

Expected: exit `1` with `deploy/docker-compose.aiwelink-dev.yml is missing`.

### Task 2: Implement The Compose And Environment Examples

**Files:**
- Create: `deploy/docker-compose.aiwelink-dev.yml`
- Create: `deploy/.env.aiwelink-dev.example`
- Test: `deploy/tests/aiwelink-growth-deployment-test.sh`

- [x] **Step 1: Add the dedicated Compose file**

Create this service model:

```yaml
name: ${COMPOSE_PROJECT_NAME:-sub2api-aiwelink-dev}

services:
  sub2api:
    image: ${SUB2API_IMAGE:?SUB2API_IMAGE must use a published AIWeLink image tag}
    pull_policy: always
    container_name: ${CONTAINER_NAME:-sub2api-8080}
    restart: unless-stopped
    stop_grace_period: 30s
    security_opt:
      - no-new-privileges:true
    ulimits:
      nofile:
        soft: 100000
        hard: 100000
    ports:
      - "${BIND_HOST:-0.0.0.0}:${SERVER_PORT:-8080}:8080"
    volumes:
      - sub2api_data:/app/data
    env_file:
      - ${SUB2API_ENV_FILE:-.env}
    environment:
      AUTO_SETUP: "true"
      SERVER_HOST: "0.0.0.0"
      SERVER_PORT: "8080"
    healthcheck:
      test: ["CMD", "wget", "-q", "-T", "5", "-O", "/dev/null", "http://localhost:8080/health"]
      interval: 10s
      timeout: 5s
      retries: 12
      start_period: 240s
    networks:
      - 1panel-network

volumes:
  sub2api_data:
    driver: local

networks:
  1panel-network:
    external: true
    name: ${ONEPANEL_NETWORK_NAME:-1panel-network}
```

- [x] **Step 2: Add the non-secret env template**

Create a template grouped into deployment, database, Redis, application-secret, and growth-registration sections. It must contain these exact operational values:

```dotenv
COMPOSE_PROJECT_NAME=sub2api-aiwelink-dev
CONTAINER_NAME=sub2api-8080
SUB2API_IMAGE=docker.aiwelink.cc/sub2api-aiwelink-dev:dev-REPLACE_WITH_12_CHARACTER_SHA
SUB2API_ENV_FILE=.env
BIND_HOST=0.0.0.0
SERVER_PORT=8080
SERVER_MODE=release
RUN_MODE=standard
TZ=Asia/Shanghai
SETUP_MIGRATION_TIMEOUT_SECONDS=600
ONEPANEL_NETWORK_NAME=1panel-network

DATABASE_HOST=replace_with_postgresql_host
DATABASE_PORT=5432
DATABASE_DBNAME=replace_with_database_name
DATABASE_USER=replace_with_database_user
DATABASE_PASSWORD=replace_with_database_password
DATABASE_SSLMODE=disable
DATABASE_MAX_OPEN_CONNS=200
DATABASE_MAX_IDLE_CONNS=20
DATABASE_CONN_MAX_LIFETIME_MINUTES=30
DATABASE_CONN_MAX_IDLE_TIME_MINUTES=5

REDIS_HOST=replace_with_redis_host
REDIS_PORT=6379
REDIS_USERNAME=
REDIS_PASSWORD=replace_with_redis_password
REDIS_DB=0
REDIS_POOL_SIZE=1024
REDIS_MIN_IDLE_CONNS=10
REDIS_ENABLE_TLS=false

JWT_SECRET=replace_with_64_character_hex_jwt_secret
TOTP_ENCRYPTION_KEY=replace_with_64_character_hex_totp_key

GROWTH_REGISTRATION_ENABLED=true
GROWTH_REGISTRATION_ENDPOINT=https://aiwelink.cc/internal/growth/registrations/bind
GROWTH_REGISTRATION_SITE_ID=aiwelink
GROWTH_REGISTRATION_SERVICE_CREDENTIAL=replace_with_traffic_aiwelink_service_credential
GROWTH_REGISTRATION_OUTBOX_ENCRYPTION_KEY=replace_with_shared_64_character_hex_outbox_key
GROWTH_REGISTRATION_COOKIE_NAME=awl_growth_sid
GROWTH_REGISTRATION_CONNECT_TIMEOUT_SECONDS=2
GROWTH_REGISTRATION_READ_TIMEOUT_SECONDS=5
```

Add comments explaining that production uses host port/container identity `8081`, real values must be stored only in `.env`, the service credential equals Traffic's `aiwelink` credential, and the outbox key is separate from that credential but shared by all enabled Sub2API instances that share the database.

- [x] **Step 3: Run the deployment test and verify it passes**

Run:

```bash
sh deploy/tests/aiwelink-growth-deployment-test.sh
```

Expected: `AIWeLink growth deployment test passed`.

- [x] **Step 4: Commit the tested deployment contract**

```bash
git add .github/workflows/backend-ci.yml \
  deploy/docker-compose.aiwelink-dev.yml \
  deploy/.env.aiwelink-dev.example \
  deploy/tests/aiwelink-growth-deployment-test.sh
git update-index --chmod=+x deploy/tests/aiwelink-growth-deployment-test.sh
git commit -m "docs: add AIWeLink growth deployment example"
```

### Task 3: Add The Chinese Operations Guide

**Files:**
- Create: `deploy/AIWELINK_GROWTH_REGISTRATION_CN.md`
- Modify: `deploy/README.md`

- [x] **Step 1: Write the operations guide**

Create a guide with these exact sections and outcomes:

```markdown
# AIWeLink 推广注册绑定部署

## 部署关系
Explain host port 8080 for development/gray, 8081 for production, container port 8080, shared PostgreSQL/Redis, the public HTTPS Traffic endpoint, and the `.aiwelink.cc` cookie.

## 准备 `.env`
Show copying `.env.aiwelink-dev.example` to `.env`, setting mode 0600, replacing every `replace_` value, and pinning the immutable image emitted after the `aiwelink-dev` merge.

## 密钥规则
Show `openssl rand -hex 32` for JWT, TOTP, and outbox keys. State that the Traffic service credential must exactly equal `SITE_SERVICE_CREDENTIALS_JSON.aiwelink`, while the outbox key is independent and must be shared by all enabled Sub2API instances using the same database.

## 校验并启动
Show `docker compose ... config --quiet`, `pull`, `up -d`, `ps`, and health commands from the `deploy` directory.

## 检查容器配置
Show a secret-safe `docker exec` command that prints public growth settings and only secret lengths, then an unauthenticated Traffic POST probe whose expected status is 401.

## 端到端验收
Show the `/r/{code}` browser flow, email registration, Traffic binding verification, and the requirement that the browser sends `awl_growth_sid` to the Sub2API registration request.

## Outbox 排障
Show a PostgreSQL query for pending/dead-letter rows without exposing encrypted sessions. Explain `401/403`, `404`, `422`, `decrypt_failed`, and retryable `503` outcomes.

## 更新和回滚
Show changing only the immutable image tag, running `pull` and `up -d`, and reverting to the previous immutable tag without changing the database.
```

Use fully executable shell commands and the actual file/container names. Do not include a real credential, database address, password, or encryption key.

- [x] **Step 2: Link the AIWeLink artifacts from the deploy README**

Add these rows to the deployment file table:

```markdown
| `docker-compose.aiwelink-dev.yml` | AIWeLink development/gray deployment with external PostgreSQL and Redis |
| `.env.aiwelink-dev.example` | AIWeLink non-secret environment template |
| `AIWELINK_GROWTH_REGISTRATION_CN.md` | AIWeLink 推广注册绑定部署与排障说明 |
```

- [x] **Step 3: Check documentation and commit**

Run:

```bash
rg -n 'GROWTH_LOGIN_|replace_|GROWTH_REGISTRATION_|8080|8081' \
  deploy/.env.aiwelink-dev.example \
  deploy/AIWELINK_GROWTH_REGISTRATION_CN.md
git diff --check
```

Expected: `GROWTH_LOGIN_` appears only in a warning explaining that it is obsolete; all eight current variables appear, both host ports are explained, and `git diff --check` exits `0`.

Commit:

```bash
git add deploy/AIWELINK_GROWTH_REGISTRATION_CN.md deploy/README.md
git commit -m "docs: document AIWeLink growth rollout"
```

### Task 4: Verify And Publish The Pull Request

**Files:**
- Verify all changed files from Tasks 1-3.

- [x] **Step 1: Run focused deployment and growth tests**

```bash
sh deploy/tests/aiwelink-growth-deployment-test.sh
sh deploy/tests/docker-compose-security-test.sh
cd backend
go test ./internal/config ./internal/server/middleware ./internal/repository ./internal/service
```

Expected: all commands exit `0` and every package reports `ok`.

- [x] **Step 2: Run repository hygiene checks**

```bash
git diff origin/aiwelink-dev...HEAD --check
git status --short --branch
git diff --name-only origin/aiwelink-dev...HEAD
```

Expected: no whitespace errors, a clean worktree, and exactly the eight intended files: the design, plan, Compose file, env template, shell test, Chinese guide, CI workflow, and deploy README.

- [ ] **Step 3: Push the feature branch**

```bash
git push -u origin codex/aiwelink-growth-deployment
```

- [ ] **Step 4: Open a pull request to `aiwelink-dev`**

Use a title such as `docs: add AIWeLink growth deployment example`. The PR body must summarize the corrected `GROWTH_REGISTRATION_*` configuration, shared-database key rule, test coverage, and explicitly state that `main` is not the target.

- [ ] **Step 5: Verify GitHub checks and image behavior**

Confirm the PR CI checks start successfully. Document that image publication occurs only after merge to `aiwelink-dev`; the workflow then publishes `dev`, the application version tag, and immutable `dev-<merge-sha>`.
