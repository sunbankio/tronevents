package redis

import (
	"context"
	"time"

	"github.com/go-redis/redis/v8"
)

// NewClient creates a new Redis client.
func NewClient(cfg Config) (*redis.Client, error) {
	client := redis.NewClient(&redis.Options{
		Addr:     cfg.Addr,
		Password: cfg.Password,
		DB:       cfg.DB,
	})

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := client.Ping(ctx).Err(); err != nil {
		return nil, err
	}

	return client, nil
}

// ToRedisOptions converts the config to redis.Options.
func ToRedisOptions(cfg Config) *redis.Options {
	return &redis.Options{
		Addr:     cfg.Addr,
		Password: cfg.Password,
		DB:       cfg.DB,
	}
}
