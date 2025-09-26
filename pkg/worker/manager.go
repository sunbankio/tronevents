package worker

import (
	"github.com/hibiken/asynq"
	"github.com/sunbankio/tronevents/pkg/logging"
)

// Manager handles the lifecycle of the workers.
type Manager struct {
	server *asynq.Server
	logger *logging.Logger
}

// NewManager creates a new worker Manager.
func NewManager(server *asynq.Server, logger *logging.Logger) *Manager {
	return &Manager{
		server: server,
		logger: logger,
	}
}

// StartWithMux starts the worker manager with a custom handler mux.
func (m *Manager) StartWithMux(mux *asynq.ServeMux) error {
	return m.server.Start(mux)
}

// Stop stops the worker manager.
func (m *Manager) Stop() {
	m.server.Stop()
}
