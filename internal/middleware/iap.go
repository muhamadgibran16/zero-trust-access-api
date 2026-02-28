package middleware

import (
	"log"
	"strings"

	"github.com/gibran/go-gin-boilerplate/pkg/response"
	"github.com/gibran/go-gin-boilerplate/pkg/security"
	"github.com/gin-gonic/gin"
)

// IdentityAwareProxy represents the IAP layer. It verifies identity before the request reaches the actual endpoints.
func IdentityAwareProxy(secret string) gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			// Check if specifically passed via proxy header for IAP
			authHeader = c.GetHeader("X-IAP-Token")
		}

		if authHeader == "" {
			log.Printf("[IAP] Blocked unauthenticated attempt from IP: %s", c.ClientIP())
			response.Unauthorized(c, "IAP: Identity verification required")
			c.Abort()
			return
		}

		parts := strings.SplitN(authHeader, " ", 2)
		tokenString := ""
		if len(parts) == 2 && parts[0] == "Bearer" {
			tokenString = parts[1]
		} else {
			tokenString = authHeader // fallback if just token is sent
		}

		claims, err := security.ValidateToken(tokenString, secret)
		if err != nil {
			log.Printf("[IAP] Invalid identity token from IP: %s", c.ClientIP())
			response.Unauthorized(c, "IAP: Invalid or expired identity token")
			c.Abort()
			return
		}

		// Inject user info into context for downstream
		c.Set("userID", claims.UserID)
		c.Set("role", claims.Role)

		c.Next()
	}
}
