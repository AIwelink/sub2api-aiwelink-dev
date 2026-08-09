package service

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"math/rand/v2"
	"net"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/google/uuid"
	"golang.org/x/net/http/httpproxy"
)

const (
	growthRegistrationDefaultBatchSize       = 25
	growthRegistrationDefaultLease           = 30 * time.Second
	growthRegistrationDefaultPollInterval    = 500 * time.Millisecond
	growthRegistrationTransitionTimeout      = 2 * time.Second
	growthRegistrationMaxResponseBodyBytes   = 4 * 1024
	growthRegistrationMaxResponseHeaderBytes = 16 * 1024
	growthRegistrationMaxErrorCodeBytes      = 100
	growthRegistrationMaxRequestIDBytes      = 64
)

type GrowthRegistrationWorkerOptions struct {
	Endpoint          string
	ServiceCredential string
	HTTPClient        *http.Client
	WorkerID          string
	BatchSize         int
	Lease             time.Duration
	PollInterval      time.Duration
	TransitionTimeout time.Duration
	Now               func() time.Time
	RetryDelay        func(attempt int) time.Duration
}

type GrowthRegistrationWorker struct {
	repository        GrowthRegistrationOutboxRepository
	cipher            *GrowthRegistrationCipher
	endpoint          string
	serviceCredential string
	httpClient        *http.Client
	workerID          string
	batchSize         int
	lease             time.Duration
	pollInterval      time.Duration
	transitionTimeout time.Duration
	now               func() time.Time
	retryDelay        func(int) time.Duration

	lifecycleMu sync.Mutex
	started     bool
	stopped     bool
	cancel      context.CancelFunc
	done        chan struct{}
}

type growthRegistrationPayload struct {
	SiteID               string  `json:"site_id"`
	ExternalUserID       string  `json:"external_user_id"`
	SourceRegistrationID string  `json:"source_registration_id"`
	RegisteredAt         string  `json:"registered_at"`
	GrowthSession        *string `json:"growth_session"`
}

type growthRegistrationErrorResponse struct {
	Error struct {
		Code string `json:"code"`
	} `json:"error"`
	RequestID string `json:"request_id"`
}

func NewGrowthRegistrationWorker(
	repository GrowthRegistrationOutboxRepository,
	cipher *GrowthRegistrationCipher,
	options GrowthRegistrationWorkerOptions,
) (*GrowthRegistrationWorker, error) {
	if repository == nil {
		return nil, errors.New("growth registration outbox repository is required")
	}
	if cipher == nil {
		return nil, errors.New("growth registration cipher is required")
	}
	endpoint, err := validateGrowthRegistrationEndpoint(options.Endpoint)
	if err != nil {
		return nil, err
	}
	credential := strings.TrimSpace(options.ServiceCredential)
	if credential == "" {
		return nil, errors.New("growth registration service credential is required")
	}
	httpClient, err := newGrowthRegistrationHTTPClient(
		options.HTTPClient,
		endpoint,
		2*time.Second,
		5*time.Second,
	)
	if err != nil {
		return nil, err
	}
	workerID := strings.TrimSpace(options.WorkerID)
	if workerID == "" {
		workerID = uuid.NewString()
	}
	if options.BatchSize <= 0 {
		options.BatchSize = growthRegistrationDefaultBatchSize
	}
	if options.Lease <= 0 {
		options.Lease = growthRegistrationDefaultLease
	}
	if options.PollInterval <= 0 {
		options.PollInterval = growthRegistrationDefaultPollInterval
	}
	if options.TransitionTimeout <= 0 {
		options.TransitionTimeout = growthRegistrationTransitionTimeout
	}
	if options.Now == nil {
		options.Now = func() time.Time { return time.Now().UTC() }
	}
	if options.RetryDelay == nil {
		options.RetryDelay = growthRegistrationRetryDelay
	}
	return &GrowthRegistrationWorker{
		repository:        repository,
		cipher:            cipher,
		endpoint:          endpoint.String(),
		serviceCredential: credential,
		httpClient:        httpClient,
		workerID:          workerID,
		batchSize:         options.BatchSize,
		lease:             options.Lease,
		pollInterval:      options.PollInterval,
		transitionTimeout: options.TransitionTimeout,
		now:               options.Now,
		retryDelay:        options.RetryDelay,
	}, nil
}

func newGrowthRegistrationHTTPClient(
	base *http.Client,
	endpoint *url.URL,
	connectTimeout time.Duration,
	readTimeout time.Duration,
) (*http.Client, error) {
	if endpoint == nil || endpoint.Hostname() == "" {
		return nil, errors.New("growth registration endpoint is required")
	}
	var client http.Client
	var transport *http.Transport
	if base == nil {
		transport = &http.Transport{}
	} else {
		client = *base
		switch configured := base.Transport.(type) {
		case nil:
			transport = http.DefaultTransport.(*http.Transport).Clone()
		case *http.Transport:
			transport = configured.Clone()
		default:
			return nil, errors.New("growth registration HTTP client transport must be *http.Transport")
		}
	}
	if connectTimeout <= 0 {
		connectTimeout = 2 * time.Second
	}
	if readTimeout <= 0 {
		readTimeout = 5 * time.Second
	}
	directProxy := func(*http.Request) (*url.URL, error) { return nil, nil }
	proxy := directProxy
	if strings.EqualFold(endpoint.Scheme, "https") {
		proxyFromEnvironment := httpproxy.FromEnvironment().ProxyFunc()
		proxy = func(request *http.Request) (*url.URL, error) {
			if request == nil || request.URL == nil {
				return nil, nil
			}
			return proxyFromEnvironment(request.URL)
		}
	}
	transport.Proxy = proxy
	if transport.DialContext == nil {
		transport.DialContext = (&net.Dialer{Timeout: connectTimeout, KeepAlive: 30 * time.Second}).DialContext
	}
	transport.ForceAttemptHTTP2 = true
	if transport.MaxIdleConns <= 0 {
		transport.MaxIdleConns = 20
	}
	if transport.MaxIdleConnsPerHost <= 0 {
		transport.MaxIdleConnsPerHost = 10
	}
	if transport.MaxConnsPerHost <= 0 {
		transport.MaxConnsPerHost = 10
	}
	if transport.IdleConnTimeout <= 0 {
		transport.IdleConnTimeout = 90 * time.Second
	}
	if transport.TLSHandshakeTimeout <= 0 || transport.TLSHandshakeTimeout > connectTimeout {
		transport.TLSHandshakeTimeout = connectTimeout
	}
	if transport.ResponseHeaderTimeout <= 0 || transport.ResponseHeaderTimeout > readTimeout {
		transport.ResponseHeaderTimeout = readTimeout
	}
	if transport.ExpectContinueTimeout <= 0 || transport.ExpectContinueTimeout > time.Second {
		transport.ExpectContinueTimeout = time.Second
	}
	if transport.MaxResponseHeaderBytes <= 0 || transport.MaxResponseHeaderBytes > growthRegistrationMaxResponseHeaderBytes {
		transport.MaxResponseHeaderBytes = growthRegistrationMaxResponseHeaderBytes
	}
	client.Transport = transport
	maxTimeout := connectTimeout + readTimeout
	if client.Timeout <= 0 || client.Timeout > maxTimeout {
		client.Timeout = maxTimeout
	}
	client.CheckRedirect = rejectGrowthRegistrationRedirect
	return &client, nil
}

func rejectGrowthRegistrationRedirect(*http.Request, []*http.Request) error {
	return http.ErrUseLastResponse
}

func (w *GrowthRegistrationWorker) Start() {
	if w == nil {
		return
	}
	w.lifecycleMu.Lock()
	if w.started {
		w.lifecycleMu.Unlock()
		return
	}
	ctx, cancel := context.WithCancel(context.Background())
	done := make(chan struct{})
	w.stopped = false
	w.started = true
	w.cancel = cancel
	w.done = done
	w.lifecycleMu.Unlock()

	go func() {
		defer close(done)
		w.run(ctx)
	}()
}

func (w *GrowthRegistrationWorker) Stop() {
	if w == nil {
		return
	}
	w.lifecycleMu.Lock()
	if !w.started {
		w.lifecycleMu.Unlock()
		return
	}
	w.stopped = true
	cancel := w.cancel
	done := w.done
	w.lifecycleMu.Unlock()
	if cancel != nil {
		cancel()
	}
	if done != nil {
		<-done
	}
	w.lifecycleMu.Lock()
	if w.done == done {
		w.started = false
		w.cancel = nil
		w.done = nil
	}
	w.lifecycleMu.Unlock()
}

func (w *GrowthRegistrationWorker) Running() bool {
	if w == nil {
		return false
	}
	w.lifecycleMu.Lock()
	defer w.lifecycleMu.Unlock()
	return w.started && !w.stopped
}

func (w *GrowthRegistrationWorker) run(ctx context.Context) {
	ticker := time.NewTicker(w.pollInterval)
	defer ticker.Stop()
	for {
		if err := w.processBatch(ctx); err != nil && ctx.Err() == nil {
			slog.Warn("growth registration worker batch failed", "error_class", "database")
		}
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
		}
	}
}

func (w *GrowthRegistrationWorker) processBatch(ctx context.Context) error {
	if ctx == nil {
		ctx = context.Background()
	}
	events, err := w.repository.Claim(ctx, w.workerID, w.batchSize, w.lease)
	if err != nil {
		return fmt.Errorf("claim growth registration outbox: %w", err)
	}
	for _, event := range events {
		if err := ctx.Err(); err != nil {
			return err
		}
		if err := w.processEvent(ctx, event); err != nil {
			slog.Warn(
				"growth registration event transition failed",
				"outbox_id", event.OutboxID,
				"source_registration_id", event.SourceRegistrationID.String(),
				"error_class", "database",
			)
		}
	}
	return nil
}

func (w *GrowthRegistrationWorker) processEvent(
	ctx context.Context,
	event GrowthRegistrationOutboxEvent,
) error {
	if ctx == nil {
		ctx = context.Background()
	}
	payload, err := w.payload(event)
	if err != nil {
		return w.deadLetter(event, nil, "decrypt_failed", "")
	}
	body, err := json.Marshal(payload)
	if err != nil {
		return w.deadLetter(event, nil, "invalid_outbox_payload", "")
	}
	request, err := http.NewRequestWithContext(ctx, http.MethodPost, w.endpoint, bytes.NewReader(body))
	if err != nil {
		return w.deadLetter(event, nil, "invalid_outbox_payload", "")
	}
	request.Header.Set("Authorization", "Service "+w.serviceCredential)
	request.Header.Set("Content-Type", "application/json")
	request.Header.Set("X-Request-ID", uuid.NewString())

	response, err := w.httpClient.Do(request)
	if err != nil {
		return w.retry(event, nil, "", "")
	}
	defer func() { _ = response.Body.Close() }()

	responseBody, err := io.ReadAll(io.LimitReader(response.Body, growthRegistrationMaxResponseBodyBytes+1))
	if err != nil {
		return w.retry(event, nil, "", "")
	}
	status := response.StatusCode
	if len(responseBody) > growthRegistrationMaxResponseBodyBytes {
		return w.deadLetter(event, &status, "invalid_response", "")
	}
	if status == http.StatusOK || status == http.StatusCreated {
		return w.delete(event)
	}

	errorCode, requestID, validJSON := parseGrowthRegistrationError(responseBody)
	if !validJSON {
		errorCode = "invalid_response"
		requestID = ""
	}
	if status == http.StatusServiceUnavailable &&
		(errorCode == "temporarily_unavailable" || errorCode == "source_adapter_unavailable") {
		return w.retry(event, &status, errorCode, requestID)
	}
	return w.deadLetter(event, &status, errorCode, requestID)
}

func (w *GrowthRegistrationWorker) payload(
	event GrowthRegistrationOutboxEvent,
) (growthRegistrationPayload, error) {
	var growthSession *string
	if event.GrowthSessionCiphertext != nil {
		plaintext, err := w.cipher.Decrypt(*event.GrowthSessionCiphertext)
		if err != nil {
			return growthRegistrationPayload{}, err
		}
		growthSession = &plaintext
	}
	return growthRegistrationPayload{
		SiteID:               event.SiteID,
		ExternalUserID:       event.ExternalUserID,
		SourceRegistrationID: event.SourceRegistrationID.String(),
		RegisteredAt:         event.RegisteredAt.UTC().Format(time.RFC3339Nano),
		GrowthSession:        growthSession,
	}, nil
}

func (w *GrowthRegistrationWorker) delete(event GrowthRegistrationOutboxEvent) error {
	ctx, cancel := context.WithTimeout(context.Background(), w.transitionTimeout)
	defer cancel()
	err := w.repository.DeleteClaimed(ctx, event.OutboxID, w.workerID)
	if err != nil {
		return err
	}
	return nil
}

func (w *GrowthRegistrationWorker) retry(
	event GrowthRegistrationOutboxEvent,
	httpStatus *int,
	errorCode string,
	requestID string,
) error {
	availableAt := w.now().UTC().Add(w.retryDelay(event.AttemptCount + 1))
	ctx, cancel := context.WithTimeout(context.Background(), w.transitionTimeout)
	defer cancel()
	err := w.repository.RetryClaimed(
		ctx,
		event.OutboxID,
		w.workerID,
		availableAt,
		httpStatus,
		errorCode,
		requestID,
	)
	if err != nil {
		return err
	}
	w.logTransition("growth registration delivery scheduled for retry", event, httpStatus, errorCode, "scheduled")
	return nil
}

func (w *GrowthRegistrationWorker) deadLetter(
	event GrowthRegistrationOutboxEvent,
	httpStatus *int,
	errorCode string,
	requestID string,
) error {
	ctx, cancel := context.WithTimeout(context.Background(), w.transitionTimeout)
	defer cancel()
	err := w.repository.DeadLetterClaimed(
		ctx,
		event.OutboxID,
		w.workerID,
		httpStatus,
		errorCode,
		requestID,
	)
	if err != nil {
		return err
	}
	w.logTransition("growth registration delivery moved to dead letter", event, httpStatus, errorCode, "dead_lettered")
	return nil
}

func (w *GrowthRegistrationWorker) logTransition(
	message string,
	event GrowthRegistrationOutboxEvent,
	httpStatus *int,
	errorCode string,
	state string,
) {
	attributes := []any{
		"outbox_id", event.OutboxID,
		"source_registration_id", event.SourceRegistrationID.String(),
		"delivery_state", state,
	}
	if httpStatus != nil {
		attributes = append(attributes, "http_status", *httpStatus)
	}
	if errorCode != "" {
		attributes = append(attributes, "error_code", errorCode)
	}
	slog.Warn(message, attributes...)
}

func parseGrowthRegistrationError(body []byte) (string, string, bool) {
	if len(body) == 0 {
		return "", "", false
	}
	var payload growthRegistrationErrorResponse
	if err := json.Unmarshal(body, &payload); err != nil {
		return "", "", false
	}
	return boundedGrowthRegistrationIdentifier(payload.Error.Code, growthRegistrationMaxErrorCodeBytes),
		boundedGrowthRegistrationIdentifier(payload.RequestID, growthRegistrationMaxRequestIDBytes),
		true
}

func boundedGrowthRegistrationIdentifier(value string, maxBytes int) string {
	value = strings.TrimSpace(value)
	if value == "" || len(value) > maxBytes {
		return ""
	}
	for index := 0; index < len(value); index++ {
		character := value[index]
		if (character < 'a' || character > 'z') &&
			(character < 'A' || character > 'Z') &&
			(character < '0' || character > '9') &&
			character != '_' && character != '-' && character != ':' && character != '.' {
			return ""
		}
	}
	return value
}

func growthRegistrationRetryDelay(attempt int) time.Duration {
	if attempt < 1 {
		attempt = 1
	}
	var base time.Duration
	if attempt >= 12 {
		base = time.Hour
	} else {
		base = 2 * time.Second * time.Duration(1<<(attempt-1))
		if base > time.Hour {
			base = time.Hour
		}
	}
	delay := time.Duration(float64(base) * (0.8 + rand.Float64()*0.4))
	if delay > time.Hour {
		return time.Hour
	}
	return delay
}

func validateGrowthRegistrationEndpoint(rawEndpoint string) (*url.URL, error) {
	endpoint, err := url.Parse(strings.TrimSpace(rawEndpoint))
	if err != nil || endpoint.Scheme == "" || endpoint.Host == "" || endpoint.Hostname() == "" ||
		endpoint.User != nil || endpoint.Fragment != "" {
		return nil, errors.New("growth registration endpoint must be an HTTP(S) URL without credentials or fragment")
	}
	hostname := strings.ToLower(endpoint.Hostname())
	if !validGrowthRegistrationEndpointHost(hostname) || !validGrowthRegistrationEndpointPort(endpoint) {
		return nil, errors.New("growth registration endpoint host or port is invalid")
	}
	switch strings.ToLower(endpoint.Scheme) {
	case "https":
	case "http":
		if !isAllowedGrowthRegistrationHTTPHost(hostname) {
			return nil, errors.New("growth registration HTTP endpoint must use an explicit private host")
		}
	default:
		return nil, errors.New("growth registration endpoint must use HTTP or HTTPS")
	}
	endpoint.Scheme = strings.ToLower(endpoint.Scheme)
	return endpoint, nil
}

func validGrowthRegistrationEndpointHost(hostname string) bool {
	if net.ParseIP(hostname) != nil {
		return true
	}
	return validGrowthRegistrationDNSName(hostname)
}

func validGrowthRegistrationEndpointPort(endpoint *url.URL) bool {
	if endpoint == nil || strings.HasSuffix(endpoint.Host, ":") {
		return false
	}
	port := endpoint.Port()
	if port == "" {
		return true
	}
	value, err := strconv.Atoi(port)
	return err == nil && value > 0 && value <= 65535
}

func isAllowedGrowthRegistrationHTTPHost(hostname string) bool {
	hostname = strings.ToLower(strings.TrimSpace(hostname))
	if hostname == "localhost" || hostname == "host.docker.internal" {
		return true
	}
	if ip := net.ParseIP(hostname); ip != nil {
		return ip.IsLoopback() || ip.IsPrivate()
	}
	if strings.HasSuffix(hostname, ".internal") {
		return validGrowthRegistrationDNSName(hostname)
	}
	return !strings.Contains(hostname, ".") && validGrowthRegistrationDNSLabel(hostname)
}

func validGrowthRegistrationDNSName(hostname string) bool {
	if len(hostname) == 0 || len(hostname) > 253 {
		return false
	}
	for _, label := range strings.Split(hostname, ".") {
		if !validGrowthRegistrationDNSLabel(label) {
			return false
		}
	}
	return true
}

func validGrowthRegistrationDNSLabel(label string) bool {
	if len(label) == 0 || len(label) > 63 ||
		!isGrowthRegistrationDNSAlphaNumeric(label[0]) ||
		!isGrowthRegistrationDNSAlphaNumeric(label[len(label)-1]) {
		return false
	}
	for index := 1; index < len(label)-1; index++ {
		if !isGrowthRegistrationDNSAlphaNumeric(label[index]) && label[index] != '-' {
			return false
		}
	}
	return true
}

func isGrowthRegistrationDNSAlphaNumeric(character byte) bool {
	return (character >= 'a' && character <= 'z') ||
		(character >= '0' && character <= '9')
}
