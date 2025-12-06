package log

import (
	"bytes"
	"strings"
	"testing"
)

// TestDefaultLogger tests the default logger functionality
func TestDefaultLogger(t *testing.T) {
	var buf bytes.Buffer
	logger := NewCustomLogger(&buf, LogLevelDebug)

	logger.Debug("debug message")
	logger.Info("info message")
	logger.Warn("warn message")
	logger.Error("error message")

	output := buf.String()

	expectedMessages := []string{
		"[DEBUG] debug message",
		"[INFO] info message",
		"[WARN] warn message",
		"[ERROR] error message",
	}

	for _, expected := range expectedMessages {
		if !strings.Contains(output, expected) {
			t.Errorf("Expected output to contain '%s', got: %s", expected, output)
		}
	}
}

// TestLogLevels tests different log levels
func TestLogLevels(t *testing.T) {
	tests := []struct {
		name             string
		level            LogLevel
		shouldContain    []string
		shouldNotContain []string
	}{
		{
			name:          "Debug level",
			level:         LogLevelDebug,
			shouldContain: []string{"[DEBUG]", "[INFO]", "[WARN]", "[ERROR]"},
		},
		{
			name:             "Info level",
			level:            LogLevelInfo,
			shouldContain:    []string{"[INFO]", "[WARN]", "[ERROR]"},
			shouldNotContain: []string{"[DEBUG]"},
		},
		{
			name:             "Warn level",
			level:            LogLevelWarn,
			shouldContain:    []string{"[WARN]", "[ERROR]"},
			shouldNotContain: []string{"[DEBUG]", "[INFO]"},
		},
		{
			name:             "Error level",
			level:            LogLevelError,
			shouldContain:    []string{"[ERROR]"},
			shouldNotContain: []string{"[DEBUG]", "[INFO]", "[WARN]"},
		},
		{
			name:             "None level",
			level:            LogLevelNone,
			shouldNotContain: []string{"[DEBUG]", "[INFO]", "[WARN]", "[ERROR]"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var buf bytes.Buffer
			logger := NewCustomLogger(&buf, tt.level)

			logger.Debug("debug")
			logger.Info("info")
			logger.Warn("warn")
			logger.Error("error")

			output := buf.String()

			for _, expected := range tt.shouldContain {
				if !strings.Contains(output, expected) {
					t.Errorf("Expected output to contain '%s'", expected)
				}
			}

			for _, unexpected := range tt.shouldNotContain {
				if strings.Contains(output, unexpected) {
					t.Errorf("Expected output NOT to contain '%s', got: %s", unexpected, output)
				}
			}
		})
	}
}

// TestNoOpLogger tests that NoOpLogger doesn't produce any output
func TestNoOpLogger(t *testing.T) {
	logger := &NoOpLogger{}

	// These should not panic
	logger.Debug("test")
	logger.Info("test")
	logger.Warn("test")
	logger.Error("test")
}

// TestLogLevelString tests LogLevel.String()
func TestLogLevelString(t *testing.T) {
	tests := []struct {
		level    LogLevel
		expected string
	}{
		{LogLevelDebug, "DEBUG"},
		{LogLevelInfo, "INFO"},
		{LogLevelWarn, "WARN"},
		{LogLevelError, "ERROR"},
		{LogLevelNone, "NONE"},
		{LogLevel(999), "UNKNOWN(999)"},
	}

	for _, tt := range tests {
		if got := tt.level.String(); got != tt.expected {
			t.Errorf("LogLevel(%d).String() = %s, want %s", tt.level, got, tt.expected)
		}
	}
}

// TestNewDefaultLogger tests creating a default logger
func TestNewDefaultLogger(t *testing.T) {
	logger := NewDefaultLogger(LogLevelInfo)
	if logger == nil {
		t.Error("NewDefaultLogger returned nil")
	}
}
