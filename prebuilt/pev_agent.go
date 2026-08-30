package prebuilt

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/smallnest/langgraphgo/graph"
	"github.com/smallnest/langgraphgo/llms"
	"github.com/smallnest/langgraphgo/tool"
)

// PEVAgentConfig configures a Plan-Execute-Verify agent.
//
// Deprecated: prefer the shared functional options on CreatePEVAgentMap
// (WithMaxRetries, WithSystemMessage, WithVerificationPrompt, WithVerbose), which
// align the PEV agent with the rest of the plan-execute family. The config struct
// remains supported; non-zero config fields take precedence over options.
type PEVAgentConfig struct {
	Model              llms.Model
	Tools              []tool.Tool
	MaxRetries         int
	SystemMessage      string
	VerificationPrompt string
	Verbose            bool
}

// VerificationResult represents the result of verification
type VerificationResult struct {
	IsSuccessful bool   `json:"is_successful"`
	Reasoning    string `json:"reasoning"`
}

// applyPEVOptions folds shared functional options into a PEVAgentConfig. Non-zero
// config fields win; options only fill fields the caller left unset.
func applyPEVOptions(config PEVAgentConfig, opts ...CreateAgentOption) PEVAgentConfig {
	if len(opts) == 0 {
		return config
	}
	o := &CreateAgentOptions{}
	for _, opt := range opts {
		opt(o)
	}
	if config.MaxRetries == 0 {
		config.MaxRetries = o.MaxRetries
	}
	if config.SystemMessage == "" {
		config.SystemMessage = o.SystemMessage
	}
	if config.VerificationPrompt == "" {
		config.VerificationPrompt = o.VerificationPrompt
	}
	if !config.Verbose {
		config.Verbose = o.Verbose
	}
	return config
}

// CreatePEVAgentMap creates a new PEV Agent with map[string]any state.
//
// Behavior can be configured via the PEVAgentConfig struct or the shared
// functional options (WithMaxRetries, WithSystemMessage, WithVerificationPrompt,
// WithVerbose). Non-zero config fields take precedence over options.
func CreatePEVAgentMap(config PEVAgentConfig, opts ...CreateAgentOption) (*graph.StateRunnable[map[string]any], error) {
	config = applyPEVOptions(config, opts...)
	if config.Model == nil {
		return nil, fmt.Errorf("model is required")
	}
	if config.MaxRetries == 0 {
		config.MaxRetries = 3
	}
	if config.SystemMessage == "" {
		config.SystemMessage = buildPEVDefaultPlannerPrompt()
	}
	if config.VerificationPrompt == "" {
		config.VerificationPrompt = buildPEVDefaultVerificationPrompt()
	}

	toolExecutor := NewToolExecutor(config.Tools)
	workflow := graph.NewStateGraph[map[string]any]()
	agentSchema := graph.NewMapSchema()
	agentSchema.RegisterReducer("messages", graph.AppendReducer)
	agentSchema.RegisterReducer("intermediate_steps", graph.AppendReducer)
	workflow.SetSchema(agentSchema)

	workflow.AddNode("planner", "Create or revise execution plan", func(ctx context.Context, state map[string]any) (map[string]any, error) {
		retries, _ := state["retries"].(int)
		messages, ok := state["messages"].([]llms.MessageContent)
		if !ok || len(messages) == 0 {
			return nil, fmt.Errorf("no messages found")
		}

		var promptMessages []llms.MessageContent
		if retries == 0 {
			promptMessages = append([]llms.MessageContent{{Role: llms.ChatMessageTypeSystem, Parts: []llms.ContentPart{llms.TextPart(config.SystemMessage)}}}, messages...)
		} else {
			lastResult, _ := state["last_tool_result"].(string)
			vResult, _ := state["verification_result"].(string)
			replanPrompt := fmt.Sprintf("Previous failed verification. New plan needed.\nRequest: %s\nLast result: %s\nFeedback: %s", getPEVOriginalRequest(messages), lastResult, vResult)
			promptMessages = []llms.MessageContent{
				{Role: llms.ChatMessageTypeSystem, Parts: []llms.ContentPart{llms.TextPart(config.SystemMessage)}},
				{Role: llms.ChatMessageTypeHuman, Parts: []llms.ContentPart{llms.TextPart(replanPrompt)}},
			}
		}

		resp, err := config.Model.GenerateContent(ctx, promptMessages)
		if err != nil {
			return nil, err
		}
		steps := parsePEVPlanSteps(resp.Choices[0].Content)
		return map[string]any{"plan": steps, "current_step": 0}, nil
	})

	workflow.AddNode("executor", "Execute step", func(ctx context.Context, state map[string]any) (map[string]any, error) {
		plan, _ := state["plan"].([]string)
		currentStep, _ := state["current_step"].(int)
		if currentStep >= len(plan) {
			return nil, fmt.Errorf("step out of bounds")
		}
		stepDesc := plan[currentStep]

		result, err := executePEVStep(ctx, stepDesc, toolExecutor, config.Model)
		if err != nil {
			result = fmt.Sprintf("Error: %v", err)
		}

		return map[string]any{
			"last_tool_result":   result,
			"intermediate_steps": []string{fmt.Sprintf("Step %d: %s -> %s", currentStep+1, stepDesc, result)},
		}, nil
	})

	workflow.AddNode("verifier", "Verify result", func(ctx context.Context, state map[string]any) (map[string]any, error) {
		lastResult, _ := state["last_tool_result"].(string)
		plan, _ := state["plan"].([]string)
		currentStep, _ := state["current_step"].(int)
		stepDesc := plan[currentStep]

		verifyPrompt := fmt.Sprintf("Verify: Action: %s\nResult: %s", stepDesc, lastResult)
		promptMessages := []llms.MessageContent{
			{Role: llms.ChatMessageTypeSystem, Parts: []llms.ContentPart{llms.TextPart(config.VerificationPrompt)}},
			{Role: llms.ChatMessageTypeHuman, Parts: []llms.ContentPart{llms.TextPart(verifyPrompt)}},
		}
		resp, err := config.Model.GenerateContent(ctx, promptMessages)
		if err != nil {
			return nil, err
		}

		var vResult VerificationResult
		_ = json.Unmarshal([]byte(extractPEVJSON(resp.Choices[0].Content)), &vResult)
		return map[string]any{"verification_result": vResult.Reasoning, "is_successful": vResult.IsSuccessful}, nil
	})

	workflow.AddNode("synthesizer", "Synthesize final answer", func(ctx context.Context, state map[string]any) (map[string]any, error) {
		messages, _ := state["messages"].([]llms.MessageContent)
		steps, _ := state["intermediate_steps"].([]string)
		prompt := fmt.Sprintf("Synthesize: Request: %s\nSteps: %s", getPEVOriginalRequest(messages), strings.Join(steps, "\n"))
		resp, err := config.Model.GenerateContent(ctx, []llms.MessageContent{{Role: llms.ChatMessageTypeHuman, Parts: []llms.ContentPart{llms.TextPart(prompt)}}})
		if err != nil {
			return nil, err
		}
		answer := resp.Choices[0].Content
		return map[string]any{
			"messages":     []llms.MessageContent{{Role: llms.ChatMessageTypeAI, Parts: []llms.ContentPart{llms.TextPart(answer)}}},
			"final_answer": answer,
		}, nil
	})

	workflow.SetEntryPoint("planner")
	workflow.AddConditionalEdge("planner", func(ctx context.Context, state map[string]any) string {
		if p, ok := state["plan"].([]string); ok && len(p) > 0 {
			return "executor"
		}
		return graph.END
	})
	workflow.AddEdge("executor", "verifier")
	workflow.AddConditionalEdge("verifier", func(ctx context.Context, state map[string]any) string {
		success, _ := state["is_successful"].(bool)
		currentStep, _ := state["current_step"].(int)
		plan, _ := state["plan"].([]string)
		if success {
			if currentStep+1 >= len(plan) {
				return "synthesizer"
			}
			state["current_step"] = currentStep + 1
			return "executor"
		}
		retries, _ := state["retries"].(int)
		if retries >= config.MaxRetries {
			return "synthesizer"
		}
		state["retries"] = retries + 1
		return "planner"
	})
	workflow.AddEdge("synthesizer", graph.END)

	return workflow.Compile()
}

// CreatePEVAgent creates a generic PEV Agent
func CreatePEVAgent[S any](
	config PEVAgentConfig,
	getMessages func(S) []llms.MessageContent,
	setMessages func(S, []llms.MessageContent) S,
	getPlan func(S) []string,
	setPlan func(S, []string) S,
	getCurrentStep func(S) int,
	setCurrentStep func(S, int) S,
	getLastToolResult func(S) string,
	setLastToolResult func(S, string) S,
	getIntermediateSteps func(S) []string,
	setIntermediateSteps func(S, []string) S,
	getRetries func(S) int,
	setRetries func(S, int) S,
	getVerificationResult func(S) string,
	setVerificationResult func(S, string) S,
	getFinalAnswer func(S) string,
	setFinalAnswer func(S, string) S,
	opts ...CreateAgentOption,
) (*graph.StateRunnable[S], error) {
	config = applyPEVOptions(config, opts...)
	if config.Model == nil {
		return nil, fmt.Errorf("model is required")
	}
	if config.MaxRetries == 0 {
		config.MaxRetries = 3
	}
	if config.SystemMessage == "" {
		config.SystemMessage = buildPEVDefaultPlannerPrompt()
	}
	if config.VerificationPrompt == "" {
		config.VerificationPrompt = buildPEVDefaultVerificationPrompt()
	}

	toolExecutor := NewToolExecutor(config.Tools)
	workflow := graph.NewStateGraph[S]()

	workflow.AddNode("planner", "Create or revise execution plan", func(ctx context.Context, state S) (S, error) {
		retries := getRetries(state)
		messages := getMessages(state)
		if len(messages) == 0 {
			return state, fmt.Errorf("no messages")
		}

		var promptMessages []llms.MessageContent
		if retries == 0 {
			promptMessages = append([]llms.MessageContent{{Role: llms.ChatMessageTypeSystem, Parts: []llms.ContentPart{llms.TextPart(config.SystemMessage)}}}, messages...)
		} else {
			replanPrompt := fmt.Sprintf("Re-plan: Request: %s\nLast result: %s\nFeedback: %s", getPEVOriginalRequest(messages), getLastToolResult(state), getVerificationResult(state))
			promptMessages = []llms.MessageContent{
				{Role: llms.ChatMessageTypeSystem, Parts: []llms.ContentPart{llms.TextPart(config.SystemMessage)}},
				{Role: llms.ChatMessageTypeHuman, Parts: []llms.ContentPart{llms.TextPart(replanPrompt)}},
			}
		}

		resp, err := config.Model.GenerateContent(ctx, promptMessages)
		if err != nil {
			return state, err
		}
		state = setPlan(state, parsePEVPlanSteps(resp.Choices[0].Content))
		state = setCurrentStep(state, 0)
		return state, nil
	})

	workflow.AddNode("executor", "Execute step", func(ctx context.Context, state S) (S, error) {
		plan := getPlan(state)
		currentStep := getCurrentStep(state)
		if currentStep >= len(plan) {
			return state, fmt.Errorf("out of bounds")
		}
		result, err := executePEVStep(ctx, plan[currentStep], toolExecutor, config.Model)
		if err != nil {
			result = "Error: " + err.Error()
		}
		state = setLastToolResult(state, result)
		state = setIntermediateSteps(state, append(getIntermediateSteps(state), fmt.Sprintf("Step %d: %s -> %s", currentStep+1, plan[currentStep], result)))
		return state, nil
	})

	workflow.AddNode("verifier", "Verify result", func(ctx context.Context, state S) (S, error) {
		prompt := fmt.Sprintf("Verify: Action: %s\nResult: %s", getPlan(state)[getCurrentStep(state)], getLastToolResult(state))
		resp, err := config.Model.GenerateContent(ctx, []llms.MessageContent{
			{Role: llms.ChatMessageTypeSystem, Parts: []llms.ContentPart{llms.TextPart(config.VerificationPrompt)}},
			{Role: llms.ChatMessageTypeHuman, Parts: []llms.ContentPart{llms.TextPart(prompt)}},
		})
		if err != nil {
			return state, err
		}
		var vResult VerificationResult
		_ = json.Unmarshal([]byte(extractPEVJSON(resp.Choices[0].Content)), &vResult)
		// We need a way to pass isSuccessful to the router. For generic S, we can't easily add a field.
		// So we encode it in VerificationResult string or assume the state can hold it.
		if vResult.IsSuccessful {
			state = setVerificationResult(state, "SUCCESS: "+vResult.Reasoning)
		} else {
			state = setVerificationResult(state, "FAILED: "+vResult.Reasoning)
		}
		return state, nil
	})

	workflow.AddNode("synthesizer", "Synthesize final answer", func(ctx context.Context, state S) (S, error) {
		prompt := fmt.Sprintf("Synthesize: Request: %s\nSteps: %s", getPEVOriginalRequest(getMessages(state)), strings.Join(getIntermediateSteps(state), "\n"))
		resp, err := config.Model.GenerateContent(ctx, []llms.MessageContent{{Role: llms.ChatMessageTypeHuman, Parts: []llms.ContentPart{llms.TextPart(prompt)}}})
		if err != nil {
			return state, err
		}
		answer := resp.Choices[0].Content
		state = setMessages(state, append(getMessages(state), llms.MessageContent{Role: llms.ChatMessageTypeAI, Parts: []llms.ContentPart{llms.TextPart(answer)}}))
		state = setFinalAnswer(state, answer)
		return state, nil
	})

	workflow.SetEntryPoint("planner")
	workflow.AddConditionalEdge("planner", func(ctx context.Context, state S) string {
		if len(getPlan(state)) > 0 {
			return "executor"
		}
		return graph.END
	})
	workflow.AddEdge("executor", "verifier")
	workflow.AddConditionalEdge("verifier", func(ctx context.Context, state S) string {
		vResult := getVerificationResult(state)
		success := strings.HasPrefix(vResult, "SUCCESS:")
		currentStep := getCurrentStep(state)
		plan := getPlan(state)
		if success {
			if currentStep+1 >= len(plan) {
				return "synthesizer"
			}
			setCurrentStep(state, currentStep+1)
			return "executor"
		}
		retries := getRetries(state)
		if retries >= config.MaxRetries {
			return "synthesizer"
		}
		setRetries(state, retries+1)
		return "planner"
	})
	workflow.AddEdge("synthesizer", graph.END)

	return workflow.Compile()
}

// CreatePEVAgentForState creates a type-safe PEV Agent graph pre-configured for PEVAgentState.
// This provides compile-time type safety without requiring manual getter/setter closures.
func CreatePEVAgentForState(config PEVAgentConfig, opts ...CreateAgentOption) (*graph.StateRunnable[PEVAgentState], error) {
	return CreatePEVAgent[PEVAgentState](
		config,
		func(s PEVAgentState) []llms.MessageContent { return s.Messages },
		func(s PEVAgentState, m []llms.MessageContent) PEVAgentState { s.Messages = m; return s },
		func(s PEVAgentState) []string { return s.Plan },
		func(s PEVAgentState, p []string) PEVAgentState { s.Plan = p; return s },
		func(s PEVAgentState) int { return s.CurrentStep },
		func(s PEVAgentState, i int) PEVAgentState { s.CurrentStep = i; return s },
		func(s PEVAgentState) string { return s.LastToolResult },
		func(s PEVAgentState, r string) PEVAgentState { s.LastToolResult = r; return s },
		func(s PEVAgentState) []string { return s.IntermediateSteps },
		func(s PEVAgentState, steps []string) PEVAgentState { s.IntermediateSteps = steps; return s },
		func(s PEVAgentState) int { return s.Retries },
		func(s PEVAgentState, r int) PEVAgentState { s.Retries = r; return s },
		func(s PEVAgentState) string { return s.VerificationResult },
		func(s PEVAgentState, v string) PEVAgentState { s.VerificationResult = v; return s },
		func(s PEVAgentState) string { return s.FinalAnswer },
		func(s PEVAgentState, a string) PEVAgentState { s.FinalAnswer = a; return s },
		opts...,
	)
}

func parsePEVPlanSteps(planText string) []string {
	var steps []string
	for line := range strings.SplitSeq(planText, "\n") {
		if line = strings.TrimSpace(line); line != "" {
			steps = append(steps, line)
		}
	}
	return steps
}

func executePEVStep(ctx context.Context, step string, te *ToolExecutor, model llms.Model) (string, error) {
	if te == nil || len(te.Tools) == 0 {
		return "Error: No tools", nil
	}
	var toolsInfo strings.Builder
	for name, tool := range te.Tools {
		toolsInfo.WriteString(fmt.Sprintf("- %s: %s\n", name, tool.Description()))
	}
	prompt := fmt.Sprintf("Select tool for: %s\nTools:\n%s\nReturn JSON: {\"tool\": \"name\", \"tool_input\": \"input\"}", step, toolsInfo.String())
	resp, err := model.GenerateContent(ctx, []llms.MessageContent{{Role: llms.ChatMessageTypeHuman, Parts: []llms.ContentPart{llms.TextPart(prompt)}}})
	if err != nil {
		return "", err
	}
	var inv ToolInvocation
	if err := json.Unmarshal([]byte(extractPEVJSON(resp.Choices[0].Content)), &inv); err != nil {
		return "", err
	}
	return te.Execute(ctx, inv)
}

func extractPEVJSON(text string) string {
	start := strings.Index(text, "{")
	end := strings.LastIndex(text, "}")
	if start != -1 && end != -1 {
		return text[start : end+1]
	}
	return text
}

func getPEVOriginalRequest(messages []llms.MessageContent) string {
	for _, m := range messages {
		if m.Role == llms.ChatMessageTypeHuman {
			for _, p := range m.Parts {
				if t, ok := p.(llms.TextContent); ok {
					return t.Text
				}
			}
		}
	}
	return ""
}

func buildPEVDefaultPlannerPrompt() string {
	return "Expert planner. Break request into numbered steps."
}
func buildPEVDefaultVerificationPrompt() string {
	return "Verification specialist. Determine success/failure. Return JSON with is_successful and reasoning."
}
