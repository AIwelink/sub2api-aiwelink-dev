package routes

import (
	"net/http"
	"net/url"
	"regexp"
	"strings"

	"github.com/Wei-Shaw/sub2api/internal/config"
	"github.com/gin-gonic/gin"
)

var growthReferralCodePattern = regexp.MustCompile("^[a-hj-km-np-z2-9]{8}$")

// RegisterGrowthReferralRoutes registers the browser-facing referral entry
// before the embedded frontend middleware can serve an SPA fallback.
func RegisterGrowthReferralRoutes(r *gin.Engine, cfg config.GrowthRegistrationConfig) {
	r.GET("/r/:code", func(c *gin.Context) {
		c.Header("Cache-Control", "no-store")
		if !cfg.Enabled {
			c.Redirect(http.StatusFound, "/register")
			return
		}

		code := strings.ToLower(c.Param("code"))
		if !growthReferralCodePattern.MatchString(code) {
			c.Redirect(http.StatusFound, "/register")
			return
		}

		base, err := url.Parse(cfg.ReferralBaseURL)
		if err != nil || base.Scheme != "https" || base.Host == "" ||
			base.User != nil || base.RawQuery != "" || base.Fragment != "" ||
			base.Path != "/r" || base.RawPath != "" {
			c.Redirect(http.StatusFound, "/register")
			return
		}
		base.Path += "/" + code
		base.RawQuery = url.Values{"entry": {"api"}}.Encode()
		c.Redirect(http.StatusFound, base.String())
	})
}
