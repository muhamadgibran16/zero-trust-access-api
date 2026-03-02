package monitoring

import (
	"runtime"
	"time"

	"github.com/gibran/go-gin-boilerplate/database"
	"github.com/gibran/go-gin-boilerplate/pkg/response"
	"github.com/gin-gonic/gin"
)

// Handler handles system monitoring and health check endpoints
type Handler struct {
	startTime time.Time
}

// NewHandler creates a new Monitoring Handler
func NewHandler() *Handler {
	return &Handler{
		startTime: time.Now(),
	}
}

// GetHealth handles GET /admin/monitoring/health
func (h *Handler) GetHealth(c *gin.Context) {
	var mem runtime.MemStats
	runtime.ReadMemStats(&mem)

	uptimeSeconds := int(time.Since(h.startTime).Seconds())

	data := gin.H{
		"uptimeSeconds": uptimeSeconds,
		"goroutines":    runtime.NumGoroutine(),
		"memoryAllocMB": mem.Alloc / 1024 / 1024,
		"memorySysMB":   mem.Sys / 1024 / 1024,
		"cpuCores":      runtime.NumCPU(),
	}

	response.Success(c, "System health retrieved", data)
}

// GetDBStatus handles GET /admin/monitoring/db
func (h *Handler) GetDBStatus(c *gin.Context) {
	sqlDB, err := database.DB.DB()
	if err != nil {
		response.InternalServerError(c, "Failed to get DB instance")
		return
	}

	stats := sqlDB.Stats()

	redisStatus := "disconnected"
	if database.RedisClient != nil {
		if err := database.RedisClient.Ping(c.Request.Context()).Err(); err == nil {
			redisStatus = "connected"
		}
	}

	data := gin.H{
		"postgres": gin.H{
			"status":          "connected",
			"openConnections": stats.OpenConnections,
			"inUse":           stats.InUse,
			"idle":            stats.Idle,
		},
		"redis": gin.H{
			"status": redisStatus,
		},
	}

	response.Success(c, "Database status retrieved", data)
}
