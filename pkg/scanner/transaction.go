package scanner

import (
	"encoding/hex"
	"time"

	"github.com/kslamph/tronlib/pb/api"
	"github.com/kslamph/tronlib/pb/core"
	"github.com/kslamph/tronlib/pkg/eventdecoder"
)

// Transaction represents a parsed TRON transaction
type Transaction struct {
	ID             string    `json:"id"`
	Contract       Contract  `json:"contract"`
	Ret            RetInfo   `json:"ret"`
	Timestamp      time.Time `json:"timestamp"`
	BlockNumber    int64     `json:"block_number,omitempty"`
	BlockTimestamp time.Time `json:"block_timestamp,omitempty"`
	Expiration     time.Time `json:"expiration,omitempty"`
	Receipt        Receipt   `json:"receipt,omitempty"`
	Logs           []LogInfo `json:"logs,omitempty"`
	Signers        []string  `json:"signers,omitempty"` // All signers for the transaction
}

// RetInfo represents the return information of a transaction
type RetInfo struct {
	ContractRet string `json:"contractRet"`
}

// Receipt represents the receipt information of a transaction
type Receipt struct {
	EnergyUsage        int64 `json:"energy_usage,omitempty"`
	EnergyFee          int64 `json:"energy_fee,omitempty"`
	OriginEnergyUsage  int64 `json:"origin_energy_usage,omitempty"`
	EnergyUsageTotal   int64 `json:"energy_usage_total,omitempty"`
	NetUsage           int64 `json:"net_usage,omitempty"`
	NetFee             int64 `json:"net_fee,omitempty"`
}


// LogInfo represents a decoded log event
type LogInfo struct {
	EventName string       `json:"event_name"`
	Signature string       `json:"signature"`
	Inputs    []EventInput `json:"inputs,omitempty"`
	Address   string       `json:"address,omitempty"`
}

// EventInput represents a parameter of a decoded event
type EventInput struct {
	Name  string      `json:"name"`
	Type  string      `json:"type"`
	Value interface{} `json:"value"`
}

// Contract represents the contract details
type Contract struct {
	Type         string      `json:"type"`
	Parameter    interface{} `json:"parameter"`
	PermissionID int         `json:"permission_id"`
}

// parseTransaction converts a raw TRON transaction to a structured format
func parseTransaction(tx *api.TransactionExtention) Transaction {
	transaction := Transaction{
		ID: hex.EncodeToString(tx.Txid),
	}

	// Parse return info
	if tx.Result != nil {
		transaction.Ret = RetInfo{
			ContractRet: tx.Result.Code.String(),
		}
	}

	// Parse timestamp
	if tx.Transaction != nil && tx.Transaction.RawData != nil {
		transaction.Timestamp = time.Unix(tx.Transaction.RawData.Timestamp/1000, 0)

		// Parse expiration time
		if tx.Transaction.RawData.Expiration > 0 {
			transaction.Expiration = time.Unix(tx.Transaction.RawData.Expiration/1000, 0)
		}

		// Parse contract if exists
		if len(tx.Transaction.RawData.Contract) > 0 {
			contract := tx.Transaction.RawData.Contract[0]
			transaction.Contract = parseContract(contract)
		}
	}

	// Extract signers from transaction signatures
	signers, err := recoverSignersFromTransaction(tx)
	if err == nil && len(signers) > 0 {
		transaction.Signers = signers
	}

	return transaction
}

// parseTransactionWithInfo enhances a structured transaction with additional info from TransactionInfo
func parseTransactionWithInfo(tx *api.TransactionExtention, txInfo *core.TransactionInfo) Transaction {
	// First parse the basic transaction
	transaction := parseTransaction(tx)

	// Enhance with additional info from TransactionInfo
	if txInfo != nil {
		// Add block information
		transaction.BlockNumber = txInfo.BlockNumber
		if txInfo.BlockTimeStamp > 0 {
			transaction.BlockTimestamp = time.Unix(txInfo.BlockTimeStamp/1000, 0)
		}

		// Add energy and network usage info
		if txInfo.Receipt != nil {
			transaction.Receipt.EnergyUsage = txInfo.Receipt.EnergyUsage
			transaction.Receipt.EnergyFee = txInfo.Receipt.EnergyFee
			transaction.Receipt.OriginEnergyUsage = txInfo.Receipt.OriginEnergyUsage
			transaction.Receipt.EnergyUsageTotal = txInfo.Receipt.EnergyUsageTotal
			transaction.Receipt.NetUsage = txInfo.Receipt.NetUsage
			transaction.Receipt.NetFee = txInfo.Receipt.NetFee
		}

		// Add logs from TransactionInfo (these are typically more complete)
		if len(txInfo.Log) > 0 {
			transaction.Logs = make([]LogInfo, 0, len(txInfo.Log))
			for _, log := range txInfo.Log {
				// Decode the log using eventdecoder
				decodedEvent, err := eventdecoder.DecodeLog(log.Topics, log.Data)
				if err == nil {
					logInfo := LogInfo{
						EventName: decodedEvent.EventName,
						Address:   byteAddrToString(log.Address),
					}

					// Add signature from the first topic if available
					if len(log.Topics) > 0 {
						logInfo.Signature = hex.EncodeToString(log.Topics[0])
					}

					// Convert decoded event parameters
					if len(decodedEvent.Parameters) > 0 {
						logInfo.Inputs = make([]EventInput, len(decodedEvent.Parameters))
						for i, param := range decodedEvent.Parameters {
							logInfo.Inputs[i] = EventInput{
								Name:  param.Name,
								Type:  param.Type,
								Value: param.Value,
							}
						}
					}

					transaction.Logs = append(transaction.Logs, logInfo)
				} else {
					// If we can't decode the log, still include basic info
					logInfo := LogInfo{
						Address: byteAddrToString(log.Address),
					}
					// Add signature from the first topic if available even when decoding fails
					if len(log.Topics) > 0 {
						logInfo.Signature = hex.EncodeToString(log.Topics[0])
					}
					transaction.Logs = append(transaction.Logs, logInfo)
				}
			}
		}
	}

	return transaction
}
