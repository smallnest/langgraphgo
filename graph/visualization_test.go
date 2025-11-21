package graph_test

import (
	"context"
	"fmt"
	"strings"
	"testing"

	"github.com/smallnest/langgraphgo/graph"
)

func TestExporter_DrawMermaid(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name          string
		buildGraph    func() *graph.MessageGraph
		expectedLines []string
	}{
		{
			name: "Simple linear graph",
			buildGraph: func() *graph.MessageGraph {
				g := graph.NewMessageGraph()
				g.AddNode("node1", func(_ context.Context, state interface{}) (interface{}, error) {
					return state, nil
				})
				g.AddNode("node2", func(_ context.Context, state interface{}) (interface{}, error) {
					return state, nil
				})
				g.AddEdge("node1", "node2")
				g.AddEdge("node2", graph.END)
				g.SetEntryPoint("node1")
				return g
			},
			expectedLines: []string{
				"flowchart TD",
				`node1[["node1"]]`,
				"START --> node1",
				"START([\"START\"])",
				"style START fill:#90EE90",
				`node2["node2"]`,
				"END([\"END\"])",
				"style END fill:#FFB6C1",
				"node1 --> node2",
				"node2 --> END",
				"style node1 fill:#87CEEB",
			},
		},
		{
			name: "Graph with branching",
			buildGraph: func() *graph.MessageGraph {
				g := graph.NewMessageGraph()
				g.AddNode("start", func(_ context.Context, state interface{}) (interface{}, error) {
					return state, nil
				})
				g.AddNode("branch1", func(_ context.Context, state interface{}) (interface{}, error) {
					return state, nil
				})
				g.AddNode("branch2", func(_ context.Context, state interface{}) (interface{}, error) {
					return state, nil
				})
				g.AddEdge("start", "branch1")
				g.AddEdge("start", "branch2")
				g.AddEdge("branch1", graph.END)
				g.AddEdge("branch2", graph.END)
				g.SetEntryPoint("start")
				return g
			},
			expectedLines: []string{
				"flowchart TD",
				`start[["start"]]`,
				"START --> start",
				`branch1["branch1"]`,
				`branch2["branch2"]`,
				"END([\"END\"])",
				"start --> branch1",
				"start --> branch2",
				"branch1 --> END",
				"branch2 --> END",
				"style start fill:#87CEEB",
			},
		},
		{
			name: "Empty graph with no entry point",
			buildGraph: func() *graph.MessageGraph {
				g := graph.NewMessageGraph()
				g.AddNode("orphan", func(_ context.Context, state interface{}) (interface{}, error) {
					return state, nil
				})
				return g
			},
			expectedLines: []string{
				"flowchart TD",
				`orphan["orphan"]`,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			g := tt.buildGraph()
			exporter := graph.NewExporter(g)

			result := exporter.DrawMermaid()

			// Check that all expected lines are present
			for _, expectedLine := range tt.expectedLines {
				if !strings.Contains(result, expectedLine) {
					t.Errorf("Expected line not found in Mermaid output: %q\nActual output:\n%s", expectedLine, result)
				}
			}

			// Verify it starts with flowchart TD
			if !strings.HasPrefix(result, "flowchart TD\n") {
				t.Errorf("Mermaid output should start with 'flowchart TD\\n', got: %s", result)
			}
		})
	}
}

func TestExporter_DrawDOT(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name          string
		buildGraph    func() *graph.MessageGraph
		expectedLines []string
	}{
		{
			name: "Simple linear graph",
			buildGraph: func() *graph.MessageGraph {
				g := graph.NewMessageGraph()
				g.AddNode("node1", func(_ context.Context, state interface{}) (interface{}, error) {
					return state, nil
				})
				g.AddNode("node2", func(_ context.Context, state interface{}) (interface{}, error) {
					return state, nil
				})
				g.AddEdge("node1", "node2")
				g.AddEdge("node2", graph.END)
				g.SetEntryPoint("node1")
				return g
			},
			expectedLines: []string{
				"digraph G {",
				"rankdir=TD;",
				"node [shape=box];",
				`START [label="START", shape=ellipse, style=filled, fillcolor=lightgreen];`,
				"START -> node1;",
				"node1 [style=filled, fillcolor=lightblue];",
				`END [label="END", shape=ellipse, style=filled, fillcolor=lightpink];`,
				"node1 -> node2;",
				"node2 -> END;",
				"}",
			},
		},
		{
			name: "Graph without entry point",
			buildGraph: func() *graph.MessageGraph {
				g := graph.NewMessageGraph()
				g.AddNode("node1", func(_ context.Context, state interface{}) (interface{}, error) {
					return state, nil
				})
				g.AddNode("node2", func(_ context.Context, state interface{}) (interface{}, error) {
					return state, nil
				})
				g.AddEdge("node1", "node2")
				return g
			},
			expectedLines: []string{
				"digraph G {",
				"rankdir=TD;",
				"node [shape=box];",
				"node1 -> node2;",
				"}",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			g := tt.buildGraph()
			exporter := graph.NewExporter(g)

			result := exporter.DrawDOT()

			// Check that all expected lines are present
			for _, expectedLine := range tt.expectedLines {
				if !strings.Contains(result, expectedLine) {
					t.Errorf("Expected line not found in DOT output: %q\nActual output:\n%s", expectedLine, result)
				}
			}

			// Verify it starts and ends correctly
			if !strings.HasPrefix(result, "digraph G {") {
				t.Errorf("DOT output should start with 'digraph G {', got: %s", result)
			}

			if !strings.HasSuffix(strings.TrimSpace(result), "}") {
				t.Errorf("DOT output should end with '}', got: %s", result)
			}
		})
	}
}

func TestExporter_DrawASCII(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name          string
		buildGraph    func() *graph.MessageGraph
		expectedLines []string
	}{
		{
			name: "Simple linear graph",
			buildGraph: func() *graph.MessageGraph {
				g := graph.NewMessageGraph()
				g.AddNode("node1", func(_ context.Context, state interface{}) (interface{}, error) {
					return state, nil
				})
				g.AddNode("node2", func(_ context.Context, state interface{}) (interface{}, error) {
					return state, nil
				})
				g.AddEdge("node1", "node2")
				g.AddEdge("node2", graph.END)
				g.SetEntryPoint("node1")
				return g
			},
			expectedLines: []string{
				"Graph Execution Flow:",
				"├── START",
				"│   └── node1",
				"│       └── node2",
				"│           └── END",
			},
		},
		{
			name: "Graph with branching",
			buildGraph: func() *graph.MessageGraph {
				g := graph.NewMessageGraph()
				g.AddNode("start", func(_ context.Context, state interface{}) (interface{}, error) {
					return state, nil
				})
				g.AddNode("branch1", func(_ context.Context, state interface{}) (interface{}, error) {
					return state, nil
				})
				g.AddNode("branch2", func(_ context.Context, state interface{}) (interface{}, error) {
					return state, nil
				})
				g.AddEdge("start", "branch1")
				g.AddEdge("start", "branch2")
				g.AddEdge("branch1", graph.END)
				g.AddEdge("branch2", graph.END)
				g.SetEntryPoint("start")
				return g
			},
			expectedLines: []string{
				"Graph Execution Flow:",
				"├── START",
				"│   └── start",
				"│       ├── branch1",
				"│       │   └── END",
				"│       └── branch2",
				"│           └── END",
			},
		},
		{
			name: "Graph with no entry point",
			buildGraph: func() *graph.MessageGraph {
				g := graph.NewMessageGraph()
				g.AddNode("orphan", func(_ context.Context, state interface{}) (interface{}, error) {
					return state, nil
				})
				return g
			},
			expectedLines: []string{
				"No entry point set",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			g := tt.buildGraph()
			exporter := graph.NewExporter(g)

			result := exporter.DrawASCII()

			// Check that all expected lines are present
			for _, expectedLine := range tt.expectedLines {
				if !strings.Contains(result, expectedLine) {
					t.Errorf("Expected line not found in ASCII output: %q\nActual output:\n%s", expectedLine, result)
				}
			}
		})
	}
}

func TestRunnable_GetGraph(t *testing.T) {
	t.Parallel()

	g := graph.NewMessageGraph()
	g.AddNode("test", func(_ context.Context, state interface{}) (interface{}, error) {
		return state, nil
	})
	g.AddEdge("test", graph.END)
	g.SetEntryPoint("test")

	runnable, err := g.Compile()
	if err != nil {
		t.Fatalf("Failed to compile graph: %v", err)
	}

	exporter := runnable.GetGraph()
	if exporter == nil {
		t.Fatal("GetGraph() returned nil")
	}

	// Test that we can use the exporter
	mermaid := exporter.DrawMermaid()
	if !strings.Contains(mermaid, "test") {
		t.Errorf("Mermaid output should contain test node, got: %s", mermaid)
	}
}

func TestExporter_CycleDetection(t *testing.T) {
	t.Parallel()

	g := graph.NewMessageGraph()
	g.AddNode("node1", func(_ context.Context, state interface{}) (interface{}, error) {
		return state, nil
	})
	g.AddNode("node2", func(_ context.Context, state interface{}) (interface{}, error) {
		return state, nil
	})

	// Create a cycle
	g.AddEdge("node1", "node2")
	g.AddEdge("node2", "node1")
	g.SetEntryPoint("node1")

	exporter := graph.NewExporter(g)
	ascii := exporter.DrawASCII()

	if !strings.Contains(ascii, "(cycle)") {
		t.Errorf("ASCII output should detect cycle, got: %s", ascii)
	}
}

// Benchmark tests
func BenchmarkExporter_DrawMermaid(b *testing.B) {
	g := graph.NewMessageGraph()
	for i := 0; i < 100; i++ {
		nodeName := fmt.Sprintf("node%d", i)
		g.AddNode(nodeName, func(_ context.Context, state interface{}) (interface{}, error) {
			return state, nil
		})
		if i > 0 {
			prevNode := fmt.Sprintf("node%d", i-1)
			g.AddEdge(prevNode, nodeName)
		}
	}
	g.AddEdge("node99", graph.END)
	g.SetEntryPoint("node0")

	exporter := graph.NewExporter(g)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		exporter.DrawMermaid()
	}
}

func BenchmarkExporter_DrawDOT(b *testing.B) {
	g := graph.NewMessageGraph()
	for i := 0; i < 100; i++ {
		nodeName := fmt.Sprintf("node%d", i)
		g.AddNode(nodeName, func(_ context.Context, state interface{}) (interface{}, error) {
			return state, nil
		})
		if i > 0 {
			prevNode := fmt.Sprintf("node%d", i-1)
			g.AddEdge(prevNode, nodeName)
		}
	}
	g.AddEdge("node99", graph.END)
	g.SetEntryPoint("node0")

	exporter := graph.NewExporter(g)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		exporter.DrawDOT()
	}
}

func BenchmarkExporter_DrawASCII(b *testing.B) {
	g := graph.NewMessageGraph()
	for i := 0; i < 10; i++ { // Smaller graph for ASCII due to recursive nature
		nodeName := fmt.Sprintf("node%d", i)
		g.AddNode(nodeName, func(_ context.Context, state interface{}) (interface{}, error) {
			return state, nil
		})
		if i > 0 {
			prevNode := fmt.Sprintf("node%d", i-1)
			g.AddEdge(prevNode, nodeName)
		}
	}
	g.AddEdge("node9", graph.END)
	g.SetEntryPoint("node0")

	exporter := graph.NewExporter(g)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		exporter.DrawASCII()
	}
}
