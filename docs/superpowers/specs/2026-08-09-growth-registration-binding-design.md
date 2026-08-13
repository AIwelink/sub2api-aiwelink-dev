# Growth Registration Binding Design

**Date:** 2026-08-09

## Goal

Bind a successful ordinary email registration in Sub2API to the promotion session established by the Traffic-owned `/r/{code}` flow, without coupling registration availability to Traffic.

## Scope

- Sub2API only consumes the configured promotion-session cookie.
- The only HTTP request inspected is `POST /api/v1/auth/register`.
- A registration event is recorded only after the user is created and the access token is generated successfully.
- Missing or invalid promotion cookies produce a registration event with `growth_session: null`.
- The feature is opt-in and disabled by default.

## Explicit Non-Goals

- Do not add or modify `/r/{code}` in Sub2API.
- Do not modify the Traffic service or its `/internal/growth/registrations/bind` implementation.
- Do not modify `/internal/growth/logins`.
- Do not record OAuth, SSO, passkey, or login activity.
- Do not reuse or change Sub2API native affiliate and invitation-code binding.

## Data Flow

1. The Traffic-owned `/r/{code}` flow establishes a cookie such as `awl_growth_sid`.
2. `GrowthRegistrationSession` runs only for the exact ordinary registration method and path.
3. A valid cookie value is copied into the request context; the middleware never changes the request body.
4. `AuthService.RegisterWithVerification` completes all existing registration work and generates the access token.
5. `GrowthRegistrationRecorder` creates a fresh source UUID and an outbox event.
6. The promotion session is encrypted before the event is inserted into PostgreSQL.
7. `GrowthRegistrationWorker` claims available rows and sends the stable payload to the configured Traffic endpoint.
8. A `200` or `201` response deletes the claimed row. Retryable failures are rescheduled; permanent failures are dead-lettered.

```mermaid
sequenceDiagram
    participant Browser
    participant Sub2 as Sub2API
    participant DB as PostgreSQL outbox
    participant Traffic
    Browser->>Sub2: POST /api/v1/auth/register + cookie
    Sub2->>Sub2: Create user and generate token
    Sub2->>DB: Insert encrypted registration event
    Sub2-->>Browser: Registration success
    Sub2->>DB: Claim event
    Sub2->>Traffic: POST /internal/growth/registrations/bind
    Traffic-->>Sub2: 200 or 201
    Sub2->>DB: Delete claimed event
```

## Stable Delivery Contract

The worker sends:

```json
{
  "site_id": "aiwelink",
  "external_user_id": "92",
  "source_registration_id": "8f4e59ce-2eb0-4d24-97d1-c248918da19e",
  "registered_at": "2026-08-09T12:00:00Z",
  "growth_session": "promotion-session"
}
```

- `external_user_id` is the decimal Sub2API user ID.
- `source_registration_id` is generated once when the outbox row is created and provides idempotency.
- `registered_at` is UTC and serialized as RFC3339Nano.
- `growth_session` is nullable.
- Requests use `Authorization: Service <credential>`, `Content-Type: application/json`, and a fresh `X-Request-ID`.

## Outbox Model

`backend/migrations/194_growth_registration_outbox.sql` defines a durable queue with:

- a unique `source_registration_id`;
- site, external user, registration timestamp, and encrypted session fields;
- availability, attempt, lease owner, and lease timestamp fields;
- bounded response metadata for operations and diagnosis;
- a dead-letter timestamp.

Workers claim with `FOR UPDATE SKIP LOCKED`. Claim transitions require the same worker ID, and expired claims are eligible for recovery. Dead-letter transitions clear the encrypted session value.

## Failure Semantics

- Registration recording is fail-open. Insert or encryption errors are logged and do not fail the user registration.
- The recorder detaches request cancellation and bounds the insert operation so a client disconnect does not lose an already successful registration.
- Disabled configuration produces a no-op runtime.
- Enabled configuration with a missing repository, credential, site ID, encryption key, endpoint, or positive timeout fails application startup.
- Transport failures and explicitly retryable `503` Traffic responses use bounded jittered exponential backoff.
- Other HTTP responses, malformed responses, oversized bodies, and decryption failures are dead-lettered.
- Outbox transition failures retain the row for lease recovery.

## Security Boundaries

- Promotion sessions are limited to 64 bytes.
- Sessions are encrypted with AES-256-GCM using a random nonce and fixed versioned additional authenticated data.
- The database never stores the session in plaintext.
- Logs do not include the session, credential, response body, or ciphertext.
- The encryption key is exactly 32 bytes encoded as 64 hexadecimal characters.
- HTTP endpoints are accepted only for explicit private hosts. Public endpoints require HTTPS.
- Private HTTP delivery bypasses environment proxies. HTTPS honors the environment proxy.
- Redirects are rejected so service credentials cannot be forwarded to another host.
- Only `*http.Transport` is accepted and cloned; arbitrary custom round trippers are rejected.
- Connect and response-header timeouts are bounded by configuration.
- Response headers are limited to 16 KiB and response bodies to 4 KiB.

## Configuration

| Environment variable | Default | Purpose |
| --- | --- | --- |
| `GROWTH_REGISTRATION_ENABLED` | `false` | Enables middleware, recorder, and worker behavior. |
| `GROWTH_REGISTRATION_ENDPOINT` | Docker: `http://traffic:8081/internal/growth/registrations/bind` | Traffic binding endpoint. |
| `GROWTH_REGISTRATION_SITE_ID` | `aiwelink` | Stable site identifier sent to Traffic. |
| `GROWTH_REGISTRATION_SERVICE_CREDENTIAL` | empty | Service authorization credential. |
| `GROWTH_REGISTRATION_OUTBOX_ENCRYPTION_KEY` | empty | AES-256-GCM key; generate with `openssl rand -hex 32`. |
| `GROWTH_REGISTRATION_COOKIE_NAME` | `awl_growth_sid` | Promotion-session cookie read at registration. |
| `GROWTH_REGISTRATION_CONNECT_TIMEOUT_SECONDS` | `2` | Connection timeout. |
| `GROWTH_REGISTRATION_READ_TIMEOUT_SECONDS` | `5` | Response-header and client timeout component. |

## Lifecycle

`ProvideGrowthRegistrationRuntime` validates enabled configuration, constructs the cipher, recorder, HTTP client, and worker, and starts the worker. Server cleanup calls `GrowthRegistrationRuntime.Stop`, which is nil-safe and supports `Start -> Stop -> Start` worker lifecycle tests.

## Acceptance Criteria

- Only the exact ordinary email registration endpoint can capture the promotion cookie.
- OAuth, passkey, login growth, and native affiliate flows remain unchanged.
- Successful ordinary registration records one event and recorder failure remains fail-open.
- No event is recorded before token generation succeeds.
- The outbox survives process and Traffic outages and safely retries eligible failures.
- Promotion sessions are encrypted at rest and absent from logs.
- Disabled configuration preserves existing behavior.
- Unit tests, package tests, `go vet ./...`, deployment-template validation, and `git diff --check` pass.
