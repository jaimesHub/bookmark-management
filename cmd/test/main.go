package main

import (
	"context"
	"time"

	"github.com/jaimesHub/bookmark-management/pkg/redis"
)

func main() {
	// example 1
	ctx := context.Background()

	redisClient, err := redis.NewClient("")

	if err != nil {
		panic(err)
	}

	redisClient.Set(ctx, "key", "1234", time.Hour)

	// example 2
	cacheDB, err := redis.NewClient("CACHE")
	if err != nil {
		panic(err)
	}

	// export CACHE_REDIS_DB=2 - Redis có 0-15 sub-db
	// docker exec -it redis redis-cli
	// keys *
	// switch db: select index , index = [0, 15] -> select 2

	cacheDB.Set(ctx, "key", "4567", time.Hour)
}
