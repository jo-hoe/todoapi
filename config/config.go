// Package config provides configuration management for the application
package config

import (
	"fmt"
	"os"
	"strconv"
	"time"
)

// Config holds all configuration for the application
type Config struct {
	Server    ServerConfig
	Todoist   TodoistConfig
	Microsoft MicrosoftConfig
}

// ServerConfig holds server-related configuration
type ServerConfig struct {
	Port         int           `json:"port"`
	ReadTimeout  time.Duration `json:"read_timeout"`
	WriteTimeout time.Duration `json:"write_timeout"`
	IdleTimeout  time.Duration `json:"idle_timeout"`
}

// TodoistConfig holds Todoist API configuration
type TodoistConfig struct {
	APIToken string `json:"api_token"`
	BaseURL  string `json:"base_url"`
}

// MicrosoftConfig holds Microsoft To Do API configuration
type MicrosoftConfig struct {
	ClientID     string `json:"client_id"`
	ClientSecret string `json:"client_secret"`
	TenantID     string `json:"tenant_id"`
	BaseURL      string `json:"base_url"`
}

// Load loads configuration from environment variables
func Load() (*Config, error) {
	config := &Config{
		Server: ServerConfig{
			Port:         getEnvAsInt("PORT", 8080),
			ReadTimeout:  getEnvAsDuration("READ_TIMEOUT", 30*time.Second),
			WriteTimeout: getEnvAsDuration("WRITE_TIMEOUT", 30*time.Second),
			IdleTimeout:  getEnvAsDuration("IDLE_TIMEOUT", 60*time.Second),
		},
		Todoist: TodoistConfig{
			APIToken: getEnv("TODOIST_API_TOKEN", ""),
			BaseURL:  getEnv("TODOIST_BASE_URL", "https://api.todoist.com/rest/v2/"),
		},
		Microsoft: MicrosoftConfig{
			ClientID:     getEnv("MS_CLIENT_ID", ""),
			ClientSecret: getEnv("MS_CLIENT_SECRET", ""),
			TenantID:     getEnv("MS_TENANT_ID", ""),
			BaseURL:      getEnv("MS_BASE_URL", "https://graph.microsoft.com/v1.0/me/todo/"),
		},
	}

	if err := config.Validate(); err != nil {
		return nil, fmt.Errorf("invalid configuration: %w", err)
	}

	return config, nil
}

// Validate validates the configuration
func (c *Config) Validate() error {
	if c.Server.Port <= 0 || c.Server.Port > 65535 {
		return fmt.Errorf("invalid server port: %d", c.Server.Port)
	}

	return nil
}

// Helper functions
func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

func getEnvAsInt(key string, defaultValue int) int {
	if value := os.Getenv(key); value != "" {
		if intValue, err := strconv.Atoi(value); err == nil {
			return intValue
		}
	}
	return defaultValue
}

func getEnvAsDuration(key string, defaultValue time.Duration) time.Duration {
	if value := os.Getenv(key); value != "" {
		if duration, err := time.ParseDuration(value); err == nil {
			return duration
		}
	}
	return defaultValue
}
