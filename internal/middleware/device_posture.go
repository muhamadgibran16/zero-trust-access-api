package middleware

import (
	"fmt"
	"log"
	"strings"

	"github.com/gibran/go-gin-boilerplate/database"
	"github.com/gibran/go-gin-boilerplate/internal/model"
	"github.com/gibran/go-gin-boilerplate/pkg/response"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

// DevicePosture ensures that only secure and compliant devices can access the platform
func DevicePosture() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Exempt self-service and device management endpoints from posture checks.
		path := c.Request.URL.Path
		exemptPrefixes := []string{
			"/api/v1/users/devices",
			"/api/v1/users/admin/devices",
			"/api/v1/users/profile",
			"/api/v1/users/notifications",
			"/api/v1/users/logout",
			"/api/v1/users/setup-mfa",
			"/api/v1/users/enable-mfa",
		}
		for _, prefix := range exemptPrefixes {
			if strings.HasPrefix(path, prefix) {
				c.Next()
				return
			}
		}

		// Expect client to send device MAC
		deviceMAC := c.GetHeader("X-Device-MAC")
		if deviceMAC == "" {
			response.Forbidden(c, "Device Posture: Missing device hardware identifier (X-Device-MAC)")
			c.Abort()
			return
		}

		// Check if device is registered and approved in database
		var device model.Device
		if err := database.DB.Where("mac_address = ?", deviceMAC).First(&device).Error; err != nil {
			// Device not found — try to auto-register for admin users
			userIDVal, exists := c.Get("userID")
			log.Printf("[DevicePosture] Device not found. MAC: %s, userID exists in ctx: %v, userIDVal: %v (type: %T)", deviceMAC, exists, userIDVal, userIDVal)

			if exists && userIDVal != nil {
				var userID uuid.UUID
				switch v := userIDVal.(type) {
				case uuid.UUID:
					userID = v
				case string:
					parsed, parseErr := uuid.Parse(v)
					if parseErr == nil {
						userID = parsed
					}
				default:
					str := fmt.Sprintf("%v", v)
					parsed, parseErr := uuid.Parse(str)
					if parseErr == nil {
						userID = parsed
					}
				}

				log.Printf("[DevicePosture] Parsed userID: %s (zero: %v)", userID, userID == uuid.Nil)

				if userID != uuid.Nil {
					var user model.User
					if dbErr := database.DB.First(&user, "id = ?", userID).Error; dbErr == nil {
						log.Printf("[DevicePosture] Found user: %s, role: %s", user.Email, user.Role)
						if user.Role == model.RoleAdmin {
							// Auto-register AND auto-approve for admin users
							newDevice := model.Device{
								UserID:     userID,
								MacAddress: deviceMAC,
								Name:       "Auto-registered (Admin)",
								IsApproved: true,
								CertThumb:  fmt.Sprintf("auto-%s", uuid.New().String()[:8]),
							}
							if createErr := database.DB.Create(&newDevice).Error; createErr != nil {
								log.Printf("[DevicePosture] FAILED to auto-register device: %v", createErr)
							} else {
								log.Printf("[DevicePosture] ✅ Auto-registered and approved device for admin %s. MAC: %s", user.Email, deviceMAC)
								c.Next()
								return
							}
						}
					} else {
						log.Printf("[DevicePosture] Failed to find user in DB: %v", dbErr)
					}
				}
			}

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
