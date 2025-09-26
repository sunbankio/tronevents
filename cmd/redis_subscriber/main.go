package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/go-redis/redis/v8"
	troneventsredis "github.com/sunbankio/tronevents/pkg/redis"
)

const (
	streamName = "tron:events"
	checkpointFile = "last_processed_id.txt"
)

var groupName string

var resumeFlag = flag.Bool("resume", false, "Resume from last processed transaction ID stored in checkpoint file")

// SafeTransaction mirrors the structure used by the publisher
type SafeTransaction struct {
	ID             string    `json:"id"`
	Contract       Contract  `json:"contract"`
	Timestamp      SafeTime  `json:"timestamp"`
	BlockNumber    int64     `json:"block_number,omitempty"`
	BlockTimestamp SafeTime  `json:"block_timestamp,omitempty"`
}

type Contract struct {
	Type         string      `json:"type"`
	Parameter    interface{} `json:"parameter"`
	PermissionID int         `json:"permission_id"`
}

type SafeTime struct {
	time.Time
}

// UnmarshalJSON implements custom unmarshaling for SafeTime
func (st *SafeTime) UnmarshalJSON(data []byte) error {
	var timestamp string
	if err := json.Unmarshal(data, &timestamp); err != nil {
		return err
	}

	parsedTime, err := time.Parse(time.RFC3339, timestamp)
	if err != nil {
		// Try Unix timestamp format
		var unixTime int64
		if err := json.Unmarshal(data, &unixTime); err != nil {
			return err
		}
		parsedTime = time.Unix(unixTime, 0)
	}

	st.Time = parsedTime
	return nil
}

// CheckpointData holds the checkpoint information including group name
type CheckpointData struct {
	Group string `json:"group"`
	ID    string `json:"id"`
}

func main() {
	flag.Parse()

	// Determine the group name based on resume flag and checkpoint
	if *resumeFlag {
		// When explicitly resuming, read the checkpoint to get the group name
		checkpoint, err := readCheckpoint()
		if err == nil && checkpoint.Group != "" {
			// Use the group name from the checkpoint
			groupName = checkpoint.Group
		} else {
			// Fallback: generate a new unique group name based on PID
			groupName = fmt.Sprintf("example-consumer-group-pid%d", os.Getpid())
		}
	} else {
		// For non-resume, generate a new unique group name
		startTime := time.Now().Unix()
		groupName = fmt.Sprintf("example-consumer-group-%d-%d", startTime, os.Getpid())
	}

	// Load Redis configuration (use default values, or from environment variables)
	cfg := troneventsredis.Config{
		Addr:     getEnv("REDIS_ADDR", "localhost:6379"),
		Password: getEnv("REDIS_PASSWORD", ""),
		DB:       0,
	}

	// Create Redis client
	client := redis.NewClient(&redis.Options{
		Addr:     cfg.Addr,
		Password: cfg.Password,
		DB:       cfg.DB,
	})

	// Test the connection
	_, err := client.Ping(context.Background()).Result()
	if err != nil {
		log.Fatalf("Failed to connect to Redis: %v", err)
	}
	defer client.Close()

	fmt.Printf("Connected to Redis. Using consumer group: %s. Subscribing to stream: %s\n", groupName, streamName)

	// Create consumer group if it doesn't exist
	err = client.XGroupCreateMkStream(context.Background(), streamName, groupName, "0").Err()
	if err != nil && err.Error() != "BUSYGROUP Consumer Group name already exists" {
		log.Printf("Warning: Failed to create consumer group: %v", err)
	}

	// Determine starting ID based on resume flag
	var startID string
	if *resumeFlag {
		if checkpoint, err := readCheckpoint(); err == nil && checkpoint.ID != "" {
			startID = checkpoint.ID
			fmt.Printf("Resuming from ID: %s\n", startID)
		} else {
			startID = ">"
			fmt.Println("Resume flag set but no checkpoint found, starting from latest messages")
		}
	} else {
		startID = ">"  // Always use ">" when not resuming to get only new messages
		fmt.Println("Starting from latest messages (use -resume to start from last processed)")
	}

	// Create context with cancellation
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// If not resuming, clear any pending messages for this consumer to start fresh
	if !*resumeFlag {
		clearPendingMessages(ctx, client)
	}

	// Handle graceful shutdown
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
	go func() {
		<-sigChan
		fmt.Println("\nReceived shutdown signal. Closing...")
		cancel()
	}()

	// Subscribe to Redis stream
	go subscribeToStream(ctx, client, startID)

	// Wait for context cancellation
	<-ctx.Done()
	fmt.Println("Subscriber stopped.")
}

func subscribeToStream(ctx context.Context, client *redis.Client, startID string) {
	fmt.Printf("Starting to subscribe to Redis stream with start ID: %s...\n", startID)

	// Initialize the IDs array with the starting ID for the stream
	streamIDs := []string{startID}

	for {
		select {
		case <-ctx.Done():
			return
		default:
			// For go-redis v8, the stream and ID are provided as alternating entries in the Streams slice
			// Format: [streamName, streamID]
			streamResults, err := client.XReadGroup(ctx, &redis.XReadGroupArgs{
				Group:    groupName,
				Consumer: "example-consumer",
				Streams:  []string{streamName, streamIDs[0]}, // Format: [streamName, streamID]
				Count:    10, // Read up to 10 messages at a time
				Block:    0,  // Block indefinitely until messages are available
			}).Result()

			if err != nil {
				if err == context.Canceled {
					fmt.Println("Context canceled, exiting subscriber...")
					return
				}
				log.Printf("Error reading from stream: %v", err)
				time.Sleep(1 * time.Second)
				continue
			}

			// Process messages
			for _, stream := range streamResults {
				for _, msg := range stream.Messages {
					// Get the payload from the message
					payload, ok := msg.Values["payload"].(string)
					if !ok {
						log.Printf("Invalid payload format in message ID: %s", msg.ID)
						continue
					}

					// Parse the transaction from JSON payload
					var safeTx SafeTransaction
					if err := json.Unmarshal([]byte(payload), &safeTx); err != nil {
						log.Printf("Error parsing transaction: %v", err)
						continue
					}

					// Print the required fields: txid, block#, transaction timestamp, contract type
					fmt.Printf("TXID: %s, Block#: %d, TransactionTimestamp: %s, ContractType: %s\n",
						safeTx.ID,
						safeTx.BlockNumber,
						safeTx.Timestamp.Format("2006-01-02 15:04:05"),
						safeTx.Contract.Type,
					)

					// Acknowledge the message so it's not delivered again
					client.XAck(ctx, streamName, groupName, msg.ID)

					// Update checkpoint file with the current group name and message ID
					if err := writeCheckpoint(groupName, msg.ID); err != nil {
						log.Printf("Error writing checkpoint: %v", err)
					}
				}
			}
		}
	}
}

// readCheckpoint reads the checkpoint data (group name and last processed ID) from the checkpoint file
func readCheckpoint() (CheckpointData, error) {
	var checkpoint CheckpointData

	data, err := os.ReadFile(checkpointFile)
	if err != nil {
		return checkpoint, err
	}

	err = json.Unmarshal(data, &checkpoint)
	if err != nil {
		// For backward compatibility, if it's not JSON, treat it as just the ID
		// and return with an empty group name
		return CheckpointData{
			Group: "",
			ID:    string(data),
		}, nil
	}

	return checkpoint, nil
}

// writeCheckpoint writes the current group name and ID to the checkpoint file
func writeCheckpoint(groupName, id string) error {
	checkpoint := CheckpointData{
		Group: groupName,
		ID:    id,
	}

	data, err := json.Marshal(checkpoint)
	if err != nil {
		return err
	}

	return os.WriteFile(checkpointFile, data, 0644)
}

// clearPendingMessages reads and acknowledges any pending messages for this consumer
// This ensures that when not resuming, we start processing only new messages
func clearPendingMessages(ctx context.Context, client *redis.Client) {
	fmt.Println("Clearing any pending messages to start fresh...")

	// Get detailed info for all pending messages in the consumer group
	pendingExtResult, err := client.XPendingExt(ctx, &redis.XPendingExtArgs{
		Stream: streamName,
		Group:  groupName,
		Start:  "-",
		End:    "+",
		Count:  1000, // Get up to 1000 pending messages
	}).Result()

	if err != nil {
		// If the group has no pending messages, this might return an error, which is fine
		if err.Error() != "NOGROUP No such key 'tron:events' or consumer group 'example-consumer-group'" {
			log.Printf("Error getting pending messages: %v", err)
		}
		return
	}

	if len(pendingExtResult) == 0 {
		fmt.Println("No pending messages to clear")
		return
	}

	// Acknowledge all pending messages to clear them
	messageIDs := make([]string, len(pendingExtResult))
	for i, msg := range pendingExtResult {
		messageIDs[i] = msg.ID
	}

	if len(messageIDs) > 0 {
		err = client.XAck(ctx, streamName, groupName, messageIDs...).Err()
		if err != nil {
			log.Printf("Error acknowledging pending messages: %v", err)
		} else {
			fmt.Printf("Cleared %d pending messages\n", len(messageIDs))
		}
	}
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}