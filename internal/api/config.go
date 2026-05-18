// Package api contains API-layer configuration and utilities.
package api

import "github.com/kelseyhightower/envconfig"

// Config holds API-layer configuration loaded from environment variables.
type Config struct {
	// ContainerPort is the port the HTTP server listens on inside the container.
	// Wire env var: API_CONTAINER_PORT. docker-compose maps host port (HOST_PORT
	// in .env) to this value; the two may differ.
	ContainerPort  string `default:"8080" envconfig:"CONTAINER_PORT"`
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
