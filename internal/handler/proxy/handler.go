package proxy

import (
	"fmt"
	"log"
	"net/http"
	"net/http/httputil"
	"net/url"
	"strings"

	"github.com/gibran/go-gin-boilerplate/database"
	"github.com/gibran/go-gin-boilerplate/internal/model"
	"github.com/gibran/go-gin-boilerplate/pkg/response"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type Handler struct{}

func NewHandler() *Handler {
	return &Handler{}
}

// GetPortalApps handles GET /users/portal/apps
func (h *Handler) GetPortalApps(c *gin.Context) {
	var routes []model.AppRoute
	if err := database.DB.Where("is_active = ?", true).Find(&routes).Error; err != nil {
		response.InternalServerError(c, "Failed to fetch apps")
		return
	}

	response.Success(c, "Portal apps retrieved", routes)
}

// ReverseProxy handles ANY /proxy/*target_path
func (h *Handler) ReverseProxy(c *gin.Context) {
	// The full path will be e.g. /api/v1/users/proxy/hr-app/api/users
	// We need to extract the app identifier (hr-app)
	proxyPath := c.Param("target_path") // e.g. /hr-app/api/users

	if proxyPath == "" || proxyPath == "/" {
		response.BadRequest(c, "App prefix is required")
		return
	}

	// Extract the prefix, ex: from "/hr-app/api/users" -> "hr-app"
	parts := strings.SplitN(strings.TrimPrefix(proxyPath, "/"), "/", 2)
	appPrefix := parts[0]
	
	// AppPrefix in DB is saved as "/hr-app"
	searchPrefix := "/" + appPrefix

	var appRoute model.AppRoute
	if err := database.DB.Where("path_prefix = ? AND is_active = ?", searchPrefix, true).First(&appRoute).Error; err != nil {
		log.Printf("[Proxy] App route not found or inactive: %s", searchPrefix)
		response.NotFound(c, "Application not found or disabled in portal")
		return
	}

	targetUrl, err := url.Parse(appRoute.TargetURL)
	if err != nil {
		response.InternalServerError(c, "Failed to parse target URL configuration")
		return
	}

	// Setup the reverse proxy
	proxy := httputil.NewSingleHostReverseProxy(targetUrl)

	// Modify the request before it goes to the internal app
	proxy.Director = func(req *http.Request) {
		req.Header.Set("X-Forwarded-Host", req.Header.Get("Host"))
		req.Header.Set("X-ZTA-Authenticated", "true")

		// Inject shared secret so target app can verify requests come from FortiGateX
		if appRoute.ProxySecret != "" {
			req.Header.Set("X-Proxy-Secret", appRoute.ProxySecret)
		}

		// Forward UserID to internal app for identity context
		if userID, exists := c.Get("userID"); exists {
			req.Header.Set("X-ZTA-User-ID", fmt.Sprintf("%v", userID))
		}

		req.Host = targetUrl.Host
		req.URL.Scheme = targetUrl.Scheme
		req.URL.Host = targetUrl.Host

		// Rewrite path: strip the ZTA prefix
		// e.g. /api/v1/users/proxy/hr-app/api/users -> /api/users
		if len(parts) > 1 {
			req.URL.Path = "/" + parts[1]
		} else {
			req.URL.Path = "/"
		}
	}

	// Serve HTTP
	proxy.ServeHTTP(c.Writer, c.Request)
}

// ---- Admin CRUD Endpoints for AppRoutes ----

// GetAllRoutes handles GET /admin/proxy-routes
func (h *Handler) GetAllRoutes(c *gin.Context) {
	var routes []model.AppRoute
	if err := database.DB.Order("created_at DESC").Find(&routes).Error; err != nil {
		response.InternalServerError(c, "Failed to fetch routes")
		return
	}
	response.Success(c, "Routes retrieved successfully", routes)
}

// CreateRoute handles POST /admin/proxy-routes
func (h *Handler) CreateRoute(c *gin.Context) {
	var req struct {
		Name        string `json:"name" binding:"required"`
		Description string `json:"description"`
		PathPrefix  string `json:"pathPrefix" binding:"required"`
		TargetURL   string `json:"targetUrl" binding:"required"`
		Icon        string `json:"icon"`
		IsActive    bool   `json:"isActive"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, err.Error())
		return
	}
	
	if !strings.HasPrefix(req.PathPrefix, "/") {
		req.PathPrefix = "/" + req.PathPrefix
	}

	appRoute := model.AppRoute{
		Name:        req.Name,
		Description: req.Description,
		PathPrefix:  req.PathPrefix,
		TargetURL:   req.TargetURL,
		ProxySecret: strings.ReplaceAll(uuid.New().String(), "-", ""),
		Icon:        req.Icon,
		IsActive:    req.IsActive,
	}

	if err := database.DB.Create(&appRoute).Error; err != nil {
		if strings.Contains(err.Error(), "duplicate key value") || strings.Contains(err.Error(), "unique constraint") {
			response.BadRequest(c, "Path prefix already exists")
			return
		}
		response.InternalServerError(c, "Failed to create route")
		return
	}

	response.Created(c, "Route created successfully", appRoute)
}

// UpdateRoute handles PUT /admin/proxy-routes/:id
func (h *Handler) UpdateRoute(c *gin.Context) {
	id := c.Param("id")

	var req struct {
		Name        string `json:"name" binding:"required"`
		Description string `json:"description"`
		PathPrefix  string `json:"pathPrefix" binding:"required"`
		TargetURL   string `json:"targetUrl" binding:"required"`
		Icon        string `json:"icon"`
		IsActive    *bool  `json:"isActive" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	if !strings.HasPrefix(req.PathPrefix, "/") {
		req.PathPrefix = "/" + req.PathPrefix
	}

	var route model.AppRoute
	if err := database.DB.First(&route, "id = ?", id).Error; err != nil {
		response.NotFound(c, "Route not found")
		return
	}

	route.Name = req.Name
	route.Description = req.Description
	route.PathPrefix = req.PathPrefix
	route.TargetURL = req.TargetURL
	route.Icon = req.Icon
	if req.IsActive != nil {
		route.IsActive = *req.IsActive
	}

	if err := database.DB.Save(&route).Error; err != nil {
		if strings.Contains(err.Error(), "duplicate key value") || strings.Contains(err.Error(), "unique constraint") {
			response.BadRequest(c, "Path prefix already exists")
			return
		}
		response.InternalServerError(c, "Failed to update route")
		return
	}

	response.Success(c, "Route updated successfully", route)
}

// DeleteRoute handles DELETE /admin/proxy-routes/:id
func (h *Handler) DeleteRoute(c *gin.Context) {
	id := c.Param("id")

	if err := database.DB.Delete(&model.AppRoute{}, "id = ?", id).Error; err != nil {
		response.InternalServerError(c, "Failed to delete route")
		return
	}

	response.Success(c, "Route deleted successfully", nil)
}
