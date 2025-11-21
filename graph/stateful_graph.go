package graph

import (
	"context"
	"encoding/json" // Added for JSON marshalling/unmarshalling Checkpoint state
	"fmt"
	"reflect"
	"sync" // Added for mutex to protect listeners slice
)



// StatefulGraph is a graph that manages a state object automatically.
// Nodes are functions that receive the current state and return a partial
// state, which is then merged back into the main state.
type StatefulGraph struct {
	stateType   reflect.Type
	nodes       map[string]reflect.Value
	edges       map[string]string
	entryPoint  string
	finishPoint string
	listeners   []NodeListener // New field for listeners
	mu          sync.RWMutex   // Mutex to protect listeners slice
}

// NewStatefulGraph creates a new StatefulGraph. It takes an instance of the
// state struct as a schema. For example: NewStatefulGraph(MyState{}).
func NewStatefulGraph(stateSchema interface{}) (*StatefulGraph, error) {
	t := reflect.TypeOf(stateSchema)
	if t.Kind() != reflect.Struct {
		return nil, fmt.Errorf("state schema must be a struct, but got %s", t.Kind())
	}

	return &StatefulGraph{
		stateType: t,
		nodes:     make(map[string]reflect.Value),
		edges:     make(map[string]string),
	}, nil
}

// AddListener adds a NodeListener to the graph.
func (g *StatefulGraph) AddListener(listener NodeListener) {
	g.mu.Lock()
	defer g.mu.Unlock()
	g.listeners = append(g.listeners, listener)
}

// RemoveListener removes a NodeListener from the graph.
func (g *StatefulGraph) RemoveListener(listener NodeListener) {
	g.mu.Lock()
	defer g.mu.Unlock()
	for i, l := range g.listeners {
		if l == listener {
			g.listeners = append(g.listeners[:i], g.listeners[i+1:]...)
			return
		}
	}
}

// notifyListeners calls OnNodeEvent for all registered listeners.
func (g *StatefulGraph) notifyListeners(ctx context.Context, event NodeEvent, nodeName string, state interface{}, err error) {
	g.mu.RLock()
	defer g.mu.RUnlock()
	for _, listener := range g.listeners {
		listener.OnNodeEvent(ctx, event, nodeName, state, err)
	}
}

// AddNode adds a new node to the graph. The node must be a function with a
// signature like: func(context.Context, T) (P, error)
// where T is the graph's state type and P is a struct or map[string]interface{}
// representing the partial state to be merged.
func (g *StatefulGraph) AddNode(name string, nodeFn interface{}) error {
	if _, exists := g.nodes[name]; exists {
		return fmt.Errorf("node with name '%s' already exists", name)
	}

	fnVal := reflect.ValueOf(nodeFn)
	if err := g.validateNodeFn(fnVal); err != nil {
		return fmt.Errorf("invalid node function '%s': %w", name, err)
	}

	g.nodes[name] = fnVal
	return nil
}

// validateNodeFn checks if the provided function has the expected signature.
func (g *StatefulGraph) validateNodeFn(fnVal reflect.Value) error {
	fnType := fnVal.Type()
	if fnType.Kind() != reflect.Func {
		return fmt.Errorf("node must be a function")
	}
	if fnType.NumIn() != 2 {
		return fmt.Errorf("node function must have 2 arguments (context, state), but has %d", fnType.NumIn())
	}
	if fnType.In(0) != reflect.TypeOf((*context.Context)(nil)).Elem() {
		return fmt.Errorf("first argument must be context.Context")
	}
	if fnType.In(1) != g.stateType {
		return fmt.Errorf("second argument must be of type %s, but got %s", g.stateType.Name(), fnType.In(1).Name())
	}
	if fnType.NumOut() != 2 {
		return fmt.Errorf("node function must have 2 return values (partial state, error), but has %d", fnType.NumOut())
	}
	if fnType.Out(1) != reflect.TypeOf((*error)(nil)).Elem() {
		return fmt.Errorf("second return value must be an error")
	}
	return nil
}

// SetEntryPoint defines the starting node of the graph.
func (g *StatefulGraph) SetEntryPoint(name string) error {
	if _, exists := g.nodes[name]; !exists {
		return fmt.Errorf("node '%s' not found in graph", name)
	}
	g.entryPoint = name
	return nil
}

// SetFinishPoint defines the terminal node of the graph.
func (g *StatefulGraph) SetFinishPoint(name string) error {
	if _, exists := g.nodes[name]; !exists {
		return fmt.Errorf("node '%s' not found in graph", name)
	}
	g.finishPoint = name
	return nil
}

// AddEdge creates a directed edge from a source node to a destination node.
func (g *StatefulGraph) AddEdge(source, dest string) error {
	if _, exists := g.nodes[source]; !exists {
		return fmt.Errorf("source node '%s' not found", source)
	}
	if _, exists := g.nodes[dest]; !exists {
		return fmt.Errorf("destination node '%s' not found", dest)
	}
	g.edges[source] = dest
	return nil
}

// runGraph is a shared execution loop for Invoke and Resume.
func (g *StatefulGraph) runGraph(ctx context.Context, initialState interface{}, startNode string) (interface{}, error) {
	statePtr := reflect.New(g.stateType)
	stateVal := statePtr.Elem()

	if initialState != nil {
		if err := mergeState(stateVal, reflect.ValueOf(initialState)); err != nil {
			return nil, fmt.Errorf("failed to merge initial state: %w", err)
		}
	}

	currentNodeName := startNode
	for {
		if currentNodeName == "" {
			return nil, fmt.Errorf("execution path ended unexpectedly")
		}

		nodeFn, exists := g.nodes[currentNodeName]
		if !exists {
			return nil, fmt.Errorf("node '%s' not found in execution path", currentNodeName)
		}

		g.notifyListeners(ctx, NodeEventStart, currentNodeName, stateVal.Interface(), nil)

		inputs := []reflect.Value{reflect.ValueOf(ctx), stateVal}
		outputs := nodeFn.Call(inputs)

		var nodeErr error
		if !outputs[1].IsNil() {
			nodeErr = outputs[1].Interface().(error)
		}

		if nodeErr != nil {
			g.notifyListeners(ctx, NodeEventError, currentNodeName, stateVal.Interface(), nodeErr)
			return nil, nodeErr
		}

		if err := mergeState(stateVal, outputs[0]); err != nil {
			g.notifyListeners(ctx, NodeEventError, currentNodeName, stateVal.Interface(), err)
			return nil, fmt.Errorf("failed to merge state from node '%s': %w", currentNodeName, err)
		}
		
		g.notifyListeners(ctx, NodeEventComplete, currentNodeName, stateVal.Interface(), nil)

		if currentNodeName == g.finishPoint {
			break
		}

		nextNode, hasNext := g.edges[currentNodeName]
		if !hasNext {
			break
		}
		currentNodeName = nextNode
	}

	return stateVal.Interface(), nil
}

// Invoke executes the graph with an initial state, starting from the entry point.
func (g *StatefulGraph) Invoke(ctx context.Context, initialState interface{}) (interface{}, error) {
	if g.entryPoint == "" {
		return nil, fmt.Errorf("entry point not set for the graph")
	}
	return g.runGraph(ctx, initialState, g.entryPoint)
}

// Resume executes the graph from a given checkpoint.
func (g *StatefulGraph) Resume(ctx context.Context, checkpoint *Checkpoint) (interface{}, error) {
	if checkpoint == nil {
		return nil, fmt.Errorf("cannot resume from a nil checkpoint")
	}

	// Deserialize the checkpoint state into a struct of the graph's stateType
	checkpointStateVal := reflect.New(g.stateType).Elem()
	if checkpoint.State != nil {
		// Marshal and Unmarshal to ensure proper type conversion for the interface{} state
		stateBytes, err := json.Marshal(checkpoint.State)
		if err != nil {
			return nil, fmt.Errorf("failed to marshal checkpoint state: %w", err)
		}
		if err := json.Unmarshal(stateBytes, checkpointStateVal.Addr().Interface()); err != nil {
			return nil, fmt.Errorf("failed to unmarshal checkpoint state into %s: %w", g.stateType.Name(), err)
		}
	}

	// Now, find the next node to execute based on the checkpointed node
	nextNode, hasNext := g.edges[checkpoint.NodeName]
	if !hasNext {
		// If the checkpointed node had no outgoing edge, it was a terminal node.
		// If it's also the finishPoint, the graph is effectively done.
		if checkpoint.NodeName == g.finishPoint {
			return checkpointStateVal.Interface(), nil // Graph already finished at checkpoint
		}
		// If not the finish point and no outgoing edge, it's an unexpected termination.
		return nil, fmt.Errorf("checkpointed node '%s' has no outgoing edge and is not a finish point", checkpoint.NodeName)
	}

	// Now run the graph from the next node with the deserialized state
	return g.runGraph(ctx, checkpointStateVal.Interface(), nextNode)
}

// GetGraph returns a StatefulGraphExporter for visualization.
func (g *StatefulGraph) GetGraph() *StatefulGraphExporter {
	return NewStatefulGraphExporter(g)
}

// mergeState merges fields from a source (map or struct) into a destination struct.
func mergeState(destStruct, src reflect.Value) error {
	if !destStruct.CanAddr() {
		return fmt.Errorf("destination struct must be addressable")
	}

	src = reflect.Indirect(src)

	switch src.Kind() {
	case reflect.Map:
		for _, key := range src.MapKeys() {
			fieldName := key.String()
			field := destStruct.FieldByName(fieldName)
			if !field.IsValid() || !field.CanSet() {
				continue // Field not found or not settable, skip
			}

			srcVal := src.MapIndex(key)
			// If srcVal is an interface{}, get its underlying element.
			if srcVal.IsValid() && srcVal.Kind() == reflect.Interface && !srcVal.IsNil() {
				srcVal = srcVal.Elem()
			}
            
            // If srcVal is still invalid after unwrapping (e.g., nil map value), skip.
            if !srcVal.IsValid() {
                continue
            }

			if srcVal.Type().AssignableTo(field.Type()) {
				field.Set(srcVal)
			} else if srcVal.Type().ConvertibleTo(field.Type()) {
				field.Set(srcVal.Convert(field.Type()))
			} else {
				return fmt.Errorf("cannot assign or convert value for field '%s' (type %s) from source type %s", fieldName, field.Type(), srcVal.Type())
			}
		}
	case reflect.Struct:
		for i := 0; i < src.NumField(); i++ {
			srcField := src.Field(i)
			fieldName := src.Type().Field(i).Name
			destField := destStruct.FieldByName(fieldName)
			if destField.IsValid() && destField.CanSet() {
				if srcField.Type().AssignableTo(destField.Type()) {
					destField.Set(srcField)
				} else if srcField.Type().ConvertibleTo(destField.Type()) {
					destField.Set(srcField.Convert(destField.Type()))
				} else {
					return fmt.Errorf("cannot assign or convert value for field '%s' (type %s) from source type %s", fieldName, destField.Type(), srcField.Type())
				}
			}
		}
	default:
		return fmt.Errorf("unsupported source type for merging: %s", src.Kind())
	}
	return nil
}