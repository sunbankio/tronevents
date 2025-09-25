package processor

import (
	"encoding/json"
	"log"
	"time"

	"github.com/hibiken/asynq"
	"github.com/sunbankio/tronevents/pkg/storage"
	"github.com/sunbankio/tronevents/pkg/scanner"
)

// SlightBacklog handles the processing of a slight block backlog.
func SlightBacklog(s scanner.IScanner, q *asynq.Client, bs *storage.BlockStorage, logger *log.Logger) {
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

	// Process blocks in the slight backlog (up to 20 blocks)
	for blockNum := lastSyncedBlock + 1; blockNum <= currentBlock && blockNum <= lastSyncedBlock + 20; blockNum++ {
		// Prepare payload for the task
		payload, err := json.Marshal(map[string]interface{}{"block_number": blockNum})
		if err != nil {
			logger.Printf("Error marshaling payload for block %d: %v", blockNum, err)
			continue
		}

		// Create a task to process this block
		task := asynq.NewTask("block:process", payload)

		// Add the task to the priority queue
		if _, err := q.Enqueue(task, asynq.Queue("priority")); err != nil {
			logger.Printf("Error enqueuing block %d: %v", blockNum, err)
		}
	}

	// Update the last synced block number to current block
	if err := bs.Save(currentBlock); err != nil {
		logger.Printf("Error saving last synced block: %v", err)
	}

	// Wait 3 seconds before continuing
	time.Sleep(3 * time.Second)
}
