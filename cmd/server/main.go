package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"

	"github-env-manager/internal/config"
	"github-env-manager/internal/server"
)

var (
	port    = 8080
	host    = "localhost"
	version = "dev" // This will be set by the build process
)

func main() {
	var rootCmd = &cobra.Command{
		Use:     "github-env-manager",
		Version: version,
		Short:   "A tool to manage GitHub environment variables and secrets",
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

	rootCmd.Flags().IntVarP(&port, "port", "p", 8080, "Port to run the server on")
	rootCmd.Flags().StringVarP(&host, "host", "H", "localhost", "Host to bind the server to")

	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

func startServer() {
	// Initialize configuration
	cfg := &config.Config{
		Server: config.ServerConfig{
			Host: host,
			Port: port,
		},
	}

	// Initialize logger
	logger := logrus.New()
	logger.SetFormatter(&logrus.JSONFormatter{})
	logger.SetLevel(logrus.InfoLevel)

	// Create server instance
	srv := server.New(cfg, logger)

	// Start server in a goroutine
	go func() {
		logger.Infof("Starting GitHub Environment Manager v%s on http://%s:%d", version, host, port)
		if err := srv.Start(); err != nil {
			logger.Fatal("Failed to start server:", err)
		}
	}()

	// Wait for interrupt signal to gracefully shutdown the server
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	logger.Info("Shutting down server...")

	// Create a deadline for server shutdown
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// Attempt graceful shutdown
	if err := srv.Shutdown(ctx); err != nil {
		logger.Fatal("Server forced to shutdown:", err)
	}

	logger.Info("Server exited")
}
