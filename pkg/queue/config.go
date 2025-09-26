package queue

import "github.com/hibiken/asynq"

// Config holds the configuration for the Asynq server.
type Config struct {
	MaxWorkers      int `yaml:"max_workers"`
	PriorityWorkers int `yaml:"priority_workers"`
	BacklogWorkers  int `yaml:"backlog_workers"`
}

// ToAsynqConfig converts queue config to asynq.Config
func (qc *Config) ToAsynqConfig() asynq.Config {
	// Use default values if config values are 0
	maxWorkers := qc.MaxWorkers
	if maxWorkers == 0 {
		maxWorkers = 15
	}

	priorityWorkers := qc.PriorityWorkers
	if priorityWorkers == 0 {
		priorityWorkers = 2
	}

	backlogWorkers := qc.BacklogWorkers
	if backlogWorkers == 0 {
		backlogWorkers = 12
	}

	// Default queue workers (for any default tasks) - keeping as 1 since we removed new_block_workers
	defaultWorkers := 1

	return asynq.Config{
		Concurrency: maxWorkers,
		Queues: map[string]int{
			"priority": priorityWorkers,  // for slight backlog
			"backlog":  backlogWorkers,  // for large backlog
			"default":  defaultWorkers, // for default tasks
		},
	}
}

// DefaultConfig returns the default Asynq server configuration.
func DefaultConfig() asynq.Config {
	return asynq.Config{
		Concurrency: 15,
		Queues: map[string]int{
			"priority": 2, // for slight backlog
			"backlog":  12, // for large backlog
			"default":  1,  // for default tasks
		},
	}
}
