package storage

import (
	"context"
	"fmt"
	"strconv"

	"github.com/go-redis/redis/v8"
)

// LastSyncedStorage handles the persistence of the last processed block number using Redis.
type LastSyncedStorage struct {
	client *redis.Client
	key    string
}

// NewLastSyncedStorage creates a new LastSyncedStorage.
func NewLastSyncedStorage(client *redis.Client, key string) *LastSyncedStorage {
	return &LastSyncedStorage{
		client: client,
		key:    key,
	}
}

// Save saves the last processed block number to Redis.
func (s *LastSyncedStorage) Save(ctx context.Context, blockNumber int64) error {
	return s.client.Set(ctx, s.key, blockNumber, 0).Err()
}

// Load loads the last processed block number from Redis.
func (s *LastSyncedStorage) Load(ctx context.Context) (int64, error) {
	val, err := s.client.Get(ctx, s.key).Result()
	if err == redis.Nil {
		// Key doesn't exist, return 0 as default
		return 0, nil
	} else if err != nil {
		return 0, err
	}

	blockNumber, err := strconv.ParseInt(val, 10, 64)
	if err != nil {
		return 0, fmt.Errorf("failed to parse block number from Redis: %v", err)
	}

	return blockNumber, nil
}

