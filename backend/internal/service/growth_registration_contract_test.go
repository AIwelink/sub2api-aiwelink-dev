//go:build unit

package service

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"sync"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
)

type growthContractRequest struct {
	Authorization string
	RequestID     string
	Payload       growthRegistrationPayload
}

type growthContractRetry struct {
	HTTPStatus *int
	ErrorCode  string
	RequestID  string
}

type growthContractRepository struct {
	mu         sync.Mutex
	nextID     int64
	pending    []GrowthRegistrationOutboxEvent
	delivered  chan int64
	retried    chan growthContractRetry
	deadLetter chan string
}

func newGrowthContractRepository() *growthContractRepository {
	return &growthContractRepository{
		delivered:  make(chan int64, 1),
		retried:    make(chan growthContractRetry, 1),
		deadLetter: make(chan string, 1),
	}
}

func (r *growthContractRepository) RecordSuccessfulRegistration(
	_ context.Context,
	sourceRegistrationID uuid.UUID,
	siteID string,
	externalUserID string,
	registeredAt time.Time,
	growthSessionCiphertext *string,
) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.nextID++
	r.pending = append(r.pending, GrowthRegistrationOutboxEvent{
		OutboxID:                r.nextID,
		SourceRegistrationID:    sourceRegistrationID,
		SiteID:                  siteID,
		ExternalUserID:          externalUserID,
		RegisteredAt:            registeredAt,
		GrowthSessionCiphertext: growthSessionCiphertext,
	})
	return nil
}

func (r *growthContractRepository) Claim(
	_ context.Context,
	_ string,
	limit int,
	_ time.Duration,
) ([]GrowthRegistrationOutboxEvent, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	if limit > len(r.pending) {
		limit = len(r.pending)
	}
	claimed := append([]GrowthRegistrationOutboxEvent(nil), r.pending[:limit]...)
	r.pending = append([]GrowthRegistrationOutboxEvent(nil), r.pending[limit:]...)
	return claimed, nil
}

func (r *growthContractRepository) DeleteClaimed(
	_ context.Context,
	outboxID int64,
	_ string,
) error {
	r.delivered <- outboxID
	return nil
}

func (r *growthContractRepository) RetryClaimed(
	_ context.Context,
	_ int64,
	_ string,
	_ time.Time,
	httpStatus *int,
	errorCode string,
	requestID string,
) error {
	r.retried <- growthContractRetry{
		HTTPStatus: cloneInt(httpStatus),
		ErrorCode:  errorCode,
		RequestID:  requestID,
	}
	return nil
}

func (r *growthContractRepository) DeadLetterClaimed(
	_ context.Context,
	_ int64,
	_ string,
	_ *int,
	errorCode string,
	_ string,
) error {
	r.deadLetter <- errorCode
	return nil
}

func requireGrowthContractEvent[T any](t *testing.T, channel <-chan T) T {
	t.Helper()
	select {
	case value := <-channel:
		return value
	case <-time.After(2 * time.Second):
		t.Fatal("timed out waiting for growth contract event")
		var zero T
		return zero
	}
}

func TestGrowthRegistrationContractDeliversEncryptedSessionToTraffic(t *testing.T) {
	requests := make(chan growthContractRequest, 1)
	traffic := httptest.NewServer(http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		var payload growthRegistrationPayload
		require.NoError(t, json.NewDecoder(request.Body).Decode(&payload))
		requests <- growthContractRequest{
			Authorization: request.Header.Get("Authorization"),
			RequestID:     request.Header.Get("X-Request-ID"),
			Payload:       payload,
		}
		writer.Header().Set("X-Request-ID", "traffic-request")
		writer.WriteHeader(http.StatusCreated)
	}))
	defer traffic.Close()

	repository := newGrowthContractRepository()
	cipher, err := NewGrowthRegistrationCipher(growthRegistrationTestKey)
	require.NoError(t, err)
	recorder, err := NewGrowthRegistrationRecorder(repository, cipher, "aiwelink")
	require.NoError(t, err)
	worker, err := NewGrowthRegistrationWorker(repository, cipher, GrowthRegistrationWorkerOptions{
		Endpoint:          traffic.URL,
		ServiceCredential: "contract-secret",
		PollInterval:      5 * time.Millisecond,
		RetryDelay:        func(int) time.Duration { return time.Millisecond },
	})
	require.NoError(t, err)
	worker.Start()
	t.Cleanup(worker.Stop)

	registeredAt := time.Date(2026, 8, 11, 8, 0, 0, 0, time.UTC)
	ctx := WithGrowthRegistrationSession(context.Background(), "growth-session")
	require.NoError(t, recorder.RecordSuccessfulRegistration(ctx, &User{ID: 42, CreatedAt: registeredAt}))

	request := requireGrowthContractEvent(t, requests)
	require.Equal(t, "Service contract-secret", request.Authorization)
	require.NotEmpty(t, request.RequestID)
	require.Equal(t, "aiwelink", request.Payload.SiteID)
	require.Equal(t, "42", request.Payload.ExternalUserID)
	require.Equal(t, "growth-session", *request.Payload.GrowthSession)
	requireGrowthContractEvent(t, repository.delivered)
}

func TestGrowthRegistrationContractRetriesTemporaryTrafficFailureWithoutBlockingRegistration(t *testing.T) {
	traffic := httptest.NewServer(http.HandlerFunc(func(writer http.ResponseWriter, _ *http.Request) {
		writer.Header().Set("Content-Type", "application/json")
		writer.WriteHeader(http.StatusServiceUnavailable)
		_, _ = writer.Write([]byte(`{"error":{"code":"temporarily_unavailable"},"request_id":"traffic-503"}`))
	}))
	defer traffic.Close()

	repository := newGrowthContractRepository()
	cipher, err := NewGrowthRegistrationCipher(growthRegistrationTestKey)
	require.NoError(t, err)
	recorder, err := NewGrowthRegistrationRecorder(repository, cipher, "aiwelink")
	require.NoError(t, err)
	worker, err := NewGrowthRegistrationWorker(repository, cipher, GrowthRegistrationWorkerOptions{
		Endpoint:          traffic.URL,
		ServiceCredential: "contract-secret",
		PollInterval:      5 * time.Millisecond,
		RetryDelay:        func(int) time.Duration { return time.Millisecond },
	})
	require.NoError(t, err)
	worker.Start()
	t.Cleanup(worker.Stop)

	started := time.Now()
	err = recorder.RecordSuccessfulRegistration(context.Background(), &User{
		ID:        43,
		CreatedAt: time.Date(2026, 8, 11, 8, 1, 0, 0, time.UTC),
	})
	require.NoError(t, err)
	require.Less(t, time.Since(started), 100*time.Millisecond)

	retry := requireGrowthContractEvent(t, repository.retried)
	require.NotNil(t, retry.HTTPStatus)
	require.Equal(t, http.StatusServiceUnavailable, *retry.HTTPStatus)
	require.Equal(t, "temporarily_unavailable", retry.ErrorCode)
	require.Equal(t, "traffic-503", retry.RequestID)
	select {
	case errorCode := <-repository.deadLetter:
		t.Fatalf("temporary failure was dead-lettered: %s", errorCode)
	default:
	}
}
