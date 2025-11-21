package graph

import (
	"context"

	"testing"
	"github.com/stretchr/testify/assert"
)

// Define a sample state struct for testing visualization.
type VisTestState struct {
	Count int
}

// nodeInc is a sample node.
func nodeInc(ctx context.Context, state VisTestState) (map[string]interface{}, error) {
	return map[string]interface{}{"Count": state.Count + 1}, nil
}

// nodeDec is a sample node.
func nodeDec(ctx context.Context, state VisTestState) (map[string]interface{}, error) {
	return map[string]interface{}{"Count": state.Count - 1}, nil
}

func TestStatefulGraphExporter_DrawMermaid(t *testing.T) {
	g, err := NewStatefulGraph(VisTestState{})
	assert.NoError(t, err)

	_ = g.AddNode("startNode", nodeInc)
	_ = g.AddNode("middleNode", nodeDec)
	_ = g.AddNode("endNode", nodeInc)

	_ = g.SetEntryPoint("startNode")
	_ = g.SetFinishPoint("endNode")

	_ = g.AddEdge("startNode", "middleNode")
	_ = g.AddEdge("middleNode", "endNode")

	exporter := g.GetGraph()
	mermaid := exporter.DrawMermaid()

	t.Logf("Generated Mermaid:\n%s", mermaid)

	assert.Contains(t, mermaid, "flowchart TD")
	assert.Contains(t, mermaid, "START[\"START\"]:::startNode")
	assert.Contains(t, mermaid, "START --> startNode")
	assert.Contains(t, mermaid, "startNode[[\"startNode\"]]") // Entry point styling
	assert.Contains(t, mermaid, "middleNode[\"middleNode\"]")
	assert.Contains(t, mermaid, "endNode(((\"endNode\")))") // Finish point styling
	assert.Contains(t, mermaid, "endNode --> END[\"END\"]:::endNode")

	assert.Contains(t, mermaid, "startNode --> middleNode")
	assert.Contains(t, mermaid, "middleNode --> endNode")

	assert.Contains(t, mermaid, "classDef startNode")
	assert.Contains(t, mermaid, "classDef endNode")
	assert.Contains(t, mermaid, "classDef entryPointNode")
	assert.Contains(t, mermaid, "classDef finishPointNode")
	assert.Contains(t, mermaid, "class startNode entryPointNode;")
	assert.Contains(t, mermaid, "class endNode finishPointNode;")
}

func TestStatefulGraphExporter_DrawDOT(t *testing.T) {
	g, err := NewStatefulGraph(VisTestState{})
	assert.NoError(t, err)

	_ = g.AddNode("startNode", nodeInc)
	_ = g.AddNode("middleNode", nodeDec)
	_ = g.AddNode("endNode", nodeInc)

	_ = g.SetEntryPoint("startNode")
	_ = g.SetFinishPoint("endNode")

	_ = g.AddEdge("startNode", "middleNode")
	_ = g.AddEdge("middleNode", "endNode")

	exporter := g.GetGraph()
	dot := exporter.DrawDOT()

	t.Logf("Generated DOT:\n%s", dot)

	assert.Contains(t, dot, "digraph G {")
	assert.Contains(t, dot, "rankdir=TD;")
	assert.Contains(t, dot, "node [shape=box];")

	assert.Contains(t, dot, "START [label=\"START\", shape=ellipse, style=filled, fillcolor=\"#90EE90\"];")
	assert.Contains(t, dot, "START -> startNode;")

	assert.Contains(t, dot, "END [label=\"END\", shape=ellipse, style=filled, fillcolor=\"#FFB6C1\"];")
	assert.Contains(t, dot, "endNode -> END;")

	assert.Contains(t, dot, "startNode [label=\"startNode\", style=filled, fillcolor=\"#87CEEB\", peripheries=2];") // Entry styling
	assert.Contains(t, dot, "middleNode [label=\"middleNode\"];")
	assert.Contains(t, dot, "endNode [label=\"endNode\", shape=doubleoctagon, style=filled, fillcolor=\"#ADD8E6\"];") // Finish styling

	assert.Contains(t, dot, "startNode -> middleNode;")
	assert.Contains(t, dot, "middleNode -> endNode;")
}

// Test for graph with no entry/finish points
func TestStatefulGraphExporter_NoEntryOrFinish(t *testing.T) {
	g, err := NewStatefulGraph(VisTestState{})
	assert.NoError(t, err)

	_ = g.AddNode("node1", nodeInc)
	_ = g.AddNode("node2", nodeDec)
	_ = g.AddEdge("node1", "node2")

	exporter := g.GetGraph()
	mermaid := exporter.DrawMermaid()
	dot := exporter.DrawDOT()

	t.Logf("Generated Mermaid (no entry/finish):\n%s", mermaid)
	t.Logf("Generated DOT (no entry/finish):\n%s", dot)

	assert.NotContains(t, mermaid, "START")
	assert.NotContains(t, mermaid, "END")
	assert.Contains(t, mermaid, "node1[\"node1\"]")
	assert.Contains(t, mermaid, "node2[\"node2\"]")

	assert.NotContains(t, dot, "START")
	assert.NotContains(t, dot, "END")
	assert.Contains(t, dot, "node1 [label=\"node1\"];")
	assert.Contains(t, dot, "node2 [label=\"node2\"];")
}
