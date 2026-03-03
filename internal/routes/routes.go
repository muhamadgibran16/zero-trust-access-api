package routes

import (
	"net/http"

	adminHandler "github.com/gibran/go-gin-boilerplate/internal/handler/admin"
	analyticsHandler "github.com/gibran/go-gin-boilerplate/internal/handler/analytics"
	authHandler "github.com/gibran/go-gin-boilerplate/internal/handler/auth"
	deviceHandler "github.com/gibran/go-gin-boilerplate/internal/handler/device"
	healthHandler "github.com/gibran/go-gin-boilerplate/internal/handler/health"
	monitoringHandler "github.com/gibran/go-gin-boilerplate/internal/handler/monitoring"
	notifHandler "github.com/gibran/go-gin-boilerplate/internal/handler/notification"
	policyHandler "github.com/gibran/go-gin-boilerplate/internal/handler/policy"
	profileHandler "github.com/gibran/go-gin-boilerplate/internal/handler/profile"
	proxyHandler "github.com/gibran/go-gin-boilerplate/internal/handler/proxy"
	userHandler "github.com/gibran/go-gin-boilerplate/internal/handler/user"
	"github.com/gibran/go-gin-boilerplate/internal/middleware"
	"github.com/gin-gonic/gin"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
	"go.uber.org/zap"

	_ "github.com/gibran/go-gin-boilerplate/docs"
)

type Handlers struct {
	Health       *healthHandler.Handler
	Auth         *authHandler.Handler
	User         *userHandler.Handler
	Admin        *adminHandler.Handler
	Policy       *policyHandler.Handler
	Device       *deviceHandler.Handler
	Profile      *profileHandler.Handler
	Notification *notifHandler.Handler
	Analytics    *analyticsHandler.Handler
	Monitoring   *monitoringHandler.Handler
	Proxy        *proxyHandler.Handler
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
			auth.POST("/forgot-password", handlers.Auth.ForgotPassword)
			auth.POST("/reset-password", handlers.Auth.ResetPassword)
		}

		// Users (Protected ZTA Zone)
		users := v1.Group("/users")
		users.Use(middleware.IdentityAwareProxy(jwtSecret))
		users.Use(middleware.RiskEngine())
		users.Use(middleware.DevicePosture(jwtSecret))
		users.Use(middleware.AuditLog(logger))
		
		{
			// MFA Setup and Enable
			users.POST("/setup-mfa", handlers.Auth.SetupMFA)
			users.POST("/enable-mfa", handlers.Auth.EnableMFA)

			// User Profile (Self-Service)
			users.GET("/profile", handlers.Profile.GetProfile)
			users.PUT("/profile", handlers.Profile.UpdateProfile)
			users.PUT("/profile/password", handlers.Profile.ChangePassword)

			// Notifications
			users.GET("/notifications", handlers.Notification.GetNotifications)
			users.GET("/notifications/unread-count", handlers.Notification.GetUnreadCount)
			users.PUT("/notifications/:id/read", handlers.Notification.MarkAsRead)

			// Admin only routes
			admin := users.Group("/admin")
			admin.Use(middleware.PolicyEngine())
			{
				admin.GET("/audit-logs", handlers.Admin.GetAuditLogs)
				admin.GET("/analytics", handlers.Analytics.GetDashboardMetrics)
				
				// System Monitoring
				admin.GET("/monitoring/health", handlers.Monitoring.GetHealth)
				admin.GET("/monitoring/db", handlers.Monitoring.GetDBStatus)
				
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

				// Proxy Routes Management
				admin.GET("/proxy-routes", handlers.Proxy.GetAllRoutes)
				admin.POST("/proxy-routes", handlers.Proxy.CreateRoute)
				admin.PUT("/proxy-routes/:id", handlers.Proxy.UpdateRoute)
				admin.DELETE("/proxy-routes/:id", handlers.Proxy.DeleteRoute)

				// Device Management
				admin.GET("/devices", handlers.Device.GetAllDevices)
				admin.PUT("/devices/:mac/approve", handlers.Device.ApproveDevice)
				admin.PUT("/devices/:mac/reject", handlers.Device.RejectDevice)
			}
			
			// General protected routes
			users.POST("/logout", handlers.Auth.Logout)
			
			// User Devices
			users.POST("/devices", handlers.Device.RegisterDevice)
			users.POST("/devices/token", handlers.Device.GetDeviceToken)
			users.GET("/devices", handlers.Device.GetMyDevices)
			
			// Portal & Proxy (protected by PolicyEngine for external app access control)
		users.GET("/portal/apps", handlers.Proxy.GetPortalApps)

		proxyGroup := users.Group("/proxy")
		proxyGroup.Use(middleware.PolicyEngine())
		{
			proxyGroup.Any("/*target_path", handlers.Proxy.ReverseProxy)
		}
		}
	}
}
