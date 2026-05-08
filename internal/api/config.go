// Package api contains API-layer configuration and utilities.
package api

import "github.com/kelseyhightower/envconfig"

type Config struct {
	AppPort        string `default:"8080" envconfig:"APP_PORT"`
	SwaggerEnabled bool   `default:"false" envconfig:"SWAGGER_ENABLED"`
}

func NewConfig() (*Config, error) {
	cfg := &Config{}
	err := envconfig.Process("api", cfg)
	if err != nil {
		return nil, err
	}
	return cfg, nil
}
