package graph_test

import (
	"context"
	"errors"
	"fmt"
	"os"
	"testing"

	"github.com/smallnest/langgraphgo/graph"
	"github.com/tmc/langchaingo/llms"
	"github.com/tmc/langchaingo/llms/openai"
)

func ExampleMessageGraph() {
	// Skip if no OpenAI API key is available
	if os.Getenv("OPENAI_API_KEY") == "" {
		fmt.Println("[{human [{What is 1 + 1?}]} {ai [{1 + 1 equals 2.}]}]")
		return
	}

	model, err := openai.New()
	if err != nil {
		panic(err)
	}

	g := graph.NewMessageGraph()

	g.AddNode("oracle", func(ctx context.Context, state interface{}) (interface{}, error) {
		messages := state.([]llms.MessageContent)
		r, err := model.GenerateContent(ctx, messages, llms.WithTemperature(0.0))
		if err != nil {
			return nil, err
		}
		return append(messages,
			llms.TextParts("ai", r.Choices[0].Content),
		), nil
	})
	g.AddNode(graph.END, func(_ context.Context, state interface{}) (interface{}, error) {
		return state, nil
	})

	g.AddEdge("oracle", graph.END)
	g.SetEntryPoint("oracle")

	runnable, err := g.Compile()
	if err != nil {
		panic(err)
	}

	ctx := context.Background()
	// Let's run it!
	res, err := runnable.Invoke(ctx, []llms.MessageContent{
		llms.TextParts("human", "What is 1 + 1?"),
	})
	if err != nil {
		panic(err)
	}

	fmt.Println(res)

	// Output:
	// [{human [{What is 1 + 1?}]} {ai [{1 + 1 equals 2.}]}]
}

func TestMessageGraph(t *testing.T) {
	t.Parallel()
	testCases := []struct {
		name           string
		buildGraph     func() *graph.MessageGraph
		inputMessages  []llms.MessageContent
		expectedOutput []llms.MessageContent
		expectedError  error
	}{
		{
			name: "Simple graph",
			buildGraph: func() *graph.MessageGraph {
				g := graph.NewMessageGraph()
				g.AddNode("node1", func(_ context.Context, state interface{}) (interface{}, error) {
					messages := state.([]llms.MessageContent)
					return append(messages, llms.TextParts("ai", "Node 1")), nil
				})
				g.AddNode("node2", func(_ context.Context, state interface{}) (interface{}, error) {
					messages := state.([]llms.MessageContent)
					return append(messages, llms.TextParts("ai", "Node 2")), nil
				})
				g.AddEdge("node1", "node2")
				g.AddEdge("node2", graph.END)
				g.SetEntryPoint("node1")
				return g
			},
			inputMessages: []llms.MessageContent{llms.TextParts("human", "Input")},
			expectedOutput: []llms.MessageContent{
				llms.TextParts("human", "Input"),
				llms.TextParts("ai", "Node 1"),
				llms.TextParts("ai", "Node 2"),
			},
			expectedError: nil,
		},
		{
			name: "Entry point not set",
			buildGraph: func() *graph.MessageGraph {
				g := graph.NewMessageGraph()
				g.AddNode("node1", func(_ context.Context, state interface{}) (interface{}, error) {
					return state, nil
				})
				return g
			},
			expectedError: graph.ErrEntryPointNotSet,
		},
		{
			name: "Node not found",
			buildGraph: func() *graph.MessageGraph {
				g := graph.NewMessageGraph()
				g.AddNode("node1", func(_ context.Context, state interface{}) (interface{}, error) {
					return state, nil
				})
				g.AddEdge("node1", "node2")
				g.SetEntryPoint("node1")
				return g
			},
			expectedError: fmt.Errorf("%w: node2", graph.ErrNodeNotFound),
		},
		{
			name: "No outgoing edge",
			buildGraph: func() *graph.MessageGraph {
				g := graph.NewMessageGraph()
				g.AddNode("node1", func(_ context.Context, state interface{}) (interface{}, error) {
					return state, nil
				})
				g.SetEntryPoint("node1")
				return g
			},
			expectedError: fmt.Errorf("%w: node1", graph.ErrNoOutgoingEdge),
		},
		{
			name: "Error in node function",
			buildGraph: func() *graph.MessageGraph {
				g := graph.NewMessageGraph()
				g.AddNode("node1", func(_ context.Context, _ interface{}) (interface{}, error) {
					return nil, errors.New("node error")
				})
				g.AddEdge("node1", graph.END)
				g.SetEntryPoint("node1")
				return g
			},
			expectedError: errors.New("error in node node1: node error"),
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			g := tc.buildGraph()
			runnable, err := g.Compile()
			if err != nil {
				if tc.expectedError == nil || !errors.Is(err, tc.expectedError) {
					t.Fatalf("unexpected compile error: %v", err)
				}
				return
			}

			output, err := runnable.Invoke(context.Background(), tc.inputMessages)
			if err != nil {
				if tc.expectedError == nil || err.Error() != tc.expectedError.Error() {
					t.Fatalf("unexpected invoke error: '%v', expected '%v'", err, tc.expectedError)
				}
				return
			}

			if tc.expectedError != nil {
				t.Fatalf("expected error %v, but got nil", tc.expectedError)
			}

			outputMsgs := output.([]llms.MessageContent)
			if len(outputMsgs) != len(tc.expectedOutput) {
				t.Fatalf("expected output length %d, but got %d", len(tc.expectedOutput), len(outputMsgs))
			}

			for i, msg := range outputMsgs {
				got := fmt.Sprint(msg)
				expected := fmt.Sprint(tc.expectedOutput[i])
				if got != expected {
					t.Errorf("expected output[%d] content %q, but got %q", i, expected, got)
				}
			}
		})
	}
}
