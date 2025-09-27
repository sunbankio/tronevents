package main

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/go-redis/redis/v8"
)

func main() {
	redisStream := NewRedisStream("localhost:6379", "")

	processor := &EventProcessor{
		redisStream:  redisStream,
		streamName:   "tron:events",
		groupName:    "tronevents_group",    // Hardcoded consumer group
		consumerName: "tronevents_consumer", // Hardcoded consumer name
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
}

func (ep *EventProcessor) ProcessEvents(ctx context.Context) error {
	// Subscribe to stream starting from the last acknowledged message (>)
	// This will automatically read only new messages that haven't been delivered to any consumer in the group
	streamCh, err := ep.redisStream.SubscribeWithResume(ep.streamName, ep.groupName, ep.consumerName, ">")
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

				// Acknowledge the message - Redis will track this automatically
				ep.redisStream.client.XAck(ep.redisStream.ctx, ep.streamName, ep.groupName, msg.ID)
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

// Subscribe with resume capability - now uses server-side tracking only
func (rs *RedisStream) SubscribeWithResume(streamName, groupName, consumerName, startID string) (<-chan *redis.XStream, error) {

	// Create consumer group if it doesn't exist
	err := rs.CreateConsumerGroup(streamName, groupName)
	if err != nil && err.Error() != "BUSYGROUP Consumer Group name already exists" {
		return nil, err
	}

	ch := make(chan *redis.XStream)

	go func() {
		defer close(ch)

		for {
			// Read new messages using server-side tracking
			// If startID is ">", we only get new messages never delivered to any consumer in the group
			streams, err := rs.client.XReadGroup(rs.ctx, &redis.XReadGroupArgs{
				Group:    groupName,
				Consumer: consumerName,
				Streams:  []string{streamName, startID}, // Using the startID parameter (typically ">")
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
