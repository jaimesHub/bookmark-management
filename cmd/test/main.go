package main

import (
	"context"
	"time"

	"github.com/jaimesHub/bookmark-management/internal/repository"
	"github.com/jaimesHub/bookmark-management/internal/service"
	"github.com/jaimesHub/bookmark-management/pkg/redis"
)

const (
	expirationTime = 24 * time.Hour
)

func main() {
	// example 1
	ctx := context.Background()

	//redisClient, err := redis.NewClient("")

	urlStorage, err := redis.NewClient("")
	if err != nil {
		panic(err)
	}

	//redisClient.Set(ctx, "key", "1234", time.Hour)
	//
	//// example 2
	//cacheDB, err := redis.NewClient("CACHE")
	//if err != nil {
	//	panic(err)
	//}
	//
	//// export CACHE_REDIS_DB=2 - Redis có 0-15 sub-db
	//// docker run --name redis -d -p 6379:6379 redis:alpine
	//// docker exec -it redis redis-cli
	//// keys *
	//// switch db: seleLct index , index = [0, 15] -> select 2
	//
	//cacheDB.Set(ctx, "key", "4567", time.Hour)

	urlRepo := repository.NewUrlStorage(urlStorage)
	// _ = urlRepo.StoreURL(ctx, "112233", "youtube.com")

	urlService := service.NewShortenService(urlRepo)

	key, _ := urlService.ShortenURL(ctx, "https://instagram.com", expirationTime)

	println(">>> shortened key:", key)

}
