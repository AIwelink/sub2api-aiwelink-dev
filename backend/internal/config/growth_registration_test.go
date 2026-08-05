package config

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestLoadGrowthRegistrationDefaults(t *testing.T) {
	resetViperWithJWTSecret(t)

	cfg, err := Load()
	require.NoError(t, err)

	require.False(t, cfg.GrowthRegistration.Enabled)
	require.Equal(t, "http://127.0.0.1:8081/internal/growth/registrations/bind", cfg.GrowthRegistration.Endpoint)
	require.Equal(t, "aiwelink", cfg.GrowthRegistration.SiteID)
	require.Empty(t, cfg.GrowthRegistration.ServiceCredential)
	require.Empty(t, cfg.GrowthRegistration.OutboxEncryptionKey)
	require.Equal(t, "awl_growth_sid", cfg.GrowthRegistration.CookieName)
	require.Equal(t, 2, cfg.GrowthRegistration.ConnectTimeoutSeconds)
	require.Equal(t, 5, cfg.GrowthRegistration.ReadTimeoutSeconds)
}

func TestLoadGrowthRegistrationFromEnvironment(t *testing.T) {
	resetViperWithJWTSecret(t)
	t.Setenv("GROWTH_REGISTRATION_ENABLED", "true")
	t.Setenv("GROWTH_REGISTRATION_ENDPOINT", "https://growth.example.com/internal/growth/registrations/bind")
	t.Setenv("GROWTH_REGISTRATION_SITE_ID", "sub2-test")
	t.Setenv("GROWTH_REGISTRATION_SERVICE_CREDENTIAL", "service-secret")
	t.Setenv("GROWTH_REGISTRATION_OUTBOX_ENCRYPTION_KEY", "0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef")
	t.Setenv("GROWTH_REGISTRATION_COOKIE_NAME", "growth_session")
	t.Setenv("GROWTH_REGISTRATION_CONNECT_TIMEOUT_SECONDS", "3")
	t.Setenv("GROWTH_REGISTRATION_READ_TIMEOUT_SECONDS", "7")

	cfg, err := Load()
	require.NoError(t, err)

	require.True(t, cfg.GrowthRegistration.Enabled)
	require.Equal(t, "https://growth.example.com/internal/growth/registrations/bind", cfg.GrowthRegistration.Endpoint)
	require.Equal(t, "sub2-test", cfg.GrowthRegistration.SiteID)
	require.Equal(t, "service-secret", cfg.GrowthRegistration.ServiceCredential)
	require.Equal(t, "0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef", cfg.GrowthRegistration.OutboxEncryptionKey)
	require.Equal(t, "growth_session", cfg.GrowthRegistration.CookieName)
	require.Equal(t, 3, cfg.GrowthRegistration.ConnectTimeoutSeconds)
	require.Equal(t, 7, cfg.GrowthRegistration.ReadTimeoutSeconds)
}
