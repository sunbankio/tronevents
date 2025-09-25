package scanner

import (
	"encoding/hex"
	"time"

	"github.com/kslamph/tronlib/pb/api"
	"github.com/kslamph/tronlib/pb/core"
	"github.com/kslamph/tronlib/pkg/eventdecoder"
	"github.com/kslamph/tronlib/pkg/utils"
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
	EnergyUsed     int64     `json:"energy_used,omitempty"`
	BandwidthUsed  int64     `json:"bandwidth_used,omitempty"`
	Logs           []LogInfo `json:"logs,omitempty"`
	Signers        []string  `json:"signers,omitempty"` // All signers for the transaction
}

// RetInfo represents the return information of a transaction
type RetInfo struct {
	ContractRet string `json:"contractRet"`
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

// TransferContract represents a transfer transaction
type TransferContract struct {
	OwnerAddress string `json:"owner_address"`
	ToAddress    string `json:"to_address"`
	Amount       int64  `json:"amount"`
}

// DelegateResourceContract represents a resource delegation transaction
type DelegateResourceContract struct {
	OwnerAddress    string `json:"owner_address"`
	ReceiverAddress string `json:"receiver_address"`
	Resource        string `json:"resource,omitempty"`
	Balance         int64  `json:"balance"`
	Lock            bool   `json:"lock,omitempty"`
	LockPeriod      int64  `json:"lock_period,omitempty"`
}

// UnDelegateResourceContract represents a resource undelegation transaction
type UnDelegateResourceContract struct {
	OwnerAddress    string `json:"owner_address"`
	ReceiverAddress string `json:"receiver_address"`
	Resource        string `json:"resource,omitempty"`
	Balance         int64  `json:"balance"`
}

// TriggerSmartContract represents a smart contract trigger transaction
type TriggerSmartContract struct {
	OwnerAddress    string `json:"owner_address"`
	ContractAddress string `json:"contract_address"`
	Data            string `json:"data"`
	CallValue       int64  `json:"call_value,omitempty"`
	FeeLimit        int64  `json:"fee_limit,omitempty"`
}

// FreezeBalanceV2Contract represents a balance freezing transaction
type FreezeBalanceV2Contract struct {
	OwnerAddress  string `json:"owner_address"`
	FrozenBalance int64  `json:"frozen_balance"`
	Resource      string `json:"resource"`
}

// TransferAssetContract represents an asset transfer transaction
type TransferAssetContract struct {
	AssetName    string `json:"asset_name"`
	OwnerAddress string `json:"owner_address"`
	ToAddress    string `json:"to_address"`
	Amount       int64  `json:"amount"`
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

			// Parse contract based on type
			switch contract.Type {
			case core.Transaction_Contract_TransferContract:
				// Extract the TransferContract from the Any type
				if contract.Parameter != nil {
					transferContract := &core.TransferContract{}
					if err := contract.Parameter.UnmarshalTo(transferContract); err == nil {
						transaction.Contract = Contract{
							Type: "TransferContract",
							Parameter: TransferContract{
								OwnerAddress: byteAddrToString(transferContract.OwnerAddress),
								ToAddress:    byteAddrToString(transferContract.ToAddress),
								Amount:       transferContract.Amount,
							},
						}
					}
				}
			case core.Transaction_Contract_DelegateResourceContract:
				// Extract the DelegateResourceContract from the Any type
				if contract.Parameter != nil {
					delegateContract := &core.DelegateResourceContract{}
					if err := contract.Parameter.UnmarshalTo(delegateContract); err == nil {
						contractData := DelegateResourceContract{
							OwnerAddress:    byteAddrToString(delegateContract.OwnerAddress),
							ReceiverAddress: byteAddrToString(delegateContract.ReceiverAddress),
							Balance:         delegateContract.Balance,
						}
						if delegateContract.Resource != core.ResourceCode_BANDWIDTH {
							contractData.Resource = "ENERGY"
						} else {
							contractData.Resource = "BANDWIDTH"
						}
						transaction.Contract = Contract{
							Type:      "DelegateResourceContract",
							Parameter: contractData,
						}
					}
				}
			case core.Transaction_Contract_UnDelegateResourceContract:
				// Extract the UnDelegateResourceContract from the Any type
				if contract.Parameter != nil {
					undelegateContract := &core.UnDelegateResourceContract{}
					if err := contract.Parameter.UnmarshalTo(undelegateContract); err == nil {
						contractData := UnDelegateResourceContract{
							OwnerAddress:    byteAddrToString(undelegateContract.OwnerAddress),
							ReceiverAddress: byteAddrToString(undelegateContract.ReceiverAddress),
							Balance:         undelegateContract.Balance,
						}
						if undelegateContract.Resource != core.ResourceCode_BANDWIDTH {
							contractData.Resource = "ENERGY"
						} else {
							contractData.Resource = "BANDWIDTH"
						}
						transaction.Contract = Contract{
							Type:      "UnDelegateResourceContract",
							Parameter: contractData,
						}
					}
				}
			case core.Transaction_Contract_TriggerSmartContract:
				// Extract the TriggerSmartContract from the Any type
				if contract.Parameter != nil {
					triggerContract := &core.TriggerSmartContract{}
					if err := contract.Parameter.UnmarshalTo(triggerContract); err == nil {
						contractData := TriggerSmartContract{
							OwnerAddress:    byteAddrToString(triggerContract.OwnerAddress),
							ContractAddress: byteAddrToString(triggerContract.ContractAddress),
							Data:            hex.EncodeToString(triggerContract.Data),
						}
						if triggerContract.CallValue != 0 {
							contractData.CallValue = triggerContract.CallValue
						}
						// Skip FeeLimit as it may not be available in this version
						transaction.Contract = Contract{
							Type:      "TriggerSmartContract",
							Parameter: contractData,
						}
					}
				}
			case core.Transaction_Contract_FreezeBalanceV2Contract:
				// Extract the FreezeBalanceV2Contract from the Any type
				if contract.Parameter != nil {
					freezeContract := &core.FreezeBalanceV2Contract{}
					if err := contract.Parameter.UnmarshalTo(freezeContract); err == nil {
						contractData := FreezeBalanceV2Contract{
							OwnerAddress:  byteAddrToString(freezeContract.OwnerAddress),
							FrozenBalance: freezeContract.FrozenBalance,
						}
						if freezeContract.Resource != core.ResourceCode_BANDWIDTH {
							contractData.Resource = "ENERGY"
						} else {
							contractData.Resource = "BANDWIDTH"
						}
						transaction.Contract = Contract{
							Type:      "FreezeBalanceV2Contract",
							Parameter: contractData,
						}
					}
				}
			case core.Transaction_Contract_TransferAssetContract:
				// Extract the TransferAssetContract from the Any type
				if contract.Parameter != nil {
					transferAssetContract := &core.TransferAssetContract{}
					if err := contract.Parameter.UnmarshalTo(transferAssetContract); err == nil {
						transaction.Contract = Contract{
							Type: "TransferAssetContract",
							Parameter: TransferAssetContract{
								AssetName:    string(transferAssetContract.AssetName),
								OwnerAddress: byteAddrToString(transferAssetContract.OwnerAddress),
								ToAddress:    byteAddrToString(transferAssetContract.ToAddress),
								Amount:       transferAssetContract.Amount,
							},
						}
					}
				}
			default:
				transaction.Contract = Contract{
					Type: contract.Type.String(),
				}
			}

			// Handle permission ID
			transaction.Contract.PermissionID = int(contract.PermissionId)
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

		// Add energy usage info
		if txInfo.Receipt != nil {
			if txInfo.Receipt.EnergyUsage > 0 {
				transaction.EnergyUsed = txInfo.Receipt.EnergyUsage
			}
			// Add net usage info
			if txInfo.Receipt.NetUsage > 0 {
				transaction.BandwidthUsed = txInfo.Receipt.NetUsage
			}
			// We could also add other energy-related fields if needed:
			// EnergyUsageTotal, OriginEnergyUsage, EnergyPenaltyTotal
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
					transaction.Logs = append(transaction.Logs, logInfo)
				}
			}
		}
	}

	return transaction
}

// recoverSignersFromTransaction recovers all signer addresses from transaction signatures using the tronlib utility
func recoverSignersFromTransaction(tx *api.TransactionExtention) ([]string, error) {
	if tx.Transaction == nil {
		return nil, nil
	}

	// Use the tronlib utility function to extract signers
	signerAddresses, err := utils.ExtractSigners(tx.Transaction)
	if err != nil {
		return nil, err
	}

	// Convert the types.Address objects to strings
	signers := make([]string, 0, len(signerAddresses))
	for _, addr := range signerAddresses {
		if addr != nil {
			signers = append(signers, addr.String())
		}
	}

	return signers, nil
}
