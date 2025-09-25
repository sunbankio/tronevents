package queue

import (
	"time"

	"github.com/hibiken/asynq"
)

const (
	// QueuePriority is the name of the priority queue.
	QueuePriority = "priority"
	// QueueBacklog is the name of the backlog queue.
	QueueBacklog = "backlog"
	// QueueRetry is the name of the retry queue.
	QueueRetry = "retry"
	// QueueDead is the name of the dead-letter queue.
	QueueDead = "dead"
)

// NewServer creates a new Asynq server.
func NewServer(redisOpt asynq.RedisClientOpt, config asynq.Config) *asynq.Server {
	return asynq.NewServer(redisOpt, config)
}

// RetryDurations are the backoff times for retrying failed tasks.
var RetryDurations = []time.Duration{
	5 * time.Second,
	10 * time.Second,
	30 * time.Second,
	60 * time.Second,
	180 * time.Second,
	300 * time.Second,
	600 * time.Second,
	1800 * time.Second,
	3600 * time.Second,
}
