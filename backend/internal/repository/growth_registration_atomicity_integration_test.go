//go:build integration

package repository

import (
	"context"
	"fmt"
	"strconv"
	"strings"
	"testing"
	"time"

	dbent "github.com/Wei-Shaw/sub2api/ent"
	"github.com/Wei-Shaw/sub2api/internal/config"
	"github.com/Wei-Shaw/sub2api/internal/service"
	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
)

type growthRegistrationSettingRepository struct {
	values map[string]string
}

func (r *growthRegistrationSettingRepository) Get(context.Context, string) (*service.Setting, error) {
	return nil, service.ErrSettingNotFound
}

func (r *growthRegistrationSettingRepository) GetValue(_ context.Context, key string) (string, error) {
	if value, ok := r.values[key]; ok {
		return value, nil
	}
	return "", service.ErrSettingNotFound
}

func (r *growthRegistrationSettingRepository) Set(context.Context, string, string) error {
	return nil
}

func (r *growthRegistrationSettingRepository) GetMultiple(_ context.Context, keys []string) (map[string]string, error) {
	values := make(map[string]string, len(keys))
	for _, key := range keys {
		if value, ok := r.values[key]; ok {
			values[key] = value
		}
	}
	return values, nil
}

func (r *growthRegistrationSettingRepository) SetMultiple(context.Context, map[string]string) error {
	return nil
}

func (r *growthRegistrationSettingRepository) GetAll(context.Context) (map[string]string, error) {
	values := make(map[string]string, len(r.values))
	for key, value := range r.values {
		values[key] = value
	}
	return values, nil
}

func (r *growthRegistrationSettingRepository) Delete(context.Context, string) error {
	return nil
}

func TestGrowthRegistrationRepositoriesReuseOuterTransaction(t *testing.T) {
	ctx := context.Background()
	client := testEntClient(t)
	userRepo := NewUserRepository(client, integrationDB)
	outboxRepo := NewGrowthRegistrationOutboxRepository(integrationDB)
	email := fmt.Sprintf("growth-outer-tx-%s@example.com", uuid.NewString())
	sourceRegistrationID := uuid.New()
	t.Cleanup(func() {
		_, _ = integrationDB.Exec(`DELETE FROM growth_registration_outbox WHERE source_registration_id = $1`, sourceRegistrationID)
		_, _ = integrationDB.Exec(`DELETE FROM auth_identities WHERE provider_subject = $1`, email)
		_, _ = integrationDB.Exec(`DELETE FROM users WHERE email = $1`, email)
	})

	tx, err := client.Tx(ctx)
	require.NoError(t, err)
	rolledBack := false
	defer func() {
		if !rolledBack {
			_ = tx.Rollback()
		}
	}()
	txCtx := dbent.NewTxContext(ctx, tx)
	user := &service.User{
		Email:        email,
		PasswordHash: "hash",
		Role:         service.RoleUser,
		Status:       service.StatusActive,
		Concurrency:  1,
	}
	require.NoError(t, userRepo.CreateWithEmailAliasGuard(txCtx, user))
	require.Positive(t, user.ID)
	require.False(t, user.CreatedAt.IsZero())
	require.NoError(t, outboxRepo.RecordSuccessfulRegistration(
		txCtx,
		sourceRegistrationID,
		"aiwelink",
		strconv.FormatInt(user.ID, 10),
		user.CreatedAt,
		nil,
	))
	require.NoError(t, tx.Rollback())
	rolledBack = true

	require.Equal(t, 0, growthRegistrationRowCount(t, `SELECT COUNT(*) FROM users WHERE email = $1`, email))
	require.Equal(t, 0, growthRegistrationRowCount(
		t,
		`SELECT COUNT(*) FROM growth_registration_outbox WHERE source_registration_id = $1`,
		sourceRegistrationID,
	))
}

func TestAuthRegistrationRollsBackWhenGrowthOutboxInsertFails(t *testing.T) {
	email := fmt.Sprintf("growth-failed-registration-%s@example.com", uuid.NewString())
	t.Cleanup(func() {
		_, _ = integrationDB.Exec(`DELETE FROM auth_identities WHERE provider_subject = $1`, email)
		_, _ = integrationDB.Exec(`DELETE FROM users WHERE email = $1`, email)
	})

	authService := newGrowthRegistrationAuthService(t, strings.Repeat("s", 101))
	token, user, err := authService.Register(context.Background(), email, "password")

	require.ErrorIs(t, err, service.ErrServiceUnavailable)
	require.Empty(t, token)
	require.Nil(t, user)
	require.Equal(t, 0, growthRegistrationRowCount(t, `SELECT COUNT(*) FROM users WHERE email = $1`, email))
}

func TestAuthRegistrationCommitsUserAndGrowthOutboxTogether(t *testing.T) {
	email := fmt.Sprintf("growth-success-registration-%s@example.com", uuid.NewString())
	siteID := "aiwelink-test-" + uuid.NewString()
	var userID int64
	t.Cleanup(func() {
		_, _ = integrationDB.Exec(`DELETE FROM growth_registration_outbox WHERE site_id = $1`, siteID)
		_, _ = integrationDB.Exec(`DELETE FROM auth_identities WHERE provider_subject = $1`, email)
		if userID > 0 {
			_, _ = integrationDB.Exec(`DELETE FROM users WHERE id = $1`, userID)
		} else {
			_, _ = integrationDB.Exec(`DELETE FROM users WHERE email = $1`, email)
		}
	})

	authService := newGrowthRegistrationAuthService(t, siteID)
	token, user, err := authService.Register(context.Background(), email, "password")

	require.NoError(t, err)
	require.NotEmpty(t, token)
	require.NotNil(t, user)
	userID = user.ID
	require.Equal(t, 1, growthRegistrationRowCount(t, `SELECT COUNT(*) FROM users WHERE id = $1`, user.ID))
	require.Equal(t, 1, growthRegistrationRowCount(
		t,
		`SELECT COUNT(*) FROM growth_registration_outbox WHERE site_id = $1 AND external_user_id = $2`,
		siteID,
		strconv.FormatInt(user.ID, 10),
	))
}

func newGrowthRegistrationAuthService(t *testing.T, siteID string) *service.AuthService {
	t.Helper()
	cfg := &config.Config{
		JWT: config.JWTConfig{
			Secret:     "growth-registration-integration-secret",
			ExpireHour: 1,
		},
		Default: config.DefaultConfig{
			UserBalance:     3.5,
			UserConcurrency: 2,
		},
	}
	settingService := service.NewSettingService(&growthRegistrationSettingRepository{values: map[string]string{
		service.SettingKeyRegistrationEnabled:                 "true",
		service.SettingKeyAuthSourceDefaultEmailGrantOnSignup: "false",
	}}, cfg)
	userRepo := NewUserRepository(integrationEntClient, integrationDB)
	authService := service.NewAuthService(
		integrationEntClient,
		userRepo,
		nil,
		nil,
		cfg,
		settingService,
		nil,
		nil,
		nil,
		nil,
		nil,
		nil,
		nil,
	)
	cipher, err := service.NewGrowthRegistrationCipher(
		strings.Repeat("01", 32),
	)
	require.NoError(t, err)
	recorder, err := service.NewGrowthRegistrationRecorder(
		NewGrowthRegistrationOutboxRepository(integrationDB),
		cipher,
		siteID,
	)
	require.NoError(t, err)
	authService.SetGrowthRegistrationRecorder(recorder)
	return authService
}

func growthRegistrationRowCount(t *testing.T, query string, args ...any) int {
	t.Helper()
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	var count int
	require.NoError(t, integrationDB.QueryRowContext(ctx, query, args...).Scan(&count))
	return count
}
