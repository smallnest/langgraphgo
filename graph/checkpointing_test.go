package graph_test

import (
	"bytes"
	"context"
	"fmt"
	"strings"
	"testing"
	"time"

	"github.com/smallnest/langgraphgo/graph"
)

const (
	testNode   = "test_node"
	testResult = "test_result"
)

func TestMemoryCheckpointStore_SaveAndLoad(t *testing.T) {
	t.Parallel()

	store := graph.NewMemoryCheckpointStore()
	ctx := context.Background()

	checkpoint := &graph.Checkpoint{
		ID:        "test_checkpoint_1",
		NodeName:  testNode,
		State:     "test_state",
		Timestamp: time.Now(),
		Version:   1,
		Metadata: map[string]interface{}{
			"execution_id": "exec_123",
		},
	}

	// Test Save
	err := store.Save(ctx, checkpoint)
	if err != nil {
		t.Fatalf("Failed to save checkpoint: %v", err)
	}

	// Test Load
	loaded, err := store.Load(ctx, "test_checkpoint_1")
	if err != nil {
		t.Fatalf("Failed to load checkpoint: %v", err)
	}

	if loaded.ID != checkpoint.ID {
		t.Errorf("Expected ID %s, got %s", checkpoint.ID, loaded.ID)
	}

	if loaded.NodeName != checkpoint.NodeName {
		t.Errorf("Expected NodeName %s, got %s", checkpoint.NodeName, loaded.NodeName)
	}

	if loaded.State != checkpoint.State {
		t.Errorf("Expected State %v, got %v", checkpoint.State, loaded.State)
	}
}

func TestMemoryCheckpointStore_LoadNonExistent(t *testing.T) {
	t.Parallel()

	store := graph.NewMemoryCheckpointStore()
	ctx := context.Background()

	_, err := store.Load(ctx, "non_existent")
	if err == nil {
		t.Error("Expected error for non-existent checkpoint")
	}

	if !strings.Contains(err.Error(), "checkpoint not found") {
		t.Errorf("Expected 'checkpoint not found' error, got: %v", err)
	}
}

func TestMemoryCheckpointStore_List(t *testing.T) {
	t.Parallel()

	store := graph.NewMemoryCheckpointStore()
	ctx := context.Background()
	executionID := "exec_123"

	// Save multiple checkpoints
	checkpoints := []*graph.Checkpoint{
		{
			ID: "checkpoint_1",
			Metadata: map[string]interface{}{
				"execution_id": executionID,
			},
		},
		{
			ID: "checkpoint_2",
			Metadata: map[string]interface{}{
				"execution_id": executionID,
			},
		},
		{
			ID: "checkpoint_3",
			Metadata: map[string]interface{}{
				"execution_id": "different_exec",
			},
		},
	}

	for _, checkpoint := range checkpoints {
		err := store.Save(ctx, checkpoint)
		if err != nil {
			t.Fatalf("Failed to save checkpoint: %v", err)
		}
	}

	// List checkpoints for specific execution
	listed, err := store.List(ctx, executionID)
	if err != nil {
		t.Fatalf("Failed to list checkpoints: %v", err)
	}

	if len(listed) != 2 {
		t.Errorf("Expected 2 checkpoints, got %d", len(listed))
	}

	// Verify correct checkpoints returned
	ids := make(map[string]bool)
	for _, checkpoint := range listed {
		ids[checkpoint.ID] = true
	}

	if !ids["checkpoint_1"] || !ids["checkpoint_2"] {
		t.Error("Wrong checkpoints returned")
	}
}

func TestMemoryCheckpointStore_Delete(t *testing.T) {
	t.Parallel()

	store := graph.NewMemoryCheckpointStore()
	ctx := context.Background()

	checkpoint := &graph.Checkpoint{
		ID: "test_checkpoint",
	}

	// Save and verify
	err := store.Save(ctx, checkpoint)
	if err != nil {
		t.Fatalf("Failed to save checkpoint: %v", err)
	}

	_, err = store.Load(ctx, "test_checkpoint")
	if err != nil {
		t.Error("Checkpoint should exist before deletion")
	}

	// Delete
	err = store.Delete(ctx, "test_checkpoint")
	if err != nil {
		t.Fatalf("Failed to delete checkpoint: %v", err)
	}

	// Verify deletion
	_, err = store.Load(ctx, "test_checkpoint")
	if err == nil {
		t.Error("Checkpoint should not exist after deletion")
	}
}

func TestMemoryCheckpointStore_Clear(t *testing.T) {
	t.Parallel()

	store := graph.NewMemoryCheckpointStore()
	ctx := context.Background()
	executionID := "exec_123"

	// Save checkpoints
	checkpoints := []*graph.Checkpoint{
		{
			ID: "checkpoint_1",
			Metadata: map[string]interface{}{
				"execution_id": executionID,
			},
		},
		{
			ID: "checkpoint_2",
			Metadata: map[string]interface{}{
				"execution_id": executionID,
			},
		},
		{
			ID: "checkpoint_3",
			Metadata: map[string]interface{}{
				"execution_id": "different_exec",
			},
		},
	}

	for _, checkpoint := range checkpoints {
		err := store.Save(ctx, checkpoint)
		if err != nil {
			t.Fatalf("Failed to save checkpoint: %v", err)
		}
	}

	// Clear execution
	err := store.Clear(ctx, executionID)
	if err != nil {
		t.Fatalf("Failed to clear checkpoints: %v", err)
	}

	// Verify clearing
	listed, err := store.List(ctx, executionID)
	if err != nil {
		t.Fatalf("Failed to list checkpoints: %v", err)
	}

	if len(listed) != 0 {
		t.Errorf("Expected 0 checkpoints after clear, got %d", len(listed))
	}

	// Verify other execution's checkpoints still exist
	listed, err = store.List(ctx, "different_exec")
	if err != nil {
		t.Fatalf("Failed to list other execution's checkpoints: %v", err)
	}

	if len(listed) != 1 {
		t.Errorf("Expected 1 checkpoint for other execution, got %d", len(listed))
	}
}

func TestFileCheckpointStore_SaveAndLoad(t *testing.T) {
	t.Parallel()

	var buf bytes.Buffer
	store := graph.NewFileCheckpointStore(&buf, &buf)
	ctx := context.Background()

	checkpoint := &graph.Checkpoint{
		ID:       "test_checkpoint",
		NodeName: testNode,
		State:    "test_state",
		Version:  1,
	}

	// Test Save
	err := store.Save(ctx, checkpoint)
	if err != nil {
		t.Fatalf("Failed to save checkpoint: %v", err)
	}

	// Test Load
	loaded, err := store.Load(ctx, "test_checkpoint")
	if err != nil {
		t.Fatalf("Failed to load checkpoint: %v", err)
	}

	if loaded.ID != checkpoint.ID {
		t.Errorf("Expected ID %s, got %s", checkpoint.ID, loaded.ID)
	}
}

func TestCheckpointableRunnable_Basic(t *testing.T) {
	t.Parallel()

	// Create graph
	g := graph.NewListenableMessageGraph()

	g.AddNode("step1", func(ctx context.Context, state interface{}) (interface{}, error) {
		return "step1_result", nil
	})

	g.AddNode("step2", func(ctx context.Context, state interface{}) (interface{}, error) {
		return "step2_result", nil
	})

	g.AddEdge("step1", "step2")
	g.AddEdge("step2", graph.END)
	g.SetEntryPoint("step1")

	// Compile listenable runnable
	listenableRunnable, err := g.CompileListenable()
	if err != nil {
		t.Fatalf("Failed to compile listenable runnable: %v", err)
	}

	// Create checkpointable runnable
	config := graph.DefaultCheckpointConfig()
	checkpointableRunnable := graph.NewCheckpointableRunnable(listenableRunnable, config)

	ctx := context.Background()
	result, err := checkpointableRunnable.Invoke(ctx, "input")
	if err != nil {
		t.Fatalf("Execution failed: %v", err)
	}

	if result != "step2_result" {
		t.Errorf("Expected 'step2_result', got %v", result)
	}

	// Wait for async checkpoint operations
	time.Sleep(100 * time.Millisecond)

	// Check that checkpoints were created
	checkpoints, err := checkpointableRunnable.ListCheckpoints(ctx)
	if err != nil {
		t.Fatalf("Failed to list checkpoints: %v", err)
	}

	if len(checkpoints) != 2 {
		t.Errorf("Expected 2 checkpoints (one per completed node), got %d", len(checkpoints))
	}

	// Verify checkpoint contents
	nodeNames := make(map[string]bool)
	for _, checkpoint := range checkpoints {
		nodeNames[checkpoint.NodeName] = true
		if checkpoint.State == nil {
			t.Error("Checkpoint state should not be nil")
		}
		if checkpoint.Timestamp.IsZero() {
			t.Error("Checkpoint timestamp should be set")
		}
	}

	if !nodeNames["step1"] || !nodeNames["step2"] {
		t.Error("Expected checkpoints for both step1 and step2")
	}
}

func TestCheckpointableRunnable_ManualCheckpoint(t *testing.T) {
	t.Parallel()

	g := graph.NewListenableMessageGraph()
	g.AddNode(testNode, func(ctx context.Context, state interface{}) (interface{}, error) {
		return testResult, nil
	})
	g.AddEdge(testNode, graph.END)
	g.SetEntryPoint(testNode)

	listenableRunnable, err := g.CompileListenable()
	if err != nil {
		t.Fatalf("Failed to compile: %v", err)
	}

	config := graph.DefaultCheckpointConfig()
	checkpointableRunnable := graph.NewCheckpointableRunnable(listenableRunnable, config)

	ctx := context.Background()

	// Manual checkpoint save
	err = checkpointableRunnable.SaveCheckpoint(ctx, testNode, "manual_state")
	if err != nil {
		t.Fatalf("Failed to save manual checkpoint: %v", err)
	}

	checkpoints, err := checkpointableRunnable.ListCheckpoints(ctx)
	if err != nil {
		t.Fatalf("Failed to list checkpoints: %v", err)
	}

	if len(checkpoints) != 1 {
		t.Errorf("Expected 1 checkpoint, got %d", len(checkpoints))
	}

	checkpoint := checkpoints[0]
	if checkpoint.NodeName != testNode {
		t.Errorf("Expected node name 'test_node', got %s", checkpoint.NodeName)
	}

	if checkpoint.State != "manual_state" {
		t.Errorf("Expected state 'manual_state', got %v", checkpoint.State)
	}
}

func TestCheckpointableRunnable_LoadCheckpoint(t *testing.T) {
	t.Parallel()

	g := graph.NewListenableMessageGraph()
	g.AddNode(testNode, func(ctx context.Context, state interface{}) (interface{}, error) {
		return testResult, nil
	})
	g.AddEdge(testNode, graph.END)
	g.SetEntryPoint(testNode)

	listenableRunnable, err := g.CompileListenable()
	if err != nil {
		t.Fatalf("Failed to compile: %v", err)
	}

	config := graph.DefaultCheckpointConfig()
	checkpointableRunnable := graph.NewCheckpointableRunnable(listenableRunnable, config)

	ctx := context.Background()

	// Save checkpoint
	err = checkpointableRunnable.SaveCheckpoint(ctx, testNode, "saved_state")
	if err != nil {
		t.Fatalf("Failed to save checkpoint: %v", err)
	}

	checkpoints, err := checkpointableRunnable.ListCheckpoints(ctx)
	if err != nil {
		t.Fatalf("Failed to list checkpoints: %v", err)
	}

	if len(checkpoints) == 0 {
		t.Fatal("No checkpoints found")
	}

	checkpointID := checkpoints[0].ID

	// Load checkpoint
	loaded, err := checkpointableRunnable.LoadCheckpoint(ctx, checkpointID)
	if err != nil {
		t.Fatalf("Failed to load checkpoint: %v", err)
	}

	if loaded.State != "saved_state" {
		t.Errorf("Expected loaded state 'saved_state', got %v", loaded.State)
	}
}

func TestCheckpointableRunnable_ClearCheckpoints(t *testing.T) {
	t.Parallel()

	g := graph.NewListenableMessageGraph()
	g.AddNode(testNode, func(ctx context.Context, state interface{}) (interface{}, error) {
		return testResult, nil
	})
	g.AddEdge(testNode, graph.END)
	g.SetEntryPoint(testNode)

	listenableRunnable, err := g.CompileListenable()
	if err != nil {
		t.Fatalf("Failed to compile: %v", err)
	}

	config := graph.DefaultCheckpointConfig()
	checkpointableRunnable := graph.NewCheckpointableRunnable(listenableRunnable, config)

	ctx := context.Background()

	// Save some checkpoints
	err = checkpointableRunnable.SaveCheckpoint(ctx, "test_node1", "state1")
	if err != nil {
		t.Fatalf("Failed to save checkpoint 1: %v", err)
	}

	err = checkpointableRunnable.SaveCheckpoint(ctx, "test_node2", "state2")
	if err != nil {
		t.Fatalf("Failed to save checkpoint 2: %v", err)
	}

	// Verify checkpoints exist
	checkpoints, err := checkpointableRunnable.ListCheckpoints(ctx)
	if err != nil {
		t.Fatalf("Failed to list checkpoints: %v", err)
	}

	if len(checkpoints) != 2 {
		t.Errorf("Expected 2 checkpoints, got %d", len(checkpoints))
	}

	// Clear checkpoints
	err = checkpointableRunnable.ClearCheckpoints(ctx)
	if err != nil {
		t.Fatalf("Failed to clear checkpoints: %v", err)
	}

	// Verify checkpoints cleared
	checkpoints, err = checkpointableRunnable.ListCheckpoints(ctx)
	if err != nil {
		t.Fatalf("Failed to list checkpoints after clear: %v", err)
	}

	if len(checkpoints) != 0 {
		t.Errorf("Expected 0 checkpoints after clear, got %d", len(checkpoints))
	}
}

func TestCheckpointableMessageGraph_CompileCheckpointable(t *testing.T) {
	t.Parallel()

	g := graph.NewCheckpointableMessageGraph()

	g.AddNode(testNode, func(ctx context.Context, state interface{}) (interface{}, error) {
		return testResult, nil
	})
	g.AddEdge(testNode, graph.END)
	g.SetEntryPoint(testNode)

	checkpointableRunnable, err := g.CompileCheckpointable()
	if err != nil {
		t.Fatalf("Failed to compile checkpointable: %v", err)
	}

	ctx := context.Background()
	result, err := checkpointableRunnable.Invoke(ctx, "input")
	if err != nil {
		t.Fatalf("Execution failed: %v", err)
	}

	if result != testResult {
		t.Errorf("Expected 'test_result', got %v", result)
	}
}

func TestCheckpointableMessageGraph_CustomConfig(t *testing.T) {
	t.Parallel()

	store := graph.NewMemoryCheckpointStore()
	config := graph.CheckpointConfig{
		Store:          store,
		AutoSave:       false,
		SaveInterval:   time.Minute,
		MaxCheckpoints: 5,
	}

	g := graph.NewCheckpointableMessageGraphWithConfig(config)

	// Verify config is set
	actualConfig := g.GetCheckpointConfig()
	if actualConfig.AutoSave != false {
		t.Error("Expected AutoSave to be false")
	}

	if actualConfig.SaveInterval != time.Minute {
		t.Error("Expected SaveInterval to be 1 minute")
	}

	if actualConfig.MaxCheckpoints != 5 {
		t.Error("Expected MaxCheckpoints to be 5")
	}
}

// Integration test with comprehensive workflow
//
//nolint:gocognit,cyclop // Comprehensive integration test requires multiple scenarios
func TestCheckpointing_Integration(t *testing.T) {
	t.Parallel()

	// Create checkpointable graph
	g := graph.NewCheckpointableMessageGraph()

	// Build a multi-step pipeline
	g.AddNode("analyze", func(ctx context.Context, state interface{}) (interface{}, error) {
		data := state.(map[string]interface{})
		data["analyzed"] = true
		return data, nil
	})

	g.AddNode("process", func(ctx context.Context, state interface{}) (interface{}, error) {
		data := state.(map[string]interface{})
		data["processed"] = true
		return data, nil
	})

	g.AddNode("finalize", func(ctx context.Context, state interface{}) (interface{}, error) {
		data := state.(map[string]interface{})
		data["finalized"] = true
		return data, nil
	})

	g.AddEdge("analyze", "process")
	g.AddEdge("process", "finalize")
	g.AddEdge("finalize", graph.END)
	g.SetEntryPoint("analyze")

	// Compile checkpointable runnable
	runnable, err := g.CompileCheckpointable()
	if err != nil {
		t.Fatalf("Failed to compile: %v", err)
	}

	// Execute with initial state
	initialState := map[string]interface{}{
		"input": "test_data",
	}

	ctx := context.Background()
	result, err := runnable.Invoke(ctx, initialState)
	if err != nil {
		t.Fatalf("Execution failed: %v", err)
	}

	// Verify final result
	finalState := result.(map[string]interface{})
	if !finalState["analyzed"].(bool) {
		t.Error("Expected analyzed to be true")
	}
	if !finalState["processed"].(bool) {
		t.Error("Expected processed to be true")
	}
	if !finalState["finalized"].(bool) {
		t.Error("Expected finalized to be true")
	}

	// Wait for async checkpoint operations
	time.Sleep(100 * time.Millisecond)

	// Check checkpoints
	checkpoints, err := runnable.ListCheckpoints(ctx)
	if err != nil {
		t.Fatalf("Failed to list checkpoints: %v", err)
	}

	if len(checkpoints) != 3 {
		t.Errorf("Expected 3 checkpoints, got %d", len(checkpoints))
	}

	// Verify each checkpoint has the correct state progression
	checkpointsByNode := make(map[string]*graph.Checkpoint)
	for _, checkpoint := range checkpoints {
		checkpointsByNode[checkpoint.NodeName] = checkpoint
	}

	// Check analyze checkpoint
	if analyzeCP, exists := checkpointsByNode["analyze"]; exists {
		state := analyzeCP.State.(map[string]interface{})
		if !state["analyzed"].(bool) {
			t.Error("Analyze checkpoint should have analyzed=true")
		}
	} else {
		t.Error("Missing checkpoint for analyze node")
	}

	// Check process checkpoint
	if processCP, exists := checkpointsByNode["process"]; exists {
		state := processCP.State.(map[string]interface{})
		if !state["processed"].(bool) {
			t.Error("Process checkpoint should have processed=true")
		}
	} else {
		t.Error("Missing checkpoint for process node")
	}

	// Check finalize checkpoint
	if finalizeCP, exists := checkpointsByNode["finalize"]; exists {
		state := finalizeCP.State.(map[string]interface{})
		if !state["finalized"].(bool) {
			t.Error("Finalize checkpoint should have finalized=true")
		}
	} else {
		t.Error("Missing checkpoint for finalize node")
	}
}

func TestCheckpointListener_ErrorHandling(t *testing.T) {
	t.Parallel()

	g := graph.NewListenableMessageGraph()

	// Node that will fail
	g.AddNode("failing_node", func(ctx context.Context, state interface{}) (interface{}, error) {
		return nil, fmt.Errorf("simulated failure")
	})

	g.AddEdge("failing_node", graph.END)
	g.SetEntryPoint("failing_node")

	listenableRunnable, err := g.CompileListenable()
	if err != nil {
		t.Fatalf("Failed to compile: %v", err)
	}

	config := graph.DefaultCheckpointConfig()
	checkpointableRunnable := graph.NewCheckpointableRunnable(listenableRunnable, config)

	ctx := context.Background()

	// This should fail
	_, err = checkpointableRunnable.Invoke(ctx, "input")
	if err == nil {
		t.Error("Expected execution to fail")
	}

	// Wait for async operations
	time.Sleep(100 * time.Millisecond)

	// Should not have checkpoints for failed nodes
	checkpoints, err := checkpointableRunnable.ListCheckpoints(ctx)
	if err != nil {
		t.Fatalf("Failed to list checkpoints: %v", err)
	}

	// Should have no checkpoints since node failed
	if len(checkpoints) != 0 {
		t.Errorf("Expected no checkpoints for failed execution, got %d", len(checkpoints))
	}
}
