package main

import (
	"context"
	"fmt"
	"log"

	"github.com/smallnest/langgraphgo/graph"
	"github.com/tmc/langchaingo/llms"
)

// CreateResearcherGraph creates the researcher subgraph for conducting focused research
func CreateResearcherGraph(config *Configuration, model llms.Model) (*graph.MessageGraph, error) {
	workflow := graph.NewMessageGraph()

	// Set up schema with reducers
	schema := graph.NewMapSchema()
	schema.RegisterReducer("messages", graph.AppendReducer)
	schema.RegisterReducer("raw_notes", graph.AppendReducer)
	workflow.SetSchema(schema)

	// Initialize search tool
	searchTool := &TavilySearchTool{APIKey: config.TavilyAPIKey}
	thinkTool := &ThinkToolImpl{}

	// Researcher node - conducts research using search tools
	workflow.AddNode("researcher", func(ctx context.Context, state interface{}) (interface{}, error) {
		mState, ok := state.(map[string]interface{})
		if !ok {
			return nil, fmt.Errorf("invalid state type")
		}

		messages, ok := mState["messages"].([]llms.MessageContent)
		if !ok {
			return nil, fmt.Errorf("messages not found in state")
		}

		toolCallIterations, _ := mState["tool_call_iterations"].(int)

		// Check iteration limit
		if toolCallIterations >= config.MaxToolCallIterations {
			log.Printf("[Researcher] Reached max tool call iterations (%d), ending research", config.MaxToolCallIterations)
			return map[string]interface{}{
				"messages": []llms.MessageContent{
					llms.TextParts(llms.ChatMessageTypeAI, "Research complete - reached iteration limit."),
				},
			}, nil
		}

		// Prepare messages with system prompt
		systemPrompt := GetResearcherSystemPrompt(config.MaxToolCallIterations)

		var msgs []llms.MessageContent
		// Always start with system message
		msgs = append(msgs, llms.TextParts(llms.ChatMessageTypeSystem, systemPrompt))
		// Then add conversation history
		msgs = append(msgs, messages...)

		// Define tools for the model
		toolDefs := []llms.Tool{
			{
				Type: "function",
				Function: &llms.FunctionDefinition{
					Name:        "tavily_search",
					Description: "Search the web for information. Input should be a search query string.",
					Parameters: map[string]interface{}{
						"type": "object",
						"properties": map[string]interface{}{
							"query": map[string]interface{}{
								"type":        "string",
								"description": "The search query",
							},
						},
						"required": []string{"query"},
					},
				},
			},
			{
				Type: "function",
				Function: &llms.FunctionDefinition{
					Name:        "think_tool",
					Description: "Use this to reflect on your progress and plan next steps.",
					Parameters: map[string]interface{}{
						"type": "object",
						"properties": map[string]interface{}{
							"reflection": map[string]interface{}{
								"type":        "string",
								"description": "Your reflection on the current state and next steps",
							},
						},
						"required": []string{"reflection"},
					},
				},
			},
		}

		// Call model with tools
		resp, err := model.GenerateContent(ctx, msgs, llms.WithTools(toolDefs), llms.WithMaxTokens(config.ResearchModelMaxTokens))
		if err != nil {
			return nil, fmt.Errorf("model call failed: %w", err)
		}

		choice := resp.Choices[0]
		aiMsg := llms.MessageContent{Role: llms.ChatMessageTypeAI}

		if choice.Content != "" {
			aiMsg.Parts = append(aiMsg.Parts, llms.TextPart(choice.Content))
		}

		// Add tool calls to message
		for _, tc := range choice.ToolCalls {
			aiMsg.Parts = append(aiMsg.Parts, tc)
		}

		return map[string]interface{}{
			"messages":             []llms.MessageContent{aiMsg},
			"tool_call_iterations": toolCallIterations + 1,
		}, nil
	})

	// Researcher tools node - executes tool calls
	workflow.AddNode("researcher_tools", func(ctx context.Context, state interface{}) (interface{}, error) {
		mState, ok := state.(map[string]interface{})
		if !ok {
			return nil, fmt.Errorf("invalid state type")
		}

		messages := mState["messages"].([]llms.MessageContent)
		lastMsg := messages[len(messages)-1]

		var toolMessages []llms.MessageContent
		var rawNotes []string

		toolCalls := GetToolCalls(lastMsg)
		for _, tc := range toolCalls {
			args, err := ParseToolArguments(tc)
			if err != nil {
				log.Printf("[Researcher Tools] Failed to parse args: %v", err)
				continue
			}

			var result string
			switch tc.FunctionCall.Name {
			case "tavily_search":
				query, _ := args["query"].(string)
				result, err = searchTool.Call(ctx, query)
				if err != nil {
					result = fmt.Sprintf("Search error: %v", err)
				} else {
					// Store raw search results
					rawNotes = append(rawNotes, result)
				}

			case "think_tool":
				reflection, _ := args["reflection"].(string)
				result, _ = thinkTool.Call(ctx, reflection)

			default:
				result = fmt.Sprintf("Unknown tool: %s", tc.FunctionCall.Name)
			}

			toolMsg := CreateToolMessage(tc.ID, tc.FunctionCall.Name, result)
			toolMessages = append(toolMessages, toolMsg)
		}

		return map[string]interface{}{
			"messages":  toolMessages,
			"raw_notes": rawNotes,
		}, nil
	})

	// Compress research node - summarizes findings
	workflow.AddNode("compress_research", func(ctx context.Context, state interface{}) (interface{}, error) {
		mState := state.(map[string]interface{})

		researchTopic, _ := mState["research_topic"].(string)
		rawNotes, _ := mState["raw_notes"].([]string)

		if len(rawNotes) == 0 {
			return map[string]interface{}{
				"compressed_research": "No research findings to compress.",
			}, nil
		}

		// Create compression prompt
		prompt := GetCompressionPrompt(researchTopic, JoinNotes(rawNotes))

		// Call model for compression
		resp, err := model.GenerateContent(ctx, []llms.MessageContent{
			llms.TextParts(llms.ChatMessageTypeHuman, prompt),
		}, llms.WithMaxTokens(config.CompressionModelMaxTokens))

		if err != nil {
			return nil, fmt.Errorf("compression failed: %w", err)
		}

		compressed := resp.Choices[0].Content

		return map[string]interface{}{
			"compressed_research": compressed,
			"raw_notes":           rawNotes,
		}, nil
	})

	// Define edges
	workflow.SetEntryPoint("researcher")

	// Conditional edge from researcher
	workflow.AddConditionalEdge("researcher", func(ctx context.Context, state interface{}) string {
		mState := state.(map[string]interface{})
		messages := mState["messages"].([]llms.MessageContent)
		lastMsg := messages[len(messages)-1]

		if HasToolCalls(lastMsg) {
			return "researcher_tools"
		}
		return "compress_research"
	})

	workflow.AddEdge("researcher_tools", "researcher")
	workflow.AddEdge("compress_research", graph.END)

	return workflow, nil
}
