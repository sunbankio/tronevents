package storage

import (
	"encoding/json"
	"io/ioutil"
	"os"
	"sync"
)

// BlockStorage handles the persistence of the last processed block number.
type BlockStorage struct {
	mu       sync.Mutex
	filePath string
}

// NewBlockStorage creates a new BlockStorage.
func NewBlockStorage(filePath string) *BlockStorage {
	return &BlockStorage{filePath: filePath}
}

// Save saves the last processed block number to a file.
func (s *BlockStorage) Save(blockNumber int64) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	data, err := json.Marshal(blockNumber)
	if err != nil {
		return err
	}

	return ioutil.WriteFile(s.filePath, data, 0644)
}

// Load loads the last processed block number from a file.
func (s *BlockStorage) Load() (int64, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	if _, err := os.Stat(s.filePath); os.IsNotExist(err) {
		return 0, nil
	}

	data, err := ioutil.ReadFile(s.filePath)
	if err != nil {
		return 0, err
	}

	var blockNumber int64
	if err := json.Unmarshal(data, &blockNumber); err != nil {
		return 0, err
	}

	return blockNumber, nil
}
