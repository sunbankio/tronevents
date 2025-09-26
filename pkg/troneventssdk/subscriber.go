package troneventssdk

import (
	"context"
	"fmt"
	"time"

	"github.com/go-redis/redis/v8"
)

// Subscriber is the main struct that handles subscribing to the Redis stream and processing events.
type Subscriber struct {
	config       SubscriberConfig
	redisClient  *redis.Client
	stateStore   StateStore
	logger       Logger
	eventHandler EventHandler
}

// NewSubscriber creates a new instance of the Subscriber with the provided configuration.
// It initializes the Redis client, the state store, and the logger.
func NewSubscriber(config SubscriberConfig) *Subscriber {
	// Initialize Redis client
	redisClient := redis.NewClient(&redis.Options{
		Addr:     config.RedisAddr,
		Password: config.RedisPassword,
		DB:       config.RedisDB,
	})

	// Initialize state store
	var stateStore StateStore
	if config.StateStore != nil {
		stateStore = config.StateStore
	} else {
		// Use default file-based state store if none is provided
		storePath := config.StateStorePath
		if storePath == "" {
			storePath = "last_processed_id.txt"
		}
		stateStore = &FileStateStore{filename: storePath}
	}

	// Initialize logger
	var logger Logger
	if config.Logger != nil {
		logger = config.Logger
	} else {
		// Use default logger if none is provided
		logger = NewDefaultLogger()
	}

	return &Subscriber{
		config:      config,
		redisClient: redisClient,
		stateStore:  stateStore,
		logger:      logger,
	}
}

// Run starts the event processing loop. It blocks until the context is cancelled.
// The eventHandler function will be called for each event received from the stream.
func (s *Subscriber) Run(ctx context.Context, eventHandler EventHandler) error {
	s.eventHandler = eventHandler

	// Create consumer group if it doesn't exist.
	err := s.createConsumerGroup(s.config.StreamName, s.config.GroupName, "$")
	if err != nil {
		return fmt.Errorf("failed to create consumer group: %w", err)
	}

	// Get last processed ID from state store
	lastID, err := s.stateStore.GetLastProcessedID(s.config.GroupName, s.config.ConsumerName)
	if err != nil {
		return fmt.Errorf("failed to get last processed ID: %w", err)
	}
	if lastID == "" {
		if s.config.InitialPosition == InitialPositionEnd {
			lastID = ">"
		} else {
			lastID = "0-0"
		}
		s.logger.Infof("No previous state found, starting from: %s", lastID)
	} else {
		s.logger.Infof("Resuming from last processed ID: %s", lastID)
	}

	batchSize := s.config.BatchSize
	if batchSize <= 0 {
		batchSize = 100
	}
	readTimeout := time.Duration(s.config.ReadTimeout) * time.Second
	if readTimeout <= 0 {
		readTimeout = 5 * time.Second
	}

	// Main processing loop
	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
			// Determine read ID and block duration
			readID := lastID
			if lastID == ">" {
				// When in "live" mode, we always read new messages
				readID = ">"
			}

			// Read messages
			streams, err := s.redisClient.XReadGroup(ctx, &redis.XReadGroupArgs{
				Group:    s.config.GroupName,
				Consumer: s.config.ConsumerName,
				Streams:  []string{s.config.StreamName, readID},
				Count:    int64(batchSize),
				Block:    readTimeout,
			}).Result()

			if err != nil {
				if err != redis.Nil {
					s.logger.Errorf("Error reading from stream: %v", err)
				}
				continue
			}

			if len(streams) == 0 || len(streams[0].Messages) == 0 {
				// If we were catching up and got no messages, switch to live mode
				if lastID != ">" {
					s.logger.Infof("Finished catching up. Switching to live mode.")
					lastID = ">"
				}
				continue
			}

			// Process messages
			for _, msg := range streams[0].Messages {
				if err := s.processMessage(ctx, msg); err != nil {
					s.logger.Errorf("Error processing message %s: %v", msg.ID, err)
					continue
				}
				// Update lastID only when catching up
				if lastID != ">" {
					lastID = msg.ID
				}
			}
		}
	}
}

// processMessage handles the processing of a single Redis message.
// It calls the event handler, acknowledges the message, and saves the state.
func (s *Subscriber) processMessage(ctx context.Context, msg redis.XMessage) error {
	event := Event{
		ID:   msg.ID,
		Data: msg.Values,
		Raw:  &msg,
	}

	// Call the user-provided event handler
	if err := s.eventHandler(ctx, event); err != nil {
		return fmt.Errorf("event handler failed for message %s: %w", msg.ID, err)
	}

	// Acknowledge the message
	if err := s.redisClient.XAck(ctx, s.config.StreamName, s.config.GroupName, msg.ID).Err(); err != nil {
		return fmt.Errorf("failed to acknowledge message %s: %w", msg.ID, err)
	}

	// Save the last processed ID
	if err := s.stateStore.SaveLastProcessedID(s.config.GroupName, s.config.ConsumerName, msg.ID); err != nil {
		return fmt.Errorf("failed to save last processed ID for message %s: %w", msg.ID, err)
	}

	return nil
}

// createConsumerGroup creates the Redis consumer group if it doesn't already exist.
func (s *Subscriber) createConsumerGroup(streamName, groupName, initialPos string) error {
	err := s.redisClient.XGroupCreateMkStream(s.redisClient.Context(), streamName, groupName, initialPos).Err()
	if err != nil {
		// If the group already exists, that's fine
		if err.Error() == "BUSYGROUP Consumer Group name already exists" {
			return nil
		}
		return err
	}
	return nil
}