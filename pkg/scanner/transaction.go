package scanner

import (
	"time"
)

// Transaction represents a parsed TRON transaction
type Transaction struct {
	ID             string    `json:"id"`
	Contract       *Contract `json:"contract,omitempty"`
	Ret            *RetInfo  `json:"ret,omitempty"`
	Timestamp      time.Time `json:"timestamp"`
	BlockNumber    int64     `json:"block_number,omitempty"`
	BlockTimestamp time.Time `json:"block_timestamp,omitempty"`
	Expiration     time.Time `json:"expiration,omitempty"`
	Receipt        *Receipt  `json:"receipt,omitempty"`
	Logs           []LogInfo `json:"logs,omitempty"`
	Signers        []string  `json:"signers,omitempty"` // All signers for the transaction
}

// RetInfo represents the return information of a transaction
type RetInfo struct {
	ContractRet string `json:"contractRet"`
}

// Receipt represents the receipt information of a transaction
type Receipt struct {
	EnergyUsage       int64 `json:"energy_usage,omitempty"`
	EnergyFee         int64 `json:"energy_fee,omitempty"`
	OriginEnergyUsage int64 `json:"origin_energy_usage,omitempty"`
	EnergyUsageTotal  int64 `json:"energy_usage_total,omitempty"`
	NetUsage          int64 `json:"net_usage,omitempty"`
	NetFee            int64 `json:"net_fee,omitempty"`
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
