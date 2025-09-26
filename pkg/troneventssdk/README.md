# TRON Events SDK

A simple and easy-to-use Go SDK for subscribing to TRON blockchain events from a Redis stream.

## Features

*   **Simple API**: Just provide a function to handle events and start the subscriber.
*   **Redis Streams**: Built on top of Redis Streams for reliable message delivery.
*   **Consumer Groups**: Supports Redis consumer groups for horizontal scaling.
*   **State Persistence**: Automatically tracks the last processed event ID to resume after restarts.
*   **Configurable**: Highly configurable with options for Redis connection, batching, timeouts, etc.
*   **Pluggable Components**: Supports custom state stores and loggers.

## Installation

```bash
go get github.com/sunbankio/tronevents/pkg/troneventssdk
```

## Usage

### Basic Example

```go
package main

import (
	"context"
	"fmt"
	"log"

	"github.com/sunbankio/tronevents/pkg/troneventssdk"
)

func myEventHandler(ctx context.Context, event troneventssdk.Event) error {
	fmt.Printf("Processing event ID: %s\n", event.ID)
	fmt.Printf("Event data: %v\n", event.Data)
	// Your business logic here
	return nil
}

func main() {
	config := troneventssdk.SubscriberConfig{
		RedisAddr:    "localhost:6379",
		StreamName:   "tron:events",
		GroupName:    "my-app-group",
		ConsumerName: "consumer-1",
	}

	subscriber := troneventssdk.NewSubscriber(config)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	if err := subscriber.Run(ctx, myEventHandler); err != nil {
		log.Fatalf("Subscriber error: %v", err)
	}
}
```

### Configuration Options

The `SubscriberConfig` struct allows you to customize the behavior of the subscriber:

*   `RedisAddr`: The address of the Redis server (e.g., "localhost:6379").
*   `RedisPassword`: The password for the Redis server, if required.
*   `RedisDB`: The Redis database number to use (default is 0).
*   `StreamName`: The name of the Redis stream to subscribe to (e.g., "tron:events").
*   `GroupName`: The name of the Redis consumer group to join.
*   `ConsumerName`: The unique name for this specific consumer instance within the group.
*   `StateStore`: A custom implementation of the `StateStore` interface for persisting the last processed ID. If nil, a default file-based store is used.
*   `StateStorePath`: The path to the file used by the default file-based state store. Only used if `StateStore` is nil. Default is "last_processed_id.txt".
*   `BatchSize`: The maximum number of messages to read from the stream in a single call. Default is 100.
*   `ReadTimeout`: The duration to block when reading from the stream (in seconds). Default is 5 seconds.
*   `Logger`: A custom implementation of the `Logger` interface for SDK logs. If nil, a default logger is used.
*   `InitialPosition`: Determines where to start reading from when no valid state is found (i.e., the state file does not exist). Can be `troneventssdk.InitialPositionBeginning` to process all historical events (default), or `troneventssdk.InitialPositionEnd` to only process new events published after the subscriber starts. If a state file exists but contains a checkpoint for a different consumer group or name, the SDK will error and stop to prevent data integrity issues.

### Custom State Store

You can implement your own state store by implementing the `StateStore` interface:

```go
type StateStore interface {
	SaveLastProcessedID(consumerGroup, consumerName, lastID string) error
	GetLastProcessedID(consumerGroup, consumerName string) (string, error)
}
```

### Custom Logger

You can implement your own logger by implementing the `Logger` interface:

```go
type Logger interface {
	Infof(format string, args ...interface{})
	Errorf(format string, args ...interface{})
	Debugf(format string, args ...interface{})
}
```

## Example

See the [example](example/main.go) directory for a complete working example.