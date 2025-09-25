package queue

import "github.com/hibiken/asynq"

// Config holds the configuration for the Asynq server.
type Config struct {
	Concurrency int
	Queues      map[string]int
}

// DefaultConfig returns the default Asynq server configuration.
func DefaultConfig() asynq.Config {
	return asynq.Config{
		Concurrency: 15,
		Queues: map[string]int{
			"priority":   6,
			"backlog":    3,
			"default":    1,
		},
	}
}
