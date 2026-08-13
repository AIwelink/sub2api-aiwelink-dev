# Growth Registration Binding Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Record successful ordinary email registrations with the Traffic promotion session from `/r/{code}`, durably and securely, without blocking user registration.

**Architecture:** Traffic owns `/r/{code}` and establishes the promotion cookie. Sub2API captures that cookie only on `POST /api/v1/auth/register`, writes an encrypted PostgreSQL outbox row after successful token generation, and delivers rows asynchronously to Traffic through a constrained HTTP worker. OAuth, passkey, login-growth, and native affiliate paths remain separate.

**Tech Stack:** Go 1.26, Gin, Viper, Wire, PostgreSQL, `database/sql`, AES-256-GCM, `net/http`, Docker Compose

---

### Task 1: Capture the promotion session

**Files:**
- Modify: `backend/internal/config/config.go`
- Create: `backend/internal/server/middleware/growth_registration_session.go`
- Create: `backend/internal/service/growth_registration_context.go`
- Test: `backend/internal/config/growth_registration_test.go`
- Test: `backend/internal/server/middleware/growth_registration_session_test.go`
- Test: `backend/internal/service/growth_registration_context_test.go`

- [x] Add the opt-in `GrowthRegistrationConfig` and bind the eight `GROWTH_REGISTRATION_*` environment variables.
- [x] Add bounded context storage for a 1-64 byte promotion session.
- [x] Install middleware that matches only method `POST` and path `/api/v1/auth/register`; invalid or missing cookies are omitted.
- [x] Run `go test ./internal/config ./internal/server/middleware ./internal/service -run GrowthRegistration -count=1`.
- [x] Commit `154bbd98e feat: capture growth registration session`.

### Task 2: Encrypt and record a registration

**Files:**
- Create: `backend/internal/service/growth_registration.go`
- Create: `backend/internal/service/growth_registration_crypto.go`
- Test: `backend/internal/service/growth_registration_test.go`
- Test: `backend/internal/service/growth_registration_crypto_test.go`

- [x] Define `GrowthRegistrationRecorder` and `GrowthRegistrationOutboxRepository` contracts with `uuid.UUID`, UTC timestamps, and nullable encrypted sessions.
- [x] Implement AES-256-GCM with a random nonce, `v1:` encoding, and versioned additional authenticated data.
- [x] Build events from the user ID and context session; use a detached, bounded insert context.
- [x] Ensure missing sessions remain nullable and recorder validation rejects nil dependencies, blank site IDs, and invalid user input.
- [x] Run the focused recorder and cipher tests.
- [x] Commit `10e47de01 feat: record encrypted growth registrations`.

### Task 3: Persist the durable outbox

**Files:**
- Create: `backend/migrations/194_growth_registration_outbox.sql`
- Create: `backend/internal/repository/growth_registration_outbox_repo.go`
- Modify: `backend/internal/repository/wire.go`
- Test: `backend/internal/repository/growth_registration_outbox_repo_test.go`

- [x] Add the unique source-registration ID, encrypted session, claim lease, retry metadata, and dead-letter columns.
- [x] Implement idempotent insert, `FOR UPDATE SKIP LOCKED` claim, owned delete, retry, and dead-letter transitions.
- [x] Clamp claim limits, normalize worker IDs, require exactly one owned row transition, and clear ciphertext on dead-letter.
- [x] Run the repository and migration contract tests.
- [x] Commit `040b07828 feat: persist growth registration outbox`.

### Task 4: Deliver events asynchronously

**Files:**
- Create: `backend/internal/service/growth_registration_worker.go`
- Create: `backend/internal/service/growth_registration_runtime.go`
- Modify: `backend/internal/service/wire.go`
- Modify: `backend/cmd/server/wire.go`
- Generate: `backend/cmd/server/wire_gen.go`
- Test: `backend/internal/service/growth_registration_worker_test.go`
- Test: `backend/internal/service/growth_registration_runtime_test.go`

- [x] Validate endpoints: private HTTP only, HTTPS for public hosts, no credentials or fragments, and valid DNS/IP/port syntax.
- [x] Use direct transport for private HTTP, environment proxy for HTTPS, redirect rejection, bounded timeouts, 16 KiB response headers, and 4 KiB response bodies.
- [x] Send a stable JSON payload with service authorization and a fresh request ID.
- [x] Retry transport failures and selected `503` error codes with jittered exponential backoff; dead-letter permanent or malformed responses.
- [x] Implement nil-safe `Start`, `Stop`, restart lifecycle, runtime construction, and server cleanup.
- [x] Run `go test ./cmd/server -count=1`, the worker/runtime tests, `go vet ./cmd/server ./internal/service`, and `git diff --check`.
- [x] Commit `4e4969048 feat: deliver growth registrations asynchronously`.

### Task 5: Hook ordinary email registration

**Files:**
- Modify: `backend/internal/service/auth_service.go`
- Modify: `backend/internal/service/auth_service_register_test.go`
- Modify: `backend/internal/service/wire.go`
- Generate: `backend/cmd/server/wire_gen.go`

- [x] Add the optional `GrowthRegistrationRecorder` field and `SetGrowthRegistrationRecorder` setter to `AuthService`.
- [x] In `RegisterWithVerification`, call the recorder only after `GenerateToken` succeeds; recorder errors are logged and do not change the registration result.
- [x] Keep the hook out of `LoginOrRegisterOAuthWithTokenPair`, passkey, login-growth, and native affiliate flows.
- [x] Add a unit test proving the context session reaches the recorder and recorder failure is fail-open.
- [x] Run `go test -tags unit ./internal/service -run '^TestAuthService_Register_RecordsGrowthRegistrationAndFailsOpen$' -count=1` and `go test ./cmd/server -count=1`.

### Task 6: Publish deployment configuration

**Files:**
- Modify: `deploy/.env.example`
- Modify: `deploy/config.example.yaml`
- Modify: `deploy/docker-compose.yml`
- Modify: `deploy/docker-compose.dev.yml`
- Modify: `deploy/docker-compose.local.yml`
- Modify: `deploy/docker-compose.standalone.yml`

- [x] Add all eight variables with safe defaults, an explicit disabled default, and the `openssl rand -hex 32` key-generation instruction.
- [x] Add the `growth_registration` YAML block using the same defaults as `GrowthRegistrationConfig`.
- [x] Pass all variables as separate environment entries in all four Compose application services; use the Docker service endpoint `http://traffic:8081/internal/growth/registrations/bind`.
- [x] Check the templates for duplicate or concatenated environment entries.

### Task 7: Restore documentation and verify the branch

**Files:**
- Create: `docs/superpowers/specs/2026-08-09-growth-registration-binding-design.md`
- Create: `docs/superpowers/plans/2026-08-09-growth-registration-binding.md`

- [x] Document the data flow, stable payload, outbox transitions, failure semantics, security boundaries, lifecycle, configuration, acceptance criteria, and explicit non-goals.
- [x] Run the full verification set:

```powershell
$env:GOPROXY='https://goproxy.cn,direct'
cd backend
go test -tags unit ./internal/service -count=1
go test ./internal/config ./internal/server/middleware ./internal/repository ./internal/service ./cmd/server -count=1
go vet ./...
git diff --check
```

- [x] Validate all four Compose files with `docker compose -f <file> config` when Docker is available.
- [x] Review the complete diff for scope and commit Task 5, deployment templates, and documentation on `codex/growth-registration-binding`.

## Scope Guard

Do not add `/r/{code}` routing to Sub2API, call `/internal/growth/logins`, alter Traffic code, or merge the existing affiliate/invitation implementation into this recorder. The only cross-service contract is the configured registration-binding endpoint.
