package prebuilt

import (
	"context"
	"encoding/json"
	"fmt"
	"slices"
	"strings"

	"github.com/smallnest/goskills"
	adapter "github.com/smallnest/langgraphgo/adapter/goskills"
	"github.com/smallnest/langgraphgo/graph"
	"github.com/smallnest/langgraphgo/llms"
	"github.com/smallnest/langgraphgo/tool"
)

// CreateAgentOptions contains options for creating an agent.
//
// The plan-execute family (CreatePlanningAgent, CreateManusAgent, CreatePEVAgent)
// shares this option surface. Fields below the common set are consumed only by the
// factories they apply to: MaxRetries/VerificationPrompt by the PEV agent, and
// WorkDir/PlanPath/NotesPath/OutputPath/AutoSave by the Manus agent.
type CreateAgentOptions struct {
	skillDir               string
	Verbose                bool
	SystemMessage          string
	StateModifier          func(messages []llms.MessageContent) []llms.MessageContent
	MaxIterations          int
	DisableModelInvocation bool

	// PEV agent options.
	MaxRetries         int
	VerificationPrompt string

	// Manus agent options.
	WorkDir    string
	PlanPath   string
	NotesPath  string
	OutputPath string
	AutoSave   bool
}

type CreateAgentOption func(*CreateAgentOptions)

func WithSystemMessage(message string) CreateAgentOption {
	return func(o *CreateAgentOptions) { o.SystemMessage = message }
}

func WithStateModifier(modifier func(messages []llms.MessageContent) []llms.MessageContent) CreateAgentOption {
	return func(o *CreateAgentOptions) { o.StateModifier = modifier }
}

func WithSkillDir(skillDir string) CreateAgentOption {
	return func(o *CreateAgentOptions) { o.skillDir = skillDir }
}

func WithVerbose(verbose bool) CreateAgentOption {
	return func(o *CreateAgentOptions) { o.Verbose = verbose }
}

func WithMaxIterations(maxIterations int) CreateAgentOption {
	return func(o *CreateAgentOptions) { o.MaxIterations = maxIterations }
}

func WithDisableModelInvocation(disable bool) CreateAgentOption {
	return func(o *CreateAgentOptions) { o.DisableModelInvocation = disable }
}

// WithMaxRetries sets the maximum replanning retries for the PEV agent.
func WithMaxRetries(maxRetries int) CreateAgentOption {
	return func(o *CreateAgentOptions) { o.MaxRetries = maxRetries }
}

// WithVerificationPrompt sets the verification system prompt for the PEV agent.
func WithVerificationPrompt(prompt string) CreateAgentOption {
	return func(o *CreateAgentOptions) { o.VerificationPrompt = prompt }
}

// WithWorkDir sets the working directory for the Manus agent's persistent files.
func WithWorkDir(dir string) CreateAgentOption {
	return func(o *CreateAgentOptions) { o.WorkDir = dir }
}

// WithPlanPath sets the plan file path for the Manus agent.
func WithPlanPath(path string) CreateAgentOption {
	return func(o *CreateAgentOptions) { o.PlanPath = path }
}

// WithNotesPath sets the notes file path for the Manus agent.
func WithNotesPath(path string) CreateAgentOption {
	return func(o *CreateAgentOptions) { o.NotesPath = path }
}

// WithOutputPath sets the output file path for the Manus agent.
func WithOutputPath(path string) CreateAgentOption {
	return func(o *CreateAgentOptions) { o.OutputPath = path }
}

// WithAutoSave toggles persisting plan/notes/output to disk for the Manus agent.
func WithAutoSave(autoSave bool) CreateAgentOption {
	return func(o *CreateAgentOptions) { o.AutoSave = autoSave }
}

// CreateAgentMap creates a new agent graph with map[string]any state
func CreateAgentMap(model llms.Model, inputTools []tool.Tool, maxIterations int, opts ...CreateAgentOption) (*graph.StateRunnable[map[string]any], error) {
	options := &CreateAgentOptions{}
	for _, opt := range opts {
		opt(options)
	}
	if maxIterations == 0 {
		maxIterations = 20
	}
	if options.MaxIterations > 0 {
		maxIterations = options.MaxIterations
	}

	workflow := graph.NewStateGraph[map[string]any]()
	agentSchema := graph.NewMapSchema()
	agentSchema.RegisterReducer("messages", graph.AppendReducer)
	agentSchema.RegisterReducer("extra_tools", graph.AppendReducer)
	agentSchema.RegisterReducer("iteration_count", graph.OverwriteReducer)
	workflow.SetSchema(agentSchema)

	if options.skillDir != "" {
		workflow.AddNode("skill", "Skill discovery node", func(ctx context.Context, state map[string]any) (map[string]any, error) {
			messages, _ := state["messages"].([]llms.MessageContent)
			if len(messages) == 0 {
				return nil, nil
			}
			userPrompt := ""
			for _, message := range slices.Backward(messages) {
				if message.Role == llms.ChatMessageTypeHuman {
					userPrompt = message.Parts[0].(llms.TextContent).Text
					break
				}
			}
			if userPrompt == "" {
				return nil, nil
			}

			availableSkills, err := discoverSkills(options.skillDir)
			if err != nil {
				return nil, err
			}
			selectedSkillName, err := selectSkill(ctx, model, userPrompt, availableSkills)
			if err != nil || selectedSkillName == "" {
				return nil, err
			}
			selectedSkill, ok := availableSkills[selectedSkillName]
			if !ok {
				return nil, nil
			}
			skillTools, err := adapter.SkillsToTools(selectedSkill)
			if err != nil {
				return nil, err
			}
			return map[string]any{"extra_tools": skillTools}, nil
		})
	}

	workflow.AddNode("agent", "Agent decision node", func(ctx context.Context, state map[string]any) (map[string]any, error) {
		messages, _ := state["messages"].([]llms.MessageContent)
		var allTools []tool.Tool
		allTools = append(allTools, inputTools...)
		if extra, ok := state["extra_tools"].([]tool.Tool); ok {
			allTools = append(allTools, extra...)
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

		var toolDefs []llms.Tool
		for _, t := range allTools {
			toolSchema := getToolSchema(t)
			toolDefs = append(toolDefs, llms.Tool{
				Type: "function",
				Function: &llms.FunctionDefinition{
					Name:        t.Name(),
					Description: t.Description(),
					Parameters:  toolSchema,
				},
			})
			// Debug logging
			if options.Verbose {
				fmt.Printf("[DEBUG] Tool: %s, Schema: %+v\n", t.Name(), toolSchema)
			}
		}

		if options.Verbose {
			fmt.Printf("[DEBUG] Total tools passed to LLM: %d\n", len(toolDefs))
		}

		msgsToSend := messages
		if options.SystemMessage != "" {
			msgsToSend = append([]llms.MessageContent{llms.TextParts(llms.ChatMessageTypeSystem, options.SystemMessage)}, msgsToSend...)
		}
		if options.StateModifier != nil {
			msgsToSend = options.StateModifier(msgsToSend)
		}

		var resp *llms.ContentResponse
		var err error

		if options.DisableModelInvocation {
			// Skip model invocation, create a dummy response
			resp = &llms.ContentResponse{
				Choices: []*llms.ContentChoice{
					{
						Content:   "",
						ToolCalls: []llms.ToolCall{},
					},
				},
			}
		} else {
			resp, err = model.GenerateContent(ctx, msgsToSend, llms.WithTools(toolDefs), llms.WithToolChoice("auto"))
			if err != nil {
				return nil, err
			}
		}

		// Debug: check for tool calls in response
		if options.Verbose {
			fmt.Printf("[DEBUG] LLM Response - Content: %s, ToolCalls: %d\n", resp.Choices[0].Content, len(resp.Choices[0].ToolCalls))
			for i, tc := range resp.Choices[0].ToolCalls {
				fmt.Printf("[DEBUG]   ToolCall %d: %s (%s)\n", i, tc.FunctionCall.Name, tc.FunctionCall.Arguments)
			}
		}

		choice := resp.Choices[0]
		aiMsg := llms.MessageContent{Role: llms.ChatMessageTypeAI}
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

	workflow.AddNode("tools", "Tool execution node", func(ctx context.Context, state map[string]any) (map[string]any, error) {
		messages := state["messages"].([]llms.MessageContent)
		lastMsg := messages[len(messages)-1]
		var allTools []tool.Tool
		allTools = append(allTools, inputTools...)
		if extra, ok := state["extra_tools"].([]tool.Tool); ok {
			allTools = append(allTools, extra...)
		}
		toolExecutor := NewToolExecutor(allTools)

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

				res, err := toolExecutor.Execute(ctx, ToolInvocation{Tool: tc.FunctionCall.Name, ToolInput: inputVal})
				if err != nil {
					res = fmt.Sprintf("Error: %v", err)
				}
				toolMessages = append(toolMessages, llms.MessageContent{
					Role:  llms.ChatMessageTypeTool,
					Parts: []llms.ContentPart{llms.ToolCallResponse{ToolCallID: tc.ID, Name: tc.FunctionCall.Name, Content: res}},
				})
			}
		}
		return map[string]any{"messages": toolMessages}, nil
	})

	if options.skillDir != "" {
		workflow.SetEntryPoint("skill")
		workflow.AddEdge("skill", "agent")
	} else {
		workflow.SetEntryPoint("agent")
	}

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

// CreateAgent creates a generic agent graph
func CreateAgent[S any](
	model llms.Model,
	inputTools []tool.Tool,
	getMessages func(S) []llms.MessageContent,
	setMessages func(S, []llms.MessageContent) S,
	getExtraTools func(S) []tool.Tool,
	setExtraTools func(S, []tool.Tool) S,
	opts ...CreateAgentOption,
) (*graph.StateRunnable[S], error) {
	options := &CreateAgentOptions{}
	for _, opt := range opts {
		opt(options)
	}

	workflow := graph.NewStateGraph[S]()

	workflow.AddNode("agent", "Agent decision node", func(ctx context.Context, state S) (S, error) {
		messages := getMessages(state)
		allTools := append(inputTools, getExtraTools(state)...)

		var toolDefs []llms.Tool
		for _, t := range allTools {
			toolDefs = append(toolDefs, llms.Tool{
				Type: "function",
				Function: &llms.FunctionDefinition{
					Name:        t.Name(),
					Description: t.Description(),
					Parameters:  getToolSchema(t),
				},
			})
		}

		msgsToSend := messages
		if options.SystemMessage != "" {
			msgsToSend = append([]llms.MessageContent{llms.TextParts(llms.ChatMessageTypeSystem, options.SystemMessage)}, msgsToSend...)
		}
		if options.StateModifier != nil {
			msgsToSend = options.StateModifier(msgsToSend)
		}

		var resp *llms.ContentResponse
		var err error

		if options.DisableModelInvocation {
			// Skip model invocation, create a dummy response
			resp = &llms.ContentResponse{
				Choices: []*llms.ContentChoice{
					{
						Content:   "",
						ToolCalls: []llms.ToolCall{},
					},
				},
			}
		} else {
			resp, err = model.GenerateContent(ctx, msgsToSend, llms.WithTools(toolDefs))
			if err != nil {
				return state, err
			}
		}

		choice := resp.Choices[0]
		aiMsg := llms.MessageContent{Role: llms.ChatMessageTypeAI}
		if choice.Content != "" {
			aiMsg.Parts = append(aiMsg.Parts, llms.TextPart(choice.Content))
		}
		for _, tc := range choice.ToolCalls {
			aiMsg.Parts = append(aiMsg.Parts, tc)
		}

		return setMessages(state, append(messages, aiMsg)), nil
	})

	workflow.AddNode("tools", "Tool execution node", func(ctx context.Context, state S) (S, error) {
		messages := getMessages(state)
		lastMsg := messages[len(messages)-1]
		toolExecutor := NewToolExecutor(append(inputTools, getExtraTools(state)...))

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

				res, err := toolExecutor.Execute(ctx, ToolInvocation{Tool: tc.FunctionCall.Name, ToolInput: inputVal})
				if err != nil {
					res = fmt.Sprintf("Error: %v", err)
				}
				toolMessages = append(toolMessages, llms.MessageContent{
					Role:  llms.ChatMessageTypeTool,
					Parts: []llms.ContentPart{llms.ToolCallResponse{ToolCallID: tc.ID, Name: tc.FunctionCall.Name, Content: res}},
				})
			}
		}
		return setMessages(state, append(messages, toolMessages...)), nil
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

// CreateAgentForState creates a type-safe Agent graph pre-configured for AgentState.
// This provides compile-time type safety without requiring manual getter/setter closures.
func CreateAgentForState(
	model llms.Model,
	inputTools []tool.Tool,
	opts ...CreateAgentOption,
) (*graph.StateRunnable[AgentState], error) {
	return CreateAgent[AgentState](
		model,
		inputTools,
		func(s AgentState) []llms.MessageContent { return s.Messages },
		func(s AgentState, m []llms.MessageContent) AgentState { s.Messages = m; return s },
		func(s AgentState) []tool.Tool { return s.ExtraTools },
		func(s AgentState, t []tool.Tool) AgentState { s.ExtraTools = t; return s },
		opts...,
	)
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
	var skillDescriptions strings.Builder
	for name, pkg := range availableSkills {
		skillDescriptions.WriteString(fmt.Sprintf("- %s: %s\n", name, pkg.Meta.Description))
	}
	prompt := fmt.Sprintf("Select the most appropriate skill for: \"%s\"\n\nSkills:\n%s\nReturn only the skill name or 'None'.", userPrompt, skillDescriptions.String())
	resp, err := model.GenerateContent(ctx, []llms.MessageContent{llms.TextParts(llms.ChatMessageTypeHuman, prompt)})
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(resp.Choices[0].Content), nil
}
