package troneventssdk

import (
	"fmt"
	"os"
	"strings"
)

// FileStateStore is a simple implementation of the StateStore interface that uses a file to persist the last processed ID.
type FileStateStore struct {
	filename string
}

// SaveLastProcessedID saves the last processed ID to the file.
func (s *FileStateStore) SaveLastProcessedID(consumerGroup, consumerName, lastID string) error {
	content := fmt.Sprintf("%s:%s:%s", consumerGroup, consumerName, lastID)
	return os.WriteFile(s.filename, []byte(content), 0644)
}

// GetLastProcessedID retrieves the last processed ID from the file.
// It returns an empty string if the file doesn't exist or if the format is invalid.
func (s *FileStateStore) GetLastProcessedID(consumerGroup, consumerName string) (string, error) {
	content, err := os.ReadFile(s.filename)
	if err != nil {
		// If the file doesn't exist, return an empty string
		if os.IsNotExist(err) {
			return "", nil
		}
		return "", err
	}

	parts := strings.Split(string(content), ":")
	if len(parts) != 3 {
		return "", fmt.Errorf("invalid checkpoint format in file %s", s.filename)
	}

	storedGroup := parts[0]
	storedConsumer := parts[1]
	lastID := parts[2]

	if storedGroup != consumerGroup || storedConsumer != consumerName {
		return "", fmt.Errorf("checkpoint in file %s does not match current consumer group and name", s.filename)
	}

	return lastID, nil
}