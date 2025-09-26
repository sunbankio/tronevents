package logging

import (
	"log"
	"os"
	"strings"
)

// LogLevel represents the logging level.
type LogLevel int

const (
	ERROR LogLevel = iota
	INFO
	DEBUG
)

// Logger provides logging functionality with configurable log level.
type Logger struct {
	logger   *log.Logger
	logLevel LogLevel
}

// NewLogger creates a new structured logger with the specified log level.
func NewLogger(logLevel string) *Logger {
	level := parseLogLevel(logLevel)
	return &Logger{
		logger:   log.New(os.Stdout, "", log.LstdFlags),
		logLevel: level,
	}
}

// parseLogLevel converts string log level to LogLevel enum.
func parseLogLevel(level string) LogLevel {
	switch strings.ToLower(level) {
	case "debug":
		return DEBUG
	case "info":
		return INFO
	case "error":
		return ERROR
	default:
		return INFO // default to info level
	}
}

// Debug logs a message at debug level if debug is enabled.
func (l *Logger) Debug(v ...interface{}) {
	if l.logLevel >= DEBUG {
		l.logger.Print(append([]interface{}{"DEBUG:"}, v...)...)
	}
}

// Debugf logs a formatted message at debug level if debug is enabled.
func (l *Logger) Debugf(format string, v ...interface{}) {
	if l.logLevel >= DEBUG {
		l.logger.Printf("DEBUG: "+format, v...)
	}
}

// Info logs a message at info level.
func (l *Logger) Info(v ...interface{}) {
	if l.logLevel >= INFO {
		l.logger.Print(v...)
	}
}

// Infof logs a formatted message at info level.
func (l *Logger) Infof(format string, v ...interface{}) {
	if l.logLevel >= INFO {
		l.logger.Printf(format, v...)
	}
}

// Error logs a message at error level.
func (l *Logger) Error(v ...interface{}) {
	if l.logLevel >= ERROR {
		l.logger.Print(v...)
	}
}

// Errorf logs a formatted message at error level.
func (l *Logger) Errorf(format string, v ...interface{}) {
	if l.logLevel >= ERROR {
		l.logger.Printf(format, v...)
	}
}

// Print is a wrapper that always prints (used for non-debug logs).
func (l *Logger) Print(v ...interface{}) {
	l.logger.Print(v...)
}

// Printf is a wrapper that always prints (used for non-debug logs).
func (l *Logger) Printf(format string, v ...interface{}) {
	l.logger.Printf(format, v...)
}

// Println is a wrapper that always prints with newline (used for compatibility).
func (l *Logger) Println(v ...interface{}) {
	l.logger.Println(v...)
}

// Fatal is a wrapper that prints and exits (used for compatibility).
func (l *Logger) Fatal(v ...interface{}) {
	l.logger.Fatal(v...)
}

// Fatalf is a wrapper that prints formatted and exits (used for compatibility).
func (l *Logger) Fatalf(format string, v ...interface{}) {
	l.logger.Fatalf(format, v...)
}
