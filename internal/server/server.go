package server

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gibran/go-gin-boilerplate/config"
	"github.com/gibran/go-gin-boilerplate/database"
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
	auditRepo "github.com/gibran/go-gin-boilerplate/internal/repository/audit"
	deviceRepo "github.com/gibran/go-gin-boilerplate/internal/repository/device"
	notifRepo "github.com/gibran/go-gin-boilerplate/internal/repository/notification"
	policyRepo "github.com/gibran/go-gin-boilerplate/internal/repository/policy"
	userRepo "github.com/gibran/go-gin-boilerplate/internal/repository/user"
	"github.com/gibran/go-gin-boilerplate/internal/routes"
	adminSvc "github.com/gibran/go-gin-boilerplate/internal/service/audit"
	authSvc "github.com/gibran/go-gin-boilerplate/internal/service/auth"
	deviceSvc "github.com/gibran/go-gin-boilerplate/internal/service/device"
	notifSvc "github.com/gibran/go-gin-boilerplate/internal/service/notification"
	policySvc "github.com/gibran/go-gin-boilerplate/internal/service/policy"
	userSvc "github.com/gibran/go-gin-boilerplate/internal/service/user"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

// Server holds the HTTP server and its dependencies
type Server struct {
	config *config.Config
	logger *zap.Logger
	db     *gorm.DB
	engine *gin.Engine
}

// New creates a new Server instance
func New(cfg *config.Config, logger *zap.Logger, db *gorm.DB) *Server {
	if cfg.AppEnv == "production" {
		gin.SetMode(gin.ReleaseMode)
	}

	engine := gin.New()

	// Global middleware
	engine.Use(middleware.RequestID())
	engine.Use(middleware.RateLimiter())
	engine.Use(middleware.Recovery(logger))
	engine.Use(middleware.Logger(logger))
	engine.Use(middleware.CORS())
	engine.Use(middleware.Security())

	// Initialize Layers
	uRepo := userRepo.NewUserRepository(db)
	aRepo := auditRepo.NewAuditLogRepository(db)
	pRepo := policyRepo.NewPolicyRepository(db)
	dRepo := deviceRepo.NewDeviceRepository(db)
	nRepo := notifRepo.NewNotificationRepository(db)
	
	aService := authSvc.NewAuthService(uRepo, cfg)
	uService := userSvc.NewUserService(uRepo)
	adminService := adminSvc.NewAuditLogService(aRepo)
	pService := policySvc.NewPolicyService(pRepo)
	dService := deviceSvc.NewDeviceService(dRepo)
	nService := notifSvc.NewNotificationService(nRepo, uRepo)

	handlers := &routes.Handlers{
		Health:       healthHandler.NewHandler(),
		Auth:         authHandler.NewHandler(aService),
		User:         userHandler.NewHandler(uService),
		Admin:        adminHandler.NewHandler(adminService),
		Policy:       policyHandler.NewHandler(pService),
		Device:       deviceHandler.NewHandler(dService),
		Profile:      profileHandler.NewHandler(uRepo),
		Notification: notifHandler.NewHandler(nService),
		Analytics:    analyticsHandler.NewHandler(database.DB),
		Monitoring:   monitoringHandler.NewHandler(),
		Proxy:        proxyHandler.NewHandler(),
	}

	// Setup Routes
	routes.Setup(engine, handlers, cfg.JWTSecret, logger)

	return &Server{
		config: cfg,
		logger: logger,
		db:     db,
		engine: engine,
	}
}

// Run starts the HTTP server with graceful shutdown
func (s *Server) Run() {
	srv := &http.Server{
		Addr:         fmt.Sprintf(":%s", s.config.AppPort),
		Handler:      s.engine,
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 10 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	go func() {
		s.logger.Info("Starting server",
			zap.String("name", s.config.AppName),
			zap.String("env", s.config.AppEnv),
			zap.String("port", s.config.AppPort),
		)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			s.logger.Fatal("Failed to start server", zap.Error(err))
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	s.logger.Info("Shutting down server...")

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		s.logger.Fatal("Server forced to shutdown", zap.Error(err))
	}

	s.logger.Info("Server exited")
}
