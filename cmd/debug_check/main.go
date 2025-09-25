package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"strconv"
	"strings"

	"github.com/go-redis/redis/v8"
	"github.com/sunbankio/tronevents/pkg/scanner"
)

func main() {
	if len(os.Args) < 2 {
		fmt.Println("Usage: go run debug_check.go <block_number> [redis_addr]")
		fmt.Println("Example: go run debug_check.go 1234567")
		os.Exit(1)
	}

	blockNumber, err := strconv.ParseInt(os.Args[1], 10, 64)
	if err != nil {
		log.Fatal("Invalid block number: ", err)
	}

	redisAddr := "localhost:6379"
	if len(os.Args) > 2 {
		redisAddr = os.Args[2]
	}

	// Create Redis client
	rdb := redis.NewClient(&redis.Options{
		Addr: redisAddr,
	})

	// Check Redis stream for transactions from the specific block
	ctx := context.Background()

	// Check if stream exists by trying to get its length
	streamLen, err := rdb.XLen(ctx, "tron:events").Result()
	if err != nil || streamLen == 0 {
		fmt.Printf("Stream 'tron:events' doesn't exist or is empty (length: %d, error: %v)\n", streamLen, err)
		os.Exit(0) // Exit normally since this is not an error with the tool itself
	}

	fmt.Printf("Stream Info:\n")
	fmt.Printf("  Length: %d\n", streamLen)

	// Try to find entries related to the specific block
	// Scan the stream looking for transactions from the specified block
	streamEntries, err := rdb.XRange(ctx, "tron:events", "-", "+").Result()
	if err != nil {
		log.Printf("Error reading stream: %v", err)
		os.Exit(1)
	}

	fmt.Printf("\nTotal entries in stream: %d\n", len(streamEntries))

	foundBlock := false
	for _, entry := range streamEntries {
		for fieldName, value := range entry.Values {
			if fieldName == "payload" {
				// Try to unmarshal the payload to check if it contains the block
				var tx scanner.Transaction
				if err := json.Unmarshal([]byte(value.(string)), &tx); err == nil {
					if tx.BlockNumber == blockNumber {
						fmt.Printf("\nFound transaction from block %d:\n", blockNumber)
						fmt.Printf("  Entry ID: %s\n", entry.ID)
						fmt.Printf("  Transaction ID: %s\n", tx.ID)
						fmt.Printf("  Block Number: %d\n", tx.BlockNumber)
						fmt.Printf("  Block Timestamp: %v\n", tx.BlockTimestamp)
						fmt.Printf("  Transaction Data: %+v\n", tx)
						foundBlock = true
					}
				}
			}
		}
	}

	if !foundBlock {
		fmt.Printf("\nNo transactions found for block %d in the Redis stream.\n", blockNumber)
	}

	// Check if the block exists in any of the queues
	fmt.Printf("\nChecking queues for block %d:\n", blockNumber)
	checkQueuesForBlock(rdb, ctx, blockNumber)
}

func checkQueuesForBlock(rdb *redis.Client, ctx context.Context, blockNumber int64) {
	// Asynq uses specific queue names prefixed with asynq
	queues := []string{"asynq:queue:priority", "asynq:queue:backlog", "asynq:queue:default"}

	for _, queue := range queues {
		fmt.Printf("  Checking queue %s: ", queue)

		// Get the queue length
		queueLen, err := rdb.LLen(ctx, queue).Result()
		if err != nil {
			fmt.Printf("Error getting queue length: %v\n", err)
			continue
		}

		if queueLen > 0 {
			fmt.Printf("Length: %d\n", queueLen)

			// Get a sample of tasks to check if our block number is there
			// We'll use LRANGE to get some tasks from the queue
			// For Asynq, the tasks are stored as JSON with specific format
			// We'll check first up to 10 tasks in the queue
			endIdx := int64(9) // 0 to 9 = 10 items
			if queueLen < 10 {
				endIdx = queueLen - 1
			}

			tasks, err := rdb.LRange(ctx, queue, 0, endIdx).Result()
			if err != nil {
				fmt.Printf("    Error getting tasks from queue: %v\n", err)
				continue
			}

			blockFound := false
			for i, taskData := range tasks {
				// Check if this task contains our block number
				if strings.Contains(taskData, fmt.Sprintf("\"block_number\":%d", blockNumber)) {
					fmt.Printf("    FOUND BLOCK %d in task %d of queue %s!\n", blockNumber, i, queue)
					blockFound = true
					break
				}
			}

			if !blockFound {
				fmt.Printf("    Block %d not found in the first %d tasks of %s\n", blockNumber, len(tasks), queue)
			}
		} else {
			fmt.Printf("Empty\n")
		}
	}

	// Also check Asynq's other key patterns
	fmt.Printf("  Checking additional Asynq structures:\n")

	// Check scheduled tasks (stored in sorted sets)
	scheduledKeys, err := rdb.Keys(ctx, "asynq:scheduled:*").Result()
	if err == nil && len(scheduledKeys) > 0 {
		for _, key := range scheduledKeys {
			scheduledLen, _ := rdb.ZCard(ctx, key).Result()
			if scheduledLen > 0 {
				fmt.Printf("    Scheduled key %s: %d tasks\n", key, scheduledLen)
				// Check a few scheduled tasks for our block number
				members, err := rdb.ZRange(ctx, key, 0, 4).Result() // Check first 5
				if err == nil {
					for _, member := range members {
						if strings.Contains(member, fmt.Sprintf("\"block_number\":%d", blockNumber)) {
							fmt.Printf("    FOUND BLOCK %d in scheduled tasks!\n", blockNumber)
							break
						}
					}
				}
			}
		}
	}

	// Check retry tasks (also stored in sorted sets)
	retryKeys, err := rdb.Keys(ctx, "asynq:retry:*").Result()
	if err == nil && len(retryKeys) > 0 {
		for _, key := range retryKeys {
			retryLen, _ := rdb.ZCard(ctx, key).Result()
			if retryLen > 0 {
				fmt.Printf("    Retry key %s: %d tasks\n", key, retryLen)
				// Check a few retry tasks for our block number
				members, err := rdb.ZRange(ctx, key, 0, 4).Result() // Check first 5
				if err == nil {
					for _, member := range members {
						if strings.Contains(member, fmt.Sprintf("\"block_number\":%d", blockNumber)) {
							fmt.Printf("    FOUND BLOCK %d in retry tasks!\n", blockNumber)
							break
						}
					}
				}
			}
		}
	}

	// Check dead letter tasks
	deadKeys, err := rdb.Keys(ctx, "asynq:dead:*").Result()
	if err == nil && len(deadKeys) > 0 {
		for _, key := range deadKeys {
			deadLen, _ := rdb.ZCard(ctx, key).Result()
			if deadLen > 0 {
				fmt.Printf("    Dead key %s: %d tasks\n", key, deadLen)
				// Check a few dead tasks for our block number
				members, err := rdb.ZRange(ctx, key, 0, 4).Result() // Check first 5
				if err == nil {
					for _, member := range members {
						if strings.Contains(member, fmt.Sprintf("\"block_number\":%d", blockNumber)) {
							fmt.Printf("    FOUND BLOCK %d in dead tasks!\n", blockNumber)
							break
						}
					}
				}
			}
		}
	}
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
