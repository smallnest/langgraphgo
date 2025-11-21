package main

import (
	"context"
	"fmt"

	"github.com/smallnest/langgraphgo/graph"
	"github.com/tmc/langchaingo/llms"
	"github.com/tmc/langchaingo/llms/openai"
)

// DocumentState represents the state flowing through the RAG pipeline
type DocumentState struct {
	Query          string
	Documents      []string
	RelevanceScore float64
	Answer         string
	Citations      []string
}

func main() {
	// Initialize the LLM
	model, err := openai.New()
	if err != nil {
		panic(err)
	}

	// Create the graph
	g := graph.NewMessageGraph()

	// Query Classification - Route based on intent
	g.AddNode("query_classifier", func(ctx context.Context, state interface{}) (interface{}, error) {
		docs := state.(DocumentState)
		// Classify query type (factual, analytical, etc.)
		docs.Query = "classified: " + docs.Query
		return docs, nil
	})

	// Document Retrieval - Vector search
	g.AddNode("retrieve_docs", func(ctx context.Context, state interface{}) (interface{}, error) {
		docs := state.(DocumentState)
		// Perform vector similarity search
		docs.Documents = []string{"doc1", "doc2", "doc3"}
		return docs, nil
	})

	// Document Reranking - Score relevance
	g.AddNode("rerank_docs", func(ctx context.Context, state interface{}) (interface{}, error) {
		docs := state.(DocumentState)
		// Calculate relevance scores
		docs.RelevanceScore = 0.85 // Example score
		return docs, nil
	})

	// Web Search Fallback - External sources
	g.AddNode("fallback_search", func(ctx context.Context, state interface{}) (interface{}, error) {
		docs := state.(DocumentState)
		// Search external sources
		docs.Documents = append(docs.Documents, "web_result1", "web_result2")
		return docs, nil
	})

	// Generate Answer - LLM with context
	g.AddNode("generate_answer", func(ctx context.Context, state interface{}) (interface{}, error) {
		docs := state.(DocumentState)

		// Build context from documents
		context := fmt.Sprintf("Query: %s\nContext: %v", docs.Query, docs.Documents)

		messages := []llms.MessageContent{
			llms.TextParts("system", "Answer based on the provided context."),
			llms.TextParts("human", context),
		}

		response, err := model.GenerateContent(ctx, messages)
		if err != nil {
			return nil, err
		}
		docs.Answer = response.Choices[0].Content
		return docs, nil
	})

	// Format Response - Add citations
	g.AddNode("format_response", func(ctx context.Context, state interface{}) (interface{}, error) {
		docs := state.(DocumentState)
		docs.Citations = []string{"[1] doc1", "[2] doc2"}
		return docs, nil
	})

	// Build the pipeline flow
	g.SetEntryPoint("query_classifier")
	g.AddEdge("query_classifier", "retrieve_docs")
	g.AddEdge("retrieve_docs", "rerank_docs")

	// Conditional routing based on relevance
	g.AddConditionalEdge("rerank_docs", func(ctx context.Context, state interface{}) string {
		docs := state.(DocumentState)
		if docs.RelevanceScore > 0.7 {
			return "generate_answer"
		}
		return "fallback_search"
	})

	g.AddEdge("fallback_search", "generate_answer")
	g.AddEdge("generate_answer", "format_response")
	g.AddEdge("format_response", graph.END)

	// Compile and visualize
	runnable, err := g.Compile()
	if err != nil {
		panic(err)
	}

	// Export visualization
	exporter := graph.NewExporter(g)
	fmt.Println("Graph Visualization (Mermaid):")
	fmt.Println(exporter.DrawMermaid())
	fmt.Println()

	// Execute the pipeline
	result, err := runnable.Invoke(context.Background(), DocumentState{
		Query: "What are the benefits of graph-based AI pipelines?",
	})
	if err != nil {
		panic(err)
	}

	finalState := result.(DocumentState)
	fmt.Printf("Query: %s\n", finalState.Query)
	fmt.Printf("Answer: %s\n", finalState.Answer)
	fmt.Printf("Citations: %v\n", finalState.Citations)
}
