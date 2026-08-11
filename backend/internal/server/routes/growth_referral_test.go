package routes

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/Wei-Shaw/sub2api/internal/config"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/require"
)

func growthReferralRouter(enabled bool) *gin.Engine {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	RegisterGrowthReferralRoutes(r, config.GrowthRegistrationConfig{
		Enabled:         enabled,
		ReferralBaseURL: "https://aiwelink.cc/r",
	})
	return r
}

func TestGrowthReferralRedirectsValidCodeWithoutBackendDependency(t *testing.T) {
	r := growthReferralRouter(true)
	req := httptest.NewRequest(http.MethodGet, "/r/BCDEFGHJ?next=https://evil.example&foo=bar", nil)
	rec := httptest.NewRecorder()

	r.ServeHTTP(rec, req)

	require.Equal(t, http.StatusFound, rec.Code)
	require.Equal(t, "https://aiwelink.cc/r/bcdefghj?entry=api", rec.Header().Get("Location"))
	require.Equal(t, "no-store", rec.Header().Get("Cache-Control"))
}

func TestGrowthReferralFallsBackForInvalidCode(t *testing.T) {
	r := growthReferralRouter(true)
	req := httptest.NewRequest(http.MethodGet, "/r/abcd0fgh", nil)
	rec := httptest.NewRecorder()

	r.ServeHTTP(rec, req)

	require.Equal(t, http.StatusFound, rec.Code)
	require.Equal(t, "/", rec.Header().Get("Location"))
	require.Equal(t, "no-store", rec.Header().Get("Cache-Control"))
}

func TestGrowthReferralFallsBackWhenDisabled(t *testing.T) {
	r := growthReferralRouter(false)
	req := httptest.NewRequest(http.MethodGet, "/r/abcdefgh", nil)
	rec := httptest.NewRecorder()

	r.ServeHTTP(rec, req)

	require.Equal(t, http.StatusFound, rec.Code)
	require.Equal(t, "/", rec.Header().Get("Location"))
}

func TestGrowthReferralDoesNotRegisterNonGetMethods(t *testing.T) {
	r := growthReferralRouter(true)
	req := httptest.NewRequest(http.MethodPost, "/r/abcdefgh", nil)
	rec := httptest.NewRecorder()

	r.ServeHTTP(rec, req)

	require.Equal(t, http.StatusNotFound, rec.Code)
}
