package storage

import (
	"context"
	"time"

	"github.com/redis/go-redis/v9"

	"github.com/pegov/fauth-backend-go/internal/log"
)

func GetCache(ctx context.Context, logger log.Logger, url string) (*redis.Client, error) {
	logger.Infof("Parsing CACHE config...")
	opts, err := redis.ParseURL(url)
	if err != nil {
		return nil, err
	}

	logger.Infof("Creating CACHE client...")
	client := redis.NewClient(opts)

	ctx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()
	logger.Infof("Pinging CACHE...")
	if err := client.Ping(ctx).Err(); err != nil {
		return nil, err
	}

	logger.Infof("CACHE is online!")

	return client, nil
}
