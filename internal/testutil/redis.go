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
