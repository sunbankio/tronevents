package scanner

import (
	"fmt"

	"github.com/kslamph/tronlib/pkg/types"
)

// PrintTransaction prints a transaction in a human-readable format
func PrintTransaction(tx Transaction) {
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
	case TransferContract:
		fmt.Printf("From: %s\n", param.OwnerAddress)
		fmt.Printf("To: %s\n", param.ToAddress)
		fmt.Printf("Amount: %d\n", param.Amount)
	case DelegateResourceContract:
		fmt.Printf("From: %s\n", param.OwnerAddress)
		fmt.Printf("To: %s\n", param.ReceiverAddress)
		fmt.Printf("Resource: %s\n", param.Resource)
		fmt.Printf("Balance: %d\n", param.Balance)
	case UnDelegateResourceContract:
		fmt.Printf("From: %s\n", param.OwnerAddress)
		fmt.Printf("To: %s\n", param.ReceiverAddress)
		fmt.Printf("Resource: %s\n", param.Resource)
		fmt.Printf("Balance: %d\n", param.Balance)
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

	fmt.Printf("Result: %s\n", tx.Ret.ContractRet)
	fmt.Println()
}

// only use this func if the addr []byte is guaranteed to be a valid address
func byteAddrToString(addr []byte) string {
	return types.MustNewAddressFromBytes(addr).String() // validate
}
