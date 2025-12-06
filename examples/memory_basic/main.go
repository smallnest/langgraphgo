package main

import (
	"context"
	"fmt"
	"log"

	"github.com/smallnest/langgraphgo/memory"
	"github.com/tmc/langchaingo/llms"
	langchainmemory "github.com/tmc/langchaingo/memory"
)

func main() {
	ctx := context.Background()

	// Example 1: Basic ConversationBuffer
	fmt.Println("=== Example 1: Basic ConversationBuffer ===")
	basicMemoryExample(ctx)

	// Example 2: ConversationWindowBuffer (keeps last N turns)
	fmt.Println("\n=== Example 2: ConversationWindowBuffer ===")
	windowMemoryExample(ctx)

	// Example 3: ChatMessageHistory with custom messages
	fmt.Println("\n=== Example 3: ChatMessageHistory ===")
	chatHistoryExample(ctx)

	// Example 4: Custom memory keys
	fmt.Println("\n=== Example 4: Custom Memory Keys ===")
	customKeysExample(ctx)

	// Example 5: Memory integration pattern
	fmt.Println("\n=== Example 5: Memory Integration Pattern ===")
	memoryIntegrationPattern(ctx)
}

// basicMemoryExample demonstrates basic conversation buffer usage
func basicMemoryExample(ctx context.Context) {
	// Create a conversation buffer memory with return messages enabled
	mem := memory.NewConversationBufferMemory(
		langchainmemory.WithReturnMessages(true),
	)

	// Simulate a conversation
	conversations := []struct {
		input  string
		output string
	}{
		{"Hello, my name is Alice", "Hi Alice! Nice to meet you."},
		{"What's my name?", "Your name is Alice."},
		{"What did I just ask you?", "You asked me what your name is."},
	}

	for _, conv := range conversations {
		// Save the conversation turn
		err := mem.SaveContext(ctx, map[string]any{
			"input": conv.input,
		}, map[string]any{
			"output": conv.output,
		})
		if err != nil {
			log.Fatalf("Failed to save context: %v", err)
		}

		fmt.Printf("User: %s\n", conv.input)
		fmt.Printf("AI: %s\n", conv.output)
	}

	// Get all messages
	messages, err := mem.GetMessages(ctx)
	if err != nil {
		log.Fatalf("Failed to get messages: %v", err)
	}

	fmt.Printf("\nTotal messages in memory: %d\n", len(messages))
	for i, msg := range messages {
		fmt.Printf("  [%d] %s: %s\n", i+1, msg.GetType(), msg.GetContent())
	}
}

// windowMemoryExample demonstrates conversation window buffer
func windowMemoryExample(ctx context.Context) {
	// Create a window buffer that keeps only the last 2 conversation turns
	mem := memory.NewConversationWindowBufferMemory(2,
		langchainmemory.WithReturnMessages(true),
	)

	// Simulate multiple conversation turns
	conversations := []struct {
		input  string
		output string
	}{
		{"First question", "First answer"},
		{"Second question", "Second answer"},
		{"Third question", "Third answer"},
		{"Fourth question", "Fourth answer"},
	}

	for i, conv := range conversations {
		err := mem.SaveContext(ctx, map[string]any{
			"input": conv.input,
		}, map[string]any{
			"output": conv.output,
		})
		if err != nil {
			log.Fatalf("Failed to save context: %v", err)
		}

		fmt.Printf("Turn %d - User: %s | AI: %s\n", i+1, conv.input, conv.output)
	}

	// Get messages - should only have the last 2 turns (4 messages)
	messages, err := mem.GetMessages(ctx)
	if err != nil {
		log.Fatalf("Failed to get messages: %v", err)
	}

	fmt.Printf("\nMessages in window (last 2 turns): %d\n", len(messages))
	for i, msg := range messages {
		fmt.Printf("  [%d] %s: %s\n", i+1, msg.GetType(), msg.GetContent())
	}
}

// chatHistoryExample demonstrates direct chat message history usage
func chatHistoryExample(ctx context.Context) {
	// Create a chat message history
	history := memory.NewChatMessageHistory()

	// Add different types of messages
	err := history.AddMessage(ctx, llms.SystemChatMessage{
		Content: "You are a helpful assistant.",
	})
	if err != nil {
		log.Fatalf("Failed to add system message: %v", err)
	}

	err = history.AddUserMessage(ctx, "Hello!")
	if err != nil {
		log.Fatalf("Failed to add user message: %v", err)
	}

	err = history.AddAIMessage(ctx, "Hi! How can I help you today?")
	if err != nil {
		log.Fatalf("Failed to add AI message: %v", err)
	}

	// Get all messages
	messages, err := history.Messages(ctx)
	if err != nil {
		log.Fatalf("Failed to get messages: %v", err)
	}

	fmt.Printf("Total messages: %d\n", len(messages))
	for i, msg := range messages {
		fmt.Printf("  [%d] %s: %s\n", i+1, msg.GetType(), msg.GetContent())
	}
}

// customKeysExample demonstrates using custom input/output/memory keys
func customKeysExample(ctx context.Context) {
	// Create memory with custom keys
	mem := memory.NewConversationBufferMemory(
		langchainmemory.WithInputKey("user_input"),
		langchainmemory.WithOutputKey("ai_response"),
		langchainmemory.WithMemoryKey("chat_history"),
		langchainmemory.WithReturnMessages(true),
		langchainmemory.WithHumanPrefix("User"),
		langchainmemory.WithAIPrefix("Assistant"),
	)

	// Save context with custom keys
	err := mem.SaveContext(ctx, map[string]any{
		"user_input": "What's the weather like?",
	}, map[string]any{
		"ai_response": "I don't have access to real-time weather data.",
	})
	if err != nil {
		log.Fatalf("Failed to save context: %v", err)
	}

	// Load memory variables
	memVars, err := mem.LoadMemoryVariables(ctx, map[string]any{})
	if err != nil {
		log.Fatalf("Failed to load memory variables: %v", err)
	}

	fmt.Printf("Memory variables keys: %v\n", getKeys(memVars))

	// Get messages
	messages, err := mem.GetMessages(ctx)
	if err != nil {
		log.Fatalf("Failed to get messages: %v", err)
	}

	fmt.Printf("Messages: %d\n", len(messages))
	for i, msg := range messages {
		fmt.Printf("  [%d] %s: %s\n", i+1, msg.GetType(), msg.GetContent())
	}
}

// memoryIntegrationPattern demonstrates how to integrate memory with a graph
func memoryIntegrationPattern(ctx context.Context) {
	fmt.Println("This example shows the pattern for integrating memory with LangGraph:")
	fmt.Println()

	// Create memory
	mem := memory.NewConversationBufferMemory(
		langchainmemory.WithReturnMessages(true),
	)

	// Simulate conversation turns
	conversations := []struct {
		input  string
		output string
	}{
		{"Hello, my name is Bob", "Hi Bob! Nice to meet you."},
		{"What's my name?", "Your name is Bob."},
	}

	for i, conv := range conversations {
		fmt.Printf("[Turn %d]\n", i+1)

		// This is what would happen in a graph node:
		// 1. Load memory
		memVars, _ := mem.LoadMemoryVariables(ctx, map[string]any{})
		var historyMessages []llms.ChatMessage
		if history, ok := memVars["history"]; ok {
			if msgs, ok := history.([]llms.ChatMessage); ok {
				historyMessages = msgs
			}
		}
		fmt.Printf("  Memory contains %d messages\n", len(historyMessages))

		// 2. Process (simulate LLM call with history + current input)
		fmt.Printf("  User: %s\n", conv.input)
		fmt.Printf("  AI: %s\n", conv.output)

		// 3. Save to memory
		mem.SaveContext(ctx, map[string]any{
			"input": conv.input,
		}, map[string]any{
			"output": conv.output,
		})

		fmt.Println()
	}

	// Show final memory state
	messages, _ := mem.GetMessages(ctx)
	fmt.Printf("Final memory contains %d messages\n", len(messages))
	fmt.Println("\nIn a real graph node, you would:")
	fmt.Println("  1. Load memory variables before LLM call")
	fmt.Println("  2. Combine history + current input")
	fmt.Println("  3. Call LLM with combined messages")
	fmt.Println("  4. Save input/output to memory")
}

// getKeys returns the keys of a map
func getKeys(m map[string]any) []string {
	keys := make([]string, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	return keys
}
