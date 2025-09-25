package scanner

import (
	"context"
	"encoding/hex"
	"fmt"
	"time"

	"github.com/kslamph/tronlib/pb/api"
	"github.com/kslamph/tronlib/pb/core"
	"github.com/kslamph/tronlib/pkg/client"
	"github.com/kslamph/tronlib/pkg/eventdecoder"
)

const (
	DefaultTimeout     = 10 // seconds
	DefaultPoolSize    = 5
	DefaultMaxPoolSize = 10
)

type Scanner struct {
	tronclient *client.Client
	ctx        context.Context
}

// Default values

func NewScanner(nodeAddress string) (*Scanner, error) {
	// Create the client with the original NewClient function
	tronclient, err := client.NewClient(nodeAddress,
		client.WithTimeout(time.Duration(DefaultTimeout)*time.Second),
		client.WithPool(DefaultPoolSize, DefaultMaxPoolSize),
	)
	if err != nil {
		return nil, err
	}

	return &Scanner{
		tronclient: tronclient,
		ctx:        context.Background(),
	}, nil
}

func (s *Scanner) Close() {
	s.tronclient.Close()
}

func (s *Scanner) Scan(blockNumber int64) ([]Transaction, error) {
	// Get both block data and transaction info
	block, err := s.getBlockByNumber(blockNumber)
	if err != nil {
		return nil, err
	}

	txInfoList, err := s.getTransactionInfoByNumber(blockNumber)
	if err != nil {
		// If we can't get transaction info, fall back to basic parsing
		transactions := make([]Transaction, 0, len(block.Transactions))
		for _, tx := range block.Transactions {
			transaction := ParseTransaction(tx)
			transactions = append(transactions, transaction)
		}
		return transactions, nil
	}

	// Create a map of transaction info by transaction ID for easy lookup
	txInfoMap := make(map[string]*core.TransactionInfo)
	for _, txInfo := range txInfoList.TransactionInfo {
		txID := hex.EncodeToString(txInfo.Id)
		txInfoMap[txID] = txInfo
	}

	// Process each transaction with enhanced data
	transactions := make([]Transaction, 0, len(block.Transactions))
	for _, tx := range block.Transactions {
		txID := hex.EncodeToString(tx.Txid)

		// Parse the basic transaction first
		transaction := ParseTransaction(tx)

		// Look for corresponding transaction info and enhance the transaction
		if txInfo, exists := txInfoMap[txID]; exists {
			// Extract energy usage and logs from txInfo and add to transaction
			transaction = ParseTransactionWithInfo(tx, txInfo)
		}

		transactions = append(transactions, transaction)
	}

	return transactions, nil
}

func (s *Scanner) ScanStructured(blockNumber int64) error {
	// Get both block data and transaction info
	block, err := s.getBlockByNumber(blockNumber)
	if err != nil {
		return err
	}

	txInfoList, err := s.getTransactionInfoByNumber(blockNumber)
	if err != nil {
		// If we can't get transaction info, fall back to basic parsing
		fmt.Println("Warning: Could not retrieve transaction info, using basic transaction parsing")
		for _, tx := range block.Transactions {
			transaction := ParseTransaction(tx)
			PrintTransaction(transaction)
		}
		return nil
	}

	// Create a map of transaction info by transaction ID for easy lookup
	txInfoMap := make(map[string]*core.TransactionInfo)
	for _, txInfo := range txInfoList.TransactionInfo {
		txID := hex.EncodeToString(txInfo.Id)
		txInfoMap[txID] = txInfo
	}

	// Process each transaction with enhanced data
	for _, tx := range block.Transactions {
		txID := hex.EncodeToString(tx.Txid)

		// Parse the basic transaction first
		transaction := ParseTransaction(tx)

		// Look for corresponding transaction info and enhance the transaction
		if txInfo, exists := txInfoMap[txID]; exists {
			// Extract energy usage and logs from txInfo and add to transaction
			transaction = ParseTransactionWithInfo(tx, txInfo)
		}

		PrintTransaction(transaction)
	}
	return nil
}

func (s *Scanner) ScanTriggerDebug(blockNumber int64) error {
	block, err := s.getBlockByNumber(blockNumber)
	if err != nil {
		return err
	}

	fmt.Printf("Found %d transactions in block\n", len(block.Transactions))

	for _, tx := range block.Transactions {
		// Check if this is a TriggerSmartContract transaction
		if tx.Transaction != nil && tx.Transaction.RawData != nil && len(tx.Transaction.RawData.Contract) > 0 {
			contract := tx.Transaction.RawData.Contract[0]
			if contract.Type.String() == "TriggerSmartContract" {
				fmt.Printf("Processing TriggerSmartContract transaction\n")
				transaction := ParseTransaction(tx)
				PrintTransaction(transaction)
			}
		}
	}
	return nil
}

// ScanTransactionInfo scans a block and displays combined information from both
// TransactionExtension (structured transaction data) and TransactionInfo (energy/logs)
func (s *Scanner) ScanTransactionInfo(blockNumber int64) error {
	// Get transaction info list (this contains energy usage and logs)
	txInfoList, err := s.getTransactionInfoByNumber(blockNumber)
	if err != nil {
		return err
	}

	fmt.Printf("Found %d transaction info entries\n", len(txInfoList.TransactionInfo))

	// Print transaction info with energy usage and logs
	for i, txInfo := range txInfoList.TransactionInfo {
		fmt.Printf("\n--- Transaction Info %d ---\n", i+1)
		fmt.Printf("Transaction ID: %x\n", txInfo.Id)
		fmt.Println("raw data:", txInfo)
		// Print energy usage information from receipt
		if txInfo.Receipt != nil {
			if txInfo.Receipt.EnergyUsage > 0 {
				fmt.Printf("Energy Usage: %d\n", txInfo.Receipt.EnergyUsage)
			}
			if txInfo.Receipt.EnergyUsageTotal > 0 {
				fmt.Printf("Energy Usage Total: %d\n", txInfo.Receipt.EnergyUsageTotal)
			}
			if txInfo.Receipt.OriginEnergyUsage > 0 {
				fmt.Printf("Origin Energy Usage: %d\n", txInfo.Receipt.OriginEnergyUsage)
			}
			if txInfo.Receipt.EnergyPenaltyTotal > 0 {
				fmt.Printf("Energy Penalty Total: %d\n", txInfo.Receipt.EnergyPenaltyTotal)
			}
			if txInfo.Receipt.NetUsage > 0 {
				fmt.Printf("Net Usage: %d\n", txInfo.Receipt.NetUsage)
			}
		}

		// Print logs if available
		if len(txInfo.Log) > 0 {
			fmt.Printf("Logs (%d):\n", len(txInfo.Log))
			for j, log := range txInfo.Log {
				fmt.Printf("  Log %d:\n", j+1)
				fmt.Printf("    Address: %s\n", byteAddrToString(log.Address))
				fmt.Printf("    Number of Topics: %d\n", len(log.Topics))
				if len(log.Data) > 0 {
					fmt.Printf("    Data Length: %d bytes\n", len(log.Data))
				} else {
					fmt.Printf("    Data: None\n")
				}

				// Try to decode logs using eventdecoder
				if len(log.Topics) > 0 {
					decodedEvent, err := eventdecoder.DecodeLog(log.Topics, log.Data)
					if err == nil {
						fmt.Printf("    Decoded Event Name: %s\n", decodedEvent.EventName)
						if len(decodedEvent.Parameters) > 0 {
							fmt.Printf("    Event Parameters:\n")
							for _, param := range decodedEvent.Parameters {
								fmt.Printf("      %s (%s): %v\n", param.Name, param.Type, param.Value)
							}
						}
					}
				}
			}
		} else {
			fmt.Printf("Logs: None\n")
		}

		fmt.Printf("Result: %s\n", txInfo.Result)

		// Limit output for readability
		// if i >= 9 {
		// 	fmt.Printf("\n... (%d more transactions)\n", len(txInfoList.TransactionInfo)-i-1)
		// 	break
		// }
	}

	return nil
}

func (s *Scanner) getBlockByNumber(blockNumber int64) (*api.BlockExtention, error) {
	return s.tronclient.Network().GetBlockByNumber(s.ctx, blockNumber)
}

func (s *Scanner) getTransactionInfoByNumber(blockNumber int64) (*api.TransactionInfoList, error) {
	return s.tronclient.Network().GetTransactionInfoByBlockNum(s.ctx, blockNumber)
}
