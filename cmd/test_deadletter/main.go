package main

import (
	"encoding/json"
	"fmt"
	"log"
	"time"

	"github.com/hibiken/asynq"
)

func main() {
	// Create Asynq client
	client := asynq.NewClient(asynq.RedisClientOpt{
		Addr: "localhost:6379",
		DB:   0,
	})
	defer client.Close()

	// Create a payload with an impossible block number that will fail
	payload, err := json.Marshal(map[string]interface{}{
		"block_number": 9999999999, // This block number is impossible to process
	})
	if err != nil {
		log.Fatal("Failed to marshal payload:", err)
	}

	// Create the task with the "block:process" type that matches your handler
	task := asynq.NewTask("block:process", payload)

	// Set up task options to ensure it will eventually go to dead letter queue
	taskInfo, err := client.Enqueue(
		task,
		asynq.MaxRetry(5), // Use same retry count as daemon - 5 retries
		asynq.Timeout(30*time.Second),
		asynq.Queue("backlog"), // Use existing backlog queue that has workers
		asynq.ProcessAt(time.Now()), // Process immediately
	)
	if err != nil {
		log.Fatal("Failed to enqueue task:", err)
	}

	fmt.Printf("Enqueued task: %+v\n", taskInfo)
	fmt.Printf("Task ID: %s\n", taskInfo.ID)
	fmt.Printf("Task Type: %s\n", taskInfo.Type)
	fmt.Printf("Task Queue: %s\n", taskInfo.Queue)
	fmt.Println("The task will fail and eventually move to the dead letter queue after retries.")
}