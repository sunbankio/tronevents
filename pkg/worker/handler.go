package worker

import (
	"context"
	"encoding/json"
	"fmt"
	"log"

	"github.com/hibiken/asynq"
	"github.com/sunbankio/tronevents/pkg/publisher"
	"github.com/sunbankio/tronevents/pkg/scanner"
)

// Handler processes Asynq tasks
type Handler struct {
	tronScanner *scanner.Scanner
	publisher   *publisher.EventPublisher
	logger      *log.Logger
}

// NewHandler creates a new task handler
func NewHandler(tronScanner *scanner.Scanner, publisher *publisher.EventPublisher, logger *log.Logger) *Handler {
	return &Handler{
		tronScanner: tronScanner,
		publisher:   publisher,
		logger:      logger,
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

	h.logger.Printf("DEBUG: Worker processing block %d", int64(blockNum))

	// Get transactions for this specific block
	transactions, err := h.tronScanner.GetTransactionsByBlock(int64(blockNum))
	if err != nil {
		h.logger.Printf("ERROR: Failed to get transactions for block %d: %v", int64(blockNum), err)
		return err
	}

	h.logger.Printf("DEBUG: Retrieved %d transactions for block %d", len(transactions), int64(blockNum))

	// Publish these transactions to the Redis stream
	publishedCount := 0
	errorCount := 0

	for _, tx := range transactions {
		if err := h.publisher.Publish(context.Background(), &tx); err != nil {
			errorCount++
		} else {
			publishedCount++
		}
	}

	h.logger.Printf("DEBUG: Successfully processed block %d, published %d transactions, %d errors", int64(blockNum), publishedCount, errorCount)

	return nil
}

// RegisterHandlers registers task handlers with the Asynq mux
func RegisterHandlers(mux *asynq.ServeMux, handler *Handler) {
	mux.HandleFunc("block:process", handler.HandleTask)
}