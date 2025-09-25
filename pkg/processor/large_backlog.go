package processor

import (
	"encoding/json"
	"log"

	"github.com/hibiken/asynq"
	"github.com/sunbankio/tronevents/pkg/storage"
	"github.com/sunbankio/tronevents/pkg/scanner"
)

// LargeBacklog handles the processing of a large block backlog.
func LargeBacklog(s scanner.IScanner, q *asynq.Client, bs *storage.BlockStorage, logger *log.Logger) {
	// Load the last synced block number
	lastSyncedBlock, err := bs.Load()
	if err != nil {
		logger.Printf("Error loading last synced block: %v", err)
		return
	}

	// Get the current block number
	currentBlock, err := s.GetCurrentBlockNumber()
	if err != nil {
		logger.Printf("Error getting current block number: %v", err)
		return
	}

	// Calculate the starting block for the backlog range
	startBlock := lastSyncedBlock + 1
	if currentBlock - 201600 > startBlock {
		startBlock = currentBlock - 201600
	}

	// Process blocks in the large backlog range
	for blockNum := startBlock; blockNum < currentBlock; blockNum++ {
		// Prepare payload for the task
		payload, err := json.Marshal(map[string]interface{}{"block_number": blockNum})
		if err != nil {
			logger.Printf("Error marshaling payload for block %d: %v", blockNum, err)
			continue
		}

		// Create a task to process this block
		task := asynq.NewTask("block:process", payload)

		// Add the task to the backlog queue
		if _, err := q.Enqueue(task, asynq.Queue("backlog")); err != nil {
			logger.Printf("Error enqueuing block %d: %v", blockNum, err)
		}
	}

	// Update the last synced block number
	if err := bs.Save(currentBlock - 1); err != nil {
		logger.Printf("Error saving last synced block: %v", err)
	}
}
