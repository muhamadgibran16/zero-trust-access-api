package middleware

import (
	"log"
	"strings"

	"github.com/gibran/go-gin-boilerplate/database"
	"github.com/gibran/go-gin-boilerplate/internal/model"
	"github.com/gibran/go-gin-boilerplate/pkg/response"
	"github.com/gin-gonic/gin"
)

// DevicePosture ensures that only secure and compliant devices can access the platform
func DevicePosture() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Exempt device registration & admin device management endpoints from posture checks
		path := c.Request.URL.Path
		if strings.HasPrefix(path, "/api/v1/users/devices") || strings.HasPrefix(path, "/api/v1/users/admin/devices") {
			c.Next()
			return
		}
		
		// Expect client to send some basic device health information.
		// In production, this can be integrated with MDM like Jamf or Intune
		// using certificates or external trust tokens. 
		
		deviceMAC := c.GetHeader("X-Device-MAC")
		if deviceMAC == "" {
			response.Forbidden(c, "Device Posture: Missing device hardware identifier (X-Device-MAC)")
			c.Abort()
			return
		}

		// Check if device is registered and approved in database
		var device model.Device
		if err := database.DB.Where("mac_address = ?", deviceMAC).First(&device).Error; err != nil {
			log.Printf("[DevicePosture] Unregistered device blocked. MAC: %s", deviceMAC)
			response.Forbidden(c, "Device Posture: Unregistered device. Please contact IT.")
			c.Abort()
			return
		}

		if !device.IsApproved {
			log.Printf("[DevicePosture] Unapproved device blocked. MAC: %s", deviceMAC)
			response.Forbidden(c, "Device Posture: Device registered but pending approval from IT.")
			c.Abort()
			return
		}

		c.Next()
	}
}
