package troneventssdk

import (
	"log"
	"os"
)

// Logger is an interface that allows users to provide their own logging implementation.
type Logger interface {
	Infof(format string, args ...interface{})
	Errorf(format string, args ...interface{})
	Debugf(format string, args ...interface{})
}

// DefaultLogger is a simple implementation of the Logger interface using the standard log package.
type DefaultLogger struct {
	logger *log.Logger
}

// NewDefaultLogger creates a new instance of DefaultLogger.
func NewDefaultLogger() *DefaultLogger {
	return &DefaultLogger{
		logger: log.New(os.Stdout, "[troneventssdk] ", log.LstdFlags),
	}
}

// Infof logs an informational message.
func (l *DefaultLogger) Infof(format string, args ...interface{}) {
	l.logger.Printf("[INFO] "+format, args...)
}

// Errorf logs an error message.
func (l *DefaultLogger) Errorf(format string, args ...interface{}) {
	l.logger.Printf("[ERROR] "+format, args...)
}

// Debugf logs a debug message.
func (l *DefaultLogger) Debugf(format string, args ...interface{}) {
	l.logger.Printf("[DEBUG] "+format, args...)
}