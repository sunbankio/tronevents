package storage

import (
	"context"
	"fmt"
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
	count, err := s.client.ZCount(ctx, s.key, fmt.Sprintf("%d", blockNumber), fmt.Sprintf("%d", blockNumber)).Result()
	if err != nil && err != redis.Nil {
		return false, err
	}
	return count > 0, nil
}

// MarkProcessed marks a block as processed with a 7-day expiration using ZSET.
func (s *BlockProcessedStorage) MarkProcessed(ctx context.Context, blockNumber int64) error {
	timestamp := float64(time.Now().Unix())
	return s.client.ZAdd(ctx, s.key, &redis.Z{
		Score:  timestamp,
		Member: blockNumber,
	}).Err()
}

// CleanupOldEntries removes entries older than the specified duration.
func (s *BlockProcessedStorage) CleanupOldEntries(ctx context.Context, maxAge time.Duration) error {
	cutoffTime := float64(time.Now().Add(-maxAge).Unix())
	return s.client.ZRemRangeByScore(ctx, s.key, "-inf", fmt.Sprintf("%f", cutoffTime)).Err()
}

// StartCleanup starts a background goroutine to periodically clean up old entries.
func (s *BlockProcessedStorage) StartCleanup(ctx context.Context, interval time.Duration, maxAge time.Duration) {
	go func() {
		ticker := time.NewTicker(interval)
		defer ticker.Stop()

		for {
			select {
			case <-ctx.Done():
				return
			case <-ticker.C:
				if err := s.CleanupOldEntries(ctx, maxAge); err != nil {
					// In a real application, you'd want to use a proper logger
					fmt.Printf("Error cleaning up old processed blocks: %v\n", err)
				}
			}
		}
	}()
}
