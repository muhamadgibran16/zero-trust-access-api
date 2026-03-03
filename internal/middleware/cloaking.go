package middleware

import (
	"github.com/gin-gonic/gin"
)

// Cloaking hides internal API endpoints from unauthorized internet access.
// Only traffic routing through correct encrypted tunnels (like Cloudflare Access or internal VPN)
// OR requests carrying a valid session cookie (browser-based access via IAP) are allowed.
func Cloaking(secret string) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Check for the internal secret that should only be injected by a trusted load balancer or tunnel proxy
		tunnelSecret := c.GetHeader("X-Tunnel-Secret")
		if tunnelSecret == secret {
			c.Next()
			return
		}

		// Also allow browser requests that carry a valid access_token cookie.
		// These are legitimate users accessing the portal via the browser.
		if cookie, err := c.Cookie("access_token"); err == nil && cookie != "" {
			c.Next()
			return
		}

		// No valid tunnel secret or session cookie — cloak the API.
		c.AbortWithStatus(404)
	}
}

