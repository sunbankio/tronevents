package scanner

import (
	"encoding/hex"
	"fmt"
	"time"

	"github.com/kslamph/tronlib/pb/api"
	"github.com/kslamph/tronlib/pb/core"
	"github.com/kslamph/tronlib/pkg/eventdecoder"
	"github.com/kslamph/tronlib/pkg/types"
)

// Transaction represents a parsed TRON transaction
type Transaction struct {
	ID         string    `json:"id"`
	Contract   Contract  `json:"contract"`
	Ret        RetInfo   `json:"ret"`
	Timestamp  time.Time `json:"timestamp"`
	EnergyUsed int64     `json:"energy_used,omitempty"`
	Logs       []LogInfo `json:"logs,omitempty"`
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
	PermissionID int         `json:"permission_id,omitempty"`
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
	PermissionID    int    `json:"permission_id,omitempty"`
	Lock            bool   `json:"lock,omitempty"`
	LockPeriod      int64  `json:"lock_period,omitempty"`
}

// UnDelegateResourceContract represents a resource undelegation transaction
type UnDelegateResourceContract struct {
	OwnerAddress    string `json:"owner_address"`
	ReceiverAddress string `json:"receiver_address"`
	Resource        string `json:"resource,omitempty"`
	Balance         int64  `json:"balance"`
	PermissionID    int    `json:"permission_id,omitempty"`
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

// ParseTransaction converts a raw TRON transaction to a structured format
func ParseTransaction(tx *api.TransactionExtention) Transaction {
	transaction := Transaction{
		ID: hex.EncodeToString(tx.Txid),
	}

	// Parse return info
	if tx.Result != nil {
		transaction.Ret = RetInfo{
			ContractRet: tx.Result.Code.String(),
		}
	}

	// Parse energy used
	if tx.EnergyUsed > 0 {
		transaction.EnergyUsed = tx.EnergyUsed
	}

	// Parse logs
	if len(tx.Logs) > 0 {
		transaction.Logs = make([]LogInfo, 0, len(tx.Logs))
		for _, log := range tx.Logs {
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

	// Parse timestamp
	if tx.Transaction != nil && tx.Transaction.RawData != nil {
		transaction.Timestamp = time.Unix(tx.Transaction.RawData.Timestamp/1000, 0)

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
						// Skip PermissionId as it may not be available in this version
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
						// Skip PermissionId as it may not be available in this version
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

			// Handle permission ID if present (skip for now as it may not be available)
			// if contract.PermissionId != 0 {
			// 	transaction.Contract.PermissionID = int(contract.PermissionId)
			// }
		}
	}

	return transaction
}

// PrintTransaction prints a transaction in a human-readable format
func PrintTransaction(tx Transaction) {
	fmt.Println("----- Transaction -----")
	fmt.Printf("Transaction ID: %s\n", tx.ID)
	fmt.Printf("Timestamp: %s\n", tx.Timestamp.Format("2006-01-02 15:04:05"))
	fmt.Printf("Contract Type: %s\n", tx.Contract.Type)

	switch param := tx.Contract.Parameter.(type) {
	case TransferContract:
		fmt.Printf("From: %s\n", param.OwnerAddress)
		fmt.Printf("To: %s\n", param.ToAddress)
		fmt.Printf("Amount: %d\n", param.Amount)
	case DelegateResourceContract:
		fmt.Printf("From: %s\n", param.OwnerAddress)
		fmt.Printf("To: %s\n", param.ReceiverAddress)
		fmt.Printf("Resource: %s\n", param.Resource)
		fmt.Printf("Balance: %d\n", param.Balance)
		if param.PermissionID != 0 {
			fmt.Printf("Permission ID: %d\n", param.PermissionID)
		}
	case UnDelegateResourceContract:
		fmt.Printf("From: %s\n", param.OwnerAddress)
		fmt.Printf("To: %s\n", param.ReceiverAddress)
		fmt.Printf("Resource: %s\n", param.Resource)
		fmt.Printf("Balance: %d\n", param.Balance)
		if param.PermissionID != 0 {
			fmt.Printf("Permission ID: %d\n", param.PermissionID)
		}
	case TriggerSmartContract:
		fmt.Printf("Owner: %s\n", param.OwnerAddress)
		fmt.Printf("Contract: %s\n", param.ContractAddress)
		fmt.Printf("Data: %s\n", param.Data)
		if param.CallValue != 0 {
			fmt.Printf("Call Value: %d\n", param.CallValue)
		}
		if param.FeeLimit != 0 {
			fmt.Printf("Fee Limit: %d\n", param.FeeLimit)
		}
	case FreezeBalanceV2Contract:
		fmt.Printf("Owner: %s\n", param.OwnerAddress)
		fmt.Printf("Resource: %s\n", param.Resource)
		fmt.Printf("Frozen Balance: %d\n", param.FrozenBalance)
	case TransferAssetContract:
		fmt.Printf("Asset Name: %s\n", param.AssetName)
		fmt.Printf("From: %s\n", param.OwnerAddress)
		fmt.Printf("To: %s\n", param.ToAddress)
		fmt.Printf("Amount: %d\n", param.Amount)
	}

	// Display energy used if available
	if tx.EnergyUsed > 0 {
		fmt.Printf("Energy Used: %d\n", tx.EnergyUsed)
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

	fmt.Printf("Result: %s\n", tx.Ret.ContractRet)
	fmt.Println()
}

// only use this func if the addr []byte is guaranteed to be a valid address
func byteAddrToString(addr []byte) string {
	return types.MustNewAddressFromBytes(addr).String() // validate
}
