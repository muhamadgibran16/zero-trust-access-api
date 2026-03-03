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
// Supports per-app policies: each AppRoute can have its own set of rules.
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

		// Determine the target AppRoute for proxy requests
		var targetAppRouteID *string
		if strings.HasPrefix(path, "/api/v1/users/proxy/") {
			// Extract app prefix: /api/v1/users/proxy/news-api/... -> news-api
			proxyPath := strings.TrimPrefix(path, "/api/v1/users/proxy/")
			parts := strings.SplitN(proxyPath, "/", 2)
			if len(parts) > 0 && parts[0] != "" {
				appPrefix := "/" + parts[0]
				var appRoute model.AppRoute
				if err := database.DB.Select("id").Where("path_prefix = ? AND is_active = ?", appPrefix, true).First(&appRoute).Error; err == nil {
					idStr := appRoute.ID.String()
					targetAppRouteID = &idStr
				}
			}
		}

		var policies []model.PolicyRule
		if err := database.DB.Where("is_active = ?", true).Find(&policies).Error; err == nil {
			for _, policy := range policies {
				// Filter policies by scope:
				// - If request is to a proxy route, only apply policies that are linked to this specific app OR are global
				// - If request is to an admin route, only apply policies that have no AppRouteID (global)
				if targetAppRouteID != nil {
					// Proxy request: apply per-app policies (matching AppRouteID) + global policies (no AppRouteID, no Resource or matching Resource)
					if policy.AppRouteID != nil {
						if policy.AppRouteID.String() != *targetAppRouteID {
							continue // This policy is for a different app
						}
					} else {
						// Global policy: check Resource prefix match if set
						if policy.Resource != "" && !strings.HasPrefix(path, policy.Resource) {
							continue
						}
					}
				} else {
					// Admin/other request: only apply global policies (no AppRouteID)
					if policy.AppRouteID != nil {
						continue // Skip per-app policies for non-proxy routes
					}
					// Check resource matching
					if policy.Resource != "" && !strings.HasPrefix(path, policy.Resource) {
						continue
					}
				}

				// Evaluate rules
				if blocked := evaluatePolicy(c, policy, role, clientIP); blocked {
					return
				}
			}
		}

		// Hardcoded fallback admin restriction (only for admin endpoints, not proxy)
		if strings.HasPrefix(path, "/api/v1/users/admin") && role != model.RoleAdmin {
			response.Forbidden(c, "Policy Engine: Requires Admin privileges")
			c.Abort()
			return
		}

		c.Next()
	}
}

// evaluatePolicy checks a single policy rule. Returns true if blocked.
func evaluatePolicy(c *gin.Context, policy model.PolicyRule, role interface{}, clientIP string) bool {
	switch policy.Type {
	case model.PolicyTypeDenyIP:
		if clientIP == policy.Value {
			response.Forbidden(c, fmt.Sprintf("Policy Engine: Access from blocked IP %s", clientIP))
			c.Abort()
			return true
		}
	case model.PolicyTypeRequire:
		roleStr := ""
		if role != nil {
			roleStr = role.(string)
		}
		if roleStr != policy.Value {
			response.Forbidden(c, fmt.Sprintf("Policy Engine: Requires %s privileges", policy.Value))
			c.Abort()
			return true
		}
	case model.PolicyTypeTime:
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

				if currentMin < startMin || currentMin > endMin {
					response.Forbidden(c, fmt.Sprintf("Policy Engine: Access restricted during this time (Allowed: %s)", policy.Value))
					c.Abort()
					return true
				}
			}
		}
	case model.PolicyTypeGeo:
		userCountry := c.GetHeader("X-Mock-Country")
		if userCountry == "" {
			userCountry = "UNKNOWN"
		}
		if strings.ToUpper(userCountry) != strings.ToUpper(policy.Value) {
			response.Forbidden(c, fmt.Sprintf("Policy Engine: Access from region %s is blocked by Geographic restrictions", userCountry))
			c.Abort()
			return true
		}
	}
	return false
}

