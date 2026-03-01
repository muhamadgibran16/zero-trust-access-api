package routes

import (
	"net/http"

	adminHandler "github.com/gibran/go-gin-boilerplate/internal/handler/admin"
	authHandler "github.com/gibran/go-gin-boilerplate/internal/handler/auth"
	healthHandler "github.com/gibran/go-gin-boilerplate/internal/handler/health"
	policyHandler "github.com/gibran/go-gin-boilerplate/internal/handler/policy"
	userHandler "github.com/gibran/go-gin-boilerplate/internal/handler/user"
	deviceHandler "github.com/gibran/go-gin-boilerplate/internal/handler/device"
	"github.com/gibran/go-gin-boilerplate/internal/middleware"
	"github.com/gin-gonic/gin"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
	"go.uber.org/zap"

	_ "github.com/gibran/go-gin-boilerplate/docs"
)

type Handlers struct {
	Health *healthHandler.Handler
	Auth   *authHandler.Handler
	User   *userHandler.Handler
	Admin  *adminHandler.Handler
	Policy *policyHandler.Handler
	Device *deviceHandler.Handler
}

// Setup registers all routes to the Gin engine
func Setup(r *gin.Engine, handlers *Handlers, jwtSecret string, logger *zap.Logger) {
	// Default route
	r.GET("/", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"status":  "success",
			"message": "Hello World",
		})
	})

	// Health check route (no version prefix)
	r.GET("/health", handlers.Health.Check)

	// Swagger route
	r.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))

	// Apply Cloaking tunnel secret globally to /api/v1 (Mock tunnel secret)
	v1 := r.Group("/api/v1")
	v1.Use(middleware.Cloaking("zT-tunnel-s3cr3t"))
	{
		// Health
		v1.GET("/health", handlers.Health.Check)

		// Auth (Public)
		auth := v1.Group("/auth")
		{
			auth.POST("/register", handlers.Auth.Register)
			auth.POST("/login", handlers.Auth.Login)
			auth.POST("/verify-mfa", handlers.Auth.VerifyMFA)
			auth.GET("/google", handlers.Auth.GoogleLogin)
			auth.GET("/google/callback", handlers.Auth.GoogleCallback)
			auth.POST("/refresh", handlers.Auth.Refresh)
		}

		// Users (Protected ZTA Zone)
		users := v1.Group("/users")
		users.Use(middleware.IdentityAwareProxy(jwtSecret)) // Replaces regular Auth middleware for strict identity verification
		users.Use(middleware.RiskEngine())                    // Continuous authorization: evaluates dynamic risk thresholds	
		users.Use(middleware.DevicePosture())                 // Ensures device compliance (e.g. valid OS version)
		users.Use(middleware.AuditLog(logger))                // Audit logging for access control
		
		{
			// MFA Setup and Enable
			users.POST("/setup-mfa", handlers.Auth.SetupMFA)
			users.POST("/enable-mfa", handlers.Auth.EnableMFA)
			// Admin only routes
			admin := users.Group("/admin")
			admin.Use(middleware.PolicyEngine()) // Replaces old simple RolesAllowed with complex Policy Engine
			{
				admin.GET("/audit-logs", handlers.Admin.GetAuditLogs)
				
				// User Management
				admin.GET("/users", middleware.ValidateQueryParams([]string{"page", "perPage"}), handlers.User.GetMany)
				admin.GET("/users/:id", handlers.User.GetOne)
				admin.PUT("/users/:id/role", handlers.User.Update)
				admin.POST("/users/:id/revoke", handlers.User.Revoke)
				admin.DELETE("/users/:id", handlers.User.Delete)

				// Policy Engine Builder
				admin.GET("/policies", handlers.Policy.GetMany)
				admin.POST("/policies", handlers.Policy.Create)
				admin.PUT("/policies/:id", handlers.Policy.Update)
				admin.DELETE("/policies/:id", handlers.Policy.Delete)

				// Device Management
				admin.GET("/devices", handlers.Device.GetAllDevices)
				admin.PUT("/devices/:mac/approve", handlers.Device.ApproveDevice)
				admin.PUT("/devices/:mac/reject", handlers.Device.RejectDevice)
			}
			
			// General protected routes
			users.POST("/logout", handlers.Auth.Logout)
			
			// User Devices
			users.POST("/devices", handlers.Device.RegisterDevice)
			users.GET("/devices", handlers.Device.GetMyDevices)
		}
	}
}
