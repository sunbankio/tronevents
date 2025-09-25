package storage

import (
	"context"
	"time"

	"github.com/go-redis/redis/v8"
)

// BlockProcessedStorage handles tracking of processed blocks to prevent duplicates.
type BlockProcessedStorage struct {
	client *redis.Client
	key    string
}

// NewBlockProcessedStorage creates a new BlockProcessedStorage.
func NewBlockProcessedStorage(client *redis.Client, key string) *BlockProcessedStorage {
	return &BlockProcessedStorage{
		client: client,
		key:    key,
	}
}

// IsProcessed checks if a block has already been processed.
func (s *BlockProcessedStorage) IsProcessed(ctx context.Context, blockNumber int64) (bool, error) {
	exists, err := s.client.SIsMember(ctx, s.key, blockNumber).Result()
	if err != nil && err != redis.Nil {
		return false, err
	}
	return exists, nil
}

// MarkProcessed marks a block as processed with a 7-day expiration.
func (s *BlockProcessedStorage) MarkProcessed(ctx context.Context, blockNumber int64) error {
	pipe := s.client.TxPipeline()
	pipe.SAdd(ctx, s.key, blockNumber)
	pipe.Expire(ctx, s.key, 7*24*time.Hour) // 7 days expiration
	_, err := pipe.Exec(ctx)
	return err
}