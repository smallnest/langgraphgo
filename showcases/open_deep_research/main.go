package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"strings"

	"github.com/tmc/langchaingo/llms"
)

func main() {
	// Check required environment variables
	if os.Getenv("OPENAI_API_KEY") == "" {
		log.Fatal("OPENAI_API_KEY environment variable is required")
	}

	if os.Getenv("TAVILY_API_KEY") == "" {
		log.Fatal("TAVILY_API_KEY environment variable is required")
	}

	// Create configuration
	config := NewConfiguration()

	log.Println("=== Open Deep Research ===")
	log.Printf("Research Model: %s", config.ResearchModel)
	log.Printf("Final Report Model: %s", config.FinalReportModel)
	log.Printf("Max Researcher Iterations: %d", config.MaxResearcherIterations)
	log.Printf("Max Concurrent Research Units: %d", config.MaxConcurrentResearchUnits)
	log.Println()

	// Create deep researcher graph
	log.Println("Initializing Deep Researcher...")
	deepResearcher, err := CreateDeepResearcherGraph(config)
	if err != nil {
		log.Fatalf("Failed to create deep researcher: %v", err)
	}

	// Define research query
	query := `What are the latest advances in large language models in 2024? 
Focus on:
1. New model architectures and techniques
2. Improvements in reasoning capabilities
3. Advances in efficiency and scaling`

	if len(os.Args) > 1 {
		query = os.Args[1]
	}

	log.Printf("Research Query: %s\n\n", query)

	// Prepare initial state
	initialState := map[string]interface{}{
		"messages": []llms.MessageContent{
			llms.TextParts(llms.ChatMessageTypeHuman, query),
		},
	}

	// Execute research
	ctx := context.Background()
	log.Println("Starting research process...")
	log.Println("---")

	result, err := deepResearcher.Invoke(ctx, initialState)
	if err != nil {
		log.Fatalf("Research execution failed: %v", err)
	}

	// Extract final report
	resultState := result.(map[string]interface{})
	finalReport, ok := resultState["final_report"].(string)
	if !ok {
		log.Fatal("No final report generated")
	}

	// Display results
	log.Println("\n" + strings.Repeat("=", 80))
	log.Println("RESEARCH COMPLETE")
	log.Println(strings.Repeat("=", 80))
	fmt.Println()
	fmt.Println(finalReport)
	fmt.Println()
	log.Println(strings.Repeat("=", 80))

	// Display metadata
	notes, _ := resultState["notes"].([]string)
	rawNotes, _ := resultState["raw_notes"].([]string)

	log.Printf("\nMetadata:")
	log.Printf("- Research iterations: %v", resultState["research_iterations"])
	log.Printf("- Research findings collected: %d", len(notes))
	log.Printf("- Raw search results: %d", len(rawNotes))
	log.Printf("- Final report length: %d characters", len(finalReport))
}
