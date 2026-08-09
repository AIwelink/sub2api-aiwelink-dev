package service

import (
	"context"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/Wei-Shaw/sub2api/internal/config"
	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
)

type growthRegistrationRuntimeRecord struct {
	sourceRegistrationID    uuid.UUID
	siteID                  string
	externalUserID          string
	registeredAt            time.Time
	growthSessionCiphertext *string
}

type growthRegistrationRuntimeRepo struct {
	growthRegistrationWorkerRepoStub
	records []growthRegistrationRuntimeRecord
}

func (r *growthRegistrationRuntimeRepo) RecordSuccessfulRegistration(
	_ context.Context,
	sourceRegistrationID uuid.UUID,
	siteID string,
	externalUserID string,
	registeredAt time.Time,
	growthSessionCiphertext *string,
) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.records = append(r.records, growthRegistrationRuntimeRecord{
		sourceRegistrationID:    sourceRegistrationID,
		siteID:                  siteID,
		externalUserID:          externalUserID,
		registeredAt:            registeredAt,
		growthSessionCiphertext: growthSessionCiphertext,
	})
	return nil
}

func validGrowthRegistrationRuntimeConfig() *config.Config {
	return &config.Config{
		GrowthRegistration: config.GrowthRegistrationConfig{
			Enabled:               true,
			Endpoint:              "http://127.0.0.1:8081/internal/growth/registrations/bind",
			SiteID:                "aiwelink",
			ServiceCredential:     "service-credential",
			OutboxEncryptionKey:   strings.Repeat("01", 32),
			CookieName:            "awl_growth_sid",
			ConnectTimeoutSeconds: 2,
			ReadTimeoutSeconds:    5,
		},
	}
}

func TestGrowthRegistrationRuntimeImplementsRecorder(t *testing.T) {
	var recorder GrowthRegistrationRecorder = &GrowthRegistrationRuntime{}
	require.NotNil(t, recorder)
}

func TestProvideGrowthRegistrationRuntimeDisabledIsNoOpWithoutDependencies(t *testing.T) {
	cfg := &config.Config{
		GrowthRegistration: config.GrowthRegistrationConfig{
			Enabled:               false,
			Endpoint:              "http://public.example.com/invalid",
			ConnectTimeoutSeconds: -1,
			ReadTimeoutSeconds:    -1,
		},
	}

	runtime, err := ProvideGrowthRegistrationRuntime(cfg, nil)
	require.NoError(t, err)
	require.NotNil(t, runtime)
	require.Nil(t, runtime.recorder)
	require.Nil(t, runtime.worker)
	require.NoError(t, runtime.RecordSuccessfulRegistration(nil, nil))
	require.NotPanics(t, runtime.Stop)
	require.NotPanics(t, runtime.Stop)
}

func TestGrowthRegistrationRuntimeNilReceiverIsNoOp(t *testing.T) {
	var runtime *GrowthRegistrationRuntime
	require.NoError(t, runtime.RecordSuccessfulRegistration(nil, nil))
	require.NotPanics(t, runtime.Stop)
}

func TestProvideGrowthRegistrationRuntimeEnabledStartsAndDelegates(t *testing.T) {
	repository := &growthRegistrationRuntimeRepo{}
	runtime, err := ProvideGrowthRegistrationRuntime(validGrowthRegistrationRuntimeConfig(), repository)
	require.NoError(t, err)
	require.NotNil(t, runtime)
	require.NotNil(t, runtime.recorder)
	require.NotNil(t, runtime.worker)
	t.Cleanup(runtime.Stop)

	require.Eventually(t, func() bool {
		repository.mu.Lock()
		defer repository.mu.Unlock()
		return repository.claimWorker != ""
	}, time.Second, 10*time.Millisecond)

	createdAt := time.Date(2026, time.August, 9, 16, 0, 0, 123, time.FixedZone("UTC+8", 8*60*60))
	ctx := WithGrowthRegistrationSession(context.Background(), "growth-session-from-runtime")
	require.NoError(t, runtime.RecordSuccessfulRegistration(ctx, &User{ID: 42, CreatedAt: createdAt}))

	repository.mu.Lock()
	require.Len(t, repository.records, 1)
	record := repository.records[0]
	repository.mu.Unlock()
	require.NotEqual(t, uuid.Nil, record.sourceRegistrationID)
	require.Equal(t, "aiwelink", record.siteID)
	require.Equal(t, "42", record.externalUserID)
	require.Equal(t, createdAt.UTC(), record.registeredAt)
	require.NotNil(t, record.growthSessionCiphertext)
	require.NotContains(t, *record.growthSessionCiphertext, "growth-session-from-runtime")
	plaintext, err := runtime.worker.cipher.Decrypt(*record.growthSessionCiphertext)
	require.NoError(t, err)
	require.Equal(t, "growth-session-from-runtime", plaintext)

	require.NotPanics(t, runtime.Stop)
	require.NotPanics(t, runtime.Stop)
}

func TestProvideGrowthRegistrationRuntimeValidatesEnabledDependencies(t *testing.T) {
	runtime, err := ProvideGrowthRegistrationRuntime(nil, &growthRegistrationRuntimeRepo{})
	require.Error(t, err)
	require.Nil(t, runtime)

	runtime, err = ProvideGrowthRegistrationRuntime(validGrowthRegistrationRuntimeConfig(), nil)
	require.Error(t, err)
	require.Nil(t, runtime)
}

func TestProvideGrowthRegistrationRuntimeRejectsInvalidConfigWithoutStarting(t *testing.T) {
	tests := []struct {
		name   string
		mutate func(*config.GrowthRegistrationConfig)
	}{
		{name: "missing endpoint", mutate: func(cfg *config.GrowthRegistrationConfig) { cfg.Endpoint = "" }},
		{name: "endpoint userinfo", mutate: func(cfg *config.GrowthRegistrationConfig) {
			cfg.Endpoint = "https://user:password@growth.example.com/bind"
		}},
		{name: "endpoint fragment", mutate: func(cfg *config.GrowthRegistrationConfig) {
			cfg.Endpoint = "https://growth.example.com/bind#secret"
		}},
		{name: "public HTTP endpoint", mutate: func(cfg *config.GrowthRegistrationConfig) {
			cfg.Endpoint = "http://growth.example.com/bind"
		}},
		{name: "missing site id", mutate: func(cfg *config.GrowthRegistrationConfig) { cfg.SiteID = " " }},
		{name: "missing credential", mutate: func(cfg *config.GrowthRegistrationConfig) {
			cfg.ServiceCredential = " "
		}},
		{name: "invalid encryption key", mutate: func(cfg *config.GrowthRegistrationConfig) {
			cfg.OutboxEncryptionKey = "invalid"
		}},
		{name: "zero connect timeout", mutate: func(cfg *config.GrowthRegistrationConfig) {
			cfg.ConnectTimeoutSeconds = 0
		}},
		{name: "negative connect timeout", mutate: func(cfg *config.GrowthRegistrationConfig) {
			cfg.ConnectTimeoutSeconds = -1
		}},
		{name: "zero read timeout", mutate: func(cfg *config.GrowthRegistrationConfig) {
			cfg.ReadTimeoutSeconds = 0
		}},
		{name: "negative read timeout", mutate: func(cfg *config.GrowthRegistrationConfig) {
			cfg.ReadTimeoutSeconds = -1
		}},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			cfg := validGrowthRegistrationRuntimeConfig()
			test.mutate(&cfg.GrowthRegistration)
			repository := &growthRegistrationRuntimeRepo{}

			runtime, err := ProvideGrowthRegistrationRuntime(cfg, repository)
			require.Error(t, err)
			require.Nil(t, runtime)
			time.Sleep(10 * time.Millisecond)
			repository.mu.Lock()
			defer repository.mu.Unlock()
			require.Empty(t, repository.claimWorker)
		})
	}
}

func TestGrowthRegistrationProviderSetContainsRuntimeAndRecorderBinding(t *testing.T) {
	content, err := os.ReadFile("wire.go")
	require.NoError(t, err)
	wireSource := string(content)
	require.Contains(t, wireSource, "ProvideGrowthRegistrationRuntime")
	require.Contains(t, wireSource, "wire.Bind(new(GrowthRegistrationRecorder), new(*GrowthRegistrationRuntime))")
}
