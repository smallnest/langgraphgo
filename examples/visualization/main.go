package main

import (
	"context"
	"fmt"

	"github.com/smallnest/langgraphgo/graph"
)

func main() {
	// Create a graph representing a data processing pipeline
	g := graph.NewMessageGraph()

	// Add nodes representing different processing stages
	g.AddNode("validate_input", func(ctx context.Context, state interface{}) (interface{}, error) {
		return fmt.Sprintf("%v → validated", state), nil
	})

	g.AddNode("fetch_data", func(ctx context.Context, state interface{}) (interface{}, error) {
		return fmt.Sprintf("%v → fetched", state), nil
	})

	g.AddNode("transform", func(ctx context.Context, state interface{}) (interface{}, error) {
		return fmt.Sprintf("%v → transformed", state), nil
	})

	g.AddNode("enrich", func(ctx context.Context, state interface{}) (interface{}, error) {
		return fmt.Sprintf("%v → enriched", state), nil
	})

	g.AddNode("validate_output", func(ctx context.Context, state interface{}) (interface{}, error) {
		return fmt.Sprintf("%v → output_validated", state), nil
	})

	g.AddNode("save", func(ctx context.Context, state interface{}) (interface{}, error) {
		return fmt.Sprintf("%v → saved", state), nil
	})

	g.AddNode("notify", func(ctx context.Context, state interface{}) (interface{}, error) {
		return fmt.Sprintf("%v → notified", state), nil
	})

	// Build the pipeline flow
	g.SetEntryPoint("validate_input")
	g.AddEdge("validate_input", "fetch_data")
	g.AddEdge("fetch_data", "transform")

	// Add conditional edge for data quality check
	g.AddConditionalEdge("transform", func(ctx context.Context, state interface{}) string {
		// In a real scenario, this would check data quality
		return "enrich" // Always enrich for this example
	})

	g.AddEdge("enrich", "validate_output")
	g.AddEdge("validate_output", "save")
	g.AddEdge("save", "notify")
	g.AddEdge("notify", graph.END)

	// Compile the graph
	runnable, err := g.Compile()
	if err != nil {
		panic(err)
	}

	// Create graph exporter for visualization
	exporter := runnable.GetGraph()

	fmt.Println("=== MERMAID DIAGRAM ===")
	fmt.Println("Copy this to view in Mermaid Live Editor (https://mermaid.live/)")
	fmt.Println()
	mermaid := exporter.DrawMermaid()
	fmt.Println(mermaid)
	fmt.Println()

	fmt.Println("=== DOT FORMAT (Graphviz) ===")
	fmt.Println("Save this as .dot file and use 'dot -Tpng file.dot -o graph.png' to generate image")
	fmt.Println()
	dot := exporter.DrawDOT()
	fmt.Println(dot)
	fmt.Println()

	fmt.Println("=== ASCII VISUALIZATION ===")
	ascii := exporter.DrawASCII()
	fmt.Println(ascii)
	fmt.Println()

	// Execute the graph to show it works
	fmt.Println("=== EXECUTING GRAPH ===")
	ctx := context.Background()
	result, err := runnable.Invoke(ctx, "input_data")
	if err != nil {
		panic(err)
	}

	fmt.Printf("Final result: %v\n", result)
}
