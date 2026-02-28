package middleware

import (
	"fmt"
	"strings"

	"github.com/gibran/go-gin-boilerplate/database"
	"github.com/gibran/go-gin-boilerplate/internal/model"
	"github.com/gibran/go-gin-boilerplate/pkg/response"
	"github.com/gin-gonic/gin"
)

// PolicyEngine dynamically evaluates access rules combining Identity, Role, and Network context against the database.
func PolicyEngine() gin.HandlerFunc {
	return func(c *gin.Context) {
		role, exists := c.Get("role")
		if !exists {
			response.Unauthorized(c, "Policy Engine: Missing role context")
			c.Abort()
			return
		}

		path := c.Request.URL.Path
		clientIP := c.ClientIP()

		var policies []model.PolicyRule
		// We fetch active policies that either apply globally (Resource="") or match the prefix of the current path
		// Notice: In production, memory caching (e.g. Redis) is heavily recommended over querying the DB per HTTP request.
		if err := database.DB.Where("is_active = ?", true).Find(&policies).Error; err == nil {
			for _, policy := range policies {
				// Check resource matching
				if policy.Resource != "" && !strings.HasPrefix(path, policy.Resource) {
					continue // Rule doesn't apply to this path
				}

				// Evaluate rules
				switch policy.Type {
				case model.PolicyTypeDenyIP:
					if clientIP == policy.Value {
						response.Forbidden(c, fmt.Sprintf("Policy Engine: Access from blocked IP %s", clientIP))
						c.Abort()
						return
					}
				case model.PolicyTypeRequire:
					// Require a certain role
					roleStr := ""
					if role != nil {
						roleStr = role.(string)
					}
					if roleStr != policy.Value {
						response.Forbidden(c, fmt.Sprintf("Policy Engine: Requires %s privileges", policy.Value))
						c.Abort()
						return
					}
				}
			}
		}

		// Hardcoded fallback admin restriction from original phase
		if strings.HasPrefix(path, "/api/v1/users/admin") && role != model.RoleAdmin {
			response.Forbidden(c, "Policy Engine: Requires Admin privileges")
			c.Abort()
			return
		}

		c.Next()
	}
}
