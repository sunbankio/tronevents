package processor

import (
	"context"
	"log"
	"time"

	"github.com/sunbankio/tronevents/pkg/scanner"
	"github.com/sunbankio/tronevents/pkg/publisher"
)

// NormalOperation handles the processing of blocks when the daemon is in sync.
func NormalOperation(s scanner.IScanner, p *publisher.EventPublisher, logger *log.Logger) {
	for {
		// Get the latest block number
		currentBlock, err := s.GetCurrentBlockNumber()
		if err != nil {
			logger.Printf("Error getting current block number: %v", err)
			time.Sleep(1 * time.Second)
			continue
		}

		// Get transactions for the current block
		transactions, err := s.GetTransactionsByBlock(currentBlock)
		if err != nil {
			logger.Printf("Error getting transactions for block %d: %v", currentBlock, err)
			time.Sleep(1 * time.Second)
			continue
		}

		// Publish each transaction to the Redis stream
		for _, tx := range transactions {
			if err := p.Publish(context.Background(), &tx); err != nil {
				logger.Printf("Error publishing transaction: %v", err)
			}
		}

		// Wait 1 second before checking for the next block
		time.Sleep(1 * time.Second)
	}
}
