package scanner

import (
	"encoding/hex"

	"github.com/kslamph/tronlib/pb/core"
)

// parseContract parses a contract based on its type
func parseContract(contract *core.Transaction_Contract) Contract {
	result := Contract{
		PermissionID: int(contract.PermissionId),
	}

	// Parse contract based on type
	switch contract.Type {
	case core.Transaction_Contract_TransferContract:
		// Extract the TransferContract from the Any type
		if contract.Parameter != nil {
			transferContract := &core.TransferContract{}
			if err := contract.Parameter.UnmarshalTo(transferContract); err == nil {
				result.Type = "TransferContract"
				result.Parameter = TransferContract{
					OwnerAddress: byteAddrToString(transferContract.OwnerAddress),
					ToAddress:    byteAddrToString(transferContract.ToAddress),
					Amount:       transferContract.Amount,
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
				if delegateContract.Lock {
					contractData.Lock = true
					contractData.LockPeriod = delegateContract.LockPeriod
				}
				result.Type = "DelegateResourceContract"
				result.Parameter = contractData
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
				result.Type = "UnDelegateResourceContract"
				result.Parameter = contractData
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
				result.Type = "TriggerSmartContract"
				result.Parameter = contractData
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
				result.Type = "FreezeBalanceV2Contract"
				result.Parameter = contractData
			}
		}
	case core.Transaction_Contract_TransferAssetContract:
		// Extract the TransferAssetContract from the Any type
		if contract.Parameter != nil {
			transferAssetContract := &core.TransferAssetContract{}
			if err := contract.Parameter.UnmarshalTo(transferAssetContract); err == nil {
				result.Type = "TransferAssetContract"
				result.Parameter = TransferAssetContract{
					AssetName:    string(transferAssetContract.AssetName),
					OwnerAddress: byteAddrToString(transferAssetContract.OwnerAddress),
					ToAddress:    byteAddrToString(transferAssetContract.ToAddress),
					Amount:       transferAssetContract.Amount,
				}
			}
		}
	case core.Transaction_Contract_AccountCreateContract:
		// Extract the AccountCreateContract from the Any type
		if contract.Parameter != nil {
			accountCreateContract := &core.AccountCreateContract{}
			if err := contract.Parameter.UnmarshalTo(accountCreateContract); err == nil {
				result.Type = "AccountCreateContract"
				result.Parameter = AccountCreateContract{
					OwnerAddress:   byteAddrToString(accountCreateContract.OwnerAddress),
					AccountAddress: byteAddrToString(accountCreateContract.AccountAddress),
					AccountType:    int32(accountCreateContract.Type),
				}
			}
		}
	case core.Transaction_Contract_AccountUpdateContract:
		// Extract the AccountUpdateContract from the Any type
		if contract.Parameter != nil {
			accountUpdateContract := &core.AccountUpdateContract{}
			if err := contract.Parameter.UnmarshalTo(accountUpdateContract); err == nil {
				result.Type = "AccountUpdateContract"
				result.Parameter = AccountUpdateContract{
					OwnerAddress: byteAddrToString(accountUpdateContract.OwnerAddress),
					AccountName:  string(accountUpdateContract.AccountName),
				}
			}
		}
	case core.Transaction_Contract_SetAccountIdContract:
		// Extract the SetAccountIdContract from the Any type
		if contract.Parameter != nil {
			setAccountIdContract := &core.SetAccountIdContract{}
			if err := contract.Parameter.UnmarshalTo(setAccountIdContract); err == nil {
				result.Type = "SetAccountIdContract"
				result.Parameter = SetAccountIdContract{
					OwnerAddress: byteAddrToString(setAccountIdContract.OwnerAddress),
					AccountId:    string(setAccountIdContract.AccountId),
				}
			}
		}
	case core.Transaction_Contract_AccountPermissionUpdateContract:
		// Extract the AccountPermissionUpdateContract from the Any type
		if contract.Parameter != nil {
			permissionUpdateContract := &core.AccountPermissionUpdateContract{}
			if err := contract.Parameter.UnmarshalTo(permissionUpdateContract); err == nil {
				result.Type = "AccountPermissionUpdateContract"
				result.Parameter = AccountPermissionUpdateContract{
					OwnerAddress: byteAddrToString(permissionUpdateContract.OwnerAddress),
					// For now, just set the permission data - we might need to handle the permissions more completely
					OwnerPermission:   permissionUpdateContract.Owner,
					WitnessPermission: permissionUpdateContract.Witness,
					ActivesPermission: permissionUpdateContract.Actives,
				}
			}
		}
	case core.Transaction_Contract_FreezeBalanceContract:
		// Extract the FreezeBalanceContract from the Any type
		if contract.Parameter != nil {
			freezeContract := &core.FreezeBalanceContract{}
			if err := contract.Parameter.UnmarshalTo(freezeContract); err == nil {
				contractData := FreezeBalanceContract{
					OwnerAddress:  byteAddrToString(freezeContract.OwnerAddress),
					FrozenBalance: freezeContract.FrozenBalance,
				}
				if freezeContract.Resource != core.ResourceCode_BANDWIDTH {
					contractData.Resource = "ENERGY"
				} else {
					contractData.Resource = "BANDWIDTH"
				}
				result.Type = "FreezeBalanceContract"
				result.Parameter = contractData
			}
		}
	case core.Transaction_Contract_UnfreezeBalanceContract:
		// Extract the UnfreezeBalanceContract from the Any type
		if contract.Parameter != nil {
			unfreezeContract := &core.UnfreezeBalanceContract{}
			if err := contract.Parameter.UnmarshalTo(unfreezeContract); err == nil {
				contractData := UnfreezeBalanceContract{
					OwnerAddress: byteAddrToString(unfreezeContract.OwnerAddress),
				}
				if unfreezeContract.Resource != core.ResourceCode_BANDWIDTH {
					contractData.Resource = "ENERGY"
				} else {
					contractData.Resource = "BANDWIDTH"
				}
				result.Type = "UnfreezeBalanceContract"
				result.Parameter = contractData
			}
		}
	case core.Transaction_Contract_WithdrawBalanceContract:
		// Extract the WithdrawBalanceContract from the Any type
		if contract.Parameter != nil {
			withdrawContract := &core.WithdrawBalanceContract{}
			if err := contract.Parameter.UnmarshalTo(withdrawContract); err == nil {
				result.Type = "WithdrawBalanceContract"
				result.Parameter = WithdrawBalanceContract{
					OwnerAddress: byteAddrToString(withdrawContract.OwnerAddress),
				}
			}
		}
	case core.Transaction_Contract_UnfreezeBalanceV2Contract:
		// Extract the UnfreezeBalanceV2Contract from the Any type
		if contract.Parameter != nil {
			unfreezeContract := &core.UnfreezeBalanceV2Contract{}
			if err := contract.Parameter.UnmarshalTo(unfreezeContract); err == nil {
				contractData := UnfreezeBalanceV2Contract{
					OwnerAddress:    byteAddrToString(unfreezeContract.OwnerAddress),
					UnfreezeBalance: unfreezeContract.UnfreezeBalance,
				}
				if unfreezeContract.Resource != core.ResourceCode_BANDWIDTH {
					contractData.Resource = "ENERGY"
				} else {
					contractData.Resource = "BANDWIDTH"
				}
				result.Type = "UnfreezeBalanceV2Contract"
				result.Parameter = contractData
			}
		}
	case core.Transaction_Contract_WithdrawExpireUnfreezeContract:
		// Extract the WithdrawExpireUnfreezeContract from the Any type
		if contract.Parameter != nil {
			withdrawContract := &core.WithdrawExpireUnfreezeContract{}
			if err := contract.Parameter.UnmarshalTo(withdrawContract); err == nil {
				result.Type = "WithdrawExpireUnfreezeContract"
				result.Parameter = WithdrawExpireUnfreezeContract{
					OwnerAddress: byteAddrToString(withdrawContract.OwnerAddress),
				}
			}
		}
	case core.Transaction_Contract_CancelAllUnfreezeV2Contract:
		// Extract the CancelAllUnfreezeV2Contract from the Any type
		if contract.Parameter != nil {
			cancelContract := &core.CancelAllUnfreezeV2Contract{}
			if err := contract.Parameter.UnmarshalTo(cancelContract); err == nil {
				result.Type = "CancelAllUnfreezeV2Contract"
				result.Parameter = CancelAllUnfreezeV2Contract{
					OwnerAddress: byteAddrToString(cancelContract.OwnerAddress),
				}
			}
		}
	case core.Transaction_Contract_CreateSmartContract:
		// Extract the CreateSmartContract from the Any type
		if contract.Parameter != nil {
			createContract := &core.CreateSmartContract{}
			if err := contract.Parameter.UnmarshalTo(createContract); err == nil {
				result.Type = "CreateSmartContract"
				result.Parameter = CreateSmartContract{
					OwnerAddress: byteAddrToString(createContract.OwnerAddress),
					NewContract:  createContract.NewContract,
				}
			}
		}

	case core.Transaction_Contract_UpdateSettingContract:
		// Extract the UpdateSettingContract from the Any type
		if contract.Parameter != nil {
			updateSettingContract := &core.UpdateSettingContract{}
			if err := contract.Parameter.UnmarshalTo(updateSettingContract); err == nil {
				result.Type = "UpdateSettingContract"
				result.Parameter = UpdateSettingContract{
					OwnerAddress:               byteAddrToString(updateSettingContract.OwnerAddress),
					ContractAddress:            byteAddrToString(updateSettingContract.ContractAddress),
					ConsumeUserResourcePercent: updateSettingContract.ConsumeUserResourcePercent,
				}
			}
		}
	case core.Transaction_Contract_UpdateEnergyLimitContract:
		// Extract the UpdateEnergyLimitContract from the Any type
		if contract.Parameter != nil {
			updateEnergyContract := &core.UpdateEnergyLimitContract{}
			if err := contract.Parameter.UnmarshalTo(updateEnergyContract); err == nil {
				result.Type = "UpdateEnergyLimitContract"
				result.Parameter = UpdateEnergyLimitContract{
					OwnerAddress:      byteAddrToString(updateEnergyContract.OwnerAddress),
					ContractAddress:   byteAddrToString(updateEnergyContract.ContractAddress),
					OriginEnergyLimit: updateEnergyContract.OriginEnergyLimit,
				}
			}
		}
	case core.Transaction_Contract_ClearABIContract:
		// Extract the ClearABIContract from the Any type
		if contract.Parameter != nil {
			clearContract := &core.ClearABIContract{}
			if err := contract.Parameter.UnmarshalTo(clearContract); err == nil {
				result.Type = "ClearABIContract"
				result.Parameter = ClearABIContract{
					OwnerAddress:    byteAddrToString(clearContract.OwnerAddress),
					ContractAddress: byteAddrToString(clearContract.ContractAddress),
				}
			}
		}
	case core.Transaction_Contract_VoteAssetContract:
		// Extract the VoteAssetContract from the Any type
		if contract.Parameter != nil {
			voteAssetContract := &core.VoteAssetContract{}
			if err := contract.Parameter.UnmarshalTo(voteAssetContract); err == nil {
				votes := make([]VoteAsset, 0, len(voteAssetContract.VoteAddress))
				for _, voteAddr := range voteAssetContract.VoteAddress {
					votes = append(votes, VoteAsset{
						Support:   voteAssetContract.Support,
						AssetName: string(voteAddr),               // Use the vote address as asset name
						VoteCount: int64(voteAssetContract.Count), // Use the count for all votes
					})
				}
				result.Type = "VoteAssetContract"
				result.Parameter = VoteAssetContract{
					OwnerAddress: byteAddrToString(voteAssetContract.OwnerAddress),
					Support:      voteAssetContract.Support,
					Votes:        votes,
				}
			}
		}
	case core.Transaction_Contract_VoteWitnessContract:
		// Extract the VoteWitnessContract from the Any type
		if contract.Parameter != nil {
			voteWitnessContract := &core.VoteWitnessContract{}
			if err := contract.Parameter.UnmarshalTo(voteWitnessContract); err == nil {
				votes := make([]VoteWitness, 0, len(voteWitnessContract.Votes))
				for _, vote := range voteWitnessContract.Votes {
					votes = append(votes, VoteWitness{
						VoteAddress: byteAddrToString(vote.VoteAddress),
						VoteCount:   vote.VoteCount,
					})
				}
				result.Type = "VoteWitnessContract"
				result.Parameter = VoteWitnessContract{
					OwnerAddress: byteAddrToString(voteWitnessContract.OwnerAddress),
					Support:      voteWitnessContract.Support,
					Votes:        votes,
				}
			}
		}
	case core.Transaction_Contract_WitnessCreateContract:
		// Extract the WitnessCreateContract from the Any type
		if contract.Parameter != nil {
			witnessCreateContract := &core.WitnessCreateContract{}
			if err := contract.Parameter.UnmarshalTo(witnessCreateContract); err == nil {
				result.Type = "WitnessCreateContract"
				result.Parameter = WitnessCreateContract{
					OwnerAddress: byteAddrToString(witnessCreateContract.OwnerAddress),
					Url:          string(witnessCreateContract.Url),
				}
			}
		}
	case core.Transaction_Contract_WitnessUpdateContract:
		// Extract the WitnessUpdateContract from the Any type
		if contract.Parameter != nil {
			witnessUpdateContract := &core.WitnessUpdateContract{}
			if err := contract.Parameter.UnmarshalTo(witnessUpdateContract); err == nil {
				result.Type = "WitnessUpdateContract"
				result.Parameter = WitnessUpdateContract{
					OwnerAddress: byteAddrToString(witnessUpdateContract.OwnerAddress),
					UpdateUrl:    string(witnessUpdateContract.UpdateUrl),
				}
			}
		}
	case core.Transaction_Contract_ProposalCreateContract:
		// Extract the ProposalCreateContract from the Any type
		if contract.Parameter != nil {
			proposalCreateContract := &core.ProposalCreateContract{}
			if err := contract.Parameter.UnmarshalTo(proposalCreateContract); err == nil {
				parameters := make(map[int64]int64)
				for key, value := range proposalCreateContract.Parameters {
					parameters[key] = value
				}
				result.Type = "ProposalCreateContract"
				result.Parameter = ProposalCreateContract{
					OwnerAddress: byteAddrToString(proposalCreateContract.OwnerAddress),
					Parameters:   parameters,
				}
			}
		}
	case core.Transaction_Contract_ProposalApproveContract:
		// Extract the ProposalApproveContract from the Any type
		if contract.Parameter != nil {
			proposalApproveContract := &core.ProposalApproveContract{}
			if err := contract.Parameter.UnmarshalTo(proposalApproveContract); err == nil {
				result.Type = "ProposalApproveContract"
				result.Parameter = ProposalApproveContract{
					OwnerAddress: byteAddrToString(proposalApproveContract.OwnerAddress),
					ProposalID:   proposalApproveContract.ProposalId,
					IsApprove:    proposalApproveContract.IsAddApproval,
				}
			}
		}
	case core.Transaction_Contract_ProposalDeleteContract:
		// Extract the ProposalDeleteContract from the Any type
		if contract.Parameter != nil {
			proposalDeleteContract := &core.ProposalDeleteContract{}
			if err := contract.Parameter.UnmarshalTo(proposalDeleteContract); err == nil {
				result.Type = "ProposalDeleteContract"
				result.Parameter = ProposalDeleteContract{
					OwnerAddress: byteAddrToString(proposalDeleteContract.OwnerAddress),
					ProposalID:   proposalDeleteContract.ProposalId,
				}
			}
		}
	case core.Transaction_Contract_ExchangeCreateContract:
		// Extract the ExchangeCreateContract from the Any type
		if contract.Parameter != nil {
			exchangeCreateContract := &core.ExchangeCreateContract{}
			if err := contract.Parameter.UnmarshalTo(exchangeCreateContract); err == nil {
				result.Type = "ExchangeCreateContract"
				result.Parameter = ExchangeCreateContract{
					OwnerAddress:       byteAddrToString(exchangeCreateContract.OwnerAddress),
					FirstTokenId:       string(exchangeCreateContract.FirstTokenId),
					FirstTokenBalance:  exchangeCreateContract.FirstTokenBalance,
					SecondTokenId:      string(exchangeCreateContract.SecondTokenId),
					SecondTokenBalance: exchangeCreateContract.SecondTokenBalance,
				}
			}
		}
	case core.Transaction_Contract_ExchangeInjectContract:
		// Extract the ExchangeInjectContract from the Any type
		if contract.Parameter != nil {
			exchangeInjectContract := &core.ExchangeInjectContract{}
			if err := contract.Parameter.UnmarshalTo(exchangeInjectContract); err == nil {
				result.Type = "ExchangeInjectContract"
				result.Parameter = ExchangeInjectContract{
					OwnerAddress: byteAddrToString(exchangeInjectContract.OwnerAddress),
					ExchangeId:   exchangeInjectContract.ExchangeId,
					TokenId:      string(exchangeInjectContract.TokenId),
					Quant:        exchangeInjectContract.Quant,
				}
			}
		}
	case core.Transaction_Contract_ExchangeWithdrawContract:
		// Extract the ExchangeWithdrawContract from the Any type
		if contract.Parameter != nil {
			exchangeWithdrawContract := &core.ExchangeWithdrawContract{}
			if err := contract.Parameter.UnmarshalTo(exchangeWithdrawContract); err == nil {
				result.Type = "ExchangeWithdrawContract"
				result.Parameter = ExchangeWithdrawContract{
					OwnerAddress: byteAddrToString(exchangeWithdrawContract.OwnerAddress),
					ExchangeId:   exchangeWithdrawContract.ExchangeId,
					TokenId:      string(exchangeWithdrawContract.TokenId),
					Quant:        exchangeWithdrawContract.Quant,
				}
			}
		}
	case core.Transaction_Contract_ExchangeTransactionContract:
		// Extract the ExchangeTransactionContract from the Any type
		if contract.Parameter != nil {
			exchangeTxContract := &core.ExchangeTransactionContract{}
			if err := contract.Parameter.UnmarshalTo(exchangeTxContract); err == nil {
				result.Type = "ExchangeTransactionContract"
				result.Parameter = ExchangeTransactionContract{
					OwnerAddress: byteAddrToString(exchangeTxContract.OwnerAddress),
					ExchangeId:   exchangeTxContract.ExchangeId,
					Symbol:       string(exchangeTxContract.TokenId),
					Quant:        exchangeTxContract.Quant,
					Expected:     exchangeTxContract.Expected,
				}
			}
		}
	case core.Transaction_Contract_MarketSellAssetContract:
		// Extract the MarketSellAssetContract from the Any type
		if contract.Parameter != nil {
			marketSellContract := &core.MarketSellAssetContract{}
			if err := contract.Parameter.UnmarshalTo(marketSellContract); err == nil {
				result.Type = "MarketSellAssetContract"
				result.Parameter = MarketSellAssetContract{
					OwnerAddress:      byteAddrToString(marketSellContract.OwnerAddress),
					SellTokenId:       string(marketSellContract.SellTokenId),
					SellTokenQuantity: marketSellContract.SellTokenQuantity,
					BuyTokenId:        string(marketSellContract.BuyTokenId),
					BuyTokenQuantity:  marketSellContract.BuyTokenQuantity,
				}
			}
		}
	case core.Transaction_Contract_MarketCancelOrderContract:
		// Extract the MarketCancelOrderContract from the Any type
		if contract.Parameter != nil {
			marketCancelContract := &core.MarketCancelOrderContract{}
			if err := contract.Parameter.UnmarshalTo(marketCancelContract); err == nil {
				result.Type = "MarketCancelOrderContract"
				result.Parameter = MarketCancelOrderContract{
					OwnerAddress: byteAddrToString(marketCancelContract.OwnerAddress),
					OrderId:      hex.EncodeToString(marketCancelContract.OrderId),
				}
			}
		}
	case core.Transaction_Contract_CustomContract:
		// For custom contracts, we just store the raw data
		result.Type = "CustomContract"
		result.Parameter = CustomContract{
			Data: "custom contract data",
		}
	case core.Transaction_Contract_UpdateBrokerageContract:
		// Extract the UpdateBrokerageContract from the Any type
		if contract.Parameter != nil {
			updateBrokerageContract := &core.UpdateBrokerageContract{}
			if err := contract.Parameter.UnmarshalTo(updateBrokerageContract); err == nil {
				result.Type = "UpdateBrokerageContract"
				result.Parameter = UpdateBrokerageContract{
					OwnerAddress: byteAddrToString(updateBrokerageContract.OwnerAddress),
					Brokerage:    updateBrokerageContract.Brokerage,
				}
			}
		}
	case core.Transaction_Contract_ShieldedTransferContract:
		// For shielded transfers, just store placeholder
		result.Type = "ShieldedTransferContract"
		result.Parameter = ShieldedTransferContract{}
	default:
		result.Type = contract.Type.String()
	}

	return result
}
