# AIWeLink Test Docker Deployment Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Produce a server-ready Sub2API Compose file and env files that run the immutable development image against the existing remote test PostgreSQL and Redis.

**Architecture:** A single Sub2API container is pulled from `docker.aiwelink.cc`, publishes host port `0.0.0.0:8081`, persists `/app/data`, and joins the external `1panel-network`. Compose consumes a real ignored env file copied from the verified local development settings; no database or Redis container is created.

**Tech Stack:** Docker Compose v2, OCI image registry, Sub2API 0.1.168, PostgreSQL, Redis, 1Panel network

---

### Task 1: Secret Boundary and Env Template

**Files:**
- Modify: `.gitignore`
- Create: `deploy/.env.aiwelink-test.example`

- [ ] **Step 1: Ignore the real server env**

Add this exact entry beside `deploy/.env.develop`:

```gitignore
deploy/.env.aiwelink-test
```

- [ ] **Step 2: Create the redacted env template**

Create `deploy/.env.aiwelink-test.example` with all required keys, fixed non-secret runtime values, and `replace_*` placeholders for secrets and remote hosts. The fixed values must include:

```dotenv
SUB2API_IMAGE=docker.aiwelink.cc/sub2api-aiwelink-dev:dev-0c4864e109
BIND_HOST=0.0.0.0
SERVER_PORT=8081
SERVER_MODE=release
RUN_MODE=standard
TZ=Asia/Shanghai
SETUP_MIGRATION_TIMEOUT_SECONDS=600
DATABASE_SSLMODE=disable
GROWTH_LOGIN_ENABLED=true
GROWTH_LOGIN_ENDPOINT=https://aiwelink.cc/internal/growth/logins
GROWTH_SITE_ID=aiwelink
GROWTH_CONNECT_TIMEOUT_SECONDS=5
GROWTH_READ_TIMEOUT_SECONDS=5
ONEPANEL_NETWORK_NAME=1panel-network
```

- [ ] **Step 3: Verify the secret boundary**

Run:

```powershell
git check-ignore -v deploy/.env.aiwelink-test
rg -n "replace_" deploy/.env.aiwelink-test.example
```

Expected: the real env is ignored and every secret-bearing template field remains a placeholder.

### Task 2: Server Compose

**Files:**
- Create: `deploy/docker-compose.aiwelink-test.yml`

- [ ] **Step 1: Create the single-service Compose file**

The service must use `${SUB2API_IMAGE}`, `pull_policy: always`, `restart: unless-stopped`, `0.0.0.0:8081:8080`, the named `/app/data` volume, and the external `${ONEPANEL_NETWORK_NAME}` network. It must load `.env.aiwelink-test`, force the container listener to `0.0.0.0:8080`, and define this health check:

```yaml
healthcheck:
  test: ["CMD", "wget", "-q", "-T", "5", "-O", "/dev/null", "http://localhost:8080/health"]
  interval: 10s
  timeout: 5s
  retries: 12
  start_period: 240s
```

The file must not contain `build`, PostgreSQL, Redis, or literal credentials.

- [ ] **Step 2: Validate Compose structure without printing resolved secrets**

Run:

```powershell
docker compose --env-file deploy/.env.aiwelink-test -f deploy/docker-compose.aiwelink-test.yml config --quiet
docker compose --env-file deploy/.env.aiwelink-test -f deploy/docker-compose.aiwelink-test.yml config --images
```

Expected: exit 0 and only `docker.aiwelink.cc/sub2api-aiwelink-dev:dev-0c4864e109` is listed.

### Task 3: Materialize the Real Test Env

**Files:**
- Create (ignored): `deploy/.env.aiwelink-test`

- [ ] **Step 1: Copy the existing verified test credentials**

Copy `deploy/.env.develop` to `deploy/.env.aiwelink-test` without displaying its contents.

- [ ] **Step 2: Append deployment-only overrides**

Append the immutable image, release mode, `0.0.0.0:8081`, migration timeout, SSL mode, timezone, and 1Panel network values from Task 1. Later dotenv entries intentionally override duplicated development values.

- [ ] **Step 3: Verify required values without printing values**

Parse only key names and assert that all database, Redis, JWT/TOTP, Growth credential, Growth encryption, image, port, and network keys exist and are non-empty. Do not run `docker compose config` without `--quiet` because expanded output may expose secrets.

- [ ] **Step 4: Verify the remote image**

Run:

```powershell
docker buildx imagetools inspect docker.aiwelink.cc/sub2api-aiwelink-dev:dev-0c4864e109
```

Expected digest: `sha256:b7430ef4c59bc593ffe63ee91de4c5e32195ed70c9588fecfc670b849e73375c`.

### Task 4: Final Safety Review

**Files:**
- Verify: `.gitignore`
- Verify: `deploy/docker-compose.aiwelink-test.yml`
- Verify: `deploy/.env.aiwelink-test.example`
- Verify ignored: `deploy/.env.aiwelink-test`

- [ ] **Step 1: Scan tracked diffs for accidental credentials**

Run a key-name and placeholder scan against the tracked Compose/template diff. Confirm no real database host, password, service credential, encryption key, JWT secret, or TOTP key is staged.

- [ ] **Step 2: Verify files and worktree scope**

Run:

```powershell
git diff --check
git status --short
```

Expected: the pre-existing `.gitignore` modification remains preserved, the two older untracked Compose files remain untouched, and the real env does not appear because it is ignored.
