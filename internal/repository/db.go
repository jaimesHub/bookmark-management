package repository

import (
	"fmt"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

// DBConfig holds connection parameters; mapped from envconfig in api package.
// Test path dùng SQLite trực tiếp qua testdb.go (không dùng DBConfig).
type DBConfig struct {
	Host     string
	Port     int
	User     string
	Password string
	Name     string
	SSLMode  string
}

// NewPostgresDB opens a *gorm.DB on Postgres. Used by production main.
// Logger ở Warn level — tránh log mọi SQL ở prod, nhưng vẫn surface
// slow query / row affected khi cần debug.
func NewPostgresDB(cfg DBConfig) (*gorm.DB, error) {
	dsn := fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=%s",
		cfg.Host, cfg.Port, cfg.User, cfg.Password, cfg.Name, cfg.SSLMode)

	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Warn),
	})
	if err != nil {
		return nil, fmt.Errorf("open postgres: %w", err)
	}
	return db, nil
}
