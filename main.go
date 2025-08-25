package main

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"os/exec"
	"runtime"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

var (
	port = 8005
	host = "localhost"
)

func main() {
	var rootCmd = &cobra.Command{
		Use:   "github-env-manager",
		Short: "A tool to manage GitHub environment variables and secrets",
		Long: `GitHub Environment Manager is a web-based tool that allows developers to:
- Manage GitHub variables and secrets across environments
- Compare variables between multiple environments
- Sync variables between environments
- Export/import environment variables as .env files
- Search and select repositories and environments`,
		Run: func(cmd *cobra.Command, args []string) {
			startServer()
		},
	}

	rootCmd.Flags().IntVarP(&port, "port", "p", 8005, "Port to run the server on")
	rootCmd.Flags().StringVarP(&host, "host", "H", "localhost", "Host to bind the server to")

	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

func startServer() {
	// Set Gin to release mode for production
	gin.SetMode(gin.ReleaseMode)

	// Create router
	router := gin.Default()

	// Serve static files
	router.Static("/static", "./static")
	router.LoadHTMLGlob("templates/*")

	// API routes
	api := router.Group("/api")
	{
		api.GET("/auth/url", getAuthURL)
		api.POST("/auth/pat", authenticateWithPAT)
		api.POST("/auth/validate", validateToken)
		api.GET("/auth/callback", handleAuthCallback)
		api.GET("/auth/status", getAuthStatus)
		api.GET("/repos", getRepositories)
		api.GET("/repos/:owner/:repo/environments", getEnvironments)
		api.POST("/repos/:owner/:repo/environments", createEnvironment)
		api.GET("/repos/:owner/:repo/variables", getVariables)
		api.GET("/repos/:owner/:repo/secrets", getSecrets)
		api.GET("/repos/:owner/:repo/environments/:env/variables", getEnvironmentVariables)
		api.GET("/repos/:owner/:repo/environments/:env/secrets", getEnvironmentSecrets)
		api.POST("/repos/:owner/:repo/environments/:env/variables", createEnvironmentVariable)
		api.PUT("/repos/:owner/:repo/environments/:env/variables/:name", updateEnvironmentVariable)
		api.DELETE("/repos/:owner/:repo/environments/:env/variables/:name", deleteEnvironmentVariable)
		api.POST("/repos/:owner/:repo/environments/:env/secrets", createEnvironmentSecret)
		api.PUT("/repos/:owner/:repo/environments/:env/secrets/:name", updateEnvironmentSecret)
		api.DELETE("/repos/:owner/:repo/environments/:env/secrets/:name", deleteEnvironmentSecret)
		api.POST("/repos/:owner/:repo/variables", createVariable)
		api.PUT("/repos/:owner/:repo/variables/:name", updateVariable)
		api.DELETE("/repos/:owner/:repo/variables/:name", deleteVariable)
		api.POST("/repos/:owner/:repo/secrets", createSecret)
		api.PUT("/repos/:owner/:repo/secrets/:name", updateSecret)
		api.DELETE("/repos/:owner/:repo/secrets/:name", deleteSecret)
		api.POST("/sync", syncVariables)
		api.POST("/export", exportVariables)
		api.POST("/import", importVariables)
		api.GET("/compare", compareEnvironments)
	}

	// WebSocket for real-time updates
	router.GET("/ws", handleWebSocket)

	// Serve the main page
	router.GET("/", func(c *gin.Context) {
		c.HTML(http.StatusOK, "index.html", gin.H{
			"title": "GitHub Environment Manager",
		})
	})

	// Start server
	serverURL := fmt.Sprintf("http://%s:%d", host, port)
	logrus.Infof("Starting GitHub Environment Manager on %s", serverURL)

	// Open browser
	go func() {
		time.Sleep(1 * time.Second)
		openBrowser(serverURL)
	}()

	if err := router.Run(fmt.Sprintf("%s:%d", host, port)); err != nil {
		log.Fatal("Failed to start server:", err)
	}
}

func openBrowser(url string) {
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
		logrus.Warnf("Failed to open browser: %v", err)
	}
}
