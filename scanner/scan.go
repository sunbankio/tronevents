package scanner

import (
	"context"
	"encoding/hex"
	"time"

	"github.com/kslamph/tronlib/pb/api"
	"github.com/kslamph/tronlib/pb/core"
	"github.com/kslamph/tronlib/pkg/client"
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

func (s *Scanner) Scan(blockNumber int64) (int64, time.Time, []Transaction, error) {
	var block *api.BlockExtention

	if blockNumber > 0 {
		var err error
		block, err = s.getBlockByNumber(blockNumber)
		if err != nil {
			return 0, time.Time{}, nil, err
		}
	} else {
		var err error
		// Get the latest block
		block, err = s.tronclient.Network().GetNowBlock(s.ctx)
		if err != nil {
			return 0, time.Time{}, nil, err
		}
		blockNumber = block.BlockHeader.RawData.Number
	}
	blockTime := time.Unix(0, block.BlockHeader.RawData.Timestamp*int64(time.Millisecond))

	// Get transaction info - assuming it always exists for block transactions
	txInfoList, err := s.getTransactionInfoByNumber(blockNumber)
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

func (s *Scanner) getBlockByNumber(blockNumber int64) (*api.BlockExtention, error) {
	return s.tronclient.Network().GetBlockByNumber(s.ctx, blockNumber)
}

func (s *Scanner) getTransactionInfoByNumber(blockNumber int64) (*api.TransactionInfoList, error) {
	return s.tronclient.Network().GetTransactionInfoByBlockNum(s.ctx, blockNumber)
}
