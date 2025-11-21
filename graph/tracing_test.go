package graph_test

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"testing"

	"github.com/smallnest/langgraphgo/graph"
)

func TestTracer_StartEndSpan(t *testing.T) {
	t.Parallel()

	tracer := graph.NewTracer()
	ctx := context.Background()

	// Test starting a span
	span := tracer.StartSpan(ctx, graph.TraceEventNodeStart, "test_node")

	if span.ID == "" {
		t.Error("Span ID should not be empty")
	}

	if span.Event != graph.TraceEventNodeStart {
		t.Errorf("Expected event %v, got %v", graph.TraceEventNodeStart, span.Event)
	}

	if span.NodeName != "test_node" {
		t.Errorf("Expected node name 'test_node', got %v", span.NodeName)
	}

	if span.StartTime.IsZero() {
		t.Error("Start time should be set")
	}

	if !span.EndTime.IsZero() {
		t.Error("End time should not be set yet")
	}

	// Test ending a span
	testState := "test_state"
	tracer.EndSpan(ctx, span, testState, nil)

	if span.EndTime.IsZero() {
		t.Error("End time should be set after ending span")
	}

	if span.Duration <= 0 {
		t.Error("Duration should be positive after ending span")
	}

	if span.State != testState {
		t.Errorf("Expected state %v, got %v", testState, span.State)
	}

	if span.Event != graph.TraceEventNodeEnd {
		t.Errorf("Expected event to be updated to %v, got %v", graph.TraceEventNodeEnd, span.Event)
	}
}

func TestTracer_SpanWithError(t *testing.T) {
	t.Parallel()

	tracer := graph.NewTracer()
	ctx := context.Background()

	span := tracer.StartSpan(ctx, graph.TraceEventNodeStart, "error_node")
	testError := fmt.Errorf("test error")

	tracer.EndSpan(ctx, span, nil, testError)

	if !errors.Is(span.Error, testError) {
		t.Errorf("Expected error %v, got %v", testError, span.Error)
	}

	if span.Event != graph.TraceEventNodeError {
		t.Errorf("Expected event to be %v for error case, got %v", graph.TraceEventNodeError, span.Event)
	}
}

func TestTracer_Hooks(t *testing.T) {
	t.Parallel()

	tracer := graph.NewTracer()
	ctx := context.Background()

	// Track hook calls
	hookCalls := make([]graph.TraceEvent, 0)

	hook := graph.TraceHookFunc(func(ctx context.Context, span *graph.TraceSpan) {
		hookCalls = append(hookCalls, span.Event)
	})

	tracer.AddHook(hook)

	// Create and end a span
	span := tracer.StartSpan(ctx, graph.TraceEventNodeStart, "hooked_node")
	tracer.EndSpan(ctx, span, "state", nil)

	// Should have 2 hook calls: start and end
	if len(hookCalls) != 2 {
		t.Errorf("Expected 2 hook calls, got %d", len(hookCalls))
	}

	// Verify the hook calls
	if hookCalls[0] != graph.TraceEventNodeStart {
		t.Errorf("First hook call should be start event, got %v", hookCalls[0])
	}

	if hookCalls[1] != graph.TraceEventNodeEnd {
		t.Errorf("Second hook call should be end event, got %v", hookCalls[1])
	}
}

func TestTracer_EdgeTraversal(t *testing.T) {
	t.Parallel()

	tracer := graph.NewTracer()
	ctx := context.Background()

	var edgeSpan *graph.TraceSpan
	hook := graph.TraceHookFunc(func(ctx context.Context, span *graph.TraceSpan) {
		if span.Event == graph.TraceEventEdgeTraversal {
			edgeSpan = span
		}
	})

	tracer.AddHook(hook)
	tracer.TraceEdgeTraversal(ctx, "node1", "node2")

	if edgeSpan == nil {
		t.Fatal("Edge traversal span was not captured")
	}

	if edgeSpan.FromNode != "node1" {
		t.Errorf("Expected FromNode 'node1', got %v", edgeSpan.FromNode)
	}

	if edgeSpan.ToNode != "node2" {
		t.Errorf("Expected ToNode 'node2', got %v", edgeSpan.ToNode)
	}

	if edgeSpan.Event != graph.TraceEventEdgeTraversal {
		t.Errorf("Expected event %v, got %v", graph.TraceEventEdgeTraversal, edgeSpan.Event)
	}
}

func TestContextWithSpan(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	span := &graph.TraceSpan{ID: "test_span"}

	// Test storing span in context
	newCtx := graph.ContextWithSpan(ctx, span)

	// Test retrieving span from context
	retrievedSpan := graph.SpanFromContext(newCtx)

	if retrievedSpan == nil {
		t.Fatal("Should be able to retrieve span from context")
	}

	if retrievedSpan.ID != "test_span" {
		t.Errorf("Expected span ID 'test_span', got %v", retrievedSpan.ID)
	}

	// Test retrieving from context without span
	emptySpan := graph.SpanFromContext(ctx)
	if emptySpan != nil {
		t.Error("Should return nil when no span in context")
	}
}

func TestTracedRunnable_Invoke(t *testing.T) {
	t.Parallel()

	// Create a simple graph
	g := graph.NewMessageGraph()

	g.AddNode("node1", func(ctx context.Context, state interface{}) (interface{}, error) {
		return fmt.Sprintf("processed_%v", state), nil
	})

	g.AddNode("node2", func(ctx context.Context, state interface{}) (interface{}, error) {
		return fmt.Sprintf("final_%v", state), nil
	})

	g.AddEdge("node1", "node2")
	g.AddEdge("node2", graph.END)
	g.SetEntryPoint("node1")

	runnable, err := g.Compile()
	if err != nil {
		t.Fatalf("Failed to compile graph: %v", err)
	}

	// Create tracer and traced runnable
	tracer := graph.NewTracer()
	tracedRunnable := graph.NewTracedRunnable(runnable, tracer)

	// Collect trace events
	events := make([]string, 0)
	hook := graph.TraceHookFunc(func(ctx context.Context, span *graph.TraceSpan) {
		events = append(events, fmt.Sprintf("%v:%v", span.Event, span.NodeName))
	})
	tracer.AddHook(hook)

	// Execute the graph
	ctx := context.Background()
	result, err := tracedRunnable.Invoke(ctx, "test")

	if err != nil {
		t.Fatalf("Execution failed: %v", err)
	}

	if result != "final_processed_test" {
		t.Errorf("Expected result 'final_processed_test', got %v", result)
	}

	// Verify trace events
	expectedEvents := []string{
		"graph_start:",
		"node_start:node1",
		"node_end:node1",
		"edge_traversal:",
		"node_start:node2",
		"node_end:node2",
		"edge_traversal:",
		"graph_end:",
	}

	if len(events) != len(expectedEvents) {
		t.Errorf("Expected %d trace events, got %d: %v", len(expectedEvents), len(events), events)
	}

	for i, expected := range expectedEvents {
		if i < len(events) && events[i] != expected {
			t.Errorf("Event %d: expected %v, got %v", i, expected, events[i])
		}
	}
}

func TestTracedRunnable_WithError(t *testing.T) {
	t.Parallel()

	// Create a graph with an error-producing node
	g := graph.NewMessageGraph()

	g.AddNode("error_node", func(ctx context.Context, state interface{}) (interface{}, error) {
		return nil, fmt.Errorf("intentional error")
	})

	g.AddEdge("error_node", graph.END)
	g.SetEntryPoint("error_node")

	runnable, err := g.Compile()
	if err != nil {
		t.Fatalf("Failed to compile graph: %v", err)
	}

	// Create tracer and traced runnable
	tracer := graph.NewTracer()
	tracedRunnable := graph.NewTracedRunnable(runnable, tracer)

	// Collect error events
	errorEvents := make([]*graph.TraceSpan, 0)
	hook := graph.TraceHookFunc(func(ctx context.Context, span *graph.TraceSpan) {
		if span.Event == graph.TraceEventNodeError {
			errorEvents = append(errorEvents, span)
		}
	})
	tracer.AddHook(hook)

	// Execute the graph (should fail)
	ctx := context.Background()
	_, err = tracedRunnable.Invoke(ctx, "test")

	if err == nil {
		t.Fatal("Expected execution to fail")
	}

	// Should have captured error event
	if len(errorEvents) != 1 {
		t.Errorf("Expected 1 error event, got %d", len(errorEvents))
	}

	if len(errorEvents) > 0 {
		errorSpan := errorEvents[0]
		if errorSpan.Error == nil {
			t.Error("Error span should contain the error")
		}

		if !strings.Contains(errorSpan.Error.Error(), "intentional error") {
			t.Errorf("Expected error to contain 'intentional error', got %v", errorSpan.Error)
		}
	}
}

func TestTracer_SpanHierarchy(t *testing.T) {
	t.Parallel()

	tracer := graph.NewTracer()
	ctx := context.Background()

	// Create parent span
	parentSpan := tracer.StartSpan(ctx, graph.TraceEventGraphStart, "")
	parentCtx := graph.ContextWithSpan(ctx, parentSpan)

	// Create child span
	childSpan := tracer.StartSpan(parentCtx, graph.TraceEventNodeStart, "child_node")

	// Child span should have parent ID
	if childSpan.ParentID != parentSpan.ID {
		t.Errorf("Expected child span parent ID %v, got %v", parentSpan.ID, childSpan.ParentID)
	}

	// Parent span should not have parent ID
	if parentSpan.ParentID != "" {
		t.Errorf("Expected parent span to have empty parent ID, got %v", parentSpan.ParentID)
	}
}

// Benchmark tests
func BenchmarkTracer_StartEndSpan(b *testing.B) {
	tracer := graph.NewTracer()
	ctx := context.Background()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		span := tracer.StartSpan(ctx, graph.TraceEventNodeStart, "benchmark_node")
		tracer.EndSpan(ctx, span, "state", nil)
	}
}

func BenchmarkTracedRunnable_Invoke(b *testing.B) {
	// Create a simple graph
	g := graph.NewMessageGraph()
	g.AddNode("node", func(ctx context.Context, state interface{}) (interface{}, error) {
		return state, nil
	})
	g.AddEdge("node", graph.END)
	g.SetEntryPoint("node")

	runnable, _ := g.Compile()
	tracer := graph.NewTracer()
	tracedRunnable := graph.NewTracedRunnable(runnable, tracer)

	ctx := context.Background()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = tracedRunnable.Invoke(ctx, "test")
		tracer.Clear() // Clear spans to avoid memory buildup
	}
}
