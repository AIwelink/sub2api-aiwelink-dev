package repository

import (
	"context"
	"errors"
	"net/http"
	"strings"
	"testing"
	"time"

	sqlmock "github.com/DATA-DOG/go-sqlmock"
	"github.com/Wei-Shaw/sub2api/migrations"
	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
)

func TestGrowthRegistrationOutboxMigrationDefinesDurableEncryptedQueue(t *testing.T) {
	content, err := migrations.FS.ReadFile("194_growth_registration_outbox.sql")
	require.NoError(t, err)
	sqlText := string(content)

	for _, required := range []string{
		"source_registration_id UUID NOT NULL UNIQUE",
		"site_id VARCHAR(100) NOT NULL",
		"external_user_id VARCHAR(255) NOT NULL",
		"registered_at TIMESTAMPTZ NOT NULL",
		"growth_session_ciphertext TEXT NULL",
		"octet_length(growth_session_ciphertext) <= 512",
		"attempt_count INTEGER NOT NULL DEFAULT 0",
		"available_at TIMESTAMPTZ NOT NULL DEFAULT NOW()",
		"claimed_at TIMESTAMPTZ NULL",
		"claimed_by VARCHAR(100) NULL",
		"last_http_status SMALLINT NULL",
		"last_http_status BETWEEN 100 AND 599",
		"last_error_code VARCHAR(100) NULL",
		"last_request_id VARCHAR(64) NULL",
		"dead_lettered_at TIMESTAMPTZ NULL",
		"created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()",
		"WHERE dead_lettered_at IS NULL AND claimed_at IS NULL",
		"WHERE dead_lettered_at IS NULL AND claimed_at IS NOT NULL",
	} {
		require.Contains(t, sqlText, required)
	}

	lowerSQL := strings.ToLower(sqlText)
	require.NotContains(t, lowerSQL, "references users")
	require.NotContains(t, lowerSQL, "password")
	require.NotContains(t, lowerSQL, "authorization")
	require.NotContains(t, lowerSQL, "response_body")
}

func TestGrowthRegistrationOutboxRepositoryRecordsRegistrationIdempotently(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer func() { _ = db.Close() }()

	sourceRegistrationID := uuid.MustParse("019c02a8-7ac0-7ab6-bbf9-8c309fbe8ca1")
	registeredAt := time.Date(2026, 8, 5, 8, 0, 0, 123456000, time.UTC)
	ciphertext := "encrypted-growth-session"

	mock.ExpectExec("(?s)INSERT INTO growth_registration_outbox.*source_registration_id.*registered_at.*growth_session_ciphertext.*ON CONFLICT \\(source_registration_id\\) DO NOTHING").
		WithArgs(sourceRegistrationID, "aiwelink", "84521", registeredAt, &ciphertext).
		WillReturnResult(sqlmock.NewResult(1, 1))

	repository := NewGrowthRegistrationOutboxRepository(db)
	err = repository.RecordSuccessfulRegistration(
		context.Background(),
		sourceRegistrationID,
		"aiwelink",
		"84521",
		registeredAt,
		&ciphertext,
	)
	require.NoError(t, err)
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestGrowthRegistrationOutboxRepositoryTreatsDuplicateSourceAsSuccess(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer func() { _ = db.Close() }()

	sourceRegistrationID := uuid.MustParse("019c02a8-7ac0-7ab6-bbf9-8c309fbe8ca1")
	registeredAt := time.Date(2026, 8, 5, 8, 0, 0, 0, time.UTC)

	mock.ExpectExec("(?s)INSERT INTO growth_registration_outbox.*ON CONFLICT \\(source_registration_id\\) DO NOTHING").
		WithArgs(sourceRegistrationID, "aiwelink", "84521", registeredAt, nil).
		WillReturnResult(sqlmock.NewResult(0, 0))

	repository := NewGrowthRegistrationOutboxRepository(db)
	err = repository.RecordSuccessfulRegistration(
		context.Background(), sourceRegistrationID, "aiwelink", "84521", registeredAt, nil,
	)
	require.NoError(t, err)
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestGrowthRegistrationOutboxRepositoryClaimsAvailableAndExpiredLeases(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer func() { _ = db.Close() }()

	registeredAt := time.Date(2026, 8, 5, 8, 0, 0, 0, time.UTC)
	sourceRegistrationID := uuid.MustParse("019c02a8-7ac0-7ab6-bbf9-8c309fbe8ca1")
	mock.ExpectQuery("(?s)WITH candidates AS.*available_at <= NOW\\(\\).*claimed_at IS NULL.*claimed_at < NOW\\(\\) - \\(\\$3 \\* INTERVAL '1 microsecond'\\).*FOR UPDATE SKIP LOCKED.*UPDATE growth_registration_outbox AS outbox.*FROM candidates.*outbox\\.outbox_id = candidates\\.outbox_id.*RETURNING").
		WithArgs("worker-a", 25, int64(1_500_000)).
		WillReturnRows(sqlmock.NewRows([]string{
			"outbox_id", "source_registration_id", "site_id", "external_user_id",
			"registered_at", "growth_session_ciphertext", "attempt_count",
		}).AddRow(
			int64(7), sourceRegistrationID, "aiwelink", "84521", registeredAt, "ciphertext", 2,
		).AddRow(
			int64(8), uuid.MustParse("019c02a8-7ac0-7ab6-bbf9-8c309fbe8ca2"),
			"aiwelink", "84522", registeredAt.Add(time.Second), nil, 0,
		))

	repository := NewGrowthRegistrationOutboxRepository(db)
	events, err := repository.Claim(context.Background(), "worker-a", 25, 1500*time.Millisecond)
	require.NoError(t, err)
	require.Len(t, events, 2)
	require.Equal(t, int64(7), events[0].OutboxID)
	require.Equal(t, sourceRegistrationID, events[0].SourceRegistrationID)
	require.NotNil(t, events[0].GrowthSessionCiphertext)
	require.Equal(t, "ciphertext", *events[0].GrowthSessionCiphertext)
	require.Nil(t, events[1].GrowthSessionCiphertext)
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestGrowthRegistrationOutboxRepositoryClaimCapsExcessiveLimitAndDefaultsNonPositiveLease(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer func() { _ = db.Close() }()

	mock.ExpectQuery("(?s)WITH candidates AS.*LIMIT \\$2.*FOR UPDATE SKIP LOCKED").
		WithArgs("worker-a", 100, int64(30_000_000)).
		WillReturnRows(sqlmock.NewRows([]string{
			"outbox_id", "source_registration_id", "site_id", "external_user_id",
			"registered_at", "growth_session_ciphertext", "attempt_count",
		}))

	repository := NewGrowthRegistrationOutboxRepository(db)
	events, err := repository.Claim(context.Background(), "worker-a", 1_000_000, 0)
	require.NoError(t, err)
	require.Empty(t, events)
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestGrowthRegistrationOutboxRepositoryClaimClampsPositiveSubMicrosecondLease(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer func() { _ = db.Close() }()

	mock.ExpectQuery("(?s)WITH candidates AS.*INTERVAL '1 microsecond'.*FOR UPDATE SKIP LOCKED").
		WithArgs("worker-a", 25, int64(1)).
		WillReturnRows(sqlmock.NewRows([]string{
			"outbox_id", "source_registration_id", "site_id", "external_user_id",
			"registered_at", "growth_session_ciphertext", "attempt_count",
		}))

	repository := NewGrowthRegistrationOutboxRepository(db)
	events, err := repository.Claim(context.Background(), "worker-a", 25, time.Nanosecond)
	require.NoError(t, err)
	require.Empty(t, events)
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestGrowthRegistrationOutboxRepositoryDeletesOnlyOwnedClaim(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer func() { _ = db.Close() }()

	mock.ExpectExec("(?s)DELETE FROM growth_registration_outbox.*outbox_id = \\$1.*claimed_by = \\$2.*dead_lettered_at IS NULL").
		WithArgs(int64(7), "worker-a").
		WillReturnResult(sqlmock.NewResult(0, 1))

	repository := NewGrowthRegistrationOutboxRepository(db)
	require.NoError(t, repository.DeleteClaimed(context.Background(), int64(7), "worker-a"))
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestGrowthRegistrationOutboxRepositoryRetryClearsClaimAndStoresNullableMetadata(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer func() { _ = db.Close() }()

	nextAttempt := time.Date(2026, 8, 5, 9, 0, 0, 0, time.FixedZone("UTC+8", 8*60*60))
	mock.ExpectExec("(?s)UPDATE growth_registration_outbox.*attempt_count = attempt_count \\+ 1.*available_at = \\$3.*last_http_status = \\$4.*last_error_code = \\$5.*last_request_id = \\$6.*claimed_at = NULL.*claimed_by = NULL.*claimed_by = \\$2.*dead_lettered_at IS NULL").
		WithArgs(int64(7), "worker-a", nextAttempt.UTC(), nil, nil, nil).
		WillReturnResult(sqlmock.NewResult(0, 1))

	repository := NewGrowthRegistrationOutboxRepository(db)
	require.NoError(t, repository.RetryClaimed(
		context.Background(), int64(7), "worker-a", nextAttempt, nil, "  ", "",
	))
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestGrowthRegistrationOutboxRepositoryDeadLetterClearsCiphertext(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer func() { _ = db.Close() }()

	status := http.StatusUnprocessableEntity
	mock.ExpectExec("(?s)UPDATE growth_registration_outbox.*attempt_count = attempt_count \\+ 1.*growth_session_ciphertext = NULL.*dead_lettered_at = NOW\\(\\).*claimed_at = NULL.*claimed_by = NULL.*claimed_by = \\$2.*dead_lettered_at IS NULL").
		WithArgs(int64(8), "worker-a", status, "invalid_request", "remote-2").
		WillReturnResult(sqlmock.NewResult(0, 1))

	repository := NewGrowthRegistrationOutboxRepository(db)
	require.NoError(t, repository.DeadLetterClaimed(
		context.Background(), int64(8), "worker-a", &status, "invalid_request", "remote-2",
	))
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestGrowthRegistrationOutboxRepositoryNormalizesWorkerIDAcrossClaimTransitions(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer func() { _ = db.Close() }()

	mock.ExpectQuery("(?s)WITH candidates AS.*SET claimed_at = NOW\\(\\), claimed_by = \\$1").
		WithArgs("worker-a", 25, int64(30_000_000)).
		WillReturnRows(sqlmock.NewRows([]string{
			"outbox_id", "source_registration_id", "site_id", "external_user_id",
			"registered_at", "growth_session_ciphertext", "attempt_count",
		}))
	mock.ExpectExec("(?s)DELETE FROM growth_registration_outbox.*claimed_by = \\$2").
		WithArgs(int64(7), "worker-a").
		WillReturnResult(sqlmock.NewResult(0, 1))

	nextAttempt := time.Date(2026, 8, 5, 9, 30, 0, 0, time.UTC)
	mock.ExpectExec("(?s)UPDATE growth_registration_outbox.*available_at = \\$3.*claimed_by = \\$2").
		WithArgs(int64(8), "worker-a", nextAttempt, nil, nil, nil).
		WillReturnResult(sqlmock.NewResult(0, 1))

	status := http.StatusBadGateway
	mock.ExpectExec("(?s)UPDATE growth_registration_outbox.*dead_lettered_at = NOW\\(\\).*claimed_by = \\$2").
		WithArgs(int64(9), "worker-a", status, "upstream_failure", "remote-9").
		WillReturnResult(sqlmock.NewResult(0, 1))

	repository := NewGrowthRegistrationOutboxRepository(db)
	_, err = repository.Claim(context.Background(), " worker-a ", 0, 0)
	require.NoError(t, err)
	require.NoError(t, repository.DeleteClaimed(context.Background(), int64(7), " worker-a "))
	require.NoError(t, repository.RetryClaimed(
		context.Background(), int64(8), " worker-a ", nextAttempt, nil, "", "",
	))
	require.NoError(t, repository.DeadLetterClaimed(
		context.Background(), int64(9), " worker-a ", &status, "upstream_failure", "remote-9",
	))
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestGrowthRegistrationOutboxRepositoryRejectsBlankWorkerIDAcrossClaimTransitions(t *testing.T) {
	db, _, err := sqlmock.New()
	require.NoError(t, err)
	defer func() { _ = db.Close() }()

	repository := NewGrowthRegistrationOutboxRepository(db)
	nextAttempt := time.Date(2026, 8, 5, 9, 30, 0, 0, time.UTC)
	tests := []struct {
		name string
		call func() error
	}{
		{
			name: "claim",
			call: func() error {
				_, err := repository.Claim(context.Background(), "  ", 25, time.Second)
				return err
			},
		},
		{
			name: "delete",
			call: func() error {
				return repository.DeleteClaimed(context.Background(), int64(7), "  ")
			},
		},
		{
			name: "retry",
			call: func() error {
				return repository.RetryClaimed(
					context.Background(), int64(8), "  ", nextAttempt, nil, "", "",
				)
			},
		},
		{
			name: "dead letter",
			call: func() error {
				return repository.DeadLetterClaimed(
					context.Background(), int64(9), "  ", nil, "", "",
				)
			},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			require.ErrorContains(t, test.call(), "growth registration worker id is required")
		})
	}
}

func TestGrowthRegistrationOutboxRepositoryTransitionsRejectLostClaim(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer func() { _ = db.Close() }()

	mock.ExpectExec("DELETE FROM growth_registration_outbox").
		WithArgs(int64(9), "worker-a").
		WillReturnResult(sqlmock.NewResult(0, 0))

	repository := NewGrowthRegistrationOutboxRepository(db)
	err = repository.DeleteClaimed(context.Background(), int64(9), "worker-a")
	require.ErrorContains(t, err, "no longer owned")
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestGrowthRegistrationOutboxRepositoryWrapsRowsAffectedErrors(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer func() { _ = db.Close() }()

	driverErr := errors.New("rows affected unavailable")
	mock.ExpectExec("DELETE FROM growth_registration_outbox").
		WithArgs(int64(10), "worker-a").
		WillReturnResult(sqlmock.NewErrorResult(driverErr))

	repository := NewGrowthRegistrationOutboxRepository(db)
	err = repository.DeleteClaimed(context.Background(), int64(10), "worker-a")
	require.ErrorContains(t, err, "inspect growth registration outbox claim 10 rows affected")
	require.ErrorIs(t, err, driverErr)
	require.NoError(t, mock.ExpectationsWereMet())
}
