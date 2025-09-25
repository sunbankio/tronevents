package main

import (
	"fmt"
	"log"
	"reflect"
	"strings"

	"github.com/sunbankio/tron--events/scanner"
)

func main() {
	nodeAddress := "grpc://127.0.0.1:50051"
	blockNumber := int64(0)

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

	// Print contract parameters dynamically using reflection
	printContractParameters(tx.Contract.Parameter)

	// Display receipt information if available
	if tx.Receipt.EnergyUsage > 0 || tx.Receipt.EnergyFee > 0 || tx.Receipt.OriginEnergyUsage > 0 ||
		tx.Receipt.EnergyUsageTotal > 0 || tx.Receipt.NetUsage > 0 || tx.Receipt.NetFee > 0 {
		fmt.Printf("Receipt Information:\n")
		if tx.Receipt.EnergyUsage > 0 {
			fmt.Printf("  Energy Usage: %d\n", tx.Receipt.EnergyUsage)
		}
		if tx.Receipt.EnergyFee > 0 {
			fmt.Printf("  Energy Fee: %d\n", tx.Receipt.EnergyFee)
		}
		if tx.Receipt.OriginEnergyUsage > 0 {
			fmt.Printf("  Origin Energy Usage: %d\n", tx.Receipt.OriginEnergyUsage)
		}
		if tx.Receipt.EnergyUsageTotal > 0 {
			fmt.Printf("  Energy Usage Total: %d\n", tx.Receipt.EnergyUsageTotal)
		}
		if tx.Receipt.NetUsage > 0 {
			fmt.Printf("  Net Usage: %d\n", tx.Receipt.NetUsage)
		}
		if tx.Receipt.NetFee > 0 {
			fmt.Printf("  Net Fee: %d\n", tx.Receipt.NetFee)
		}
		fmt.Println()
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

// printContractParameters prints contract parameters dynamically using reflection
func printContractParameters(param interface{}) {
	if param == nil {
		return
	}

	v := reflect.ValueOf(param)
	if v.Kind() == reflect.Ptr {
		v = v.Elem()
	}

	if !v.IsValid() {
		return
	}

	t := v.Type()

	for i := 0; i < v.NumField(); i++ {
		field := v.Field(i)
		fieldType := t.Field(i)

		// Get the field name from struct tag if available, otherwise use field name
		name := fieldType.Name
		if jsonTag := fieldType.Tag.Get("json"); jsonTag != "" {
			if commaIdx := strings.Index(jsonTag, ","); commaIdx != -1 {
				name = jsonTag[:commaIdx]
			} else {
				name = jsonTag
			}
		}

		// Skip fields that are meant to be omitted
		if name == "-" {
			continue
		}

		// Format the field name nicely (convert from CamelCase to readable format)
		formattedName := formatFieldName(name)

		// Print the field value based on its type
		switch field.Kind() {
		case reflect.String:
			if field.String() != "" {
				fmt.Printf("%s: %s\n", formattedName, field.String())
			}
		case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
			if field.Int() != 0 {
				fmt.Printf("%s: %d\n", formattedName, field.Int())
			}
		case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
			if field.Uint() != 0 {
				fmt.Printf("%s: %d\n", formattedName, field.Uint())
			}
		case reflect.Bool:
			fmt.Printf("%s: %t\n", formattedName, field.Bool())
		case reflect.Slice:
			if field.Len() > 0 {
				fmt.Printf("%s: %v\n", formattedName, field.Interface())
			}
		case reflect.Array:
			if field.Len() > 0 {
				fmt.Printf("%s: %v\n", formattedName, field.Interface())
			}
		default:
			if field.Interface() != nil {
				fmt.Printf("%s: %v\n", formattedName, field.Interface())
			}
		}
	}
}

// formatFieldName converts a field name from CamelCase to a more readable format
func formatFieldName(name string) string {
	if name == "" {
		return name
	}

	// Simple conversion from CamelCase to readable format
	result := ""
	for i, r := range name {
		if i > 0 && r >= 'A' && r <= 'Z' {
			result += " "
		}
		result += string(r)
	}
	return result
}
