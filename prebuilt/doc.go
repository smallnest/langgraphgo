// Package prebuilt provides ready-to-use agent implementations for common AI patterns.
//
// This package offers a collection of pre-built agents that implement various reasoning
// and execution patterns, from simple tool-using agents to complex multi-agent systems.
// Each agent is implemented using the core graph package and can be easily customized
// or extended for specific use cases.
//
// # Available Agents
//
// ## ReAct Agent (Reason + Act)
// The ReAct agent combines reasoning and acting by having the model think about what to do,
// choose tools to use, and act on the results. It's suitable for general-purpose tasks.
//
//	import (
//		"github.com/smallnest/langgraphgo/prebuilt"
//		"github.com/smallnest/langgraphgo/llms"
//		"github.com/smallnest/langgraphgo/tool"
//		"github.com/tmc/langchaingo/tools" // langchaingo tools satisfy tool.Tool
//	)
//
//	// Create a ReAct agent with tools
//	agent, err := prebuilt.CreateReactAgent(
//		llm,           // Language model
//		[]tool.Tool{  // Available tools
//			&tools.CalculatorTool{},
//			weatherTool,
//		},
//		10, // Max iterations
//	)
//
//	// Execute agent
//	result, err := agent.Invoke(ctx, map[string]any{
//		"messages": []llms.MessageContent{
//			{
//				Role: llms.ChatMessageTypeHuman,
//				Parts: []llms.ContentPart{
//					llms.TextPart("What's the weather in London and calculate 15% of 100?"),
//				},
//			},
//		},
//	})
//
// ## Typed ReAct Agent
// A type-safe version of the ReAct agent using Go generics:
//
//	type AgentState struct {
//		Messages       []llms.MessageContent `json:"messages"`
//		IterationCount int                    `json:"iteration_count"`
//	}
//
//	agent, err := prebuilt.CreateReactAgentTyped[AgentState](
//		llm,
//		tools,
//		10,
//		func() AgentState { return AgentState{} },
//	)
//
// ## Supervisor Agent
// Orchestrates multiple specialized agents, routing tasks to the appropriate agent:
//
//	// Create specialized agents
//	weatherAgent, _ := prebuilt.CreateReactAgent(llm, weatherTools, 5)
//	calcAgent, _ := prebuilt.CreateReactAgent(llm, calcTools, 5)
//	searchAgent, _ := prebuilt.CreateReactAgent(llm, searchTools, 5)
//
//	// Create supervisor
//	members := map[string]*graph.StateRunnableUntyped
//		"weather": weatherAgent,
//		"calculator": calcAgent,
//		"search": searchAgent,
//	}
//
//	supervisor, err := prebuilt.CreateSupervisor(
//		llm,
//		members,
//		"Router", // Router agent name
//	)
//
//	// Use supervisor to route tasks
//	result, err := supervisor.Invoke(ctx, map[string]any{
//		"messages": []llms.MessageContent{
//			{
//				Role: llms.ChatMessageTypeHuman,
//				Parts: []llms.ContentPart{
//					llms.TextPart("Calculate the distance between London and Paris"),
//				},
//			},
//		},
//	})
//
// ## Planning Agent
// Creates and executes plans for complex tasks:
//
//	planner, err := prebuilt.CreatePlanningAgent(
//		llm,
//		planningTools,
//		executionTools,
//	)
//
//	// The agent will create a plan, then execute each step
//	result, err := planner.Invoke(ctx, map[string]any{
//		"messages": []llms.MessageContent{
//			{
//				Role: llms.ChatMessageTypeHuman,
//				Parts: []llms.ContentPart{
//					llms.TextPart("Plan and execute a research report on renewable energy"),
//				},
//			},
//		},
//	})
//
// ## Choosing a Plan-Execute Agent
// Three factories implement the "plan then execute" idea with different topologies;
// pick by how execution and verification should work:
//
//   - CreatePlanningAgent: plans a workflow graph, then runs it once. Best when the
//     task decomposes into a static DAG of the available nodes.
//   - CreateManusAgent: multi-phase plan persisted to Markdown files, executed phase
//     by phase with progress checkboxes. Best for long-running, resumable tasks.
//   - CreatePEVAgent: plan → execute → verify loop that replans failed steps up to a
//     retry budget. Best when steps need validation before proceeding.
//
// All three share one functional-option surface (WithVerbose, WithSystemMessage,
// WithMaxRetries, WithVerificationPrompt, WithWorkDir, WithAutoSave, ...). The older
// ManusConfig / PEVAgentConfig struct entry points still work but are deprecated in
// favor of these options.
//
// ## Reflection Agent
// Uses self-reflection to improve responses:
//
//	reflectionAgent, err := prebuilt.CreateReflectionAgent(
//		llm,
//		tools,
//	)
//
//	// The agent will reflect on and potentially revise its answers
//	result, err := reflectionAgent.Invoke(ctx, map[string]any{
//		"messages": []llms.MessageContent{
//			{
//				Role: llms.ChatMessageTypeHuman,
//				Parts: []llms.ContentPart{
//					llms.TextPart("Explain quantum computing"),
//				},
//			},
//		},
//	})
//
// ## Tree of Thoughts Agent
// Explores multiple reasoning paths before choosing the best:
//
//	totAgent, err := prebuilt.CreateTreeOfThoughtsAgent(
//		llm,
//		3, // Number of thoughts to generate
//		5, // Maximum steps
//	)
//
//	// The agent will generate and evaluate multiple reasoning paths
//	result, err := totAgent.Invoke(ctx, map[string]any{
//		"messages": []llms.MessageContent{
//			{
//				Role: llms.ChatMessageTypeHuman,
//				Parts: []llms.ContentPart{
//					llms.TextPart("Solve this complex math problem step by step"),
//				},
//			},
//		},
//	})
//
// # Agent State Constructor Tiers
//
// Every agent pattern provides three constructor tiers to balance speed, type safety, and customizability:
//
//   - Tier 1: Map Constructors (Create<X>AgentMap)
//     Operate on map[string]any and register a MapSchema with reducers (e.g. "messages"
//     uses AppendReducer), enabling automated delta merging and parallel fan-out.
//     Zero accessor boilerplate — pass only model/tools/config. Recommended for quick
//     prototyping and dynamic workflows.
//
//   - Tier 2: Pre-Wired Struct Constructors (Create<X>AgentForState)
//     Provide compile-time type safety using standard prebuilt state structs
//     (e.g., PEVAgentState, ReflectionAgentState, TreeOfThoughtsState, AgentState).
//     Zero closure boilerplate — accessors are wired internally. Recommended for
//     type-safe production applications.
//
//   - Tier 3: Generic Custom State Constructors (Create<X>Agent[S])
//     Full compile-time type safety over your own domain-specific struct type S.
//     Requires injecting getter/setter closures for field transformations.
//     Recommended when embedding the agent within an existing proprietary data model.
//

// # RAG (Retrieval-Augmented Generation)
//
// ## Basic RAG Agent
// Combines document retrieval with generation:
//
//	rag, err := prebuilt.CreateRAGAgent(
//		llm,
//		documentLoader,   // Loads documents
//		textSplitter,     // Splits text into chunks
//		embedder,         // Creates embeddings
//		vectorStore,      // Stores and searches embeddings
//		5,                // Number of documents to retrieve
//	)
//
//	// The agent will retrieve relevant documents and generate answers
//	result, err := rag.Invoke(ctx, map[string]any{
//		"messages": []llms.MessageContent{
//			{
//				Role: llms.ChatMessageTypeHuman,
//				Parts: []llms.ContentPart{
//					llms.TextPart("What are the benefits of renewable energy?"),
//				},
//			},
//		},
//	})
//
// ## Advanced RAG with Conditional Processing
//
//	rag, err := prebuilt.CreateConditionalRAGAgent(
//		llm,
//		loader,
//		splitter,
//		embedder,
//		vectorStore,
//		3, // Retrieve count
//		// Condition function to decide whether to use RAG
//		func(ctx context.Context, query string) bool {
//			return len(strings.Fields(query)) > 5
//		},
//	)
//
// # Chat Agent
// For conversational applications:
//
//	chatAgent, err := prebuilt.CreateChatAgent(
//		llm,
//		systemPrompt, // Optional system prompt
//		memory,        // Memory for conversation history
//	)
//
//	// The agent maintains conversation context
//	result, err := chatAgent.Invoke(ctx, map[string]any{
//		"messages": []llms.MessageContent{
//			{
//				Role: llms.ChatMessageTypeHuman,
//				Parts: []llms.ContentPart{
//					llms.TextPart("Hello! How are you?"),
//				},
//			},
//		},
//	})
//
// # Custom Tools
//
// Create custom tools for agents:
//
//	type WeatherTool struct{}
//
//	func (t *WeatherTool) Name() string { return "get_weather" }
//	func (t *WeatherTool) Description() string {
//		return "Get current weather for a city"
//	}
//
//	func (t *WeatherTool) Call(ctx context.Context, input string) (string, error) {
//		// Parse the city from input
//		var data struct {
//			City string `json:"city"`
//		}
//		if err := json.Unmarshal([]byte(input), &data); err != nil {
//			return "", err
//		}
//
//		// Call weather API
//		// Implementation here...
//
//		return fmt.Sprintf("The weather in %s is 22°C and sunny", data.City), nil
//	}
//
//	// Use with any agent
//	weatherTool := &WeatherTool{}
//	agent, err := prebuilt.CreateReactAgent(llm, []tool.Tool{weatherTool}, 10)
//
// # Agent Configuration
//
// Most agents support configuration through options:
//
//	agent, err := prebuilt.CreateReactAgent(llm, tools, 10,
//		prebuilt.WithMaxTokens(4000),
//		prebuilt.WithTemperature(0.7),
//		prebuilt.WithStreaming(true),
//		prebuilt.WithCheckpointing(checkpointer),
//		prebuilt.WithMemory(memory),
//	)
//
// # Streaming Support
//
// Enable real-time streaming of agent thoughts and actions:
//
//	// Create streaming agent
//	agent, _ := prebuilt.CreateReactAgent(llm, tools, 10)
//	streaming := prebuilt.NewStreamingAgent(agent)
//
//	// Stream execution
//	stream, _ := streaming.Stream(ctx, input)
//	for event := range stream.Events {
//		fmt.Printf("Event: %v\n", event)
//	}
//
// # Memory Integration
//
// Agents can integrate with various memory strategies:
//
//	import "github.com/smallnest/langgraphgo/memory"
//
//	// Use buffer memory
//	bufferMemory := memory.NewBufferMemory(100)
//	agent, _ := prebuilt.CreateChatAgent(llm, "", bufferMemory)
//
//	// Use summarization memory
//	summMemory := memory.NewSummarizationMemory(llm, 2000)
//	agent, _ := prebuilt.CreateChatAgent(llm, "", summMemory)
//
// # Best Practices
//
//  1. Choose the right agent pattern for your use case
//  2. Provide clear tool descriptions and examples
//  3. Set appropriate iteration limits to prevent infinite loops
//  4. Use memory for conversational applications
//  5. Enable streaming for better user experience
//  6. Use checkpointing for long-running tasks
//  7. Test with various input patterns
//  8. Monitor token usage and costs
//
// # Error Handling
//
// Agents include built-in error handling:
//
//   - Tool execution failures
//   - LLM API errors
//   - Timeout protection
//   - Iteration limit enforcement
//   - Graceful degradation strategies
//
// # Performance Considerations
//
//   - Use typed agents for better performance
//   - Cache tool results when appropriate
//   - Batch tool calls when possible
//   - Monitor resource usage
//   - Consider parallel execution for independent tasks
package prebuilt
