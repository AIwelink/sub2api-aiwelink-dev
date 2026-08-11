CREATE TABLE IF NOT EXISTS growth_registration_outbox (
    outbox_id BIGSERIAL PRIMARY KEY,
    source_registration_id UUID NOT NULL UNIQUE,
    site_id VARCHAR(100) NOT NULL CHECK (site_id <> ''),
    external_user_id VARCHAR(255) NOT NULL CHECK (external_user_id <> ''),
    registered_at TIMESTAMPTZ NOT NULL,
    growth_session_ciphertext TEXT NULL CHECK (
        growth_session_ciphertext IS NULL
        OR octet_length(growth_session_ciphertext) <= 512
    ),
    attempt_count INTEGER NOT NULL DEFAULT 0 CHECK (attempt_count >= 0),
    available_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    claimed_at TIMESTAMPTZ NULL,
    claimed_by VARCHAR(100) NULL,
    last_http_status SMALLINT NULL CHECK (
        last_http_status IS NULL OR last_http_status BETWEEN 100 AND 599
    ),
    last_error_code VARCHAR(100) NULL,
    last_request_id VARCHAR(64) NULL,
    dead_lettered_at TIMESTAMPTZ NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_growth_registration_outbox_available
    ON growth_registration_outbox (available_at, outbox_id)
    WHERE dead_lettered_at IS NULL AND claimed_at IS NULL;

CREATE INDEX IF NOT EXISTS idx_growth_registration_outbox_expired_lease
    ON growth_registration_outbox (claimed_at, outbox_id)
    WHERE dead_lettered_at IS NULL AND claimed_at IS NOT NULL;

CREATE INDEX IF NOT EXISTS idx_growth_registration_outbox_created_at
    ON growth_registration_outbox (created_at);

COMMENT ON TABLE growth_registration_outbox IS
    'Durable Sub2 successful-registration events awaiting Traffic Analysis delivery';
