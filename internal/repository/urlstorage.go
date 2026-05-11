package repository

import (
	"context"
	"time"

	"github.com/redis/go-redis/v9"
)

const (
	urlExpTime = 24 * time.Hour
)

// TODO: DELETE THIS AFTER REFACTORING DONE
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

// StoreURL stores a URL with its shortened code in Redis with a 24-hour expiration
func (s *urlStorage) StoreURL(ctx context.Context, code, url string) error {
	return s.c.Set(ctx, code, url, urlExpTime).Err()
}

// GetURL retrieves the original URL by its shortened code from Redis
func (s *urlStorage) GetURL(ctx context.Context, code string) (string, error) {
	return s.c.Get(ctx, code).Result()
}
