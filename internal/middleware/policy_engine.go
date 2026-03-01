package middleware

import (
	"fmt"
	"strings"
	"time"

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
				case model.PolicyTypeTime:
					// Parse "HH:MM-HH:MM" e.g., "08:00-17:00"
					parts := strings.Split(policy.Value, "-")
					if len(parts) == 2 {
						now := time.Now()
						currentMin := now.Hour()*60 + now.Minute()

						startParts := strings.Split(parts[0], ":")
						endParts := strings.Split(parts[1], ":")

						if len(startParts) == 2 && len(endParts) == 2 {
							var sh, sm, eh, em int
							fmt.Sscanf(startParts[0], "%d", &sh)
							fmt.Sscanf(startParts[1], "%d", &sm)
							fmt.Sscanf(endParts[0], "%d", &eh)
							fmt.Sscanf(endParts[1], "%d", &em)

							startMin := sh*60 + sm
							endMin := eh*60 + em

							// If current time is OUTSIDE the allowed window
							if currentMin < startMin || currentMin > endMin {
								response.Forbidden(c, fmt.Sprintf("Policy Engine: Access restricted during this time (Allowed: %s)", policy.Value))
								c.Abort()
								return
							}
						}
					}
				case model.PolicyTypeGeo:
					// Mock GeoIP detection via header for demonstration
					userCountry := c.GetHeader("X-Mock-Country")
					if userCountry == "" {
						userCountry = "UNKNOWN" // Default if not passed
					}
					
					// Rule value could be "ID" (Indonesia), "US", etc.
					// If the user's country does NOT match the required allowed country, block.
					if strings.ToUpper(userCountry) != strings.ToUpper(policy.Value) {
						response.Forbidden(c, fmt.Sprintf("Policy Engine: Access from region %s is blocked by Geographic restrictions", userCountry))
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
