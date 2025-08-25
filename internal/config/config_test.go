package config

import (
	"os"
	"testing"
)

func TestConfig_Validate(t *testing.T) {
	tests := []struct {
		name    string
		config  *Config
		wantErr bool
	}{
		{
			name: "valid config",
			config: &Config{
				Server: ServerConfig{
					Host: "localhost",
					Port: 8080,
				},
			},
			wantErr: false,
		},
		{
			name: "invalid port - too low",
			config: &Config{
				Server: ServerConfig{
					Host: "localhost",
					Port: 0,
				},
			},
			wantErr: true,
		},
		{
			name: "invalid port - too high",
			config: &Config{
				Server: ServerConfig{
					Host: "localhost",
					Port: 70000,
				},
			},
			wantErr: true,
		},
		{
			name: "invalid host - empty",
			config: &Config{
				Server: ServerConfig{
					Host: "",
					Port: 8080,
				},
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.config.Validate()
			if (err != nil) != tt.wantErr {
				t.Errorf("Config.Validate() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestConfig_LoadFromEnv(t *testing.T) {
	// Test environment variable loading
	os.Setenv("HOST", "0.0.0.0")
	os.Setenv("PORT", "3000")

	defer func() {
		os.Unsetenv("HOST")
		os.Unsetenv("PORT")
	}()

	config := &Config{}
	config.LoadFromEnv()

	if config.Server.Host != "0.0.0.0" {
		t.Errorf("Expected host to be '0.0.0.0', got '%s'", config.Server.Host)
	}

	if config.Server.Port != 3000 {
		t.Errorf("Expected port to be 3000, got %d", config.Server.Port)
	}
}
