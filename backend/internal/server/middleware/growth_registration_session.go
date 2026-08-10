package middleware

import (
	"net/http"
	"strings"

	"github.com/Wei-Shaw/sub2api/internal/config"
	"github.com/Wei-Shaw/sub2api/internal/service"

	"github.com/gin-gonic/gin"
)

const growthRegistrationPath = "/api/v1/auth/register"

// GrowthRegistrationSession snapshots the attribution cookie for the ordinary
// email registration request. The feature is opt-in and invalid values are
// omitted so downstream payloads can use a JSON null session.
func GrowthRegistrationSession(cfg config.GrowthRegistrationConfig) gin.HandlerFunc {
	return func(c *gin.Context) {
		if c.Request == nil || !cfg.Enabled ||
			c.Request.Method != http.MethodPost || c.Request.URL.Path != growthRegistrationPath {
			c.Next()
			return
		}

		cookieName := strings.TrimSpace(cfg.CookieName)
		if cookieName != "" {
			if cookie, err := c.Request.Cookie(cookieName); err == nil {
				ctx := service.WithGrowthRegistrationSession(c.Request.Context(), cookie.Value)
				c.Request = c.Request.WithContext(ctx)
			}
		}
		c.Next()
	}
}
