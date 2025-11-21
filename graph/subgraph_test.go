package graph_test

import (
	"context"
	"fmt"
	"strings"
	"testing"

	"github.com/smallnest/langgraphgo/graph"
)

func TestSubgraph(t *testing.T) {
	t.Parallel()

	// Create main graph
	main := graph.NewMessageGraph()

	// Create a subgraph
	subgraph := graph.NewMessageGraph()
	subgraph.AddNode("sub_process", func(ctx context.Context, state interface{}) (interface{}, error) {
		s := state.(string)
		return s + "_processed_by_subgraph", nil
	})
	subgraph.AddEdge("sub_process", graph.END)
	subgraph.SetEntryPoint("sub_process")

	// Add regular node
	main.AddNode("pre_process", func(ctx context.Context, state interface{}) (interface{}, error) {
		s := state.(string)
		return s + "_preprocessed", nil
	})

	// Add subgraph to main
	err := main.AddSubgraph("subgraph_node", subgraph)
	if err != nil {
		t.Fatalf("Failed to add subgraph: %v", err)
	}

	// Add post-process node
	main.AddNode("post_process", func(ctx context.Context, state interface{}) (interface{}, error) {
		s := state.(string)
		return s + "_postprocessed", nil
	})

	// Connect nodes
	main.AddEdge("pre_process", "subgraph_node")
	main.AddEdge("subgraph_node", "post_process")
	main.AddEdge("post_process", graph.END)
	main.SetEntryPoint("pre_process")

	// Compile and run
	runnable, err := main.Compile()
	if err != nil {
		t.Fatalf("Failed to compile: %v", err)
	}

	result, err := runnable.Invoke(context.Background(), "input")
	if err != nil {
		t.Fatalf("Execution failed: %v", err)
	}

	expected := "input_preprocessed_processed_by_subgraph_postprocessed"
	if result != expected {
		t.Errorf("Expected %s, got %v", expected, result)
	}
}

func TestCreateSubgraph(t *testing.T) {
	t.Parallel()

	main := graph.NewMessageGraph()

	// Create subgraph using builder function
	err := main.CreateSubgraph("validation_subgraph", func(sg *graph.MessageGraph) {
		sg.AddNode("validate", func(ctx context.Context, state interface{}) (interface{}, error) {
			data := state.(map[string]interface{})
			if val, ok := data["value"].(int); ok && val > 0 {
				data["valid"] = true
			} else {
				data["valid"] = false
			}
			return data, nil
		})

		sg.AddNode("format", func(ctx context.Context, state interface{}) (interface{}, error) {
			data := state.(map[string]interface{})
			data["formatted"] = true
			return data, nil
		})

		sg.AddEdge("validate", "format")
		sg.AddEdge("format", graph.END)
		sg.SetEntryPoint("validate")
	})

	if err != nil {
		t.Fatalf("Failed to create subgraph: %v", err)
	}

	main.AddEdge("validation_subgraph", graph.END)
	main.SetEntryPoint("validation_subgraph")

	runnable, err := main.Compile()
	if err != nil {
		t.Fatalf("Failed to compile: %v", err)
	}

	input := map[string]interface{}{"value": 42}
	result, err := runnable.Invoke(context.Background(), input)
	if err != nil {
		t.Fatalf("Execution failed: %v", err)
	}

	data := result.(map[string]interface{})
	if !data["valid"].(bool) || !data["formatted"].(bool) {
		t.Errorf("Expected valid and formatted to be true, got %v", data)
	}
}

func TestNestedSubgraphs(t *testing.T) {
	t.Parallel()

	// Create innermost subgraph
	inner := graph.NewMessageGraph()
	inner.AddNode("inner_node", func(ctx context.Context, state interface{}) (interface{}, error) {
		s := state.(string)
		return s + "_inner", nil
	})
	inner.AddEdge("inner_node", graph.END)
	inner.SetEntryPoint("inner_node")

	// Create middle subgraph
	middle := graph.NewMessageGraph()
	middle.AddNode("middle_start", func(ctx context.Context, state interface{}) (interface{}, error) {
		s := state.(string)
		return s + "_middle", nil
	})

	// Add inner subgraph to middle
	err := middle.AddSubgraph("inner_subgraph", inner)
	if err != nil {
		t.Fatalf("Failed to add inner subgraph: %v", err)
	}

	middle.AddEdge("middle_start", "inner_subgraph")
	middle.AddEdge("inner_subgraph", graph.END)
	middle.SetEntryPoint("middle_start")

	// Create main graph
	main := graph.NewMessageGraph()
	main.AddNode("main_start", func(ctx context.Context, state interface{}) (interface{}, error) {
		s := state.(string)
		return s + "_main", nil
	})

	// Add middle subgraph to main
	err = main.AddSubgraph("middle_subgraph", middle)
	if err != nil {
		t.Fatalf("Failed to add middle subgraph: %v", err)
	}

	main.AddEdge("main_start", "middle_subgraph")
	main.AddEdge("middle_subgraph", graph.END)
	main.SetEntryPoint("main_start")

	// Compile and run
	runnable, err := main.Compile()
	if err != nil {
		t.Fatalf("Failed to compile: %v", err)
	}

	result, err := runnable.Invoke(context.Background(), "start")
	if err != nil {
		t.Fatalf("Execution failed: %v", err)
	}

	expected := "start_main_middle_inner"
	if result != expected {
		t.Errorf("Expected %s, got %v", expected, result)
	}
}

func TestRecursiveSubgraph(t *testing.T) {
	t.Parallel()

	main := graph.NewMessageGraph()

	// Add recursive subgraph that counts down
	main.AddRecursiveSubgraph(
		"countdown",
		5, // max depth
		func(state interface{}, depth int) bool {
			n := state.(int)
			return n > 0 && depth < 5
		},
		func(sg *graph.MessageGraph) {
			sg.AddNode("decrement", func(ctx context.Context, state interface{}) (interface{}, error) {
				n := state.(int)
				return n - 1, nil
			})
			sg.AddEdge("decrement", graph.END)
			sg.SetEntryPoint("decrement")
		},
	)

	main.AddEdge("countdown", graph.END)
	main.SetEntryPoint("countdown")

	runnable, err := main.Compile()
	if err != nil {
		t.Fatalf("Failed to compile: %v", err)
	}

	// Start with 3, should decrement to 0 (but limited by condition)
	result, err := runnable.Invoke(context.Background(), 3)
	if err != nil {
		t.Fatalf("Execution failed: %v", err)
	}

	if result != 0 {
		t.Errorf("Expected countdown to reach 0, got %v", result)
	}

	// Test max depth limit
	result, err = runnable.Invoke(context.Background(), 10)
	if err != nil {
		t.Fatalf("Execution failed: %v", err)
	}

	// Should stop at 5 due to max depth
	if result != 5 {
		t.Errorf("Expected countdown to stop at 5 (max depth), got %v", result)
	}
}

func TestNestedConditionalSubgraph(t *testing.T) {
	t.Parallel()

	main := graph.NewMessageGraph()

	// Create different subgraphs for different types
	stringProcessor := graph.NewMessageGraph()
	stringProcessor.AddNode("uppercase", func(ctx context.Context, state interface{}) (interface{}, error) {
		s := state.(map[string]interface{})["data"].(string)
		s = strings.ToUpper(s)
		return map[string]interface{}{"type": "string", "data": s, "processed": true}, nil
	})
	stringProcessor.AddEdge("uppercase", graph.END)
	stringProcessor.SetEntryPoint("uppercase")

	numberProcessor := graph.NewMessageGraph()
	numberProcessor.AddNode("double", func(ctx context.Context, state interface{}) (interface{}, error) {
		m := state.(map[string]interface{})
		n := m["data"].(int)
		return map[string]interface{}{"type": "number", "data": n * 2, "processed": true}, nil
	})
	numberProcessor.AddEdge("double", graph.END)
	numberProcessor.SetEntryPoint("double")

	// Router function
	router := func(state interface{}) string {
		m := state.(map[string]interface{})
		dataType := m["type"].(string)
		switch dataType {
		case "string":
			return "string_processor"
		case "number":
			return "number_processor"
		default:
			return "string_processor"
		}
	}

	// Add nested conditional subgraph
	err := main.AddNestedConditionalSubgraph(
		"type_processor",
		router,
		map[string]*graph.MessageGraph{
			"string_processor": stringProcessor,
			"number_processor": numberProcessor,
		},
	)
	if err != nil {
		t.Fatalf("Failed to add nested conditional subgraph: %v", err)
	}

	main.AddEdge("type_processor", graph.END)
	main.SetEntryPoint("type_processor")

	runnable, err := main.Compile()
	if err != nil {
		t.Fatalf("Failed to compile: %v", err)
	}

	// Test with string
	stringInput := map[string]interface{}{
		"type": "string",
		"data": "hello",
	}
	result, err := runnable.Invoke(context.Background(), stringInput)
	if err != nil {
		t.Fatalf("String execution failed: %v", err)
	}

	stringResult := result.(map[string]interface{})
	if stringResult["data"] != "HELLO" {
		t.Errorf("Expected uppercase HELLO, got %v", stringResult["data"])
	}

	// Test with number
	numberInput := map[string]interface{}{
		"type": "number",
		"data": 21,
	}
	result, err = runnable.Invoke(context.Background(), numberInput)
	if err != nil {
		t.Fatalf("Number execution failed: %v", err)
	}

	numberResult := result.(map[string]interface{})
	if numberResult["data"] != 42 {
		t.Errorf("Expected doubled value 42, got %v", numberResult["data"])
	}
}

func TestCompositeGraph(t *testing.T) {
	t.Parallel()

	composite := graph.NewCompositeGraph()

	// Create first graph
	graph1 := graph.NewMessageGraph()
	graph1.AddNode("step1", func(ctx context.Context, state interface{}) (interface{}, error) {
		return state.(int) + 10, nil
	})
	graph1.AddEdge("step1", graph.END)
	graph1.SetEntryPoint("step1")

	// Create second graph
	graph2 := graph.NewMessageGraph()
	graph2.AddNode("step2", func(ctx context.Context, state interface{}) (interface{}, error) {
		return state.(int) * 2, nil
	})
	graph2.AddEdge("step2", graph.END)
	graph2.SetEntryPoint("step2")

	// Add graphs to composite
	composite.AddGraph("adder", graph1)
	composite.AddGraph("multiplier", graph2)

	// Connect graphs with transformation
	err := composite.Connect("adder", "step1", "multiplier", "step2", func(state interface{}) interface{} {
		// Transform between graphs if needed
		return state
	})
	if err != nil {
		t.Fatalf("Failed to connect graphs: %v", err)
	}

	// This test primarily validates the composite graph structure
	// Full execution would require more complex setup
}

func BenchmarkSubgraphExecution(b *testing.B) {
	main := graph.NewMessageGraph()

	// Create a subgraph
	subgraph := graph.NewMessageGraph()
	subgraph.AddNode("process", func(ctx context.Context, state interface{}) (interface{}, error) {
		n := state.(int)
		return n * 2, nil
	})
	subgraph.AddEdge("process", graph.END)
	subgraph.SetEntryPoint("process")

	// Add multiple subgraphs
	for i := 0; i < 5; i++ {
		sg := graph.NewMessageGraph()
		sg.AddNode("process", func(ctx context.Context, state interface{}) (interface{}, error) {
			n := state.(int)
			return n + i, nil
		})
		sg.AddEdge("process", graph.END)
		sg.SetEntryPoint("process")

		nodeName := fmt.Sprintf("subgraph_%d", i)
		_ = main.AddSubgraph(nodeName, sg)

		if i > 0 {
			prevNode := fmt.Sprintf("subgraph_%d", i-1)
			main.AddEdge(prevNode, nodeName)
		}
	}

	main.AddEdge("subgraph_4", graph.END)
	main.SetEntryPoint("subgraph_0")

	runnable, err := main.Compile()
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
