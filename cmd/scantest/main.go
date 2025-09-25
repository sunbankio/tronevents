package main

import (
	"fmt"
	"log"
	"os"

	"github.com/sunbankio/tron--events/scanner"
)

func main() {
	nodeAddress := "grpc://127.0.0.1:50051"
	blockNumber := int64(76036779)

	// Check command line arguments
	args := os.Args[1:]
	structured := false
	for _, arg := range args {
		if arg == "--structured" || arg == "-s" {
			structured = true
		}
	}

	fmt.Printf("Creating scanner to connect to node: %s\n", nodeAddress)
	fmt.Printf("Scanning block number: %d\n", blockNumber)

	scanner, err := scanner.NewScanner(nodeAddress)
	if err != nil {
		log.Fatalf("Failed to create scanner: %v", err)
	}
	defer scanner.Close()

	if structured {
		fmt.Println("Using structured output:")
		err = scanner.ScanStructured(blockNumber)
	} else {
		fmt.Println("Using raw output:")
		err = scanner.Scan(blockNumber)
	}

	if err != nil {
		log.Fatalf("Failed to scan block %d: %v", blockNumber, err)
	}

	fmt.Printf("Successfully scanned block %d\n", blockNumber)
}
