package routes

import (
	"net/http"

	adminHandler "github.com/gibran/go-gin-boilerplate/internal/handler/admin"
	authHandler "github.com/gibran/go-gin-boilerplate/internal/handler/auth"
	healthHandler "github.com/gibran/go-gin-boilerplate/internal/handler/health"
	userHandler "github.com/gibran/go-gin-boilerplate/internal/handler/user"
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
			auth.POST("/sso-login", handlers.Auth.SsoLogin)
			auth.POST("/refresh", handlers.Auth.Refresh)
		}

		// Users (Protected ZTA Zone)
		users := v1.Group("/users")
		users.Use(middleware.IdentityAwareProxy(jwtSecret)) // Replaces regular Auth middleware for strict identity verification
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
			}
			
			// We can also route the original User admin handlers here if needed, but for ZTA let's keep it simple
			// We will just leave them inside users group or admin group as originally structured.
			
			// Note: The previous User.GetMany was assigned to admin group without prefix `""`, we'll restore user standard endpoints.
			users.GET("", middleware.PolicyEngine(), middleware.ValidateQueryParams([]string{"page", "perPage"}), handlers.User.GetMany)
			users.GET("/:id", handlers.User.GetOne)
			users.PUT("/:id", handlers.User.Update)
			users.DELETE("/:id", handlers.User.Delete)
			
			// General protected routes
			users.POST("/logout", handlers.Auth.Logout)
		}
	}
}
