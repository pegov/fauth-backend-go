package storage

import (
	"context"
	"time"

	"github.com/redis/go-redis/v9"

	"github.com/pegov/fauth-backend-go/internal/logger"
)

func GetCache(logger logger.Logger, url string) *redis.Client {
	logger.Infof("Parsing CACHE config...")
	opts, err := redis.ParseURL(url)
	if err != nil {
		panic(err)
	}

	logger.Infof("Creating CACHE client...")
	client := redis.NewClient(opts)

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	logger.Infof("Pinging CACHE...")
	if err := client.Ping(ctx).Err(); err != nil {
		panic(err)
	}

	logger.Infof("CACHE is online!")

	return client
}
