package graph

import (
	"context"

	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

// Define a sample state struct for testing. (Same as in stateful_graph_test.go)
type MyComplexState struct {
	Value      int
	Message    string
	History    []string
	IsFinished bool
}

// nodeIncValue increments Value and adds a message to History.
func nodeIncValue(ctx context.Context, state MyComplexState) (map[string]interface{}, error) {
	return map[string]interface{}{
		"Value":   state.Value + 1,
		"History": append(state.History, "nodeIncValue executed"),
	}, nil
}

// nodeAddMessage adds a specific message and sets IsFinished.
func nodeAddMessage(ctx context.Context, state MyComplexState) (map[string]interface{}, error) {
	return map[string]interface{}{
		"Message":    "Processed by nodeAddMessage",
		"History":    append(state.History, "nodeAddMessage executed"),
		"IsFinished": true,
	}, nil
}

// nodeConditional routes based on Value.
func nodeConditional(ctx context.Context, state MyComplexState) (string, error) {
	if state.Value > 10 {
		return "nodeFinished", nil
	}
	return "nodeIncValue", nil // Loop back for more processing
}

// nodeFinished marks the end.
func nodeFinished(ctx context.Context, state MyComplexState) (map[string]interface{}, error) {
	return map[string]interface{}{
		"Message": "Graph finished as Value > 10",
		"History": append(state.History, "nodeFinished executed"),
	}, nil
}

func TestNewCheckpointableStatefulGraph(t *testing.T) {
	t.Run("should create with default config", func(t *testing.T) {
		csg, err := NewCheckpointableStatefulGraph(MyComplexState{})
		assert.NoError(t, err)
		assert.NotNil(t, csg)
		assert.NotNil(t, csg.StatefulGraph)
		assert.NotNil(t, csg.config.Store)
		assert.True(t, csg.config.AutoSave)
		assert.NotEmpty(t, csg.executionID)
	})

	t.Run("should create with custom config", func(t *testing.T) {
		customConfig := CheckpointConfig{
			Store:          NewMemoryCheckpointStore(),
			AutoSave:       false,
			SaveInterval:   5 * time.Second,
			MaxCheckpoints: 5,
		}
		csg, err := NewCheckpointableStatefulGraphWithConfig(MyComplexState{}, customConfig)
		assert.NoError(t, err)
		assert.NotNil(t, csg)
		assert.False(t, csg.config.AutoSave)
		assert.Equal(t, 5*time.Second, csg.config.SaveInterval)
	})
}

func TestCheckpointableStatefulGraph_InvokeAndResume(t *testing.T) {
	t.Run("should save checkpoint and resume execution", func(t *testing.T) {
		store := NewMemoryCheckpointStore()
		config := CheckpointConfig{
			Store:    store,
			AutoSave: true, // Enable auto-save after each node
		}

		// --- First run: partial execution and save checkpoints ---
		graph1, _ := NewCheckpointableStatefulGraphWithConfig(MyComplexState{}, config)
		_ = graph1.AddNode("incValue", nodeIncValue)
		_ = graph1.AddNode("addMessage", nodeAddMessage)
		_ = graph1.AddEdge("incValue", "addMessage")
		_ = graph1.SetEntryPoint("incValue")
		_ = graph1.SetFinishPoint("addMessage")

		initialState1 := MyComplexState{Value: 5, Message: "Start", History: []string{}}
		// Invoke will run "incValue" then "addMessage", saving checkpoints for each
		finalState1, err := graph1.Invoke(context.Background(), initialState1)
		assert.NoError(t, err)
		assert.NotNil(t, finalState1)

		result1, ok := finalState1.(MyComplexState)
		assert.True(t, ok)
		assert.Equal(t, 6, result1.Value)
		assert.Equal(t, "Processed by nodeAddMessage", result1.Message)
		assert.Len(t, result1.History, 2)
		assert.Contains(t, result1.History, "nodeIncValue executed")
		assert.Contains(t, result1.History, "nodeAddMessage executed")
		assert.True(t, result1.IsFinished)

		// Check if checkpoints were saved
		checkpoints, err := store.List(context.Background(), graph1.executionID)
		assert.NoError(t, err)
		assert.GreaterOrEqual(t, len(checkpoints), 3, "Expected at least 3 checkpoints (incValue, addMessage, and final __end__)")

		// Find the checkpoint after "incValue"
		var checkpointAfterIncValue *Checkpoint
		for _, cp := range checkpoints {
			if cp.NodeName == "incValue" { // Corrected: Checkpoint after "incValue" execution has NodeName "incValue"
				checkpointAfterIncValue = cp
				break
			}
		}
		assert.NotNil(t, checkpointAfterIncValue, "Checkpoint after incValue not found")

		// --- Second run: resume from checkpoint ---
		graph2, _ := NewCheckpointableStatefulGraphWithConfig(MyComplexState{}, config)
		_ = graph2.AddNode("incValue", nodeIncValue)
		_ = graph2.AddNode("addMessage", nodeAddMessage)
		_ = graph2.AddEdge("incValue", "addMessage")
		_ = graph2.SetEntryPoint("incValue") // Entry point doesn't matter for resume
		_ = graph2.SetFinishPoint("addMessage")

		resumedFinalState, err := graph2.ResumeFromCheckpoint(context.Background(), checkpointAfterIncValue.ID)
		assert.NoError(t, err)
		assert.NotNil(t, resumedFinalState)

		result2, ok := resumedFinalState.(MyComplexState)
		assert.True(t, ok)

		// Verify state after resume
		assert.Equal(t, 6, result2.Value) // Value should be 6 from first run
		assert.Equal(t, "Processed by nodeAddMessage", result2.Message)
		assert.Len(t, result2.History, 2)
		assert.Contains(t, result2.History, "nodeIncValue executed")
		assert.Contains(t, result2.History, "nodeAddMessage executed")
		assert.True(t, result2.IsFinished)
	})

	t.Run("should handle no checkpoint store configured", func(t *testing.T) {
		config := CheckpointConfig{
			Store:    nil, // No store
			AutoSave: true,
		}
		csg, err := NewCheckpointableStatefulGraphWithConfig(MyComplexState{}, config)
		assert.NoError(t, err)

		_ = csg.AddNode("incValue", nodeIncValue)
		_ = csg.AddNode("addMessage", nodeAddMessage)
		_ = csg.AddEdge("incValue", "addMessage")
		_ = csg.SetEntryPoint("incValue")
		_ = csg.SetFinishPoint("addMessage")

		initialState := MyComplexState{Value: 5}
		finalState, err := csg.Invoke(context.Background(), initialState)
		assert.NoError(t, err)
		assert.NotNil(t, finalState)

		// No error expected, and no checkpoints should be returned
		checkpoints, err := csg.ListCheckpoints(context.Background())
		assert.NoError(t, err)
		assert.Empty(t, checkpoints)
	})

	// TODO: Add more tests for:
	// - Manual SaveCheckpoint and LoadCheckpoint
	// - ClearCheckpoints
	// - Graph with conditional edges and resume
	// - Error handling during node execution and checkpointing
}
