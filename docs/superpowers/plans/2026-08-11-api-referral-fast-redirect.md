# API Referral Fast Redirect Implementation Plan

> For agentic workers: use superpowers:executing-plans to implement this plan task-by-task.

**Goal:** Add a backend-only /r/{code} entry point to Sub2 and make Traffic redirect API-domain referrals directly to the API registration page while preserving attribution and the existing homepage flow.

**Architecture:** Sub2 validates the referral code format locally and returns an immediate 302 to the public Traffic URL with fixed entry=api. It does not query a database, call Traffic, or create a cookie. Traffic remains the only authority for code lookup, visit recording, attribution cookie issuance, and final target selection; entry=api selects a fixed configured URL and never accepts a caller-controlled redirect.

**Tech Stack:** Go, Gin, Viper, Go test; Python 3.12, FastAPI, Pydantic Settings, SQLAlchemy, pytest, Ruff, mypy.

---

### Task 1: Confirm isolated worktrees and baselines

**Files:**
- No source changes.

- [x] Verify Sub2 is isolated at D:/Data/Codex 项目文件夹/sub2api-aiwelink-dev/.worktrees/growth-registration-binding on codex/api-referral-fast-redirect.
- [x] Verify Traffic is isolated at C:/Users/Achernar/AppData/Local/Temp/codex-worktrees/traffic-api-referral on codex/api-referral-api-target, based on origin/main.
- [x] Run the baseline commands:
  - In Sub2 backend: go test ./internal/config ./internal/server/routes
  - In Traffic: python -m uv run pytest tests/unit/test_config.py tests/unit/test_redirect_service.py tests/http/test_redirect.py -q
- [x] Expected baseline: Sub2 packages pass and Traffic reports 35 passed.

### Task 2: Add Sub2 configuration and route tests

**Files:**
- Modify: backend/internal/config/config.go
- Modify: backend/internal/config/growth_registration_test.go
- Create: backend/internal/server/routes/growth_referral.go
- Create: backend/internal/server/routes/growth_referral_test.go
- Modify: backend/internal/server/router.go
- Modify: backend/internal/web/embed_on.go
- Modify: backend/internal/web/embed_test.go

- [ ] Write failing configuration tests for the default and GROWTH_REGISTRATION_REFERRAL_BASE_URL environment value. Add a Load validation case for an enabled feature with an http URL.
- [ ] Run go test ./internal/config -run 'TestLoadGrowthRegistration' and confirm failure because ReferralBaseURL does not yet exist.
- [ ] Write failing route tests for valid uppercase normalization, exact Location https://aiwelink.cc/r/{code}?entry=api, Cache-Control no-store, invalid code, disabled feature, non-GET, discarded input query, and no dependency on a database or HTTP client.
- [ ] Write a web test proving /r/ is bypassed by the embedded SPA middleware.
- [ ] Run go test ./internal/server/routes ./internal/web -run 'GrowthReferral|EmbeddedFrontend.*Referral' and confirm the new tests fail for the missing implementation.

### Task 3: Implement Sub2 fast redirect minimally

**Files:**
- Modify: backend/internal/config/config.go
- Create/modify: backend/internal/server/routes/growth_referral.go
- Modify: backend/internal/server/router.go
- Modify: backend/internal/web/embed_on.go

- [ ] Add ReferralBaseURL to GrowthRegistrationConfig, register its environment key, trim it during normalization, and default it to https://aiwelink.cc/r.
- [ ] When growth registration is enabled, parse the base with net/url and require HTTPS, a non-empty host, no userinfo, no query, no fragment, and exact path /r. Return a configuration error before startup for invalid values.
- [ ] Add RegisterGrowthReferralRoutes(r *gin.Engine, cfg config.GrowthRegistrationConfig). Register only GET /r/:code and validate code with ^[a-hj-km-np-z2-9]{8}$. Normalize uppercase to lowercase.
- [ ] For an enabled valid code, build the target with url.URL and url.Values and return 302 with only entry=api and Cache-Control: no-store. For invalid code or disabled feature, return 302 to /register with the same cache policy.
- [ ] Never forward the request query, create a client, touch a repository, or write a cookie.
- [ ] Register this route after core logger/session/CORS/security/timing middleware but before embedded frontend middleware. Ensure the embedded frontend bypass predicate includes /r/ for defense in depth.
- [ ] Run gofmt on changed Go files and go test ./internal/config ./internal/server/routes ./internal/web.

### Task 4: Document Sub2 deployment configuration

**Files:**
- Modify: deploy/config.example.yaml
- Modify: deploy/.env.example
- Modify: deploy/.env.aiwelink-dev.example
- Modify: deploy/AIWELINK_GROWTH_REGISTRATION_CN.md

- [ ] Add growth_registration.referral_base_url: https://aiwelink.cc/r and GROWTH_REGISTRATION_REFERRAL_BASE_URL=https://aiwelink.cc/r.
- [ ] Explain that this is only the browser-visible Traffic entry, has no Sub2 database or HTTP dependency, must be HTTPS with exact /r path, and is separate from GROWTH_LOGIN_*.
- [ ] Document api.aiwelink.cc/r/{code} -> Sub2 302 -> Traffic /r/{code}?entry=api -> API registration, Traffic-first deployment, and private port 8300 with no reverse proxy.
- [ ] Run git diff --check.

### Task 5: Add Traffic API target tests

**Files:**
- Modify: src/aiwelink_growth/config.py
- Modify: tests/unit/test_config.py
- Modify: src/aiwelink_growth/public_gateway/service.py
- Modify: src/aiwelink_growth/public_gateway/router.py
- Modify: tests/unit/test_redirect_service.py
- Modify: tests/http/test_redirect.py

- [ ] Write failing Settings tests for API_REGISTRATION_URL parsing and production rejection of any value other than exact HTTPS https://api.aiwelink.cc/register, including query, fragment, host, port, and path changes.
- [ ] Write a pure target-selection test: missing entry and unknown entry use PUBLIC_HOMEPAGE_URL; entry=api uses API_REGISTRATION_URL.
- [ ] Add an HTTP test proving entry=api reaches the service and next=https://evil.example cannot change the target.
- [ ] Run python -m uv run pytest tests/unit/test_config.py tests/unit/test_redirect_service.py tests/http/test_redirect.py -q and confirm the new tests fail because the setting and selector do not exist.

### Task 6: Implement Traffic fixed API target

**Files:**
- Modify: src/aiwelink_growth/config.py
- Modify: src/aiwelink_growth/public_gateway/service.py
- Modify: src/aiwelink_growth/public_gateway/router.py

- [ ] Add api_registration_url with default https://api.aiwelink.cc/register. In production require that exact string; reject userinfo, query, fragment, non-HTTPS, another host, another port, and another path.
- [ ] Implement redirect_target(entry, settings) -> str returning only the API URL when entry == api, otherwise the homepage URL. Do not parse a URL from request input and do not add next, redirect, origin, or similar parameters.
- [ ] Add entry: str | None to RedirectRequestData. Read only request.query_params.get('entry') in the router and pass it to the service.
- [ ] Compute the selected target once at the beginning of RedirectService.redirect; pass it to _fallback for invalid_code, excluded, and database_error and use it for attribution_updated. Preserve cookie issuance, attribution writes, rate limits, and result labels.
- [ ] Run the focused tests, ruff format --check ., ruff check ., and mypy src through python -m uv run.

### Task 7: Document Traffic deployment configuration

**Files:**
- Modify: .env.example
- Modify: README.md

- [ ] Add API_REGISTRATION_URL=https://api.aiwelink.cc/register beside the homepage/fallback settings.
- [ ] Explain missing entry keeps the homepage flow, entry=api records attribution and redirects directly to API registration, unknown entries fall back to homepage, and port 8300 stays loopback/private.
- [ ] Run git diff --check.

### Task 8: Final verification and separate PRs

**Files:**
- All files listed in Tasks 2-7.

- [ ] Run Sub2:
  - go test ./internal/config ./internal/server/routes ./internal/web
  - go test ./internal/...
- [ ] Run Traffic:
  - python -m uv run pytest tests/unit/test_config.py tests/unit/test_redirect_service.py tests/http/test_redirect.py -q
  - python -m uv run ruff format --check .
  - python -m uv run ruff check .
  - python -m uv run mypy src
- [ ] Review git status -sb, git diff --stat, and git diff --check separately. Stage only feature files; never stage the dirty Traffic main checkout or .venv.
- [ ] Commit Sub2 as feat: add API referral fast redirect and Traffic as feat: route API referrals to registration.
- [ ] Push codex/api-referral-fast-redirect and create a PR to aiwelink-dev in AIwelink/sub2api-aiwelink-dev.
- [ ] Push codex/api-referral-api-target and create a PR to main in AIwelink/traffic-analysis.
- [ ] Do not merge either PR.
