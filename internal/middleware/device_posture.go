package middleware

import (
	"strings"

	"github.com/gibran/go-gin-boilerplate/database"
	"github.com/gibran/go-gin-boilerplate/internal/model"
	"github.com/gibran/go-gin-boilerplate/pkg/response"
	"github.com/gibran/go-gin-boilerplate/pkg/security"
	"github.com/gin-gonic/gin"
)

// DevicePosture ensures that only secure and compliant devices can access the platform
func DevicePosture(secret string) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Exempt self-service and device management endpoints from posture checks.
		path := c.Request.URL.Path
		exemptPrefixes := []string{
			"/api/v1/users/devices",
			"/api/v1/users/admin",  // admin routes are protected by PolicyEngine
			"/api/v1/users/profile",
			"/api/v1/users/notifications",
			"/api/v1/users/logout",
			"/api/v1/users/setup-mfa",
			"/api/v1/users/enable-mfa",
			"/api/v1/users/portal",
			"/api/v1/users/proxy",  // proxy routes are protected by IAP
		}
		for _, prefix := range exemptPrefixes {
			if strings.HasPrefix(path, prefix) {
				c.Next()
				return
			}
		}

		deviceTokenStr := c.GetHeader("X-Device-Token")
		if deviceTokenStr == "" {
			response.Forbidden(c, "Device Posture: Missing cryptographic device token (X-Device-Token). Please register your device.")
			c.Abort()
			return
		}

		claims, err := security.ValidateDeviceToken(deviceTokenStr, secret)
		if err != nil {
			response.Forbidden(c, "Device Posture: Invalid device token. Signature verification failed.")
			c.Abort()
			return
		}

		// Double-check DB to ensure device wasn't revoked since token was issued
		var device model.Device
		if err := database.DB.Select("is_approved").Where("id = ?", claims.DeviceID).First(&device).Error; err != nil {
			response.Forbidden(c, "Device Posture: Device not found in registry.")
			c.Abort()
			return
		}

		if !device.IsApproved {
			response.Forbidden(c, "Device Posture: Device access has been revoked or is pending approval.")
			c.Abort()
			return
		}

		c.Next()
	}
}
