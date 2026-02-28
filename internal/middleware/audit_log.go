package middleware

import (
	"bytes"
	"fmt"
	"io"
	"time"

	"github.com/gibran/go-gin-boilerplate/database"
	"github.com/gibran/go-gin-boilerplate/internal/model"
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

type bodyLogWriter struct {
	gin.ResponseWriter
	body *bytes.Buffer
}

func (w bodyLogWriter) Write(b []byte) (int, error) {
	w.body.Write(b)
	return w.ResponseWriter.Write(b)
}

// AuditLog intercepts requests to record specific administrative/access events
func AuditLog(logger *zap.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()

		path := c.Request.URL.Path
		method := c.Request.Method
		clientIP := c.ClientIP()
		userAgent := c.Request.UserAgent()
		
		blw := &bodyLogWriter{body: bytes.NewBufferString(""), ResponseWriter: c.Writer}
		c.Writer = blw

		// Read request body non-destructively for logging (only if needed/safe - simplified here)
		var reqBodyBytes []byte
		if c.Request.Body != nil {
			reqBodyBytes, _ = io.ReadAll(c.Request.Body)
			c.Request.Body = io.NopCloser(bytes.NewBuffer(reqBodyBytes))
		}

		c.Next()

		statusCode := c.Writer.Status()

		userIDStr := ""
		if uid, exists := c.Get("userID"); exists {
			if str, ok := uid.(string); ok {
				userIDStr = str
			} else {
				userIDStr = fmt.Sprintf("%v", uid)
			}
		}

		details := ""
		if len(c.Errors) > 0 {
			details = c.Errors.String()
		}

		// Save the log to DB asynchronously to not block response
		go func(uid, m, p, ip, ua, d string, status int) {
			audit := &model.AuditLog{
				UserID:    uid,
				Action:    "API_ACCESS",
				Method:    m,
				Path:      p,
				IPAddress: ip,
				UserAgent: ua,
				Status:    status,
				Details:   d,
			}
			if err := database.DB.Create(audit).Error; err != nil {
				logger.Error("Failed to create audit log", zap.Error(err))
			}
		}(userIDStr, method, path, clientIP, userAgent, details, statusCode)
		
		logger.Info("Audit Logged",
			zap.String("method", method),
			zap.String("path", path),
			zap.Int("status", statusCode),
			zap.Duration("latency", time.Since(start)),
		)
	}
}
