package api

import (
	"log/slog"

	"github.com/gin-gonic/gin"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/tempest-concorde/fw-app/internal/api/handlers"
	"github.com/tempest-concorde/fw-app/internal/api/middleware"
	"github.com/tempest-concorde/fw-app/internal/audit"
	"github.com/tempest-concorde/fw-app/internal/auth"
	"github.com/tempest-concorde/fw-app/internal/storage"
)

// RouterConfig contains all dependencies for router setup
type RouterConfig struct {
	DB             *storage.DB
	JWTManager     *auth.JWTManager
	GitHubAuth     *auth.GitHubAuth
	AuditWriter    *audit.Writer
	SwaggerEnabled bool
}

// NewRouter creates and configures the Gin router
func NewRouter(cfg RouterConfig) *gin.Engine {
	logger := slog.Default()
	
	gin.SetMode(gin.ReleaseMode)
	r := gin.New()

	// Global middleware
	r.Use(middleware.Logging(logger))
	r.Use(middleware.Metrics())
	r.Use(gin.Recovery())

	// Static files
	r.Static("/static", "./web/static")
	r.StaticFile("/", "./web/static/index.html")
	r.StaticFile("/app", "./web/static/app.html")

	// Health endpoints (no auth)
	healthHandler := handlers.NewHealthHandler(cfg.DB.DB)
	r.GET("/health", healthHandler.Health)
	r.GET("/readyz", healthHandler.Readiness)

	// Metrics (no auth)
	r.GET("/metrics", gin.WrapH(promhttp.Handler()))

	// Auth endpoints (no auth)
	authHandler := handlers.NewAuthHandler(cfg.GitHubAuth, cfg.JWTManager)
	authGroup := r.Group("/auth")
	{
		authGroup.GET("/login", authHandler.Login)
		authGroup.GET("/callback", authHandler.Callback)
		authGroup.POST("/logout", authHandler.Logout)
	}

	// API endpoints (auth required)
	sampleHandler := handlers.NewSampleHandler(cfg.DB.DB)
	apiGroup := r.Group("/api/v1")
	apiGroup.Use(middleware.Auth(cfg.JWTManager))
	apiGroup.Use(middleware.Audit(cfg.AuditWriter))
	{
		apiGroup.GET("/samples", sampleHandler.ListSamples)
		apiGroup.POST("/samples", sampleHandler.CreateSample)
		apiGroup.GET("/samples/:id", sampleHandler.GetSample)
		apiGroup.PUT("/samples/:id", sampleHandler.UpdateSample)
		apiGroup.DELETE("/samples/:id", sampleHandler.DeleteSample)
	}

	return r
}
