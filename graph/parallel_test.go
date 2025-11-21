package graph_test

import (
	"context"
	"fmt"
	"sync/atomic"
	"testing"
	"time"

	"github.com/smallnest/langgraphgo/graph"
)

func TestParallelNodes(t *testing.T) {
	t.Parallel()

	g := graph.NewMessageGraph()

	// Track execution order
	var counter int32

	// Add parallel nodes
	parallelFuncs := make(map[string]func(context.Context, interface{}) (interface{}, error))
	for i := 0; i < 5; i++ {
		id := fmt.Sprintf("worker_%d", i)
		parallelFuncs[id] = func(workerID string) func(context.Context, interface{}) (interface{}, error) {
			return func(ctx context.Context, state interface{}) (interface{}, error) {
				// Simulate work
				time.Sleep(10 * time.Millisecond)
				atomic.AddInt32(&counter, 1)
				return fmt.Sprintf("result_%s", workerID), nil
			}
		}(id)
	}

	g.AddParallelNodes("parallel_group", parallelFuncs)
	g.AddEdge("parallel_group", graph.END)
	g.SetEntryPoint("parallel_group")

	runnable, err := g.Compile()
	if err != nil {
		t.Fatalf("Failed to compile: %v", err)
	}

	start := time.Now()
	result, err := runnable.Invoke(context.Background(), "input")
	duration := time.Since(start)

	if err != nil {
		t.Fatalf("Execution failed: %v", err)
	}

	// Check that all workers executed
	if atomic.LoadInt32(&counter) != 5 {
		t.Errorf("Expected 5 workers to execute, got %d", counter)
	}

	// Check that execution was parallel (should be faster than sequential)
	// 5 workers with 10ms each = 50ms sequential, should be ~10ms parallel
	if duration > 30*time.Millisecond {
		t.Logf("Warning: Parallel execution took %v, might not be parallel", duration)
	}

	// Check results
	results := result.([]interface{})
	if len(results) != 5 {
		t.Errorf("Expected 5 results, got %d", len(results))
	}
}

func TestMapReduceNode(t *testing.T) {
	t.Parallel()

	g := graph.NewMessageGraph()

	// Create map functions that process parts of data
	mapFuncs := map[string]func(context.Context, interface{}) (interface{}, error){
		"map1": func(ctx context.Context, state interface{}) (interface{}, error) {
			nums := state.([]int)
			sum := 0
			for i := 0; i < len(nums)/2; i++ {
				sum += nums[i]
			}
			return sum, nil
		},
		"map2": func(ctx context.Context, state interface{}) (interface{}, error) {
			nums := state.([]int)
			sum := 0
			for i := len(nums) / 2; i < len(nums); i++ {
				sum += nums[i]
			}
			return sum, nil
		},
	}

	// Reducer function
	reducer := func(results []interface{}) (interface{}, error) {
		total := 0
		for _, r := range results {
			total += r.(int)
		}
		return total, nil
	}

	g.AddMapReduceNode("sum_parallel", mapFuncs, reducer)
	g.AddEdge("sum_parallel", graph.END)
	g.SetEntryPoint("sum_parallel")

	runnable, err := g.Compile()
	if err != nil {
		t.Fatalf("Failed to compile: %v", err)
	}

	// Test with array of numbers
	input := []int{1, 2, 3, 4, 5, 6, 7, 8, 9, 10}
	result, err := runnable.Invoke(context.Background(), input)
	if err != nil {
		t.Fatalf("Execution failed: %v", err)
	}

	// Sum of 1-10 is 55
	if result != 55 {
		t.Errorf("Expected sum of 55, got %v", result)
	}
}

func TestFanOutFanIn(t *testing.T) {
	t.Parallel()

	g := graph.NewMessageGraph()

	// Source node
	g.AddNode("source", func(ctx context.Context, state interface{}) (interface{}, error) {
		return state, nil
	})

	// Worker functions
	workers := map[string]func(context.Context, interface{}) (interface{}, error){
		"worker1": func(ctx context.Context, state interface{}) (interface{}, error) {
			n := state.(int)
			return n * 2, nil
		},
		"worker2": func(ctx context.Context, state interface{}) (interface{}, error) {
			n := state.(int)
			return n * 3, nil
		},
		"worker3": func(ctx context.Context, state interface{}) (interface{}, error) {
			n := state.(int)
			return n * 4, nil
		},
	}

	// Collector function
	collector := func(results []interface{}) (interface{}, error) {
		sum := 0
		for _, r := range results {
			sum += r.(int)
		}
		return sum, nil
	}

	g.FanOutFanIn("source", []string{"worker1", "worker2", "worker3"}, "collector", workers, collector)
	g.AddEdge("collector", graph.END)
	g.SetEntryPoint("source")

	runnable, err := g.Compile()
	if err != nil {
		t.Fatalf("Failed to compile: %v", err)
	}

	result, err := runnable.Invoke(context.Background(), 10)
	if err != nil {
		t.Fatalf("Execution failed: %v", err)
	}

	// 10*2 + 10*3 + 10*4 = 20 + 30 + 40 = 90
	if result != 90 {
		t.Errorf("Expected 90, got %v", result)
	}
}

func TestParallelErrorHandling(t *testing.T) {
	t.Parallel()

	g := graph.NewMessageGraph()

	// Add parallel nodes where one fails
	parallelFuncs := map[string]func(context.Context, interface{}) (interface{}, error){
		"success1": func(ctx context.Context, state interface{}) (interface{}, error) {
			return "ok1", nil
		},
		"failure": func(ctx context.Context, state interface{}) (interface{}, error) {
			return nil, fmt.Errorf("deliberate failure")
		},
		"success2": func(ctx context.Context, state interface{}) (interface{}, error) {
			return "ok2", nil
		},
	}

	g.AddParallelNodes("parallel_with_error", parallelFuncs)
	g.AddEdge("parallel_with_error", graph.END)
	g.SetEntryPoint("parallel_with_error")

	runnable, err := g.Compile()
	if err != nil {
		t.Fatalf("Failed to compile: %v", err)
	}

	_, err = runnable.Invoke(context.Background(), "input")
	if err == nil {
		t.Error("Expected error from parallel execution")
	}
}

func TestParallelContextCancellation(t *testing.T) {
	t.Parallel()

	g := graph.NewMessageGraph()

	// Add parallel nodes with different delays
	parallelFuncs := map[string]func(context.Context, interface{}) (interface{}, error){
		"fast": func(ctx context.Context, _ interface{}) (interface{}, error) {
			select {
			case <-time.After(10 * time.Millisecond):
				return "fast_done", nil
			case <-ctx.Done():
				return nil, ctx.Err()
			}
		},
		"slow": func(ctx context.Context, _ interface{}) (interface{}, error) {
			select {
			case <-time.After(1 * time.Second):
				return "slow_done", nil
			case <-ctx.Done():
				return nil, ctx.Err()
			}
		},
	}

	g.AddParallelNodes("parallel_cancellable", parallelFuncs)
	g.AddEdge("parallel_cancellable", graph.END)
	g.SetEntryPoint("parallel_cancellable")

	runnable, err := g.Compile()
	if err != nil {
		t.Fatalf("Failed to compile: %v", err)
	}

	// Create context with short timeout
	ctx, cancel := context.WithTimeout(context.Background(), 50*time.Millisecond)
	defer cancel()

	start := time.Now()
	_, err = runnable.Invoke(ctx, "input")
	duration := time.Since(start)

	if err == nil {
		t.Error("Expected timeout error")
	}

	// Should timeout quickly, not wait for slow node
	if duration > 100*time.Millisecond {
		t.Errorf("Cancellation took too long: %v", duration)
	}
}

func BenchmarkParallelExecution(b *testing.B) {
	g := graph.NewMessageGraph()

	// Create many parallel workers
	workers := make(map[string]func(context.Context, interface{}) (interface{}, error))
	for i := 0; i < 10; i++ {
		workerID := fmt.Sprintf("worker_%d", i)
		workers[workerID] = func(ctx context.Context, state interface{}) (interface{}, error) {
			// Simulate some work
			n := state.(int)
			result := 0
			for j := 0; j < 100; j++ {
				result += n * j
			}
			return result, nil
		}
	}

	g.AddParallelNodes("parallel", workers)
	g.AddEdge("parallel", graph.END)
	g.SetEntryPoint("parallel")

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

func BenchmarkSequentialVsParallel(b *testing.B) {
	workFunc := func(ctx context.Context, state interface{}) (interface{}, error) {
		// Simulate CPU-bound work
		n := state.(int)
		result := 0
		for i := 0; i < 1000; i++ {
			result += n * i
		}
		return result, nil
	}

	b.Run("Sequential", func(b *testing.B) {
		g := graph.NewMessageGraph()

		// Chain nodes sequentially
		for i := 0; i < 5; i++ {
			nodeName := fmt.Sprintf("node_%d", i)
			g.AddNode(nodeName, workFunc)
			if i > 0 {
				prevNode := fmt.Sprintf("node_%d", i-1)
				g.AddEdge(prevNode, nodeName)
			}
		}
		g.AddEdge("node_4", graph.END)
		g.SetEntryPoint("node_0")

		runnable, _ := g.Compile()
		ctx := context.Background()

		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			_, _ = runnable.Invoke(ctx, i)
		}
	})

	b.Run("Parallel", func(b *testing.B) {
		g := graph.NewMessageGraph()

		// Add nodes in parallel
		workers := make(map[string]func(context.Context, interface{}) (interface{}, error))
		for i := 0; i < 5; i++ {
			workers[fmt.Sprintf("worker_%d", i)] = workFunc
		}

		g.AddParallelNodes("parallel", workers)
		g.AddEdge("parallel", graph.END)
		g.SetEntryPoint("parallel")

		runnable, _ := g.Compile()
		ctx := context.Background()

		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			_, _ = runnable.Invoke(ctx, i)
		}
	})
}
