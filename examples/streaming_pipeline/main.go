package main

import (
	"context"
	"fmt"
	"time"

	"github.com/smallnest/langgraphgo/graph"
)

func main() {
	// Create streaming graph
	g := graph.NewStreamingMessageGraph()

	// Add nodes with listeners
	analyze := g.AddNode("analyze", func(ctx context.Context, state interface{}) (interface{}, error) {
		time.Sleep(100 * time.Millisecond) // Simulate processing
		return fmt.Sprintf("analyzed_%v", state), nil
	})

	enhance := g.AddNode("enhance", func(ctx context.Context, state interface{}) (interface{}, error) {
		time.Sleep(200 * time.Millisecond) // Simulate processing
		return fmt.Sprintf("enhanced_%v", state), nil
	})

	// Build pipeline
	g.AddEdge("analyze", "enhance")
	g.AddEdge("enhance", graph.END)
	g.SetEntryPoint("analyze")

	// Add real-time listeners
	progressListener := graph.NewProgressListener()
	chatListener := graph.NewChatListener()
	metricsListener := graph.NewMetricsListener()

	analyze.AddListener(progressListener)
	analyze.AddListener(chatListener)
	analyze.AddListener(metricsListener)

	enhance.AddListener(progressListener)
	enhance.AddListener(chatListener)
	enhance.AddListener(metricsListener)

	// Compile and execute with streaming
	streamingRunnable, err := g.CompileStreaming()
	if err != nil {
		panic(err)
	}

	executor := graph.NewStreamingExecutor(streamingRunnable)

	err = executor.ExecuteWithCallback(
		context.Background(),
		"input_document",
		// Event callback - receives real-time updates
		func(event graph.StreamEvent) {
			fmt.Printf("[%s] Event: %s from %s\n",
				time.Now().Format("15:04:05.000"),
				event.Event,
				event.NodeName)
		},
		// Result callback - receives final result
		func(result interface{}, err error) {
			if err != nil {
				fmt.Printf("Error: %v\n", err)
			} else {
				fmt.Printf("Final result: %v\n", result)
			}
		},
	)

	if err != nil {
		panic(err)
	}

	// Print metrics summary
	fmt.Println("\nMetrics Summary:")
	metricsListener.PrintSummary(nil)
}
