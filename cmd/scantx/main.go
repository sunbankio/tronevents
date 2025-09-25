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
		printTransaction(tx)
	}

	fmt.Printf("Successfully scanned block %d and found %d transactions\n", blockNumber, len(transactions))
}

// PrintTransaction prints a transaction in a human-readable format
func printTransaction(tx scanner.Transaction) {
	fmt.Println("----- Transaction -----")
	fmt.Printf("Transaction ID: %s\n", tx.ID)
	fmt.Printf("Timestamp: %s\n", tx.Timestamp.Format("2006-01-02 15:04:05"))
	if tx.BlockNumber > 0 {
		fmt.Printf("Block Number: %d\n", tx.BlockNumber)
	}
	if !tx.BlockTimestamp.IsZero() {
		fmt.Printf("Block Timestamp: %s\n", tx.BlockTimestamp.Format("2006-01-02 15:04:05"))
	}
	if !tx.Expiration.IsZero() {
		fmt.Printf("Expiration: %s\n", tx.Expiration.Format("2006-01-02 15:04:05"))
	}
	fmt.Printf("Contract Type: %s\n", tx.Contract.Type)
	if tx.Contract.PermissionID != 0 {
		fmt.Printf("Contract Permission ID: %d\n", tx.Contract.PermissionID)
	}

	switch param := tx.Contract.Parameter.(type) {
	case scanner.TransferContract:
		fmt.Printf("From: %s\n", param.OwnerAddress)
		fmt.Printf("To: %s\n", param.ToAddress)
		fmt.Printf("Amount: %d\n", param.Amount)
	case scanner.DelegateResourceContract:
		fmt.Printf("From: %s\n", param.OwnerAddress)
		fmt.Printf("To: %s\n", param.ReceiverAddress)
		fmt.Printf("Resource: %s\n", param.Resource)
		fmt.Printf("Balance: %d\n", param.Balance)
	case scanner.UnDelegateResourceContract:
		fmt.Printf("From: %s\n", param.OwnerAddress)
		fmt.Printf("To: %s\n", param.ReceiverAddress)
		fmt.Printf("Resource: %s\n", param.Resource)
		fmt.Printf("Balance: %d\n", param.Balance)
	case scanner.TriggerSmartContract:
		fmt.Printf("Owner: %s\n", param.OwnerAddress)
		fmt.Printf("Contract: %s\n", param.ContractAddress)
		fmt.Printf("Data: %s\n", param.Data)
		if param.CallValue != 0 {
			fmt.Printf("Call Value: %d\n", param.CallValue)
		}
		if param.FeeLimit != 0 {
			fmt.Printf("Fee Limit: %d\n", param.FeeLimit)
		}
	case scanner.FreezeBalanceV2Contract:
		fmt.Printf("Owner: %s\n", param.OwnerAddress)
		fmt.Printf("Resource: %s\n", param.Resource)
		fmt.Printf("Frozen Balance: %d\n", param.FrozenBalance)
	case scanner.TransferAssetContract:
		fmt.Printf("Asset Name: %s\n", param.AssetName)
		fmt.Printf("From: %s\n", param.OwnerAddress)
		fmt.Printf("To: %s\n", param.ToAddress)
		fmt.Printf("Amount: %d\n", param.Amount)
	}

	// Display energy used if available
	if tx.EnergyUsed > 0 {
		fmt.Printf("Energy Used: %d\n", tx.EnergyUsed)
	}

	// Display bandwidth used if available
	if tx.BandwidthUsed > 0 {
		fmt.Printf("Bandwidth Used: %d\n", tx.BandwidthUsed)
	}

	// Display logs if available
	if len(tx.Logs) > 0 {
		fmt.Printf("Logs (%d):\n", len(tx.Logs))
		for i, log := range tx.Logs {
			fmt.Printf("  Log %d:\n", i+1)
			if log.EventName != "" {
				fmt.Printf("    Event: %s\n", log.EventName)
				fmt.Printf("    Signature: %s\n", log.Signature)
				if len(log.Inputs) > 0 {
					fmt.Printf("    Inputs:\n")
					for _, input := range log.Inputs {
						fmt.Printf("      %s (%s): %v\n", input.Name, input.Type, input.Value)
					}
				}
			}
			if log.Address != "" {
				fmt.Printf("    Address: %s\n", log.Address)
			}
		}
	}

	// Display signers if available
	if len(tx.Signers) > 0 {
		fmt.Printf("Signers (%d):\n", len(tx.Signers))
		for i, signer := range tx.Signers {
			fmt.Printf("  Signer %d: %s\n", i+1, signer)
		}
	}

	fmt.Printf("Result: %s\n", tx.Ret.ContractRet)
	fmt.Println()
}
