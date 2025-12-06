package main

import (
	"context"
	"fmt"
	"log"
	"os"

	"github.com/smallnest/goskills"
	"github.com/smallnest/langgraphgo/adapter/goskills"
	"github.com/smallnest/langgraphgo/ptc"
	"github.com/tmc/langchaingo/llms"
	"github.com/tmc/langchaingo/llms/openai"
)

// This example demonstrates how to use goskills tools with PTC
// goskills provides local tool execution (shell, python, file operations, etc.)
// PTC allows LLM to generate code that calls these tools programmatically

func main() {
	ctx := context.Background()

	// 1. Load goskills (empty skill for basic tools)
	skill := goskills.SkillPackage{
		Name:        "basic_tools",
		Description: "Basic system tools",
		Path:        "",
	}

	// 2. Convert goskills to tools.Tool interface
	tools, err := goskills.SkillsToTools(skill)
	if err != nil {
		log.Fatalf("Failed to convert skills to tools: %v", err)
	}

	fmt.Printf("Loaded %d tools from goskills:\n", len(tools))
	for _, tool := range tools {
		fmt.Printf("  - %s: %s\n", tool.Name(), tool.Description())
	}

	// 3. Create LLM
	apiKey := os.Getenv("OPENAI_API_KEY")
	if apiKey == "" {
		log.Fatal("OPENAI_API_KEY environment variable not set")
	}

	model, err := openai.New()
	if err != nil {
		log.Fatalf("Failed to create LLM: %v", err)
	}

	// 4. Create PTC agent with goskills tools
	agent, err := ptc.CreatePTCAgent(ptc.PTCAgentConfig{
		Model:         model,
		Tools:         tools,
		Language:      ptc.LanguagePython, // Use Python for code generation
		ExecutionMode: ptc.ModeDirect,     // Direct mode for efficiency
		MaxIterations: 10,
		SystemPrompt:  "You are a helpful assistant with access to system tools. Use the tools to help users with their tasks.",
	})
	if err != nil {
		log.Fatalf("Failed to create PTC agent: %v", err)
	}

	// 5. Run a query that requires tool usage
	query := "Create a file named 'test.txt' with the content 'Hello from PTC + goskills!' and then read it back to verify."

	initialState := map[string]interface{}{
		"messages": []llms.MessageContent{
			{
				Role: llms.ChatMessageTypeHuman,
				Parts: []llms.ContentPart{
					llms.TextPart(query),
				},
			},
		},
	}

	fmt.Printf("\nQuery: %s\n\n", query)
	fmt.Println("Processing...")

	// 6. Execute the agent
	result, err := agent.Invoke(ctx, initialState)
	if err != nil {
		log.Fatalf("Agent execution failed: %v", err)
	}

	// 7. Display results
	messages := result.(map[string]interface{})["messages"].([]llms.MessageContent)
	fmt.Println("\n=== Execution Results ===\n")

	for i, msg := range messages {
		role := msg.Role
		fmt.Printf("[Message %d - %s]\n", i+1, role)

		for _, part := range msg.Parts {
			if textPart, ok := part.(llms.TextContent); ok {
				fmt.Println(textPart.Text)
			}
		}
		fmt.Println()
	}

	// The agent will:
	// 1. Generate Python code to call write_file tool
	// 2. Execute the code through PTC
	// 3. goskills adapter will handle actual file writing
	// 4. Generate code to call read_file tool
	// 5. Verify the content and respond to user
}
