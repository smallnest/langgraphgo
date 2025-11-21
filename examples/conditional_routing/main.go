package main

import (
	"context"
	"fmt"

	"github.com/smallnest/langgraphgo/graph"
)

type Task struct {
	Priority string
	Content  string
	Result   string
}

func main() {
	g := graph.NewMessageGraph()

	// Router node - analyzes the task
	g.AddNode("router", func(ctx context.Context, state interface{}) (interface{}, error) {
		task := state.(Task)
		fmt.Printf("Routing task with priority: %s\n", task.Priority)
		return task, nil
	})

	// High priority handler
	g.AddNode("urgent_handler", func(ctx context.Context, state interface{}) (interface{}, error) {
		task := state.(Task)
		task.Result = fmt.Sprintf("URGENT: Handled %s immediately", task.Content)
		fmt.Println("→ Handled by urgent handler")
		return task, nil
	})

	// Normal priority handler
	g.AddNode("normal_handler", func(ctx context.Context, state interface{}) (interface{}, error) {
		task := state.(Task)
		task.Result = fmt.Sprintf("Normal: Processed %s in queue", task.Content)
		fmt.Println("→ Handled by normal handler")
		return task, nil
	})

	// Low priority handler
	g.AddNode("batch_handler", func(ctx context.Context, state interface{}) (interface{}, error) {
		task := state.(Task)
		task.Result = fmt.Sprintf("Batch: Queued %s for later", task.Content)
		fmt.Println("→ Handled by batch handler")
		return task, nil
	})

	// Set up the flow
	g.SetEntryPoint("router")

	// Add conditional routing based on priority
	g.AddConditionalEdge("router", func(ctx context.Context, state interface{}) string {
		task := state.(Task)
		switch task.Priority {
		case "high", "urgent":
			return "urgent_handler"
		case "low":
			return "batch_handler"
		default:
			return "normal_handler"
		}
	})

	// All handlers lead to END
	g.AddEdge("urgent_handler", graph.END)
	g.AddEdge("normal_handler", graph.END)
	g.AddEdge("batch_handler", graph.END)

	// Compile the graph
	runnable, err := g.Compile()
	if err != nil {
		panic(err)
	}

	// Test with different priorities
	testCases := []Task{
		{Priority: "high", Content: "Fix production bug"},
		{Priority: "normal", Content: "Update documentation"},
		{Priority: "low", Content: "Refactor old code"},
		{Priority: "urgent", Content: "Security patch"},
	}

	ctx := context.Background()
	for _, task := range testCases {
		fmt.Printf("\n--- Processing: %s ---\n", task.Content)
		result, err := runnable.Invoke(ctx, task)
		if err != nil {
			fmt.Printf("Error: %v\n", err)
			continue
		}
		finalTask := result.(Task)
		fmt.Printf("Result: %s\n", finalTask.Result)
	}
}
