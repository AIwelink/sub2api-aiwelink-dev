# API Referral Homepage Redirect Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Keep referral attribution for `api.aiwelink.cc/r/{code}`, then return the browser to the API homepage so its existing `HomepageIntro` animation plays instead of opening registration directly.

**Architecture:** Sub2 continues to return an immediate backend `302` to the public Traffic route with the closed `entry=api` selector. Traffic remains responsible for attribution and the parent-domain cookie, but `entry=api` selects the fixed API homepage `https://api.aiwelink.cc/`. A cross-domain round trip creates a new API document, so the existing homepage intro runs without a new query flag or frontend component.

**Tech Stack:** Go, Gin, Go test; Python 3.12, FastAPI, Pydantic Settings, pytest, Ruff, mypy; Vue 3 and Vitest for existing homepage animation verification.

---

### Task 1: Change Traffic's fixed API target

**Files:**
- Modify: `tests/unit/test_config.py`
- Modify: `tests/unit/test_redirect_service.py`
- Modify: `tests/http/test_redirect.py`
- Modify: `src/aiwelink_growth/config.py`
- Modify: `src/aiwelink_growth/public_gateway/service.py`
- Modify: `.env.example`

- [ ] **Step 1: Write failing target tests**

Rename the setting contract from `api_registration_url` / `API_REGISTRATION_URL` to `api_homepage_url` / `API_HOMEPAGE_URL`. Assert that `entry="api"` returns `https://api.aiwelink.cc/`, while missing, unknown, and URL-shaped entries still return the public homepage.

```python
assert select_redirect_target(settings, "api") == "https://api.aiwelink.cc/"
assert select_redirect_target(settings, "https://evil.example") == "https://aiwelink.cc/"
```

- [ ] **Step 2: Verify RED**

Run:

```powershell
.venv\Scripts\python.exe -m pytest tests/unit/test_config.py tests/unit/test_redirect_service.py tests/http/test_redirect.py -q
```

Expected: failures still report `/register` or the missing `api_homepage_url` setting.

- [ ] **Step 3: Implement the fixed homepage target**

Add the Pydantic setting with the exact production value and select it only for the closed `api` entry:

```python
api_homepage_url: AnyHttpUrl = AnyHttpUrl("https://api.aiwelink.cc/")

def select_redirect_target(settings: Settings, entry: str | None) -> str:
    if entry == "api":
        return str(settings.api_homepage_url)
    return str(settings.public_homepage_url)
```

Production validation must reject non-HTTPS, userinfo, alternate hosts, ports, paths, query strings, and fragments by requiring the exact string `https://api.aiwelink.cc/`.

- [ ] **Step 4: Verify GREEN and static checks**

Run the focused pytest command, then:

```powershell
.venv\Scripts\python.exe -m ruff check .
.venv\Scripts\python.exe -m mypy
```

Expected: focused tests pass, Ruff reports `All checks passed!`, and mypy reports no issues.

### Task 2: Make Sub2 referral fallbacks return home

**Files:**
- Modify: `backend/internal/server/routes/growth_referral_test.go`
- Modify: `backend/internal/server/routes/growth_referral.go`

- [ ] **Step 1: Write failing fallback tests**

Change the invalid-code and disabled-feature assertions from `/register` to `/`:

```go
require.Equal(t, "/", rec.Header().Get("Location"))
```

- [ ] **Step 2: Verify RED**

Run:

```powershell
go test -count=1 ./internal/server/routes -run GrowthReferral
```

Expected: the two fallback tests fail because the current handler returns `/register`.

- [ ] **Step 3: Implement the homepage fallback**

Return the local homepage for both disabled and invalid referrals:

```go
c.Redirect(http.StatusFound, "/")
```

Keep valid-code normalization, `entry=api`, `Cache-Control: no-store`, and the no-database/no-HTTP-client behavior unchanged.

- [ ] **Step 4: Verify GREEN**

Run:

```powershell
gofmt -w internal/server/routes/growth_referral.go internal/server/routes/growth_referral_test.go
go test -count=1 ./internal/config ./internal/server/routes ./internal/web
```

Expected: all three packages pass.

### Task 3: Align deployment documentation

**Files:**
- Modify: `deploy/AIWELINK_GROWTH_REGISTRATION_CN.md`
- Modify: `docs/superpowers/specs/2026-08-11-api-referral-fast-redirect-design.md` only if implementation exposes a contradiction
- Modify: `.env.example` in Traffic as covered by Task 1

- [ ] Replace Traffic's old registration target name/value with `API_HOMEPAGE_URL=https://api.aiwelink.cc/`.
- [ ] Document the chain as `API /r` -> Traffic attribution -> API homepage -> existing `HomepageIntro`.
- [ ] State that no referral-specific frontend state, animation component, or query parameter is added.
- [ ] Run `git diff --check` in both worktrees.

### Task 4: Verify the existing homepage animation contract

**Files:**
- No frontend production changes expected.
- Verify: `frontend/src/composables/__tests__/useHomepageIntro.spec.ts`
- Verify: `frontend/src/components/home/__tests__/AIWeLinkHome.spec.ts`

- [ ] Run:

```powershell
pnpm exec vitest run src/composables/__tests__/useHomepageIntro.spec.ts src/components/home/__tests__/AIWeLinkHome.spec.ts
```

Expected: the first mount starts at `preparing`, proceeds through `composing` and `revealing`, and reaches `ready`. Because the Traffic redirect loads a new API document, this existing first-mount behavior supplies the requested animation.

### Task 5: Commit, push, and update existing Draft PRs

**Files:**
- All files changed in Tasks 1-3.

- [ ] Run Traffic unit/http/contract tests and Sub2 affected Go tests.
- [ ] Confirm each worktree is clean except for the intended files and run `git diff --check`.
- [ ] Commit Traffic with `fix: return API referrals to homepage` and push `codex/api-referral-api-target`.
- [ ] Commit Sub2 with `fix: return referral fallbacks to homepage` and push `codex/api-referral-fast-redirect`.
- [ ] Confirm Traffic PR #1 still targets `main` and Sub2 PR #20 still targets `aiwelink-dev`.
- [ ] Keep both PRs unmerged; do not modify Sub2 `main`.
