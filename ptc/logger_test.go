package ptc

import (
	"bytes"
	"context"
	"strings"
	"testing"

	"github.com/smallnest/langgraphgo/log"
	"github.com/tmc/langchaingo/tools"
)

// TestLogger tests the logging functionality
func TestLogger(t *testing.T) {
	// Create a buffer to capture log output
	var buf bytes.Buffer
	logger := log.NewCustomLogger(&buf, log.LogLevelDebug)

	toolList := []tools.Tool{
		newMockTool("test", "Test tool", "ok"),
	}

	executor := NewCodeExecutor(LanguagePython, toolList)
	executor.SetLogger(logger)

	ctx := context.Background()
	if err := executor.Start(ctx); err != nil {
		t.Fatalf("Failed to start executor: %v", err)
	}
	defer executor.Stop(ctx)

	code := `
result = test("hello")
print(result)
`

	_, err := executor.Execute(ctx, code)
	if err != nil {
		t.Fatalf("Failed to execute code: %v", err)
	}

	// Check that logs were written
	logOutput := buf.String()
	if logOutput == "" {
		t.Error("Expected log output, got none")
	}

	// Check for expected log entries
	expectedLogs := []string{
		"Tool server starting on port",
		"Tool server started successfully",
		"Executing code in",
		"Code execution succeeded",
	}

	for _, expected := range expectedLogs {
		if !strings.Contains(logOutput, expected) {
			t.Errorf("Expected log to contain '%s', got: %s", expected, logOutput)
		}
	}
}

// TestLogLevels tests different log levels
func TestLogLevels(t *testing.T) {
	tests := []struct {
		name          string
		level         log.LogLevel
		shouldContain []string
		shouldNotContain []string
	}{
		{
			name:          "Debug level",
			level:         log.LogLevelDebug,
			shouldContain: []string{"[DEBUG]", "[INFO]", "[WARN]", "[ERROR]"},
		},
		{
			name:          "Info level",
			level:         log.LogLevelInfo,
			shouldContain: []string{"[INFO]", "[WARN]", "[ERROR]"},
			shouldNotContain: []string{"[DEBUG]"},
		},
		{
			name:          "Error level",
			level:         log.LogLevelError,
			shouldContain: []string{"[ERROR]"},
			shouldNotContain: []string{"[DEBUG]", "[INFO]", "[WARN]"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var buf bytes.Buffer
			logger := log.NewCustomLogger(&buf, tt.level)

			// Log messages at all levels
			logger.Debug("debug message")
			logger.Info("info message")
			logger.Warn("warn message")
			logger.Error("error message")

			output := buf.String()

			for _, expected := range tt.shouldContain {
				if !strings.Contains(output, expected) {
					t.Errorf("Expected output to contain '%s'", expected)
				}
			}

			for _, unexpected := range tt.shouldNotContain {
				if strings.Contains(output, unexpected) {
					t.Errorf("Expected output NOT to contain '%s'", unexpected)
				}
			}
		})
	}
}

// TestNoOpLogger tests that NoOpLogger doesn't produce any output
func TestNoOpLogger(t *testing.T) {
	logger := &log.NoOpLogger{}

	// These should not panic or produce output
	logger.Debug("test")
	logger.Info("test")
	logger.Warn("test")
	logger.Error("test")
}

// TestLogLevelString tests LogLevel.String()
func TestLogLevelString(t *testing.T) {
	tests := []struct {
		level    log.LogLevel
		expected string
	}{
		{log.LogLevelDebug, "DEBUG"},
		{log.LogLevelInfo, "INFO"},
		{log.LogLevelWarn, "WARN"},
		{log.LogLevelError, "ERROR"},
		{log.LogLevelNone, "NONE"},
		{log.LogLevel(999), "UNKNOWN(999)"},
	}

	for _, tt := range tests {
		if got := tt.level.String(); got != tt.expected {
			t.Errorf("LogLevel(%d).String() = %s, want %s", tt.level, got, tt.expected)
		}
	}
}
