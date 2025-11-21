package main

import (
	"context"
	"fmt"
	"time"

	"github.com/smallnest/langgraphgo/graph"
)

// Simple example demonstrating all major features
func main() {
	fmt.Println("🚀 LangGraphGo Basic Example")
	fmt.Println("============================")

	runBasicExample()
	runStreamingExample()
	runCheckpointingExample()
	runVisualizationExample()
}

func runBasicExample() {
	fmt.Println("\n1️⃣ Basic Graph Execution")

	g := graph.NewMessageGraph()

	g.AddNode("process", func(ctx context.Context, state interface{}) (interface{}, error) {
		input := state.(string)
		return fmt.Sprintf("processed_%s", input), nil
	})

	g.AddEdge("process", graph.END)
	g.SetEntryPoint("process")

	runnable, _ := g.Compile()
	result, _ := runnable.Invoke(context.Background(), "input")

	fmt.Printf("   Result: %s\n", result)
}

func runStreamingExample() {
	fmt.Println("\n2️⃣ Streaming with Listeners")

	g := graph.NewListenableMessageGraph()

	node := g.AddNode("stream_process", func(ctx context.Context, state interface{}) (interface{}, error) {
		time.Sleep(100 * time.Millisecond) // Simulate work
		return fmt.Sprintf("streamed_%v", state), nil
	})

	g.AddEdge("stream_process", graph.END)
	g.SetEntryPoint("stream_process")

	// Add progress listener
	progressListener := graph.NewProgressListener().WithTiming(false)
	progressListener.SetNodeStep("stream_process", "🔄 Processing with streaming")
	node.AddListener(progressListener)

	runnable, _ := g.CompileListenable()
	result, _ := runnable.Invoke(context.Background(), "stream_input")

	fmt.Printf("   Streamed Result: %s\n", result)
}

func runCheckpointingExample() {
	fmt.Println("\n3️⃣ Checkpointing Example")

	g := graph.NewCheckpointableMessageGraph()

	g.AddNode("checkpoint_step1", func(ctx context.Context, state interface{}) (interface{}, error) {
		data := state.(map[string]interface{})
		data["step1"] = "completed"
		return data, nil
	})

	g.AddNode("checkpoint_step2", func(ctx context.Context, state interface{}) (interface{}, error) {
		data := state.(map[string]interface{})
		data["step2"] = "completed"
		return data, nil
	})

	g.AddEdge("checkpoint_step1", "checkpoint_step2")
	g.AddEdge("checkpoint_step2", graph.END)
	g.SetEntryPoint("checkpoint_step1")

	// Configure checkpointing
	config := graph.CheckpointConfig{
		Store:          graph.NewMemoryCheckpointStore(),
		AutoSave:       true,
		MaxCheckpoints: 5,
	}
	g.SetCheckpointConfig(config)

	runnable, _ := g.CompileCheckpointable()

	initialState := map[string]interface{}{
		"input": "checkpoint_test",
	}

	result, _ := runnable.Invoke(context.Background(), initialState)

	// Wait for async checkpoints
	time.Sleep(100 * time.Millisecond)

	checkpoints, _ := runnable.ListCheckpoints(context.Background())
	fmt.Printf("   Final State: %v\n", result)
	fmt.Printf("   Created %d checkpoints\n", len(checkpoints))
}

func runVisualizationExample() {
	fmt.Println("\n4️⃣ Graph Visualization")

	g := graph.NewMessageGraph()

	g.AddNode("visualize_step1", func(ctx context.Context, state interface{}) (interface{}, error) {
		return state, nil
	})

	g.AddNode("visualize_step2", func(ctx context.Context, state interface{}) (interface{}, error) {
		return state, nil
	})

	g.AddEdge("visualize_step1", "visualize_step2")
	g.AddEdge("visualize_step2", graph.END)
	g.SetEntryPoint("visualize_step1")

	runnable, _ := g.Compile()
	exporter := runnable.GetGraph()

	fmt.Println("   📊 Mermaid Diagram:")
	mermaid := exporter.DrawMermaid()
	fmt.Printf("      %s\n", mermaid[:100]+"...")

	fmt.Println("   🌳 ASCII Tree:")
	ascii := exporter.DrawASCII()
	fmt.Printf("      %s\n", ascii[:50]+"...")

	fmt.Println("\n✅ All examples completed successfully!")
}
