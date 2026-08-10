# AIWeLink Versioning Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Implement AIWeLink versions such as `0.1.170-1` and `0.1.170-2.4`, expose the official Sub2API baseline, and make updates, releases, and containers use AIWeLink artifacts exclusively.

**Architecture:** A focused Go package owns strict parsing, validation, and comparison. Build metadata carries both the full AIWeLink version and its upstream baseline through existing dependency injection. The update service filters AIWeLink GitHub Releases with the same parser, while build scripts and CI reject inconsistent files or tags before artifacts are produced.

**Tech Stack:** Go 1.26, Gin, Vue 3, TypeScript, Vitest, shell scripts, GitHub Actions, GoReleaser, Docker Buildx.

---

### Task 1: Version model and committed metadata

**Files:**
- Create: `backend/internal/versioninfo/version.go`
- Create: `backend/internal/versioninfo/version_test.go`
- Create: `backend/cmd/server/UPSTREAM_VERSION`
- Modify: `backend/cmd/server/VERSION`

- [ ] **Step 1: Write failing parser and comparison tests**

Add table tests requiring `Parse("0.1.170-1")` and `Parse("0.1.170-2.4")` to succeed, rejecting official-only, zero revision, empty segments, signs, and suffix text. Add comparisons proving `0.1.170-2.4 > 0.1.170-2`, `0.1.170-2.4 < 0.1.170-3`, and `0.1.171-1 > 0.1.170-99`.

- [ ] **Step 2: Run the focused tests and verify failure**

Run: `go test -tags=unit ./internal/versioninfo`

Expected: FAIL because the package implementation does not exist.

- [ ] **Step 3: Implement strict parsing and validation**

Create an immutable parsed representation with `Upstream() string`, `String() string`, `Compare(a, b string) (int, error)`, and `Validate(full, upstream string) error`. Parse the upstream triplet and every positive integer revision component with `strconv.Atoi`; do not silently coerce malformed values to zero.

- [ ] **Step 4: Set initial committed versions**

Set `VERSION` to `0.1.170-1` and `UPSTREAM_VERSION` to `0.1.170`.

- [ ] **Step 5: Run focused tests**

Run: `go test -tags=unit ./internal/versioninfo`

Expected: PASS.

### Task 2: Build metadata and APIs

**Files:**
- Modify: `backend/cmd/server/main.go`
- Modify: `backend/cmd/server/wire.go`
- Modify: `backend/cmd/server/wire_gen.go`
- Modify: `backend/cmd/server/wire_gen_test.go`
- Modify: `backend/internal/handler/handler.go`
- Modify: `backend/internal/handler/wire.go`
- Modify: `backend/internal/handler/setting_handler.go`
- Modify: `backend/internal/handler/dto/settings.go`
- Modify: `backend/internal/handler/setting_handler_public_test.go`
- Modify: `backend/internal/handler/admin/system_handler.go`
- Modify: `backend/internal/handler/admin/system_handler_test.go`
- Modify: `backend/internal/service/wire.go`

- [ ] **Step 1: Add failing metadata propagation and response tests**

Require `BuildInfo{Version: "0.1.170-1", UpstreamVersion: "0.1.170"}` to survive handler-to-service conversion. Require public settings and admin version/update JSON to contain `upstream_version: "0.1.170"` without changing existing version keys.

- [ ] **Step 2: Run the focused backend tests and verify failure**

Run: `go test -tags=unit ./cmd/server ./internal/handler ./internal/handler/admin`

Expected: FAIL on missing `UpstreamVersion` fields.

- [ ] **Step 3: Embed and propagate the upstream baseline**

Embed `UPSTREAM_VERSION` beside `VERSION`, add linker variable `UpstreamVersion`, and pass it through both BuildInfo structs, Wire providers, setting handler, and update service. Change `--version` output to identify `AIWeLink <version> (based on Sub2API <upstream>)`.

- [ ] **Step 4: Expose additive API fields**

Add `upstream_version` to `dto.PublicSettings`, `/admin/system/version`, and `service.UpdateInfo`. Preserve `version`, `current_version`, and all existing response fields.

- [ ] **Step 5: Run focused tests**

Run: `go test -tags=unit ./cmd/server ./internal/handler ./internal/handler/admin`

Expected: PASS.

### Task 3: AIWeLink update and rollback selection

**Files:**
- Modify: `backend/internal/service/update_service.go`
- Modify: `backend/internal/service/update_service_test.go`

- [ ] **Step 1: Add failing update-source and filtering tests**

Record the repository passed to the GitHub client and require `AIwelink/sub2api-aiwelink-dev`. Add unordered releases containing valid AIWeLink tags, official-only tags, malformed tags, drafts, and prereleases. Require update selection to choose the highest valid stable AIWeLink version and rollback selection to retain only older valid AIWeLink versions.

- [ ] **Step 2: Add failing multi-part comparison tests**

Require comparison of all upstream and AIWeLink revision components, including `0.1.170-2.4`, and require malformed values to be excluded rather than treated as `0.0.0`.

- [ ] **Step 3: Run update service tests and verify failure**

Run: `go test -tags=unit ./internal/service -run UpdateService`

Expected: FAIL because the service still queries the official repository and only compares three numeric components.

- [ ] **Step 4: Implement AIWeLink release selection**

Replace the repository constant, fetch recent releases for update checks, filter through `versioninfo.Parse`, sort with strict comparison, and construct update details from the highest valid stable release. Use the same validation path for rollback allowlists.

- [ ] **Step 5: Run update service tests**

Run: `go test -tags=unit ./internal/service -run UpdateService`

Expected: PASS.

### Task 4: Frontend version presentation

**Files:**
- Modify: `frontend/src/types/index.ts`
- Modify: `frontend/src/api/admin/system.ts`
- Modify: `frontend/src/stores/app.ts`
- Modify: `frontend/src/stores/__tests__/app.spec.ts`
- Modify: `frontend/src/components/common/VersionBadge.vue`
- Create: `frontend/src/components/common/__tests__/VersionBadge.spec.ts`
- Modify: `frontend/src/i18n/locales/zh/misc.ts`
- Modify: `frontend/src/i18n/locales/en/misc.ts`

- [ ] **Step 1: Add failing store and component tests**

Require the app store to retain `upstream_version`, the badge to render `AIWeLink v0.1.170-1`, and the expanded admin panel to render the localized Sub2API baseline.

- [ ] **Step 2: Run focused frontend tests and verify failure**

Run: `pnpm run test:run -- src/stores/__tests__/app.spec.ts src/components/common/__tests__/VersionBadge.spec.ts`

Expected: FAIL on missing upstream state and display text.

- [ ] **Step 3: Implement additive frontend metadata**

Add optional `upstream_version` to public settings for backward compatibility and required `upstream_version` to AIWeLink update responses. Store it alongside the existing cached version state and render it in the current version area. Change GitHub and Docker constants to `AIwelink/sub2api-aiwelink-dev` and `docker.aiwelink.cc/sub2api-aiwelink-dev`.

- [ ] **Step 4: Add Chinese and English labels**

Use `基于 Sub2API v{version}` and `Based on Sub2API v{version}` without changing existing update-control wording.

- [ ] **Step 5: Run focused frontend tests**

Run: `pnpm run test:run -- src/stores/__tests__/app.spec.ts src/components/common/__tests__/VersionBadge.spec.ts`

Expected: PASS.

### Task 5: Build and release validation

**Files:**
- Create: `backend/scripts/validate-version.sh`
- Modify: `backend/scripts/resolve-version.sh`
- Modify: `backend/Makefile`
- Modify: `backend/Dockerfile`
- Modify: `deploy/Dockerfile`
- Modify: `Dockerfile`
- Modify: `.goreleaser.yaml`
- Modify: `.goreleaser.simple.yaml`
- Modify: `.github/workflows/release.yml`

- [ ] **Step 1: Add strict shell validation**

Validate the two committed files and an optional tag argument. Reject mismatched baselines and tags before returning success. Make `resolve-version.sh` accept exact AIWeLink tags only and invoke the validator for file fallback.

- [ ] **Step 2: Verify invalid and valid inputs**

Run the validator against committed values and `v0.1.170-1`, then run it against `v0.1.170-2` and require a non-zero exit.

- [ ] **Step 3: Inject both linker values**

Update Makefiles, Dockerfiles, and both GoReleaser configurations with `-X main.Version=<full>` and `-X main.UpstreamVersion=<baseline>`.

- [ ] **Step 4: Make releases PR-safe**

Add a validation job that checks the tag against committed metadata. Remove workflow jobs that rewrite and directly push `VERSION`. Force GoReleaser releases to `prerelease: false`.

- [ ] **Step 5: Add optional private-registry publication**

When `AIWELINK_REGISTRY_USERNAME` and `AIWELINK_REGISTRY_PASSWORD` secrets are configured, log in to `docker.aiwelink.cc` and publish immutable and `latest` tags under `sub2api-aiwelink-dev`. Keep GHCR publication as a fallback release channel.

### Task 6: Full verification and branch handoff

**Files:**
- Verify all modified files; do not stage user-owned `.gitignore`, `.superpowers/`, `AIWELINK_GIT_WORKFLOW.md`, or deployment drafts.

- [ ] **Step 1: Format changed Go files**

Run: `gofmt -w internal/versioninfo/version.go internal/versioninfo/version_test.go cmd/server/main.go cmd/server/wire.go cmd/server/wire_gen.go cmd/server/wire_gen_test.go internal/handler/handler.go internal/handler/wire.go internal/handler/setting_handler.go internal/handler/dto/settings.go internal/handler/setting_handler_public_test.go internal/handler/admin/system_handler.go internal/handler/admin/system_handler_test.go internal/service/wire.go internal/service/update_service.go internal/service/update_service_test.go` from `backend`.

- [ ] **Step 2: Run backend verification**

Run: `go test -tags=unit ./...`

Expected: PASS.

- [ ] **Step 3: Run frontend verification**

Run: `pnpm run lint:check`, `pnpm run typecheck`, `pnpm run test:run`, and `pnpm run build`.

Expected: PASS.

- [ ] **Step 4: Verify versions and production build**

Run the version validator with `v0.1.170-1`, run `go build` with both linker values, and build the Docker image tagged `docker.aiwelink.cc/sub2api-aiwelink-dev:0.1.170-1` without pushing it.

- [ ] **Step 5: Review scope and commit**

Stage only implementation files, confirm unrelated changes remain unstaged, and commit with `feat: add AIWeLink release versioning`.

- [ ] **Step 6: Push and create PR**

Push `chore/aiwelink-versioning` and create a PR targeting `aiwelink-dev` with the version rules, update-source safety change, test evidence, required registry secrets, and rollback behavior.
