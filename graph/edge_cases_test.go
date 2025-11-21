package graph_test

import (
	"context"
	"errors"
	"fmt"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/smallnest/langgraphgo/graph"
	"github.com/tmc/langchaingo/llms"
)

// TestEmptyGraph tests behavior with empty graphs
func TestEmptyGraph(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name        string
		buildGraph  func() *graph.MessageGraph
		expectError bool
		errorMsg    string
	}{
		{
			name: "Graph with no nodes",
			buildGraph: func() *graph.MessageGraph {
				g := graph.NewMessageGraph()
				return g
			},
			expectError: true,
			errorMsg:    "entry point not set",
		},
		{
			name: "Graph with nodes but no entry point",
			buildGraph: func() *graph.MessageGraph {
				g := graph.NewMessageGraph()
				g.AddNode("node1", func(ctx context.Context, state interface{}) (interface{}, error) {
					return state, nil
				})
				return g
			},
			expectError: true,
			errorMsg:    "entry point not set",
		},
		{
			name: "Graph with self-referencing node",
			buildGraph: func() *graph.MessageGraph {
				g := graph.NewMessageGraph()
				g.AddNode("node1", func(ctx context.Context, state interface{}) (interface{}, error) {
					return state, nil
				})
				g.AddEdge("node1", "node1") // Self-loop
				g.SetEntryPoint("node1")
				return g
			},
			expectError: false, // Will create infinite loop, but that's valid
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			g := tt.buildGraph()
			_, err := g.Compile()

			if tt.expectError && err == nil {
				t.Error("Expected error but got none")
			}
			if !tt.expectError && err != nil {
				t.Errorf("Unexpected error: %v", err)
			}
		})
	}
}

// TestLargeGraph tests performance with large graphs
func TestLargeGraph(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping large graph test in short mode")
	}

	g := graph.NewMessageGraph()

	// Create a chain of 1000 nodes
	nodeCount := 1000
	for i := 0; i < nodeCount; i++ {
		nodeName := fmt.Sprintf("node_%d", i)
		g.AddNode(nodeName, func(ctx context.Context, state interface{}) (interface{}, error) {
			counter := state.(int)
			return counter + 1, nil
		})

		if i > 0 {
			prevNode := fmt.Sprintf("node_%d", i-1)
			g.AddEdge(prevNode, nodeName)
		}
	}

	lastNode := fmt.Sprintf("node_%d", nodeCount-1)
	g.AddEdge(lastNode, graph.END)
	g.SetEntryPoint("node_0")

	runnable, err := g.Compile()
	if err != nil {
		t.Fatalf("Failed to compile large graph: %v", err)
	}

	start := time.Now()
	result, err := runnable.Invoke(context.Background(), 0)
	duration := time.Since(start)

	if err != nil {
		t.Fatalf("Failed to execute large graph: %v", err)
	}

	if result != nodeCount {
		t.Errorf("Expected result %d, got %v", nodeCount, result)
	}

	t.Logf("Large graph with %d nodes executed in %v", nodeCount, duration)
}

// TestConcurrentExecution tests thread safety
func TestConcurrentExecution(t *testing.T) {
	t.Parallel()

	g := graph.NewMessageGraph()

	var counter int32
	g.AddNode("increment", func(ctx context.Context, state interface{}) (interface{}, error) {
		atomic.AddInt32(&counter, 1)
		return state, nil
	})

	g.AddEdge("increment", graph.END)
	g.SetEntryPoint("increment")

	runnable, err := g.Compile()
	if err != nil {
		t.Fatalf("Failed to compile: %v", err)
	}

	// Run multiple executions concurrently
	concurrency := 100
	var wg sync.WaitGroup
	errors := make(chan error, concurrency)

	for i := 0; i < concurrency; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()
			ctx := context.Background()
			_, err := runnable.Invoke(ctx, id)
			if err != nil {
				errors <- err
			}
		}(i)
	}

	wg.Wait()
	close(errors)

	// Check for errors
	for err := range errors {
		t.Errorf("Concurrent execution error: %v", err)
	}

	// Verify all executions completed
	finalCount := atomic.LoadInt32(&counter)
	if finalCount != int32(concurrency) {
		t.Errorf("Expected %d executions, got %d", concurrency, finalCount)
	}
}

// TestContextCancellation tests context cancellation handling
func TestContextCancellation(t *testing.T) {
	t.Parallel()

	g := graph.NewMessageGraph()

	// Add a slow node
	g.AddNode("slow_node", func(ctx context.Context, state interface{}) (interface{}, error) {
		select {
		case <-time.After(5 * time.Second):
			return state, nil
		case <-ctx.Done():
			return nil, ctx.Err()
		}
	})

	g.AddEdge("slow_node", graph.END)
	g.SetEntryPoint("slow_node")

	runnable, err := g.Compile()
	if err != nil {
		t.Fatalf("Failed to compile: %v", err)
	}

	// Create a context with timeout
	ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
	defer cancel()

	start := time.Now()
	_, err = runnable.Invoke(ctx, "test")
	duration := time.Since(start)

	if err == nil {
		t.Error("Expected context cancellation error")
	}

	if duration > 200*time.Millisecond {
		t.Errorf("Cancellation took too long: %v", duration)
	}
}

// TestPanicRecovery tests panic handling in node functions
func TestPanicRecovery(t *testing.T) {
	t.Parallel()

	g := graph.NewMessageGraph()

	g.AddNode("panic_node", func(ctx context.Context, state interface{}) (interface{}, error) {
		panic("intentional panic")
	})

	g.AddEdge("panic_node", graph.END)
	g.SetEntryPoint("panic_node")

	runnable, err := g.Compile()
	if err != nil {
		t.Fatalf("Failed to compile: %v", err)
	}

	// This should handle the panic gracefully
	defer func() {
		if r := recover(); r == nil {
			t.Error("Expected panic to propagate")
		}
	}()

	ctx := context.Background()
	_, _ = runnable.Invoke(ctx, "test")
}

// TestComplexConditionalRouting tests complex conditional edge scenarios
func TestComplexConditionalRouting(t *testing.T) {
	t.Parallel()

	g := graph.NewMessageGraph()

	// Create a decision tree with multiple levels
	g.AddNode("root", func(ctx context.Context, state interface{}) (interface{}, error) {
		return state, nil
	})

	g.AddNode("branch_a", func(ctx context.Context, state interface{}) (interface{}, error) {
		m := state.(map[string]interface{})
		m["path"] = append(m["path"].([]string), "A")
		return m, nil
	})

	g.AddNode("branch_b", func(ctx context.Context, state interface{}) (interface{}, error) {
		m := state.(map[string]interface{})
		m["path"] = append(m["path"].([]string), "B")
		return m, nil
	})

	g.AddNode("leaf_a1", func(ctx context.Context, state interface{}) (interface{}, error) {
		m := state.(map[string]interface{})
		m["path"] = append(m["path"].([]string), "A1")
		return m, nil
	})

	g.AddNode("leaf_a2", func(ctx context.Context, state interface{}) (interface{}, error) {
		m := state.(map[string]interface{})
		m["path"] = append(m["path"].([]string), "A2")
		return m, nil
	})

	g.AddNode("leaf_b1", func(ctx context.Context, state interface{}) (interface{}, error) {
		m := state.(map[string]interface{})
		m["path"] = append(m["path"].([]string), "B1")
		return m, nil
	})

	// Root conditional
	g.AddConditionalEdge("root", func(ctx context.Context, state interface{}) string {
		m := state.(map[string]interface{})
		if m["choice"].(string) == "A" {
			return "branch_a"
		}
		return "branch_b"
	})

	// Branch A conditional
	g.AddConditionalEdge("branch_a", func(ctx context.Context, state interface{}) string {
		m := state.(map[string]interface{})
		if m["subchoice"].(int) == 1 {
			return "leaf_a1"
		}
		return "leaf_a2"
	})

	// Branch B always goes to B1
	g.AddEdge("branch_b", "leaf_b1")

	// All leaves go to END
	g.AddEdge("leaf_a1", graph.END)
	g.AddEdge("leaf_a2", graph.END)
	g.AddEdge("leaf_b1", graph.END)

	g.SetEntryPoint("root")

	runnable, err := g.Compile()
	if err != nil {
		t.Fatalf("Failed to compile: %v", err)
	}

	// Test different paths
	tests := []struct {
		input        map[string]interface{}
		expectedPath []string
	}{
		{
			input: map[string]interface{}{
				"choice":    "A",
				"subchoice": 1,
				"path":      []string{},
			},
			expectedPath: []string{"A", "A1"},
		},
		{
			input: map[string]interface{}{
				"choice":    "A",
				"subchoice": 2,
				"path":      []string{},
			},
			expectedPath: []string{"A", "A2"},
		},
		{
			input: map[string]interface{}{
				"choice":    "B",
				"subchoice": 1,
				"path":      []string{},
			},
			expectedPath: []string{"B", "B1"},
		},
	}

	for i, tt := range tests {
		t.Run(fmt.Sprintf("Path_%d", i), func(t *testing.T) {
			result, err := runnable.Invoke(context.Background(), tt.input)
			if err != nil {
				t.Fatalf("Execution failed: %v", err)
			}

			m := result.(map[string]interface{})
			path := m["path"].([]string)

			if len(path) != len(tt.expectedPath) {
				t.Errorf("Expected path %v, got %v", tt.expectedPath, path)
			} else {
				for j, p := range path {
					if p != tt.expectedPath[j] {
						t.Errorf("Path mismatch at %d: expected %s, got %s", j, tt.expectedPath[j], p)
					}
				}
			}
		})
	}
}

// TestStateModification tests various state modification scenarios
//
//nolint:gocognit,cyclop // Complex state modification scenarios require extensive testing
func TestStateModification(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name           string
		buildGraph     func() *graph.MessageGraph
		initialState   interface{}
		expectedResult interface{}
	}{
		{
			name: "Accumulator pattern",
			buildGraph: func() *graph.MessageGraph {
				g := graph.NewMessageGraph()
				g.AddNode("accumulate", func(ctx context.Context, state interface{}) (interface{}, error) {
					acc := state.([]int)
					return append(acc, len(acc)+1), nil
				})
				g.AddConditionalEdge("accumulate", func(ctx context.Context, state interface{}) string {
					acc := state.([]int)
					if len(acc) >= 5 {
						return graph.END
					}
					return "accumulate"
				})
				g.SetEntryPoint("accumulate")
				return g
			},
			initialState:   []int{},
			expectedResult: []int{1, 2, 3, 4, 5},
		},
		{
			name: "Map transformation",
			buildGraph: func() *graph.MessageGraph {
				g := graph.NewMessageGraph()
				g.AddNode("transform", func(ctx context.Context, state interface{}) (interface{}, error) {
					m := state.(map[string]interface{})
					// Transform each value
					for k, v := range m {
						if num, ok := v.(int); ok {
							m[k] = num * 2
						}
					}
					return m, nil
				})
				g.AddEdge("transform", graph.END)
				g.SetEntryPoint("transform")
				return g
			},
			initialState: map[string]interface{}{
				"a": 1,
				"b": 2,
				"c": 3,
			},
			expectedResult: map[string]interface{}{
				"a": 2,
				"b": 4,
				"c": 6,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			g := tt.buildGraph()
			runnable, err := g.Compile()
			if err != nil {
				t.Fatalf("Failed to compile: %v", err)
			}

			result, err := runnable.Invoke(context.Background(), tt.initialState)
			if err != nil {
				t.Fatalf("Execution failed: %v", err)
			}

			// Compare results
			switch expected := tt.expectedResult.(type) {
			case []int:
				actual := result.([]int)
				if len(actual) != len(expected) {
					t.Errorf("Expected %v, got %v", expected, actual)
				} else {
					for i := range expected {
						if actual[i] != expected[i] {
							t.Errorf("Mismatch at index %d: expected %d, got %d", i, expected[i], actual[i])
						}
					}
				}
			case map[string]interface{}:
				actual := result.(map[string]interface{})
				for k, v := range expected {
					if actual[k] != v {
						t.Errorf("Map mismatch for key %s: expected %v, got %v", k, v, actual[k])
					}
				}
			}
		})
	}
}

// TestErrorPropagation tests error handling and propagation
func TestErrorPropagation(t *testing.T) {
	t.Parallel()

	g := graph.NewMessageGraph()

	g.AddNode("node1", func(ctx context.Context, state interface{}) (interface{}, error) {
		return "step1", nil
	})

	g.AddNode("node2", func(ctx context.Context, state interface{}) (interface{}, error) {
		return nil, errors.New("deliberate error in node2")
	})

	g.AddNode("node3", func(ctx context.Context, state interface{}) (interface{}, error) {
		t.Error("node3 should not be executed after error in node2")
		return state, nil
	})

	g.AddEdge("node1", "node2")
	g.AddEdge("node2", "node3")
	g.AddEdge("node3", graph.END)
	g.SetEntryPoint("node1")

	runnable, err := g.Compile()
	if err != nil {
		t.Fatalf("Failed to compile: %v", err)
	}

	_, err = runnable.Invoke(context.Background(), "start")
	if err == nil {
		t.Error("Expected error from node2")
	} else if !errors.Is(err, errors.New("deliberate error in node2")) {
		// Check error message contains expected text
		if err.Error() != "error in node node2: deliberate error in node2" {
			t.Errorf("Unexpected error: %v", err)
		}
	}
}

// TestMessageContentEdgeCases tests edge cases with message content
func TestMessageContentEdgeCases(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name         string
		initialState []llms.MessageContent
		transform    func([]llms.MessageContent) []llms.MessageContent
		validate     func(*testing.T, []llms.MessageContent)
	}{
		{
			name:         "Empty message list",
			initialState: []llms.MessageContent{},
			transform: func(msgs []llms.MessageContent) []llms.MessageContent {
				return append(msgs, llms.TextParts("ai", "Hello"))
			},
			validate: func(t *testing.T, msgs []llms.MessageContent) {
				if len(msgs) != 1 {
					t.Errorf("Expected 1 message, got %d", len(msgs))
				}
			},
		},
		{
			name: "Multiple parts in message",
			initialState: []llms.MessageContent{
				{
					Role: "human",
					Parts: []llms.ContentPart{
						llms.TextContent{Text: "Part 1"},
						llms.TextContent{Text: "Part 2"},
					},
				},
			},
			transform: func(msgs []llms.MessageContent) []llms.MessageContent {
				// Count total parts
				totalParts := 0
				for _, msg := range msgs {
					totalParts += len(msg.Parts)
				}
				return append(msgs, llms.TextParts("ai", fmt.Sprintf("You have %d parts", totalParts)))
			},
			validate: func(t *testing.T, msgs []llms.MessageContent) {
				if len(msgs) != 2 {
					t.Errorf("Expected 2 messages, got %d", len(msgs))
				}
				lastMsg := msgs[len(msgs)-1]
				if lastMsg.Parts[0].(llms.TextContent).Text != "You have 2 parts" {
					t.Errorf("Unexpected response: %v", lastMsg.Parts[0])
				}
			},
		},
		{
			name: "Very long message",
			initialState: []llms.MessageContent{
				llms.TextParts("human", string(make([]byte, 10000))), // 10KB message
			},
			transform: func(msgs []llms.MessageContent) []llms.MessageContent {
				// Should handle large messages
				return append(msgs, llms.TextParts("ai", "Handled large message"))
			},
			validate: func(t *testing.T, msgs []llms.MessageContent) {
				if len(msgs) != 2 {
					t.Errorf("Expected 2 messages, got %d", len(msgs))
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			g := graph.NewMessageGraph()
			g.AddNode("process", func(ctx context.Context, state interface{}) (interface{}, error) {
				msgs := state.([]llms.MessageContent)
				return tt.transform(msgs), nil
			})
			g.AddEdge("process", graph.END)
			g.SetEntryPoint("process")

			runnable, err := g.Compile()
			if err != nil {
				t.Fatalf("Failed to compile: %v", err)
			}

			result, err := runnable.Invoke(context.Background(), tt.initialState)
			if err != nil {
				t.Fatalf("Execution failed: %v", err)
			}

			tt.validate(t, result.([]llms.MessageContent))
		})
	}
}

// BenchmarkConditionalEdges benchmarks conditional edge performance
func BenchmarkConditionalEdges(b *testing.B) {
	g := graph.NewMessageGraph()

	// Create a graph with conditional routing
	g.AddNode("router", func(ctx context.Context, state interface{}) (interface{}, error) {
		return state, nil
	})

	for i := 0; i < 10; i++ {
		nodeName := fmt.Sprintf("node_%d", i)
		g.AddNode(nodeName, func(ctx context.Context, state interface{}) (interface{}, error) {
			n := state.(int)
			return n + 1, nil
		})
		g.AddEdge(nodeName, graph.END)
	}

	g.AddConditionalEdge("router", func(ctx context.Context, state interface{}) string {
		n := state.(int)
		return fmt.Sprintf("node_%d", n%10)
	})

	g.SetEntryPoint("router")

	runnable, err := g.Compile()
	if err != nil {
		b.Fatalf("Failed to compile: %v", err)
	}

	ctx := context.Background()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := runnable.Invoke(ctx, i)
		if err != nil {
			b.Fatalf("Execution failed: %v", err)
		}
	}
}

// BenchmarkLargeStateTransfer benchmarks performance with large state objects
func BenchmarkLargeStateTransfer(b *testing.B) {
	g := graph.NewMessageGraph()

	// Create nodes that pass large state
	g.AddNode("node1", func(ctx context.Context, state interface{}) (interface{}, error) {
		return state, nil
	})
	g.AddNode("node2", func(ctx context.Context, state interface{}) (interface{}, error) {
		return state, nil
	})
	g.AddNode("node3", func(ctx context.Context, state interface{}) (interface{}, error) {
		return state, nil
	})

	g.AddEdge("node1", "node2")
	g.AddEdge("node2", "node3")
	g.AddEdge("node3", graph.END)
	g.SetEntryPoint("node1")

	runnable, err := g.Compile()
	if err != nil {
		b.Fatalf("Failed to compile: %v", err)
	}

	// Create large state object (1MB)
	largeState := make([]byte, 1024*1024)
	ctx := context.Background()

	b.ResetTimer()
	b.SetBytes(int64(len(largeState)))

	for i := 0; i < b.N; i++ {
		_, err := runnable.Invoke(ctx, largeState)
		if err != nil {
			b.Fatalf("Execution failed: %v", err)
		}
	}
}
