package service

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/Wei-Shaw/sub2api/internal/config"
)

type GrowthRegistrationRuntime struct {
	recorder GrowthRegistrationRecorder
	worker   *GrowthRegistrationWorker
}

var _ GrowthRegistrationRecorder = (*GrowthRegistrationRuntime)(nil)

func ProvideGrowthRegistrationRuntime(
	cfg *config.Config,
	repository GrowthRegistrationOutboxRepository,
) (*GrowthRegistrationRuntime, error) {
	if cfg == nil {
		return nil, errors.New("growth registration configuration is required")
	}
	if !cfg.GrowthRegistration.Enabled {
		return &GrowthRegistrationRuntime{}, nil
	}
	if repository == nil {
		return nil, errors.New("growth registration outbox repository is required")
	}
	if cfg.GrowthRegistration.ConnectTimeoutSeconds <= 0 {
		return nil, errors.New("growth registration connect timeout must be positive")
	}
	if cfg.GrowthRegistration.ReadTimeoutSeconds <= 0 {
		return nil, errors.New("growth registration read timeout must be positive")
	}
	endpoint, err := validateGrowthRegistrationEndpoint(cfg.GrowthRegistration.Endpoint)
	if err != nil {
		return nil, err
	}
	cipher, err := NewGrowthRegistrationCipher(cfg.GrowthRegistration.OutboxEncryptionKey)
	if err != nil {
		return nil, fmt.Errorf("invalid growth registration outbox encryption key: %w", err)
	}
	recorder, err := NewGrowthRegistrationRecorder(repository, cipher, cfg.GrowthRegistration.SiteID)
	if err != nil {
		return nil, err
	}
	httpClient, err := newGrowthRegistrationHTTPClient(
		nil,
		endpoint,
		time.Duration(cfg.GrowthRegistration.ConnectTimeoutSeconds)*time.Second,
		time.Duration(cfg.GrowthRegistration.ReadTimeoutSeconds)*time.Second,
	)
	if err != nil {
		return nil, err
	}
	worker, err := NewGrowthRegistrationWorker(repository, cipher, GrowthRegistrationWorkerOptions{
		Endpoint:          endpoint.String(),
		ServiceCredential: cfg.GrowthRegistration.ServiceCredential,
		HTTPClient:        httpClient,
	})
	if err != nil {
		return nil, err
	}
	runtime := &GrowthRegistrationRuntime{
		recorder: recorder,
		worker:   worker,
	}
	worker.Start()
	return runtime, nil
}

func (r *GrowthRegistrationRuntime) RecordSuccessfulRegistration(ctx context.Context, user *User) error {
	if r == nil || r.recorder == nil {
		return nil
	}
	return r.recorder.RecordSuccessfulRegistration(ctx, user)
}

func (r *GrowthRegistrationRuntime) Stop() {
	if r == nil || r.worker == nil {
		return
	}
	r.worker.Stop()
}
