package middleware

import (
	"log"

	"github.com/gibran/go-gin-boilerplate/database"
	"github.com/gibran/go-gin-boilerplate/internal/model"
	"github.com/gibran/go-gin-boilerplate/pkg/response"
	"github.com/gin-gonic/gin"
)

// RiskEngine evaluates continuous authorization by monitoring user behavior and dynamically blocking access if risk score is too high.
func RiskEngine() gin.HandlerFunc {
	return func(c *gin.Context) {
		userID, exists := c.Get("userID")
		if !exists {
			c.Next()
			return
		}

		var user model.User
		if err := database.DB.Select("id, risk_score").First(&user, "id = ?", userID).Error; err == nil {
			// Threshold for extreme risk
			if user.RiskScore >= 100 {
				log.Printf("[RiskEngine] BLOCKED UserID: %s due to Critical Risk Score: %d", userID, user.RiskScore)
				response.Forbidden(c, "Risk Engine: Access denied due to critical security risk score. Contact administrator.")
				c.Abort()
				return
			}
		}

		c.Next()
	}
}
