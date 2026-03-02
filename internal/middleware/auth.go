package middleware

import (
	"strings"

	"github.com/gibran/go-gin-boilerplate/database"
	"github.com/gibran/go-gin-boilerplate/internal/model"
	"github.com/gibran/go-gin-boilerplate/pkg/response"
	"github.com/gibran/go-gin-boilerplate/pkg/security"
	"github.com/gin-gonic/gin"
)

// Auth returns a Gin middleware that validates JWT tokens
func Auth(secret string) gin.HandlerFunc {
	return func(c *gin.Context) {
		token, err := c.Cookie("access_token")
		if err != nil || token == "" {
			authHeader := c.GetHeader("Authorization")
			if authHeader != "" {
				parts := strings.SplitN(authHeader, " ", 2)
				if len(parts) == 2 && parts[0] == "Bearer" {
					token = parts[1]
				}
			}
		}

		if token == "" {
			response.Unauthorized(c, "Authorization token is required")
			c.Abort()
			return
		}

		claims, err := security.ValidateToken(token, secret)
		if err != nil {
			response.Unauthorized(c, "Invalid or expired token")
			c.Abort()
			return
		}

		// Verify session is not blacklisted
		var blacklistCount int64
		err = database.DB.Model(&model.TokenBlacklist{}).
			Where("token_id = ? OR (user_id = ? AND created_at > ?)", claims.ID, claims.UserID, claims.IssuedAt.Time).
			Count(&blacklistCount).Error

		if err == nil && blacklistCount > 0 {
			response.Unauthorized(c, "Session has been revoked by Administrator")
			c.Abort()
			return
		}

		// Inject user info into context
		c.Set("userID", claims.UserID)
		c.Set("role", claims.Role)
		c.Set("jti", claims.ID)
		c.Set("token_exp", claims.ExpiresAt.Time)

		c.Next()
	}
}
