package prebuilt

import (
	"context"
	"encoding/json"
	"fmt"

	"strings"

	"github.com/smallnest/goskills"
	adapter "github.com/smallnest/langgraphgo/adapter/goskills"
	"github.com/smallnest/langgraphgo/graph"
	"github.com/tmc/langchaingo/llms"
	"github.com/tmc/langchaingo/tools"
)

// CreateAgentOptions contains options for creating an agent
type CreateAgentOptions struct {
	skillDir      string
	Verbose       bool
	SystemMessage string
	StateModifier func(messages []llms.MessageContent) []llms.MessageContent
	Checkpointer  graph.CheckpointStore
}

// CreateAgentOption is a function that configures CreateAgentOptions
type CreateAgentOption func(*CreateAgentOptions)

// WithSystemMessage sets the system message for the agent
func WithSystemMessage(message string) CreateAgentOption {
	return func(o *CreateAgentOptions) {
		o.SystemMessage = message
	}
}

// WithStateModifier sets a function to modify messages before they are sent to the model
func WithStateModifier(modifier func(messages []llms.MessageContent) []llms.MessageContent) CreateAgentOption {
	return func(o *CreateAgentOptions) {
		o.StateModifier = modifier
	}
}

// WithCheckpointer sets the checkpointer for the agent
// Note: Currently this is a placeholder and may not be fully integrated into the graph execution yet
func WithCheckpointer(checkpointer graph.CheckpointStore) CreateAgentOption {
	return func(o *CreateAgentOptions) {
		o.Checkpointer = checkpointer
	}
}

// WithSkillDir sets the skill directory for the agent
func WithSkillDir(skillDir string) CreateAgentOption {
	return func(o *CreateAgentOptions) {
		o.skillDir = skillDir
	}
}

// WithVerbose sets the verbose mode for the agent
func WithVerbose(verbose bool) CreateAgentOption {
	return func(o *CreateAgentOptions) {
		o.Verbose = verbose
	}
}

// CreateAgent creates a new agent graph with options
func CreateAgent(model llms.Model, inputTools []tools.Tool, opts ...CreateAgentOption) (*graph.StateRunnable, error) {
	options := &CreateAgentOptions{}
	for _, opt := range opts {
		opt(options)
	}

	// Define the graph
	workflow := graph.NewStateGraph()

	// Define the state schema
	// We use a MapSchema with AppendReducer for messages
	agentSchema := graph.NewMapSchema()
	agentSchema.RegisterReducer("messages", graph.AppendReducer)
	agentSchema.RegisterReducer("extra_tools", graph.AppendReducer)
	workflow.SetSchema(agentSchema)

	// Define the skill selection node if skillDir is provided
	if options.skillDir != "" {
		workflow.AddNode("skill", func(ctx context.Context, state interface{}) (interface{}, error) {
			mState, ok := state.(map[string]interface{})
			if !ok {
				return nil, fmt.Errorf("invalid state type: %T", state)
			}

			messages, ok := mState["messages"].([]llms.MessageContent)
			if !ok || len(messages) == 0 {
				// If no messages, we can't select a skill based on user input.
				// Just pass through.
				return nil, nil
			}

			lastMsg := messages[len(messages)-1]
			userPrompt := ""
			if lastMsg.Role == llms.ChatMessageTypeHuman {
				userPrompt = lastMsg.Parts[0].(llms.TextContent).Text
			} else {
				// Try to find the last human message
				for i := len(messages) - 1; i >= 0; i-- {
					if messages[i].Role == llms.ChatMessageTypeHuman {
						userPrompt = messages[i].Parts[0].(llms.TextContent).Text
						break
					}
				}
			}

			if userPrompt == "" {
				return nil, nil
			}

			// --- STEP 1: SKILL DISCOVERY ---
			if options.Verbose {
				fmt.Printf("🔎 Discovering available skills in %s...\n", options.skillDir)
			}
			availableSkills, err := discoverSkills(options.skillDir)
			if err != nil {
				return nil, fmt.Errorf("failed to discover skills: %w", err)
			}
			if len(availableSkills) == 0 {
				if options.Verbose {
					fmt.Println("⚠️ No skills found.")
				}
				return nil, nil
			}
			if options.Verbose {
				fmt.Printf("✅ Found %d skills.\n\n", len(availableSkills))
			}

			// --- STEP 2: SKILL SELECTION ---
			if options.Verbose {
				fmt.Println("🧠 Asking LLM to select the best skill...")
			}
			selectedSkillName, err := selectSkill(ctx, model, userPrompt, availableSkills)
			if err != nil {
				return nil, fmt.Errorf("failed during skill selection: %w", err)
			}

			if selectedSkillName == "" {
				if options.Verbose {
					fmt.Println("🤷 LLM did not select any skill.")
				}
				return nil, nil
			}

			selectedSkill, ok := availableSkills[selectedSkillName]
			if !ok {
				// LLM hallucinated a skill name
				if options.Verbose {
					fmt.Printf("⚠️ LLM selected a non-existent skill '%s'. Ignoring.\n", selectedSkillName)
				}
				return nil, nil
			}
			if options.Verbose {
				fmt.Printf("✅ LLM selected skill: %s\n\n", selectedSkillName)
			}

			// Convert skill to tools
			skillTools, err := adapter.SkillsToTools(*selectedSkill)
			if err != nil {
				return nil, fmt.Errorf("failed to convert skill to tools: %w", err)
			}

			return map[string]interface{}{
				"extra_tools": skillTools,
			}, nil
		})
	}

	// Define the agent node
	workflow.AddNode("agent", func(ctx context.Context, state interface{}) (interface{}, error) {
		mState, ok := state.(map[string]interface{})
		if !ok {
			return nil, fmt.Errorf("invalid state type: %T", state)
		}

		messages, ok := mState["messages"].([]llms.MessageContent)
		if !ok {
			return nil, fmt.Errorf("messages key not found or invalid type")
		}

		// Convert tools to ToolInfo for the model
		var toolDefs []llms.Tool
		// Combine input tools with extra tools
		var allTools []tools.Tool
		allTools = append(allTools, inputTools...)

		if extra, ok := mState["extra_tools"].([]tools.Tool); ok {
			allTools = append(allTools, extra...)
		} else if extra, ok := mState["extra_tools"].([]interface{}); ok {
			// Handle case where AppendReducer might return []interface{} if types were mixed or initial append
			// But since we append []tools.Tool, reflect might keep it as []tools.Tool or []interface{} depending on implementation.
			// Graph schema AppendReducer returns interface{}.
			// If we appended a slice to nil, it returns the slice.
			// If we appended slice to slice, it returns slice.
			// We need to be careful about type assertion.
			for _, t := range extra {
				if tool, ok := t.(tools.Tool); ok {
					allTools = append(allTools, tool)
				}
			}
		}

		// Convert tools to ToolInfo for the model
		// var toolDefs []llms.Tool // Removed redeclaration
		for _, t := range allTools {
			toolDefs = append(toolDefs, llms.Tool{
				Type: "function",
				Function: &llms.FunctionDefinition{
					Name:        t.Name(),
					Description: t.Description(),
					Parameters: map[string]interface{}{
						"type": "object",
						"properties": map[string]interface{}{
							"input": map[string]interface{}{
								"type":        "string",
								"description": "The input query for the tool",
							},
						},
						"required":             []string{"input"},
						"additionalProperties": false,
					},
				},
			})
		}

		// We need to pass tools to the model
		callOpts := []llms.CallOption{
			llms.WithTools(toolDefs),
		}

		// Apply StateModifier if provided
		msgsToSend := messages

		// Prepend system message if provided (and not handled by StateModifier)
		// If StateModifier is provided, it's responsible for the whole message list structure,
		// but usually SystemMessage is separate.
		// LangChain logic: SystemMessage is prepended. StateModifier can modify everything.
		// Let's prepend SystemMessage first, then apply StateModifier?
		// Or apply StateModifier to the raw history, then prepend SystemMessage?
		// LangChain docs say: "This is useful for doing things like... removing the system message"
		// So StateModifier should probably run AFTER SystemMessage is added?
		// But if SystemMessage is just a string, we construct it here.
		// Let's construct SystemMessage first.

		if options.SystemMessage != "" {
			sysMsg := llms.TextParts(llms.ChatMessageTypeSystem, options.SystemMessage)
			// Check if the first message is already a system message?
			// For simplicity, just prepend.
			msgsToSend = append([]llms.MessageContent{sysMsg}, msgsToSend...)
		}

		// Now apply StateModifier if it exists
		// Wait, if StateModifier is used to REMOVE system message, it must run AFTER.
		// But if it's used to filter history, it might run BEFORE.
		// LangChain `create_react_agent` source:
		// 1. `_modify_state` runs on input state.
		// 2. `system_message` is added.
		// Actually, `create_agent` in LangChain 0.2+ might be different.
		// Let's stick to: SystemMessage is added to the front. StateModifier sees the result.
		if options.StateModifier != nil {
			msgsToSend = options.StateModifier(msgsToSend)
		}

		resp, err := model.GenerateContent(ctx, msgsToSend, callOpts...)
		if err != nil {
			return nil, err
		}

		choice := resp.Choices[0]

		// Create AIMessage
		aiMsg := llms.MessageContent{
			Role: llms.ChatMessageTypeAI,
		}

		if choice.Content != "" {
			aiMsg.Parts = append(aiMsg.Parts, llms.TextPart(choice.Content))
		}

		// Handle tool calls
		if len(choice.ToolCalls) > 0 {
			for _, tc := range choice.ToolCalls {
				// ToolCall implements ContentPart
				aiMsg.Parts = append(aiMsg.Parts, tc)
			}
		}

		return map[string]interface{}{
			"messages": []llms.MessageContent{aiMsg},
		}, nil
	})

	// Define the tools node
	workflow.AddNode("tools", func(ctx context.Context, state interface{}) (interface{}, error) {
		mState, ok := state.(map[string]interface{})
		if !ok {
			return nil, fmt.Errorf("invalid state")
		}

		messages := mState["messages"].([]llms.MessageContent)
		lastMsg := messages[len(messages)-1]

		if lastMsg.Role != llms.ChatMessageTypeAI {
			return nil, fmt.Errorf("last message is not an AI message")
		}

		var toolMessages []llms.MessageContent

		for _, part := range lastMsg.Parts {
			if tc, ok := part.(llms.ToolCall); ok {
				// Parse arguments to get input
				var args map[string]interface{}
				if err := json.Unmarshal([]byte(tc.FunctionCall.Arguments), &args); err != nil {
					// If unmarshal fails, try to use the raw string if it's not JSON object
				}

				inputVal := ""
				if val, ok := args["input"].(string); ok {
					inputVal = val
				} else {
					inputVal = tc.FunctionCall.Arguments
				}

				// Combine input tools with extra tools for execution
				var allTools []tools.Tool
				allTools = append(allTools, inputTools...)
				if extra, ok := mState["extra_tools"].([]tools.Tool); ok {
					allTools = append(allTools, extra...)
				} else if extra, ok := mState["extra_tools"].([]interface{}); ok {
					for _, t := range extra {
						if tool, ok := t.(tools.Tool); ok {
							allTools = append(allTools, tool)
						}
					}
				}

				// Create a temporary executor for this run
				// Optimization: We could cache this if tools don't change often, but here they might.
				currentToolExecutor := NewToolExecutor(allTools)

				// Execute tool
				res, err := currentToolExecutor.Execute(ctx, ToolInvocation{
					Tool:      tc.FunctionCall.Name,
					ToolInput: inputVal,
				})
				if err != nil {
					res = fmt.Sprintf("Error: %v", err)
				}

				// Create ToolMessage
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

		return map[string]interface{}{
			"messages": toolMessages,
		}, nil
	})

	// Define edges
	if options.skillDir != "" {
		workflow.SetEntryPoint("skill")
		workflow.AddEdge("skill", "agent")
	} else {
		workflow.SetEntryPoint("agent")
	}

	workflow.AddConditionalEdge("agent", func(ctx context.Context, state interface{}) string {
		mState := state.(map[string]interface{})
		messages := mState["messages"].([]llms.MessageContent)
		lastMsg := messages[len(messages)-1]

		hasToolCalls := false
		for _, part := range lastMsg.Parts {
			if _, ok := part.(llms.ToolCall); ok {
				hasToolCalls = true
				break
			}
		}

		if hasToolCalls {
			return "tools"
		}
		return graph.END
	})

	workflow.AddEdge("tools", "agent")

	return workflow.Compile()
}

func discoverSkills(skillDir string) (map[string]*goskills.SkillPackage, error) {
	packages, err := goskills.ParseSkillPackages(skillDir)
	if err != nil {
		return nil, err
	}

	skills := make(map[string]*goskills.SkillPackage)
	for _, pkg := range packages {
		skills[pkg.Meta.Name] = pkg
	}
	return skills, nil
}

func selectSkill(ctx context.Context, model llms.Model, userPrompt string, availableSkills map[string]*goskills.SkillPackage) (string, error) {
	var skillDescriptions string
	for name, pkg := range availableSkills {
		skillDescriptions += fmt.Sprintf("- %s: %s\n", name, pkg.Meta.Description)
	}

	prompt := fmt.Sprintf(`You are an intelligent agent that selects the most appropriate skill to handle a user's request.

Available Skills:
%s

User Request: "%s"

Instructions:
1. Analyze the user's request.
2. Determine if any of the available skills are relevant.
3. If a skill is relevant, output ONLY the name of the skill.
4. If no skill is relevant, output "None".

Output:`, skillDescriptions, userPrompt)

	resp, err := model.GenerateContent(ctx, []llms.MessageContent{
		llms.TextParts(llms.ChatMessageTypeHuman, prompt),
	})
	if err != nil {
		return "", err
	}

	selection := resp.Choices[0].Content
	return strings.TrimSpace(selection), nil
}
