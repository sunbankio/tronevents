package worker

import (
	"context"
	"encoding/json"
	"fmt"
	"log"

	"github.com/hibiken/asynq"
)

// Manager handles the lifecycle of the workers.
type Manager struct {
	server *asynq.Server
	logger *log.Logger
}

// NewManager creates a new worker Manager.
func NewManager(server *asynq.Server, logger *log.Logger) *Manager {
	return &Manager{
		server: server,
		logger: logger,
	}
}

// Start starts the worker manager with the default handler.
func (m *Manager) Start() error {
	h := asynq.NewServeMux()

	// For now, using a simple inline handler for debugging
	h.HandleFunc("block:process", func(ctx context.Context, t *asynq.Task) error {
		// Get the queue information if available
		m.logger.Printf("DEBUG: Worker processing task from queue (task type: %s)", t.Type())

		var p map[string]interface{}
		if err := json.Unmarshal(t.Payload(), &p); err != nil {
			m.logger.Printf("ERROR: Failed to unmarshal task payload: %v", err)
			return err
		}

		blockNum, ok := p["block_number"].(float64) // JSON numbers are float64
		if !ok {
			err := fmt.Errorf("block_number not found in task payload: %v", p)
			m.logger.Printf("ERROR: %v", err)
			return err
		}

		m.logger.Printf("DEBUG: Worker processing block %d", int64(blockNum))

		return nil
	})

	return m.server.Start(h)
}

// StartWithMux starts the worker manager with a custom handler mux.
func (m *Manager) StartWithMux(mux *asynq.ServeMux) error {
	return m.server.Start(mux)
}

// Stop stops the worker manager.
func (m *Manager) Stop() {
	m.server.Stop()
}
