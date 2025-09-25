package worker

import (
	"github.com/hibiken/asynq"
)

// Config holds the configuration for the worker manager.
type Config struct {
	MaxWorkers        int `yaml:"max_workers"`
	NewBlockWorkers   int `yaml:"new_block_workers"`
	PriorityWorkers   int `yaml:"priority_workers"`
	BacklogWorkers    int `yaml:"backlog_workers"`
}

// DefaultConfig returns the default asynq configuration for the worker.
func DefaultConfig() asynq.Config {
	return asynq.Config{
		Concurrency: 15,
		Queues: map[string]int{
			"priority": 1,  // 1 worker for new blocks
			"backlog":  13, // up to 13 workers for backlog
			"default":  1,  // 1 worker for priority queue
		},
	}
}
