package service

import (
	"context"
	"errors"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/google/uuid"
)

const defaultGrowthRegistrationRecordTimeout = 2 * time.Second

type GrowthRegistrationRecorder interface {
	RecordSuccessfulRegistration(ctx context.Context, user *User) error
}

type GrowthRegistrationRecordRepository interface {
	RecordSuccessfulRegistration(
		ctx context.Context,
		sourceRegistrationID uuid.UUID,
		siteID string,
		externalUserID string,
		registeredAt time.Time,
		growthSessionCiphertext *string,
	) error
}

type GrowthRegistrationOutboxRepository interface {
	GrowthRegistrationRecordRepository
	Claim(
		ctx context.Context,
		workerID string,
		limit int,
		lease time.Duration,
	) ([]GrowthRegistrationOutboxEvent, error)
	DeleteClaimed(ctx context.Context, outboxID int64, workerID string) error
	RetryClaimed(
		ctx context.Context,
		outboxID int64,
		workerID string,
		availableAt time.Time,
		httpStatus *int,
		errorCode string,
		requestID string,
	) error
	DeadLetterClaimed(
		ctx context.Context,
		outboxID int64,
		workerID string,
		httpStatus *int,
		errorCode string,
		requestID string,
	) error
}

type GrowthRegistrationOutboxEvent struct {
	OutboxID                int64
	SourceRegistrationID    uuid.UUID
	SiteID                  string
	ExternalUserID          string
	RegisteredAt            time.Time
	GrowthSessionCiphertext *string
	AttemptCount            int
}

type growthRegistrationRecorder struct {
	repository              GrowthRegistrationRecordRepository
	cipher                  *GrowthRegistrationCipher
	siteID                  string
	newSourceRegistrationID func() (uuid.UUID, error)
}

func NewGrowthRegistrationRecorder(
	repository GrowthRegistrationRecordRepository,
	cipher *GrowthRegistrationCipher,
	siteID string,
) (GrowthRegistrationRecorder, error) {
	if repository == nil {
		return nil, errors.New("growth registration repository is required")
	}
	if cipher == nil {
		return nil, errors.New("growth registration cipher is required")
	}
	siteID = strings.TrimSpace(siteID)
	if siteID == "" {
		return nil, errors.New("growth registration site id is required")
	}
	return &growthRegistrationRecorder{
		repository:              repository,
		cipher:                  cipher,
		siteID:                  siteID,
		newSourceRegistrationID: uuid.NewRandom,
	}, nil
}

func (r *growthRegistrationRecorder) RecordSuccessfulRegistration(ctx context.Context, user *User) error {
	if ctx == nil {
		return errors.New("growth registration context is required")
	}
	if user == nil {
		return errors.New("growth registration user is required")
	}
	if user.ID <= 0 {
		return errors.New("growth registration user id must be positive")
	}
	if user.CreatedAt.IsZero() {
		return errors.New("growth registration user created at is required")
	}

	var ciphertext *string
	if growthSession, ok := GrowthRegistrationSessionFromContext(ctx); ok {
		encrypted, err := r.cipher.Encrypt(growthSession)
		if err != nil {
			return fmt.Errorf("encrypt growth registration session: %w", err)
		}
		ciphertext = &encrypted
	}

	sourceRegistrationID, err := r.newSourceRegistrationID()
	if err != nil {
		return fmt.Errorf("generate growth registration source id: %w", err)
	}

	persistCtx, cancel := context.WithTimeout(
		context.WithoutCancel(ctx),
		defaultGrowthRegistrationRecordTimeout,
	)
	defer cancel()
	err = r.repository.RecordSuccessfulRegistration(
		persistCtx,
		sourceRegistrationID,
		r.siteID,
		strconv.FormatInt(user.ID, 10),
		user.CreatedAt.UTC(),
		ciphertext,
	)
	if err != nil {
		return fmt.Errorf("record growth registration: %w", err)
	}
	return nil
}
