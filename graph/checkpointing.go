package graph

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"sync"
	"time"
)

// Checkpoint represents a saved state at a specific point in execution
type Checkpoint struct {
	ID        string                 `json:"id"`
	NodeName  string                 `json:"node_name"`
	State     interface{}            `json:"state"`
	Metadata  map[string]interface{} `json:"metadata"`
	Timestamp time.Time              `json:"timestamp"`
	Version   int                    `json:"version"`
}

// CheckpointStore defines the interface for checkpoint persistence
type CheckpointStore interface {
	// Save stores a checkpoint
	Save(ctx context.Context, checkpoint *Checkpoint) error

	// Load retrieves a checkpoint by ID
	Load(ctx context.Context, checkpointID string) (*Checkpoint, error)

	// List returns all checkpoints for a given execution
	List(ctx context.Context, executionID string) ([]*Checkpoint, error)

	// Delete removes a checkpoint
	Delete(ctx context.Context, checkpointID string) error

	// Clear removes all checkpoints for an execution
	Clear(ctx context.Context, executionID string) error
}

// MemoryCheckpointStore provides in-memory checkpoint storage
type MemoryCheckpointStore struct {
	checkpoints map[string]*Checkpoint
	mutex       sync.RWMutex
}

// NewMemoryCheckpointStore creates a new in-memory checkpoint store
func NewMemoryCheckpointStore() *MemoryCheckpointStore {
	return &MemoryCheckpointStore{
		checkpoints: make(map[string]*Checkpoint),
	}
}

// Save implements CheckpointStore interface
func (m *MemoryCheckpointStore) Save(_ context.Context, checkpoint *Checkpoint) error {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	m.checkpoints[checkpoint.ID] = checkpoint
	return nil
}

// Load implements CheckpointStore interface
func (m *MemoryCheckpointStore) Load(_ context.Context, checkpointID string) (*Checkpoint, error) {
	m.mutex.RLock()
	defer m.mutex.RUnlock()

	checkpoint, exists := m.checkpoints[checkpointID]
	if !exists {
		return nil, fmt.Errorf("checkpoint not found: %s", checkpointID)
	}

	return checkpoint, nil
}

// List implements CheckpointStore interface
func (m *MemoryCheckpointStore) List(_ context.Context, executionID string) ([]*Checkpoint, error) {
	m.mutex.RLock()
	defer m.mutex.RUnlock()

	var checkpoints []*Checkpoint
	for _, checkpoint := range m.checkpoints {
		if execID, ok := checkpoint.Metadata["execution_id"].(string); ok && execID == executionID {
			checkpoints = append(checkpoints, checkpoint)
		}
	}
	return checkpoints, nil
}

// Delete implements CheckpointStore interface
func (m *MemoryCheckpointStore) Delete(_ context.Context, checkpointID string) error {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	delete(m.checkpoints, checkpointID)
	return nil
}

// Clear implements CheckpointStore interface
func (m *MemoryCheckpointStore) Clear(_ context.Context, executionID string) error {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	for id, checkpoint := range m.checkpoints {
		if execID, ok := checkpoint.Metadata["execution_id"].(string); ok && execID == executionID {
			delete(m.checkpoints, id)
		}
	}

	return nil
}

// FileCheckpointStore provides file-based checkpoint storage
type FileCheckpointStore struct {
	writer io.Writer
	reader io.Reader
	mutex  sync.RWMutex
}

// NewFileCheckpointStore creates a new file-based checkpoint store
func NewFileCheckpointStore(writer io.Writer, reader io.Reader) *FileCheckpointStore {
	return &FileCheckpointStore{
		writer: writer,
		reader: reader,
	}
}

// Save implements CheckpointStore interface for file storage
func (f *FileCheckpointStore) Save(_ context.Context, checkpoint *Checkpoint) error {
	f.mutex.Lock()
	defer f.mutex.Unlock()

	data, err := json.Marshal(checkpoint)
	if err != nil {
		return fmt.Errorf("failed to marshal checkpoint: %w", err)
	}

	_, err = f.writer.Write(data)
	if err != nil {
		return fmt.Errorf("failed to write checkpoint: %w", err)
	}

	return nil
}

// Load implements CheckpointStore interface for file storage
func (f *FileCheckpointStore) Load(_ context.Context, checkpointID string) (*Checkpoint, error) {
	f.mutex.RLock()
	defer f.mutex.RUnlock()

	data, err := io.ReadAll(f.reader)
	if err != nil {
		return nil, fmt.Errorf("failed to read checkpoint: %w", err)
	}

	var checkpoint Checkpoint
	err = json.Unmarshal(data, &checkpoint)
	if err != nil {
		return nil, fmt.Errorf("failed to unmarshal checkpoint: %w", err)
	}

	if checkpoint.ID != checkpointID {
		return nil, fmt.Errorf("checkpoint not found: %s", checkpointID)
	}

	return &checkpoint, nil
}

// List implements CheckpointStore interface for file storage
func (f *FileCheckpointStore) List(_ context.Context, _ string) ([]*Checkpoint, error) {
	// For file storage, this would typically involve reading from multiple files
	// This is a simplified implementation
	return nil, fmt.Errorf("list operation not implemented for file store")
}

// Delete implements CheckpointStore interface for file storage
func (f *FileCheckpointStore) Delete(_ context.Context, checkpointID string) error {
	// For file storage, this would involve file system operations
	return fmt.Errorf("delete operation not implemented for file store")
}

// Clear implements CheckpointStore interface for file storage
func (f *FileCheckpointStore) Clear(_ context.Context, executionID string) error {
	// For file storage, this would involve file system operations
	return fmt.Errorf("clear operation not implemented for file store")
}

// CheckpointConfig configures checkpointing behavior
type CheckpointConfig struct {
	// Store is the checkpoint storage backend
	Store CheckpointStore

	// AutoSave enables automatic checkpointing after each node
	AutoSave bool

	// SaveInterval specifies how often to save (when AutoSave is false)
	SaveInterval time.Duration

	// MaxCheckpoints limits the number of checkpoints to keep
	MaxCheckpoints int
}

// DefaultCheckpointConfig returns a default checkpoint configuration
func DefaultCheckpointConfig() CheckpointConfig {
	return CheckpointConfig{
		Store:          NewMemoryCheckpointStore(),
		AutoSave:       true,
		SaveInterval:   30 * time.Second,
		MaxCheckpoints: 10,
	}
}

// CheckpointableRunnable wraps a runnable with checkpointing capabilities
type CheckpointableRunnable struct {
	runnable *ListenableRunnable
	config   CheckpointConfig

	executionID string
}

// NewCheckpointableRunnable creates a new checkpointable runnable
func NewCheckpointableRunnable(runnable *ListenableRunnable, config CheckpointConfig) *CheckpointableRunnable {
	return &CheckpointableRunnable{
		runnable:    runnable,
		config:      config,
		executionID: generateExecutionID(),
	}
}

// Invoke executes the graph with checkpointing
func (cr *CheckpointableRunnable) Invoke(ctx context.Context, initialState interface{}) (interface{}, error) {
	// Create checkpointing listener
	checkpointListener := &CheckpointListener{
		store:       cr.config.Store,
		executionID: cr.executionID,
		autoSave:    cr.config.AutoSave,
	}

	// Add checkpoint listener to all nodes
	for _, node := range cr.runnable.listenableNodes {
		node.AddListener(checkpointListener)
	}

	defer func() {
		// Clean up: remove checkpoint listener from all nodes
		for _, node := range cr.runnable.listenableNodes {
			node.RemoveListener(checkpointListener)
		}
	}()

	return cr.runnable.Invoke(ctx, initialState)
}

// SaveCheckpoint manually saves a checkpoint
func (cr *CheckpointableRunnable) SaveCheckpoint(ctx context.Context, nodeName string, state interface{}) error {
	checkpoint := &Checkpoint{
		ID:        generateCheckpointID(),
		NodeName:  nodeName,
		State:     state,
		Timestamp: time.Now(),
		Version:   1,
		Metadata: map[string]interface{}{
			"execution_id": cr.executionID,
		},
	}

	return cr.config.Store.Save(ctx, checkpoint)
}

// LoadCheckpoint loads a specific checkpoint
func (cr *CheckpointableRunnable) LoadCheckpoint(ctx context.Context, checkpointID string) (*Checkpoint, error) {
	return cr.config.Store.Load(ctx, checkpointID)
}

// ListCheckpoints returns all checkpoints for this execution
func (cr *CheckpointableRunnable) ListCheckpoints(ctx context.Context) ([]*Checkpoint, error) {
	return cr.config.Store.List(ctx, cr.executionID)
}

// ResumeFromCheckpoint resumes execution from a specific checkpoint
func (cr *CheckpointableRunnable) ResumeFromCheckpoint(ctx context.Context, checkpointID string) (interface{}, error) {
	checkpoint, err := cr.LoadCheckpoint(ctx, checkpointID)
	if err != nil {
		return nil, fmt.Errorf("failed to load checkpoint: %w", err)
	}

	// Resume execution from the checkpointed state
	// This would require the graph to support starting from a specific node
	// For now, we'll return the checkpointed state
	return checkpoint.State, nil
}

// ClearCheckpoints removes all checkpoints for this execution
func (cr *CheckpointableRunnable) ClearCheckpoints(ctx context.Context) error {
	return cr.config.Store.Clear(ctx, cr.executionID)
}

// CheckpointListener automatically creates checkpoints during execution
type CheckpointListener struct {
	store       CheckpointStore
	executionID string
	autoSave    bool
}

// OnNodeEvent implements the NodeListener interface for checkpointing
func (cl *CheckpointListener) OnNodeEvent(ctx context.Context, event NodeEvent, nodeName string, state interface{}, err error) {
	if !cl.autoSave || event != NodeEventComplete {
		return
	}

	if err != nil {
		// Don't save checkpoints for failed nodes
		return
	}

	checkpoint := &Checkpoint{
		ID:        generateCheckpointID(),
		NodeName:  nodeName,
		State:     state,
		Timestamp: time.Now(),
		Version:   1,
		Metadata: map[string]interface{}{
			"execution_id": cl.executionID,
			"event":        event,
		},
	}

	// Save checkpoint asynchronously to avoid blocking execution
	go func(ctx context.Context) {
		if saveErr := cl.store.Save(ctx, checkpoint); saveErr != nil {
			// Error is intentionally ignored to avoid blocking execution
			_ = saveErr
		}
	}(ctx)
}

// CheckpointableMessageGraph extends ListenableMessageGraph with checkpointing
type CheckpointableMessageGraph struct {
	*ListenableMessageGraph
	config CheckpointConfig
}

// NewCheckpointableMessageGraph creates a new checkpointable message graph
func NewCheckpointableMessageGraph() *CheckpointableMessageGraph {
	return &CheckpointableMessageGraph{
		ListenableMessageGraph: NewListenableMessageGraph(),
		config:                 DefaultCheckpointConfig(),
	}
}

// NewCheckpointableMessageGraphWithConfig creates a checkpointable graph with custom config
func NewCheckpointableMessageGraphWithConfig(config CheckpointConfig) *CheckpointableMessageGraph {
	return &CheckpointableMessageGraph{
		ListenableMessageGraph: NewListenableMessageGraph(),
		config:                 config,
	}
}

// CompileCheckpointable compiles the graph into a checkpointable runnable
func (g *CheckpointableMessageGraph) CompileCheckpointable() (*CheckpointableRunnable, error) {
	listenableRunnable, err := g.CompileListenable()
	if err != nil {
		return nil, err
	}

	return NewCheckpointableRunnable(listenableRunnable, g.config), nil
}

// SetCheckpointConfig updates the checkpointing configuration
func (g *CheckpointableMessageGraph) SetCheckpointConfig(config CheckpointConfig) {
	g.config = config
}

// GetCheckpointConfig returns the current checkpointing configuration
func (g *CheckpointableMessageGraph) GetCheckpointConfig() CheckpointConfig {
	return g.config
}

// Helper functions
func generateExecutionID() string {
	return fmt.Sprintf("exec_%d", time.Now().UnixNano())
}

func generateCheckpointID() string {
	return fmt.Sprintf("checkpoint_%d", time.Now().UnixNano())
}
