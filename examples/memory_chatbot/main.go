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

	fmt.Println("=== Memory-Enabled Chatbot Simulation ===\n")

	// Example 1: Chatbot with ConversationBuffer
	fmt.Println("--- Example 1: Full Memory (ConversationBuffer) ---")
	runChatbotSimulation(ctx, "buffer", 2)

	// Example 2: Chatbot with ConversationWindowBuffer
	fmt.Println("\n--- Example 2: Window Memory (Last 2 Turns) ---")
	runChatbotSimulation(ctx, "window", 2)

	// Example 3: Demonstrate memory persistence
	fmt.Println("\n--- Example 3: Memory Persistence ---")
	demonstrateMemoryPersistence(ctx)
}

// runChatbotSimulation simulates a chatbot conversation with memory
func runChatbotSimulation(ctx context.Context, memoryType string, windowSize int) {
	// Create memory based on type
	var mem *memory.LangChainMemory
	switch memoryType {
	case "buffer":
		mem = memory.NewConversationBufferMemory(
			langchainmemory.WithReturnMessages(true),
		)
	case "window":
		mem = memory.NewConversationWindowBufferMemory(windowSize,
			langchainmemory.WithReturnMessages(true),
		)
	default:
		mem = memory.NewConversationBufferMemory(
			langchainmemory.WithReturnMessages(true),
		)
	}

	// Simulate conversation
	conversations := []struct {
		input  string
		output string
	}{
		{
			"Hello! My name is Alice and I love programming in Go.",
			"Hi Alice! It's great to meet you. Go is an excellent programming language! What aspects of Go do you enjoy most?",
		},
		{
			"What's my name?",
			"Your name is Alice, as you mentioned in your first message.",
		},
		{
			"What programming language do I like?",
			"You mentioned that you love programming in Go!",
		},
		{
			"Can you remind me what we talked about?",
			"We talked about your name (Alice) and your love for Go programming. You asked me to confirm your name and the programming language you like.",
		},
	}

	for i, conv := range conversations {
		fmt.Printf("\n[Turn %d]\n", i+1)

		// Load memory before processing
		memVars, err := mem.LoadMemoryVariables(ctx, map[string]any{})
		if err != nil {
			log.Printf("Error loading memory: %v\n", err)
			continue
		}

		// Get historical messages
		var historyMessages []llms.ChatMessage
		if history, ok := memVars["history"]; ok {
			if msgs, ok := history.([]llms.ChatMessage); ok {
				historyMessages = msgs
			}
		}

		fmt.Printf("Memory: %d previous messages\n", len(historyMessages))
		fmt.Printf("User: %s\n", conv.input)
		fmt.Printf("Bot: %s\n", conv.output)

		// Save to memory
		err = mem.SaveContext(ctx, map[string]any{
			"input": conv.input,
		}, map[string]any{
			"output": conv.output,
		})
		if err != nil {
			log.Printf("Error saving context: %v\n", err)
		}
	}

	// Show final memory state
	fmt.Println("\n--- Final Memory State ---")
	messages, err := mem.GetMessages(ctx)
	if err != nil {
		log.Printf("Error getting messages: %v\n", err)
		return
	}

	fmt.Printf("Total messages in memory: %d\n", len(messages))
	for i, msg := range messages {
		content := msg.GetContent()
		if len(content) > 60 {
			content = content[:60] + "..."
		}
		fmt.Printf("  [%d] %s: %s\n", i+1, msg.GetType(), content)
	}
}

// demonstrateMemoryPersistence shows how memory persists across interactions
func demonstrateMemoryPersistence(ctx context.Context) {
	// Create a custom chat history
	chatHistory := memory.NewChatMessageHistory()

	// Add initial messages
	chatHistory.AddUserMessage(ctx, "I'm learning about LangGraph")
	chatHistory.AddAIMessage(ctx, "That's great! LangGraph is a powerful framework for building stateful, multi-actor applications.")

	// Create memory with the custom chat history
	mem := memory.NewConversationBufferMemory(
		langchainmemory.WithChatHistory(chatHistory.GetHistory()),
		langchainmemory.WithReturnMessages(true),
	)

	fmt.Println("Initial memory loaded from chat history:")
	messages, _ := mem.GetMessages(ctx)
	for i, msg := range messages {
		fmt.Printf("  [%d] %s: %s\n", i+1, msg.GetType(), msg.GetContent())
	}

	// Add more conversation
	fmt.Println("\nContinuing conversation:")
	newConversations := []struct {
		input  string
		output string
	}{
		{
			"What can I build with it?",
			"You can build chatbots, agents, multi-step workflows, and complex AI applications with memory and state management.",
		},
		{
			"What was I learning about?",
			"You mentioned you're learning about LangGraph!",
		},
	}

	for i, conv := range newConversations {
		fmt.Printf("\n[Turn %d]\n", i+1)
		fmt.Printf("User: %s\n", conv.input)
		fmt.Printf("Bot: %s\n", conv.output)

		mem.SaveContext(ctx, map[string]any{
			"input": conv.input,
		}, map[string]any{
			"output": conv.output,
		})
	}

	// Show complete history
	fmt.Println("\n--- Complete Conversation History ---")
	messages, _ = mem.GetMessages(ctx)
	fmt.Printf("Total messages: %d\n", len(messages))
	for i, msg := range messages {
		fmt.Printf("  [%d] %s: %s\n", i+1, msg.GetType(), msg.GetContent())
	}

	fmt.Println("\nKey Takeaway:")
	fmt.Println("  - Memory can be initialized with existing chat history")
	fmt.Println("  - New conversations build on top of existing history")
	fmt.Println("  - Perfect for resuming conversations or multi-session chats")
}
