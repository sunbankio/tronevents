package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"strings"

	"github.com/go-redis/redis/v8"
	"github.com/hibiken/asynq"
)

func main() {
	redisAddr := "localhost:6379"
	if len(os.Args) > 1 {
		redisAddr = os.Args[1]
	}

	// Create Redis client
	rdb := redis.NewClient(&redis.Options{
		Addr: redisAddr,
	})

	ctx := context.Background()

	fmt.Println("=== Redis Asynq Queue Inspector ===")
	fmt.Printf("Connecting to Redis at: %s\n\n", redisAddr)

	// Get all keys in Redis
	keys, err := rdb.Keys(ctx, "*").Result()
	if err != nil {
		log.Printf("Error getting keys: %v", err)
		return
	}

	fmt.Printf("Found %d keys in Redis:\n", len(keys))
	for _, key := range keys {
		fmt.Printf("  - %s\n", key)
	}

	fmt.Println("\n=== Asynq Queue Analysis ===")

	// Look for Asynq-related keys
	asynqKeys := make(map[string][]string)
	for _, key := range keys {
		if strings.HasPrefix(key, "asynq:") {
			parts := strings.Split(key, ":")
			if len(parts) >= 2 {
				category := parts[1]
				asynqKeys[category] = append(asynqKeys[category], key)
			}
		}
	}

	if len(asynqKeys) > 0 {
		fmt.Println("\nAsynq-related keys found:")
		for category, keys := range asynqKeys {
			fmt.Printf("  %s:\n", category)
			for _, key := range keys {
				keyType, err := rdb.Type(ctx, key).Result()
				if err != nil {
					fmt.Printf("    - %s (type: error - %v)\n", key, err)
					continue
				}

				switch keyType {
				case "string":
					value, err := rdb.Get(ctx, key).Result()
					if err != nil {
						fmt.Printf("    - %s (type: string, value: error - %v)\n", key, err)
					} else {
						fmt.Printf("    - %s (type: string, value: %s)\n", key, value)
					}
				case "list":
					length, err := rdb.LLen(ctx, key).Result()
					if err != nil {
						fmt.Printf("    - %s (type: list, length: error - %v)\n", key, err)
					} else {
						fmt.Printf("    - %s (type: list, length: %d)\n", key, length)
					}
				case "set":
					card, err := rdb.SCard(ctx, key).Result()
					if err != nil {
						fmt.Printf("    - %s (type: set, cardinality: error - %v)\n", key, err)
					} else {
						fmt.Printf("    - %s (type: set, cardinality: %d)\n", key, card)
					}
				case "zset":
					card, err := rdb.ZCard(ctx, key).Result()
					if err != nil {
						fmt.Printf("    - %s (type: zset, cardinality: error - %v)\n", key, err)
					} else {
						fmt.Printf("    - %s (type: zset, cardinality: %d)\n", key, card)
					}
				case "hash":
					card, err := rdb.HLen(ctx, key).Result()
					if err != nil {
						fmt.Printf("    - %s (type: hash, fields: error - %v)\n", key, err)
					} else {
						fmt.Printf("    - %s (type: hash, fields: %d)\n", key, card)
					}
				default:
					fmt.Printf("    - %s (type: %s)\n", key, keyType)
				}
			}
		}
	} else {
		fmt.Println("\nNo Asynq-related keys found.")
	}

	// Check for TRON-specific keys mentioned in the code
	fmt.Println("\n=== TRON-specific Keys ===")
	tronKeys := []string{"tron:events", "tron:last_synced_block", "tron:retry_queue", "tron:dlq"}
	for _, key := range tronKeys {
		exists, err := rdb.Exists(ctx, key).Result()
		if err != nil {
			fmt.Printf("  %s: error checking existence - %v\n", key, err)
			continue
		}
		if exists > 0 {
			keyType, err := rdb.Type(ctx, key).Result()
			if err != nil {
				fmt.Printf("  %s: exists, type error - %v\n", key, err)
				continue
			}

			switch keyType {
			case "stream":
				info, err := rdb.XInfoStream(ctx, key).Result()
				if err != nil {
					fmt.Printf("  %s: type: stream, info error - %v\n", key, err)
				} else {
					fmt.Printf("  %s: type: stream, length: %d, first_entry: %s, last_entry: %s\n", 
						key, info.Length, info.FirstEntry.ID, info.LastEntry.ID)
				}
			case "string":
				value, err := rdb.Get(ctx, key).Result()
				if err != nil {
					fmt.Printf("  %s: type: string, value error - %v\n", key, err)
				} else {
					fmt.Printf("  %s: type: string, value: %s\n", key, value)
				}
			case "list":
				length, err := rdb.LLen(ctx, key).Result()
				if err != nil {
					fmt.Printf("  %s: type: list, length error - %v\n", key, err)
				} else {
					fmt.Printf("  %s: type: list, length: %d\n", key, length)
				}
			case "zset":
				card, err := rdb.ZCard(ctx, key).Result()
				if err != nil {
					fmt.Printf("  %s: type: zset, cardinality error - %v\n", key, err)
				} else {
					fmt.Printf("  %s: type: zset, cardinality: %d\n", key, card)
				}
			default:
				fmt.Printf("  %s: type: %s\n", key, keyType)
			}
		} else {
			fmt.Printf("  %s: does not exist\n", key)
		}
	}

	// Check Asynq server info if available
	fmt.Println("\n=== Asynq Server Info ===")
	inspector := asynq.NewInspector(asynq.RedisClientOpt{Addr: redisAddr})
	defer inspector.Close()

	// Get queue names
	queueNames, err := inspector.Queues()
	if err != nil {
		fmt.Printf("Error getting queue names: %v\n", err)
	} else {
		fmt.Printf("Queues found: %v\n", queueNames)
		for _, queue := range queueNames {
			qInfo, err := inspector.GetQueueInfo(queue)
			if err != nil {
				fmt.Printf("  %s: error getting info - %v\n", queue, err)
				continue
			}
			// Note: The correct fields are Active, Pending, Completed, Retry, Paused
			fmt.Printf("  %s: active: %d, pending: %d, completed: %d, retry: %d, paused: %t\n",
				queue, qInfo.Active, qInfo.Pending, qInfo.Completed, qInfo.Retry, qInfo.Paused)
		}
	}

	// Get task info using correct Asynq API methods
	fmt.Println("\n=== Queue Task Details ===")
	for _, queue := range queueNames {
		// Get active tasks (running tasks)
		activeTasks, err := inspector.ListActiveTasks(queue)
		if err != nil {
			fmt.Printf("  Error getting active tasks for queue %s: %v\n", queue, err)
			continue
		}
		fmt.Printf("  Queue '%s': %d active tasks\n", queue, len(activeTasks))
		for i, task := range activeTasks {
			if i < 3 { // Show first 3 tasks as sample
				fmt.Printf("    - Type: %s, Payload: %s\n", task.Type, string(task.Payload))
			}
		}
		if len(activeTasks) > 3 {
			fmt.Printf("    ... and %d more tasks\n", len(activeTasks)-3)
		}

		// Get pending tasks (scheduled tasks)
		pendingTasks, err := inspector.ListPendingTasks(queue)
		if err != nil {
			fmt.Printf("  Error getting pending tasks for queue %s: %v\n", queue, err)
			continue
		}
		fmt.Printf("  Queue '%s': %d pending tasks\n", queue, len(pendingTasks))
		for i, task := range pendingTasks {
			if i < 3 { // Show first 3 tasks as sample
				fmt.Printf("    - Type: %s, Payload: %s\n", task.Type, string(task.Payload))
			}
		}
		if len(pendingTasks) > 3 {
			fmt.Printf("    ... and %d more tasks\n", len(pendingTasks)-3)
		}

		// Get retry tasks
		retryTasks, err := inspector.ListRetryTasks(queue)
		if err != nil {
			fmt.Printf("  Error getting retry tasks for queue %s: %v\n", queue, err)
			continue
		}
		fmt.Printf("  Queue '%s': %d retry tasks\n", queue, len(retryTasks))
		for i, task := range retryTasks {
			if i < 3 { // Show first 3 tasks as sample
				fmt.Printf("    - Type: %s, Payload: %s, Retry Count: %d\n", task.Type, string(task.Payload), task.Retried)
			}
		}
		if len(retryTasks) > 3 {
			fmt.Printf("    ... and %d more tasks\n", len(retryTasks)-3)
		}

		// Check for dead tasks in a separate "dead" queue
		deadQueueInfo, err := inspector.GetQueueInfo("dead")
		if err == nil && deadQueueInfo != nil {
			deadTasksCount := deadQueueInfo.Active + deadQueueInfo.Pending
			fmt.Printf("  Queue '%s': %d dead tasks (in separate 'dead' queue)\n", queue, deadTasksCount)
		} else {
			fmt.Printf("  Queue '%s': no separate dead letter queue found\n", queue)
		}
	}
}