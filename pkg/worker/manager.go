package worker

import (
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

// Start starts the worker manager.
func (m *Manager) Start() error {
	// Create a simple handler that just logs the tasks
	h := asynq.NewServeMux()

	// You would register your task handlers here
	// h.HandleFunc("block:process", m.handleBlockProcess)

	return m.server.Start(h)
}

// Stop stops the worker manager.
func (m *Manager) Stop() {
	m.server.Stop()
}
