package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"strings"
	"time"

	"github.com/go-redis/redis/v8"
)

func main() {
	redisStream := NewRedisStream("localhost:6379", "")

	processor := &EventProcessor{
		redisStream:  redisStream,
		streamName:   "tron:events",
		groupName:    "exampleGroup1",
		consumerName: "exampleConsumer1",
		stateStore:   &FileStateStore{filename: "last_processed_id.txt"},
	}

	// Start processing
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	if err := processor.ProcessEvents(ctx); err != nil {
		log.Fatal(err)
	}
}

type EventProcessor struct {
	redisStream  *RedisStream
	streamName   string
	groupName    string
	consumerName string
	stateStore   *FileStateStore // File-based state store
}

// File-based state store that matches the README description
type FileStateStore struct {
	filename string
}

func (s *FileStateStore) SaveLastProcessedID(consumerGroup, consumerName, lastID string) error {
	content := fmt.Sprintf("%s:%s:%s", consumerGroup, consumerName, lastID)
	return os.WriteFile(s.filename, []byte(content), 0644)
}

func (s *FileStateStore) GetLastProcessedID(consumerGroup, consumerName string) (string, error) {
	content, err := os.ReadFile(s.filename)
	if err != nil {
		return "", err
	}
	
	parts := strings.Split(string(content), ":")
	if len(parts) != 3 {
		return "", fmt.Errorf("invalid checkpoint format")
	}
	
	storedGroup := parts[0]
	storedConsumer := parts[1]
	lastID := parts[2]
	
	if storedGroup != consumerGroup || storedConsumer != consumerName {
		return "", fmt.Errorf("checkpoint does not match current consumer group")
	}
	
	return lastID, nil
}

func (ep *EventProcessor) ProcessEvents(ctx context.Context) error {
	// Get last processed ID
	lastID, err := ep.stateStore.GetLastProcessedID(ep.groupName, ep.consumerName)
	if err != nil {
		lastID = "0-0" // Start from beginning if no state found
	}

	// Subscribe to stream
	streamCh, err := ep.redisStream.SubscribeWithResume(ep.streamName, ep.groupName, ep.consumerName, lastID)
	if err != nil {
		return err
	}

	for {
		select {
		case <-ctx.Done():
			return nil
		case stream, ok := <-streamCh:
			if !ok {
				return nil
			}

			for _, msg := range stream.Messages {
				// Process the message
				if err := ep.processMessage(msg); err != nil {
					log.Printf("Error processing message %s: %v", msg.ID, err)
					continue
				}

				// Acknowledge the message
				ep.redisStream.client.XAck(ep.redisStream.ctx, ep.streamName, ep.groupName, msg.ID)

				// Save the last processed ID
				ep.stateStore.SaveLastProcessedID(ep.groupName, ep.consumerName, msg.ID)
			}
		}
	}
}

func (ep *EventProcessor) processMessage(msg redis.XMessage) error {
	// Your message processing logic here
	fmt.Printf("Processing message ID: %s, Values: %v\n", msg.ID, msg.Values)
	return nil
}

type RedisStream struct {
	client *redis.Client
	ctx    context.Context
}

func NewRedisStream(addr, password string) *RedisStream {
	rdb := redis.NewClient(&redis.Options{
		Addr:     addr,
		Password: password,
		DB:       0,
	})

	return &RedisStream{
		client: rdb,
		ctx:    context.Background(),
	}
}

// Create consumer group
func (rs *RedisStream) CreateConsumerGroup(streamName, groupName string) error {
	return rs.client.XGroupCreateMkStream(rs.ctx, streamName, groupName, "$").Err()
}

// Subscribe with resume capability
func (rs *RedisStream) SubscribeWithResume(streamName, groupName, consumerName string,
	lastProcessedID string) (<-chan *redis.XStream, error) {

	// Create consumer group if it doesn't exist
	err := rs.CreateConsumerGroup(streamName, groupName)
	if err != nil && err.Error() != "BUSYGROUP Consumer Group name already exists" {
		return nil, err
	}

	ch := make(chan *redis.XStream)

	go func() {
		defer close(ch)

		// Start position
		startID := lastProcessedID
		if startID == "" {
			startID = "0-0" // Start from beginning
		}

		for {
			// Read pending messages first (for resume capability)
			if startID != ">" {
				pendingMsgs, err := rs.client.XReadGroup(rs.ctx, &redis.XReadGroupArgs{
					Group:    groupName,
					Consumer: consumerName,
					Streams:  []string{streamName, startID},
					Count:    100,
					Block:    0, // No blocking for pending messages
				}).Result()

				if err == nil && len(pendingMsgs) > 0 {
					ch <- &pendingMsgs[0]
					startID = ">" // Switch to new messages after processing pending
					continue
				}
			}

			// Read new messages
			streams, err := rs.client.XReadGroup(rs.ctx, &redis.XReadGroupArgs{
				Group:    groupName,
				Consumer: consumerName,
				Streams:  []string{streamName, ">"},
				Count:    100,
				Block:    5 * time.Second, // Block for 5 seconds
			}).Result()

			if err != nil {
				if err != redis.Nil {
					log.Printf("Error reading from stream: %v", err)
				}
				continue
			}

			if len(streams) > 0 {
				ch <- &streams[0]
			}
		}
	}()

	return ch, nil
}
