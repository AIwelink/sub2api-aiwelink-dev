package repository

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/Wei-Shaw/sub2api/internal/service"
	"github.com/google/uuid"
)

type growthRegistrationOutboxRepository struct {
	db *sql.DB
}

const (
	defaultGrowthRegistrationClaimLimit = 25
	maxGrowthRegistrationClaimLimit     = 100
	defaultGrowthRegistrationClaimLease = 30 * time.Second
)

func NewGrowthRegistrationOutboxRepository(db *sql.DB) service.GrowthRegistrationOutboxRepository {
	return &growthRegistrationOutboxRepository{db: db}
}

func (r *growthRegistrationOutboxRepository) RecordSuccessfulRegistration(
	ctx context.Context,
	sourceRegistrationID uuid.UUID,
	siteID string,
	externalUserID string,
	registeredAt time.Time,
	growthSessionCiphertext *string,
) error {
	if err := r.validateDatabase(); err != nil {
		return err
	}
	_, err := r.db.ExecContext(ctx, `
		INSERT INTO growth_registration_outbox (
			source_registration_id,
			site_id,
			external_user_id,
			registered_at,
			growth_session_ciphertext
		) VALUES ($1, $2, $3, $4, $5)
		ON CONFLICT (source_registration_id) DO NOTHING
	`, sourceRegistrationID, siteID, externalUserID, registeredAt.UTC(), growthSessionCiphertext)
	if err != nil {
		return fmt.Errorf("insert growth registration outbox: %w", err)
	}
	return nil
}

func (r *growthRegistrationOutboxRepository) Claim(
	ctx context.Context,
	workerID string,
	limit int,
	lease time.Duration,
) ([]service.GrowthRegistrationOutboxEvent, error) {
	if err := r.validateDatabase(); err != nil {
		return nil, err
	}
	workerID, err := normalizeGrowthRegistrationWorkerID(workerID)
	if err != nil {
		return nil, err
	}
	if limit <= 0 {
		limit = defaultGrowthRegistrationClaimLimit
	} else if limit > maxGrowthRegistrationClaimLimit {
		limit = maxGrowthRegistrationClaimLimit
	}
	leaseMicroseconds := defaultGrowthRegistrationClaimLease.Microseconds()
	if lease > 0 {
		leaseMicroseconds = lease.Microseconds()
		if leaseMicroseconds < 1 {
			leaseMicroseconds = 1
		}
	}

	rows, err := r.db.QueryContext(ctx, `
		WITH candidates AS (
			SELECT outbox_id
			FROM growth_registration_outbox
			WHERE dead_lettered_at IS NULL
			  AND available_at <= NOW()
			  AND (
				claimed_at IS NULL
				OR claimed_at < NOW() - ($3 * INTERVAL '1 microsecond')
			  )
			ORDER BY available_at, outbox_id
			LIMIT $2
			FOR UPDATE SKIP LOCKED
		)
		UPDATE growth_registration_outbox AS outbox
		SET claimed_at = NOW(), claimed_by = $1
		FROM candidates
		WHERE outbox.outbox_id = candidates.outbox_id
		RETURNING
			outbox.outbox_id,
			outbox.source_registration_id,
			outbox.site_id,
			outbox.external_user_id,
			outbox.registered_at,
			outbox.growth_session_ciphertext,
			outbox.attempt_count
	`, workerID, limit, leaseMicroseconds)
	if err != nil {
		return nil, fmt.Errorf("claim growth registration outbox: %w", err)
	}
	defer func() { _ = rows.Close() }()

	events := make([]service.GrowthRegistrationOutboxEvent, 0, limit)
	for rows.Next() {
		var (
			event      service.GrowthRegistrationOutboxEvent
			ciphertext sql.NullString
		)
		if err := rows.Scan(
			&event.OutboxID,
			&event.SourceRegistrationID,
			&event.SiteID,
			&event.ExternalUserID,
			&event.RegisteredAt,
			&ciphertext,
			&event.AttemptCount,
		); err != nil {
			return nil, fmt.Errorf("scan growth registration outbox: %w", err)
		}
		if ciphertext.Valid {
			value := ciphertext.String
			event.GrowthSessionCiphertext = &value
		}
		events = append(events, event)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate growth registration outbox: %w", err)
	}
	return events, nil
}

func (r *growthRegistrationOutboxRepository) DeleteClaimed(
	ctx context.Context,
	outboxID int64,
	workerID string,
) error {
	if err := r.validateDatabase(); err != nil {
		return err
	}
	workerID, err := normalizeGrowthRegistrationWorkerID(workerID)
	if err != nil {
		return err
	}
	result, err := r.db.ExecContext(ctx, `
		DELETE FROM growth_registration_outbox
		WHERE outbox_id = $1
		  AND claimed_by = $2
		  AND dead_lettered_at IS NULL
	`, outboxID, workerID)
	if err != nil {
		return fmt.Errorf("delete delivered growth registration: %w", err)
	}
	return requireGrowthRegistrationClaim(result, outboxID)
}

func (r *growthRegistrationOutboxRepository) RetryClaimed(
	ctx context.Context,
	outboxID int64,
	workerID string,
	availableAt time.Time,
	httpStatus *int,
	errorCode string,
	requestID string,
) error {
	if err := r.validateDatabase(); err != nil {
		return err
	}
	workerID, err := normalizeGrowthRegistrationWorkerID(workerID)
	if err != nil {
		return err
	}
	result, err := r.db.ExecContext(ctx, `
		UPDATE growth_registration_outbox
		SET attempt_count = attempt_count + 1,
			available_at = $3,
			last_http_status = $4,
			last_error_code = $5,
			last_request_id = $6,
			claimed_at = NULL,
			claimed_by = NULL
		WHERE outbox_id = $1
		  AND claimed_by = $2
		  AND dead_lettered_at IS NULL
	`,
		outboxID,
		workerID,
		availableAt.UTC(),
		nullableGrowthRegistrationStatus(httpStatus),
		nullableGrowthRegistrationString(errorCode),
		nullableGrowthRegistrationString(requestID),
	)
	if err != nil {
		return fmt.Errorf("retry growth registration outbox: %w", err)
	}
	return requireGrowthRegistrationClaim(result, outboxID)
}

func (r *growthRegistrationOutboxRepository) DeadLetterClaimed(
	ctx context.Context,
	outboxID int64,
	workerID string,
	httpStatus *int,
	errorCode string,
	requestID string,
) error {
	if err := r.validateDatabase(); err != nil {
		return err
	}
	workerID, err := normalizeGrowthRegistrationWorkerID(workerID)
	if err != nil {
		return err
	}
	result, err := r.db.ExecContext(ctx, `
		UPDATE growth_registration_outbox
		SET attempt_count = attempt_count + 1,
			last_http_status = $3,
			last_error_code = $4,
			last_request_id = $5,
			growth_session_ciphertext = NULL,
			dead_lettered_at = NOW(),
			claimed_at = NULL,
			claimed_by = NULL
		WHERE outbox_id = $1
		  AND claimed_by = $2
		  AND dead_lettered_at IS NULL
	`,
		outboxID,
		workerID,
		nullableGrowthRegistrationStatus(httpStatus),
		nullableGrowthRegistrationString(errorCode),
		nullableGrowthRegistrationString(requestID),
	)
	if err != nil {
		return fmt.Errorf("dead-letter growth registration outbox: %w", err)
	}
	return requireGrowthRegistrationClaim(result, outboxID)
}

func (r *growthRegistrationOutboxRepository) validateDatabase() error {
	if r == nil || r.db == nil {
		return errors.New("nil growth registration outbox database")
	}
	return nil
}

func normalizeGrowthRegistrationWorkerID(workerID string) (string, error) {
	workerID = strings.TrimSpace(workerID)
	if workerID == "" {
		return "", errors.New("growth registration worker id is required")
	}
	return workerID, nil
}

func requireGrowthRegistrationClaim(result sql.Result, outboxID int64) error {
	affected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("inspect growth registration outbox claim %d rows affected: %w", outboxID, err)
	}
	if affected != 1 {
		return fmt.Errorf("growth registration outbox claim %d is no longer owned", outboxID)
	}
	return nil
}

func nullableGrowthRegistrationStatus(status *int) any {
	if status == nil {
		return nil
	}
	return *status
}

func nullableGrowthRegistrationString(value string) any {
	value = strings.TrimSpace(value)
	if value == "" {
		return nil
	}
	return value
}
