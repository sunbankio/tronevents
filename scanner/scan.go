package scanner

import (
	"context"
	"fmt"
	"time"

	"github.com/kslamph/tronlib/pb/api"
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

func (s *Scanner) Scan(blockNumber int64) error {
	block, err := s.getBlockByNumber(blockNumber)
	if err != nil {
		return err
	}
	// fmt.Println(block)

	for _, tx := range block.Transactions {
		fmt.Println("----- Transaction -----")
		fmt.Printf("Transaction ID: %x\n", tx.Txid)
		fmt.Printf("Contract : %v\n", tx.Transaction)
		// Further processing can be done here
	}
	return nil
}

func (s *Scanner) ScanStructured(blockNumber int64) error {
	block, err := s.getBlockByNumber(blockNumber)
	if err != nil {
		return err
	}

	for _, tx := range block.Transactions {
		transaction := ParseTransaction(tx)
		PrintTransaction(transaction)
	}
	return nil
}

func (s *Scanner) getBlockByNumber(blockNumber int64) (*api.BlockExtention, error) {
	return s.tronclient.Network().GetBlockByNumber(s.ctx, blockNumber)
}
