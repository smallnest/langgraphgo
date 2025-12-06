package ptc

import (
	"context"
	"encoding/json"
	"fmt"
	"net"
	"net/http"
	"sync"
	"time"

	"github.com/smallnest/langgraphgo/log"
	"github.com/tmc/langchaingo/tools"
)

// ToolServer provides an HTTP API for tool execution
// This allows code in any language to call Go tools via HTTP
type ToolServer struct {
	tools   map[string]tools.Tool
	server  *http.Server
	port    int
	mu      sync.RWMutex
	started bool
	logger  log.Logger // Optional logger for debugging and monitoring
}

// ToolRequest represents a tool execution request
type ToolRequest struct {
	ToolName string      `json:"tool_name"`
	Input    interface{} `json:"input"`
}

// ToolResponse represents a tool execution response
type ToolResponse struct {
	Success bool        `json:"success"`
	Result  string      `json:"result"`
	Error   string      `json:"error,omitempty"`
	Tool    string      `json:"tool"`
	Input   interface{} `json:"input"`
}

// NewToolServer creates a new tool server
func NewToolServer(toolList []tools.Tool) *ToolServer {
	toolMap := make(map[string]tools.Tool)
	for _, tool := range toolList {
		toolMap[tool.Name()] = tool
	}

	return &ToolServer{
		tools:   toolMap,
		port:    0, // Will be assigned automatically
		started: false,
		logger:  &log.NoOpLogger{}, // Default to no logging
	}
}

// SetLogger sets a custom logger for the tool server
func (ts *ToolServer) SetLogger(logger log.Logger) {
	ts.mu.Lock()
	defer ts.mu.Unlock()
	ts.logger = logger
}

// Start starts the tool server on an available port
func (ts *ToolServer) Start(ctx context.Context) error {
	ts.mu.Lock()
	defer ts.mu.Unlock()

	if ts.started {
		return fmt.Errorf("server already started")
	}

	// Find an available port
	listener, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		return fmt.Errorf("failed to find available port: %w", err)
	}
	ts.port = listener.Addr().(*net.TCPAddr).Port
	ts.logger.Info("Tool server starting on port %d", ts.port)

	mux := http.NewServeMux()
	mux.HandleFunc("/tools", ts.handleListTools)
	mux.HandleFunc("/call", ts.handleCallTool)
	mux.HandleFunc("/health", ts.handleHealth)

	ts.server = &http.Server{
		Handler:      mux,
		ReadTimeout:  30 * time.Second,
		WriteTimeout: 30 * time.Second,
	}

	ts.started = true

	// Start server in goroutine
	go func() {
		if err := ts.server.Serve(listener); err != nil && err != http.ErrServerClosed {
			ts.logger.Error("Tool server error: %v", err)
		}
	}()

	// Wait a bit for server to start
	time.Sleep(100 * time.Millisecond)
	ts.logger.Info("Tool server started successfully on http://127.0.0.1:%d", ts.port)

	return nil
}

// Stop stops the tool server
func (ts *ToolServer) Stop(ctx context.Context) error {
	ts.mu.Lock()
	defer ts.mu.Unlock()

	if !ts.started {
		return nil
	}

	ts.started = false
	if ts.server != nil {
		return ts.server.Shutdown(ctx)
	}
	return nil
}

// GetPort returns the port the server is listening on
func (ts *ToolServer) GetPort() int {
	ts.mu.RLock()
	defer ts.mu.RUnlock()
	return ts.port
}

// GetBaseURL returns the base URL of the server
func (ts *ToolServer) GetBaseURL() string {
	return fmt.Sprintf("http://127.0.0.1:%d", ts.GetPort())
}

// handleHealth handles health check requests
func (ts *ToolServer) handleHealth(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"status": "ok",
		"tools":  len(ts.tools),
	})
}

// handleListTools handles tool listing requests
func (ts *ToolServer) handleListTools(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	ts.mu.RLock()
	defer ts.mu.RUnlock()

	toolList := make([]map[string]string, 0, len(ts.tools))
	for name, tool := range ts.tools {
		toolList = append(toolList, map[string]string{
			"name":        name,
			"description": tool.Description(),
		})
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"tools": toolList,
	})
}

// handleCallTool handles tool execution requests
func (ts *ToolServer) handleCallTool(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req ToolRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		ts.logger.Warn("Invalid tool call request: %v", err)
		ts.sendErrorResponse(w, "", nil, fmt.Sprintf("Invalid request: %v", err))
		return
	}

	ts.logger.Debug("Tool call request: %s", req.ToolName)

	ts.mu.RLock()
	tool, exists := ts.tools[req.ToolName]
	ts.mu.RUnlock()

	if !exists {
		ts.logger.Warn("Tool not found: %s", req.ToolName)
		ts.sendErrorResponse(w, req.ToolName, req.Input, fmt.Sprintf("Tool not found: %s", req.ToolName))
		return
	}

	// Convert input to string for tool execution
	inputStr := ""
	switch v := req.Input.(type) {
	case string:
		inputStr = v
	case map[string]interface{}:
		inputBytes, _ := json.Marshal(v)
		inputStr = string(inputBytes)
	default:
		inputBytes, _ := json.Marshal(v)
		inputStr = string(inputBytes)
	}

	ts.logger.Debug("Executing tool %s with input length: %d bytes", req.ToolName, len(inputStr))

	// Execute tool
	ctx, cancel := context.WithTimeout(r.Context(), 30*time.Second)
	defer cancel()

	result, err := tool.Call(ctx, inputStr)
	if err != nil {
		ts.logger.Error("Tool %s execution failed: %v", req.ToolName, err)
		ts.sendErrorResponse(w, req.ToolName, req.Input, fmt.Sprintf("Tool execution failed: %v", err))
		return
	}

	ts.logger.Info("Tool %s executed successfully, result length: %d bytes", req.ToolName, len(result))
	ts.sendSuccessResponse(w, req.ToolName, req.Input, result)
}

// sendSuccessResponse sends a successful tool response
func (ts *ToolServer) sendSuccessResponse(w http.ResponseWriter, toolName string, input interface{}, result string) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(ToolResponse{
		Success: true,
		Result:  result,
		Tool:    toolName,
		Input:   input,
	})
}

// sendErrorResponse sends an error tool response
func (ts *ToolServer) sendErrorResponse(w http.ResponseWriter, toolName string, input interface{}, errorMsg string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusBadRequest)
	json.NewEncoder(w).Encode(ToolResponse{
		Success: false,
		Error:   errorMsg,
		Tool:    toolName,
		Input:   input,
	})
}
