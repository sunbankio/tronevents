package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/sunbankio/tronevents/pkg/config"
	"github.com/sunbankio/tronevents/pkg/daemon"
)

func main() {
	// Load configuration
	cfg, err := config.LoadFromFile("config.yaml")
	if err != nil {
		log.Fatal("Failed to load config: ", err)
	}

	// Create the daemon service
	service := daemon.NewService(cfg)

	// Create a context that is cancelled on interrupt signal
	ctx, cancel := context.WithCancel(context.Background())

	// Handle shutdown
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)

	// Goroutine to handle shutdown signal
	go func() {
		<-c
		log.Println("Shutting down...")
		cancel() // Cancel the context to signal shutdown
		service.WorkerManager().Stop()
	}()

	// Start the daemon with context
	service.RunWithContext(ctx)
}