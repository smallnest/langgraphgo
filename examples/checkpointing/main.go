package main

import (
	"context"
	"fmt"
	"time"

	"github.com/smallnest/langgraphgo/graph"
)

// ProcessState defines the state structure for our StatefulGraph
type ProcessState struct {
	Step    int
	Data    string
	History []string
}

func main() {
	// Configure checkpointing
	config := graph.CheckpointConfig{
		Store:          graph.NewMemoryCheckpointStore(),
		AutoSave:       true, // Auto-save after each node completes
		MaxCheckpoints: 5,
	}

	// Create a checkpointable stateful graph
	csg, err := graph.NewCheckpointableStatefulGraphWithConfig(ProcessState{}, config)
	if err != nil {
		panic(err)
	}

	// Add processing nodes
	csg.AddNode("step1", func(ctx context.Context, state ProcessState) (map[string]interface{}, error) {
		state.Step = 1
		state.Data = state.Data + " → Step1"
		state.History = append(state.History, "Completed Step 1")
		fmt.Println("Executing Step 1...")
		time.Sleep(500 * time.Millisecond) // Simulate work
		return map[string]interface{}{
			"Step":    state.Step,
			"Data":    state.Data,
			"History": state.History,
		}, nil
	})

	csg.AddNode("step2", func(ctx context.Context, state ProcessState) (map[string]interface{}, error) {
		state.Step = 2
		state.Data = state.Data + " → Step2"
		state.History = append(state.History, "Completed Step 2")
		fmt.Println("Executing Step 2...")
		time.Sleep(500 * time.Millisecond) // Simulate work
		return map[string]interface{}{
			"Step":    state.Step,
			"Data":    state.Data,
			"History": state.History,
		}, nil
	})

	csg.AddNode("step3", func(ctx context.Context, state ProcessState) (map[string]interface{}, error) {
		state.Step = 3
		state.Data = state.Data + " → Step3"
		state.History = append(state.History, "Completed Step 3")
		fmt.Println("Executing Step 3...")
		time.Sleep(500 * time.Millisecond) // Simulate work
		return map[string]interface{}{
			"Step":    state.Step,
			"Data":    state.Data,
			"History": state.History,
		}, nil
	})

	// Build the pipeline
	csg.SetEntryPoint("step1")
	csg.AddEdge("step1", "step2")
	csg.AddEdge("step2", "step3")
	csg.SetFinishPoint("step3") // Use SetFinishPoint for StatefulGraph

	ctx := context.Background()
	initialState := ProcessState{
		Step:    0,
		Data:    "Start",
		History: []string{"Initialized"},
	}

	fmt.Println("=== Starting execution with StatefulGraph checkpointing ===")

	// Execute with automatic checkpointing
	result, err := csg.Invoke(ctx, initialState)
	if err != nil {
		panic(err)
	}

	finalState := result.(ProcessState)
	fmt.Printf("\n=== Execution completed ===\n")
	fmt.Printf("Final Step: %d\n", finalState.Step)
	fmt.Printf("Final Data: %s\n", finalState.Data)
	fmt.Printf("History: %v\n", finalState.History)

	// List saved checkpoints
	checkpoints, err := csg.ListCheckpoints(ctx)
	if err != nil {
		panic(err)
	}

	fmt.Printf("\n=== Created %d checkpoints ===\n", len(checkpoints))
	for i, cp := range checkpoints {
		fmt.Printf("Checkpoint %d: ID=%s, Node=%s, Time=%v\n", i+1, cp.ID, cp.NodeName, cp.Timestamp)
	}

	// Demonstrate resuming from a checkpoint
	if len(checkpoints) >= 2 { // We need at least two checkpoints to resume from the middle
		// Let's resume from the checkpoint after step1 completes.
		// This will be the checkpoint associated with NodeName "step1"
		var checkpointToResumeFrom *graph.Checkpoint
		for _, cp := range checkpoints {
			if cp.NodeName == "step1" {
				checkpointToResumeFrom = cp
				break
			}
		}

		if checkpointToResumeFrom != nil {
			fmt.Printf("\n=== Resuming from checkpoint after %s (ID: %s) ===\n", checkpointToResumeFrom.NodeName, checkpointToResumeFrom.ID)
			resumedState, err := csg.ResumeFromCheckpoint(ctx, checkpointToResumeFrom.ID)
			if err != nil {
				fmt.Printf("Error resuming: %v\n", err)
			} else {
				resumed := resumedState.(ProcessState)
				fmt.Printf("Resumed Final Step: %d\n", resumed.Step)
				fmt.Printf("Resumed Final Data: %s\n", resumed.Data)
				fmt.Printf("Resumed History: %v\n", resumed.History)
			}
		} else {
			fmt.Println("\nCould not find a checkpoint after step1 to resume from.")
		}
	} else {
		fmt.Println("\nNot enough checkpoints to demonstrate resumption.")
	}
}
