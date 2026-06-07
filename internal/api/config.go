// Package api contains API-layer configuration and utilities.
package api

import "github.com/kelseyhightower/envconfig"

// DBConfig — Postgres connection params loaded từ env.
// Main.go load riêng với prefix "" (KHÔNG qua "api" prefix của api.Config)
// để giữ docker-compose env mapping (DB_HOST=postgres) compatibility — tránh
// cascade rename .env.example + docker-compose.yml.
// Tách struct ở api package (KHÔNG dùng repository.DBConfig) vì repository
// pkg KHÔNG depend envconfig — clean architecture boundary.
// Main.go copy fields → repository.DBConfig khi call NewPostgresDB.
type DBConfig struct {
	Host     string `envconfig:"DB_HOST"     default:"localhost"`
	Port     int    `envconfig:"DB_PORT"     default:"5432"`
	User     string `envconfig:"DB_USER"     default:"bookmark"`
	Password string `envconfig:"DB_PASSWORD" required:"true"`
	Name     string `envconfig:"DB_NAME"     default:"bookmark"`
	SSLMode  string `envconfig:"DB_SSLMODE"  default:"disable"`
}

// RSAConfig — Path tới RSA keypair PEM files cho JWT signing/verifying.
// Defaults match local dev workflow (`make generate-rsa-key` produces ./keys/).
// PROD: set absolute paths /app/keys/... qua docker-compose env (ADR-10 v3).
// Main.go load riêng với prefix "" (consistency với DBConfig pattern T8+T9).
type RSAConfig struct {
	PrivatePath string `envconfig:"RSA_PRIVATE_KEY_PATH" default:"./keys/private.pem"`
	PublicPath  string `envconfig:"RSA_PUBLIC_KEY_PATH"  default:"./keys/public.pem"`
}

// Config holds API-layer configuration loaded from environment variables.
type Config struct {
	// ContainerPort is the port the HTTP server listens on inside the container.
	// Wire env var: API_CONTAINER_PORT. docker-compose maps host port (HOST_PORT
	// in .env) to this value; the two may differ.
	ContainerPort  string `default:"8080" envconfig:"CONTAINER_PORT"`
	SwaggerEnabled bool   `default:"false" envconfig:"SWAGGER_ENABLED"`
	// SwaggerHost overrides docs.SwaggerInfo.Host at runtime.
	// Empty string → Swagger UI uses same-origin (recommended khi đi qua nginx).
	// Wire env var: API_SWAGGER_HOST.
	SwaggerHost string `default:"" envconfig:"SWAGGER_HOST"`

	// Logger config
	Env      string `envconfig:"APP_ENV" default:"dev"`
	LogLevel string `envconfig:"LOG_LEVEL" default:"info"`

	// BcryptCost cho password hashing (Lec-6 register feature).
	// Prod default 12 (~100ms/hash); test inject 4 qua NewUserServiceForTest.
	// Env var: API_BCRYPT_COST (prefix "api" applied bởi NewConfig).
	BcryptCost int `envconfig:"BCRYPT_COST" default:"12"`
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
