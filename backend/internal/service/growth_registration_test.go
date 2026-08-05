package service

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
)

type growthRegistrationRecordCall struct {
	ctx                     context.Context
	ctxErr                  error
	hasDeadline             bool
	sourceRegistrationID    uuid.UUID
	siteID                  string
	externalUserID          string
	registeredAt            time.Time
	growthSessionCiphertext *string
}

type growthRegistrationRecordRepositoryStub struct {
	calls []growthRegistrationRecordCall
	err   error
}

func (r *growthRegistrationRecordRepositoryStub) RecordSuccessfulRegistration(
	ctx context.Context,
	sourceRegistrationID uuid.UUID,
	siteID string,
	externalUserID string,
	registeredAt time.Time,
	growthSessionCiphertext *string,
) error {
	_, hasDeadline := ctx.Deadline()
	r.calls = append(r.calls, growthRegistrationRecordCall{
		ctx:                     ctx,
		ctxErr:                  ctx.Err(),
		hasDeadline:             hasDeadline,
		sourceRegistrationID:    sourceRegistrationID,
		siteID:                  siteID,
		externalUserID:          externalUserID,
		registeredAt:            registeredAt,
		growthSessionCiphertext: growthSessionCiphertext,
	})
	return r.err
}

func TestNewGrowthRegistrationRecorderValidatesDependencies(t *testing.T) {
	repository := &growthRegistrationRecordRepositoryStub{}
	cipher, err := NewGrowthRegistrationCipher(growthRegistrationTestKey)
	require.NoError(t, err)

	recorder, err := NewGrowthRegistrationRecorder(nil, cipher, "aiwelink")
	require.Error(t, err)
	require.Nil(t, recorder)

	recorder, err = NewGrowthRegistrationRecorder(repository, nil, "aiwelink")
	require.Error(t, err)
	require.Nil(t, recorder)

	recorder, err = NewGrowthRegistrationRecorder(repository, cipher, "  ")
	require.Error(t, err)
	require.Nil(t, recorder)
}

func TestGrowthRegistrationRecorderBuildsEncryptedEventFromUserAndContext(t *testing.T) {
	repository := &growthRegistrationRecordRepositoryStub{}
	cipher, err := NewGrowthRegistrationCipher(growthRegistrationTestKey)
	require.NoError(t, err)
	recorder, err := NewGrowthRegistrationRecorder(repository, cipher, " aiwelink ")
	require.NoError(t, err)

	createdAt := time.Date(2026, time.August, 5, 16, 0, 0, 123456789, time.FixedZone("UTC+8", 8*60*60))
	ctx := WithGrowthRegistrationSession(context.Background(), "growth-session-123")
	err = recorder.RecordSuccessfulRegistration(ctx, &User{ID: 84521, CreatedAt: createdAt})
	require.NoError(t, err)
	require.Len(t, repository.calls, 1)

	call := repository.calls[0]
	require.NotEqual(t, uuid.Nil, call.sourceRegistrationID)
	require.Equal(t, "aiwelink", call.siteID)
	require.Equal(t, "84521", call.externalUserID)
	require.Equal(t, createdAt.UTC(), call.registeredAt)
	require.Equal(t, time.UTC, call.registeredAt.Location())
	require.NotNil(t, call.growthSessionCiphertext)
	require.NotEqual(t, "growth-session-123", *call.growthSessionCiphertext)

	plaintext, err := cipher.Decrypt(*call.growthSessionCiphertext)
	require.NoError(t, err)
	require.Equal(t, "growth-session-123", plaintext)
}

func TestGrowthRegistrationRecorderPersistsNilForMissingSession(t *testing.T) {
	repository := &growthRegistrationRecordRepositoryStub{}
	cipher, err := NewGrowthRegistrationCipher(growthRegistrationTestKey)
	require.NoError(t, err)
	recorder, err := NewGrowthRegistrationRecorder(repository, cipher, "aiwelink")
	require.NoError(t, err)

	err = recorder.RecordSuccessfulRegistration(context.Background(), &User{
		ID:        1,
		CreatedAt: time.Date(2026, time.August, 5, 8, 0, 0, 0, time.UTC),
	})
	require.NoError(t, err)
	require.Len(t, repository.calls, 1)
	require.Nil(t, repository.calls[0].growthSessionCiphertext)
}

func TestGrowthRegistrationRecorderCreatesFreshSourceID(t *testing.T) {
	repository := &growthRegistrationRecordRepositoryStub{}
	cipher, err := NewGrowthRegistrationCipher(growthRegistrationTestKey)
	require.NoError(t, err)
	recorder, err := NewGrowthRegistrationRecorder(repository, cipher, "aiwelink")
	require.NoError(t, err)
	user := &User{ID: 1, CreatedAt: time.Now()}

	require.NoError(t, recorder.RecordSuccessfulRegistration(context.Background(), user))
	require.NoError(t, recorder.RecordSuccessfulRegistration(context.Background(), user))
	require.Len(t, repository.calls, 2)
	require.NotEqual(t, repository.calls[0].sourceRegistrationID, repository.calls[1].sourceRegistrationID)
}

func TestGrowthRegistrationRecorderValidatesRecordInput(t *testing.T) {
	repository := &growthRegistrationRecordRepositoryStub{}
	cipher, err := NewGrowthRegistrationCipher(growthRegistrationTestKey)
	require.NoError(t, err)
	recorder, err := NewGrowthRegistrationRecorder(repository, cipher, "aiwelink")
	require.NoError(t, err)

	tests := map[string]struct {
		ctx  context.Context
		user *User
	}{
		"nil context":        {ctx: nil, user: &User{ID: 1, CreatedAt: time.Now()}},
		"nil user":           {ctx: context.Background(), user: nil},
		"non-positive id":    {ctx: context.Background(), user: &User{ID: 0, CreatedAt: time.Now()}},
		"missing created at": {ctx: context.Background(), user: &User{ID: 1}},
	}
	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			err := recorder.RecordSuccessfulRegistration(test.ctx, test.user)
			require.Error(t, err)
		})
	}
	require.Empty(t, repository.calls)
}

func TestGrowthRegistrationRecorderPropagatesRepositoryError(t *testing.T) {
	repositoryError := errors.New("insert failed")
	repository := &growthRegistrationRecordRepositoryStub{err: repositoryError}
	cipher, err := NewGrowthRegistrationCipher(growthRegistrationTestKey)
	require.NoError(t, err)
	recorder, err := NewGrowthRegistrationRecorder(repository, cipher, "aiwelink")
	require.NoError(t, err)

	err = recorder.RecordSuccessfulRegistration(context.Background(), &User{ID: 1, CreatedAt: time.Now()})
	require.ErrorIs(t, err, repositoryError)
}

func TestGrowthRegistrationRecorderDetachesRequestCancellationAndBoundsInsert(t *testing.T) {
	repository := &growthRegistrationRecordRepositoryStub{}
	cipher, err := NewGrowthRegistrationCipher(growthRegistrationTestKey)
	require.NoError(t, err)
	recorder, err := NewGrowthRegistrationRecorder(repository, cipher, "aiwelink")
	require.NoError(t, err)

	requestCtx, cancel := context.WithCancel(
		WithGrowthRegistrationSession(context.Background(), "growth-session"),
	)
	cancel()

	err = recorder.RecordSuccessfulRegistration(requestCtx, &User{ID: 1, CreatedAt: time.Now()})
	require.NoError(t, err)
	require.Len(t, repository.calls, 1)
	require.NoError(t, repository.calls[0].ctxErr)
	require.True(t, repository.calls[0].hasDeadline)
}

func TestGrowthRegistrationRecorderReturnsSourceIDGenerationError(t *testing.T) {
	repository := &growthRegistrationRecordRepositoryStub{}
	cipher, err := NewGrowthRegistrationCipher(growthRegistrationTestKey)
	require.NoError(t, err)
	recorder, err := NewGrowthRegistrationRecorder(repository, cipher, "aiwelink")
	require.NoError(t, err)

	implementation := recorder.(*growthRegistrationRecorder)
	implementation.newSourceRegistrationID = func() (uuid.UUID, error) {
		return uuid.Nil, errors.New("random source unavailable")
	}

	err = recorder.RecordSuccessfulRegistration(context.Background(), &User{ID: 1, CreatedAt: time.Now()})
	require.ErrorContains(t, err, "generate growth registration source id")
	require.Empty(t, repository.calls)
}
