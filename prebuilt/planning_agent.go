package prebuilt

import (
	"context"
	"encoding/json"
	"fmt"
	"regexp"
	"strings"

	"github.com/smallnest/langgraphgo/graph"
	"github.com/smallnest/langgraphgo/llms"
	"github.com/smallnest/langgraphgo/tool"
)

// CreatePlanningAgentMap creates a planning agent with map[string]any state
func CreatePlanningAgentMap(model llms.Model, availableNodes []graph.TypedNode[map[string]any], inputTools []tool.Tool, opts ...CreateAgentOption) (*graph.StateRunnable[map[string]any], error) {
	options := &CreateAgentOptions{}
	for _, opt := range opts {
		opt(options)
	}

	nodeMap := make(map[string]graph.TypedNode[map[string]any])
	for _, node := range availableNodes {
		nodeMap[node.Name] = node
	}

	workflow := graph.NewStateGraph[map[string]any]()
	agentSchema := graph.NewMapSchema()
	agentSchema.RegisterReducer("messages", graph.AppendReducer)
	agentSchema.RegisterReducer("workflow_plan", graph.OverwriteReducer)
	workflow.SetSchema(agentSchema)

	workflow.AddNode("planner", "Generates workflow plan", func(ctx context.Context, state map[string]any) (map[string]any, error) {
		messages, ok := state["messages"].([]llms.MessageContent)
		if !ok {
			return nil, fmt.Errorf("messages not found")
		}
		nodeDescriptions := buildPlanningNodeDescriptions(availableNodes)
		planningPrompt := buildPlanningPrompt(nodeDescriptions)
		planningMessages := []llms.MessageContent{
			{Role: llms.ChatMessageTypeSystem, Parts: []llms.ContentPart{llms.TextPart(planningPrompt)}},
		}
		planningMessages = append(planningMessages, messages...)

		resp, err := model.GenerateContent(ctx, planningMessages)
		if err != nil {
			return nil, err
		}

		planText := resp.Choices[0].Content
		workflowPlan, err := parseWorkflowPlan(planText)
		if err != nil {
			return nil, err
		}

		aiMsg := llms.MessageContent{
			Role:  llms.ChatMessageTypeAI,
			Parts: []llms.ContentPart{llms.TextPart(fmt.Sprintf("Workflow plan created with %d nodes and %d edges", len(workflowPlan.Nodes), len(workflowPlan.Edges)))},
		}

		return map[string]any{
			"messages":      []llms.MessageContent{aiMsg},
			"workflow_plan": workflowPlan,
		}, nil
	})

	workflow.AddNode("executor", "Executes the planned workflow", func(ctx context.Context, state map[string]any) (map[string]any, error) {
		workflowPlan, ok := state["workflow_plan"].(*WorkflowPlan)
		if !ok {
			return nil, fmt.Errorf("workflow_plan not found in state")
		}

		dynamicWorkflow := graph.NewStateGraph[map[string]any]()
		dynamicSchema := graph.NewMapSchema()
		dynamicSchema.RegisterReducer("messages", graph.AppendReducer)
		dynamicWorkflow.SetSchema(dynamicSchema)

		for _, planNode := range workflowPlan.Nodes {
			if planNode.Name == "START" || planNode.Name == "END" {
				continue
			}
			actualNode, exists := nodeMap[planNode.Name]
			if !exists {
				return nil, fmt.Errorf("node %s not found", planNode.Name)
			}
			dynamicWorkflow.AddNode(actualNode.Name, actualNode.Description, actualNode.Function)
		}

		var entryPoint string
		endNodes := make(map[string]bool)
		for _, edge := range workflowPlan.Edges {
			if edge.From == "START" {
				entryPoint = edge.To
				continue
			}
			if edge.To == "END" {
				endNodes[edge.From] = true
				continue
			}
			if edge.Condition != "" {
				dynamicWorkflow.AddConditionalEdge(edge.From, func(ctx context.Context, state map[string]any) string {
					return edge.To
				})
			} else {
				dynamicWorkflow.AddEdge(edge.From, edge.To)
			}
		}

		for nodeName := range endNodes {
			dynamicWorkflow.AddEdge(nodeName, graph.END)
		}

		if entryPoint == "" {
			return nil, fmt.Errorf("no entry point in plan")
		}
		dynamicWorkflow.SetEntryPoint(entryPoint)

		runnable, err := dynamicWorkflow.Compile()
		if err != nil {
			return nil, err
		}

		return runnable.Invoke(ctx, state)
	})

	workflow.SetEntryPoint("planner")
	workflow.AddEdge("planner", "executor")
	workflow.AddEdge("executor", graph.END)

	return workflow.Compile()
}

// CreatePlanningAgent creates a generic planning agent
func CreatePlanningAgent[S any](
	model llms.Model,
	availableNodes []graph.TypedNode[S],
	getMessages func(S) []llms.MessageContent,
	setMessages func(S, []llms.MessageContent) S,
	getPlan func(S) *WorkflowPlan,
	setPlan func(S, *WorkflowPlan) S,
	opts ...CreateAgentOption,
) (*graph.StateRunnable[S], error) {
	options := &CreateAgentOptions{}
	for _, opt := range opts {
		opt(options)
	}

	nodeMap := make(map[string]graph.TypedNode[S])
	for _, node := range availableNodes {
		nodeMap[node.Name] = node
	}

	workflow := graph.NewStateGraph[S]()

	workflow.AddNode("planner", "Generates workflow plan", func(ctx context.Context, state S) (S, error) {
		messages := getMessages(state)
		if len(messages) == 0 {
			return state, fmt.Errorf("no messages found in state")
		}

		nodeDescriptions := buildPlanningNodeDescriptions(availableNodes)
		planningPrompt := buildPlanningPrompt(nodeDescriptions)

		planningMessages := []llms.MessageContent{
			{Role: llms.ChatMessageTypeSystem, Parts: []llms.ContentPart{llms.TextPart(planningPrompt)}},
		}
		planningMessages = append(planningMessages, messages...)

		resp, err := model.GenerateContent(ctx, planningMessages)
		if err != nil {
			return state, err
		}

		planText := resp.Choices[0].Content
		workflowPlan, err := parseWorkflowPlan(planText)
		if err != nil {
			return state, err
		}

		aiMsg := llms.MessageContent{
			Role:  llms.ChatMessageTypeAI,
			Parts: []llms.ContentPart{llms.TextPart(fmt.Sprintf("Workflow plan created with %d nodes and %d edges", len(workflowPlan.Nodes), len(workflowPlan.Edges)))},
		}

		state = setMessages(state, append(messages, aiMsg))
		state = setPlan(state, workflowPlan)
		return state, nil
	})

	workflow.AddNode("executor", "Executes the planned workflow", func(ctx context.Context, state S) (S, error) {
		workflowPlan := getPlan(state)
		if workflowPlan == nil {
			return state, fmt.Errorf("workflow_plan not found in state")
		}

		dynamicWorkflow := graph.NewStateGraph[S]()
		// Note: We can't easily use Schema here without knowing more about S
		// So we assume nodes handle their own state merging if needed or S is simple

		for _, planNode := range workflowPlan.Nodes {
			if planNode.Name == "START" || planNode.Name == "END" {
				continue
			}
			actualNode, exists := nodeMap[planNode.Name]
			if !exists {
				return state, fmt.Errorf("node %s not found", planNode.Name)
			}
			dynamicWorkflow.AddNode(actualNode.Name, actualNode.Description, actualNode.Function)
		}

		var entryPoint string
		endNodes := make(map[string]bool)
		for _, edge := range workflowPlan.Edges {
			if edge.From == "START" {
				entryPoint = edge.To
				continue
			}
			if edge.To == "END" {
				endNodes[edge.From] = true
				continue
			}
			if edge.Condition != "" {
				dynamicWorkflow.AddConditionalEdge(edge.From, func(ctx context.Context, s S) string {
					return edge.To
				})
			} else {
				dynamicWorkflow.AddEdge(edge.From, edge.To)
			}
		}

		for nodeName := range endNodes {
			dynamicWorkflow.AddEdge(nodeName, graph.END)
		}

		if entryPoint == "" {
			return state, fmt.Errorf("no entry point in plan")
		}
		dynamicWorkflow.SetEntryPoint(entryPoint)

		runnable, err := dynamicWorkflow.Compile()
		if err != nil {
			return state, err
		}

		return runnable.Invoke(ctx, state)
	})

	workflow.SetEntryPoint("planner")
	workflow.AddEdge("planner", "executor")
	workflow.AddEdge("executor", graph.END)

	return workflow.Compile()
}

// CreatePlanningAgentForState creates a type-safe planning agent pre-configured for PlanningAgentState.
// This provides compile-time type safety without requiring manual getter/setter closures.
func CreatePlanningAgentForState(
	model llms.Model,
	availableNodes []graph.TypedNode[PlanningAgentState],
	opts ...CreateAgentOption,
) (*graph.StateRunnable[PlanningAgentState], error) {
	return CreatePlanningAgent[PlanningAgentState](
		model,
		availableNodes,
		func(s PlanningAgentState) []llms.MessageContent { return s.Messages },
		func(s PlanningAgentState, m []llms.MessageContent) PlanningAgentState { s.Messages = m; return s },
		func(s PlanningAgentState) *WorkflowPlan { return s.WorkflowPlan },
		func(s PlanningAgentState, p *WorkflowPlan) PlanningAgentState { s.WorkflowPlan = p; return s },
		opts...,
	)
}

func buildPlanningNodeDescriptions[S any](nodes []graph.TypedNode[S]) string {
	var sb strings.Builder
	sb.WriteString("Available nodes:\n")
	for i, node := range nodes {
		sb.WriteString(fmt.Sprintf("%d. %s: %s\n", i+1, node.Name, node.Description))
	}
	return sb.String()
}

func buildPlanningPrompt(nodeDescriptions string) string {
	return fmt.Sprintf(`You are a workflow planning assistant. Based on the user's request, create a workflow plan using the available nodes.

%s

Generate a workflow plan in the following JSON format:
{
  "nodes": [
    {"name": "node_name", "type": "process"}
  ],
  "edges": [
    {"from": "START", "to": "first_node"},
    {"from": "first_node", "to": "second_node"},
    {"from": "last_node", "to": "END"}
  ]
}

Rules:
1. The workflow must start with an edge from "START"
2. The workflow must end with an edge to "END"
3. Only use nodes from the available nodes list
4. Each node should appear in the nodes array
5. Create a logical flow based on the user's request
6. Return ONLY the JSON object, no additional text`, nodeDescriptions)
}

func parseWorkflowPlan(planText string) (*WorkflowPlan, error) {
	jsonText := extractJSON(planText)
	var plan WorkflowPlan
	if err := json.Unmarshal([]byte(jsonText), &plan); err != nil {
		return nil, fmt.Errorf("failed to parse JSON: %w", err)
	}
	if len(plan.Nodes) == 0 || len(plan.Edges) == 0 {
		return nil, fmt.Errorf("invalid plan")
	}
	return &plan, nil
}

func extractJSON(text string) string {
	codeBlockRegex := regexp.MustCompile("(?s)```(?:json)?\\s*({.*?})\\s*```")
	matches := codeBlockRegex.FindStringSubmatch(text)
	if len(matches) > 1 {
		return matches[1]
	}
	jsonRegex := regexp.MustCompile("(?s){.*}")
	matches = jsonRegex.FindStringSubmatch(text)
	if len(matches) > 0 {
		return matches[0]
	}
	return text
}
