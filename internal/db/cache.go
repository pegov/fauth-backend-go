package db

import (
	"context"
	"time"

	"github.com/redis/go-redis/v9"
)

func GetCache(dsn string) *redis.Client {
	opts, err := redis.ParseURL(dsn)
	if err != nil {
		panic(err)
	}

	client := redis.NewClient(opts)

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	if err := client.Ping(ctx).Err(); err != nil {
		panic(err)
	}

	return client
}
