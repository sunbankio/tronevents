package main

import (
	"context"
	"fmt"
	"log"

	"github.com/sunbankio/tronevents/pkg/troneventssdk"
)

// myEventHandler is the function that will be called for each event received from the stream.
func myEventHandler(ctx context.Context, event troneventssdk.Event) error {
	fmt.Printf("Processing event ID: %s\n", event.ID)
	fmt.Printf("Event data: %v\n", event.Data)

	// Example: Access specific fields from the event data
	if contractAddr, ok := event.Data["contract_address"]; ok {
		fmt.Printf("Contract Address: %v\n", contractAddr)
	}
	if eventName, ok := event.Data["event_name"]; ok {
		fmt.Printf("Event Name: %v\n", eventName)
	}

	// Your business logic goes here
	// For example, you might store the event in a database, trigger an alert, etc.

	return nil
}

func main() {
	// Configure the subscriber
	config := troneventssdk.SubscriberConfig{
		RedisAddr:    "localhost:6379", // Replace with your Redis address
		StreamName:   "tron:events",    // The name of the Redis stream
		GroupName:    "my-app-group",   // The consumer group name
		ConsumerName: "consumer-1",     // Unique name for this consumer instance
		// Optional: Set other configuration options
		// StateStorePath: "my_checkpoint.txt", // Custom path for state file
		// BatchSize: 50,                       // Custom batch size
		// ReadTimeout: 10,                     // Custom read timeout in seconds
		// InitialPosition: troneventssdk.InitialPositionEnd, // Start from new events only (default is "beginning")
	}

	// Create the subscriber
	subscriber := troneventssdk.NewSubscriber(config)

	// Start the subscriber with the event handler
	// The Run method blocks, so we use a context that can be cancelled
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	fmt.Println("Starting TRON event subscriber...")
	if err := subscriber.Run(ctx, myEventHandler); err != nil {
		log.Fatalf("Subscriber error: %v", err)
	}
}