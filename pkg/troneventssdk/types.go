package troneventssdk

import (
	"context"
	"github.com/go-redis/redis/v8"
)

// Event represents a single event from the TRON blockchain stream.
type Event struct {
	ID    string            // The unique ID of the event in the Redis stream
	Data  map[string]interface{} // The event data as key-value pairs
	Raw   *redis.XMessage   // The raw Redis message, in case advanced users need access to it
}

// EventHandler is the function signature that developers will implement to process events.
// It receives the context and the event data.
// If the handler returns an error, the SDK will log the error but continue processing subsequent events.
type EventHandler func(ctx context.Context, event Event) error

// StateStore defines the interface for persisting and retrieving the last processed event ID.
// This allows for pluggable state storage mechanisms (file, database, etc.).
type StateStore interface {
	// SaveLastProcessedID saves the ID of the last successfully processed event.
	// consumerGroup and consumerName are used to uniquely identify the consumer instance.
	SaveLastProcessedID(consumerGroup, consumerName, lastID string) error

	// GetLastProcessedID retrieves the ID of the last processed event.
	// It returns an empty string if no previous state exists.
	// consumerGroup and consumerName are used to uniquely identify the consumer instance.
	GetLastProcessedID(consumerGroup, consumerName string) (string, error)
}

// InitialPosition defines where the subscriber should start reading from when no valid state is found.
type InitialPosition string

const (
	// InitialPositionBeginning starts reading from the very first event in the stream.
	InitialPositionBeginning InitialPosition = "beginning"

	// InitialPositionEnd starts reading from new events published after the subscriber connects.
	// This is equivalent to using '$' in Redis XREADGROUP.
	InitialPositionEnd InitialPosition = "end"
)

// SubscriberConfig holds the configuration options for the event subscriber.
type SubscriberConfig struct {
	// RedisAddr is the address of the Redis server (e.g., "localhost:6379").
	RedisAddr string

	// RedisPassword is the password for the Redis server, if required.
	RedisPassword string

	// RedisDB is the Redis database number to use (default is 0).
	RedisDB int

	// StreamName is the name of the Redis stream to subscribe to (e.g., "tron:events").
	StreamName string

	// GroupName is the name of the Redis consumer group to join.
	// Multiple instances of your application with the same GroupName will share the load.
	GroupName string

	// ConsumerName is the unique name for this specific consumer instance within the group.
	// It should be unique across all instances of your application.
	ConsumerName string

	// StateStore is the mechanism used to persist the last processed event ID.
	// If nil, a default file-based state store will be used.
	StateStore StateStore

	// StateStorePath is the path to the file used by the default file-based state store.
	// It's only used if StateStore is nil.
	// Default is "last_processed_id.txt".
	StateStorePath string

	// BatchSize is the maximum number of messages to read from the stream in a single call.
	// Default is 100.
	BatchSize int

	// ReadTimeout is the duration to block when reading from the stream.
	// Default is 5 seconds.
	ReadTimeout int

	// Logger is the logger implementation to use for SDK logs.
	// If nil, a default logger will be used.
	Logger Logger

	// InitialPosition determines where to start reading from when no valid state is found.
	// This can be either "beginning" to process all historical events, or "end" to only process new events.
	// Default is "beginning".
	InitialPosition InitialPosition
}