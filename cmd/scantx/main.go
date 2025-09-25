package main

import (
	"fmt"
	"log"

	"github.com/sunbankio/tron--events/scanner"
)

func main() {
	nodeAddress := "grpc://127.0.0.1:50051"
	blockNumber := int64(76036779)

	fmt.Printf("Creating scanner to connect to node: %s\n", nodeAddress)
	fmt.Printf("Scanning block number: %d\n", blockNumber)

	scn, err := scanner.NewScanner(nodeAddress)
	if err != nil {
		log.Fatalf("Failed to create scanner: %v", err)
	}
	defer scn.Close()

	// Use the implemented Scan function to get transactions
	transactions, err := scn.Scan(blockNumber)
	if err != nil {
		log.Fatalf("Failed to scan block %d: %v", blockNumber, err)
	}

	// Print all transactions
	for _, tx := range transactions {
		scanner.PrintTransaction(tx)
	}

	fmt.Printf("Successfully scanned block %d and found %d transactions\n", blockNumber, len(transactions))
}
