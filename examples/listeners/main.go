package main

import (
	"context"
	"fmt"
	"time"

	"github.com/smallnest/langgraphgo/graph"
)

func main() {
	// Create a graph with listener support
	g := graph.NewListenableMessageGraph()

	// Create different types of listeners
	progressListener := graph.NewProgressListener().
		WithTiming(true).
		WithDetails(true)

	metricsListener := graph.NewMetricsListener()

	chatListener := graph.NewChatListener()
	chatListener.SetNodeMessage("process", "🤖 Processing your data...")
	chatListener.SetNodeMessage("analyze", "🔍 Analyzing results...")
	chatListener.SetNodeMessage("report", "📊 Generating report...")

	loggingListener := graph.NewLoggingListener().
		WithLogLevel(graph.LogLevelInfo).
		WithState(true)

	// Add nodes with different processing times
	processNode := g.AddNode("process", func(ctx context.Context, state interface{}) (interface{}, error) {
		fmt.Println("  [Node] Starting data processing...")
		time.Sleep(300 * time.Millisecond) // Simulate work
		return fmt.Sprintf("%v → processed", state), nil
	})

	analyzeNode := g.AddNode("analyze", func(ctx context.Context, state interface{}) (interface{}, error) {
		fmt.Println("  [Node] Analyzing data...")
		time.Sleep(200 * time.Millisecond) // Simulate work
		return fmt.Sprintf("%v → analyzed", state), nil
	})

	reportNode := g.AddNode("report", func(ctx context.Context, state interface{}) (interface{}, error) {
		fmt.Println("  [Node] Creating report...")
		time.Sleep(100 * time.Millisecond) // Simulate work
		return fmt.Sprintf("%v → reported", state), nil
	})

	// Attach listeners to nodes
	processNode.AddListener(progressListener)
	processNode.AddListener(metricsListener)
	processNode.AddListener(chatListener)
	processNode.AddListener(loggingListener)

	analyzeNode.AddListener(progressListener)
	analyzeNode.AddListener(metricsListener)
	analyzeNode.AddListener(chatListener)
	analyzeNode.AddListener(loggingListener)

	reportNode.AddListener(progressListener)
	reportNode.AddListener(metricsListener)
	reportNode.AddListener(chatListener)
	reportNode.AddListener(loggingListener)

	// Set up node steps for progress listener
	progressListener.SetNodeStep("process", "Processing input data")
	progressListener.SetNodeStep("analyze", "Analyzing processed data")
	progressListener.SetNodeStep("report", "Generating final report")

	// Build the pipeline
	g.SetEntryPoint("process")
	g.AddEdge("process", "analyze")
	g.AddEdge("analyze", "report")
	g.AddEdge("report", graph.END)

	// Compile with listener support
	runnable, err := g.CompileListenable()
	if err != nil {
		panic(err)
	}

	fmt.Println("=== EXECUTING WITH LISTENERS ===")

	// Execute the graph
	ctx := context.Background()
	startTime := time.Now()

	result, err := runnable.Invoke(ctx, "raw_data")
	if err != nil {
		panic(err)
	}

	duration := time.Since(startTime)

	fmt.Printf("\n=== EXECUTION COMPLETED ===\n")
	fmt.Printf("Final result: %v\n", result)
	fmt.Printf("Total duration: %v\n\n", duration)

	// Print metrics summary
	fmt.Println("=== METRICS SUMMARY ===")
	fmt.Printf("Total executions: %d\n", metricsListener.GetTotalExecutions())

	nodeExecutions := metricsListener.GetNodeExecutions()
	avgDurations := metricsListener.GetNodeAverageDuration()
	for node, count := range nodeExecutions {
		avgDuration := avgDurations[node]
		fmt.Printf("  %s: %d executions, avg duration: %v\n", node, count, avgDuration)
	}

	// Demonstrate running multiple times to see metrics accumulate
	fmt.Println("\n=== RUNNING 3 MORE TIMES ===")
	for i := 1; i <= 3; i++ {
		fmt.Printf("\nRun %d:\n", i)
		_, err := runnable.Invoke(ctx, fmt.Sprintf("data_%d", i))
		if err != nil {
			fmt.Printf("Error: %v\n", err)
		}
	}

	fmt.Println("\n=== UPDATED METRICS ===")
	fmt.Printf("Total executions: %d\n", metricsListener.GetTotalExecutions())

	nodeExecutions = metricsListener.GetNodeExecutions()
	avgDurations = metricsListener.GetNodeAverageDuration()
	for node, count := range nodeExecutions {
		avgDuration := avgDurations[node]
		fmt.Printf("  %s: %d executions, avg: %v\n",
			node, count, avgDuration)
	}
}
