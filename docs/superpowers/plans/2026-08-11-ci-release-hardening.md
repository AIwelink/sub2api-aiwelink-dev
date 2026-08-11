# AIWeLink CI and Image Release Hardening Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Make `ci-gate` the reliable quality boundary for Sub2 changes and ensure branch/release images cannot be published before the tested commit passes all required checks.

**Architecture:** Consolidate PR and protected-branch checks in `.github/workflows/backend-ci.yml`, keep scheduled security/canary concerns separate, and publish trusted branch images only from a job that depends on `ci-gate`. Reuse the existing growth recorder, encrypted outbox, worker, middleware, and referral route in one Fake Traffic contract test rather than adding a second integration implementation.

**Tech Stack:** GitHub Actions, Bash, Go 1.26.5, Node.js 24, pnpm 9, Vitest, CodeQL, Trivy 0.73.0, govulncheck 1.1.4, actionlint 1.7.12, Docker Buildx.

---

## File Map

- Create `deploy/tests/ci-workflow-contract-test.sh`: static behavioral contract for triggers, gate dependencies, image tags, versions, and release ownership.
- Create `backend/internal/service/growth_registration_contract_test.go`: in-process Fake Traffic test for recorder -> encrypted outbox -> worker -> HTTP bind.
- Create `deploy/tests/growth-public-canary.sh`: side-effect-bounded public `/r/{code}` and health probe.
- Create `.github/workflows/growth-public-canary.yml`: scheduled/manual public probe.
- Modify `.github/workflows/backend-ci.yml`: scoped triggers, concurrency, parallel checks, aggregate gate, and gated image publication.
- Modify `.github/workflows/security-scan.yml`: scheduled/manual deep scan with pinned tool versions and Node 24.
- Modify `.github/workflows/release.yml`: tested-main-commit validation, pre-push image scan, and exclusive semantic version tags.
- Modify `Makefile`: full frontend check/test targets and CI contract target.
- Modify `backend/Makefile`: focused growth contract target and pinned vulnerability target.
- Verify `Dockerfile`: keep the existing `node:24-alpine` builder and lock that major-version alignment in the CI contract.
- Modify `.github/audit-exceptions.yml`: remove expired exceptions that no longer match current audit output and replace `security@your-domain` where an active exception remains.
- Delete `.github/workflows/publish-aiwelink-dev-image.yml`: remove the independent pre-CI publisher.
- Delete `.github/workflows/cla.yml`: remove a workflow that can only skip in the AIWeLink repository.

### Task 1: Lock the CI policy in a failing contract test

**Files:**
- Create: `deploy/tests/ci-workflow-contract-test.sh`
- Modify: `Makefile`

- [ ] **Step 1: Write the failing workflow contract**

Create an executable Bash test with strict mode and these concrete assertions:

```bash
#!/usr/bin/env bash
set -euo pipefail

ROOT_DIR=$(cd "$(dirname "${BASH_SOURCE[0]}")/../.." && pwd)
CI="$ROOT_DIR/.github/workflows/backend-ci.yml"
SECURITY="$ROOT_DIR/.github/workflows/security-scan.yml"
RELEASE="$ROOT_DIR/.github/workflows/release.yml"
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

assert_contains "$CI" 'branches: [aiwelink-dev, main]'
assert_contains "$CI" 'cancel-in-progress: true'
assert_contains "$CI" 'ci-gate:'
assert_contains "$CI" 'publish-image:'
assert_contains "$CI" 'needs: ci-gate'
assert_contains "$CI" "github.event_name == 'push'"
assert_contains "$CI" 'type=raw,value=dev,enable='
assert_contains "$CI" 'type=raw,value=latest,enable='
assert_contains "$CI" 'type=sha,prefix=dev-,format=short,enable='
assert_contains "$CI" 'type=sha,prefix=main-,format=short,enable='
assert_not_contains "$CI" 'type=raw,value=${{ steps.version.outputs.version }}'
assert_not_contains "$SECURITY" 'node-version: '\''20'\'''
assert_contains "$SECURITY" "node-version: '24'"
assert_contains "$RELEASE" 'Verify successful ci-gate for release commit'
assert_not_contains "$RELEASE" 'docker.aiwelink.cc/sub2api-aiwelink-dev:latest'
assert_contains "$MAKEFILE" 'pnpm --dir frontend run test:run'
assert_contains "$MAKEFILE" 'pnpm --dir frontend run build'
assert_not_contains "$MAKEFILE" 'FRONTEND_CRITICAL_VITEST'
assert_contains "$ROOT_DIR/Dockerfile" 'ARG NODE_IMAGE=node:24-alpine'
test ! -e "$ROOT_DIR/.github/workflows/publish-aiwelink-dev-image.yml"
test ! -e "$ROOT_DIR/.github/workflows/cla.yml"

printf 'CI workflow contract checks passed\n'
```

Add this target:

```make
.PHONY: test-ci-contract
test-ci-contract:
	@bash deploy/tests/ci-workflow-contract-test.sh
```

- [ ] **Step 2: Run the contract and verify RED**

Run: `bash deploy/tests/ci-workflow-contract-test.sh`

Expected: FAIL on the first missing scoped branch trigger. This proves the test detects the current unsafe workflow.

- [ ] **Step 3: Commit the failing contract**

```bash
git add deploy/tests/ci-workflow-contract-test.sh Makefile
git commit -m "test(ci): define protected pipeline contract"
```

### Task 2: Add the Fake Traffic growth delivery contract

**Files:**
- Create: `backend/internal/service/growth_registration_contract_test.go`
- Modify: `backend/Makefile`

- [ ] **Step 1: Write the failing end-to-end service contract**

In package `service`, define a `growthContractRepository` that implements `GrowthRegistrationOutboxRepository`. `RecordSuccessfulRegistration` must immediately enqueue a `GrowthRegistrationOutboxEvent`, `Claim` must lease the queued event, and the three transition methods must report through buffered channels. Start an `httptest.Server` as Fake Traffic and assert the real worker sends:

```go
type growthContractRequest struct {
	Authorization string
	RequestID     string
	Payload       growthRegistrationPayload
}

func TestGrowthRegistrationContractDeliversEncryptedSessionToTraffic(t *testing.T) {
	requests := make(chan growthContractRequest, 1)
	traffic := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var payload growthRegistrationPayload
		require.NoError(t, json.NewDecoder(r.Body).Decode(&payload))
		requests <- growthContractRequest{
			Authorization: r.Header.Get("Authorization"),
			RequestID:     r.Header.Get("X-Request-ID"),
			Payload:       payload,
		}
		w.Header().Set("X-Request-ID", "traffic-request")
		w.WriteHeader(http.StatusNoContent)
	}))
	defer traffic.Close()

	repository := newGrowthContractRepository()
	cipher, err := NewGrowthRegistrationCipher(growthRegistrationTestKey)
	require.NoError(t, err)
	recorder, err := NewGrowthRegistrationRecorder(repository, cipher, "aiwelink")
	require.NoError(t, err)
	worker, err := NewGrowthRegistrationWorker(repository, cipher, GrowthRegistrationWorkerOptions{
		Endpoint:          traffic.URL,
		ServiceCredential: "contract-secret",
		PollInterval:      5 * time.Millisecond,
		RetryDelay:        func(int) time.Duration { return time.Millisecond },
	})
	require.NoError(t, err)
	worker.Start()
	t.Cleanup(worker.Stop)

	registeredAt := time.Date(2026, 8, 11, 8, 0, 0, 0, time.UTC)
	ctx := WithGrowthRegistrationSession(context.Background(), "growth-session")
	require.NoError(t, recorder.RecordSuccessfulRegistration(ctx, &User{ID: 42, CreatedAt: registeredAt}))

	request := requireReceive(t, requests)
	require.Equal(t, "Service contract-secret", request.Authorization)
	require.NotEmpty(t, request.RequestID)
	require.Equal(t, "aiwelink", request.Payload.SiteID)
	require.Equal(t, "42", request.Payload.ExternalUserID)
	require.Equal(t, "growth-session", *request.Payload.GrowthSession)
	requireReceive(t, repository.delivered)
}
```

Add a second test where Fake Traffic returns `503`; assert `RecordSuccessfulRegistration` still succeeds quickly and the repository receives a retry transition instead of a delivered transition. The repository stores ciphertext only and decrypts exclusively through the production worker.

- [ ] **Step 2: Run the focused test and verify RED**

Run: `go test -tags=unit ./internal/service -run '^TestGrowthRegistrationContract' -count=1`

Expected: FAIL until the repository test harness and expected transition behavior are fully implemented.

- [ ] **Step 3: Complete the test harness without changing production behavior**

Implement all `GrowthRegistrationOutboxRepository` methods with a mutex, monotonically increasing `OutboxID`, and buffered `delivered`, `retried`, and `deadLettered` channels. Use a generic helper:

```go
func requireReceive[T any](t *testing.T, channel <-chan T) T {
	t.Helper()
	select {
	case value := <-channel:
		return value
	case <-time.After(2 * time.Second):
		t.Fatal("timed out waiting for growth contract event")
		var zero T
		return zero
	}
}
```

- [ ] **Step 4: Add and run the Make target**

```make
.PHONY: test-growth-contract
test-growth-contract:
	go test -tags=unit ./internal/service ./internal/server/routes ./internal/server/middleware \
		-run 'GrowthRegistration|GrowthReferral|AuthService_Register_RecordsGrowth' -count=1
```

Run: `make test-growth-contract`

Expected: PASS, including Fake Traffic delivery, retry, referral redirect, cookie capture, and registration fail-open tests.

- [ ] **Step 5: Commit**

```bash
git add backend/internal/service/growth_registration_contract_test.go backend/Makefile
git commit -m "test(growth): gate registration delivery contract"
```

### Task 3: Run complete frontend and parallel backend checks

**Files:**
- Modify: `Makefile`
- Modify: `.github/workflows/backend-ci.yml`

- [ ] **Step 1: Replace the six-test allowlist**

Replace `FRONTEND_CRITICAL_VITEST` and the old targets with:

```make
.PHONY: test-frontend test-frontend-checks
test-frontend:
	@pnpm --dir frontend run test:run

test-frontend-checks:
	@pnpm --dir frontend run lint:check
	@pnpm --dir frontend run typecheck
	@pnpm --dir frontend run build
```

- [ ] **Step 2: Rewrite CI triggers and concurrency**

Use this event boundary:

```yaml
on:
  push:
    branches: [aiwelink-dev, main]
  pull_request:
    branches: [aiwelink-dev, main]
  workflow_dispatch:

concurrency:
  group: ci-${{ github.event.pull_request.number || github.ref }}
  cancel-in-progress: true
```

Keep `shell` and `compose`. Split the existing `test` job into `unit-tests` and `integration-tests`, each using `actions/setup-go@v6` with `go-version-file: backend/go.mod`. Remove the duplicated literal `grep go1.26.5` checks. Add `growth-contract` running `make test-growth-contract` in `backend`.

- [ ] **Step 3: Split frontend checks from full tests with Node 24**

Create `frontend-checks` and `frontend-tests`. Both install with pnpm 9 and Node 24; one runs `make test-frontend-checks`, the other runs `make test-frontend`.

```yaml
- uses: pnpm/action-setup@v6
  with:
    version: 9
- uses: actions/setup-node@v6
  with:
    node-version: '24'
    cache: pnpm
    cache-dependency-path: frontend/pnpm-lock.yaml
```

- [ ] **Step 4: Run local frontend checks and tests**

Run with the bundled Node 24 runtime on PATH:

```powershell
$env:PATH='C:\Users\Achernar\.cache\codex-runtimes\codex-primary-runtime\dependencies\node\bin;C:\Users\Achernar\.cache\codex-runtimes\codex-primary-runtime\dependencies\bin\fallback;' + $env:PATH
pnpm --dir frontend install --frozen-lockfile
make test-frontend-checks
make test-frontend
```

Expected: lint, typecheck, production build, and all Vitest files PASS.

- [ ] **Step 5: Commit**

```bash
git add Makefile .github/workflows/backend-ci.yml
git commit -m "ci: run complete protected branch checks"
```

### Task 4: Add security jobs and aggregate `ci-gate`

**Files:**
- Modify: `.github/workflows/backend-ci.yml`
- Modify: `.github/workflows/security-scan.yml`
- Modify: `backend/Makefile`
- Modify: `.github/audit-exceptions.yml`

- [ ] **Step 1: Pin reusable security commands**

Add:

```make
.PHONY: security
security:
	go run golang.org/x/vuln/cmd/govulncheck@v1.1.4 ./...
```

Run `pnpm audit --prod --audit-level=high --json` and the existing exception checker. Remove only expired exception records absent from current audit output; any active exception must retain advisory, severity, mitigation, expiry, and a real AIWeLink owner.

- [ ] **Step 2: Add required security jobs to CI**

Add `backend-security`, `frontend-security`, `repository-scan`, and a CodeQL matrix for `go` and `javascript-typescript`. `repository-scan` uses Trivy 0.73.0 with `scan-type: fs`, scanners `vuln,secret,misconfig`, severities `HIGH,CRITICAL`, and `exit-code: 1`. CodeQL uses `github/codeql-action/init@v4`, `autobuild@v4`, and `analyze@v4` with workflow permission `security-events: write`.

- [ ] **Step 3: Add the stable aggregate gate**

```yaml
ci-gate:
  if: always()
  runs-on: ubuntu-latest
  needs:
    - shell
    - compose
    - unit-tests
    - integration-tests
    - growth-contract
    - frontend-checks
    - frontend-tests
    - golangci-lint
    - backend-security
    - frontend-security
    - repository-scan
    - codeql
  steps:
    - name: Require every CI dependency
      env:
        NEEDS_JSON: ${{ toJSON(needs) }}
      run: jq -e 'all(.[]; .result == "success")' <<<"$NEEDS_JSON"
```

- [ ] **Step 4: Restrict the standalone security workflow**

Change `.github/workflows/security-scan.yml` to `schedule` and `workflow_dispatch` only, add concurrency, Node 24, govulncheck 1.1.4, and the same audit/repository scan semantics. It no longer duplicates every push and PR.

- [ ] **Step 5: Run the static contract and security exception tests**

Run:

```bash
bash deploy/tests/ci-workflow-contract-test.sh
python tools/check_pnpm_audit_exceptions.py --audit frontend/audit.json --exceptions .github/audit-exceptions.yml
```

Expected: the workflow contract may still fail only on pending publish/release assertions; audit exceptions PASS.

- [ ] **Step 6: Commit**

```bash
git add .github/workflows/backend-ci.yml .github/workflows/security-scan.yml backend/Makefile .github/audit-exceptions.yml
git commit -m "ci(security): aggregate required scans"
```

### Task 5: Publish branch images only after `ci-gate`

**Files:**
- Modify: `.github/workflows/backend-ci.yml`
- Delete: `.github/workflows/publish-aiwelink-dev-image.yml`
- Delete: `.github/workflows/cla.yml`

- [ ] **Step 1: Delete independent/no-op workflows and verify RED advances**

Remove the standalone image publisher and AIWeLink-inapplicable CLA workflow. Run `bash deploy/tests/ci-workflow-contract-test.sh`; expected failure moves to missing gated publish tags or release validation, never the deleted-file assertions.

- [ ] **Step 2: Add `publish-image` after the gate**

The job must use:

```yaml
publish-image:
  needs: ci-gate
  if: >-
    github.event_name == 'push' &&
    (github.ref_name == 'aiwelink-dev' || github.ref_name == 'main')
  runs-on: ubuntu-latest
```

Generate only these tags:

```yaml
tags: |
  type=raw,value=dev,enable=${{ github.ref_name == 'aiwelink-dev' }}
  type=raw,value=latest,enable=${{ github.ref_name == 'main' }}
  type=sha,prefix=dev-,format=short,enable=${{ github.ref_name == 'aiwelink-dev' }}
  type=sha,prefix=main-,format=short,enable=${{ github.ref_name == 'main' }}
```

Build `linux/amd64` with `load: true` and a local immutable SHA tag, scan it with Trivy 0.73.0 before registry login, then push exactly the newline-separated tags emitted by `docker/metadata-action`. Never use `--all-tags`.

- [ ] **Step 3: Run contract and actionlint**

Run:

```bash
bash deploy/tests/ci-workflow-contract-test.sh
cd backend && go run github.com/rhysd/actionlint/cmd/actionlint@v1.7.12 ../../.github/workflows/*.yml
```

Expected: contract now fails only on release checks; actionlint reports no workflow syntax/expression errors.

- [ ] **Step 4: Commit**

```bash
git add .github/workflows/backend-ci.yml .github/workflows/publish-aiwelink-dev-image.yml .github/workflows/cla.yml
git commit -m "ci(image): publish only tested branch commits"
```

### Task 6: Make Release the sole semantic version publisher

**Files:**
- Modify: `.github/workflows/release.yml`
- Modify: `deploy/tests/ci-workflow-contract-test.sh`
- Test: `deploy/tests/install-github-token-test.sh`

- [ ] **Step 1: Extend the failing release contract**

Assert the release workflow contains `Verify release commit belongs to main`, `Verify successful ci-gate for release commit`, and a pre-push `Scan AIWeLink release image`. Assert the private registry block contains the version tag but not `:latest`.

- [ ] **Step 2: Resolve and validate the checked-out release SHA**

After checkout, write `git rev-parse HEAD` to step output, fetch `origin/main`, and run `git merge-base --is-ancestor "$RELEASE_SHA" origin/main`. Add `checks: read` permission.

Use `actions/github-script@v8` with the resolved SHA to call `github.rest.checks.listForRef`, then fail unless a check run named `ci-gate` has `conclusion === 'success'` for that SHA.

- [ ] **Step 3: Scan before semantic tag push**

Build/load the amd64 private-registry image under a temporary local tag, scan it at `HIGH,CRITICAL`, then perform the existing multi-architecture push with only:

```yaml
tags: |
  docker.aiwelink.cc/sub2api-aiwelink-dev:${{ steps.release_metadata.outputs.version }}
```

Release remains the only workflow allowed to write a semantic version tag. `main` owns `latest`.

- [ ] **Step 4: Run release-focused tests**

Run:

```bash
bash deploy/tests/ci-workflow-contract-test.sh
bash deploy/tests/install-github-token-test.sh
```

Expected: both PASS.

- [ ] **Step 5: Commit**

```bash
git add .github/workflows/release.yml deploy/tests/ci-workflow-contract-test.sh
git commit -m "ci(release): require tested main commit"
```

### Task 7: Add the bounded public growth canary

**Files:**
- Create: `deploy/tests/growth-public-canary.sh`
- Create: `.github/workflows/growth-public-canary.yml`
- Modify: `deploy/tests/ci-workflow-contract-test.sh`

- [ ] **Step 1: Write local validation tests first**

Add a `--self-test` mode that starts a local fixture server or accepts `GROWTH_CANARY_BASE_URL` and verifies failure for an invalid code, missing secure/HttpOnly cookie, non-redirect response, and unexpected final host. The script must use temp files with a cleanup trap and must never print cookie jar contents.

- [ ] **Step 2: Implement the read-only/public request sequence**

Require `GROWTH_CANARY_REFERRAL_CODE` to match `^[a-hj-km-np-z2-9]{8}$`. Check `${BASE_URL}/health`, request `${BASE_URL}/r/${CODE}` without automatic POSTs, follow redirects with a temporary cookie jar, and assert:

- first response is `302`;
- `Cache-Control` contains `no-store`;
- final host is `api.aiwelink.cc` or the explicitly configured fixture host;
- cookie jar contains `#HttpOnly_.aiwelink.cc`, `awl_growth_sid`, and the secure flag;
- no registration, bind, login, or payment endpoint is requested.

- [ ] **Step 3: Add scheduled/manual workflow**

Run every 30 minutes and on `workflow_dispatch`, with `contents: read`, five-minute timeout, concurrency cancellation, `GROWTH_CANARY_BASE_URL=https://api.aiwelink.cc`, and referral code from `secrets.GROWTH_CANARY_REFERRAL_CODE`.

- [ ] **Step 4: Run self-test and workflow contract**

Run:

```bash
bash deploy/tests/growth-public-canary.sh --self-test
bash deploy/tests/ci-workflow-contract-test.sh
```

Expected: PASS and no cookie value in stdout/stderr.

- [ ] **Step 5: Commit**

```bash
git add deploy/tests/growth-public-canary.sh .github/workflows/growth-public-canary.yml deploy/tests/ci-workflow-contract-test.sh
git commit -m "ci(growth): monitor public referral path"
```

### Task 8: Verify, publish the PR, and migrate safe repository settings

**Files:**
- Modify: `docs/superpowers/plans/2026-08-11-ci-release-hardening.md` checkbox state only if tracking is retained
- GitHub settings: AIwelink/sub2api-aiwelink-dev Actions security and branch protection

- [ ] **Step 1: Run full fresh local verification**

Run:

```bash
bash deploy/tests/ci-workflow-contract-test.sh
bash deploy/tests/growth-public-canary.sh --self-test
bash deploy/tests/aiwelink-growth-deployment-test.sh
bash deploy/tests/install-github-token-test.sh
cd backend && go test -tags=unit ./...
cd backend && go test -tags=integration ./...
cd backend && go run github.com/rhysd/actionlint/cmd/actionlint@v1.7.12 ../../.github/workflows/*.yml
pnpm --dir frontend run lint:check
pnpm --dir frontend run typecheck
pnpm --dir frontend run test:run
pnpm --dir frontend run build
git diff --check origin/aiwelink-dev...HEAD
```

Expected: every command exits 0. Record Docker daemon unavailability explicitly if local image verification cannot run; do not claim container verification until GitHub Actions completes.

- [ ] **Step 2: Request focused code review and fix all Critical/Important findings**

Review `origin/aiwelink-dev...HEAD` against the design, especially GitHub expression evaluation, secret exposure, branch/tag ownership, and canary side effects. Re-run affected tests after every fix.

- [ ] **Step 3: Push and create a PR only to `aiwelink-dev`**

```bash
git push -u origin codex/ci-release-hardening
gh pr create --repo AIwelink/sub2api-aiwelink-dev \
  --base aiwelink-dev \
  --head codex/ci-release-hardening \
  --title "ci: gate AIWeLink image publication" \
  --body "$(printf '%s\n' \
    '## Summary' \
    '- require a single ci-gate before protected branch images publish' \
    '- isolate dev, main, and release image tags' \
    '- run complete frontend, growth contract, and security checks' \
    '' \
    '## Verification' \
    '- local workflow, shell, Go, and frontend checks listed in the implementation plan' \
    '- GitHub Actions must pass before branch-protection migration')"
```

Do not create or merge a `main` PR.

- [ ] **Step 4: Wait for PR checks and validate published behavior**

Confirm the PR produces one CI run, no image publish job, all complete frontend tests, growth contract, CodeQL/Trivy, and a successful `ci-gate`. If a check fails, inspect logs and fix on the feature branch.

- [ ] **Step 5: Enable native secret scanning where supported**

Use the GitHub repository API to enable `secret_scanning`, `secret_scanning_push_protection`, and validity checks. Read back repository settings and report unsupported flags rather than silently ignoring them.

- [ ] **Step 6: Transition branch protection without deadlock**

After `ci-gate` exists on the PR, update `aiwelink-dev` required checks to `ci-gate` while retaining strict up-to-date checks, review requirements, admin enforcement, and conversation resolution. Add `compose` to `main`'s current required contexts immediately; defer replacing main contexts with `ci-gate` until the workflow reaches main through a user-authorized `aiwelink-dev -> main` PR.

- [ ] **Step 7: Final evidence**

Report the PR URL, exact successful check run URLs, produced/withheld image behavior, branch protection read-back, native security setting read-back, local environment limitations, and the fact that main was not merged or modified.
