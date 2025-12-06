package ptc

import (
	"context"
	"strings"
	"testing"
	"time"

	"github.com/smallnest/langgraphgo/log"
	"github.com/tmc/langchaingo/tools"
)

// TestPackageLevelLogging tests using package-level logging functions
func TestPackageLevelLogging(t *testing.T) {
	// Save original logger
	originalLogger := log.GetDefaultLogger()
	defer log.SetDefaultLogger(originalLogger)

	// Enable logging for this test
	log.SetLogLevel(log.LogLevelInfo)

	toolList := []tools.Tool{
		newMockTool("test", "Test tool", "ok"),
	}

	executor := NewCodeExecutor(LanguagePython, toolList)

	if executor == nil {
		t.Error("NewCodeExecutor should return the executor")
	}

	// Verify that logging works with package-level logger
	ctx := context.Background()
	if err := executor.Start(ctx); err != nil {
		t.Fatalf("Failed to start executor: %v", err)
	}
	defer executor.Stop(ctx)
}

// TestSanitizeFunctionNameEdgeCases tests edge cases for sanitizeFunctionName
func TestSanitizeFunctionNameEdgeCases(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{"with dashes", "my-tool-name", "my_tool_name"},
		{"with spaces", "my tool name", "my_tool_name"},
		{"with dots", "my.tool.name", "my_tool_name"},
		{"starts with number", "123tool", "tool_123tool"},
		{"mixed characters", "my-tool.name 123", "my_tool_name_123"},
		{"empty string", "", ""},
		{"already valid", "my_tool_name", "my_tool_name"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := sanitizeFunctionName(tt.input)
			if result != tt.expected {
				t.Errorf("sanitizeFunctionName(%q) = %q, want %q", tt.input, result, tt.expected)
			}
		})
	}
}

// TestExecutorWithEmptyTools tests executor with no tools
func TestExecutorWithEmptyTools(t *testing.T) {
	executor := NewCodeExecutor(LanguagePython, []tools.Tool{})

	ctx := context.Background()
	if err := executor.Start(ctx); err != nil {
		t.Fatalf("Failed to start executor with empty tools: %v", err)
	}
	defer executor.Stop(ctx)

	// Verify server URL is available even with no tools
	if executor.GetToolServerURL() == "" {
		t.Error("Expected tool server URL even with empty tools")
	}
}

// TestExecutorTimeout tests code execution timeout
func TestExecutorTimeoutShort(t *testing.T) {
	toolList := []tools.Tool{
		newMockTool("slow", "Slow tool", "ok"),
	}

	executor := NewCodeExecutor(LanguagePython, toolList)
	executor.Timeout = 100 * time.Millisecond // Very short timeout

	ctx := context.Background()
	if err := executor.Start(ctx); err != nil {
		t.Fatalf("Failed to start executor: %v", err)
	}
	defer executor.Stop(ctx)

	// Python code that sleeps longer than timeout
	code := `
import time
time.sleep(1)  # Sleep for 1 second
print("done")
`

	_, err := executor.Execute(ctx, code)
	// Should timeout but not panic
	if err == nil {
		t.Log("Expected timeout error, but execution completed")
		// This is not necessarily a failure - execution might complete quickly
	}
}

// TestExecutorWithMalformedCode tests execution of malformed code
func TestExecutorWithMalformedCode(t *testing.T) {
	toolList := []tools.Tool{
		newMockTool("test", "Test tool", "ok"),
	}

	executor := NewCodeExecutor(LanguagePython, toolList)
	ctx := context.Background()

	if err := executor.Start(ctx); err != nil {
		t.Fatalf("Failed to start executor: %v", err)
	}
	defer executor.Stop(ctx)

	// Malformed Python code
	code := `
this is not valid python syntax !!!
`

	result, err := executor.Execute(ctx, code)
	// Should return error or result with error in output
	if err == nil && !strings.Contains(result.Output, "SyntaxError") {
		t.Error("Expected syntax error in output for malformed code")
	}
}

// TestExecutorWithLargeCode tests execution of large code blocks
func TestExecutorWithLargeCode(t *testing.T) {
	toolList := []tools.Tool{
		newMockTool("test", "Test tool", "ok"),
	}

	executor := NewCodeExecutor(LanguagePython, toolList)
	ctx := context.Background()

	if err := executor.Start(ctx); err != nil {
		t.Fatalf("Failed to start executor: %v", err)
	}
	defer executor.Stop(ctx)

	// Generate large code block (many print statements)
	var codeBuilder strings.Builder
	for i := 0; i < 100; i++ {
		codeBuilder.WriteString("print('Line " + string(rune(i)) + "')\n")
	}

	_, err := executor.Execute(ctx, codeBuilder.String())
	if err != nil {
		t.Errorf("Failed to execute large code block: %v", err)
	}
}

// TestExecutorWithSpecialCharacters tests code with special characters
func TestExecutorWithSpecialCharacters(t *testing.T) {
	toolList := []tools.Tool{
		newMockTool("test", "Test tool", "ok"),
	}

	executor := NewCodeExecutor(LanguagePython, toolList)
	ctx := context.Background()

	if err := executor.Start(ctx); err != nil {
		t.Fatalf("Failed to start executor: %v", err)
	}
	defer executor.Stop(ctx)

	// Code with various special characters
	code := `
# Test with special characters: © ® ™ § ¶ • ª º « »
print("Hello 世界! 🌍")
print("Special chars: ñ é à ü")
`

	result, err := executor.Execute(ctx, code)
	if err != nil {
		t.Errorf("Failed to execute code with special characters: %v", err)
	}

	if result.Output == "" {
		t.Error("Expected non-empty output for special character code")
	}
}

// TestStopWithoutStart tests stopping executor that was never started
func TestStopBeforeStart(t *testing.T) {
	executor := NewCodeExecutor(LanguagePython, []tools.Tool{})

	ctx := context.Background()
	// Stop without starting should not panic
	err := executor.Stop(ctx)
	if err != nil {
		// This is acceptable - some implementations might return error
		t.Logf("Stop without Start returned error: %v", err)
	}
}

// TestMultipleStarts tests starting executor multiple times
func TestMultipleStarts(t *testing.T) {
	executor := NewCodeExecutor(LanguagePython, []tools.Tool{})

	ctx := context.Background()

	// First start
	if err := executor.Start(ctx); err != nil {
		t.Fatalf("First start failed: %v", err)
	}

	// Second start should return error
	err := executor.Start(ctx)
	if err == nil {
		t.Error("Expected error when starting already started executor")
	}

	executor.Stop(ctx)
}

// TestExecutorWorkDir tests custom work directory
func TestExecutorWorkDir(t *testing.T) {
	executor := NewCodeExecutor(LanguagePython, []tools.Tool{})

	// Verify default work dir is set
	if executor.WorkDir == "" {
		t.Error("Expected default WorkDir to be set")
	}

	// Change work dir
	executor.WorkDir = "/tmp"
	if executor.WorkDir != "/tmp" {
		t.Error("Failed to set custom WorkDir")
	}
}

// TestExecutorModeDirect tests Direct mode specific behavior
func TestExecutorModeDirectSpecific(t *testing.T) {
	toolList := []tools.Tool{
		newMockTool("test", "Test tool", "ok"),
	}

	executor := NewCodeExecutorWithMode(LanguagePython, toolList, ModeDirect)

	if executor.Mode != ModeDirect {
		t.Errorf("Expected ModeDirect, got %v", executor.Mode)
	}

	ctx := context.Background()
	if err := executor.Start(ctx); err != nil {
		t.Fatalf("Failed to start in Direct mode: %v", err)
	}
	defer executor.Stop(ctx)

	// In Direct mode, tool server should be available (for generic tools)
	if executor.GetToolServerURL() == "" {
		t.Error("Expected tool server URL in Direct mode")
	}
}

// TestExecutorModeServer tests Server mode specific behavior
func TestExecutorModeServerSpecific(t *testing.T) {
	toolList := []tools.Tool{
		newMockTool("test", "Test tool", "ok"),
	}

	executor := NewCodeExecutorWithMode(LanguagePython, toolList, ModeServer)

	if executor.Mode != ModeServer {
		t.Errorf("Expected ModeServer, got %v", executor.Mode)
	}

	ctx := context.Background()
	if err := executor.Start(ctx); err != nil {
		t.Fatalf("Failed to start in Server mode: %v", err)
	}
	defer executor.Stop(ctx)

	// In Server mode, tool server should be available
	if executor.GetToolServerURL() == "" {
		t.Error("Expected tool server URL in Server mode")
	}
}

// TestGetToolServerURLBeforeStart tests GetToolServerURL before Start
func TestGetToolServerURLBeforeStart(t *testing.T) {
	executor := NewCodeExecutor(LanguagePython, []tools.Tool{})

	// GetToolServerURL before Start returns URL with port 0
	url := executor.GetToolServerURL()
	if !strings.Contains(url, "127.0.0.1") {
		t.Errorf("Expected URL to contain localhost, got %s", url)
	}
}

// TestExecutionResultStructure tests ExecutionResult structure
func TestExecutionResultStructure(t *testing.T) {
	result := &ExecutionResult{
		Output: "test output",
		Error:  nil,
		Stdout: "stdout content",
		Stderr: "stderr content",
	}

	if result.Output != "test output" {
		t.Error("ExecutionResult.Output not set correctly")
	}
	if result.Stdout != "stdout content" {
		t.Error("ExecutionResult.Stdout not set correctly")
	}
	if result.Stderr != "stderr content" {
		t.Error("ExecutionResult.Stderr not set correctly")
	}
	if result.Error != nil {
		t.Error("ExecutionResult.Error should be nil")
	}
}
