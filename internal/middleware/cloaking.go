package middleware

import (
	"github.com/gin-gonic/gin"
)

// Cloaking hides internal API endpoints from unauthorized internet access.
// Only traffic routing through correct encrypted tunnels (like Cloudflare Access or internal VPN)
// and carrying a specific secret key is allowed.
func Cloaking(secret string) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Check for the internal secret that should only be injected by a trusted load balancer or tunnel proxy
		tunnelSecret := c.GetHeader("X-Tunnel-Secret")
		if tunnelSecret != secret {
			// Do not explicitly state it's an API. Cloak by dropping connection or returning a generic resource.
			// Returning 404 Not Found to simulate cloaking (ignoring the exact router path match).
			c.AbortWithStatus(404)
			return
		}
		c.Next()
	}
}
