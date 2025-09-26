package scanner

import (
	"context"
	"encoding/hex"
	"fmt"
	"time"

	"github.com/kslamph/tronlib/pb/api"
	"github.com/kslamph/tronlib/pb/core"
	"github.com/kslamph/tronlib/pkg/client"
)

type Scanner struct {
	tronclient *client.Client
}

func NewScanner(nodeAddress string, timeout int, poolSize int, maxPoolSize int) (*Scanner, error) {
	// Create the client with configurable timeout and pool settings
	tronclient, err := client.NewClient(nodeAddress,
		client.WithTimeout(time.Duration(timeout)*time.Second),
		client.WithPool(poolSize, maxPoolSize),
	)
	if err != nil {
		return nil, err
	}

	return &Scanner{
		tronclient: tronclient,
	}, nil
}

func (s *Scanner) Close() {
	s.tronclient.Close()
}

func (s *Scanner) Scan(ctx context.Context, blockNumber int64) (int64, time.Time, []Transaction, error) {
	var block *api.BlockExtention

	if blockNumber > 0 {
		var err error
		block, err = s.getBlockByNumber(ctx, blockNumber)
		if err != nil {
			return 0, time.Time{}, nil, err
		}
		// Check if block is nil (block doesn't exist)
		if block == nil {
			return 0, time.Time{}, nil, fmt.Errorf("block %d is nil", blockNumber)
		}
	} else {
		var err error
		// Get the latest block
		block, err = s.tronclient.Network().GetNowBlock(ctx)
		if err != nil {
			return 0, time.Time{}, nil, err
		}
		// Check if block is nil (shouldn't happen for latest block, but be safe)
		if block == nil {
			return 0, time.Time{}, nil, fmt.Errorf("current block is nil")
		}
		blockNumber = block.BlockHeader.RawData.Number
	}

	// Additional safety check for block header
	if block.BlockHeader == nil || block.BlockHeader.RawData == nil {
		return 0, time.Time{}, nil, fmt.Errorf("block %d does not exist", blockNumber)
	}

	blockTime := time.Unix(0, block.BlockHeader.RawData.Timestamp*int64(time.Millisecond))

	// Get transaction info - assuming it always exists for block transactions
	txInfoList, err := s.getTransactionInfoByNumber(ctx, blockNumber)
	if err != nil {
		return 0, time.Time{}, nil, err
	}

	// Create a map of transaction info by transaction ID for easy lookup
	txInfoMap := make(map[string]*core.TransactionInfo)
	for _, txInfo := range txInfoList.TransactionInfo {
		txID := hex.EncodeToString(txInfo.Id)
		txInfoMap[txID] = txInfo
	}

	// Process each transaction with enhanced data from txinfo
	transactions := make([]Transaction, 0, len(block.Transactions))
	for _, tx := range block.Transactions {
		txID := hex.EncodeToString(tx.Txid)

		// Look for corresponding transaction info and enhance the transaction
		if txInfo, exists := txInfoMap[txID]; exists {
			// Parse the transaction with the available info
			transaction := parseTransactionWithInfo(tx, txInfo)
			transactions = append(transactions, transaction)
		} else {
			// This should not happen if txinfo always exists, but handle gracefully
			transaction := parseTransaction(tx)
			transactions = append(transactions, transaction)
		}
	}

	return blockNumber, blockTime, transactions, nil
}

func (s *Scanner) getBlockByNumber(ctx context.Context, blockNumber int64) (*api.BlockExtention, error) {
	return s.tronclient.Network().GetBlockByNumber(ctx, blockNumber)
}

func (s *Scanner) getTransactionInfoByNumber(ctx context.Context, blockNumber int64) (*api.TransactionInfoList, error) {
	return s.tronclient.Network().GetTransactionInfoByBlockNum(ctx, blockNumber)
}

// GetTransactionsByBlock returns transactions for a given block number
func (s *Scanner) GetTransactionsByBlock(blockNum int64) ([]Transaction, error) {
	_, _, transactions, err := s.Scan(context.Background(), blockNum)
	return transactions, err
}
