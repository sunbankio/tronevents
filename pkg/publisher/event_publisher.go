package publisher

import (
	"context"
	"encoding/json"
	"time"

	"github.com/go-redis/redis/v8"
	"github.com/sunbankio/tronevents/pkg/scanner"
)

const (
	streamName = "tron:events"
	sevenDays  = 7 * 24 * time.Hour
)

// EventPublisher is responsible for publishing events to a Redis stream.
type EventPublisher struct {
	client  *redis.Client
	limiter <-chan time.Time
}

// NewEventPublisher creates a new EventPublisher.
func NewEventPublisher(client *redis.Client) *EventPublisher {
	return &EventPublisher{
		client:  client,
		limiter: time.Tick(3 * time.Second / 500),
	}
}

// Publish publishes a transaction to the Redis stream.
func (p *EventPublisher) Publish(ctx context.Context, tx *scanner.Transaction) error {
	<-p.limiter

	payload, err := json.Marshal(tx)
	if err != nil {
		return err
	}

	return p.client.XAdd(ctx, &redis.XAddArgs{
		Stream:       streamName,
		MaxLenApprox: 201600, // 7 days * 24 hours * 60 mins * 60 secs / 3 secs per block
		Values:       map[string]interface{}{"payload": payload},
	}).Err()
}
