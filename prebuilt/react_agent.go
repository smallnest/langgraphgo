package prebuilt

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/smallnest/langgraphgo/graph"
	"github.com/smallnest/langgraphgo/llms"
	"github.com/smallnest/langgraphgo/tool"
)

// CreateReactAgentMap creates a new ReAct agent graph with map[string]any state
//
// Deprecated: Use CreateAgentMap instead, which now includes the same iteration limiting functionality.
// This function is kept for backward compatibility and will be removed in a future version.
func CreateReactAgentMap(model llms.Model, inputTools []tool.Tool, maxIterations int) (*graph.StateRunnable[map[string]any], error) {
	if maxIterations == 0 {
		maxIterations = 20
	}
	// Define the tool executor
	toolExecutor := NewToolExecutor(inputTools)

	// Define the graph
	workflow := graph.NewStateGraph[map[string]any]()

	// Define the state schema
	agentSchema := graph.NewMapSchema()
	agentSchema.RegisterReducer("messages", graph.AppendReducer)
	workflow.SetSchema(agentSchema)

	// Define the agent node
	workflow.AddNode("agent", "ReAct agent decision maker", func(ctx context.Context, state map[string]any) (map[string]any, error) {
		messages, ok := state["messages"].([]llms.MessageContent)
		if !ok {
			return nil, fmt.Errorf("messages key not found or invalid type")
		}

		// Check iteration count
		iterationCount := 0
		if count, ok := state["iteration_count"].(int); ok {
			iterationCount = count
		}
		if iterationCount >= maxIterations {
			// Max iterations reached, return final message
			finalMsg := llms.MessageContent{
				Role: llms.ChatMessageTypeAI,
				Parts: []llms.ContentPart{
					llms.TextPart("Maximum iterations reached. Please try a simpler query."),
				},
			}
			return map[string]any{
				"messages": []llms.MessageContent{finalMsg},
			}, nil
		}

		// Convert tools to ToolInfo for the model
		var toolDefs []llms.Tool
		for _, t := range inputTools {
			toolDefs = append(toolDefs, llms.Tool{
				Type: "function",
				Function: &llms.FunctionDefinition{
					Name:        t.Name(),
					Description: t.Description(),
					Parameters:  getToolSchema(t),
				},
			})
		}

		// Call model with tools
		resp, err := model.GenerateContent(ctx, messages, llms.WithTools(toolDefs))
		if err != nil {
			return nil, err
		}

		choice := resp.Choices[0]
		aiMsg := llms.MessageContent{
			Role: llms.ChatMessageTypeAI,
		}
		if choice.Content != "" {
			aiMsg.Parts = append(aiMsg.Parts, llms.TextPart(choice.Content))
		}
		for _, tc := range choice.ToolCalls {
			aiMsg.Parts = append(aiMsg.Parts, tc)
		}

		return map[string]any{
			"messages":        []llms.MessageContent{aiMsg},
			"iteration_count": iterationCount + 1,
		}, nil
	})

	// Define the tools node
	workflow.AddNode("tools", "Tool execution node", func(ctx context.Context, state map[string]any) (map[string]any, error) {
		messages := state["messages"].([]llms.MessageContent)
		lastMsg := messages[len(messages)-1]

		if lastMsg.Role != llms.ChatMessageTypeAI {
			return nil, fmt.Errorf("last message is not an AI message")
		}

		var toolMessages []llms.MessageContent
		for _, part := range lastMsg.Parts {
			if tc, ok := part.(llms.ToolCall); ok {
				// Get the tool to check if it has a custom schema
				tool, hasTool := toolExecutor.Tools[tc.FunctionCall.Name]

				var inputVal string
				if hasTool {
					// Check if tool has custom schema
					if _, hasCustomSchema := tool.(ToolWithSchema); hasCustomSchema {
						// Tool has custom schema, pass JSON arguments directly
						inputVal = tc.FunctionCall.Arguments
					} else {
						// Tool uses default schema, try to extract "input" field
						var args map[string]any
						_ = json.Unmarshal([]byte(tc.FunctionCall.Arguments), &args)
						if val, ok := args["input"].(string); ok {
							inputVal = val
						} else {
							inputVal = tc.FunctionCall.Arguments
						}
					}
				} else {
					// Tool not found, use arguments as-is
					inputVal = tc.FunctionCall.Arguments
				}

				res, err := toolExecutor.Execute(ctx, ToolInvocation{
					Tool:      tc.FunctionCall.Name,
					ToolInput: inputVal,
				})
				if err != nil {
					res = fmt.Sprintf("Error: %v", err)
				}

				toolMsg := llms.MessageContent{
					Role: llms.ChatMessageTypeTool,
					Parts: []llms.ContentPart{
						llms.ToolCallResponse{
							ToolCallID: tc.ID,
							Name:       tc.FunctionCall.Name,
							Content:    res,
						},
					},
				}
				toolMessages = append(toolMessages, toolMsg)
			}
		}

		return map[string]any{
			"messages": toolMessages,
		}, nil
	})

	workflow.SetEntryPoint("agent")
	workflow.AddConditionalEdge("agent", func(ctx context.Context, state map[string]any) string {
		messages := state["messages"].([]llms.MessageContent)
		lastMsg := messages[len(messages)-1]
		for _, part := range lastMsg.Parts {
			if _, ok := part.(llms.ToolCall); ok {
				return "tools"
			}
		}
		return graph.END
	})
	workflow.AddEdge("tools", "agent")

	return workflow.Compile()
}

// CreateReactAgent creates a new typed ReAct agent graph
func CreateReactAgent[S any](
	model llms.Model,
	inputTools []tool.Tool,
	getMessages func(S) []llms.MessageContent,
	setMessages func(S, []llms.MessageContent) S,
	getIterationCount func(S) int,
	setIterationCount func(S, int) S,
	maxIterations int,
) (*graph.StateRunnable[S], error) {
	if maxIterations == 0 {
		maxIterations = 20
	}
	toolExecutor := NewToolExecutor(inputTools)
	workflow := graph.NewStateGraph[S]()

	workflow.AddNode("agent", "ReAct agent decision maker", func(ctx context.Context, state S) (S, error) {
		iterationCount := getIterationCount(state)
		if iterationCount >= maxIterations {
			finalMsg := llms.MessageContent{
				Role: llms.ChatMessageTypeAI,
				Parts: []llms.ContentPart{
					llms.TextPart("Maximum iterations reached. Please try a simpler query."),
				},
			}
			return setMessages(state, append(getMessages(state), finalMsg)), nil
		}

		var toolDefs []llms.Tool
		for _, t := range inputTools {
			toolDefs = append(toolDefs, llms.Tool{
				Type: "function",
				Function: &llms.FunctionDefinition{
					Name:        t.Name(),
					Description: t.Description(),
					Parameters:  getToolSchema(t),
				},
			})
		}

		messages := getMessages(state)
		resp, err := model.GenerateContent(ctx, messages, llms.WithTools(toolDefs))
		if err != nil {
			return state, err
		}

		choice := resp.Choices[0]
		aiMsg := llms.MessageContent{
			Role: llms.ChatMessageTypeAI,
		}
		if choice.Content != "" {
			aiMsg.Parts = append(aiMsg.Parts, llms.TextPart(choice.Content))
		}
		for _, tc := range choice.ToolCalls {
			aiMsg.Parts = append(aiMsg.Parts, tc)
		}

		state = setMessages(state, append(messages, aiMsg))
		state = setIterationCount(state, iterationCount+1)
		return state, nil
	})

	workflow.AddNode("tools", "Tool execution node", func(ctx context.Context, state S) (S, error) {
		messages := getMessages(state)
		lastMsg := messages[len(messages)-1]

		var toolMessages []llms.MessageContent
		for _, part := range lastMsg.Parts {
			if tc, ok := part.(llms.ToolCall); ok {
				// Get the tool to check if it has a custom schema
				tool, hasTool := toolExecutor.Tools[tc.FunctionCall.Name]

				var inputVal string
				if hasTool {
					// Check if tool has custom schema
					if _, hasCustomSchema := tool.(ToolWithSchema); hasCustomSchema {
						// Tool has custom schema, pass JSON arguments directly
						inputVal = tc.FunctionCall.Arguments
					} else {
						// Tool uses default schema, try to extract "input" field
						var args map[string]any
						_ = json.Unmarshal([]byte(tc.FunctionCall.Arguments), &args)
						if val, ok := args["input"].(string); ok {
							inputVal = val
						} else {
							inputVal = tc.FunctionCall.Arguments
						}
					}
				} else {
					// Tool not found, use arguments as-is
					inputVal = tc.FunctionCall.Arguments
				}

				res, err := toolExecutor.Execute(ctx, ToolInvocation{
					Tool:      tc.FunctionCall.Name,
					ToolInput: inputVal,
				})
				if err != nil {
					res = fmt.Sprintf("Error: %v", err)
				}

				toolMsg := llms.MessageContent{
					Role: llms.ChatMessageTypeTool,
					Parts: []llms.ContentPart{
						llms.ToolCallResponse{
							ToolCallID: tc.ID,
							Name:       tc.FunctionCall.Name,
							Content:    res,
						},
					},
				}
				toolMessages = append(toolMessages, toolMsg)
			}
		}

		return setMessages(state, append(getMessages(state), toolMessages...)), nil
	})

	workflow.SetEntryPoint("agent")
	workflow.AddConditionalEdge("agent", func(ctx context.Context, state S) string {
		messages := getMessages(state)
		lastMsg := messages[len(messages)-1]
		for _, part := range lastMsg.Parts {
			if _, ok := part.(llms.ToolCall); ok {
				return "tools"
			}
		}
		return graph.END
	})
	workflow.AddEdge("tools", "agent")

	return workflow.Compile()
}

// CreateReactAgentForState creates a type-safe ReAct agent graph pre-configured for ReactAgentState.
// This provides compile-time type safety without requiring manual getter/setter closures.
func CreateReactAgentForState(
	model llms.Model,
	inputTools []tool.Tool,
	maxIterations int,
) (*graph.StateRunnable[ReactAgentState], error) {
	return CreateReactAgent[ReactAgentState](
		model,
		inputTools,
		func(s ReactAgentState) []llms.MessageContent { return s.Messages },
		func(s ReactAgentState, m []llms.MessageContent) ReactAgentState { s.Messages = m; return s },
		func(s ReactAgentState) int { return s.IterationCount },
		func(s ReactAgentState, i int) ReactAgentState { s.IterationCount = i; return s },
		maxIterations,
	)
}
