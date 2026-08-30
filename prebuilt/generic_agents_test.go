package prebuilt

import (
	"context"
	"testing"

	"github.com/smallnest/langgraphgo/graph"
	"github.com/smallnest/langgraphgo/llms"
	"github.com/smallnest/langgraphgo/tool"
)

// Mock Reflection LLM
type MockReflectionLLM struct {
	responses []string
	callCount int
}

func (m *MockReflectionLLM) GenerateContent(ctx context.Context, messages []llms.MessageContent, options ...llms.CallOption) (*llms.ContentResponse, error) {
	response := m.responses[m.callCount%len(m.responses)]
	m.callCount++
	return &llms.ContentResponse{Choices: []*llms.ContentChoice{{Content: response}}}, nil
}

func (m *MockReflectionLLM) Call(ctx context.Context, prompt string, options ...llms.CallOption) (string, error) {
	return "", nil
}

// PEV Mock Tool
type PEVMockTool struct {
	name        string
	description string
	response    string
}

func (m PEVMockTool) Name() string        { return m.name }
func (m PEVMockTool) Description() string { return m.description }
func (m PEVMockTool) Call(ctx context.Context, input string) (string, error) {
	return m.response, nil
}

// PEV Mock LLM
type PEVMockLLM struct {
	responses []string
	callCount int
}

func (m *PEVMockLLM) GenerateContent(ctx context.Context, messages []llms.MessageContent, options ...llms.CallOption) (*llms.ContentResponse, error) {
	if m.callCount >= len(m.responses) {
		return &llms.ContentResponse{Choices: []*llms.ContentChoice{{Content: "Default response"}}}, nil
	}
	response := m.responses[m.callCount]
	m.callCount++
	return &llms.ContentResponse{Choices: []*llms.ContentChoice{{Content: response}}}, nil
}

func (m *PEVMockLLM) Call(ctx context.Context, prompt string, options ...llms.CallOption) (string, error) {
	return "", nil
}

// Simple Thought State
type SimpleThoughtState struct {
	isGoal  bool
	isValid bool
	desc    string
}

func (s *SimpleThoughtState) IsValid() bool          { return s.isValid }
func (s *SimpleThoughtState) IsGoal() bool           { return s.isGoal }
func (s *SimpleThoughtState) GetDescription() string { return s.desc }
func (s *SimpleThoughtState) Hash() string           { return s.desc }

type SimpleThoughtGenerator struct{}

func (g *SimpleThoughtGenerator) Generate(ctx context.Context, state ThoughtState) ([]ThoughtState, error) {
	return []ThoughtState{}, nil
}

type SimpleThoughtEvaluator struct{}

func (e *SimpleThoughtEvaluator) Evaluate(ctx context.Context, state ThoughtState, depth int) (float64, error) {
	return 1.0, nil
}

// TestCreateReflectionAgent tests the ReflectionAgent
func TestCreateReflectionAgent(t *testing.T) {
	mockLLM := &MockReflectionLLM{
		responses: []string{
			"Initial response",
			"**Strengths:** Good. **Weaknesses:** None. **Suggestions:** None.",
		},
	}
	config := ReflectionAgentConfig{Model: mockLLM, MaxIterations: 2}
	agent, err := CreateReflectionAgent(
		config,
		func(s ReflectionAgentState) []llms.MessageContent { return s.Messages },
		func(s ReflectionAgentState, m []llms.MessageContent) ReflectionAgentState {
			s.Messages = m
			return s
		},
		func(s ReflectionAgentState) string { return s.Draft },
		func(s ReflectionAgentState, d string) ReflectionAgentState { s.Draft = d; return s },
		func(s ReflectionAgentState) int { return s.Iteration },
		func(s ReflectionAgentState, i int) ReflectionAgentState { s.Iteration = i; return s },
		func(s ReflectionAgentState) string { return s.Reflection },
		func(s ReflectionAgentState, r string) ReflectionAgentState { s.Reflection = r; return s },
	)
	if err != nil {
		t.Fatalf("Failed: %v", err)
	}
	_, err = agent.Invoke(context.Background(), ReflectionAgentState{Messages: []llms.MessageContent{llms.TextParts(llms.ChatMessageTypeHuman, "Test")}})
	if err != nil {
		t.Fatalf("Invoke failed: %v", err)
	}
}

// TestCreatePEVAgent tests the PEVAgent
func TestCreatePEVAgent(t *testing.T) {
	mockLLM := &PEVMockLLM{
		responses: []string{
			"1. Step",
			`{"tool": "calculator", "tool_input": "2+2"}`,
			`{"is_successful": true, "reasoning": "Ok"}`,
			"Final",
		},
	}
	config := PEVAgentConfig{Model: mockLLM, Tools: []tool.Tool{PEVMockTool{name: "calculator"}}}
	agent, err := CreatePEVAgent(
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
		func(s PEVAgentState, r string) PEVAgentState { s.VerificationResult = r; return s },
		func(s PEVAgentState) string { return s.FinalAnswer },
		func(s PEVAgentState, a string) PEVAgentState { s.FinalAnswer = a; return s },
	)
	if err != nil {
		t.Fatalf("Failed: %v", err)
	}
	_, err = agent.Invoke(context.Background(), PEVAgentState{Messages: []llms.MessageContent{llms.TextParts(llms.ChatMessageTypeHuman, "Test")}})
	if err != nil {
		t.Fatalf("Invoke failed: %v", err)
	}
}

func TestTreeOfThoughtsAgent(t *testing.T) {
	config := TreeOfThoughtsConfig{
		Generator:    &SimpleThoughtGenerator{},
		Evaluator:    &SimpleThoughtEvaluator{},
		InitialState: &SimpleThoughtState{isGoal: true, isValid: true, desc: "Goal"},
	}
	agent, err := CreateTreeOfThoughtsAgent(
		config,
		func(s TreeOfThoughtsState) map[string]*SearchPath { return s.ActivePaths },
		func(s TreeOfThoughtsState, p map[string]*SearchPath) TreeOfThoughtsState { s.ActivePaths = p; return s },
		func(s TreeOfThoughtsState) string { return s.Solution },
		func(s TreeOfThoughtsState, sol string) TreeOfThoughtsState { s.Solution = sol; return s },
		func(s TreeOfThoughtsState) map[string]bool { return s.VisitedStates },
		func(s TreeOfThoughtsState, v map[string]bool) TreeOfThoughtsState { s.VisitedStates = v; return s },
		func(s TreeOfThoughtsState) int { return s.Iteration },
		func(s TreeOfThoughtsState, i int) TreeOfThoughtsState { s.Iteration = i; return s },
	)
	if err != nil {
		t.Fatalf("Failed: %v", err)
	}
	_, err = agent.Invoke(context.Background(), TreeOfThoughtsState{})
	if err != nil {
		t.Fatalf("Invoke failed: %v", err)
	}
}

func TestCreateForStateHelpers(t *testing.T) {
	mockLLM := &MockReflectionLLM{
		responses: []string{
			"Initial response",
			"**Strengths:** Good. **Weaknesses:** None. **Suggestions:** None.",
		},
	}

	// Test CreateReflectionAgentForState
	reflConfig := ReflectionAgentConfig{Model: mockLLM, MaxIterations: 2}
	reflAgent, err := CreateReflectionAgentForState(reflConfig)
	if err != nil {
		t.Fatalf("CreateReflectionAgentForState failed: %v", err)
	}
	_, err = reflAgent.Invoke(context.Background(), ReflectionAgentState{
		Messages: []llms.MessageContent{llms.TextParts(llms.ChatMessageTypeHuman, "Test")},
	})
	if err != nil {
		t.Fatalf("reflAgent invoke failed: %v", err)
	}

	// Test CreatePEVAgentForState
	pevMockLLM := &PEVMockLLM{
		responses: []string{
			"1. Step",
			`{"tool": "calculator", "tool_input": "2+2"}`,
			`{"is_successful": true, "reasoning": "Ok"}`,
			"Final",
		},
	}
	pevConfig := PEVAgentConfig{Model: pevMockLLM, Tools: []tool.Tool{PEVMockTool{name: "calculator"}}}
	pevAgent, err := CreatePEVAgentForState(pevConfig)
	if err != nil {
		t.Fatalf("CreatePEVAgentForState failed: %v", err)
	}
	_, err = pevAgent.Invoke(context.Background(), PEVAgentState{
		Messages: []llms.MessageContent{llms.TextParts(llms.ChatMessageTypeHuman, "Test")},
	})
	if err != nil {
		t.Fatalf("pevAgent invoke failed: %v", err)
	}

	// Test CreateTreeOfThoughtsAgentForState
	totConfig := TreeOfThoughtsConfig{
		Generator:    &SimpleThoughtGenerator{},
		Evaluator:    &SimpleThoughtEvaluator{},
		InitialState: &SimpleThoughtState{isGoal: true, isValid: true, desc: "Goal"},
	}
	totAgent, err := CreateTreeOfThoughtsAgentForState(totConfig)
	if err != nil {
		t.Fatalf("CreateTreeOfThoughtsAgentForState failed: %v", err)
	}
	_, err = totAgent.Invoke(context.Background(), TreeOfThoughtsState{})
	if err != nil {
		t.Fatalf("totAgent invoke failed: %v", err)
	}

	// Test CreateAgentForState
	agentForState, err := CreateAgentForState(mockLLM, []tool.Tool{PEVMockTool{name: "calc"}})
	if err != nil {
		t.Fatalf("CreateAgentForState failed: %v", err)
	}
	_, err = agentForState.Invoke(context.Background(), AgentState{
		Messages: []llms.MessageContent{llms.TextParts(llms.ChatMessageTypeHuman, "Test")},
	})
	if err != nil {
		t.Fatalf("agentForState invoke failed: %v", err)
	}

	// Test CreateReactAgentForState
	reactForState, err := CreateReactAgentForState(mockLLM, []tool.Tool{PEVMockTool{name: "calc"}}, 5)
	if err != nil {
		t.Fatalf("CreateReactAgentForState failed: %v", err)
	}
	_, err = reactForState.Invoke(context.Background(), ReactAgentState{
		Messages: []llms.MessageContent{llms.TextParts(llms.ChatMessageTypeHuman, "Test")},
	})
	if err != nil {
		t.Fatalf("reactForState invoke failed: %v", err)
	}

	// Test CreatePlanningAgentForState
	planLLM := &MockPlanningLLM{
		planJSON: `{"nodes": [{"name": "step1", "type": "process"}], "edges": [{"from": "START", "to": "step1"}, {"from": "step1", "to": "END"}]}`,
	}
	planNodes := []graph.TypedNode[PlanningAgentState]{
		{
			Name:        "step1",
			Description: "Step 1",
			Function: func(ctx context.Context, state PlanningAgentState) (PlanningAgentState, error) {
				return state, nil
			},
		},
	}
	planningAgent, err := CreatePlanningAgentForState(planLLM, planNodes)
	if err != nil {
		t.Fatalf("CreatePlanningAgentForState failed: %v", err)
	}
	_, err = planningAgent.Invoke(context.Background(), PlanningAgentState{
		Messages: []llms.MessageContent{llms.TextParts(llms.ChatMessageTypeHuman, "Plan something")},
	})
	if err != nil {
		t.Fatalf("planningAgent invoke failed: %v", err)
	}

	// Test CreateSupervisorForState
	supLLM := &SupervisorMockLLM{
		responses: []llms.ContentResponse{
			{
				Choices: []*llms.ContentChoice{
					{
						ToolCalls: []llms.ToolCall{
							{
								FunctionCall: &llms.FunctionCall{
									Name:      "route",
									Arguments: `{"next": "FINISH"}`,
								},
							},
						},
					},
				},
			},
		},
	}
	supAgent, err := CreateSupervisorForState(supLLM, map[string]*graph.StateRunnable[SupervisorState]{})
	if err != nil {
		t.Fatalf("CreateSupervisorForState failed: %v", err)
	}
	_, err = supAgent.Invoke(context.Background(), SupervisorState{
		Messages: []llms.MessageContent{llms.TextParts(llms.ChatMessageTypeHuman, "Coordinate this")},
	})
	if err != nil {
		t.Fatalf("supAgent invoke failed: %v", err)
	}
}


