package graph

import (
	"context"
	"testing"
	"github.com/stretchr/testify/assert"
)

// Define a sample state struct for testing.
type MyTestState struct {
	Value   int
	Message string
}

// nodeA is a sample node function for testing.
// It modifies the state and returns a partial state.
func nodeA(ctx context.Context, state MyTestState) (map[string]interface{}, error) {
	return map[string]interface{}{"Value": state.Value + 1}, nil
}

// nodeB is another sample node function.
func nodeB(ctx context.Context, state MyTestState) (map[string]interface{}, error) {
	return map[string]interface{}{"Message": "Hello from Node B"}, nil
}

// nodeC is another sample node function.
func nodeC(ctx context.Context, state MyTestState) (map[string]interface{}, error) {
	return map[string]interface{}{"Message": state.Message + ", and Node C finished!"}, nil
}

func TestNewStatefulGraph(t *testing.T) {
	t.Run("should create a new stateful graph with a valid struct", func(t *testing.T) {
		g, err := NewStatefulGraph(MyTestState{})
		assert.NoError(t, err)
		assert.NotNil(t, g)
		assert.Equal(t, "MyTestState", g.stateType.Name())
	})

	t.Run("should return an error for non-struct state", func(t *testing.T) {
		_, err := NewStatefulGraph(123)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "state schema must be a struct")
	})
}

func TestStatefulGraph_AddNodeAndEdge(t *testing.T) {
	g, _ := NewStatefulGraph(MyTestState{})

	err := g.AddNode("A", nodeA)
	assert.NoError(t, err)

	err = g.AddNode("B", nodeB)
	assert.NoError(t, err)

	err = g.AddNode("C", nodeC) // Add nodeC
	assert.NoError(t, err)

	err = g.SetEntryPoint("A")
	assert.NoError(t, err)

	err = g.AddEdge("A", "B")
	assert.NoError(t, err)

	err = g.AddEdge("B", "C") // Add edge B -> C
	assert.NoError(t, err)

	_, nodeAExists := g.nodes["A"]
	assert.True(t, nodeAExists)
	_, nodeBExists := g.nodes["B"]
	assert.True(t, nodeBExists)
	_, nodeCExists := g.nodes["C"] // Check nodeC
	assert.True(t, nodeCExists)
	assert.Equal(t, "A", g.entryPoint)
	assert.Equal(t, "B", g.edges["A"])
	assert.Equal(t, "C", g.edges["B"]) // Check edge B -> C
}

func TestStatefulGraph_Invoke(t *testing.T) {
	g, _ := NewStatefulGraph(MyTestState{})
	_ = g.AddNode("A", nodeA)
	_ = g.AddNode("B", nodeB)
	_ = g.SetEntryPoint("A")
	_ = g.AddEdge("A", "B")
	_ = g.SetFinishPoint("B")

	initialState := MyTestState{Value: 10, Message: "Initial"}

	finalState, err := g.Invoke(context.Background(), initialState)

	assert.NoError(t, err)
	assert.NotNil(t, finalState)

	result, ok := finalState.(MyTestState)
	assert.True(t, ok)

	assert.Equal(t, 11, result.Value)
	assert.Equal(t, "Hello from Node B", result.Message)
}

func TestStatefulGraph_Resume(t *testing.T) {
	t.Run("should resume graph execution from a checkpoint", func(t *testing.T) {
		g, _ := NewStatefulGraph(MyTestState{})
		_ = g.AddNode("A", nodeA)
		_ = g.AddNode("B", nodeB)
		_ = g.AddNode("C", nodeC)
		_ = g.SetEntryPoint("A") // Entry point is A
		_ = g.AddEdge("A", "B")
		_ = g.AddEdge("B", "C")
		_ = g.SetFinishPoint("C") // Finish point is C

		// Simulate checkpoint after Node A executes
		// State after A: Value=11, Message="Initial"
		checkpointState := MyTestState{Value: 11, Message: "Initial"}
		checkpoint := &Checkpoint{
			ID:       "checkpoint-123",
			NodeName: "A", // Corrected: A is the last completed node, resume from next
			State:    checkpointState,
			Metadata: map[string]interface{}{},
		}

		resumedState, err := g.Resume(context.Background(), checkpoint)
		assert.NoError(t, err)
		assert.NotNil(t, resumedState)

		result, ok := resumedState.(MyTestState)
		assert.True(t, ok)

		// Node B executed (Message updated)
		// Node C executed (Message appended)
		assert.Equal(t, 11, result.Value) // Value should remain 11 as only A modified it
		assert.Equal(t, "Hello from Node B, and Node C finished!", result.Message)
	})
}
