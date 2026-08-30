package prebuilt

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/smallnest/langgraphgo/graph"
	"github.com/smallnest/langgraphgo/llms"
)

// CreateSupervisorMap creates a supervisor graph with map[string]any state
func CreateSupervisorMap(model llms.Model, members map[string]*graph.StateRunnable[map[string]any]) (*graph.StateRunnable[map[string]any], error) {
	workflow := graph.NewStateGraph[map[string]any]()
	schema := graph.NewMapSchema()
	schema.RegisterReducer("messages", graph.AppendReducer)
	workflow.SetSchema(schema)

	var memberNames []string
	for name := range members {
		memberNames = append(memberNames, name)
	}

	workflow.AddNode("supervisor", "Supervisor orchestration node", func(ctx context.Context, state map[string]any) (map[string]any, error) {
		messages, ok := state["messages"].([]llms.MessageContent)
		if !ok {
			return nil, fmt.Errorf("messages key not found or invalid type")
		}

		options := append(memberNames, "FINISH")
		routeTool := llms.Tool{
			Type: "function",
			Function: &llms.FunctionDefinition{
				Name:        "route",
				Description: "Select the next role.",
				Parameters: map[string]any{
					"type": "object",
					"properties": map[string]any{
						"next": map[string]any{
							"type": "string",
							"enum": options,
						},
					},
					"required": []string{"next"},
				},
			},
		}

		systemPrompt := fmt.Sprintf(
			"You are a supervisor tasked with managing a conversation between: %s. Respond with the worker to act next or FINISH. Use the 'route' tool.",
			strings.Join(memberNames, ", "),
		)

		inputMessages := append([]llms.MessageContent{llms.TextParts(llms.ChatMessageTypeSystem, systemPrompt)}, messages...)

		toolChoice := llms.ToolChoice{Type: "function", Function: &llms.FunctionReference{Name: "route"}}
		resp, err := model.GenerateContent(ctx, inputMessages, llms.WithTools([]llms.Tool{routeTool}), llms.WithToolChoice(toolChoice))
		if err != nil {
			return nil, err
		}

		choice := resp.Choices[0]
		if len(choice.ToolCalls) == 0 {
			return nil, fmt.Errorf("supervisor did not select a next step")
		}

		var args struct {
			Next string `json:"next"`
		}
		if err := json.Unmarshal([]byte(choice.ToolCalls[0].FunctionCall.Arguments), &args); err != nil {
			return nil, fmt.Errorf("failed to parse route arguments: %w", err)
		}

		return map[string]any{"next": args.Next}, nil
	})

	for name, agent := range members {
		agentName := name
		agentRunnable := agent
		workflow.AddNode(agentName, "Agent: "+agentName, func(ctx context.Context, state map[string]any) (map[string]any, error) {
			return agentRunnable.Invoke(ctx, state)
		})
	}

	workflow.SetEntryPoint("supervisor")
	workflow.AddConditionalEdge("supervisor", func(ctx context.Context, state map[string]any) string {
		next, _ := state["next"].(string)
		if next == "FINISH" || next == "" {
			return graph.END
		}
		return next
	})

	for _, name := range memberNames {
		workflow.AddEdge(name, "supervisor")
	}

	return workflow.Compile()
}

// CreateSupervisor creates a generic supervisor graph
func CreateSupervisor[S any](
	model llms.Model,
	members map[string]*graph.StateRunnable[S],
	getMessages func(S) []llms.MessageContent,
	getNext func(S) string,
	setNext func(S, string) S,
) (*graph.StateRunnable[S], error) {
	workflow := graph.NewStateGraph[S]()

	var memberNames []string
	for name := range members {
		memberNames = append(memberNames, name)
	}

	workflow.AddNode("supervisor", "Supervisor orchestration node", func(ctx context.Context, state S) (S, error) {
		messages := getMessages(state)
		options := append(memberNames, "FINISH")
		routeTool := llms.Tool{
			Type: "function",
			Function: &llms.FunctionDefinition{
				Name:        "route",
				Description: "Select the next role.",
				Parameters: map[string]any{
					"type": "object",
					"properties": map[string]any{
						"next": map[string]any{
							"type": "string",
							"enum": options,
						},
					},
					"required": []string{"next"},
				},
			},
		}

		systemPrompt := fmt.Sprintf(
			"You are a supervisor tasked with managing a conversation between: %s. Respond with the worker to act next or FINISH. Use the 'route' tool.",
			strings.Join(memberNames, ", "),
		)

		inputMessages := append([]llms.MessageContent{llms.TextParts(llms.ChatMessageTypeSystem, systemPrompt)}, messages...)

		toolChoice := llms.ToolChoice{Type: "function", Function: &llms.FunctionReference{Name: "route"}}
		resp, err := model.GenerateContent(ctx, inputMessages, llms.WithTools([]llms.Tool{routeTool}), llms.WithToolChoice(toolChoice))
		if err != nil {
			return state, err
		}

		choice := resp.Choices[0]
		if len(choice.ToolCalls) == 0 {
			return state, fmt.Errorf("supervisor did not select a next step")
		}

		var args struct {
			Next string `json:"next"`
		}
		if err := json.Unmarshal([]byte(choice.ToolCalls[0].FunctionCall.Arguments), &args); err != nil {
			return state, fmt.Errorf("failed to parse route arguments: %w", err)
		}

		return setNext(state, args.Next), nil
	})

	for name, runnable := range members {
		agentName := name
		agentRunnable := runnable
		workflow.AddNode(agentName, "Agent: "+agentName, func(ctx context.Context, state S) (S, error) {
			return agentRunnable.Invoke(ctx, state)
		})
	}

	workflow.SetEntryPoint("supervisor")
	workflow.AddConditionalEdge("supervisor", func(ctx context.Context, state S) string {
		next := getNext(state)
		if next == "FINISH" || next == "" {
			return graph.END
		}
		return next
	})

	for name := range members {
		workflow.AddEdge(name, "supervisor")
	}

	return workflow.Compile()
}

// CreateSupervisorForState creates a type-safe supervisor graph pre-configured for SupervisorState.
// This provides compile-time type safety without requiring manual getter/setter closures.
func CreateSupervisorForState(
	model llms.Model,
	members map[string]*graph.StateRunnable[SupervisorState],
) (*graph.StateRunnable[SupervisorState], error) {
	return CreateSupervisor[SupervisorState](
		model,
		members,
		func(s SupervisorState) []llms.MessageContent { return s.Messages },
		func(s SupervisorState) string { return s.Next },
		func(s SupervisorState, n string) SupervisorState { s.Next = n; return s },
	)
}
