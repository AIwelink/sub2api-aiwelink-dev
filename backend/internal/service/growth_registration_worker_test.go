package service

import (
	"bytes"
	"context"
	"errors"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
)

type growthRegistrationTransition struct {
	kind        string
	outboxID    int64
	workerID    string
	availableAt time.Time
	httpStatus  *int
	errorCode   string
	requestID   string
	detached    bool
	deadline    time.Duration
}

type growthRegistrationWorkerRepoStub struct {
	mu          sync.Mutex
	events      []GrowthRegistrationOutboxEvent
	claimErr    error
	claimLimit  int
	claimLease  time.Duration
	claimWorker string
	claimCalls  int
	transitions []growthRegistrationTransition
	active      int
	maxActive   int
	block       time.Duration
}

func (*growthRegistrationWorkerRepoStub) RecordSuccessfulRegistration(context.Context, uuid.UUID, string, string, time.Time, *string) error {
	return nil
}

func (r *growthRegistrationWorkerRepoStub) Claim(_ context.Context, workerID string, limit int, lease time.Duration) ([]GrowthRegistrationOutboxEvent, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.claimWorker = workerID
	r.claimLimit = limit
	r.claimLease = lease
	r.claimCalls++
	if r.claimErr != nil {
		return nil, r.claimErr
	}
	count := limit
	if count <= 0 || count > len(r.events) {
		count = len(r.events)
	}
	events := append([]GrowthRegistrationOutboxEvent(nil), r.events[:count]...)
	r.events = append([]GrowthRegistrationOutboxEvent(nil), r.events[count:]...)
	return events, nil
}

func (r *growthRegistrationWorkerRepoStub) DeleteClaimed(ctx context.Context, outboxID int64, workerID string) error {
	r.recordTransition(ctx, growthRegistrationTransition{kind: "delete", outboxID: outboxID, workerID: workerID})
	return nil
}

func (r *growthRegistrationWorkerRepoStub) RetryClaimed(ctx context.Context, outboxID int64, workerID string, availableAt time.Time, httpStatus *int, errorCode, requestID string) error {
	r.recordTransition(ctx, growthRegistrationTransition{kind: "retry", outboxID: outboxID, workerID: workerID, availableAt: availableAt, httpStatus: cloneInt(httpStatus), errorCode: errorCode, requestID: requestID})
	return nil
}

func (r *growthRegistrationWorkerRepoStub) DeadLetterClaimed(ctx context.Context, outboxID int64, workerID string, httpStatus *int, errorCode, requestID string) error {
	r.recordTransition(ctx, growthRegistrationTransition{kind: "dead", outboxID: outboxID, workerID: workerID, httpStatus: cloneInt(httpStatus), errorCode: errorCode, requestID: requestID})
	return nil
}

func (r *growthRegistrationWorkerRepoStub) recordTransition(ctx context.Context, transition growthRegistrationTransition) {
	deadline, ok := ctx.Deadline()
	transition.detached = ctx.Err() == nil
	if ok {
		transition.deadline = time.Until(deadline)
	}
	r.mu.Lock()
	r.active++
	if r.active > r.maxActive {
		r.maxActive = r.active
	}
	r.mu.Unlock()
	if r.block > 0 {
		time.Sleep(r.block)
	}
	r.mu.Lock()
	r.active--
	r.transitions = append(r.transitions, transition)
	r.mu.Unlock()
}

func (r *growthRegistrationWorkerRepoStub) snapshot() []growthRegistrationTransition {
	r.mu.Lock()
	defer r.mu.Unlock()
	return append([]growthRegistrationTransition(nil), r.transitions...)
}

func cloneInt(value *int) *int {
	if value == nil {
		return nil
	}
	copy := *value
	return &copy
}

type growthRegistrationRoundTripper func(*http.Request) (*http.Response, error)

func (fn growthRegistrationRoundTripper) RoundTrip(request *http.Request) (*http.Response, error) {
	return fn(request)
}

type growthRegistrationErrorBody struct{}

func (growthRegistrationErrorBody) Read([]byte) (int, error) { return 0, errors.New("read failed") }
func (growthRegistrationErrorBody) Close() error             { return nil }

func newGrowthRegistrationWorkerForTest(t *testing.T, repo *growthRegistrationWorkerRepoStub, endpoint string, client *http.Client) (*GrowthRegistrationWorker, *GrowthRegistrationCipher) {
	t.Helper()
	cipher, err := NewGrowthRegistrationCipher(growthRegistrationTestKey)
	require.NoError(t, err)
	worker, err := NewGrowthRegistrationWorker(repo, cipher, GrowthRegistrationWorkerOptions{
		Endpoint:          endpoint,
		ServiceCredential: "service-secret",
		HTTPClient:        client,
		WorkerID:          "worker-test",
		Now:               func() time.Time { return time.Date(2026, 8, 9, 4, 5, 6, 0, time.UTC) },
		RetryDelay:        func(int) time.Duration { return 7 * time.Second },
	})
	require.NoError(t, err)
	return worker, cipher
}

func growthRegistrationEvent(t *testing.T, cipher *GrowthRegistrationCipher, session *string) GrowthRegistrationOutboxEvent {
	t.Helper()
	var ciphertext *string
	if session != nil {
		encrypted, err := cipher.Encrypt(*session)
		require.NoError(t, err)
		ciphertext = &encrypted
	}
	return GrowthRegistrationOutboxEvent{
		OutboxID:                42,
		SourceRegistrationID:    uuid.MustParse("11111111-2222-4333-8444-555555555555"),
		SiteID:                  "aiwelink",
		ExternalUserID:          "12345",
		RegisteredAt:            time.Date(2026, 8, 9, 12, 34, 56, 789123456, time.FixedZone("CST", 8*60*60)),
		GrowthSessionCiphertext: ciphertext,
		AttemptCount:            2,
	}
}

func TestGrowthRegistrationWorkerPostsExactStablePayloadAndFreshRequestIDs(t *testing.T) {
	var bodies [][]byte
	var requestIDs []string
	var headers http.Header
	server := httptest.NewServer(http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		body, err := io.ReadAll(request.Body)
		require.NoError(t, err)
		bodies = append(bodies, body)
		requestIDs = append(requestIDs, request.Header.Get("X-Request-ID"))
		headers = request.Header.Clone()
		if len(bodies) == 1 {
			writer.WriteHeader(http.StatusServiceUnavailable)
			_, _ = io.WriteString(writer, `{"error":{"code":"temporarily_unavailable"},"request_id":"growth-request"}`)
			return
		}
		writer.WriteHeader(http.StatusCreated)
		_, _ = io.WriteString(writer, `{}`)
	}))
	defer server.Close()

	repo := &growthRegistrationWorkerRepoStub{}
	worker, cipher := newGrowthRegistrationWorkerForTest(t, repo, server.URL, server.Client())
	session := "growth-session"
	event := growthRegistrationEvent(t, cipher, &session)
	require.NoError(t, worker.processEvent(context.Background(), event))
	require.NoError(t, worker.processEvent(context.Background(), event))

	want := `{"site_id":"aiwelink","external_user_id":"12345","source_registration_id":"11111111-2222-4333-8444-555555555555","registered_at":"2026-08-09T04:34:56.789123456Z","growth_session":"growth-session"}`
	require.Equal(t, [][]byte{[]byte(want), []byte(want)}, bodies)
	require.Equal(t, "Service service-secret", headers.Get("Authorization"))
	require.Equal(t, "application/json", headers.Get("Content-Type"))
	require.Len(t, requestIDs, 2)
	require.NotEqual(t, requestIDs[0], requestIDs[1])
	_, err := uuid.Parse(requestIDs[0])
	require.NoError(t, err)
	require.Equal(t, []string{"retry", "delete"}, []string{repo.snapshot()[0].kind, repo.snapshot()[1].kind})
}

func TestGrowthRegistrationWorkerSendsNullSession(t *testing.T) {
	var body []byte
	server := httptest.NewServer(http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		body, _ = io.ReadAll(request.Body)
		writer.WriteHeader(http.StatusOK)
	}))
	defer server.Close()
	repo := &growthRegistrationWorkerRepoStub{}
	worker, cipher := newGrowthRegistrationWorkerForTest(t, repo, server.URL, server.Client())
	require.NoError(t, worker.processEvent(context.Background(), growthRegistrationEvent(t, cipher, nil)))
	require.True(t, bytes.HasSuffix(body, []byte(`"growth_session":null}`)))
}

func TestGrowthRegistrationWorkerClassifiesResponses(t *testing.T) {
	tests := []struct {
		name      string
		status    int
		body      string
		want      string
		wantCode  string
		wantReqID string
	}{
		{name: "ok", status: http.StatusOK, body: `{}`, want: "delete"},
		{name: "created", status: http.StatusCreated, body: `{}`, want: "delete"},
		{name: "temporary exact trimmed", status: http.StatusServiceUnavailable, body: `{"error":{"code":" temporarily_unavailable "},"request_id":" request-1 "}`, want: "retry", wantCode: "temporarily_unavailable", wantReqID: "request-1"},
		{name: "adapter unavailable", status: http.StatusServiceUnavailable, body: `{"error":{"code":"source_adapter_unavailable"}}`, want: "retry", wantCode: "source_adapter_unavailable"},
		{name: "request timeout", status: http.StatusRequestTimeout, body: `{}`, want: "retry"},
		{name: "too early", status: http.StatusTooEarly, body: `{}`, want: "retry"},
		{name: "rate limited", status: http.StatusTooManyRequests, body: `{"error":{"code":"rate_limited"}}`, want: "retry", wantCode: "rate_limited"},
		{name: "internal", status: http.StatusInternalServerError, body: `{}`, want: "retry"},
		{name: "bad gateway", status: http.StatusBadGateway, body: `{}`, want: "retry"},
		{name: "other 503", status: http.StatusServiceUnavailable, body: `{"error":{"code":"TEMPORARILY_UNAVAILABLE"}}`, want: "retry", wantCode: "TEMPORARILY_UNAVAILABLE"},
		{name: "gateway timeout", status: http.StatusGatewayTimeout, body: `{}`, want: "retry"},
		{name: "unauthorized", status: http.StatusUnauthorized, body: `{"error":{"code":"unauthorized"}}`, want: "dead", wantCode: "unauthorized"},
		{name: "forbidden", status: http.StatusForbidden, body: `{}`, want: "dead"},
		{name: "conflict", status: http.StatusConflict, body: `{}`, want: "dead"},
		{name: "unprocessable", status: http.StatusUnprocessableEntity, body: `{}`, want: "dead"},
		{name: "not implemented", status: http.StatusNotImplemented, body: `{}`, want: "dead"},
		{name: "empty 503", status: http.StatusServiceUnavailable, body: ``, want: "retry", wantCode: "invalid_response"},
		{name: "malformed 503", status: http.StatusServiceUnavailable, body: `{`, want: "retry", wantCode: "invalid_response"},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(writer http.ResponseWriter, _ *http.Request) {
				writer.WriteHeader(tc.status)
				_, _ = io.WriteString(writer, tc.body)
			}))
			defer server.Close()
			repo := &growthRegistrationWorkerRepoStub{}
			worker, cipher := newGrowthRegistrationWorkerForTest(t, repo, server.URL, server.Client())
			require.NoError(t, worker.processEvent(context.Background(), growthRegistrationEvent(t, cipher, nil)))
			transition := repo.snapshot()[0]
			require.Equal(t, tc.want, transition.kind)
			require.Equal(t, tc.wantCode, transition.errorCode)
			require.Equal(t, tc.wantReqID, transition.requestID)
			if tc.want != "delete" {
				require.Equal(t, tc.status, *transition.httpStatus)
			}
		})
	}
}

func TestGrowthRegistrationWorkerRetriesTransportAndBodyReadErrors(t *testing.T) {
	tests := []struct {
		name      string
		transport growthRegistrationRoundTripper
	}{
		{name: "transport", transport: func(*http.Request) (*http.Response, error) { return nil, errors.New("network down") }},
		{name: "body read", transport: func(*http.Request) (*http.Response, error) {
			return &http.Response{StatusCode: http.StatusOK, Header: make(http.Header), Body: growthRegistrationErrorBody{}}, nil
		}},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			repo := &growthRegistrationWorkerRepoStub{}
			client := &http.Client{Transport: tc.transport}
			worker, cipher := newGrowthRegistrationWorkerForTest(t, repo, "http://worker.internal/bind", nil)
			worker.httpClient = client
			require.NoError(t, worker.processEvent(context.Background(), growthRegistrationEvent(t, cipher, nil)))
			transition := repo.snapshot()[0]
			require.Equal(t, "retry", transition.kind)
			require.Nil(t, transition.httpStatus)
			require.Empty(t, transition.errorCode)
			require.Empty(t, transition.requestID)
		})
	}
}

func TestGrowthRegistrationWorkerDeadLettersPermanentStatusWhenBodyReadFails(t *testing.T) {
	repo := &growthRegistrationWorkerRepoStub{}
	client := &http.Client{Transport: growthRegistrationRoundTripper(func(*http.Request) (*http.Response, error) {
		return &http.Response{
			StatusCode: http.StatusUnprocessableEntity,
			Header:     make(http.Header),
			Body:       growthRegistrationErrorBody{},
		}, nil
	})}
	worker, cipher := newGrowthRegistrationWorkerForTest(t, repo, "http://worker.internal/bind", nil)
	worker.httpClient = client

	require.NoError(t, worker.processEvent(context.Background(), growthRegistrationEvent(t, cipher, nil)))
	transition := repo.snapshot()[0]
	require.Equal(t, "dead", transition.kind)
	require.NotNil(t, transition.httpStatus)
	require.Equal(t, http.StatusUnprocessableEntity, *transition.httpStatus)
	require.Equal(t, "invalid_response", transition.errorCode)
	require.Empty(t, transition.requestID)
}

func TestGrowthRegistrationWorkerClassifiesOversizedBodiesByStatus(t *testing.T) {
	for _, test := range []struct {
		status int
		want   string
	}{
		{status: http.StatusCreated, want: "dead"},
		{status: http.StatusOK, want: "dead"},
		{status: http.StatusServiceUnavailable, want: "retry"},
	} {
		t.Run(http.StatusText(test.status), func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(writer http.ResponseWriter, _ *http.Request) {
				writer.WriteHeader(test.status)
				_, _ = io.WriteString(writer, strings.Repeat("x", growthRegistrationMaxResponseBodyBytes+1))
			}))
			defer server.Close()
			repo := &growthRegistrationWorkerRepoStub{}
			worker, cipher := newGrowthRegistrationWorkerForTest(t, repo, server.URL, server.Client())
			require.NoError(t, worker.processEvent(context.Background(), growthRegistrationEvent(t, cipher, nil)))
			require.Equal(t, test.want, repo.snapshot()[0].kind)
		})
	}
}

func TestGrowthRegistrationWorkerDeadLettersDecryptFailureWithoutSending(t *testing.T) {
	requests := 0
	server := httptest.NewServer(http.HandlerFunc(func(http.ResponseWriter, *http.Request) { requests++ }))
	defer server.Close()
	repo := &growthRegistrationWorkerRepoStub{}
	worker, cipher := newGrowthRegistrationWorkerForTest(t, repo, server.URL, server.Client())
	event := growthRegistrationEvent(t, cipher, nil)
	bad := "v1:not-valid"
	event.GrowthSessionCiphertext = &bad
	require.NoError(t, worker.processEvent(context.Background(), event))
	require.Zero(t, requests)
	transition := repo.snapshot()[0]
	require.Equal(t, "dead", transition.kind)
	require.Nil(t, transition.httpStatus)
	require.Equal(t, "decrypt_failed", transition.errorCode)
}

func TestGrowthRegistrationWorkerSanitizesResponseMetadata(t *testing.T) {
	longCode := strings.Repeat("a", growthRegistrationMaxErrorCodeBytes+1)
	longRequestID := strings.Repeat("b", growthRegistrationMaxRequestIDBytes+1)
	for _, body := range []string{
		`{"error":{"code":"bad\ncode","request_id":"bad\tid"}}`,
		`{"error":{"code":"` + longCode + `","request_id":"` + longRequestID + `"}}`,
	} {
		server := httptest.NewServer(http.HandlerFunc(func(writer http.ResponseWriter, _ *http.Request) {
			writer.WriteHeader(http.StatusBadRequest)
			_, _ = io.WriteString(writer, body)
		}))
		repo := &growthRegistrationWorkerRepoStub{}
		worker, cipher := newGrowthRegistrationWorkerForTest(t, repo, server.URL, server.Client())
		require.NoError(t, worker.processEvent(context.Background(), growthRegistrationEvent(t, cipher, nil)))
		transition := repo.snapshot()[0]
		require.Empty(t, transition.errorCode)
		require.Empty(t, transition.requestID)
		server.Close()
	}
}

func TestGrowthRegistrationWorkerTransitionsUseDetachedTwoSecondContext(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(writer http.ResponseWriter, _ *http.Request) { writer.WriteHeader(http.StatusOK) }))
	defer server.Close()
	repo := &growthRegistrationWorkerRepoStub{}
	worker, cipher := newGrowthRegistrationWorkerForTest(t, repo, server.URL, server.Client())
	parent, cancel := context.WithCancel(context.Background())
	cancel()
	require.NoError(t, worker.processEvent(parent, growthRegistrationEvent(t, cipher, nil)))
	transition := repo.snapshot()[0]
	require.True(t, transition.detached)
	require.Greater(t, transition.deadline, 1500*time.Millisecond)
	require.LessOrEqual(t, transition.deadline, growthRegistrationTransitionTimeout)
}

func TestGrowthRegistrationWorkerBatchDefaultsAndSerialProcessing(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(writer http.ResponseWriter, _ *http.Request) { writer.WriteHeader(http.StatusOK) }))
	defer server.Close()
	repo := &growthRegistrationWorkerRepoStub{block: 20 * time.Millisecond}
	worker, cipher := newGrowthRegistrationWorkerForTest(t, repo, server.URL, server.Client())
	repo.events = []GrowthRegistrationOutboxEvent{
		growthRegistrationEvent(t, cipher, nil),
		growthRegistrationEvent(t, cipher, nil),
		growthRegistrationEvent(t, cipher, nil),
	}
	for index := range repo.events {
		repo.events[index].OutboxID = int64(index + 1)
	}
	require.NoError(t, worker.processBatch(context.Background()))
	require.Equal(t, 1, repo.claimLimit)
	require.Equal(t, growthRegistrationDefaultLease, repo.claimLease)
	require.Equal(t, "worker-test", repo.claimWorker)
	require.Equal(t, 1, repo.maxActive)
	require.Len(t, repo.snapshot(), 3)
}

func TestGrowthRegistrationWorkerClaimsAtMostOneEventAtATime(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(writer http.ResponseWriter, _ *http.Request) {
		writer.WriteHeader(http.StatusOK)
	}))
	defer server.Close()
	repo := &growthRegistrationWorkerRepoStub{}
	cipher, err := NewGrowthRegistrationCipher(growthRegistrationTestKey)
	require.NoError(t, err)
	worker, err := NewGrowthRegistrationWorker(repo, cipher, GrowthRegistrationWorkerOptions{
		Endpoint:          server.URL,
		ServiceCredential: "service-secret",
		HTTPClient:        server.Client(),
		WorkerID:          "worker-test",
		BatchSize:         25,
	})
	require.NoError(t, err)

	require.NoError(t, worker.processBatch(context.Background()))
	require.Equal(t, 1, repo.claimLimit)
	require.Equal(t, 1, repo.claimCalls)
}

func TestGrowthRegistrationWorkerLeaseCoversSingleDeliveryTimeout(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(writer http.ResponseWriter, _ *http.Request) {
		writer.WriteHeader(http.StatusOK)
	}))
	defer server.Close()
	repo := &growthRegistrationWorkerRepoStub{}
	cipher, err := NewGrowthRegistrationCipher(growthRegistrationTestKey)
	require.NoError(t, err)
	worker, err := NewGrowthRegistrationWorker(repo, cipher, GrowthRegistrationWorkerOptions{
		Endpoint:          server.URL,
		ServiceCredential: "service-secret",
		HTTPClient:        server.Client(),
		WorkerID:          "worker-test",
		Lease:             time.Second,
		TransitionTimeout: 3 * time.Second,
	})
	require.NoError(t, err)
	require.Equal(t, worker.httpClient.Timeout+worker.transitionTimeout, worker.lease)
}

func TestGrowthRegistrationWorkerLifecycleIsNilSafeAndIdempotent(t *testing.T) {
	var nilWorker *GrowthRegistrationWorker
	require.NotPanics(t, func() { nilWorker.Start(); nilWorker.Stop() })

	server := httptest.NewServer(http.HandlerFunc(func(writer http.ResponseWriter, _ *http.Request) {
		writer.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	repo := &growthRegistrationWorkerRepoStub{}
	worker, _ := newGrowthRegistrationWorkerForTest(t, repo, server.URL, server.Client())
	require.NotPanics(t, func() { worker.Stop(); worker.Stop(); worker.Start(); worker.Start(); worker.Stop() })

	worker, _ = newGrowthRegistrationWorkerForTest(t, repo, server.URL, server.Client())
	worker.Start()
	require.Eventually(t, worker.Running, time.Second, 10*time.Millisecond)
	worker.Start()
	worker.Stop()
	worker.Stop()
	require.False(t, worker.Running())
	worker.Start()
	require.Eventually(t, worker.Running, time.Second, 10*time.Millisecond)
	worker.Stop()
	require.False(t, worker.Running())
}

func TestGrowthRegistrationRetryDelayIsJitteredExponentialAndBounded(t *testing.T) {
	seen := map[time.Duration]struct{}{}
	for attempt := 1; attempt <= 40; attempt++ {
		delay := growthRegistrationRetryDelay(attempt)
		require.Greater(t, delay, time.Duration(0))
		require.LessOrEqual(t, delay, time.Hour)
		seen[delay] = struct{}{}
		if attempt <= 8 {
			base := 2 * time.Second * time.Duration(1<<(attempt-1))
			require.GreaterOrEqual(t, delay, time.Duration(float64(base)*0.8))
			require.LessOrEqual(t, delay, time.Duration(float64(base)*1.2))
		}
	}
	require.Greater(t, len(seen), 1)
}

func TestGrowthRegistrationWorkerRejectsRedirectWithoutForwardingCredentials(t *testing.T) {
	forwarded := 0
	destination := httptest.NewServer(http.HandlerFunc(func(_ http.ResponseWriter, request *http.Request) {
		forwarded++
		require.Empty(t, request.Header.Get("Authorization"))
	}))
	defer destination.Close()
	redirect := httptest.NewServer(http.HandlerFunc(func(writer http.ResponseWriter, _ *http.Request) {
		http.Redirect(writer, &http.Request{}, destination.URL, http.StatusFound)
	}))
	defer redirect.Close()
	repo := &growthRegistrationWorkerRepoStub{}
	worker, cipher := newGrowthRegistrationWorkerForTest(t, repo, redirect.URL, redirect.Client())
	require.NoError(t, worker.processEvent(context.Background(), growthRegistrationEvent(t, cipher, nil)))
	require.Zero(t, forwarded)
	require.Equal(t, "dead", repo.snapshot()[0].kind)
}

func TestGrowthRegistrationEndpointValidationMatrix(t *testing.T) {
	valid := []string{
		"https://growth.example.com/bind", "https://203.0.113.4:8443/bind", "https://[2001:db8::1]/bind",
		"http://localhost/bind", "http://127.0.0.1/bind", "http://[::1]/bind", "http://10.0.0.2/bind",
		"http://172.16.1.2/bind", "http://172.31.255.254/bind", "http://192.168.1.2/bind",
		"http://[fc00::1]/bind", "http://host.docker.internal/bind", "http://growth-api:8080/bind", "http://growth.service.internal/bind",
	}
	invalid := []string{
		"", "ftp://growth.example.com/bind", "http://example.com/bind", "http://8.8.8.8/bind", "http://172.32.0.1/bind",
		"https://user:pass@growth.example.com/bind", "https://growth.example.com/bind#fragment", "https://-bad.example/bind",
		"https://bad_host.example/bind", "https://growth.example.com:bad/bind", "https:///bind", "http://a.b/bind",
	}
	for _, endpoint := range valid {
		t.Run("valid "+endpoint, func(t *testing.T) {
			parsed, err := validateGrowthRegistrationEndpoint(endpoint)
			require.NoError(t, err)
			require.NotNil(t, parsed)
		})
	}
	for _, endpoint := range invalid {
		t.Run("invalid "+endpoint, func(t *testing.T) {
			parsed, err := validateGrowthRegistrationEndpoint(endpoint)
			require.Error(t, err)
			require.Nil(t, parsed)
		})
	}
}

func TestGrowthRegistrationHTTPClientProxyRedirectAndBoundsPolicy(t *testing.T) {
	t.Setenv("HTTP_PROXY", "http://127.0.0.1:3128")
	t.Setenv("HTTPS_PROXY", "http://127.0.0.1:3129")
	t.Setenv("NO_PROXY", "")
	privateURL, err := url.Parse("http://10.0.0.3/bind")
	require.NoError(t, err)
	httpsURL, err := url.Parse("https://growth.example.com/bind")
	require.NoError(t, err)

	privateClient, err := newGrowthRegistrationHTTPClient(nil, privateURL, time.Second, 2*time.Second)
	require.NoError(t, err)
	httpsClient, err := newGrowthRegistrationHTTPClient(nil, httpsURL, time.Second, 2*time.Second)
	require.NoError(t, err)
	privateTransport, ok := privateClient.Transport.(*http.Transport)
	require.True(t, ok)
	httpsTransport, ok := httpsClient.Transport.(*http.Transport)
	require.True(t, ok)
	proxy, err := privateTransport.Proxy(&http.Request{URL: privateURL})
	require.NoError(t, err)
	require.Nil(t, proxy)
	proxy, err = httpsTransport.Proxy(&http.Request{URL: httpsURL})
	require.NoError(t, err)
	require.Equal(t, "http://127.0.0.1:3129", proxy.String())
	require.Equal(t, int64(growthRegistrationMaxResponseHeaderBytes), privateTransport.MaxResponseHeaderBytes)
	require.Positive(t, privateTransport.MaxIdleConns)
	require.Positive(t, privateTransport.MaxIdleConnsPerHost)
	require.Positive(t, privateTransport.MaxConnsPerHost)
	require.Equal(t, 3*time.Second, privateClient.Timeout)
	require.ErrorIs(t, privateClient.CheckRedirect(nil, nil), http.ErrUseLastResponse)

	suppliedTransport := &http.Transport{}
	supplied := &http.Client{Transport: suppliedTransport}
	clone, err := newGrowthRegistrationHTTPClient(supplied, privateURL, time.Second, 2*time.Second)
	require.NoError(t, err)
	require.NotSame(t, supplied, clone)
	require.NotSame(t, suppliedTransport, clone.Transport)
	require.Nil(t, supplied.CheckRedirect)
	require.ErrorIs(t, clone.CheckRedirect(nil, nil), http.ErrUseLastResponse)
	cloneTransport, ok := clone.Transport.(*http.Transport)
	require.True(t, ok)
	proxy, err = cloneTransport.Proxy(&http.Request{URL: privateURL})
	require.NoError(t, err)
	require.Nil(t, proxy)
	require.LessOrEqual(t, cloneTransport.MaxResponseHeaderBytes, int64(growthRegistrationMaxResponseHeaderBytes))
	require.Positive(t, cloneTransport.ResponseHeaderTimeout)
	require.LessOrEqual(t, clone.Timeout, 3*time.Second)

	unsafeClient := &http.Client{Transport: growthRegistrationRoundTripper(func(*http.Request) (*http.Response, error) {
		return nil, errors.New("not called")
	})}
	clone, err = newGrowthRegistrationHTTPClient(unsafeClient, privateURL, time.Second, 2*time.Second)
	require.Error(t, err)
	require.Nil(t, clone)
}

func TestGrowthRegistrationWorkerRejectsHeadersOverSixteenKiB(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(writer http.ResponseWriter, _ *http.Request) {
		writer.Header().Set("X-Large", strings.Repeat("x", growthRegistrationMaxResponseHeaderBytes+1))
		writer.WriteHeader(http.StatusOK)
	}))
	defer server.Close()
	repo := &growthRegistrationWorkerRepoStub{}
	worker, cipher := newGrowthRegistrationWorkerForTest(t, repo, server.URL, nil)
	require.NoError(t, worker.processEvent(context.Background(), growthRegistrationEvent(t, cipher, nil)))
	require.Equal(t, "retry", repo.snapshot()[0].kind)
}

func TestGrowthRegistrationWorkerDoesNotLogSecretsOrResponseBodies(t *testing.T) {
	var logs bytes.Buffer
	previous := slog.Default()
	slog.SetDefault(slog.New(slog.NewTextHandler(&logs, nil)))
	defer slog.SetDefault(previous)
	responseSecret := "response-body-secret"
	server := httptest.NewServer(http.HandlerFunc(func(writer http.ResponseWriter, _ *http.Request) {
		writer.WriteHeader(http.StatusBadRequest)
		_, _ = io.WriteString(writer, `{"error":{"code":"bad_request"},"detail":"`+responseSecret+`"}`)
	}))
	defer server.Close()
	repo := &growthRegistrationWorkerRepoStub{}
	worker, cipher := newGrowthRegistrationWorkerForTest(t, repo, server.URL, server.Client())
	cookieSecret := "cookie-plaintext-secret"
	event := growthRegistrationEvent(t, cipher, &cookieSecret)
	ciphertextSecret := *event.GrowthSessionCiphertext
	require.NoError(t, worker.processEvent(context.Background(), event))
	output := logs.String()
	for _, secret := range []string{"service-secret", cookieSecret, ciphertextSecret, growthRegistrationTestKey, responseSecret} {
		require.NotContains(t, output, secret)
	}
}
