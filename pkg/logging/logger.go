package logging

import (
	"log"
	"os"
)

// NewLogger creates a new structured logger.
func NewLogger() *log.Logger {
	return log.New(os.Stdout, "", log.LstdFlags)
}
