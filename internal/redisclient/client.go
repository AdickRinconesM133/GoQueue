package redisclient

import (
	"context"
	"github.com/redis/go-redis/v9"
	"log"
)

func InitRedis() *redis.Client {
	ctx := context.Background()

	rdb := redis.NewClient(&redis.Options{
		Addr: "localhost:6379",
	})

	if err := rdb.Ping(ctx).Err(); err != nil {
		log.Fatalf("Could not connect to Redis: %v", err)
	}

	log.Printf("Redis connection: Successfull")
	return rdb
}
