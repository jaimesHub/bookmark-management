package testutil

import (
	"testing"

	"github.com/alicebob/miniredis/v2"
	"github.com/redis/go-redis/v9"
)

// InitMockRedis starts an in-process miniredis server and returns a go-redis client pointed at it.
// The server is automatically stopped when t finishes.
func InitMockRedis(t *testing.T) *redis.Client {
	mock := miniredis.RunT(t)
	return redis.NewClient(&redis.Options{
		Addr: mock.Addr(),
	})
}

// InitClosedRedis returns a go-redis client pointed at an unreachable address.
// Use this to simulate a Redis connection failure in tests.
func InitClosedRedis(t *testing.T) *redis.Client {
	t.Helper()
	return redis.NewClient(&redis.Options{
		Addr: "127.0.0.1:1",
	})
}
