package models

import (
	"encoding/json"
	"time"

	tronScanner "github.com/sunbankio/tronevents/pkg/scanner"
)

// SafeTime wraps time.Time to handle invalid values
type SafeTime struct {
	time.Time
}

// MarshalJSON customizes JSON marshaling to handle invalid times
func (t SafeTime) MarshalJSON() ([]byte, error) {
	if t.Time.IsZero() || t.Time.Year() < 0 || t.Time.Year() > 9999 {
		return json.Marshal(time.Time{})
	}
	return json.Marshal(t.Time)
}

// UnmarshalJSON customizes JSON unmarshaling
func (t *SafeTime) UnmarshalJSON(data []byte) error {
	var tm time.Time
	if err := json.Unmarshal(data, &tm); err != nil {
		return err
	}
	t.Time = tm
	return nil
}

// SafeTransaction wraps scanner.Transaction with safe time handling
type SafeTransaction struct {
	ID             string              `json:"id"`
	Contract       tronScanner.Contract `json:"contract"`
	Ret            tronScanner.RetInfo `json:"ret"`
	Timestamp      SafeTime            `json:"timestamp"`
	BlockNumber    int64               `json:"block_number,omitempty"`
	BlockTimestamp SafeTime            `json:"block_timestamp,omitempty"`
	Expiration     SafeTime            `json:"expiration,omitempty"`
	Receipt        tronScanner.Receipt `json:"receipt,omitempty"`
	Logs           []tronScanner.LogInfo `json:"logs,omitempty"`
	Signers        []string            `json:"signers,omitempty"` // All signers for the transaction
}

// ConvertTransaction converts a scanner.Transaction to a SafeTransaction
func ConvertTransaction(tx tronScanner.Transaction) SafeTransaction {
	return SafeTransaction{
		ID:             tx.ID,
		Contract:       tx.Contract,
		Ret:            tx.Ret,
		Timestamp:      SafeTime{tx.Timestamp},
		BlockNumber:    tx.BlockNumber,
		BlockTimestamp: SafeTime{tx.BlockTimestamp},
		Expiration:     SafeTime{tx.Expiration},
		Receipt:        tx.Receipt,
		Logs:           tx.Logs,
		Signers:        tx.Signers,
	}
}