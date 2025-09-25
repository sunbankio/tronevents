package worker

import (
	"context"
	"encoding/json"
	"fmt"
	"log"

	"github.com/hibiken/asynq"
	"github.com/sunbankio/tronevents/pkg/publisher"
	"github.com/sunbankio/tronevents/pkg/scanner"
	"github.com/sunbankio/tronevents/pkg/storage"
)

// Handler processes Asynq tasks
type Handler struct {
	tronScanner       *scanner.Scanner
	publisher         *publisher.EventPublisher
	logger            *log.Logger
	blockProcessedStorage *storage.BlockProcessedStorage
}

// NewHandler creates a new task handler
func NewHandler(tronScanner *scanner.Scanner, publisher *publisher.EventPublisher, blockProcessedStorage *storage.BlockProcessedStorage, logger *log.Logger) *Handler {
	return &Handler{
		tronScanner:       tronScanner,
		publisher:         publisher,
		logger:            logger,
		blockProcessedStorage: blockProcessedStorage,
	}
}

// HandleTask processes a task from the queue
func (h *Handler) HandleTask(ctx context.Context, t *asynq.Task) error {
	h.logger.Printf("DEBUG: Worker processing task from queue")

	// Extract block number from the payload
	var p map[string]interface{}
	if err := json.Unmarshal(t.Payload(), &p); err != nil {
		h.logger.Printf("ERROR: Failed to unmarshal task payload: %v", err)
		return err
	}

	blockNum, ok := p["block_number"].(float64) // JSON numbers are float64
	if !ok {
		err := fmt.Errorf("block_number not found in task payload: %v", p)
		h.logger.Printf("ERROR: %v", err)
		return err
	}

	blockNumber := int64(blockNum)
	h.logger.Printf("DEBUG: Worker processing block %d", blockNumber)

	// Check if block has already been processed (idempotent processing)
	alreadyProcessed, err := h.blockProcessedStorage.IsProcessed(ctx, blockNumber)
	if err != nil {
		h.logger.Printf("ERROR: Failed to check if block %d was already processed: %v", blockNumber, err)
		return err
	}

	if alreadyProcessed {
		h.logger.Printf("DEBUG: Block %d already processed, skipping", blockNumber)
		return nil
	}

	// Get transactions for this specific block
	transactions, err := h.tronScanner.GetTransactionsByBlock(blockNumber)
	if err != nil {
		h.logger.Printf("ERROR: Failed to get transactions for block %d: %v", blockNumber, err)
		return err
	}

	h.logger.Printf("DEBUG: Retrieved %d transactions for block %d", len(transactions), blockNumber)

	// Publish these transactions to the Redis stream in batch
	// Convert slice of values to slice of pointers for batch publishing
	transactionPointers := make([]*scanner.Transaction, len(transactions))
	for i := range transactions {
		transactionPointers[i] = &transactions[i]
	}

	if err := h.publisher.PublishBatch(context.Background(), transactionPointers); err != nil {
		h.logger.Printf("ERROR: Failed to publish batch of %d transactions for block %d: %v", len(transactions), blockNumber, err)
		return err
	}
	publishedCount := len(transactions)
	errorCount := 0

	// Mark the block as processed to prevent duplicate processing
	if err := h.blockProcessedStorage.MarkProcessed(ctx, blockNumber); err != nil {
		h.logger.Printf("ERROR: Failed to mark block %d as processed: %v", blockNumber, err)
		return err
	}

	h.logger.Printf("DEBUG: Successfully processed block %d, published %d transactions, %d errors", blockNumber, publishedCount, errorCount)

	return nil
}

// RegisterHandlers registers task handlers with the Asynq mux
func RegisterHandlers(mux *asynq.ServeMux, handler *Handler) {
	mux.HandleFunc("block:process", handler.HandleTask)
}