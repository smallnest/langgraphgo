package graph_test

import (
	"context"
	"fmt"
	"sync"
	"testing"
	"time"

	"github.com/smallnest/langgraphgo/graph"
)

const (
	resultValue = "result"
)

func TestListenableNode_AddListener(t *testing.T) {
	t.Parallel()

	node := graph.Node{
		Name: testNode,
		Function: func(ctx context.Context, state interface{}) (interface{}, error) {
			return resultValue, nil
		},
	}

	listenableNode := graph.NewListenableNode(node)

	// Test adding listener
	var eventReceived bool
	var mu sync.Mutex
	listener := graph.NodeListenerFunc(func(ctx context.Context, event graph.NodeEvent, nodeName string, state interface{}, err error) {
		mu.Lock()
		eventReceived = true
		mu.Unlock()
	})

	listenableNode.AddListener(listener)

	// Verify listener was added
	listeners := listenableNode.GetListeners()
	if len(listeners) != 1 {
		t.Errorf("Expected 1 listener, got %d", len(listeners))
	}

	// Test listener is called during execution
	ctx := context.Background()
	_, err := listenableNode.Execute(ctx, "input")
	if err != nil {
		t.Fatalf("Execution failed: %v", err)
	}

	// Give some time for async listeners
	time.Sleep(10 * time.Millisecond)

	mu.Lock()
	received := eventReceived
	mu.Unlock()

	if !received {
		t.Error("Listener should have been called")
	}
}

func TestListenableNode_Execute(t *testing.T) {
	t.Parallel()

	node := graph.Node{
		Name: testNode,
		Function: func(ctx context.Context, state interface{}) (interface{}, error) {
			return fmt.Sprintf("processed_%v", state), nil
		},
	}

	listenableNode := graph.NewListenableNode(node)

	// Track events
	var events []graph.NodeEvent
	var mutex sync.Mutex

	listener := graph.NodeListenerFunc(func(ctx context.Context, event graph.NodeEvent, nodeName string, state interface{}, err error) {
		mutex.Lock()
		defer mutex.Unlock()
		events = append(events, event)
	})

	listenableNode.AddListener(listener)

	// Execute node
	ctx := context.Background()
	result, err := listenableNode.Execute(ctx, "test_input")

	if err != nil {
		t.Fatalf("Execution failed: %v", err)
	}

	if result != "processed_test_input" {
		t.Errorf("Expected 'processed_test_input', got %v", result)
	}

	// Wait for async listeners
	time.Sleep(50 * time.Millisecond)

	mutex.Lock()
	defer mutex.Unlock()

	// Should have start and complete events
	if len(events) != 2 {
		t.Errorf("Expected 2 events, got %d", len(events))
		return
	}

	// Events may arrive out of order due to goroutines, so check presence instead
	hasStart := false
	hasComplete := false

	for _, event := range events {
		switch event {
		case graph.NodeEventStart:
			hasStart = true
		case graph.NodeEventComplete:
			hasComplete = true
		case graph.NodeEventProgress, graph.NodeEventError:
			// These events are not expected in this test but handled for completeness
		}
	}

	if !hasStart {
		t.Error("Should have received start event")
	}

	if !hasComplete {
		t.Error("Should have received complete event")
	}
}

func TestListenableNode_ExecuteWithError(t *testing.T) {
	t.Parallel()

	node := graph.Node{
		Name: "error_node",
		Function: func(ctx context.Context, state interface{}) (interface{}, error) {
			return nil, fmt.Errorf("test error")
		},
	}

	listenableNode := graph.NewListenableNode(node)

	// Track events
	var events []graph.NodeEvent
	var lastError error
	var mutex sync.Mutex

	listener := graph.NodeListenerFunc(func(ctx context.Context, event graph.NodeEvent, nodeName string, state interface{}, err error) {
		mutex.Lock()
		defer mutex.Unlock()
		events = append(events, event)
		if err != nil {
			lastError = err
		}
	})

	listenableNode.AddListener(listener)

	// Execute node (should fail)
	ctx := context.Background()
	_, err := listenableNode.Execute(ctx, "test_input")

	if err == nil {
		t.Fatal("Expected execution to fail")
	}

	// Wait for async listeners
	time.Sleep(50 * time.Millisecond)

	mutex.Lock()
	defer mutex.Unlock()

	// Should have start and error events
	if len(events) != 2 {
		t.Errorf("Expected 2 events, got %d", len(events))
		return
	}

	// Events may arrive out of order due to goroutines, so check presence instead
	hasStart := false
	hasError := false

	for _, event := range events {
		switch event {
		case graph.NodeEventStart:
			hasStart = true
		case graph.NodeEventError:
			hasError = true
		case graph.NodeEventProgress, graph.NodeEventComplete:
			// These events are not expected in this test but handled for completeness
		}
	}

	if !hasStart {
		t.Error("Should have received start event")
	}

	if !hasError {
		t.Error("Should have received error event")
	}

	if lastError == nil {
		t.Error("Error event should contain the error")
	}
}

func TestListenableMessageGraph_AddNode(t *testing.T) {
	t.Parallel()

	g := graph.NewListenableMessageGraph()

	// Add a node
	node := g.AddNode(testNode, func(ctx context.Context, state interface{}) (interface{}, error) {
		return resultValue, nil
	})

	if node == nil {
		t.Fatal("AddNode should return a ListenableNode")
	}

	// Verify node was added to graph
	listenableNode := g.GetListenableNode(testNode)
	if listenableNode == nil {
		t.Fatal("Node should be retrievable")
		return
	}

	if listenableNode.Name != testNode {
		t.Errorf("Expected node name 'test_node', got %v", listenableNode.Name)
	}
}

func TestListenableMessageGraph_GlobalListeners(t *testing.T) {
	t.Parallel()

	g := graph.NewListenableMessageGraph()

	// Add multiple nodes
	node1 := g.AddNode("node1", func(ctx context.Context, state interface{}) (interface{}, error) {
		return "result1", nil
	})

	node2 := g.AddNode("node2", func(ctx context.Context, state interface{}) (interface{}, error) {
		return "result2", nil
	})

	// Add global listener
	var eventCount int
	var mutex sync.Mutex

	globalListener := graph.NodeListenerFunc(func(ctx context.Context, event graph.NodeEvent, nodeName string, state interface{}, err error) {
		mutex.Lock()
		defer mutex.Unlock()
		eventCount++
	})

	g.AddGlobalListener(globalListener)

	// Verify listeners were added to all nodes
	if len(node1.GetListeners()) != 1 {
		t.Error("Global listener should be added to node1")
	}

	if len(node2.GetListeners()) != 1 {
		t.Error("Global listener should be added to node2")
	}

	// Execute both nodes
	ctx := context.Background()
	_, _ = node1.Execute(ctx, "input1")
	_, _ = node2.Execute(ctx, "input2")

	// Wait for async listeners
	time.Sleep(20 * time.Millisecond)

	mutex.Lock()
	defer mutex.Unlock()

	// Should have 4 events total (2 nodes * 2 events each)
	if eventCount != 4 {
		t.Errorf("Expected 4 events, got %d", eventCount)
	}
}

func TestListenableRunnable_Invoke(t *testing.T) {
	t.Parallel()

	g := graph.NewListenableMessageGraph()

	// Create a simple pipeline
	node1 := g.AddNode("node1", func(ctx context.Context, state interface{}) (interface{}, error) {
		return fmt.Sprintf("step1_%v", state), nil
	})

	node2 := g.AddNode("node2", func(ctx context.Context, state interface{}) (interface{}, error) {
		return fmt.Sprintf("step2_%v", state), nil
	})

	// Add edges
	g.AddEdge("node1", "node2")
	g.AddEdge("node2", graph.END)
	g.SetEntryPoint("node1")

	// Track execution flow
	var executionFlow []string
	var mutex sync.Mutex

	listener := graph.NodeListenerFunc(func(ctx context.Context, event graph.NodeEvent, nodeName string, state interface{}, err error) {
		mutex.Lock()
		defer mutex.Unlock()
		executionFlow = append(executionFlow, fmt.Sprintf("%s:%s", nodeName, event))
	})

	node1.AddListener(listener)
	node2.AddListener(listener)

	// Compile and execute
	runnable, err := g.CompileListenable()
	if err != nil {
		t.Fatalf("Failed to compile: %v", err)
	}

	ctx := context.Background()
	result, err := runnable.Invoke(ctx, "input")
	if err != nil {
		t.Fatalf("Execution failed: %v", err)
	}

	if result != "step2_step1_input" {
		t.Errorf("Expected 'step2_step1_input', got %v", result)
	}

	// Wait for async listeners
	time.Sleep(50 * time.Millisecond)

	mutex.Lock()
	defer mutex.Unlock()

	// Should have events for both nodes (4 total events)
	if len(executionFlow) != 4 {
		t.Errorf("Expected 4 events, got %d: %v", len(executionFlow), executionFlow)
		return
	}

	// Check that we have the right events (order may vary due to goroutines)
	eventCounts := make(map[string]int)
	for _, event := range executionFlow {
		eventCounts[event]++
	}

	expectedEvents := map[string]int{
		"node1:start":    1,
		"node1:complete": 1,
		"node2:start":    1,
		"node2:complete": 1,
	}

	for expectedEvent, expectedCount := range expectedEvents {
		if eventCounts[expectedEvent] != expectedCount {
			t.Errorf("Expected %d occurrences of %s, got %d", expectedCount, expectedEvent, eventCounts[expectedEvent])
		}
	}
}

func TestListenerPanicRecovery(t *testing.T) {
	t.Parallel()

	node := graph.Node{
		Name: testNode,
		Function: func(ctx context.Context, state interface{}) (interface{}, error) {
			return resultValue, nil
		},
	}

	listenableNode := graph.NewListenableNode(node)

	// Add a panicking listener
	panicListener := graph.NodeListenerFunc(func(ctx context.Context, event graph.NodeEvent, nodeName string, state interface{}, err error) {
		panic("test panic")
	})

	// Add a normal listener
	var normalListenerCalled bool
	var mutex sync.Mutex

	normalListener := graph.NodeListenerFunc(func(ctx context.Context, event graph.NodeEvent, nodeName string, state interface{}, err error) {
		mutex.Lock()
		defer mutex.Unlock()
		normalListenerCalled = true
	})

	listenableNode.AddListener(panicListener)
	listenableNode.AddListener(normalListener)

	// Execute should not panic even though listener panics
	ctx := context.Background()
	result, err := listenableNode.Execute(ctx, "input")

	if err != nil {
		t.Fatalf("Execution failed: %v", err)
	}

	if result != resultValue {
		t.Errorf("Expected 'result', got %v", result)
	}

	// Wait for async listeners
	time.Sleep(20 * time.Millisecond)

	mutex.Lock()
	defer mutex.Unlock()

	// Normal listener should still have been called
	if !normalListenerCalled {
		t.Error("Normal listener should have been called despite panic in other listener")
	}
}

func TestStreamEvent_Creation(t *testing.T) {
	t.Parallel()

	timestamp := time.Now()
	event := &graph.StreamEvent{
		Timestamp: timestamp,
		NodeName:  testNode,
		Event:     graph.NodeEventComplete,
		State:     testState,
		Error:     nil,
		Metadata:  map[string]interface{}{"key": "value"},
		Duration:  100 * time.Millisecond,
	}

	if event.Timestamp != timestamp {
		t.Error("Timestamp should be preserved")
	}

	if event.NodeName != testNode {
		t.Error("NodeName should be preserved")
	}

	if event.Event != graph.NodeEventComplete {
		t.Error("Event should be preserved")
	}

	if event.State != testState {
		t.Error("State should be preserved")
	}

	if event.Duration != 100*time.Millisecond {
		t.Error("Duration should be preserved")
	}
}

// Benchmark tests
func BenchmarkListenableNode_Execute(b *testing.B) {
	node := graph.Node{
		Name: "benchmark_node",
		Function: func(ctx context.Context, state interface{}) (interface{}, error) {
			return state, nil
		},
	}

	listenableNode := graph.NewListenableNode(node)

	// Add a listener
	listener := graph.NodeListenerFunc(func(ctx context.Context, event graph.NodeEvent, nodeName string, state interface{}, err error) {
		// No-op listener for benchmarking
	})

	listenableNode.AddListener(listener)

	ctx := context.Background()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = listenableNode.Execute(ctx, "test")
	}
}

func BenchmarkListenableRunnable_Invoke(b *testing.B) {
	g := graph.NewListenableMessageGraph()

	node := g.AddNode("node", func(ctx context.Context, state interface{}) (interface{}, error) {
		return state, nil
	})

	g.AddEdge("node", graph.END)
	g.SetEntryPoint("node")

	// Add listener
	listener := graph.NodeListenerFunc(func(ctx context.Context, event graph.NodeEvent, nodeName string, state interface{}, err error) {
		// No-op for benchmarking
	})

	node.AddListener(listener)

	runnable, _ := g.CompileListenable()
	ctx := context.Background()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = runnable.Invoke(ctx, "test")
	}
}
