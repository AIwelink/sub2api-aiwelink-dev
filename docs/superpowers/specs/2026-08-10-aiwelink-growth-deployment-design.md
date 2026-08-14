# AIWeLink Growth Registration Deployment Design

## Goal

Provide a repository-owned Docker Compose example, environment template, and Chinese operations guide for the AIWeLink development/gray Sub2API deployment. The deployment must enable the registration-attribution integration added in PR #15 without committing production secrets or changing the generic upstream deployment examples.

## Deployment Model

- The image is built from `aiwelink-dev` and published as both mutable `dev` and immutable `dev-<12-character-sha>` tags in `docker.aiwelink.cc/sub2api-aiwelink-dev`.
- The development/gray deployment publishes Sub2API on host port `8080`; production uses host port `8081`. Sub2API always listens on port `8080` inside the container.
- PostgreSQL and Redis are external services. Development/gray and production currently share the same PostgreSQL database and Redis deployment.
- The container joins the existing `${ONEPANEL_NETWORK_NAME:-1panel-network}` Docker network.
- Traffic is reached through the existing HTTPS reverse proxy at `https://aiwelink.cc/internal/growth/registrations/bind`. The deployment does not depend on host-loopback access or `host.docker.internal`.

## Repository Artifacts

### `deploy/docker-compose.aiwelink-dev.yml`

The dedicated Compose file runs only Sub2API and leaves PostgreSQL and Redis external. It:

- requires `SUB2API_IMAGE` and defaults the published host port to `8080`;
- fixes the container listener to `0.0.0.0:8080`;
- loads runtime values from `deploy/.env`;
- uses the named volume `sub2api-aiwelink-dev` for `/app/data`;
- joins the external 1Panel network;
- uses `pull_policy: always`, `no-new-privileges`, the existing file-descriptor limit, and a health check against the container's port `8080`.

Environment-specific and secret values remain in `.env`. Compose only supplies fixed container invariants and a few documented defaults.

### `deploy/.env.aiwelink-dev.example`

The template includes placeholders for the image, external PostgreSQL, external Redis, JWT/TOTP secrets, and all eight `GROWTH_REGISTRATION_*` settings. It does not contain real credentials.

The growth settings are:

| Variable | Required value or rule |
| --- | --- |
| `GROWTH_REGISTRATION_ENABLED` | `true` |
| `GROWTH_REGISTRATION_ENDPOINT` | `https://aiwelink.cc/internal/growth/registrations/bind` |
| `GROWTH_REGISTRATION_SITE_ID` | `aiwelink` |
| `GROWTH_REGISTRATION_SERVICE_CREDENTIAL` | Exactly the Traffic `SITE_SERVICE_CREDENTIALS_JSON.aiwelink` credential |
| `GROWTH_REGISTRATION_OUTBOX_ENCRYPTION_KEY` | A dedicated 64-character hexadecimal AES-256 key |
| `GROWTH_REGISTRATION_COOKIE_NAME` | `awl_growth_sid` |
| `GROWTH_REGISTRATION_CONNECT_TIMEOUT_SECONDS` | `2` |
| `GROWTH_REGISTRATION_READ_TIMEOUT_SECONDS` | `5` |

`GROWTH_REGISTRATION_OUTBOX_ENCRYPTION_KEY` must not reuse the Traffic service credential. Because all enabled Sub2API replicas claim rows from the same database outbox, every enabled deployment sharing that database must use the same outbox key. Otherwise one replica can claim another replica's row and dead-letter it as `decrypt_failed`.

The old `GROWTH_LOGIN_*` names are explicitly documented as incompatible with registration attribution.

### `deploy/AIWELINK_GROWTH_REGISTRATION_CN.md`

The operations guide covers:

- preparing `.env` from the committed template;
- generating JWT, TOTP, and outbox keys;
- copying the `aiwelink` service credential from Traffic without exposing it;
- validating Compose interpolation;
- pulling and deploying an immutable `dev-<sha>` image;
- checking the eight variables inside the running container without printing secret values;
- probing Traffic authentication and Sub2API health;
- performing an end-to-end `/r/{code}` to email-registration test;
- querying pending/dead-letter outbox state and interpreting common failures;
- using host port `8080` for development/gray and `8081` for production.

## Registration Data Flow

1. A browser visits `https://aiwelink.cc/r/{code}`.
2. Traffic establishes the `awl_growth_sid` cookie for `.aiwelink.cc`.
3. The browser completes email registration through Sub2API and sends that cookie.
4. Sub2API generates the access token, then commits the user and encrypted `growth_registration_outbox` row atomically in the shared PostgreSQL database.
5. A Sub2API worker posts the event over HTTPS to Traffic with `Authorization: Service <credential>`.
6. Traffic returns `200` or `201`; Sub2API deletes the delivered outbox row. HTTP `408`, `425`, `429`, `500`, `502`, `503`, and `504` responses are retried, and terminal failures are retained as dead letters.

## Security And Failure Handling

- Only placeholders are committed. The real `.env` remains ignored and should be mode `0600` on Linux.
- Traffic communication uses HTTPS and rejects redirects.
- The endpoint contains no credential in its URL.
- Verification commands show whether secrets are set and report their lengths, but never print their values.
- An unauthenticated Traffic probe should return `401`, which proves the route is reachable and authentication is active.
- Shared-database deployments coordinate outbox claims with database leases; the shared encryption key is therefore an operational requirement.

## Validation

Before the pull request is opened:

1. Render the dedicated Compose file with a temporary non-secret test env using `docker compose config`.
2. Assert the rendered host/container port mapping, Traffic endpoint, external network, image, health-check port, and all eight registration settings.
3. Run the growth registration Go tests.
4. Run the deployment shell tests affected by the new example.
5. Check Markdown links and commands manually and verify that no real secret or server `.env` value is tracked.

## Non-Goals

- Do not modify or merge `main`.
- Do not change generic `deploy/docker-compose.yml`, `deploy/docker-compose.standalone.yml`, or the default disabled behavior.
- Do not deploy to the server from this pull request.
- Do not commit a real `.env`, database password, Redis password, service credential, or encryption key.
- Do not change `/r/{code}`, Traffic routing, or Sub2API registration code.

## Acceptance Criteria

- A server operator can create `deploy/.env` from the example and deploy the gray instance without guessing variable names or ports.
- The generated container receives all eight `GROWTH_REGISTRATION_*` values.
- The Traffic endpoint is the reachable HTTPS reverse-proxy URL.
- Documentation prevents reuse of `GROWTH_LOGIN_*` and explains the shared-database encryption-key requirement.
- The branch is proposed only to `aiwelink-dev` through a pull request.
