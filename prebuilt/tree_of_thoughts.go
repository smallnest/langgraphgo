package prebuilt

import (
	"context"
	"fmt"

	"github.com/smallnest/langgraphgo/graph"
)

type ThoughtState interface {
	IsValid() bool
	IsGoal() bool
	GetDescription() string
	Hash() string
}

type ThoughtGenerator interface {
	Generate(ctx context.Context, current ThoughtState) ([]ThoughtState, error)
}

type ThoughtEvaluator interface {
	Evaluate(ctx context.Context, state ThoughtState, pathLength int) (float64, error)
}

type SearchPath struct {
	States []ThoughtState
	Score  float64
}

type TreeOfThoughtsConfig struct {
	Generator    ThoughtGenerator
	Evaluator    ThoughtEvaluator
	MaxDepth     int
	MaxPaths     int
	Verbose      bool
	InitialState ThoughtState
}

// CreateTreeOfThoughtsAgentMap creates a ToT agent with map[string]any state
func CreateTreeOfThoughtsAgentMap(config TreeOfThoughtsConfig) (*graph.StateRunnable[map[string]any], error) {
	if config.Generator == nil || config.Evaluator == nil || config.InitialState == nil {
		return nil, fmt.Errorf("generator, evaluator and initial state are required")
	}
	if config.MaxDepth == 0 {
		config.MaxDepth = 10
	}
	if config.MaxPaths == 0 {
		config.MaxPaths = 5
	}

	workflow := graph.NewStateGraph[map[string]any]()

	workflow.AddNode("initialize", "Initialize search", func(ctx context.Context, state map[string]any) (map[string]any, error) {
		initialPath := SearchPath{States: []ThoughtState{config.InitialState}, Score: 0}
		visited := map[string]bool{config.InitialState.Hash(): true}
		return map[string]any{
			"active_paths":   []SearchPath{initialPath},
			"solution":       nil,
			"visited_states": visited,
			"iteration":      0,
		}, nil
	})

	workflow.AddNode("expand", "Expand paths", func(ctx context.Context, state map[string]any) (map[string]any, error) {
		activePaths, _ := state["active_paths"].([]SearchPath)
		visitedStates, ok := state["visited_states"].(map[string]bool)
		if !ok || visitedStates == nil {
			visitedStates = make(map[string]bool)
		}
		iteration, _ := state["iteration"].(int)

		var newPaths []SearchPath
		for _, path := range activePaths {
			currentState := path.States[len(path.States)-1]
			if currentState.IsGoal() {
				return map[string]any{"solution": path}, nil
			}
			if len(path.States) >= config.MaxDepth {
				continue
			}

			nextStates, _ := config.Generator.Generate(ctx, currentState)
			for _, next := range nextStates {
				if !next.IsValid() || visitedStates[next.Hash()] {
					continue
				}
				newPathStates := append([]ThoughtState{}, path.States...)
				newPathStates = append(newPathStates, next)
				newPaths = append(newPaths, SearchPath{States: newPathStates, Score: 0})
				visitedStates[next.Hash()] = true
			}
		}
		return map[string]any{"active_paths": newPaths, "visited_states": visitedStates, "iteration": iteration + 1}, nil
	})

	workflow.AddNode("evaluate", "Evaluate paths", func(ctx context.Context, state map[string]any) (map[string]any, error) {
		activePaths, _ := state["active_paths"].([]SearchPath)
		for i := range activePaths {
			last := activePaths[i].States[len(activePaths[i].States)-1]
			score, _ := config.Evaluator.Evaluate(ctx, last, len(activePaths[i].States))
			activePaths[i].Score = score
		}
		// Sort and prune (simple implementation)
		var pruned []SearchPath
		for _, p := range activePaths {
			if p.Score >= 0 {
				pruned = append(pruned, p)
			}
		}
		// Keep top MaxPaths (simplified)
		if len(pruned) > config.MaxPaths {
			pruned = pruned[:config.MaxPaths]
		}
		return map[string]any{"active_paths": pruned}, nil
	})

	workflow.SetEntryPoint("initialize")
	workflow.AddEdge("initialize", "expand")
	workflow.AddConditionalEdge("expand", func(ctx context.Context, state map[string]any) string {
		if s, ok := state["solution"].(SearchPath); ok && s.States != nil {
			return graph.END
		}
		if p, _ := state["active_paths"].([]SearchPath); len(p) == 0 {
			return graph.END
		}
		if iter, _ := state["iteration"].(int); iter >= config.MaxDepth {
			return graph.END
		}
		return "evaluate"
	})
	workflow.AddConditionalEdge("evaluate", func(ctx context.Context, state map[string]any) string {
		if p, _ := state["active_paths"].([]SearchPath); len(p) == 0 {
			return graph.END
		}
		return "expand"
	})

	return workflow.Compile()
}

// CreateTreeOfThoughtsAgent creates a generic Tree of Thoughts Agent
func CreateTreeOfThoughtsAgent[S any](
	config TreeOfThoughtsConfig,
	getActivePaths func(S) map[string]*SearchPath,
	setActivePaths func(S, map[string]*SearchPath) S,
	getSolution func(S) string,
	setSolution func(S, string) S,
	getVisited func(S) map[string]bool,
	setVisited func(S, map[string]bool) S,
	getIteration func(S) int,
	setIteration func(S, int) S,
) (*graph.StateRunnable[S], error) {
	if config.Generator == nil || config.Evaluator == nil || config.InitialState == nil {
		return nil, fmt.Errorf("generator, evaluator and initial state are required")
	}
	if config.MaxDepth == 0 {
		config.MaxDepth = 10
	}
	if config.MaxPaths == 0 {
		config.MaxPaths = 5
	}

	workflow := graph.NewStateGraph[S]()

	workflow.AddNode("initialize", "Initialize search", func(ctx context.Context, state S) (S, error) {
		initialPath := SearchPath{States: []ThoughtState{config.InitialState}, Score: 0}
		paths := map[string]*SearchPath{"initial": &initialPath}
		visited := map[string]bool{config.InitialState.Hash(): true}
		state = setActivePaths(state, paths)
		state = setVisited(state, visited)
		state = setIteration(state, 0)
		return state, nil
	})

	workflow.AddNode("expand", "Expand paths", func(ctx context.Context, state S) (S, error) {
		activePaths := getActivePaths(state)
		visitedStates := getVisited(state)
		iteration := getIteration(state)

		newPaths := make(map[string]*SearchPath)
		for id, path := range activePaths {
			currentState := path.States[len(path.States)-1]
			if currentState.IsGoal() {
				state = setSolution(state, "Goal reached in path: "+id)
				return state, nil
			}
			if len(path.States) >= config.MaxDepth {
				continue
			}

			nextStates, _ := config.Generator.Generate(ctx, currentState)
			for i, next := range nextStates {
				if !next.IsValid() || visitedStates[next.Hash()] {
					continue
				}
				newPathStates := append([]ThoughtState{}, path.States...)
				newPathStates = append(newPathStates, next)
				newPaths[fmt.Sprintf("%s-%d", id, i)] = &SearchPath{States: newPathStates, Score: 0}
				visitedStates[next.Hash()] = true
			}
		}
		state = setActivePaths(state, newPaths)
		state = setVisited(state, visitedStates)
		state = setIteration(state, iteration+1)
		return state, nil
	})

	workflow.AddNode("evaluate", "Evaluate paths", func(ctx context.Context, state S) (S, error) {
		activePaths := getActivePaths(state)
		for _, path := range activePaths {
			last := path.States[len(path.States)-1]
			score, _ := config.Evaluator.Evaluate(ctx, last, len(path.States))
			path.Score = score
		}
		// Simplified pruning and top-k
		state = setActivePaths(state, activePaths) // Update state
		return state, nil
	})

	workflow.SetEntryPoint("initialize")
	workflow.AddEdge("initialize", "expand")
	workflow.AddConditionalEdge("expand", func(ctx context.Context, state S) string {
		if getSolution(state) != "" {
			return graph.END
		}
		if len(getActivePaths(state)) == 0 {
			return graph.END
		}
		if getIteration(state) >= config.MaxDepth {
			return graph.END
		}
		return "evaluate"
	})
	workflow.AddConditionalEdge("evaluate", func(ctx context.Context, state S) string {
		if len(getActivePaths(state)) == 0 {
			return graph.END
		}
		return "expand"
	})

	return workflow.Compile()
}

// CreateTreeOfThoughtsAgentForState creates a type-safe Tree of Thoughts Agent graph pre-configured for TreeOfThoughtsState.
// This provides compile-time type safety without requiring manual getter/setter closures.
func CreateTreeOfThoughtsAgentForState(config TreeOfThoughtsConfig) (*graph.StateRunnable[TreeOfThoughtsState], error) {
	return CreateTreeOfThoughtsAgent[TreeOfThoughtsState](
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
}
