package redis

import "github.com/redis/go-redis/v9"

// NewClient creates a Redis client using config loaded via envconfig.
// envPrefix scopes env var lookup (e.g. "" → REDIS_ADDR, "CACHE" → CACHE_REDIS_ADDR),
// allowing multiple Redis instances in one process.
func NewClient(envPrefix string) (*redis.Client, error) {
	cfg, err := newConfig(envPrefix)
	if err != nil {
		return nil, err
	}
	redisClient := redis.NewClient(&redis.Options{
		Addr:     cfg.Address,
		Password: cfg.Password,
		DB:       cfg.DB,
	})

	return redisClient, nil
}
