package prebuilt

import (
	"context"
	"fmt"
	"strings"

	"github.com/smallnest/langgraphgo/graph"
	"github.com/smallnest/langgraphgo/llms"
)

// ReflectionAgentConfig configures the reflection agent
type ReflectionAgentConfig struct {
	Model            llms.Model
	ReflectionModel  llms.Model
	MaxIterations    int
	SystemMessage    string
	ReflectionPrompt string
	Verbose          bool
}

// CreateReflectionAgentMap creates a new Reflection Agent with map[string]any state
func CreateReflectionAgentMap(config ReflectionAgentConfig) (*graph.StateRunnable[map[string]any], error) {
	if config.Model == nil {
		return nil, fmt.Errorf("model is required")
	}
	if config.MaxIterations == 0 {
		config.MaxIterations = 3
	}
	reflectionModel := config.ReflectionModel
	if reflectionModel == nil {
		reflectionModel = config.Model
	}
	if config.SystemMessage == "" {
		config.SystemMessage = "You are a helpful assistant. Generate a high-quality response to the user's request."
	}
	if config.ReflectionPrompt == "" {
		config.ReflectionPrompt = buildDefaultReflectionPrompt()
	}

	workflow := graph.NewStateGraph[map[string]any]()
	agentSchema := graph.NewMapSchema()
	agentSchema.RegisterReducer("messages", graph.AppendReducer)
	workflow.SetSchema(agentSchema)

	workflow.AddNode("generate", "Generate or revise response", func(ctx context.Context, state map[string]any) (map[string]any, error) {
		iteration, _ := state["iteration"].(int)
		messages, ok := state["messages"].([]llms.MessageContent)
		if !ok || len(messages) == 0 {
			return nil, fmt.Errorf("no messages found")
		}

		var promptMessages []llms.MessageContent
		if iteration == 0 {
			promptMessages = append([]llms.MessageContent{{Role: llms.ChatMessageTypeSystem, Parts: []llms.ContentPart{llms.TextPart(config.SystemMessage)}}}, messages...)
		} else {
			reflection, _ := state["reflection"].(string)
			draft, _ := state["draft"].(string)
			revisionPrompt := fmt.Sprintf("Revise based on reflection:\nRequest: %s\nDraft: %s\nReflection: %s", getOriginalRequest(messages), draft, reflection)
			promptMessages = []llms.MessageContent{
				{Role: llms.ChatMessageTypeSystem, Parts: []llms.ContentPart{llms.TextPart(config.SystemMessage)}},
				{Role: llms.ChatMessageTypeHuman, Parts: []llms.ContentPart{llms.TextPart(revisionPrompt)}},
			}
		}

		resp, err := config.Model.GenerateContent(ctx, promptMessages)
		if err != nil {
			return nil, err
		}
		draft := resp.Choices[0].Content
		return map[string]any{
			"messages":  []llms.MessageContent{{Role: llms.ChatMessageTypeAI, Parts: []llms.ContentPart{llms.TextPart(draft)}}},
			"draft":     draft,
			"iteration": iteration + 1,
		}, nil
	})

	workflow.AddNode("reflect", "Reflect on response", func(ctx context.Context, state map[string]any) (map[string]any, error) {
		draft, _ := state["draft"].(string)
		messages := state["messages"].([]llms.MessageContent)
		reflectionMessages := []llms.MessageContent{
			{Role: llms.ChatMessageTypeSystem, Parts: []llms.ContentPart{llms.TextPart(config.ReflectionPrompt)}},
			{Role: llms.ChatMessageTypeHuman, Parts: []llms.ContentPart{llms.TextPart(fmt.Sprintf("Request: %s\nResponse: %s", getOriginalRequest(messages), draft))}},
		}
		resp, err := reflectionModel.GenerateContent(ctx, reflectionMessages)
		if err != nil {
			return nil, err
		}
		return map[string]any{"reflection": resp.Choices[0].Content}, nil
	})

	workflow.SetEntryPoint("generate")
	workflow.AddConditionalEdge("generate", func(ctx context.Context, state map[string]any) string {
		iteration, _ := state["iteration"].(int)
		if iteration >= config.MaxIterations {
			return graph.END
		}
		return "reflect"
	})
	workflow.AddConditionalEdge("reflect", func(ctx context.Context, state map[string]any) string {
		reflection, _ := state["reflection"].(string)
		if isResponseSatisfactory(reflection) {
			return graph.END
		}
		return "generate"
	})

	return workflow.Compile()
}

// CreateReflectionAgent creates a generic reflection agent
func CreateReflectionAgent[S any](
	config ReflectionAgentConfig,
	getMessages func(S) []llms.MessageContent,
	setMessages func(S, []llms.MessageContent) S,
	getDraft func(S) string,
	setDraft func(S, string) S,
	getIteration func(S) int,
	setIteration func(S, int) S,
	getReflection func(S) string,
	setReflection func(S, string) S,
) (*graph.StateRunnable[S], error) {
	if config.Model == nil {
		return nil, fmt.Errorf("model is required")
	}
	if config.MaxIterations == 0 {
		config.MaxIterations = 3
	}
	reflectionModel := config.ReflectionModel
	if reflectionModel == nil {
		reflectionModel = config.Model
	}
	if config.SystemMessage == "" {
		config.SystemMessage = "You are a helpful assistant. Generate a high-quality response to the user's request."
	}
	if config.ReflectionPrompt == "" {
		config.ReflectionPrompt = buildDefaultReflectionPrompt()
	}

	workflow := graph.NewStateGraph[S]()

	workflow.AddNode("generate", "Generate or revise response", func(ctx context.Context, state S) (S, error) {
		iteration := getIteration(state)
		messages := getMessages(state)
		if len(messages) == 0 {
			return state, fmt.Errorf("no messages found")
		}

		var promptMessages []llms.MessageContent
		if iteration == 0 {
			promptMessages = append([]llms.MessageContent{{Role: llms.ChatMessageTypeSystem, Parts: []llms.ContentPart{llms.TextPart(config.SystemMessage)}}}, messages...)
		} else {
			reflection := getReflection(state)
			draft := getDraft(state)
			revisionPrompt := fmt.Sprintf("Revise based on reflection:\nRequest: %s\nDraft: %s\nReflection: %s", getOriginalRequest(messages), draft, reflection)
			promptMessages = []llms.MessageContent{
				{Role: llms.ChatMessageTypeSystem, Parts: []llms.ContentPart{llms.TextPart(config.SystemMessage)}},
				{Role: llms.ChatMessageTypeHuman, Parts: []llms.ContentPart{llms.TextPart(revisionPrompt)}},
			}
		}

		resp, err := config.Model.GenerateContent(ctx, promptMessages)
		if err != nil {
			return state, err
		}
		draft := resp.Choices[0].Content
		aiMsg := llms.MessageContent{Role: llms.ChatMessageTypeAI, Parts: []llms.ContentPart{llms.TextPart(draft)}}
		state = setMessages(state, append(messages, aiMsg))
		state = setDraft(state, draft)
		state = setIteration(state, iteration+1)
		return state, nil
	})

	workflow.AddNode("reflect", "Reflect on response", func(ctx context.Context, state S) (S, error) {
		draft := getDraft(state)
		messages := getMessages(state)
		reflectionMessages := []llms.MessageContent{
			{Role: llms.ChatMessageTypeSystem, Parts: []llms.ContentPart{llms.TextPart(config.ReflectionPrompt)}},
			{Role: llms.ChatMessageTypeHuman, Parts: []llms.ContentPart{llms.TextPart(fmt.Sprintf("Request: %s\nResponse: %s", getOriginalRequest(messages), draft))}},
		}
		resp, err := reflectionModel.GenerateContent(ctx, reflectionMessages)
		if err != nil {
			return state, err
		}
		state = setReflection(state, resp.Choices[0].Content)
		return state, nil
	})

	workflow.SetEntryPoint("generate")
	workflow.AddConditionalEdge("generate", func(ctx context.Context, state S) string {
		if getIteration(state) >= config.MaxIterations {
			return graph.END
		}
		return "reflect"
	})
	workflow.AddConditionalEdge("reflect", func(ctx context.Context, state S) string {
		if isResponseSatisfactory(getReflection(state)) {
			return graph.END
		}
		return "generate"
	})

	return workflow.Compile()
}

// CreateReflectionAgentForState creates a type-safe Reflection Agent graph pre-configured for ReflectionAgentState.
// This provides compile-time type safety without requiring manual getter/setter closures.
func CreateReflectionAgentForState(config ReflectionAgentConfig) (*graph.StateRunnable[ReflectionAgentState], error) {
	return CreateReflectionAgent[ReflectionAgentState](
		config,
		func(s ReflectionAgentState) []llms.MessageContent { return s.Messages },
		func(s ReflectionAgentState, m []llms.MessageContent) ReflectionAgentState { s.Messages = m; return s },
		func(s ReflectionAgentState) string { return s.Draft },
		func(s ReflectionAgentState, d string) ReflectionAgentState { s.Draft = d; return s },
		func(s ReflectionAgentState) int { return s.Iteration },
		func(s ReflectionAgentState, i int) ReflectionAgentState { s.Iteration = i; return s },
		func(s ReflectionAgentState) string { return s.Reflection },
		func(s ReflectionAgentState, r string) ReflectionAgentState { s.Reflection = r; return s },
	)
}

func isResponseSatisfactory(reflection string) bool {
	reflectionLower := strings.ToLower(reflection)
	satisfactoryKeywords := []string{"excellent", "satisfactory", "no major issues", "well done", "accurate", "meets all requirements"}
	for _, keyword := range satisfactoryKeywords {
		if strings.Contains(reflectionLower, keyword) {
			return true
		}
	}
	return false
}

func getOriginalRequest(messages []llms.MessageContent) string {
	for _, msg := range messages {
		if msg.Role == llms.ChatMessageTypeHuman {
			for _, part := range msg.Parts {
				if textPart, ok := part.(llms.TextContent); ok {
					return textPart.Text
				}
			}
		}
	}
	return ""
}

func buildDefaultReflectionPrompt() string {
	return `You are a critical reviewer. Evaluate the response and provide strengths, weaknesses and suggestions.`
}
