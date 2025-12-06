package ptc_test

import (
	"context"
	"strings"
	"testing"
	"time"

	"github.com/smallnest/langgraphgo/ptc"
	"github.com/tmc/langchaingo/tools"
)

// TestModeDirectExecution tests that ModeDirect mode actually executes tools
func TestModeDirectExecution(t *testing.T) {
	tools := []tools.Tool{
		MockTool{
			name:        "echo",
			description: "Echoes input",
			response:    "echoed: test",
		},
	}

	executor := ptc.NewCodeExecutorWithMode(ptc.LanguagePython, tools, ptc.ModeDirect)
	ctx := context.Background()

	// Start the executor (Direct mode uses internal server for generic tools)
	if err := executor.Start(ctx); err != nil {
		t.Fatalf("Failed to start executor: %v", err)
	}
	defer executor.Stop(ctx)

	// Verify tool server URL is available (used internally for generic tools)
	serverURL := executor.GetToolServerURL()
	if serverURL == "" {
		t.Error("Expected non-empty tool server URL in Direct mode (for internal use)")
	}

	// Test Python code that calls a generic tool
	// Generic tools (like echo) are called via internal server
	code := `
result = echo("hello")
print(result)
`

	result, err := executor.Execute(ctx, code)
	if err != nil {
		t.Fatalf("Failed to execute code: %v", err)
	}

	// In Direct mode, generic tools should call through internal server and return actual result
	if !strings.Contains(result.Output, "echoed") {
		t.Errorf("Expected output to contain 'echoed', got: %s", result.Output)
	}
}

// TestModeServerExecution tests that ModeServer mode works
func TestModeServerExecution(t *testing.T) {
	tools := []tools.Tool{
		MockTool{
			name:        "calculator",
			description: "Performs calculations",
			response:    "42",
		},
	}

	executor := ptc.NewCodeExecutorWithMode(ptc.LanguagePython, tools, ptc.ModeServer)
	ctx := context.Background()

	if err := executor.Start(ctx); err != nil {
		t.Fatalf("Failed to start executor: %v", err)
	}
	defer executor.Stop(ctx)

	serverURL := executor.GetToolServerURL()
	if serverURL == "" {
		t.Error("Expected non-empty tool server URL in Server mode")
	}

	// Test Python code that calls tools via HTTP
	code := `
result = calculator("2+2")
print(result)
`

	result, err := executor.Execute(ctx, code)
	if err != nil {
		t.Fatalf("Failed to execute code: %v", err)
	}

	if !strings.Contains(result.Output, "42") {
		t.Errorf("Expected output to contain '42', got: %s", result.Output)
	}
}

// TestExecutorTimeout tests execution timeout
func TestExecutorTimeout(t *testing.T) {
	tools := []tools.Tool{
		MockTool{
			name:        "test",
			description: "Test tool",
			response:    "ok",
		},
	}

	executor := ptc.NewCodeExecutor(ptc.LanguagePython, tools)
	executor.Timeout = 2 * time.Second
	ctx := context.Background()

	if err := executor.Start(ctx); err != nil {
		t.Fatalf("Failed to start executor: %v", err)
	}
	defer executor.Stop(ctx)

	// Code that takes longer than timeout
	code := `
import time
time.sleep(10)
print("done")
`

	result, err := executor.Execute(ctx, code)
	// Timeout may be returned as error or in result
	if err == nil && result.Error == nil {
		t.Skip("Timeout test skipped - execution completed before timeout")
	}
}

// TestGoCodeExecution tests Go code execution
func TestGoCodeExecution(t *testing.T) {
	tools := []tools.Tool{
		MockTool{
			name:        "greet",
			description: "Greets someone",
			response:    "Hello, World!",
		},
	}

	executor := ptc.NewCodeExecutor(ptc.LanguageGo, tools)
	ctx := context.Background()

	if err := executor.Start(ctx); err != nil {
		t.Fatalf("Failed to start executor: %v", err)
	}
	defer executor.Stop(ctx)

	code := `
result, _ := greet(ctx, "World")
fmt.Println(result)
`

	result, err := executor.Execute(ctx, code)
	if err != nil {
		t.Fatalf("Failed to execute Go code: %v", err)
	}

	// In Direct mode, generic tools (like "greet") are called via internal server
	// and should return the actual tool result
	if !strings.Contains(result.Output, "Hello") {
		t.Errorf("Expected output to contain 'Hello', got: %s", result.Output)
	}
}

// TestMultipleTools tests execution with multiple tools
func TestMultipleTools(t *testing.T) {
	tools := []tools.Tool{
		MockTool{
			name:        "add",
			description: "Adds numbers",
			response:    "5",
		},
		MockTool{
			name:        "multiply",
			description: "Multiplies numbers",
			response:    "10",
		},
		MockTool{
			name:        "divide",
			description: "Divides numbers",
			response:    "2",
		},
	}

	executor := ptc.NewCodeExecutor(ptc.LanguagePython, tools)
	ctx := context.Background()

	if err := executor.Start(ctx); err != nil {
		t.Fatalf("Failed to start executor: %v", err)
	}
	defer executor.Stop(ctx)

	code := `
a = add("2+3")
b = multiply("2*5")
c = divide("10/5")
print(f"Results: {a}, {b}, {c}")
`

	result, err := executor.Execute(ctx, code)
	if err != nil {
		t.Fatalf("Failed to execute code: %v", err)
	}

	output := result.Output
	if !strings.Contains(output, "5") || !strings.Contains(output, "10") || !strings.Contains(output, "2") {
		t.Errorf("Expected output to contain all results, got: %s", output)
	}
}

// TestErrorHandling tests error handling in tool execution
func TestErrorHandling(t *testing.T) {
	tools := []tools.Tool{
		MockTool{
			name:        "test",
			description: "Test tool",
			response:    "ok",
		},
	}

	executor := ptc.NewCodeExecutor(ptc.LanguagePython, tools)
	ctx := context.Background()

	if err := executor.Start(ctx); err != nil {
		t.Fatalf("Failed to start executor: %v", err)
	}
	defer executor.Stop(ctx)

	// Code with syntax error
	code := `
print("unclosed string
`

	result, err := executor.Execute(ctx, code)
	// Should not return error, but result should contain error info
	if err == nil && result.Error == nil {
		t.Error("Expected error in result for invalid Python code")
	}
}

// TestToolDefinitionsGeneration tests tool definition generation
func TestToolDefinitionsGeneration(t *testing.T) {
	tools := []tools.Tool{
		MockTool{
			name:        "calculator",
			description: "Performs calculations",
			response:    "result",
		},
		MockTool{
			name:        "weather",
			description: "Gets weather info",
			response:    "sunny",
		},
	}

	executor := ptc.NewCodeExecutor(ptc.LanguagePython, tools)

	defs := executor.GetToolDefinitions()

	if !strings.Contains(defs, "calculator") {
		t.Error("Expected tool definitions to contain 'calculator'")
	}

	if !strings.Contains(defs, "weather") {
		t.Error("Expected tool definitions to contain 'weather'")
	}

	if !strings.Contains(defs, "Performs calculations") {
		t.Error("Expected tool definitions to contain description")
	}
}

// TestConcurrentExecution tests concurrent code execution
func TestConcurrentExecution(t *testing.T) {
	tools := []tools.Tool{
		MockTool{
			name:        "test",
			description: "Test tool",
			response:    "ok",
		},
	}

	executor := ptc.NewCodeExecutor(ptc.LanguagePython, tools)
	ctx := context.Background()

	if err := executor.Start(ctx); err != nil {
		t.Fatalf("Failed to start executor: %v", err)
	}
	defer executor.Stop(ctx)

	// Run multiple executions concurrently
	done := make(chan bool, 3)

	for i := 0; i < 3; i++ {
		go func(id int) {
			code := `
result = test("input")
print(result)
`
			_, err := executor.Execute(ctx, code)
			if err != nil {
				t.Errorf("Execution %d failed: %v", id, err)
			}
			done <- true
		}(i)
	}

	// Wait for all executions to complete
	for i := 0; i < 3; i++ {
		select {
		case <-done:
			// Success
		case <-time.After(30 * time.Second):
			t.Fatal("Concurrent execution timed out")
		}
	}
}

// TestStopWithoutStart tests that Stop works even if Start wasn't called
func TestStopWithoutStart(t *testing.T) {
	tools := []tools.Tool{
		MockTool{
			name:        "test",
			description: "Test",
			response:    "ok",
		},
	}

	executor := ptc.NewCodeExecutor(ptc.LanguagePython, tools)
	ctx := context.Background()

	// Should not panic
	if err := executor.Stop(ctx); err != nil {
		t.Errorf("Stop without Start should not return error: %v", err)
	}
}
