package main

import (
	"log"

	"github.com/sunbankio/tronevents/pkg/config"
	"github.com/sunbankio/tronevents/pkg/daemon"
)

func main() {
	// Load configuration
	cfg, err := config.LoadFromFile("config.yaml")
	if err != nil {
		log.Fatal("Failed to load config: ", err)
	}

	// Create and run the daemon service
	service := daemon.NewService(cfg)

	// Handle shutdown
	go daemon.Shutdown(service.WorkerManager(), service.Logger())

	// Start the daemon
	service.Run()
}