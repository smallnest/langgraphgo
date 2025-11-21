//go:build ignore
// +build ignore

package main

import (
	"context"
	"fmt"
	"strings"

	"github.com/smallnest/langgraphgo/graph"
	"github.com/tmc/langchaingo/llms"
)

func main() {
	fmt.Println("🔀 Conditional Edges Example")
	fmt.Println("============================\n")

	// Example 1: Simple Intent Router
	SimpleIntentRouter()

	// Example 2: Multi-step Workflow with Conditions
	MultiStepWorkflow()

	// Example 3: Dynamic Tool Selection
	DynamicToolSelection()
}

// SimpleIntentRouter demonstrates routing based on user intent
func SimpleIntentRouter() {
	fmt.Println("1️⃣ Intent-Based Routing")
	fmt.Println("------------------------")

	g := graph.NewMessageGraph()

	// Entry point - analyze intent
	g.AddNode("analyze_intent", func(ctx context.Context, state interface{}) (interface{}, error) {
		messages := state.([]llms.MessageContent)
		fmt.Printf("   Analyzing: %s\n", messages[0].Parts[0].(llms.TextContent).Text)
		return messages, nil
	})

	// Different handlers for different intents
	g.AddNode("handle_question", func(ctx context.Context, state interface{}) (interface{}, error) {
		messages := state.([]llms.MessageContent)
		response := "I'll help answer your question about that."
		fmt.Printf("   ❓ Question Handler: %s\n", response)
		return append(messages, llms.TextParts("ai", response)), nil
	})

	g.AddNode("handle_command", func(ctx context.Context, state interface{}) (interface{}, error) {
		messages := state.([]llms.MessageContent)
		response := "Executing your command..."
		fmt.Printf("   ⚡ Command Handler: %s\n", response)
		return append(messages, llms.TextParts("ai", response)), nil
	})

	g.AddNode("handle_feedback", func(ctx context.Context, state interface{}) (interface{}, error) {
		messages := state.([]llms.MessageContent)
		response := "Thank you for your feedback!"
		fmt.Printf("   💬 Feedback Handler: %s\n", response)
		return append(messages, llms.TextParts("ai", response)), nil
	})

	// Conditional routing based on intent
	g.AddConditionalEdge("analyze_intent", func(ctx context.Context, state interface{}) string {
		messages := state.([]llms.MessageContent)
		if len(messages) > 0 {
			text := messages[0].Parts[0].(llms.TextContent).Text
			text = strings.ToLower(text)

			// Route based on keywords
			if strings.Contains(text, "?") || strings.Contains(text, "what") || strings.Contains(text, "how") {
				fmt.Println("   → Routing to Question Handler")
				return "handle_question"
			}
			if strings.Contains(text, "please") || strings.Contains(text, "could you") || strings.Contains(text, "run") {
				fmt.Println("   → Routing to Command Handler")
				return "handle_command"
			}
			if strings.Contains(text, "thanks") || strings.Contains(text, "good") || strings.Contains(text, "bad") {
				fmt.Println("   → Routing to Feedback Handler")
				return "handle_feedback"
			}
		}
		// Default to question handler
		fmt.Println("   → Default: Routing to Question Handler")
		return "handle_question"
	})

	// All handlers go to END
	g.AddEdge("handle_question", graph.END)
	g.AddEdge("handle_command", graph.END)
	g.AddEdge("handle_feedback", graph.END)

	g.SetEntryPoint("analyze_intent")

	// Compile and test with different inputs
	runnable, _ := g.Compile()
	ctx := context.Background()

	// Test different intents
	testInputs := []string{
		"What is the weather today?",
		"Please run the diagnostic tool",
		"Thanks for your help!",
	}

	for _, input := range testInputs {
		fmt.Printf("\n📝 Input: %s\n", input)
		messages := []llms.MessageContent{llms.TextParts("human", input)}
		result, _ := runnable.Invoke(ctx, messages)
		finalMessages := result.([]llms.MessageContent)
		fmt.Printf("   Response: %s\n", finalMessages[len(finalMessages)-1].Parts[0].(llms.TextContent).Text)
	}
	fmt.Println()
}

// MultiStepWorkflow demonstrates a workflow with conditional branching
func MultiStepWorkflow() {
	fmt.Println("2️⃣ Multi-Step Workflow with Conditions")
	fmt.Println("---------------------------------------")

	g := graph.NewMessageGraph()

	// Data validation step
	g.AddNode("validate", func(ctx context.Context, state interface{}) (interface{}, error) {
		data := state.(map[string]interface{})
		fmt.Printf("   Validating data: %v\n", data)

		// Check if data is valid
		if value, ok := data["value"].(int); ok && value > 0 {
			data["valid"] = true
		} else {
			data["valid"] = false
		}
		return data, nil
	})

	// Process valid data
	g.AddNode("process", func(ctx context.Context, state interface{}) (interface{}, error) {
		data := state.(map[string]interface{})
		fmt.Println("   ✅ Processing valid data...")
		data["result"] = data["value"].(int) * 2
		data["status"] = "processed"
		return data, nil
	})

	// Handle invalid data
	g.AddNode("handle_error", func(ctx context.Context, state interface{}) (interface{}, error) {
		data := state.(map[string]interface{})
		fmt.Println("   ❌ Handling invalid data...")
		data["status"] = "error"
		data["error"] = "Invalid input value"
		return data, nil
	})

	// Store results
	g.AddNode("store", func(ctx context.Context, state interface{}) (interface{}, error) {
		data := state.(map[string]interface{})
		fmt.Printf("   💾 Storing result: %v\n", data["result"])
		return data, nil
	})

	// Conditional edge after validation
	g.AddConditionalEdge("validate", func(ctx context.Context, state interface{}) string {
		data := state.(map[string]interface{})
		if valid, ok := data["valid"].(bool); ok && valid {
			fmt.Println("   → Data is valid, proceeding to process")
			return "process"
		}
		fmt.Println("   → Data is invalid, handling error")
		return "handle_error"
	})

	// Conditional edge after processing
	g.AddConditionalEdge("process", func(ctx context.Context, state interface{}) string {
		data := state.(map[string]interface{})
		if result, ok := data["result"].(int); ok && result > 100 {
			fmt.Println("   → Large result, storing...")
			return "store"
		}
		fmt.Println("   → Small result, skipping storage")
		return graph.END
	})

	g.AddEdge("handle_error", graph.END)
	g.AddEdge("store", graph.END)

	g.SetEntryPoint("validate")

	// Test the workflow
	runnable, _ := g.Compile()
	ctx := context.Background()

	testCases := []map[string]interface{}{
		{"value": 60}, // Valid, large result -> will be stored
		{"value": 10}, // Valid, small result -> won't be stored
		{"value": -5}, // Invalid -> error handling
	}

	for i, testData := range testCases {
		fmt.Printf("\n   Test %d: Input = %v\n", i+1, testData)
		result, _ := runnable.Invoke(ctx, testData)
		finalData := result.(map[string]interface{})
		fmt.Printf("   Final State: %v\n", finalData)
	}
	fmt.Println()
}

// DynamicToolSelection demonstrates selecting tools based on task requirements
func DynamicToolSelection() {
	fmt.Println("3️⃣ Dynamic Tool Selection")
	fmt.Println("-------------------------")

	g := graph.NewMessageGraph()

	// Analyze task requirements
	g.AddNode("analyze_task", func(ctx context.Context, state interface{}) (interface{}, error) {
		task := state.(string)
		fmt.Printf("   Analyzing task: %s\n", task)
		return task, nil
	})

	// Different tools
	g.AddNode("calculator", func(ctx context.Context, state interface{}) (interface{}, error) {
		task := state.(string)
		fmt.Println("   🧮 Using Calculator Tool")
		return task + " -> Result: 42", nil
	})

	g.AddNode("web_search", func(ctx context.Context, state interface{}) (interface{}, error) {
		task := state.(string)
		fmt.Println("   🔍 Using Web Search Tool")
		return task + " -> Found 10 relevant results", nil
	})

	g.AddNode("code_generator", func(ctx context.Context, state interface{}) (interface{}, error) {
		task := state.(string)
		fmt.Println("   💻 Using Code Generator Tool")
		return task + " -> Generated code snippet", nil
	})

	g.AddNode("translator", func(ctx context.Context, state interface{}) (interface{}, error) {
		task := state.(string)
		fmt.Println("   🌐 Using Translator Tool")
		return task + " -> Translated to target language", nil
	})

	// Tool selection based on task keywords
	g.AddConditionalEdge("analyze_task", func(ctx context.Context, state interface{}) string {
		task := strings.ToLower(state.(string))

		if strings.Contains(task, "calculate") || strings.Contains(task, "compute") || strings.Contains(task, "math") {
			fmt.Println("   → Selecting Calculator")
			return "calculator"
		}
		if strings.Contains(task, "search") || strings.Contains(task, "find") || strings.Contains(task, "lookup") {
			fmt.Println("   → Selecting Web Search")
			return "web_search"
		}
		if strings.Contains(task, "code") || strings.Contains(task, "program") || strings.Contains(task, "function") {
			fmt.Println("   → Selecting Code Generator")
			return "code_generator"
		}
		if strings.Contains(task, "translate") || strings.Contains(task, "language") {
			fmt.Println("   → Selecting Translator")
			return "translator"
		}

		// Default to web search
		fmt.Println("   → Default: Selecting Web Search")
		return "web_search"
	})

	// All tools go to END
	g.AddEdge("calculator", graph.END)
	g.AddEdge("web_search", graph.END)
	g.AddEdge("code_generator", graph.END)
	g.AddEdge("translator", graph.END)

	g.SetEntryPoint("analyze_task")

	// Test with different tasks
	runnable, _ := g.Compile()
	ctx := context.Background()

	tasks := []string{
		"Calculate the compound interest",
		"Search for best practices in Go",
		"Generate code for sorting algorithm",
		"Translate this to Spanish",
		"Analyze market trends", // Will use default
	}

	for _, task := range tasks {
		fmt.Printf("\n📋 Task: %s\n", task)
		result, _ := runnable.Invoke(ctx, task)
		fmt.Printf("   Output: %s\n", result)
	}
	fmt.Println("\n✅ Conditional Edges Examples Complete!")
}
