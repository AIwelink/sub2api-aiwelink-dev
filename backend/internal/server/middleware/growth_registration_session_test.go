package middleware

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/Wei-Shaw/sub2api/internal/config"
	"github.com/Wei-Shaw/sub2api/internal/service"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/require"
)

func TestGrowthRegistrationSessionCapturesBoundedCookie(t *testing.T) {
	gin.SetMode(gin.TestMode)

	for _, tc := range []struct {
		name      string
		cookie    *http.Cookie
		wantValue string
		wantOK    bool
	}{
		{name: "valid", cookie: &http.Cookie{Name: "awl_growth_sid", Value: "growth-session"}, wantValue: "growth-session", wantOK: true},
		{name: "maximum length", cookie: &http.Cookie{Name: "awl_growth_sid", Value: strings.Repeat("a", 64)}, wantValue: strings.Repeat("a", 64), wantOK: true},
		{name: "missing"},
		{name: "empty", cookie: &http.Cookie{Name: "awl_growth_sid", Value: ""}},
		{name: "over limit", cookie: &http.Cookie{Name: "awl_growth_sid", Value: strings.Repeat("a", 65)}},
	} {
		t.Run(tc.name, func(t *testing.T) {
			var captured string
			var capturedOK bool
			router := gin.New()
			router.Use(GrowthRegistrationSession(config.GrowthRegistrationConfig{
				Enabled:    true,
				CookieName: "awl_growth_sid",
			}))
			router.POST("/api/v1/auth/register", func(c *gin.Context) {
				captured, capturedOK = service.GrowthRegistrationSessionFromContext(c.Request.Context())
				c.Status(http.StatusNoContent)
			})

			recorder := httptest.NewRecorder()
			request := httptest.NewRequest(http.MethodPost, "/api/v1/auth/register", nil)
			if tc.cookie != nil {
				request.AddCookie(tc.cookie)
			}
			router.ServeHTTP(recorder, request)

			require.Equal(t, http.StatusNoContent, recorder.Code)
			require.Equal(t, tc.wantOK, capturedOK)
			require.Equal(t, tc.wantValue, captured)
		})
	}
}

func TestGrowthRegistrationSessionUsesConfiguredCookieName(t *testing.T) {
	gin.SetMode(gin.TestMode)

	var captured string
	var capturedOK bool
	router := gin.New()
	router.Use(GrowthRegistrationSession(config.GrowthRegistrationConfig{
		Enabled:    true,
		CookieName: "custom_growth_session",
	}))
	router.POST("/api/v1/auth/register", func(c *gin.Context) {
		captured, capturedOK = service.GrowthRegistrationSessionFromContext(c.Request.Context())
		c.Status(http.StatusNoContent)
	})

	recorder := httptest.NewRecorder()
	request := httptest.NewRequest(http.MethodPost, "/api/v1/auth/register", nil)
	request.AddCookie(&http.Cookie{Name: "awl_growth_sid", Value: "wrong"})
	request.AddCookie(&http.Cookie{Name: "custom_growth_session", Value: "right"})
	router.ServeHTTP(recorder, request)

	require.Equal(t, http.StatusNoContent, recorder.Code)
	require.True(t, capturedOK)
	require.Equal(t, "right", captured)
}

func TestGrowthRegistrationSessionIsNoOpWhenDisabled(t *testing.T) {
	gin.SetMode(gin.TestMode)

	var captured string
	var capturedOK bool
	router := gin.New()
	router.Use(GrowthRegistrationSession(config.GrowthRegistrationConfig{
		Enabled:    false,
		CookieName: "awl_growth_sid",
	}))
	router.POST("/api/v1/auth/register", func(c *gin.Context) {
		captured, capturedOK = service.GrowthRegistrationSessionFromContext(c.Request.Context())
		c.Status(http.StatusNoContent)
	})

	recorder := httptest.NewRecorder()
	request := httptest.NewRequest(http.MethodPost, "/api/v1/auth/register", nil)
	request.AddCookie(&http.Cookie{Name: "awl_growth_sid", Value: "growth-session"})
	router.ServeHTTP(recorder, request)

	require.Equal(t, http.StatusNoContent, recorder.Code)
	require.False(t, capturedOK)
	require.Empty(t, captured)
}

func TestGrowthRegistrationSessionIgnoresNonRegistrationRequests(t *testing.T) {
	gin.SetMode(gin.TestMode)

	for _, tc := range []struct {
		name   string
		method string
		path   string
	}{
		{name: "login", method: http.MethodPost, path: "/api/v1/auth/login"},
		{name: "oauth", method: http.MethodGet, path: "/api/v1/auth/google"},
		{name: "wrong method", method: http.MethodGet, path: "/api/v1/auth/register"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			var capturedOK bool
			router := gin.New()
			router.Use(GrowthRegistrationSession(config.GrowthRegistrationConfig{
				Enabled:    true,
				CookieName: "awl_growth_sid",
			}))
			router.Handle(tc.method, tc.path, func(c *gin.Context) {
				_, capturedOK = service.GrowthRegistrationSessionFromContext(c.Request.Context())
				c.Status(http.StatusNoContent)
			})

			request := httptest.NewRequest(tc.method, tc.path, nil)
			request.AddCookie(&http.Cookie{Name: "awl_growth_sid", Value: "growth-session"})
			recorder := httptest.NewRecorder()
			router.ServeHTTP(recorder, request)

			require.Equal(t, http.StatusNoContent, recorder.Code)
			require.False(t, capturedOK)
		})
	}
}
