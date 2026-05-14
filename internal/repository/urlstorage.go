package repository

import (
	"context"
	"time"

	"github.com/redis/go-redis/v9"
)

const (
	urlExpTime = 24 * time.Hour // for testing, set a short expiration time.
)

// ref from instructor: interface originally defined here, moved to service layer (Clean Architecture)
//type UrlStorage interface {
//	StoreURL(ctx context.Context, code, url string) error
//	GetURL(ctx context.Context, code string) (string, error)
//}

type urlStorage struct {
	c *redis.Client
}

func NewUrlStorage(c *redis.Client) *urlStorage {
	return &urlStorage{c: c}
}

// StoreURL stores a URL with its shortened code in Redis with the given expiration duration.
func (s *urlStorage) StoreURL(ctx context.Context, code, url string, exp time.Duration) error {
	return s.c.Set(ctx, code, url, exp).Err()
}

// GetURL retrieves the original URL by its shortened code from Redis
func (s *urlStorage) GetURL(ctx context.Context, code string) (string, error) {
	return s.c.Get(ctx, code).Result()
}

// StoreURLIfNotExists stores the URL only if the code does not already exist.
// Returns true if stored successfully, false if the code already exists.
func (s *urlStorage) StoreURLIfNotExists(ctx context.Context, code, url string, exp time.Duration) (bool, error) {
	return s.c.SetNX(ctx, code, url, exp).Result()
}
