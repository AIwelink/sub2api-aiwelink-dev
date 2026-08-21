# CI Merge Optimizations Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Harden the AIWeLink merge CI against hung jobs and CodeQL environment drift without changing its triggers, required check name, or test scope.

**Architecture:** Keep `.github/workflows/backend-ci.yml` as the single validation workflow and `ci-gate` as its aggregate required check. Add conservative job-level timeouts, initialize Go before CodeQL, and preserve the gated post-merge image publisher. Main-branch protection and removal of the legacy publisher remain a follow-up migration after this workflow reaches `main`.

**Tech Stack:** GitHub Actions YAML, Bash contract tests, Git worktree.

---

### Task 1: Extend the CI workflow contract checks

**Files:**
- Modify: `deploy/tests/ci-workflow-contract-test.sh`

- [x] **Step 1: Add assertions for the required timeout and CodeQL ordering contracts**

Assert that `unit-tests`, `integration-tests`, `golangci-lint`, `codeql`, `ci-gate`, and `publish-image` expose the agreed timeout values, and assert that `Set up Go` appears before `Initialize CodeQL` while `Build Go database` remains after initialization.

- [x] **Step 2: Run the contract test against the unchanged workflow**

Run: `bash deploy/tests/ci-workflow-contract-test.sh`

Expected: FAIL because the current workflow lacks the new timeout/order assertions.

### Task 2: Apply minimal workflow hardening

**Files:**
- Modify: `.github/workflows/backend-ci.yml:46-257`

- [x] **Step 1: Add conservative job timeouts**

Use 15 minutes for unit tests, 20 for integration tests, 10 for growth/compose/release contracts, 15 for frontend checks/tests/security/repository scan, 35 for golangci-lint, 30 for CodeQL, 5 for `ci-gate`, and 45 for image publication. Keep all existing commands and triggers unchanged.

- [x] **Step 2: Move Go setup before CodeQL initialization**

Keep the Go setup conditional on the Go matrix entry, then initialize CodeQL, build the Go database, and analyze. Do not change the matrix languages or the `ci-gate` dependency list.

- [x] **Step 3: Run the contract test**

Run: `bash deploy/tests/ci-workflow-contract-test.sh`

Expected: PASS with `CI workflow contract checks passed`.

### Task 3: Verify the change and prepare handoff

**Files:**
- No additional source files.

- [x] **Step 1: Parse the workflow and inspect the diff**

Run: `python -c "import yaml; yaml.safe_load(open('.github/workflows/backend-ci.yml', encoding='utf-8')); print('YAML parsed')"` and `git diff --check`.

- [x] **Step 2: Confirm branch protection and migration boundaries**

Verify that `aiwelink-dev` still requires only `ci-gate`, and document that `main` remains on its existing required contexts until the new workflow is merged there. Do not call the branch-protection mutation in this change.

- [x] **Step 3: Review the final diff and commit**

Confirm that only the design/plan documents, contract test, and CI workflow changed, then commit with `ci: harden merge workflow timeouts`.
