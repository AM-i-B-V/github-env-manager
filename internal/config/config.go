package config

import (
	"os"
	"strconv"
)

// Config holds all configuration for the application
type Config struct {
	Server ServerConfig `json:"server"`
}

// ServerConfig holds server-related configuration
type ServerConfig struct {
	Host string `json:"host"`
	Port int    `json:"port"`
}

// LoadFromEnv loads configuration from environment variables
func (c *Config) LoadFromEnv() {
	// Server config
	if host := os.Getenv("HOST"); host != "" {
		c.Server.Host = host
	}
	if portStr := os.Getenv("PORT"); portStr != "" {
		if port, err := strconv.Atoi(portStr); err == nil {
			c.Server.Port = port
		}
	}
}

// Validate checks if the configuration is valid
func (c *Config) Validate() error {
	if c.Server.Port <= 0 || c.Server.Port > 65535 {
		return ErrInvalidPort
	}

	if c.Server.Host == "" {
		return ErrInvalidHost
	}

	return nil
}
