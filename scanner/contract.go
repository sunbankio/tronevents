package scanner

import "github.com/kslamph/tronlib/pb/core"

// Account contracts
// AccountCreateContract represents an account creation transaction
type AccountCreateContract struct {
	OwnerAddress   string `json:"owner_address"`
	AccountAddress string `json:"account_address"`
	AccountType    int32  `json:"account_type,omitempty"`
}

// AccountUpdateContract represents an account update transaction
type AccountUpdateContract struct {
	OwnerAddress string `json:"owner_address"`
	AccountName  string `json:"account_name"`
}

// SetAccountIdContract represents a set account ID transaction
type SetAccountIdContract struct {
	OwnerAddress string `json:"owner_address"`
	AccountId    string `json:"account_id"`
}

// AccountPermissionUpdateContract represents an account permission update transaction
type AccountPermissionUpdateContract struct {
	OwnerAddress      string             `json:"owner_address"`
	OwnerPermission   *core.Permission   `json:"owner_permission,omitempty"`
	WitnessPermission *core.Permission   `json:"witness_permission,omitempty"`
	ActivesPermission []*core.Permission `json:"actives_permission,omitempty"`
}

// Asset contracts
// TransferContract represents a transfer transaction
type TransferContract struct {
	OwnerAddress string `json:"owner_address"`
	ToAddress    string `json:"to_address"`
	Amount       int64  `json:"amount"`
}

// TransferAssetContract represents an asset transfer transaction
type TransferAssetContract struct {
	AssetName    string `json:"asset_name"`
	OwnerAddress string `json:"owner_address"`
	ToAddress    string `json:"to_address"`
	Amount       int64  `json:"amount"`
}

// Balance contracts
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

// FreezeBalanceContract represents a balance freezing transaction (V1)
type FreezeBalanceContract struct {
	OwnerAddress  string `json:"owner_address"`
	Resource      string `json:"resource"`
	FrozenBalance int64  `json:"frozen_balance"`
	ExpireTime    int64  `json:"expire_time"`
}

// UnfreezeBalanceContract represents a balance unfreezing transaction (V1)
type UnfreezeBalanceContract struct {
	OwnerAddress string `json:"owner_address"`
	Resource     string `json:"resource"`
}

// WithdrawBalanceContract represents a balance withdrawal transaction
type WithdrawBalanceContract struct {
	OwnerAddress string `json:"owner_address"`
}

// UnfreezeBalanceV2Contract represents a balance unfreezing transaction (V2)
type UnfreezeBalanceV2Contract struct {
	OwnerAddress    string `json:"owner_address"`
	UnfreezeBalance int64  `json:"unfreeze_balance"`
	Resource        string `json:"resource"`
}

// WithdrawExpireUnfreezeContract represents a withdrawal of expired unfreeze balance
type WithdrawExpireUnfreezeContract struct {
	OwnerAddress string `json:"owner_address"`
}

// CancelAllUnfreezeV2Contract represents cancel all unfreeze V2 requests
type CancelAllUnfreezeV2Contract struct {
	OwnerAddress string `json:"owner_address"`
}

// Smart contracts
// CreateSmartContract represents a smart contract creation transaction
type CreateSmartContract struct {
	OwnerAddress string              `json:"owner_address"`
	NewContract  *core.SmartContract `json:"new_contract,omitempty"`
}

// GetContract represents a get contract transaction
type GetContract struct {
	OwnerAddress    string `json:"owner_address"`
	ContractAddress string `json:"contract_address"`
}

// UpdateSettingContract represents a contract setting update transaction
type UpdateSettingContract struct {
	OwnerAddress               string `json:"owner_address"`
	ContractAddress            string `json:"contract_address"`
	ConsumeUserResourcePercent int64  `json:"consume_user_resource_percent"`
}

// UpdateEnergyLimitContract represents an energy limit update transaction
type UpdateEnergyLimitContract struct {
	OwnerAddress      string `json:"owner_address"`
	ContractAddress   string `json:"contract_address"`
	OriginEnergyLimit int64  `json:"origin_energy_limit"`
}

// ClearABIContract represents a clear ABI transaction
type ClearABIContract struct {
	OwnerAddress    string `json:"owner_address"`
	ContractAddress string `json:"contract_address"`
}

// Governance contracts
// VoteAssetContract represents a vote asset transaction
type VoteAssetContract struct {
	OwnerAddress string      `json:"owner_address"`
	Support      bool        `json:"support"`
	Votes        []VoteAsset `json:"votes"`
}

type VoteAsset struct {
	Support   bool   `json:"support"`
	AssetName string `json:"asset_name"`
	VoteCount int64  `json:"vote_count"`
}

// VoteWitnessContract represents a vote witness transaction
type VoteWitnessContract struct {
	OwnerAddress string        `json:"owner_address"`
	Support      bool          `json:"support"`
	Votes        []VoteWitness `json:"votes"`
}

type VoteWitness struct {
	VoteAddress string `json:"vote_address"`
	VoteCount   int64  `json:"vote_count"`
}

// WitnessCreateContract represents a witness creation transaction
type WitnessCreateContract struct {
	OwnerAddress string `json:"owner_address"`
	Url          string `json:"url"`
}

// WitnessUpdateContract represents a witness update transaction
type WitnessUpdateContract struct {
	OwnerAddress string `json:"owner_address"`
	UpdateUrl    string `json:"update_url"`
}

// ProposalCreateContract represents a proposal creation transaction
type ProposalCreateContract struct {
	OwnerAddress string          `json:"owner_address"`
	Parameters   map[int64]int64 `json:"parameters"`
}

// ProposalApproveContract represents a proposal approval transaction
type ProposalApproveContract struct {
	OwnerAddress string `json:"owner_address"`
	ProposalID   int64  `json:"proposal_id"`
	IsApprove    bool   `json:"is_approve"`
}

// ProposalDeleteContract represents a proposal deletion transaction
type ProposalDeleteContract struct {
	OwnerAddress string `json:"owner_address"`
	ProposalID   int64  `json:"proposal_id"`
}

// Exchange contracts
// ExchangeCreateContract represents an exchange creation transaction
type ExchangeCreateContract struct {
	OwnerAddress       string `json:"owner_address"`
	FirstTokenId       string `json:"first_token_id"`
	FirstTokenBalance  int64  `json:"first_token_balance"`
	SecondTokenId      string `json:"second_token_id"`
	SecondTokenBalance int64  `json:"second_token_balance"`
}

// ExchangeInjectContract represents an exchange injection transaction
type ExchangeInjectContract struct {
	OwnerAddress string `json:"owner_address"`
	ExchangeId   int64  `json:"exchange_id"`
	TokenId      string `json:"token_id"`
	Quant        int64  `json:"quant"`
}

// ExchangeWithdrawContract represents an exchange withdrawal transaction
type ExchangeWithdrawContract struct {
	OwnerAddress string `json:"owner_address"`
	ExchangeId   int64  `json:"exchange_id"`
	TokenId      string `json:"token_id"`
	Quant        int64  `json:"quant"`
}

// ExchangeTransactionContract represents an exchange transaction
type ExchangeTransactionContract struct {
	OwnerAddress string `json:"owner_address"`
	ExchangeId   int64  `json:"exchange_id"`
	Symbol       string `json:"symbol"`
	Quant        int64  `json:"quant"`
	Expected     int64  `json:"expected"`
}

// Market contracts
// MarketSellAssetContract represents a market sell asset transaction
type MarketSellAssetContract struct {
	OwnerAddress      string `json:"owner_address"`
	SellTokenId       string `json:"sell_token_id"`
	SellTokenQuantity int64  `json:"sell_token_quantity"`
	BuyTokenId        string `json:"buy_token_id"`
	BuyTokenQuantity  int64  `json:"buy_token_quantity"`
	OrderId           int64  `json:"order_id"`
}

// MarketCancelOrderContract represents a market cancel order transaction
type MarketCancelOrderContract struct {
	OwnerAddress string `json:"owner_address"`
	OrderId      string `json:"order_id"`
}

// Other contracts
// CustomContract represents a custom contract transaction
type CustomContract struct {
	Data string `json:"data"`
}

// UpdateBrokerageContract represents an update brokerage transaction
type UpdateBrokerageContract struct {
	OwnerAddress string `json:"owner_address"`
	Brokerage    int32  `json:"brokerage"`
}

// ShieldedTransferContract represents a shielded transfer transaction
type ShieldedTransferContract struct {
	FromAmount      int64                 `json:"from_amount"`
	ShieldedSpends  []ShieldedSpendNote   `json:"shielded_spends"`
	ShieldedOutputs []ShieldedReceiveNote `json:"shielded_outputs"`
	ToAmount        int64                 `json:"to_amount"`
	ToAddress       string                `json:"to_address"`
}

type ShieldedSpendNote struct {
	Alpha          string `json:"alpha"`
	AssetAmount    int64  `json:"asset_amount"`
	Note           string `json:"note"`
	Path           []int  `json:"path"`
	Root           string `json:"root"`
	Nullifier      string `json:"nullifier"`
	PerSig         string `json:"per_sig"`
	SpendAuthority string `json:"spend_authority"`
}

type ShieldedReceiveNote struct {
	Note            string `json:"note"`
	NoteCommitment  string `json:"note_commitment"`
	PAK             string `json:"pak"`
	ShieldedAddress string `json:"shielded_address"`
}
