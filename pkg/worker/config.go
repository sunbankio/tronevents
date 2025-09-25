package worker

import (
	"time"

	"github.com/hibiken/asynq"
	"github.com/sunbankio/tronevents/pkg/queue"
)

// Config holds the configuration for the worker manager.
type Config struct {
	MaxWorkers        int `yaml:"max_workers"`
	NewBlockWorkers   int `yaml:"new_block_workers"`
	PriorityWorkers   int `yaml:"priority_workers"`
	BacklogWorkers    int `yaml:"backlog_workers"`
}

// DefaultConfig returns the default asynq.Config for the worker.
func DefaultConfig() asynq.Config {
	return asynq.Config{
		Concurrency:    15,
		RetryDelayFunc: func(n int, err error, t *asynq.Task) time.Duration {
			// Use predefined retry durations from queue package
			// n is the number of times the task has been retried so far
			// When n=0, we're calculating delay before 1st retry (after 1st failure)
			// When n=1, we're calculating delay before 2nd retry (after 2nd failure)
			if n < len(queue.RetryDurations) {
				return queue.RetryDurations[n]
			}
			// If we exceed the defined durations, use the last duration
			return queue.RetryDurations[len(queue.RetryDurations)-1]
		},
		Queues: map[string]int{
			"priority": 1,  // 1 worker for new blocks
			"backlog":  13, // up to 13 workers for backlog
			"default":  1,  // 1 worker for priority queue
			"dead":     1,  // 1 worker for dead letter queue
		},
	}
}
