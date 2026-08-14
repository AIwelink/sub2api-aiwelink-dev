# AIWeLink Image Publish Summary Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Add an always-generated GitHub Actions Job Summary that reports the AIWeLink branch image version, tags, digest, and failure stage.

**Architecture:** Keep image publication in the existing `publish-image` job. Put deterministic Markdown rendering and failure-stage selection in a focused shell script, cover it with shell tests, and have the workflow pass step outcomes plus the registry digest into that script.

**Tech Stack:** GitHub Actions YAML, Bash, Docker Buildx, jq, existing shell contract tests.

---

## File Map

- Create `deploy/image-publish-summary.sh`: render the summary from GitHub context and step outcome environment variables.
- Create `deploy/tests/image-publish-summary-test.sh`: exercise dev success, main failure, and missing digest/version fallbacks.
- Modify `.github/workflows/backend-ci.yml`: assign step ids, resolve the published digest, invoke the summary with `if: always()`, and run the new shell test in CI.
- Modify `deploy/tests/ci-workflow-contract-test.sh`: statically require the summary wiring and preserve existing branch/tag gates.

### Task 1: Test and implement summary rendering

**Files:**
- Create: `deploy/tests/image-publish-summary-test.sh`
- Create: `deploy/image-publish-summary.sh`

- [ ] **Step 1: Write the failing shell test**

Create a temporary summary file, invoke `deploy/image-publish-summary.sh` with explicit environment variables, and assert exact fields. The success case must provide `PUSH_OUTCOME=success` and a valid `sha256:` digest; the failure case must provide `BUILD_OUTCOME=failure`, `PUSH_OUTCOME=skipped`, and `GITHUB_REF_NAME=main`.

```bash
run_summary() {
  local output=$1
  shift
  env GITHUB_STEP_SUMMARY="$output" \
    IMAGE_NAME=docker.aiwelink.cc/sub2api-aiwelink-dev \
    GITHUB_SHA=0123456789abcdef0123456789abcdef01234567 \
    CHECKOUT_OUTCOME=success BUILDX_OUTCOME=success VERSION_OUTCOME=success \
    META_OUTCOME=success IMAGE_OUTCOME=success SCAN_OUTCOME=success \
    CREDENTIALS_OUTCOME=success LOGIN_OUTCOME=success \
    "$@" bash "$SUMMARY_SCRIPT"
}

run_summary "$TMP_DIR/success.md" \
  GITHUB_REF_NAME=aiwelink-dev APPLICATION_VERSION=0.1.170-1 \
  UPSTREAM_VERSION=0.1.170 BUILD_OUTCOME=success PUSH_OUTCOME=success \
  IMAGE_DIGEST=sha256:aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa
assert_contains "$TMP_DIR/success.md" '| Publish status | `success` |'
assert_contains "$TMP_DIR/success.md" 'docker.aiwelink.cc/sub2api-aiwelink-dev:dev-0123456789ab'

run_summary "$TMP_DIR/failure.md" \
  GITHUB_REF_NAME=main APPLICATION_VERSION=0.1.170-1 \
  UPSTREAM_VERSION=0.1.170 BUILD_OUTCOME=failure PUSH_OUTCOME=skipped
assert_contains "$TMP_DIR/failure.md" '| Publish status | `failure` |'
assert_contains "$TMP_DIR/failure.md" '| Failure stage | `build` |'
assert_contains "$TMP_DIR/failure.md" 'docker.aiwelink.cc/sub2api-aiwelink-dev:latest'
assert_contains "$TMP_DIR/failure.md" '| Digest | `unavailable` |'
```

- [ ] **Step 2: Run the test and confirm the red state**

Run:

```bash
"C:/Program Files/Git/bin/bash.exe" deploy/tests/image-publish-summary-test.sh
```

Expected: non-zero exit because `deploy/image-publish-summary.sh` does not exist.

- [ ] **Step 3: Implement the minimal renderer**

Implement `deploy/image-publish-summary.sh` with `set -euo pipefail`. Compute the mutable and immutable tag from `GITHUB_REF_NAME` and the first 12 characters of `GITHUB_SHA`. Determine the first failed stage from this ordered mapping:

```bash
STEPS=(
  "checkout:${CHECKOUT_OUTCOME:-skipped}"
  "buildx:${BUILDX_OUTCOME:-skipped}"
  "version:${VERSION_OUTCOME:-skipped}"
  "metadata:${META_OUTCOME:-skipped}"
  "image-selection:${IMAGE_OUTCOME:-skipped}"
  "build:${BUILD_OUTCOME:-skipped}"
  "scan:${SCAN_OUTCOME:-skipped}"
  "credentials:${CREDENTIALS_OUTCOME:-skipped}"
  "login:${LOGIN_OUTCOME:-skipped}"
  "push:${PUSH_OUTCOME:-skipped}"
)
```

If push succeeded, report `success`; if any mapped outcome is `failure` or `cancelled`, report `failure` and that stage; otherwise report `skipped`. Read missing versions from `backend/cmd/server/VERSION` and `backend/cmd/server/UPSTREAM_VERSION`, falling back to `unknown`. Accept only `sha256:[0-9a-f]{64}` as a displayable digest. Append a Markdown table and pull commands to `$GITHUB_STEP_SUMMARY`.

- [ ] **Step 4: Run the renderer tests and shell syntax checks**

Run:

```bash
"C:/Program Files/Git/bin/bash.exe" -n deploy/image-publish-summary.sh
"C:/Program Files/Git/bin/bash.exe" -n deploy/tests/image-publish-summary-test.sh
"C:/Program Files/Git/bin/bash.exe" deploy/tests/image-publish-summary-test.sh
```

Expected: both syntax checks exit 0 and the behavior test prints `Image publish summary tests passed`.

- [ ] **Step 5: Commit the tested renderer**

```bash
git add deploy/image-publish-summary.sh deploy/tests/image-publish-summary-test.sh
git commit -m "test(ci): cover image publish summaries"
```

### Task 2: Wire the always-run summary into GitHub Actions

**Files:**
- Modify: `.github/workflows/backend-ci.yml`
- Modify: `deploy/tests/ci-workflow-contract-test.sh`

- [ ] **Step 1: Add failing workflow contract assertions**

Define `IMAGE_SUMMARY="$ROOT_DIR/deploy/image-publish-summary.sh"`, then require the new test invocation, stable step ids, digest lookup, `if: always()`, summary script call, and summary output destination:

```bash
assert_contains "$CI" '/bin/bash deploy/tests/image-publish-summary-test.sh'
assert_contains "$CI" 'id: push'
assert_contains "$CI" 'name: Resolve published image digest'
assert_contains "$CI" "--format '{{json .Manifest}}'"
assert_contains "$CI" 'name: Publish image summary'
assert_contains "$CI" 'if: always()'
assert_contains "$CI" 'bash deploy/image-publish-summary.sh'
assert_contains "$IMAGE_SUMMARY" 'GITHUB_STEP_SUMMARY'
```

- [ ] **Step 2: Run the contract and confirm the red state**

Run:

```bash
"C:/Program Files/Git/bin/bash.exe" deploy/tests/ci-workflow-contract-test.sh
```

Expected: non-zero exit reporting the first missing workflow string.

- [ ] **Step 3: Add workflow step ids and digest resolution**

Assign ids `checkout`, `buildx`, `build`, `scan`, `credentials`, `login`, and `push`. After a successful push, resolve the immutable image digest without failing publication if inspection is unavailable:

```yaml
- name: Resolve published image digest
  id: digest
  if: steps.push.outcome == 'success'
  continue-on-error: true
  env:
    IMMUTABLE_IMAGE: ${{ steps.image.outputs.scan-ref }}
  shell: bash
  run: |
    MANIFEST="$(docker buildx imagetools inspect "$IMMUTABLE_IMAGE" --format '{{json .Manifest}}')"
    DIGEST="$(jq -r '.digest // empty' <<<"$MANIFEST")"
    if ! [[ "$DIGEST" =~ ^sha256:[0-9a-f]{64}$ ]]; then
      echo "::error::Unable to resolve a valid digest for $IMMUTABLE_IMAGE"
      exit 1
    fi
    echo "digest=$DIGEST" >> "$GITHUB_OUTPUT"
```

- [ ] **Step 4: Add the always-run summary step and CI test invocation**

Append a step with `if: always()` and pass all step outcomes, version output, and digest output to `deploy/image-publish-summary.sh`. GitHub automatically supplies `GITHUB_STEP_SUMMARY`, `GITHUB_REF_NAME`, and `GITHUB_SHA`; `IMAGE_NAME` comes from the job environment. Add `/bin/bash deploy/tests/image-publish-summary-test.sh` to the existing `shell` job.

```yaml
- name: Publish image summary
  if: always()
  env:
    APPLICATION_VERSION: ${{ steps.version.outputs.version }}
    IMAGE_DIGEST: ${{ steps.digest.outputs.digest }}
    CHECKOUT_OUTCOME: ${{ steps.checkout.outcome }}
    BUILDX_OUTCOME: ${{ steps.buildx.outcome }}
    VERSION_OUTCOME: ${{ steps.version.outcome }}
    META_OUTCOME: ${{ steps.meta.outcome }}
    IMAGE_OUTCOME: ${{ steps.image.outcome }}
    BUILD_OUTCOME: ${{ steps.build.outcome }}
    SCAN_OUTCOME: ${{ steps.scan.outcome }}
    CREDENTIALS_OUTCOME: ${{ steps.credentials.outcome }}
    LOGIN_OUTCOME: ${{ steps.login.outcome }}
    PUSH_OUTCOME: ${{ steps.push.outcome }}
  shell: bash
  run: bash deploy/image-publish-summary.sh
```

- [ ] **Step 5: Run targeted verification**

Run:

```bash
"C:/Program Files/Git/bin/bash.exe" deploy/tests/image-publish-summary-test.sh
"C:/Program Files/Git/bin/bash.exe" deploy/tests/ci-workflow-contract-test.sh
git diff --check
```

Expected: both scripts pass and `git diff --check` exits 0.

- [ ] **Step 6: Commit the workflow integration**

```bash
git add .github/workflows/backend-ci.yml deploy/tests/ci-workflow-contract-test.sh
git commit -m "ci: summarize AIWeLink image publication"
```

### Task 3: Verify and publish the pull request

**Files:**
- Verify all files changed by Tasks 1 and 2.

- [ ] **Step 1: Run final local checks**

Run the two shell syntax checks, both targeted test scripts, YAML parsing with `python -c "import pathlib, yaml; yaml.safe_load(pathlib.Path('.github/workflows/backend-ci.yml').read_text(encoding='utf-8'))"`, `git diff --check`, and `git status --short --branch`. Expected: every executable check exits 0 and only intentional commits differ from `origin/aiwelink-dev`.

- [ ] **Step 2: Push the feature branch**

```bash
git push -u origin codex/image-publish-summary
```

- [ ] **Step 3: Create a ready PR targeting `aiwelink-dev`**

Create a PR titled `ci: summarize AIWeLink image publication`. The body must describe always-on success/failure summaries, displayed image metadata, digest fallback behavior, and local test evidence. Set base to `aiwelink-dev` and head to `codex/image-publish-summary`.

- [ ] **Step 4: Inspect PR checks**

Use GitHub Actions metadata to confirm the PR targets `aiwelink-dev`, the expected CI run exists, and `publish-image` is skipped for the PR event as designed. Report any running checks without claiming they passed.
