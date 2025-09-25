package monitoring

import (
	"sync"
)

// Metrics holds the monitoring metrics for the daemon.
type Metrics struct {
	mu             sync.Mutex
	EventsProcessed uint64
	BlocksScanned   uint64
	QueueStatus    map[string]uint64
}

// NewMetrics creates a new Metrics.
func NewMetrics() *Metrics {
	return &Metrics{
		QueueStatus: make(map[string]uint64),
	}
}

// IncrEventsProcessed increments the number of events processed.
func (m *Metrics) IncrEventsProcessed() {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.EventsProcessed++
}

// IncrBlocksScanned increments the number of blocks scanned.
func (m *Metrics) IncrBlocksScanned() {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.BlocksScanned++
}

// SetQueueStatus sets the status of a queue.
func (m *Metrics) SetQueueStatus(queue string, value uint64) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.QueueStatus[queue] = value
}
