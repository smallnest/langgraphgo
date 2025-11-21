package graph

import (
	"context"
	"fmt"
	"reflect" // Added
	"time"
)

// CheckpointableStatefulGraph extends StatefulGraph with checkpointing capabilities.
type CheckpointableStatefulGraph struct {
	*StatefulGraph
	config      CheckpointConfig
	executionID string
}

// NewCheckpointableStatefulGraph creates a new CheckpointableStatefulGraph with a default config.
func NewCheckpointableStatefulGraph(stateSchema interface{}) (*CheckpointableStatefulGraph, error) {
	sg, err := NewStatefulGraph(stateSchema)
	if err != nil {
		return nil, err
	}
	return &CheckpointableStatefulGraph{
		StatefulGraph: sg,
		config:        DefaultCheckpointConfig(),
		executionID:   generateExecutionID(),
	}, nil // Fixed: removed extra comma
}

// NewCheckpointableStatefulGraphWithConfig creates a new CheckpointableStatefulGraph with a custom config.
func NewCheckpointableStatefulGraphWithConfig(stateSchema interface{}, config CheckpointConfig) (*CheckpointableStatefulGraph, error) {
	sg, err := NewStatefulGraph(stateSchema)
	if err != nil {
		return nil, err
	}
	return &CheckpointableStatefulGraph{
		StatefulGraph: sg,
		config:        config,
		executionID:   generateExecutionID(),
	}, nil // Fixed: removed extra comma
}

// Invoke executes the graph with checkpointing.
// It wraps the underlying StatefulGraph's Invoke method and handles checkpointing.
func (csg *CheckpointableStatefulGraph) Invoke(ctx context.Context, initialState interface{}) (interface{}, error) {
	// Add the StatefulCheckpointListener to the StatefulGraph
	if csg.config.AutoSave && csg.config.Store != nil {
		checkpointListener := &StatefulCheckpointListener{
			store:       csg.config.Store,
			executionID: csg.executionID,
			stateType:   csg.stateType,
		}
		csg.StatefulGraph.AddListener(checkpointListener)
		// Ensure listener is removed after Invoke completes
		defer csg.StatefulGraph.RemoveListener(checkpointListener)
	}

	finalState, err := csg.StatefulGraph.Invoke(ctx, initialState)
	if err != nil {
		return nil, err
	}

	// Save a final checkpoint after successful completion if auto-save is enabled
	if csg.config.AutoSave && csg.config.Store != nil {
		checkpoint := &Checkpoint{
			ID:        generateCheckpointID(),
			NodeName:  "__end__", // Special node name for final state
			State:     finalState,
			Timestamp: time.Now(),
			Version:   1,
			Metadata: map[string]interface{}{
				"execution_id": csg.executionID,
				"status":       "completed",
			},
		}
		if saveErr := csg.config.Store.Save(ctx, checkpoint); saveErr != nil {
			fmt.Printf("Warning: failed to save final checkpoint: %v\n", saveErr)
		}
	}

	return finalState, nil
}

// SaveCheckpoint manually saves a checkpoint.
func (csg *CheckpointableStatefulGraph) SaveCheckpoint(ctx context.Context, nodeName string, state interface{}) error {
	checkpoint := &Checkpoint{
		ID:        generateCheckpointID(),
		NodeName:  nodeName,
		State:     state,
		Timestamp: time.Now(),
		Version:   1,
		Metadata: map[string]interface{}{
			"execution_id": csg.executionID,
		},
	}
	return csg.config.Store.Save(ctx, checkpoint)
}

// LoadCheckpoint loads a specific checkpoint.
func (csg *CheckpointableStatefulGraph) LoadCheckpoint(ctx context.Context, checkpointID string) (*Checkpoint, error) {
	return csg.config.Store.Load(ctx, checkpointID)
}

// ResumeFromCheckpoint resumes execution from a specific checkpoint.
func (csg *CheckpointableStatefulGraph) ResumeFromCheckpoint(ctx context.Context, checkpointID string) (interface{}, error) {
	checkpoint, err := csg.LoadCheckpoint(ctx, checkpointID)
	if err != nil {
		return nil, fmt.Errorf("failed to load checkpoint %s: %w", checkpointID, err)
	}

	// Add the StatefulCheckpointListener before resuming
	if csg.config.AutoSave && csg.config.Store != nil {
		checkpointListener := &StatefulCheckpointListener{
			store:       csg.config.Store,
			executionID: csg.executionID,
			stateType:   csg.stateType,
		}
		csg.StatefulGraph.AddListener(checkpointListener)
		// Ensure listener is removed after Resume completes
		defer csg.StatefulGraph.RemoveListener(checkpointListener)
	}

	// Use the embedded StatefulGraph's Resume method
	return csg.StatefulGraph.Resume(ctx, checkpoint)
}

// ClearCheckpoints removes all checkpoints for this execution.
func (csg *CheckpointableStatefulGraph) ClearCheckpoints(ctx context.Context) error {
	if csg.config.Store == nil {
		return nil // No store configured, nothing to clear
	}
	return csg.config.Store.Clear(ctx, csg.executionID)
}

// ListCheckpoints returns all checkpoints for this execution.
func (csg *CheckpointableStatefulGraph) ListCheckpoints(ctx context.Context) ([]*Checkpoint, error) {
	if csg.config.Store == nil {
		return nil, nil // No store configured, no checkpoints to list
	}
	return csg.config.Store.List(ctx, csg.executionID)
}

// StatefulCheckpointListener is a listener for StatefulGraph nodes to save checkpoints.
// It relies on NodeListener interface defined in stateful_graph.go.
type StatefulCheckpointListener struct {
	store       CheckpointStore
	executionID string
	stateType   reflect.Type
}

// OnNodeEvent implements the NodeListener interface for checkpointing.
// It is called after a node completes (NodeEventComplete) or errors (NodeEventError).
func (scl *StatefulCheckpointListener) OnNodeEvent(ctx context.Context, event NodeEvent, nodeName string, state interface{}, err error) {
	// Only save checkpoints on successful completion events
	if event != NodeEventComplete || err != nil {
		return
	}

	checkpoint := &Checkpoint{
		ID:        generateCheckpointID(),
		NodeName:  nodeName,
		State:     state, // state is already the full state of StatefulGraph
		Timestamp: time.Now(),
		Version:   1,
		Metadata: map[string]interface{}{
			"execution_id": scl.executionID,
		},
	}
	if saveErr := scl.store.Save(ctx, checkpoint); saveErr != nil {
		fmt.Printf("Warning: failed to save checkpoint for node %s: %v\n", nodeName, saveErr)
	}
}