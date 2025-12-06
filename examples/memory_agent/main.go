package main

import (
	"bufio"
	"context"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/smallnest/langgraphgo/memory"
)

// AgentState represents the state of our chat agent
type AgentState struct {
	Messages     []string
	UserInput    string
	Response     string
	MemoryStats  *memory.Stats
	CurrentTopic string
}

// ChatAgent is a simple agent that uses memory strategies
type ChatAgent struct {
	memory   memory.Strategy
	strategy string
	ctx      context.Context
}

// NewChatAgent creates a new chat agent with specified memory strategy
func NewChatAgent(strategyName string) *ChatAgent {
	ctx := context.Background()
	var mem memory.Strategy

	switch strategyName {
	case "sequential":
		mem = memory.NewSequentialMemory()
	case "sliding":
		mem = memory.NewSlidingWindowMemory(5) // Keep last 5 messages
	case "buffer":
		mem = memory.NewBufferMemory(&memory.BufferConfig{
			MaxMessages: 8,
			MaxTokens:   500,
		})
	case "summarization":
		mem = memory.NewSummarizationMemory(&memory.SummarizationConfig{
			RecentWindowSize: 3,
			SummarizeAfter:   6,
		})
	case "retrieval":
		mem = memory.NewRetrievalMemory(&memory.RetrievalConfig{
			TopK: 3,
		})
	case "hierarchical":
		mem = memory.NewHierarchicalMemory(&memory.HierarchicalConfig{
			RecentLimit:    3,
			ImportantLimit: 5,
		})
	case "graph":
		mem = memory.NewGraphBasedMemory(&memory.GraphConfig{
			TopK: 4,
		})
	case "compression":
		mem = memory.NewCompressionMemory(&memory.CompressionConfig{
			CompressionTrigger: 5,
		})
	case "oslike":
		mem = memory.NewOSLikeMemory(&memory.OSLikeConfig{
			ActiveLimit:  3,
			CacheLimit:   5,
			AccessWindow: time.Minute * 5,
		})
	default:
		mem = memory.NewBufferMemory(&memory.BufferConfig{
			MaxMessages: 10,
		})
	}

	return &ChatAgent{
		memory:   mem,
		strategy: strategyName,
		ctx:      ctx,
	}
}

// ProcessMessage handles incoming user messages
func (a *ChatAgent) ProcessMessage(userMsg string) (string, error) {
	// Add user message to memory
	msg := memory.NewMessage("user", userMsg)

	// Mark important messages
	if strings.Contains(strings.ToLower(userMsg), "important") ||
		strings.Contains(strings.ToLower(userMsg), "remember") {
		msg.Metadata["importance"] = 0.9
	}

	if err := a.memory.AddMessage(a.ctx, msg); err != nil {
		return "", err
	}

	// Get relevant context from memory
	context, err := a.memory.GetContext(a.ctx, userMsg)
	if err != nil {
		return "", err
	}

	// Generate response based on context
	response := a.generateResponse(userMsg, context)

	// Add response to memory
	responseMsg := memory.NewMessage("assistant", response)
	if err := a.memory.AddMessage(a.ctx, responseMsg); err != nil {
		return "", err
	}

	return response, nil
}

// generateResponse simulates an LLM response based on context
func (a *ChatAgent) generateResponse(input string, context []*memory.Message) string {
	inputLower := strings.ToLower(input)

	// Greetings
	if strings.Contains(inputLower, "hello") || strings.Contains(inputLower, "hi") {
		return "Hello! I'm your assistant. How can I help you today?"
	}

	// Product price queries
	if strings.Contains(inputLower, "price") {
		// Check if we mentioned price before in context
		for _, msg := range context {
			if strings.Contains(strings.ToLower(msg.Content), "$99") {
				return "As I mentioned before, the product is priced at $99."
			}
		}
		return "Our premium product is priced at $99, which includes free shipping!"
	}

	// Name queries
	if strings.Contains(inputLower, "my name") || strings.Contains(inputLower, "i am") {
		// Extract name from input
		words := strings.Fields(input)
		for i, word := range words {
			if (strings.ToLower(word) == "am" || strings.ToLower(word) == "name") && i+1 < len(words) {
				name := words[i+1]
				return fmt.Sprintf("Nice to meet you, %s! I'll remember your name.", name)
			}
		}
	}

	// Check if we know the user's name from context
	userName := ""
	for _, msg := range context {
		if msg.Role == "user" && (strings.Contains(msg.Content, "I am") || strings.Contains(msg.Content, "My name")) {
			words := strings.Fields(msg.Content)
			for i, word := range words {
				if (strings.ToLower(word) == "am" || strings.ToLower(word) == "name") && i+1 < len(words) {
					userName = words[i+1]
					break
				}
			}
		}
	}

	if userName != "" && (strings.Contains(inputLower, "who am i") || strings.Contains(inputLower, "remember me")) {
		return fmt.Sprintf("Of course I remember you, %s!", userName)
	}

	// Features query
	if strings.Contains(inputLower, "feature") {
		return "Our product has amazing features: waterproof design, 24-hour battery life, and AI-powered assistance!"
	}

	// Warranty query
	if strings.Contains(inputLower, "warranty") {
		return "Yes! We offer a 2-year warranty covering all manufacturing defects."
	}

	// Shipping query
	if strings.Contains(inputLower, "shipping") || strings.Contains(inputLower, "delivery") {
		return "We offer free standard shipping (3-5 business days) and express shipping ($15, 1-2 days)."
	}

	// Context-based responses
	if len(context) > 2 {
		return fmt.Sprintf("Based on our conversation (I remember %d messages), I'm here to help with any questions about our products!", len(context))
	}

	// Default response
	return "I understand. Could you please provide more details about what you're looking for?"
}

// GetStats returns current memory statistics
func (a *ChatAgent) GetStats() (*memory.Stats, error) {
	return a.memory.GetStats(a.ctx)
}

// Demo functions for different scenarios

func demoCustomerSupport() {
	fmt.Println("\n=== Customer Support Scenario ===")
	fmt.Println("Strategy: Sliding Window (keeps last 5 messages)")
	fmt.Println("Use case: Recent conversation context is most important\n")

	agent := NewChatAgent("sliding")
	conversation := []string{
		"Hello!",
		"What's the price of your product?",
		"Does it have good features?",
		"Tell me about the warranty",
		"What about shipping?",
		"Can you remind me of the price again?", // Tests memory recall
	}

	for _, msg := range conversation {
		fmt.Printf("User: %s\n", msg)
		response, _ := agent.ProcessMessage(msg)
		fmt.Printf("Agent: %s\n", response)

		stats, _ := agent.GetStats()
		fmt.Printf("  [Memory: %d messages, %d tokens]\n\n", stats.TotalMessages, stats.TotalTokens)

		time.Sleep(100 * time.Millisecond)
	}
}

func demoLongConsultation() {
	fmt.Println("\n=== Long Consultation Scenario ===")
	fmt.Println("Strategy: Summarization (summarizes old, keeps recent)")
	fmt.Println("Use case: Long sessions where history matters\n")

	agent := NewChatAgent("summarization")

	// Simulate longer conversation
	conversation := []string{
		"Hi, I'm John",
		"I'm interested in your product",
		"IMPORTANT: I need it to be waterproof",
		"What's the price?",
		"Tell me about features",
		"Any warranty?",
		"What are shipping options?",
		"Do you remember my name?", // Tests long-term memory
		"And my waterproof requirement?",
	}

	for _, msg := range conversation {
		fmt.Printf("User: %s\n", msg)
		response, _ := agent.ProcessMessage(msg)
		fmt.Printf("Agent: %s\n", response)

		stats, _ := agent.GetStats()
		fmt.Printf("  [Memory: %d total, %d active messages]\n\n",
			stats.TotalMessages, stats.ActiveMessages)

		time.Sleep(100 * time.Millisecond)
	}
}

func demoKnowledgeBase() {
	fmt.Println("\n=== Knowledge Base Scenario ===")
	fmt.Println("Strategy: Retrieval (retrieves relevant messages)")
	fmt.Println("Use case: Large history, query-driven retrieval\n")

	agent := NewChatAgent("retrieval")

	// Add various information
	conversation := []string{
		"What's the price?",
		"Tell me about features",
		"Shipping information?",
		"Warranty details?",
		"Available colors?",
		"Tell me about the price again", // Should retrieve price-related messages
	}

	for _, msg := range conversation {
		fmt.Printf("User: %s\n", msg)
		response, _ := agent.ProcessMessage(msg)
		fmt.Printf("Agent: %s\n", response)

		stats, _ := agent.GetStats()
		fmt.Printf("  [Memory: %d total messages stored]\n\n", stats.TotalMessages)

		time.Sleep(100 * time.Millisecond)
	}
}

func demoImportantInfo() {
	fmt.Println("\n=== Important Information Tracking ===")
	fmt.Println("Strategy: Hierarchical (keeps important + recent)")
	fmt.Println("Use case: Some messages more important than others\n")

	agent := NewChatAgent("hierarchical")

	conversation := []string{
		"Hello!",
		"IMPORTANT: Remember I'm allergic to latex",
		"What's the price?",
		"Tell me about features",
		"Any latex in the materials?", // Should remember the allergy
		"What about warranty?",
		"Shipping options?",
		"Just to confirm - you remember my allergy?",
	}

	for _, msg := range conversation {
		fmt.Printf("User: %s\n", msg)
		response, _ := agent.ProcessMessage(msg)
		fmt.Printf("Agent: %s\n", response)

		stats, _ := agent.GetStats()
		fmt.Printf("  [Memory: %d total, %d active]\n\n",
			stats.TotalMessages, stats.ActiveMessages)

		time.Sleep(100 * time.Millisecond)
	}
}

func demoGraphRelationships() {
	fmt.Println("\n=== Topic Relationship Tracking ===")
	fmt.Println("Strategy: Graph-Based (tracks topic relationships)")
	fmt.Println("Use case: Related topics and cross-references\n")

	agent := NewChatAgent("graph")

	conversation := []string{
		"What's the price of the product?",
		"Tell me about features",
		"Does the price include warranty?", // Relates price + warranty
		"What features justify the price?", // Relates features + price
		"Shipping costs?",
	}

	for _, msg := range conversation {
		fmt.Printf("User: %s\n", msg)
		response, _ := agent.ProcessMessage(msg)
		fmt.Printf("Agent: %s\n", response)

		stats, _ := agent.GetStats()

		// Show relationships for graph strategy
		if graphMem, ok := agent.memory.(*memory.GraphBasedMemory); ok {
			relations := graphMem.GetRelationships()
			fmt.Printf("  [Topics tracked: %v]\n\n", getKeys(relations))
		} else {
			fmt.Printf("  [Memory: %d messages]\n\n", stats.TotalMessages)
		}

		time.Sleep(100 * time.Millisecond)
	}
}

func getKeys(m map[string]int) []string {
	keys := make([]string, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	return keys
}

func interactiveDemo() {
	fmt.Println("\n=== Interactive Mode ===")
	fmt.Println("Choose a memory strategy:")
	fmt.Println("1. Sequential (keep all)")
	fmt.Println("2. Sliding Window (last 5)")
	fmt.Println("3. Buffer (max 8 messages)")
	fmt.Println("4. Summarization")
	fmt.Println("5. Retrieval")
	fmt.Println("6. Hierarchical")
	fmt.Println("7. Graph-Based")
	fmt.Println("8. Compression")
	fmt.Println("9. OS-Like")

	reader := bufio.NewReader(os.Stdin)
	fmt.Print("\nEnter choice (1-9): ")
	choiceStr, _ := reader.ReadString('\n')
	choiceStr = strings.TrimSpace(choiceStr)

	strategies := map[string]string{
		"1": "sequential",
		"2": "sliding",
		"3": "buffer",
		"4": "summarization",
		"5": "retrieval",
		"6": "hierarchical",
		"7": "graph",
		"8": "compression",
		"9": "oslike",
	}

	strategyName, ok := strategies[choiceStr]
	if !ok {
		strategyName = "buffer"
	}

	agent := NewChatAgent(strategyName)
	fmt.Printf("\nUsing %s strategy. Type 'quit' to exit.\n\n", strategyName)

	for {
		fmt.Print("You: ")
		input, _ := reader.ReadString('\n')
		input = strings.TrimSpace(input)

		if strings.ToLower(input) == "quit" {
			break
		}

		if input == "" {
			continue
		}

		response, err := agent.ProcessMessage(input)
		if err != nil {
			fmt.Printf("Error: %v\n", err)
			continue
		}

		fmt.Printf("Agent: %s\n", response)

		stats, _ := agent.GetStats()
		fmt.Printf("  [Memory: %d total, %d active messages, %.0f tokens]\n\n",
			stats.TotalMessages, stats.ActiveMessages, float64(stats.TotalTokens))
	}
}

func main() {
	if len(os.Args) > 1 && os.Args[1] == "interactive" {
		interactiveDemo()
		return
	}

	fmt.Println("=== Memory-Powered Agent Demonstrations ===")
	fmt.Println("Showing how different memory strategies affect agent behavior\n")

	// Run different scenarios
	demoCustomerSupport()
	time.Sleep(500 * time.Millisecond)

	demoLongConsultation()
	time.Sleep(500 * time.Millisecond)

	demoKnowledgeBase()
	time.Sleep(500 * time.Millisecond)

	demoImportantInfo()
	time.Sleep(500 * time.Millisecond)

	demoGraphRelationships()

	fmt.Println("\n=== Summary ===")
	fmt.Println("Different memory strategies provide different benefits:")
	fmt.Println("- Sliding Window: Great for customer support (recent context)")
	fmt.Println("- Summarization: Best for long consultations")
	fmt.Println("- Retrieval: Perfect for knowledge bases")
	fmt.Println("- Hierarchical: Excellent when importance varies")
	fmt.Println("- Graph: Ideal for tracking topic relationships")
	fmt.Println("\nRun with 'interactive' argument for interactive mode:")
	fmt.Println("  go run main.go interactive")
}
