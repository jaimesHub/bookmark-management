package repository

import (
	"context"

	"github.com/redis/go-redis/v9"
)

type pingRepo struct {
	c *redis.Client
}

// NewPingRepo creates a repository responsible for checking Redis connection liveness.
// Kept separate from urlStorage so each repo has a single responsibility,
// and future DB connections can be added here without touching URL logic.
func NewPingRepo(c *redis.Client) *pingRepo {
	return &pingRepo{c: c}
}

// Ping checks if the Redis connection is alive.
func (r *pingRepo) Ping(ctx context.Context) error {
	return r.c.Ping(ctx).Err()
}
