package graph_test

import (
	"context"
	"strings"
	"testing"

	"github.com/smallnest/langgraphgo/graph"
	"github.com/tmc/langchaingo/llms"
)

//nolint:gocognit,dupl,cyclop // This is a comprehensive test that needs to check multiple scenarios with similar setup
func TestConditionalEdges(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name           string
		buildGraph     func() *graph.MessageGraph
		initialState   interface{}
		expectedResult interface{}
		expectError    bool
	}{
		{
			name: "Simple conditional routing based on content",
			buildGraph: func() *graph.MessageGraph {
				g := graph.NewMessageGraph()

				// Add nodes
				g.AddNode("start", func(ctx context.Context, state interface{}) (interface{}, error) {
					// Just pass through
					return state, nil
				})

				g.AddNode("calculator", func(ctx context.Context, state interface{}) (interface{}, error) {
					messages := state.([]llms.MessageContent)
					return append(messages, llms.TextParts("ai", "Calculating: 2+2=4")), nil
				})

				g.AddNode("general", func(ctx context.Context, state interface{}) (interface{}, error) {
					messages := state.([]llms.MessageContent)
					return append(messages, llms.TextParts("ai", "General response")), nil
				})

				// Add conditional edge from start
				g.AddConditionalEdge("start", func(ctx context.Context, state interface{}) string {
					messages := state.([]llms.MessageContent)
					if len(messages) > 0 {
						lastMessage := messages[len(messages)-1]
						if content, ok := lastMessage.Parts[0].(llms.TextContent); ok {
							if strings.Contains(content.Text, "calculate") || strings.Contains(content.Text, "math") {
								return "calculator"
							}
						}
					}
					return "general"
				})

				// Add regular edges to END
				g.AddEdge("calculator", graph.END)
				g.AddEdge("general", graph.END)

				g.SetEntryPoint("start")
				return g
			},
			initialState: []llms.MessageContent{
				llms.TextParts("human", "I need to calculate something"),
			},
			expectedResult: []llms.MessageContent{
				llms.TextParts("human", "I need to calculate something"),
				llms.TextParts("ai", "Calculating: 2+2=4"),
			},
			expectError: false,
		},
		{
			name: "Conditional routing to general path",
			buildGraph: func() *graph.MessageGraph {
				g := graph.NewMessageGraph()

				g.AddNode("start", func(ctx context.Context, state interface{}) (interface{}, error) {
					return state, nil
				})

				g.AddNode("calculator", func(ctx context.Context, state interface{}) (interface{}, error) {
					messages := state.([]llms.MessageContent)
					return append(messages, llms.TextParts("ai", "Calculating: 2+2=4")), nil
				})

				g.AddNode("general", func(ctx context.Context, state interface{}) (interface{}, error) {
					messages := state.([]llms.MessageContent)
					return append(messages, llms.TextParts("ai", "General response")), nil
				})

				g.AddConditionalEdge("start", func(ctx context.Context, state interface{}) string {
					messages := state.([]llms.MessageContent)
					if len(messages) > 0 {
						lastMessage := messages[len(messages)-1]
						if content, ok := lastMessage.Parts[0].(llms.TextContent); ok {
							if strings.Contains(content.Text, "calculate") || strings.Contains(content.Text, "math") {
								return "calculator"
							}
						}
					}
					return "general"
				})

				g.AddEdge("calculator", graph.END)
				g.AddEdge("general", graph.END)

				g.SetEntryPoint("start")
				return g
			},
			initialState: []llms.MessageContent{
				llms.TextParts("human", "Tell me a story"),
			},
			expectedResult: []llms.MessageContent{
				llms.TextParts("human", "Tell me a story"),
				llms.TextParts("ai", "General response"),
			},
			expectError: false,
		},
		{
			name: "Multi-level conditional routing",
			buildGraph: func() *graph.MessageGraph {
				g := graph.NewMessageGraph()

				g.AddNode("router", func(ctx context.Context, state interface{}) (interface{}, error) {
					return state, nil
				})

				g.AddNode("urgent", func(ctx context.Context, state interface{}) (interface{}, error) {
					s := state.(string)
					return s + " -> handled urgently", nil
				})

				g.AddNode("normal", func(ctx context.Context, state interface{}) (interface{}, error) {
					s := state.(string)
					return s + " -> handled normally", nil
				})

				g.AddNode("low", func(ctx context.Context, state interface{}) (interface{}, error) {
					s := state.(string)
					return s + " -> handled with low priority", nil
				})

				// Conditional routing based on priority keywords
				g.AddConditionalEdge("router", func(ctx context.Context, state interface{}) string {
					s := state.(string)
					if strings.Contains(s, "URGENT") || strings.Contains(s, "ASAP") {
						return "urgent"
					}
					if strings.Contains(s, "NORMAL") || strings.Contains(s, "REGULAR") {
						return "normal"
					}
					return "low"
				})

				g.AddEdge("urgent", graph.END)
				g.AddEdge("normal", graph.END)
				g.AddEdge("low", graph.END)

				g.SetEntryPoint("router")
				return g
			},
			initialState:   "URGENT: Fix the bug",
			expectedResult: "URGENT: Fix the bug -> handled urgently",
			expectError:    false,
		},
		{
			name: "Conditional edge to END",
			buildGraph: func() *graph.MessageGraph {
				g := graph.NewMessageGraph()

				g.AddNode("check", func(ctx context.Context, state interface{}) (interface{}, error) {
					return state, nil
				})

				g.AddNode("process", func(ctx context.Context, state interface{}) (interface{}, error) {
					n := state.(int)
					return n * 2, nil
				})

				// Conditional edge that can go directly to END
				g.AddConditionalEdge("check", func(ctx context.Context, state interface{}) string {
					n := state.(int)
					if n < 0 {
						return graph.END
					}
					return "process"
				})

				g.AddEdge("process", graph.END)

				g.SetEntryPoint("check")
				return g
			},
			initialState:   -5,
			expectedResult: -5, // Should go directly to END without processing
			expectError:    false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			g := tt.buildGraph()
			runnable, err := g.Compile()
			if err != nil {
				t.Fatalf("Failed to compile graph: %v", err)
			}

			ctx := context.Background()
			result, err := runnable.Invoke(ctx, tt.initialState)

			if tt.expectError && err == nil {
				t.Error("Expected error but got none")
			}
			if !tt.expectError && err != nil {
				t.Errorf("Unexpected error: %v", err)
			}

			if !tt.expectError {
				// For message content, compare the last messages
				if messages, ok := result.([]llms.MessageContent); ok {
					expectedMessages := tt.expectedResult.([]llms.MessageContent)
					if len(messages) != len(expectedMessages) {
						t.Errorf("Expected %d messages, got %d", len(expectedMessages), len(messages))
					} else {
						for i := range messages {
							if messages[i].Role != expectedMessages[i].Role {
								t.Errorf("Message %d: expected role %s, got %s", i, expectedMessages[i].Role, messages[i].Role)
							}
							expectedText := expectedMessages[i].Parts[0].(llms.TextContent).Text
							actualText := messages[i].Parts[0].(llms.TextContent).Text
							if actualText != expectedText {
								t.Errorf("Message %d: expected text %q, got %q", i, expectedText, actualText)
							}
						}
					}
				} else {
					// For other types, direct comparison
					if result != tt.expectedResult {
						t.Errorf("Expected result %v, got %v", tt.expectedResult, result)
					}
				}
			}
		})
	}
}

func TestConditionalEdges_ChainedConditions(t *testing.T) {
	t.Parallel()

	g := graph.NewMessageGraph()

	// Create a chain of conditional decisions
	g.AddNode("start", func(ctx context.Context, state interface{}) (interface{}, error) {
		return state, nil
	})

	g.AddNode("step1", func(ctx context.Context, state interface{}) (interface{}, error) {
		n := state.(int)
		return n + 10, nil
	})

	g.AddNode("step2", func(ctx context.Context, state interface{}) (interface{}, error) {
		n := state.(int)
		return n * 2, nil
	})

	g.AddNode("step3", func(ctx context.Context, state interface{}) (interface{}, error) {
		n := state.(int)
		return n - 5, nil
	})

	// First conditional
	g.AddConditionalEdge("start", func(ctx context.Context, state interface{}) string {
		n := state.(int)
		if n > 0 {
			return "step1"
		}
		return "step2"
	})

	// Second conditional
	g.AddConditionalEdge("step1", func(ctx context.Context, state interface{}) string {
		n := state.(int)
		if n > 15 {
			return "step3"
		}
		return graph.END
	})

	g.AddEdge("step2", graph.END)
	g.AddEdge("step3", graph.END)
	g.SetEntryPoint("start")

	runnable, err := g.Compile()
	if err != nil {
		t.Fatalf("Failed to compile graph: %v", err)
	}

	// Test with positive number (should go: start -> step1 -> step3 -> END)
	ctx := context.Background()
	result, err := runnable.Invoke(ctx, 10)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	// 10 + 10 = 20 (step1), then 20 > 15 so go to step3, 20 - 5 = 15
	if result != 15 {
		t.Errorf("Expected result 15, got %v", result)
	}

	// Test with negative number (should go: start -> step2 -> END)
	result, err = runnable.Invoke(ctx, -5)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	// -5 * 2 = -10 (step2)
	if result != -10 {
		t.Errorf("Expected result -10, got %v", result)
	}
}
