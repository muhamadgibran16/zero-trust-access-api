package middleware

import (
	"log"
	"strings"

	"github.com/gibran/go-gin-boilerplate/database"
	"github.com/gibran/go-gin-boilerplate/internal/model"
	"github.com/gibran/go-gin-boilerplate/pkg/response"
	"github.com/gibran/go-gin-boilerplate/pkg/security"
	"github.com/gin-gonic/gin"
)

// IdentityAwareProxy represents the IAP layer. It verifies identity before the request reaches the actual endpoints.
func IdentityAwareProxy(secret string) gin.HandlerFunc {
	return func(c *gin.Context) {
		tokenString, err := c.Cookie("access_token")
		if err != nil || tokenString == "" {
			// Fallback to headers
			authHeader := c.GetHeader("Authorization")
			if authHeader == "" {
				authHeader = c.GetHeader("X-IAP-Token")
			}

			if authHeader != "" {
				parts := strings.SplitN(authHeader, " ", 2)
				if len(parts) == 2 && parts[0] == "Bearer" {
					tokenString = parts[1]
				} else {
					tokenString = authHeader
				}
			}
		}

		if tokenString == "" {
			log.Printf("[IAP] Blocked unauthenticated attempt from IP: %s", c.ClientIP())
			response.Unauthorized(c, "IAP: Identity verification required")
			c.Abort()
			return
		}

		claims, err := security.ValidateToken(tokenString, secret)
		if err != nil {
			log.Printf("[IAP] Invalid identity token from IP: %s - Error: %v", c.ClientIP(), err)
			response.Unauthorized(c, "IAP: Invalid or expired identity token")
			c.Abort()
			return
		}

		// Verify session is not blacklisted via DB instead of Redis
		var blacklistCount int64
		err = database.DB.Model(&model.TokenBlacklist{}).
			Where("token_id = ? OR (user_id = ? AND created_at > ?)", claims.ID, claims.UserID, claims.IssuedAt.Time).
			Count(&blacklistCount).Error

		if err == nil && blacklistCount > 0 {
			log.Printf("[IAP] Blocked revoked session for UserID: %s, IP: %s", claims.UserID, c.ClientIP())
			response.Unauthorized(c, "IAP: Session has been completely revoked by Administrator")
			c.Abort()
			return
		}

		// Inject user info into context for downstream
		c.Set("userID", claims.UserID)
		c.Set("role", claims.Role)
		c.Set("jti", claims.ID)
		c.Set("token_exp", claims.ExpiresAt.Time)

		c.Next()
	}
}
