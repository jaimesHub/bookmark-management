// Package api contains API-layer configuration and utilities.
package api

import "github.com/kelseyhightower/envconfig"

// Config holds API-layer configuration loaded from environment variables.
type Config struct {
	AppPort        string `default:"8080" envconfig:"APP_PORT"`
	SwaggerEnabled bool   `default:"false" envconfig:"SWAGGER_ENABLED"`

	// Logger config
	Env      string `envconfig:"APP_ENV" default:"dev"`
	LogLevel string `envconfig:"LOG_LEVEL" default:"info"`
}

// NewConfig creates a new API Config instance from environment variables.
func NewConfig() (*Config, error) {
	cfg := &Config{}
	err := envconfig.Process("api", cfg)
	if err != nil {
		return nil, err
	}
	return cfg, nil
}
