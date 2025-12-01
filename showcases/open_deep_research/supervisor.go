package main

import (
	"context"
	"fmt"
	"log"
	"sync"

	"github.com/smallnest/langgraphgo/graph"
	"github.com/tmc/langchaingo/llms"
)

// CreateSupervisorGraph creates the supervisor subgraph for managing research delegation
func CreateSupervisorGraph(config *Configuration, model llms.Model, researcherGraph *graph.MessageGraph) (*graph.MessageGraph, error) {
	workflow := graph.NewMessageGraph()

	// Set up schema with reducers
	schema := graph.NewMapSchema()
	schema.RegisterReducer("supervisor_messages", graph.AppendReducer)
	schema.RegisterReducer("notes", graph.AppendReducer)
	schema.RegisterReducer("raw_notes", graph.AppendReducer)
	workflow.SetSchema(schema)

	// Supervisor node - delegates research tasks
	workflow.AddNode("supervisor", func(ctx context.Context, state interface{}) (interface{}, error) {
		mState, ok := state.(map[string]interface{})
		if !ok {
			return nil, fmt.Errorf("invalid state type")
		}

		messages, ok := mState["supervisor_messages"].([]llms.MessageContent)
		if !ok {
			return nil, fmt.Errorf("supervisor_messages not found")
		}

		researchBrief, _ := mState["research_brief"].(string)
		researchIterations, _ := mState["research_iterations"].(int)

		log.Printf("[Supervisor] Starting iteration %d with %d existing messages", researchIterations+1, len(messages))

		// Check iteration limit
		if researchIterations >= config.MaxResearcherIterations {
			log.Printf("[Supervisor] Reached max research iterations (%d), completing research", config.MaxResearcherIterations)

			// Create completion message
			completeMsg := llms.MessageContent{
				Role: llms.ChatMessageTypeAI,
				Parts: []llms.ContentPart{
					llms.ToolCall{
						ID: "complete_1",
						FunctionCall: &llms.FunctionCall{
							Name:      "ResearchComplete",
							Arguments: `{"complete": true}`,
						},
					},
				},
			}

			return map[string]interface{}{
				"supervisor_messages": []llms.MessageContent{completeMsg},
			}, nil
		}

		// Prepare system prompt
		systemPrompt := GetSupervisorSystemPrompt(config.MaxResearcherIterations, config.MaxConcurrentResearchUnits)

		// Build message list properly
		var msgs []llms.MessageContent

		// Always start with system message
		msgs = append(msgs, llms.TextParts(llms.ChatMessageTypeSystem, systemPrompt))

		// Add research brief as initial human message only if no conversation history
		if len(messages) == 0 {
			msgs = append(msgs, llms.TextParts(llms.ChatMessageTypeHuman, fmt.Sprintf("Research Brief: %s\n\nPlease analyze this research brief and decide how to delegate the research. Use the think_tool first to plan your approach, then call ConductResearch to delegate tasks.", researchBrief)))
		} else {
			// For subsequent iterations, append the conversation history
			msgs = append(msgs, messages...)
		}

		// Define tools for supervisor
		toolDefs := []llms.Tool{
			{
				Type: "function",
				Function: &llms.FunctionDefinition{
					Name:        "ConductResearch",
					Description: "Delegate a research task to a specialized sub-agent. Provide a detailed research topic.",
					Parameters: map[string]interface{}{
						"type": "object",
						"properties": map[string]interface{}{
							"research_topic": map[string]interface{}{
								"type":        "string",
								"description": "The topic to research. Should be a single topic described in high detail (at least a paragraph).",
							},
						},
						"required": []string{"research_topic"},
					},
				},
			},
			{
				Type: "function",
				Function: &llms.FunctionDefinition{
					Name:        "ResearchComplete",
					Description: "Call this when research is complete and you have enough information.",
					Parameters: map[string]interface{}{
						"type": "object",
						"properties": map[string]interface{}{
							"complete": map[string]interface{}{
								"type":        "boolean",
								"description": "Set to true when research is complete",
							},
						},
						"required": []string{"complete"},
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

		// Call model
		resp, err := model.GenerateContent(ctx, msgs, llms.WithTools(toolDefs), llms.WithMaxTokens(config.ResearchModelMaxTokens))
		if err != nil {
			return nil, fmt.Errorf("supervisor model call failed: %w", err)
		}

		choice := resp.Choices[0]
		aiMsg := llms.MessageContent{Role: llms.ChatMessageTypeAI}

		if choice.Content != "" {
			aiMsg.Parts = append(aiMsg.Parts, llms.TextPart(choice.Content))
		}

		for _, tc := range choice.ToolCalls {
			aiMsg.Parts = append(aiMsg.Parts, tc)
		}

		return map[string]interface{}{
			"supervisor_messages": []llms.MessageContent{aiMsg},
			"research_iterations": researchIterations + 1,
		}, nil
	})

	// Supervisor tools node - executes research delegation
	workflow.AddNode("supervisor_tools", func(ctx context.Context, state interface{}) (interface{}, error) {
		mState := state.(map[string]interface{})
		messages := mState["supervisor_messages"].([]llms.MessageContent)
		lastMsg := messages[len(messages)-1]

		var toolMessages []llms.MessageContent
		var allRawNotes []string
		var allNotes []string

		toolCalls := GetToolCalls(lastMsg)

		// Separate tool calls by type
		conductResearchCalls := ExtractToolCallsByName(toolCalls, "ConductResearch")
		thinkCalls := ExtractToolCallsByName(toolCalls, "think_tool")

		// Handle think tool calls
		thinkTool := &ThinkToolImpl{}
		for _, tc := range thinkCalls {
			args, _ := ParseToolArguments(tc)
			reflection, _ := args["reflection"].(string)
			result, _ := thinkTool.Call(ctx, reflection)
			toolMsg := CreateToolMessage(tc.ID, tc.FunctionCall.Name, result)
			toolMessages = append(toolMessages, toolMsg)
		}

		// Handle ConductResearch calls in parallel (up to max concurrent)
		if len(conductResearchCalls) > 0 {
			// Limit concurrent research
			maxConcurrent := config.MaxConcurrentResearchUnits
			if len(conductResearchCalls) > maxConcurrent {
				log.Printf("[Supervisor] Limiting research tasks from %d to %d", len(conductResearchCalls), maxConcurrent)

				// Add error messages for overflow
				for _, tc := range conductResearchCalls[maxConcurrent:] {
					errMsg := CreateToolMessage(
						tc.ID,
						tc.FunctionCall.Name,
						fmt.Sprintf("Error: Exceeded maximum concurrent research units (%d). Please try again with fewer tasks.", maxConcurrent),
					)
					toolMessages = append(toolMessages, errMsg)
				}

				conductResearchCalls = conductResearchCalls[:maxConcurrent]
			}

			// Execute research tasks in parallel
			type researchResult struct {
				toolCallID string
				result     string
				rawNotes   []string
				err        error
			}

			resultsChan := make(chan researchResult, len(conductResearchCalls))
			var wg sync.WaitGroup

			for _, tc := range conductResearchCalls {
				wg.Add(1)
				go func(toolCall llms.ToolCall) {
					defer wg.Done()

					args, err := ParseToolArguments(toolCall)
					if err != nil {
						resultsChan <- researchResult{
							toolCallID: toolCall.ID,
							err:        err,
						}
						return
					}

					researchTopic, _ := args["research_topic"].(string)

					// Compile and invoke researcher subgraph
					researcherRunnable, err := researcherGraph.Compile()
					if err != nil {
						resultsChan <- researchResult{
							toolCallID: toolCall.ID,
							err:        fmt.Errorf("failed to compile researcher: %w", err),
						}
						return
					}

					// Prepare researcher state
					researcherState := map[string]interface{}{
						"messages": []llms.MessageContent{
							llms.TextParts(llms.ChatMessageTypeHuman, researchTopic),
						},
						"research_topic":       researchTopic,
						"tool_call_iterations": 0,
						"raw_notes":            []string{},
					}

					result, err := researcherRunnable.Invoke(ctx, researcherState)
					if err != nil {
						resultsChan <- researchResult{
							toolCallID: toolCall.ID,
							err:        fmt.Errorf("researcher execution failed: %w", err),
						}
						return
					}

					resultState := result.(map[string]interface{})
					compressed, _ := resultState["compressed_research"].(string)
					rawNotes, _ := resultState["raw_notes"].([]string)

					resultsChan <- researchResult{
						toolCallID: toolCall.ID,
						result:     compressed,
						rawNotes:   rawNotes,
					}
				}(tc)
			}

			// Wait for all research tasks to complete
			go func() {
				wg.Wait()
				close(resultsChan)
			}()

			// Collect results
			for res := range resultsChan {
				var content string
				if res.err != nil {
					content = fmt.Sprintf("Research error: %v", res.err)
				} else {
					content = res.result
					allNotes = append(allNotes, res.result)
					allRawNotes = append(allRawNotes, res.rawNotes...)
				}

				toolMsg := CreateToolMessage(res.toolCallID, "ConductResearch", content)
				toolMessages = append(toolMessages, toolMsg)
			}
		}

		return map[string]interface{}{
			"supervisor_messages": toolMessages,
			"notes":               allNotes,
			"raw_notes":           allRawNotes,
		}, nil
	})

	// Define edges
	workflow.SetEntryPoint("supervisor")

	workflow.AddConditionalEdge("supervisor", func(ctx context.Context, state interface{}) string {
		mState := state.(map[string]interface{})
		messages := mState["supervisor_messages"].([]llms.MessageContent)
		lastMsg := messages[len(messages)-1]

		toolCalls := GetToolCalls(lastMsg)

		// Check if ResearchComplete was called
		for _, tc := range toolCalls {
			if tc.FunctionCall.Name == "ResearchComplete" {
				return graph.END
			}
		}

		// If there are other tool calls, go to tools node
		if len(toolCalls) > 0 {
			return "supervisor_tools"
		}

		return graph.END
	})

	workflow.AddEdge("supervisor_tools", "supervisor")

	return workflow, nil
}
