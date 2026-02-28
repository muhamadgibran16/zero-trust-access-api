package middleware

import (
	"github.com/gibran/go-gin-boilerplate/pkg/response"
	"github.com/gin-gonic/gin"
)

// DevicePosture ensures that only secure and compliant devices can access the platform
func DevicePosture() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Expect client to send some basic device health information.
		// In production, this can be integrated with MDM like Jamf or Intune
		// using certificates or external trust tokens. Here we look at headers for demonstration.
		
		deviceOS := c.GetHeader("X-Device-OS")
		deviceSecure := c.GetHeader("X-Device-Secure")
		deviceRooted := c.GetHeader("X-Device-Rooted")

		if deviceRooted == "true" {
			response.Forbidden(c, "Device Posture: Access restricted from rooted or jailbroken devices")
			c.Abort()
			return
		}

		if deviceSecure != "true" || deviceOS == "" {
			response.Forbidden(c, "Device Posture: Device non-compliant. Ensure OS is updated and security agents are active.")
			c.Abort()
			return
		}

		c.Next()
	}
}
