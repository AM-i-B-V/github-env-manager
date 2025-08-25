package server

import (
	"context"
	"fmt"
	"net/http"
	"os/exec"
	"runtime"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"

	"github-env-manager/internal/config"
	"github-env-manager/internal/handlers"
	"github-env-manager/internal/middleware"
)

// Server represents the HTTP server
type Server struct {
	config *config.Config
	logger *logrus.Logger
	router *gin.Engine
	server *http.Server
}

// New creates a new server instance
func New(cfg *config.Config, logger *logrus.Logger) *Server {
	// Set Gin mode
	if cfg.Server.Host != "localhost" {
		gin.SetMode(gin.ReleaseMode)
	}

	router := gin.New()

	// Create server instance
	srv := &Server{
		config: cfg,
		logger: logger,
		router: router,
	}

	// Setup middleware
	srv.setupMiddleware()

	// Setup routes
	srv.setupRoutes()

	// Create HTTP server
	srv.server = &http.Server{
		Addr:    fmt.Sprintf("%s:%d", cfg.Server.Host, cfg.Server.Port),
		Handler: router,
	}

	return srv
}

// setupMiddleware configures middleware for the server
func (s *Server) setupMiddleware() {
	// Recovery middleware
	s.router.Use(gin.Recovery())

	// Logging middleware
	s.router.Use(middleware.Logger(s.logger))

	// CORS middleware
	s.router.Use(middleware.CORS())

	// Request ID middleware
	s.router.Use(middleware.RequestID())
}

// setupRoutes configures all routes for the server
func (s *Server) setupRoutes() {
	// Serve static files
	s.router.Static("/static", "./static")
	s.router.LoadHTMLGlob("templates/*")

	// Health check endpoint
	s.router.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "ok", "timestamp": time.Now().Unix()})
	})

	// API routes
	api := s.router.Group("/api")
	{
		// Auth routes
		auth := api.Group("/auth")
		{
			auth.GET("/url", handlers.GetAuthURL)
			auth.POST("/pat", handlers.AuthenticateWithPAT)
			auth.POST("/validate", handlers.ValidateToken)
			auth.GET("/callback", handlers.HandleAuthCallback)
			auth.GET("/status", handlers.GetAuthStatus)
		}

		// Repository routes
		repos := api.Group("/repos")
		{
			repos.GET("", handlers.GetRepositories)
			repos.GET("/:owner/:repo/environments", handlers.GetEnvironments)
			repos.POST("/:owner/:repo/environments", handlers.CreateEnvironment)
			repos.GET("/:owner/:repo/variables", handlers.GetVariables)
			repos.GET("/:owner/:repo/secrets", handlers.GetSecrets)
		}

		// Environment-specific routes
		envs := api.Group("/repos/:owner/:repo/environments/:env")
		{
			envs.GET("/variables", handlers.GetEnvironmentVariables)
			envs.GET("/secrets", handlers.GetEnvironmentSecrets)
			envs.POST("/variables", handlers.CreateEnvironmentVariable)
			envs.PUT("/variables/:name", handlers.UpdateEnvironmentVariable)
			envs.DELETE("/variables/:name", handlers.DeleteEnvironmentVariable)
			envs.POST("/secrets", handlers.CreateEnvironmentSecret)
			envs.PUT("/secrets/:name", handlers.UpdateEnvironmentSecret)
			envs.DELETE("/secrets/:name", handlers.DeleteEnvironmentSecret)
		}

		// Repository-level variable/secret routes
		repoVars := api.Group("/repos/:owner/:repo")
		{
			repoVars.POST("/variables", handlers.CreateVariable)
			repoVars.PUT("/variables/:name", handlers.UpdateVariable)
			repoVars.DELETE("/variables/:name", handlers.DeleteVariable)
			repoVars.POST("/secrets", handlers.CreateSecret)
			repoVars.PUT("/secrets/:name", handlers.UpdateSecret)
			repoVars.DELETE("/secrets/:name", handlers.DeleteSecret)
		}

		// Tool routes
		tools := api.Group("")
		{
			tools.POST("/sync", handlers.SyncVariables)
			tools.POST("/export", handlers.ExportVariables)
			tools.POST("/import", handlers.ImportVariables)
			tools.GET("/compare", handlers.CompareEnvironments)
		}
	}

	// WebSocket endpoint
	s.router.GET("/ws", handlers.HandleWebSocket)

	// Serve the main page
	s.router.GET("/", func(c *gin.Context) {
		c.HTML(http.StatusOK, "index.html", gin.H{
			"title": "GitHub Environment Manager",
		})
	})
}

// Start starts the server
func (s *Server) Start() error {
	// Open browser in development mode
	if s.config.Server.Host == "localhost" {
		go func() {
			time.Sleep(1 * time.Second)
			s.openBrowser()
		}()
	}

	return s.server.ListenAndServe()
}

// Shutdown gracefully shuts down the server
func (s *Server) Shutdown(ctx context.Context) error {
	return s.server.Shutdown(ctx)
}

// openBrowser opens the default browser to the application URL
func (s *Server) openBrowser() {
	url := fmt.Sprintf("http://%s:%d", s.config.Server.Host, s.config.Server.Port)
	var err error

	switch runtime.GOOS {
	case "linux":
		err = exec.Command("xdg-open", url).Start()
	case "windows":
		err = exec.Command("rundll32", "url.dll,FileProtocolHandler", url).Start()
	case "darwin":
		err = exec.Command("open", url).Start()
	default:
		err = fmt.Errorf("unsupported platform")
	}

	if err != nil {
		s.logger.Warnf("Failed to open browser: %v", err)
	}
}
