package main

import (
	"context"
	"fmt"
	"strings"

	"github.com/smallnest/langgraphgo/graph"
)

type Document struct {
	Content   string
	Validated bool
	Processed bool
}

func main() {
	// Create main graph
	main := graph.NewMessageGraph()

	// Create a validation subgraph
	validationSubgraph := graph.NewMessageGraph()
	validationSubgraph.AddNode("check_format", func(ctx context.Context, state interface{}) (interface{}, error) {
		doc := state.(Document)
		fmt.Println("  [Subgraph] Checking format...")
		// Simple validation: check if content is not empty
		if len(doc.Content) > 0 {
			doc.Validated = true
		}
		return doc, nil
	})
	validationSubgraph.AddNode("sanitize", func(ctx context.Context, state interface{}) (interface{}, error) {
		doc := state.(Document)
		fmt.Println("  [Subgraph] Sanitizing content...")
		doc.Content = strings.TrimSpace(doc.Content)
		return doc, nil
	})
	validationSubgraph.AddEdge("check_format", "sanitize")
	validationSubgraph.AddEdge("sanitize", graph.END)
	validationSubgraph.SetEntryPoint("check_format")

	// Create a processing subgraph
	processingSubgraph := graph.NewMessageGraph()
	processingSubgraph.AddNode("transform", func(ctx context.Context, state interface{}) (interface{}, error) {
		doc := state.(Document)
		fmt.Println("  [Subgraph] Transforming content...")
		doc.Content = strings.ToUpper(doc.Content)
		return doc, nil
	})
	processingSubgraph.AddNode("enrich", func(ctx context.Context, state interface{}) (interface{}, error) {
		doc := state.(Document)
		fmt.Println("  [Subgraph] Enriching content...")
		doc.Content = fmt.Sprintf("[PROCESSED] %s [END]", doc.Content)
		doc.Processed = true
		return doc, nil
	})
	processingSubgraph.AddEdge("transform", "enrich")
	processingSubgraph.AddEdge("enrich", graph.END)
	processingSubgraph.SetEntryPoint("transform")

	// Add main graph nodes
	main.AddNode("receive", func(ctx context.Context, state interface{}) (interface{}, error) {
		doc := state.(Document)
		fmt.Println("[Main] Receiving document...")
		fmt.Printf("  Initial content: %s\n", doc.Content)
		return doc, nil
	})

	// Add subgraphs to main graph
	err := main.AddSubgraph("validation", validationSubgraph)
	if err != nil {
		panic(err)
	}

	err = main.AddSubgraph("processing", processingSubgraph)
	if err != nil {
		panic(err)
	}

	main.AddNode("finalize", func(ctx context.Context, state interface{}) (interface{}, error) {
		doc := state.(Document)
		fmt.Println("[Main] Finalizing document...")
		return doc, nil
	})

	// Connect the workflow
	main.SetEntryPoint("receive")
	main.AddEdge("receive", "validation")

	// Conditional edge based on validation
	main.AddConditionalEdge("validation", func(ctx context.Context, state interface{}) string {
		doc := state.(Document)
		if doc.Validated {
			return "processing"
		}
		return "finalize" // Skip processing if not validated
	})

	main.AddEdge("processing", "finalize")
	main.AddEdge("finalize", graph.END)

	// Compile and run
	runnable, err := main.Compile()
	if err != nil {
		panic(err)
	}

	// Test with different documents
	testDocs := []Document{
		{Content: "  Hello World  "},
		{Content: ""},
		{Content: "Test Document"},
	}

	ctx := context.Background()
	for i, doc := range testDocs {
		fmt.Printf("\n=== DOCUMENT %d ===\n", i+1)
		result, err := runnable.Invoke(ctx, doc)
		if err != nil {
			fmt.Printf("Error: %v\n", err)
			continue
		}

		finalDoc := result.(Document)
		fmt.Printf("\nFinal state:\n")
		fmt.Printf("  Content: %s\n", finalDoc.Content)
		fmt.Printf("  Validated: %v\n", finalDoc.Validated)
		fmt.Printf("  Processed: %v\n", finalDoc.Processed)
	}

	// You can also create subgraphs using builder functions
	fmt.Println("\n=== USING BUILDER FUNCTION ===")

	main2 := graph.NewMessageGraph()
	err = main2.CreateSubgraph("simple_sub", func(sg *graph.MessageGraph) {
		sg.AddNode("step1", func(ctx context.Context, state interface{}) (interface{}, error) {
			return fmt.Sprintf("%v → step1", state), nil
		})
		sg.AddNode("step2", func(ctx context.Context, state interface{}) (interface{}, error) {
			return fmt.Sprintf("%v → step2", state), nil
		})
		sg.AddEdge("step1", "step2")
		sg.AddEdge("step2", graph.END)
		sg.SetEntryPoint("step1")
	})

	if err != nil {
		panic(err)
	}

	main2.AddEdge("simple_sub", graph.END)
	main2.SetEntryPoint("simple_sub")

	runnable2, err := main2.Compile()
	if err != nil {
		panic(err)
	}

	result2, err := runnable2.Invoke(ctx, "start")
	if err != nil {
		panic(err)
	}

	fmt.Printf("Builder result: %v\n", result2)
}
