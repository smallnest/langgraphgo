package log

import (
	"fmt"
	"io"
	"log"
	"os"
)

// LogLevel represents logging severity
type LogLevel int

const (
	// LogLevelDebug for detailed debugging information
	LogLevelDebug LogLevel = iota
	// LogLevelInfo for general informational messages
	LogLevelInfo
	// LogLevelWarn for warning messages
	LogLevelWarn
	// LogLevelError for error messages
	LogLevelError
	// LogLevelNone disables all logging
	LogLevelNone
)

// Logger interface for PTC logging
type Logger interface {
	Debug(format string, v ...interface{})
	Info(format string, v ...interface{})
	Warn(format string, v ...interface{})
	Error(format string, v ...interface{})
}

// DefaultLogger implements Logger using Go's standard log package
type DefaultLogger struct {
	logger *log.Logger
	level  LogLevel
}

// NewDefaultLogger creates a new default logger
func NewDefaultLogger(level LogLevel) *DefaultLogger {
	return &DefaultLogger{
		logger: log.New(os.Stderr, "[PTC] ", log.LstdFlags),
		level:  level,
	}
}

// NewCustomLogger creates a logger with custom output
func NewCustomLogger(out io.Writer, level LogLevel) *DefaultLogger {
	return &DefaultLogger{
		logger: log.New(out, "[PTC] ", log.LstdFlags),
		level:  level,
	}
}

// Debug logs debug messages
func (l *DefaultLogger) Debug(format string, v ...interface{}) {
	if l.level <= LogLevelDebug {
		l.logger.Printf("[DEBUG] "+format, v...)
	}
}

// Info logs informational messages
func (l *DefaultLogger) Info(format string, v ...interface{}) {
	if l.level <= LogLevelInfo {
		l.logger.Printf("[INFO] "+format, v...)
	}
}

// Warn logs warning messages
func (l *DefaultLogger) Warn(format string, v ...interface{}) {
	if l.level <= LogLevelWarn {
		l.logger.Printf("[WARN] "+format, v...)
	}
}

// Error logs error messages
func (l *DefaultLogger) Error(format string, v ...interface{}) {
	if l.level <= LogLevelError {
		l.logger.Printf("[ERROR] "+format, v...)
	}
}

// NoOpLogger is a logger that doesn't log anything
type NoOpLogger struct{}

// Debug does nothing
func (l *NoOpLogger) Debug(format string, v ...interface{}) {}

// Info does nothing
func (l *NoOpLogger) Info(format string, v ...interface{}) {}

// Warn does nothing
func (l *NoOpLogger) Warn(format string, v ...interface{}) {}

// Error does nothing
func (l *NoOpLogger) Error(format string, v ...interface{}) {}

// String returns the string representation of LogLevel
func (l LogLevel) String() string {
	switch l {
	case LogLevelDebug:
		return "DEBUG"
	case LogLevelInfo:
		return "INFO"
	case LogLevelWarn:
		return "WARN"
	case LogLevelError:
		return "ERROR"
	case LogLevelNone:
		return "NONE"
	default:
		return fmt.Sprintf("UNKNOWN(%d)", l)
	}
}

// Package-level logger (default is NoOpLogger for zero overhead)
var defaultLogger Logger = &NoOpLogger{}

// SetDefaultLogger sets the package-level logger
// This allows users to enable logging globally without passing logger objects around
func SetDefaultLogger(logger Logger) {
	defaultLogger = logger
}

// GetDefaultLogger returns the current package-level logger
func GetDefaultLogger() Logger {
	return defaultLogger
}

// SetLogLevel creates and sets a default logger with the specified log level
// This is a convenience function for quick logging setup
func SetLogLevel(level LogLevel) {
	defaultLogger = NewDefaultLogger(level)
}

// Debug logs a debug message using the package-level logger
func Debug(format string, v ...interface{}) {
	defaultLogger.Debug(format, v...)
}

// Info logs an informational message using the package-level logger
func Info(format string, v ...interface{}) {
	defaultLogger.Info(format, v...)
}

// Warn logs a warning message using the package-level logger
func Warn(format string, v ...interface{}) {
	defaultLogger.Warn(format, v...)
}

// Error logs an error message using the package-level logger
func Error(format string, v ...interface{}) {
	defaultLogger.Error(format, v...)
}
