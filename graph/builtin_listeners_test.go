package graph_test

import (
	"bytes"
	"context"
	"fmt"
	"log"
	"strings"
	"testing"
	"time"

	"github.com/smallnest/langgraphgo/graph"
)

const (
	testState   = "test_state"
	step2Result = "step2_result"
)

func TestProgressListener_OnNodeEvent(t *testing.T) {
	t.Parallel()

	var buf bytes.Buffer
	listener := graph.NewProgressListenerWithWriter(&buf).
		WithTiming(false). // Disable timing for predictable output
		WithPrefix("🔄")

	ctx := context.Background()

	// Test start event
	listener.OnNodeEvent(ctx, graph.NodeEventStart, "test_node", nil, nil)
	output := buf.String()

	if !strings.Contains(output, "🔄 Starting test_node") {
		t.Errorf("Expected start message, got: %s", output)
	}

	// Test complete event
	buf.Reset()
	listener.OnNodeEvent(ctx, graph.NodeEventComplete, "test_node", nil, nil)
	output = buf.String()

	if !strings.Contains(output, "✅ test_node completed") {
		t.Errorf("Expected complete message, got: %s", output)
	}

	// Test error event
	buf.Reset()
	listener.OnNodeEvent(ctx, graph.NodeEventError, "test_node", nil, fmt.Errorf("test error"))
	output = buf.String()

	if !strings.Contains(output, "❌ test_node failed: test error") {
		t.Errorf("Expected error message, got: %s", output)
	}
}

func TestProgressListener_CustomSteps(t *testing.T) {
	t.Parallel()

	var buf bytes.Buffer
	listener := graph.NewProgressListenerWithWriter(&buf).
		WithTiming(false)

	// Set custom step message
	listener.SetNodeStep("process", "Analyzing data")

	ctx := context.Background()
	listener.OnNodeEvent(ctx, graph.NodeEventStart, "process", nil, nil)

	output := buf.String()
	if !strings.Contains(output, "🔄 Analyzing data") {
		t.Errorf("Expected custom message, got: %s", output)
	}
}

func TestProgressListener_WithDetails(t *testing.T) {
	t.Parallel()

	var buf bytes.Buffer
	listener := graph.NewProgressListenerWithWriter(&buf).
		WithTiming(false).
		WithDetails(true)

	ctx := context.Background()
	state := map[string]string{"key": "value"}

	listener.OnNodeEvent(ctx, graph.NodeEventComplete, "test_node", state, nil)

	output := buf.String()
	if !strings.Contains(output, "State: map[key:value]") {
		t.Errorf("Expected state details, got: %s", output)
	}
}

func TestLoggingListener_OnNodeEvent(t *testing.T) {
	t.Parallel()

	var buf bytes.Buffer
	logger := log.New(&buf, "[TEST] ", 0) // No timestamp for predictable output

	listener := graph.NewLoggingListenerWithLogger(logger).
		WithLogLevel(graph.LogLevelDebug)

	ctx := context.Background()

	// Test different event types
	listener.OnNodeEvent(ctx, graph.NodeEventStart, "test_node", nil, nil)
	listener.OnNodeEvent(ctx, graph.NodeEventComplete, "test_node", nil, nil)
	listener.OnNodeEvent(ctx, graph.NodeEventError, "test_node", nil, fmt.Errorf("test error"))

	output := buf.String()

	if !strings.Contains(output, "[TEST] START test_node") {
		t.Errorf("Expected start log, got: %s", output)
	}

	if !strings.Contains(output, "[TEST] COMPLETE test_node") {
		t.Errorf("Expected complete log, got: %s", output)
	}

	if !strings.Contains(output, "[TEST] ERROR test_node: test error") {
		t.Errorf("Expected error log, got: %s", output)
	}
}

func TestLoggingListener_LogLevel(t *testing.T) {
	t.Parallel()

	var buf bytes.Buffer
	logger := log.New(&buf, "[TEST] ", 0)

	listener := graph.NewLoggingListenerWithLogger(logger).
		WithLogLevel(graph.LogLevelError) // Only error level and above

	ctx := context.Background()

	// These should be filtered out
	listener.OnNodeEvent(ctx, graph.NodeEventStart, "test_node", nil, nil)
	listener.OnNodeEvent(ctx, graph.NodeEventProgress, "test_node", nil, nil)

	// This should be logged
	listener.OnNodeEvent(ctx, graph.NodeEventError, "test_node", nil, fmt.Errorf("test error"))

	output := buf.String()

	if strings.Contains(output, "START") || strings.Contains(output, "PROGRESS") {
		t.Errorf("Expected debug/info messages to be filtered, got: %s", output)
	}

	if !strings.Contains(output, "ERROR test_node") {
		t.Errorf("Expected error message, got: %s", output)
	}
}

func TestLoggingListener_WithState(t *testing.T) {
	t.Parallel()

	var buf bytes.Buffer
	logger := log.New(&buf, "[TEST] ", 0)

	listener := graph.NewLoggingListenerWithLogger(logger).
		WithState(true)

	ctx := context.Background()
	state := testState

	listener.OnNodeEvent(ctx, graph.NodeEventComplete, "test_node", state, nil)

	output := buf.String()
	if !strings.Contains(output, "State: test_state") {
		t.Errorf("Expected state in log, got: %s", output)
	}
}

func TestMetricsListener_OnNodeEvent(t *testing.T) {
	t.Parallel()

	listener := graph.NewMetricsListener()
	ctx := context.Background()

	// Simulate node execution
	listener.OnNodeEvent(ctx, graph.NodeEventStart, "test_node", nil, nil)
	time.Sleep(1 * time.Millisecond) // Small delay to measure
	listener.OnNodeEvent(ctx, graph.NodeEventComplete, "test_node", nil, nil)

	// Check metrics
	executions := listener.GetNodeExecutions()
	if executions["test_node"] != 1 {
		t.Errorf("Expected 1 execution, got %d", executions["test_node"])
	}

	avgDurations := listener.GetNodeAverageDuration()
	if _, exists := avgDurations["test_node"]; !exists {
		t.Error("Expected duration to be recorded")
	}

	if listener.GetTotalExecutions() != 1 {
		t.Errorf("Expected 1 total execution, got %d", listener.GetTotalExecutions())
	}
}

func TestMetricsListener_ErrorTracking(t *testing.T) {
	t.Parallel()

	listener := graph.NewMetricsListener()
	ctx := context.Background()

	// Simulate node execution with error
	listener.OnNodeEvent(ctx, graph.NodeEventStart, "error_node", nil, nil)
	listener.OnNodeEvent(ctx, graph.NodeEventError, "error_node", nil, fmt.Errorf("test error"))

	// Check error metrics
	errors := listener.GetNodeErrors()
	if errors["error_node"] != 1 {
		t.Errorf("Expected 1 error, got %d", errors["error_node"])
	}
}

func TestMetricsListener_PrintSummary(t *testing.T) {
	t.Parallel()

	listener := graph.NewMetricsListener()
	ctx := context.Background()

	// Generate some metrics
	listener.OnNodeEvent(ctx, graph.NodeEventStart, "node1", nil, nil)
	listener.OnNodeEvent(ctx, graph.NodeEventComplete, "node1", nil, nil)

	listener.OnNodeEvent(ctx, graph.NodeEventStart, "node2", nil, nil)
	listener.OnNodeEvent(ctx, graph.NodeEventError, "node2", nil, fmt.Errorf("error"))

	var buf bytes.Buffer
	listener.PrintSummary(&buf)

	output := buf.String()

	if !strings.Contains(output, "Node Execution Metrics") {
		t.Error("Expected metrics header")
	}

	if !strings.Contains(output, "Total Executions: 2") {
		t.Error("Expected total executions count")
	}

	if !strings.Contains(output, "node1: 1") {
		t.Error("Expected node1 execution count")
	}

	if !strings.Contains(output, "node2: 1 errors") {
		t.Error("Expected node2 error count")
	}
}

func TestMetricsListener_Reset(t *testing.T) {
	t.Parallel()

	listener := graph.NewMetricsListener()
	ctx := context.Background()

	// Generate some metrics
	listener.OnNodeEvent(ctx, graph.NodeEventStart, "test_node", nil, nil)
	listener.OnNodeEvent(ctx, graph.NodeEventComplete, "test_node", nil, nil)

	// Verify metrics exist
	if listener.GetTotalExecutions() != 1 {
		t.Error("Expected metrics to be recorded")
	}

	// Reset and verify metrics are cleared
	listener.Reset()

	if listener.GetTotalExecutions() != 0 {
		t.Error("Expected metrics to be reset")
	}

	executions := listener.GetNodeExecutions()
	if len(executions) != 0 {
		t.Error("Expected executions to be cleared")
	}
}

func TestChatListener_OnNodeEvent(t *testing.T) {
	t.Parallel()

	var buf bytes.Buffer
	listener := graph.NewChatListenerWithWriter(&buf).
		WithTime(false) // Disable time for predictable output

	ctx := context.Background()

	// Test start event
	listener.OnNodeEvent(ctx, graph.NodeEventStart, "process", nil, nil)
	output := buf.String()

	if !strings.Contains(output, "🤖 Starting process...") {
		t.Errorf("Expected start message, got: %s", output)
	}

	// Test complete event
	buf.Reset()
	listener.OnNodeEvent(ctx, graph.NodeEventComplete, "process", nil, nil)
	output = buf.String()

	if !strings.Contains(output, "✅ process finished") {
		t.Errorf("Expected complete message, got: %s", output)
	}
}

func TestChatListener_CustomMessages(t *testing.T) {
	t.Parallel()

	var buf bytes.Buffer
	listener := graph.NewChatListenerWithWriter(&buf).
		WithTime(false)

	// Set custom message
	listener.SetNodeMessage("analyze", "Analyzing your document")

	ctx := context.Background()
	listener.OnNodeEvent(ctx, graph.NodeEventStart, "analyze", nil, nil)

	output := buf.String()
	if !strings.Contains(output, "Analyzing your document") {
		t.Errorf("Expected custom message, got: %s", output)
	}
}

func TestChatListener_WithTime(t *testing.T) {
	t.Parallel()

	var buf bytes.Buffer
	listener := graph.NewChatListenerWithWriter(&buf).
		WithTime(true)

	ctx := context.Background()
	listener.OnNodeEvent(ctx, graph.NodeEventStart, "test", nil, nil)

	output := buf.String()
	// Should contain timestamp in format [HH:MM:SS]
	if !strings.Contains(output, "[") || !strings.Contains(output, "]") {
		t.Errorf("Expected timestamp in output, got: %s", output)
	}
}

// Integration test with actual graph execution
func TestBuiltinListeners_Integration(t *testing.T) {
	t.Parallel()

	// Create graph
	g := graph.NewListenableMessageGraph()

	node1 := g.AddNode("step1", func(_ context.Context, _ interface{}) (interface{}, error) {
		return "step1_result", nil
	})

	node2 := g.AddNode("step2", func(_ context.Context, _ interface{}) (interface{}, error) {
		return step2Result, nil
	})

	g.AddEdge("step1", "step2")
	g.AddEdge("step2", graph.END)
	g.SetEntryPoint("step1")

	// Add listeners
	var progressBuf, logBuf, chatBuf bytes.Buffer

	progressListener := graph.NewProgressListenerWithWriter(&progressBuf).WithTiming(false)
	logListener := graph.NewLoggingListenerWithLogger(log.New(&logBuf, "[GRAPH] ", 0))
	chatListener := graph.NewChatListenerWithWriter(&chatBuf).WithTime(false)
	metricsListener := graph.NewMetricsListener()

	node1.AddListener(progressListener)
	node1.AddListener(logListener)
	node1.AddListener(chatListener)
	node1.AddListener(metricsListener)

	node2.AddListener(progressListener)
	node2.AddListener(logListener)
	node2.AddListener(chatListener)
	node2.AddListener(metricsListener)

	// Execute graph
	runnable, err := g.CompileListenable()
	if err != nil {
		t.Fatalf("Failed to compile: %v", err)
	}

	ctx := context.Background()
	result, err := runnable.Invoke(ctx, "input")
	if err != nil {
		t.Fatalf("Execution failed: %v", err)
	}

	if result != step2Result {
		t.Errorf("Expected 'step2_result', got %v", result)
	}

	// Wait for async listeners
	time.Sleep(50 * time.Millisecond)

	// Check outputs
	progressOutput := progressBuf.String()
	if !strings.Contains(progressOutput, "Starting step1") {
		t.Errorf("Progress listener should show step1, got: %s", progressOutput)
	}

	logOutput := logBuf.String()
	if !strings.Contains(logOutput, "START step1") {
		t.Errorf("Log listener should show START step1, got: %s", logOutput)
	}

	chatOutput := chatBuf.String()
	if !strings.Contains(chatOutput, "🤖 Starting step1") {
		t.Errorf("Chat listener should show start message, got: %s", chatOutput)
	}

	// Check metrics
	executions := metricsListener.GetNodeExecutions()
	if executions["step1"] != 1 || executions["step2"] != 1 {
		t.Errorf("Expected 1 execution each, got: %v", executions)
	}

	if metricsListener.GetTotalExecutions() != 2 {
		t.Errorf("Expected 2 total executions, got %d", metricsListener.GetTotalExecutions())
	}
}
