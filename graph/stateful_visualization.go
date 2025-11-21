package graph

import (
	"fmt"
	"sort"
	"strings"
)

// StatefulGraphExporter provides methods to export StatefulGraph in different formats
type StatefulGraphExporter struct {
	graph *StatefulGraph
}

// NewStatefulGraphExporter creates a new graph exporter for the given StatefulGraph
func NewStatefulGraphExporter(graph *StatefulGraph) *StatefulGraphExporter {
	return &StatefulGraphExporter{graph: graph}
}

// DrawMermaid generates a Mermaid diagram representation of the StatefulGraph
func (ge *StatefulGraphExporter) DrawMermaid() string {
	var sb strings.Builder

	// Start Mermaid flowchart
	sb.WriteString("flowchart TD\n")

	// Collect all unique node names
	nodeNamesMap := make(map[string]struct{})
	for name := range ge.graph.nodes {
		nodeNamesMap[name] = struct{}{}
	}

	// Add entry point
	if ge.graph.entryPoint != "" {
		sb.WriteString(fmt.Sprintf("    START[\"START\"]:::startNode\n"))
		sb.WriteString(fmt.Sprintf("    START --> %s\n", ge.graph.entryPoint))
		nodeNamesMap[ge.graph.entryPoint] = struct{}{} // Ensure entry point is in nodes
	}

	// Add finish point
	if ge.graph.finishPoint != "" {
		sb.WriteString(fmt.Sprintf("    %s --> END[\"END\"]:::endNode\n", ge.graph.finishPoint))
		nodeNamesMap[ge.graph.finishPoint] = struct{}{} // Ensure finish point is in nodes
	}


	// Get sorted node names for consistent output
	sortedNodeNames := make([]string, 0, len(nodeNamesMap))
	for name := range nodeNamesMap {
		if name != "START" && name != "END" { // Exclude special nodes
			sortedNodeNames = append(sortedNodeNames, name)
		}
	}
	sort.Strings(sortedNodeNames)

	// Define all nodes (except START/END which are styled separately)
	for _, name := range sortedNodeNames {
		shape := "[%s]" // Default shape for generic nodes
		if name == ge.graph.entryPoint {
			shape = "[[\" %s \"]]" // Entry point specific shape (double border)
		} else if name == ge.graph.finishPoint {
			shape = "(((\" %s \")))" // Finish point specific shape (rounded rectangle)
		} else {
			shape = "[\"%s\"]" // Regular rectangle
		}
		sb.WriteString(fmt.Sprintf("    %s%s\n", name, fmt.Sprintf(shape, name)))
	}

	// Add edges
	for source, dest := range ge.graph.edges {
		sb.WriteString(fmt.Sprintf("    %s --> %s\n", source, dest))
	}

	// Apply styles
	sb.WriteString("    classDef startNode fill:#90EE90,stroke:#3C3,stroke-width:2px,color:#000;\n")
	sb.WriteString("    classDef endNode fill:#FFB6C1,stroke:#F00,stroke-width:2px,color:#000;\n")
	if ge.graph.entryPoint != "" {
		sb.WriteString(fmt.Sprintf("    class %s entryPointNode;\n", ge.graph.entryPoint))
		sb.WriteString("    classDef entryPointNode fill:#87CEEB,stroke:#369,stroke-width:2px,color:#000;\n")
	}
	if ge.graph.finishPoint != "" {
		sb.WriteString(fmt.Sprintf("    class %s finishPointNode;\n", ge.graph.finishPoint))
		sb.WriteString("    classDef finishPointNode fill:#ADD8E6,stroke:#00F,stroke-width:2px,color:#000;\n")
	}


	return sb.String()
}

// DrawDOT generates a DOT (Graphviz) representation of the StatefulGraph
func (ge *StatefulGraphExporter) DrawDOT() string {
	var sb strings.Builder

	sb.WriteString("digraph G {\n")
	sb.WriteString("    rankdir=TD;\n")
	sb.WriteString("    node [shape=box];\n")

	// Collect all unique node names
	nodeNamesMap := make(map[string]struct{})
	for name := range ge.graph.nodes {
		nodeNamesMap[name] = struct{}{}
	}

	// Add START node if there's an entry point
	if ge.graph.entryPoint != "" {
		sb.WriteString("    START [label=\"START\", shape=ellipse, style=filled, fillcolor=\"#90EE90\"];\n")
		sb.WriteString(fmt.Sprintf("    START -> %s;\n", ge.graph.entryPoint))
		nodeNamesMap[ge.graph.entryPoint] = struct{}{}
	}

	// Add END node if there's a finish point
	if ge.graph.finishPoint != "" {
		sb.WriteString("    END [label=\"END\", shape=ellipse, style=filled, fillcolor=\"#FFB6C1\"];\n")
		sb.WriteString(fmt.Sprintf("    %s -> END;\n", ge.graph.finishPoint))
		nodeNamesMap[ge.graph.finishPoint] = struct{}{}
	}

	// Add regular nodes and apply specific styling
	sortedNodeNames := make([]string, 0, len(nodeNamesMap))
	for name := range nodeNamesMap {
		if name != "START" && name != "END" {
			sortedNodeNames = append(sortedNodeNames, name)
		}
	}
	sort.Strings(sortedNodeNames)

	for _, name := range sortedNodeNames {
		var style string
		if name == ge.graph.entryPoint {
			style = ", style=filled, fillcolor=\"#87CEEB\", peripheries=2" // Double border for entry
		} else if name == ge.graph.finishPoint {
			style = ", shape=doubleoctagon, style=filled, fillcolor=\"#ADD8E6\"" // Octagon for finish
		}
		sb.WriteString(fmt.Sprintf("    %s [label=\" %s \"%s];\n", name, name, style))
	}

	// Add edges
	for source, dest := range ge.graph.edges {
		sb.WriteString(fmt.Sprintf("    %s -> %s;\n", source, dest))
	}

	sb.WriteString("}\n")
	return sb.String()
}

// DrawASCII (Placeholder for future implementation)
func (ge *StatefulGraphExporter) DrawASCII() string {
	return "ASCII visualization not yet implemented for StatefulGraph."
}
