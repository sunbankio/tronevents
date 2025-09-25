package processor

import (
	"context"
	"log"
	"time"

	"github.com/sunbankio/tronevents/pkg/scanner"
	"github.com/sunbankio/tronevents/pkg/publisher"
	"github.com/sunbankio/tronevents/pkg/storage"
)

// SyncLogic handles the processing of blocks on the first run or when in sync.
func SyncLogic(s scanner.IScanner, p *publisher.EventPublisher, bs *storage.BlockStorage, logger *log.Logger) {
	for {
		// Load the last synced block number
		lastSyncedBlock, err := bs.Load()
		if err != nil {
			logger.Printf("Error loading last synced block: %v", err)
			time.Sleep(3 * time.Second)
			continue
		}

		// Get the next block to process
		nextBlock := lastSyncedBlock + 1

		// Get transactions for the next block
		transactions, err := s.GetTransactionsByBlock(nextBlock)
		if err != nil {
			logger.Printf("Error getting transactions for block %d: %v", nextBlock, err)
			time.Sleep(3 * time.Second)
			continue
		}

		// Publish each transaction to the Redis stream
		for _, tx := range transactions {
			if err := p.Publish(context.Background(), &tx); err != nil {
				logger.Printf("Error publishing transaction: %v", err)
			}
		}

		// Update the last synced block number
		if err := bs.Save(nextBlock); err != nil {
			logger.Printf("Error saving last synced block: %v", err)
		}

		// Wait 3 seconds before processing the next block
		time.Sleep(3 * time.Second)
	}
}
