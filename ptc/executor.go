package ptc

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"github.com/smallnest/langgraphgo/log"
	"github.com/tmc/langchaingo/tools"
)

// ExecutionLanguage defines the programming language for code execution
type ExecutionLanguage string

const (
	LanguagePython ExecutionLanguage = "python"
	LanguageGo     ExecutionLanguage = "go"
)

// ExecutionMode defines how tools are executed in the code
type ExecutionMode string

const (
	// ModeServer: All tools are called via HTTP server (alternative)
	// - Server URL exposed to user-generated code
	// - Tools accessed via HTTP calls in Python/Go code
	// - Better isolation (sandboxed)
	// - Reliable tool execution
	ModeServer ExecutionMode = "server"

	// ModeDirect: Hybrid approach for optimal performance (default, recommended)
	// - Shell/Python/File tools: Embedded subprocess execution (true local)
	// - Generic tools: Internal HTTP server (hidden from user code)
	// - Server starts automatically but not exposed to user
	// - Best of both worlds: performance + compatibility
	ModeDirect ExecutionMode = "direct"
)

// CodeExecutor handles the execution of programmatic tool calling code
type CodeExecutor struct {
	Language   ExecutionLanguage
	Tools      []tools.Tool
	Timeout    time.Duration
	WorkDir    string
	Mode       ExecutionMode
	toolServer *ToolServer
	logger     log.Logger // Optional logger for debugging and monitoring
}

// ExecutionResult contains the result of code execution
type ExecutionResult struct {
	Output string
	Error  error
	Stdout string
	Stderr string
}

// NewCodeExecutor creates a new code executor for PTC
// Default mode is ModeDirect for simplicity
func NewCodeExecutor(language ExecutionLanguage, toolList []tools.Tool) *CodeExecutor {
	return NewCodeExecutorWithMode(language, toolList, ModeDirect)
}

// NewCodeExecutorWithMode creates a new code executor with specified execution mode
func NewCodeExecutorWithMode(language ExecutionLanguage, toolList []tools.Tool, mode ExecutionMode) *CodeExecutor {
	executor := &CodeExecutor{
		Language: language,
		Tools:    toolList,
		Timeout:  5 * time.Minute,
		WorkDir:  os.TempDir(),
		Mode:     mode,
		logger:   &log.NoOpLogger{}, // Default to no logging
	}

	// Create tool server for both modes
	// In Direct mode: Internal server for generic tools (shell/python/file use embedded execution)
	// In Server mode: Exposed server for all tools via HTTP
	executor.toolServer = NewToolServer(toolList)

	return executor
}

// SetLogger sets a custom logger for the executor
func (ce *CodeExecutor) SetLogger(logger log.Logger) {
	ce.logger = logger
	if ce.toolServer != nil {
		ce.toolServer.SetLogger(logger)
	}
}

// WithLogger is a fluent method to set a logger
func (ce *CodeExecutor) WithLogger(logger log.Logger) *CodeExecutor {
	ce.SetLogger(logger)
	return ce
}

// Start starts the code executor and its tool server
// In both modes, the server is started for tool access:
// - Direct mode: Internal server for generic tools (not exposed in wrappers)
// - Server mode: Server URL exposed to user code
func (ce *CodeExecutor) Start(ctx context.Context) error {
	if ce.toolServer != nil {
		return ce.toolServer.Start(ctx)
	}
	return nil
}

// Stop stops the code executor and its tool server
func (ce *CodeExecutor) Stop(ctx context.Context) error {
	if ce.toolServer != nil {
		return ce.toolServer.Stop(ctx)
	}
	return nil
}

// GetToolServerURL returns the URL of the tool server
// In Server mode, this URL is exposed to user code
// In Direct mode, returns URL for internal use (not exposed to user)
func (ce *CodeExecutor) GetToolServerURL() string {
	if ce.toolServer != nil {
		return ce.toolServer.GetBaseURL()
	}
	return ""
}

// Execute runs the generated code with access to tools
func (ce *CodeExecutor) Execute(ctx context.Context, code string) (*ExecutionResult, error) {
	ce.logger.Debug("Executing code in %s mode with language %s", ce.Mode, ce.Language)
	ce.logger.Debug("Code length: %d bytes", len(code))

	var result *ExecutionResult
	var err error

	switch ce.Language {
	case LanguagePython:
		result, err = ce.executePython(ctx, code)
	case LanguageGo:
		result, err = ce.executeGo(ctx, code)
	default:
		err = fmt.Errorf("unsupported language: %s", ce.Language)
		ce.logger.Error("Unsupported language: %s", ce.Language)
		return nil, err
	}

	if err != nil {
		ce.logger.Error("Code execution failed: %v", err)
	} else {
		ce.logger.Info("Code execution succeeded, output length: %d bytes", len(result.Output))
	}

	return result, err
}

// executePython executes Python code with tool bindings
func (ce *CodeExecutor) executePython(ctx context.Context, code string) (*ExecutionResult, error) {
	// Create a temporary Python script
	scriptPath := filepath.Join(ce.WorkDir, fmt.Sprintf("ptc_script_%d.py", time.Now().UnixNano()))
	defer os.Remove(scriptPath)

	// Generate Python tool wrapper functions based on execution mode
	var toolWrappers string
	if ce.Mode == ModeServer {
		toolWrappers = ce.generatePythonToolWrappersServer()
	} else {
		toolWrappers = ce.generatePythonToolWrappersDirect()
	}

	// Combine tool wrappers and user code
	fullScript := fmt.Sprintf(`
import json
import sys

# Tool wrapper functions
%s

# User code
%s
`, toolWrappers, code)

	if err := os.WriteFile(scriptPath, []byte(fullScript), 0644); err != nil {
		return nil, fmt.Errorf("failed to write script: %w", err)
	}

	// Execute Python script
	execCtx, cancel := context.WithTimeout(ctx, ce.Timeout)
	defer cancel()

	cmd := exec.CommandContext(execCtx, "python3", scriptPath)
	output, err := cmd.CombinedOutput()

	result := &ExecutionResult{
		Output: string(output),
		Stdout: string(output),
	}

	if err != nil {
		result.Error = err
	}

	return result, nil
}

// executeGo executes Go code with tool bindings
func (ce *CodeExecutor) executeGo(ctx context.Context, code string) (*ExecutionResult, error) {
	// Create a temporary Go file
	scriptPath := filepath.Join(ce.WorkDir, fmt.Sprintf("ptc_script_%d.go", time.Now().UnixNano()))
	defer os.Remove(scriptPath)

	// Generate Go tool wrapper functions based on execution mode
	var toolWrappers string
	if ce.Mode == ModeServer {
		toolWrappers = ce.generateGoToolWrappersServer()
	} else {
		toolWrappers = ce.generateGoToolWrappersDirect()
	}

	// Combine tool wrappers and user code
	fullScript := fmt.Sprintf(`
package main

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"os"
	"os/exec"
	"strings"
)

// Prevent unused import errors
var _ = json.Marshal
var _ = fmt.Println
var _ = strings.Contains
var _ = bytes.NewBuffer
var _ = http.Client{}
var _ = io.ReadAll

// Tool wrapper functions
%s

func main() {
	ctx := context.Background()
	%s
}
`, toolWrappers, code)

	if err := os.WriteFile(scriptPath, []byte(fullScript), 0644); err != nil {
		return nil, fmt.Errorf("failed to write script: %w", err)
	}

	// Execute Go script
	execCtx, cancel := context.WithTimeout(ctx, ce.Timeout)
	defer cancel()

	cmd := exec.CommandContext(execCtx, "go", "run", scriptPath)
	output, err := cmd.CombinedOutput()

	result := &ExecutionResult{
		Output: string(output),
		Stdout: string(output),
	}

	if err != nil {
		result.Error = err
	}

	return result, nil
}

// generatePythonToolWrappersServer creates Python wrapper functions for tools (server mode)
func (ce *CodeExecutor) generatePythonToolWrappersServer() string {
	var wrappers []string

	serverURL := ce.toolServer.GetBaseURL()

	// Create a mapping of tools that can be called via HTTP
	toolsMap := make(map[string]string)
	for _, tool := range ce.Tools {
		toolsMap[tool.Name()] = tool.Description()
	}

	// Serialize tools map for the wrapper
	toolsJSON, _ := json.Marshal(toolsMap)

	wrapper := fmt.Sprintf(`
# Available tools: %s
import json
try:
    import urllib.request
except ImportError:
    import urllib2 as urllib

TOOL_SERVER_URL = "%s"

def call_tool(tool_name, tool_input):
    """Call a tool through the HTTP tool server"""
    try:
        url = TOOL_SERVER_URL + "/call"
        data = json.dumps({
            "tool_name": tool_name,
            "input": tool_input
        }).encode('utf-8')

        req = urllib.request.Request(url, data=data, headers={'Content-Type': 'application/json'})
        response = urllib.request.urlopen(req)
        result = json.loads(response.read().decode('utf-8'))

        if result.get("success"):
            return result.get("result", "")
        else:
            return f"Error calling tool {tool_name}: {result.get('error', 'Unknown error')}"
    except Exception as e:
        return f"Error calling tool {tool_name}: {str(e)}"
`, string(toolsJSON), serverURL)

	wrappers = append(wrappers, wrapper)

	// Generate individual tool functions
	for _, tool := range ce.Tools {
		funcWrapper := fmt.Sprintf(`
def %s(input_data):
    """
    %s
    """
    return call_tool("%s", input_data)
`, sanitizeFunctionName(tool.Name()), tool.Description(), tool.Name())
		wrappers = append(wrappers, funcWrapper)
	}

	return strings.Join(wrappers, "\n")
}

// generatePythonToolWrappersDirect creates Python wrapper functions for tools (direct mode)
// In direct mode, shell/python/file tools are embedded; generic tools use internal server
func (ce *CodeExecutor) generatePythonToolWrappersDirect() string {
	var wrappers []string

	serverURL := ce.toolServer.GetBaseURL()

	// Add common imports and utilities for direct tool execution
	wrapper := fmt.Sprintf(`
# Direct tool execution (embedded tools for shell/python/file, internal server for generic tools)
import subprocess
import json
import os
import tempfile
import sys
try:
    import urllib.request
except ImportError:
    import urllib2 as urllib

INTERNAL_TOOL_SERVER = "%s"

# Helper function to call generic tools via internal server
def _call_generic_tool(tool_name, tool_input):
    """Call a generic tool through the internal tool server"""
    try:
        url = INTERNAL_TOOL_SERVER + "/call"
        data = json.dumps({
            "tool_name": tool_name,
            "input": tool_input
        }).encode('utf-8')

        req = urllib.request.Request(url, data=data, headers={'Content-Type': 'application/json'})
        response = urllib.request.urlopen(req)
        result = json.loads(response.read().decode('utf-8'))

        if result.get("success"):
            return result.get("result", "")
        else:
            return f"Error calling tool {tool_name}: {result.get('error', 'Unknown error')}"
    except Exception as e:
        return f"Error calling tool {tool_name}: {str(e)}"

# Helper function to run shell commands
def _run_shell(code, args=None):
    """Execute shell code directly"""
    try:
        with tempfile.NamedTemporaryFile(mode='w', suffix='.sh', delete=False) as f:
            f.write(code)
            script_path = f.name

        try:
            cmd = ['bash', script_path]
            if args:
                cmd.extend(args)
            result = subprocess.run(cmd, capture_output=True, text=True, timeout=30)
            return result.stdout + result.stderr
        finally:
            os.unlink(script_path)
    except Exception as e:
        return f"Shell execution error: {str(e)}"

# Helper function to run Python code
def _run_python(code, args=None):
    """Execute Python code directly"""
    try:
        with tempfile.NamedTemporaryFile(mode='w', suffix='.py', delete=False) as f:
            f.write(code)
            script_path = f.name

        try:
            python_cmd = sys.executable
            cmd = [python_cmd, script_path]
            if args:
                cmd.extend(args)
            result = subprocess.run(cmd, capture_output=True, text=True, timeout=30)
            return result.stdout + result.stderr
        finally:
            os.unlink(script_path)
    except Exception as e:
        return f"Python execution error: {str(e)}"

# Helper function to read files
def _read_file(file_path):
    """Read file content"""
    try:
        with open(file_path, 'r', encoding='utf-8') as f:
            return f.read()
    except Exception as e:
        return f"File read error: {str(e)}"

# Helper function to write files
def _write_file(file_path, content):
    """Write file content"""
    try:
        with open(file_path, 'w', encoding='utf-8') as f:
            f.write(content)
        return f"Successfully wrote to {file_path}"
    except Exception as e:
        return f"File write error: {str(e)}"
`, serverURL)
	wrappers = append(wrappers, wrapper)

	// Generate embedded tool functions based on tool name patterns
	for _, tool := range ce.Tools {
		funcName := sanitizeFunctionName(tool.Name())
		toolName := tool.Name()

		// Generate appropriate embedded implementation based on tool name
		var funcImpl string

		// Detect tool type and generate embedded implementation
		if strings.Contains(strings.ToLower(toolName), "shell") {
			// Shell execution tool
			funcImpl = fmt.Sprintf(`
def %s(input_data):
    """
    %s
    Direct shell execution (embedded)
    """
    try:
        if isinstance(input_data, str):
            # Simple string input - treat as code
            return _run_shell(input_data)
        elif isinstance(input_data, dict):
            # Structured input
            code = input_data.get('code', input_data.get('command', ''))
            args = input_data.get('args', [])
            if isinstance(args, dict):
                # Template-style args, inject into code
                for key, value in args.items():
                    code = code.replace('{{.%%s}}' %% key, str(value))
                args = []
            return _run_shell(code, args)
        else:
            return _run_shell(str(input_data))
    except Exception as e:
        return f"Error in %s: {str(e)}"
`, funcName, tool.Description(), tool.Name())
		} else if strings.Contains(strings.ToLower(toolName), "python") {
			// Python execution tool
			funcImpl = fmt.Sprintf(`
def %s(input_data):
    """
    %s
    Direct Python execution (embedded)
    """
    try:
        if isinstance(input_data, str):
            return _run_python(input_data)
        elif isinstance(input_data, dict):
            code = input_data.get('code', input_data.get('script', ''))
            args = input_data.get('args', [])
            if isinstance(args, dict):
                for key, value in args.items():
                    code = code.replace('{{.%%s}}' %% key, str(value))
                args = []
            return _run_python(code, args)
        else:
            return _run_python(str(input_data))
    except Exception as e:
        return f"Error in %s: {str(e)}"
`, funcName, tool.Description(), tool.Name())
		} else if strings.Contains(strings.ToLower(toolName), "read") && strings.Contains(strings.ToLower(toolName), "file") {
			// File read tool
			funcImpl = fmt.Sprintf(`
def %s(input_data):
    """
    %s
    Direct file reading (embedded)
    """
    try:
        if isinstance(input_data, str):
            return _read_file(input_data)
        elif isinstance(input_data, dict):
            file_path = input_data.get('filePath', input_data.get('file_path', input_data.get('path', '')))
            return _read_file(file_path)
        else:
            return _read_file(str(input_data))
    except Exception as e:
        return f"Error in %s: {str(e)}"
`, funcName, tool.Description(), tool.Name())
		} else if strings.Contains(strings.ToLower(toolName), "write") && strings.Contains(strings.ToLower(toolName), "file") {
			// File write tool
			funcImpl = fmt.Sprintf(`
def %s(input_data):
    """
    %s
    Direct file writing (embedded)
    """
    try:
        if isinstance(input_data, dict):
            file_path = input_data.get('filePath', input_data.get('file_path', input_data.get('path', '')))
            content = input_data.get('content', '')
            return _write_file(file_path, content)
        else:
            return "Error: write_file requires dict with 'filePath' and 'content'"
    except Exception as e:
        return f"Error in %s: {str(e)}"
`, funcName, tool.Description(), tool.Name())
		} else {
			// Generic tool - call via internal tool server
			funcImpl = fmt.Sprintf(`
def %s(input_data):
    """
    %s
    Generic tool called via internal server.
    """
    # Convert input to JSON string if it's a dict
    if isinstance(input_data, dict):
        input_str = json.dumps(input_data)
    else:
        input_str = str(input_data)

    return _call_generic_tool("%s", input_str)
`, funcName, tool.Description(), tool.Name())
		}

		wrappers = append(wrappers, funcImpl)
	}

	return strings.Join(wrappers, "\n")
}

// generateGoToolWrappersServer creates Go wrapper functions for tools (server mode)
func (ce *CodeExecutor) generateGoToolWrappersServer() string {
	var wrappers []string

	serverURL := ce.toolServer.GetBaseURL()

	// Create the call_tool function
	wrapper := fmt.Sprintf(`
import (
	"bytes"
	"io"
	"net/http"
)

const toolServerURL = "%s"

// callTool calls a tool through the HTTP tool server
func callTool(ctx context.Context, toolName string, toolInput interface{}) (string, error) {
	requestBody := map[string]interface{}{
		"tool_name": toolName,
		"input":     toolInput,
	}

	jsonData, err := json.Marshal(requestBody)
	if err != nil {
		return "", fmt.Errorf("failed to marshal request: %%w", err)
	}

	req, err := http.NewRequestWithContext(ctx, "POST", toolServerURL+"/call", bytes.NewBuffer(jsonData))
	if err != nil {
		return "", fmt.Errorf("failed to create request: %%w", err)
	}
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return "", fmt.Errorf("failed to call tool: %%w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("failed to read response: %%w", err)
	}

	var result map[string]interface{}
	if err := json.Unmarshal(body, &result); err != nil {
		return "", fmt.Errorf("failed to unmarshal response: %%w", err)
	}

	if success, ok := result["success"].(bool); ok && success {
		if resultStr, ok := result["result"].(string); ok {
			return resultStr, nil
		}
	}

	errorMsg := "unknown error"
	if errStr, ok := result["error"].(string); ok {
		errorMsg = errStr
	}
	return "", fmt.Errorf("tool execution failed: %%s", errorMsg)
}
`, serverURL)
	wrappers = append(wrappers, wrapper)

	// Generate individual tool functions
	for _, tool := range ce.Tools {
		funcWrapper := fmt.Sprintf(`
// %s: %s
func %s(ctx context.Context, input string) (string, error) {
	return callTool(ctx, "%s", input)
}
`, tool.Name(), tool.Description(), sanitizeFunctionName(tool.Name()), tool.Name())
		wrappers = append(wrappers, funcWrapper)
	}

	return strings.Join(wrappers, "\n")
}

// generateGoToolWrappersDirect creates Go wrapper functions for tools (direct mode)
// In direct mode, shell/python/file tools are embedded; generic tools use internal server
func (ce *CodeExecutor) generateGoToolWrappersDirect() string {
	var wrappers []string

	serverURL := ce.toolServer.GetBaseURL()

	// Add common helper functions for direct tool execution
	// (imports are in the main template)
	wrapper := fmt.Sprintf(`
// Internal tool server URL for generic tools
const internalToolServer = "%s"

// Helper function to call generic tools via internal server
func callGenericTool(ctx context.Context, toolName string, input string) (string, error) {
	requestBody := map[string]interface{}{
		"tool_name": toolName,
		"input":     input,
	}

	jsonData, err := json.Marshal(requestBody)
	if err != nil {
		return "", fmt.Errorf("failed to marshal request: %%w", err)
	}

	req, err := http.NewRequestWithContext(ctx, "POST", internalToolServer+"/call", bytes.NewBuffer(jsonData))
	if err != nil {
		return "", fmt.Errorf("failed to create request: %%w", err)
	}
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return "", fmt.Errorf("failed to call tool: %%w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("failed to read response: %%w", err)
	}

	var result map[string]interface{}
	if err := json.Unmarshal(body, &result); err != nil {
		return "", fmt.Errorf("failed to unmarshal response: %%w", err)
	}

	if success, ok := result["success"].(bool); ok && success {
		if resultStr, ok := result["result"].(string); ok {
			return resultStr, nil
		}
	}

	errorMsg := "unknown error"
	if errStr, ok := result["error"].(string); ok {
		errorMsg = errStr
	}
	return "", fmt.Errorf("tool execution failed: %%s", errorMsg)
}

// Helper function to run shell commands
func runShell(ctx context.Context, code string, args []string) (string, error) {
	tmpfile, err := ioutil.TempFile("", "shell-*.sh")
	if err != nil {
		return "", err
	}
	defer os.Remove(tmpfile.Name())

	if _, err := tmpfile.Write([]byte(code)); err != nil {
		return "", err
	}
	if err := tmpfile.Close(); err != nil {
		return "", err
	}

	cmd := exec.CommandContext(ctx, "bash", append([]string{tmpfile.Name()}, args...)...)
	output, err := cmd.CombinedOutput()
	return string(output), err
}

// Helper function to run Python scripts
func runPython(ctx context.Context, code string, args []string) (string, error) {
	tmpfile, err := ioutil.TempFile("", "python-*.py")
	if err != nil {
		return "", err
	}
	defer os.Remove(tmpfile.Name())

	if _, err := tmpfile.Write([]byte(code)); err != nil {
		return "", err
	}
	if err := tmpfile.Close(); err != nil {
		return "", err
	}

	pythonCmd := "python3"
	if _, err := exec.LookPath("python3"); err != nil {
		pythonCmd = "python"
	}

	cmd := exec.CommandContext(ctx, pythonCmd, append([]string{tmpfile.Name()}, args...)...)
	output, err := cmd.CombinedOutput()
	return string(output), err
}

// Helper function to read files
func readFile(filePath string) (string, error) {
	data, err := ioutil.ReadFile(filePath)
	if err != nil {
		return "", err
	}
	return string(data), nil
}

// Helper function to write files
func writeFile(filePath string, content string) (string, error) {
	err := ioutil.WriteFile(filePath, []byte(content), 0644)
	if err != nil {
		return "", err
	}
	return fmt.Sprintf("Successfully wrote to %%s", filePath), nil
}
`, serverURL)
	wrappers = append(wrappers, wrapper)

	// Generate embedded tool functions based on tool name patterns
	for _, tool := range ce.Tools {
		funcName := sanitizeFunctionName(tool.Name())
		toolName := tool.Name()

		// Generate appropriate embedded implementation based on tool name
		var funcImpl string

		// Detect tool type and generate embedded implementation
		if strings.Contains(strings.ToLower(toolName), "shell") {
			// Shell execution tool
			funcImpl = fmt.Sprintf(`
// %s: %s (Direct shell execution - embedded)
func %s(ctx context.Context, input string) (string, error) {
	// Parse input as JSON if possible
	var params map[string]interface{}
	if err := json.Unmarshal([]byte(input), &params); err == nil {
		// Structured input
		code := ""
		if codeVal, ok := params["code"]; ok {
			code = fmt.Sprintf("%%v", codeVal)
		} else if cmdVal, ok := params["command"]; ok {
			code = fmt.Sprintf("%%v", cmdVal)
		}

		args := []string{}
		if argsVal, ok := params["args"]; ok {
			if argsList, ok := argsVal.([]interface{}); ok {
				for _, arg := range argsList {
					args = append(args, fmt.Sprintf("%%v", arg))
				}
			}
		}

		return runShell(ctx, code, args)
	}

	// Simple string input - treat as shell code
	return runShell(ctx, input, nil)
}`, tool.Name(), tool.Description(), funcName)
		} else if strings.Contains(strings.ToLower(toolName), "python") {
			// Python execution tool
			funcImpl = fmt.Sprintf(`
// %s: %s (Direct Python execution - embedded)
func %s(ctx context.Context, input string) (string, error) {
	var params map[string]interface{}
	if err := json.Unmarshal([]byte(input), &params); err == nil {
		code := ""
		if codeVal, ok := params["code"]; ok {
			code = fmt.Sprintf("%%v", codeVal)
		} else if scriptVal, ok := params["script"]; ok {
			code = fmt.Sprintf("%%v", scriptVal)
		}

		args := []string{}
		if argsVal, ok := params["args"]; ok {
			if argsList, ok := argsVal.([]interface{}); ok {
				for _, arg := range argsList {
					args = append(args, fmt.Sprintf("%%v", arg))
				}
			}
		}

		return runPython(ctx, code, args)
	}

	return runPython(ctx, input, nil)
}`, tool.Name(), tool.Description(), funcName)
		} else if strings.Contains(strings.ToLower(toolName), "read") && strings.Contains(strings.ToLower(toolName), "file") {
			// File read tool
			funcImpl = fmt.Sprintf(`
// %s: %s (Direct file reading - embedded)
func %s(ctx context.Context, input string) (string, error) {
	var params map[string]interface{}
	if err := json.Unmarshal([]byte(input), &params); err == nil {
		if filePathVal, ok := params["filePath"]; ok {
			return readFile(fmt.Sprintf("%%v", filePathVal))
		}
		if filePathVal, ok := params["file_path"]; ok {
			return readFile(fmt.Sprintf("%%v", filePathVal))
		}
		if pathVal, ok := params["path"]; ok {
			return readFile(fmt.Sprintf("%%v", pathVal))
		}
	}

	return readFile(input)
}`, tool.Name(), tool.Description(), funcName)
		} else if strings.Contains(strings.ToLower(toolName), "write") && strings.Contains(strings.ToLower(toolName), "file") {
			// File write tool
			funcImpl = fmt.Sprintf(`
// %s: %s (Direct file writing - embedded)
func %s(ctx context.Context, input string) (string, error) {
	var params map[string]interface{}
	if err := json.Unmarshal([]byte(input), &params); err != nil {
		return "", fmt.Errorf("write_file requires JSON input with filePath and content")
	}

	filePath := ""
	if val, ok := params["filePath"]; ok {
		filePath = fmt.Sprintf("%%v", val)
	} else if val, ok := params["file_path"]; ok {
		filePath = fmt.Sprintf("%%v", val)
	} else if val, ok := params["path"]; ok {
		filePath = fmt.Sprintf("%%v", val)
	}

	content := ""
	if val, ok := params["content"]; ok {
		content = fmt.Sprintf("%%v", val)
	}

	if filePath == "" {
		return "", fmt.Errorf("filePath is required")
	}

	return writeFile(filePath, content)
}`, tool.Name(), tool.Description(), funcName)
		} else {
			// Generic tool - call via internal tool server
			funcImpl = fmt.Sprintf(`
// %s: %s (Generic tool called via internal server)
func %s(ctx context.Context, input string) (string, error) {
	return callGenericTool(ctx, "%s", input)
}`, tool.Name(), tool.Description(), funcName, tool.Name())
		}

		wrappers = append(wrappers, funcImpl)
	}

	return strings.Join(wrappers, "\n")
}

// createToolHelperProgram creates a helper executable for direct tool execution
func (ce *CodeExecutor) createToolHelperProgram() string {
	// Create a temporary Go program that can execute tools
	helperPath := filepath.Join(ce.WorkDir, fmt.Sprintf("tool_helper_%d", time.Now().UnixNano()))

	// Generate Go source code for the helper
	helperSource := ce.generateHelperSource()

	sourcePath := helperPath + ".go"
	if err := os.WriteFile(sourcePath, []byte(helperSource), 0644); err != nil {
		// If we can't create the helper, return empty path
		// The calling code will handle the error
		return ""
	}

	// Compile the helper program
	cmd := exec.Command("go", "build", "-o", helperPath, sourcePath)
	if err := cmd.Run(); err != nil {
		// Compilation failed, clean up and return empty
		os.Remove(sourcePath)
		return ""
	}

	// Clean up source file
	os.Remove(sourcePath)

	return helperPath
}

// generateHelperSource generates the Go source code for the tool helper program
func (ce *CodeExecutor) generateHelperSource() string {
	serverURL := ce.GetToolServerURL()

	// Build tool call implementations
	var toolCases []string
	for _, tool := range ce.Tools {
		toolCase := fmt.Sprintf(`	case "%s":
		result, err = tool_%s(ctx, req.Input)`,
			tool.Name(),
			sanitizeFunctionName(tool.Name()))
		toolCases = append(toolCases, toolCase)
	}

	// Build tool function implementations that call the tool server
	var toolFuncs []string
	for _, tool := range ce.Tools {
		toolFunc := fmt.Sprintf(`
func tool_%s(ctx context.Context, input string) (string, error) {
	// Call tool via internal server: %s
	return callToolServer(ctx, "%s", input)
}`,
			sanitizeFunctionName(tool.Name()),
			tool.Description(),
			tool.Name())
		toolFuncs = append(toolFuncs, toolFunc)
	}

	source := fmt.Sprintf(`package main

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
)

const toolServerURL = "%s"

type Request struct {
	ToolName string ` + "`json:\"tool_name\"`" + `
	Input    string ` + "`json:\"input\"`" + `
}

type Response struct {
	Success bool   ` + "`json:\"success\"`" + `
	Result  string ` + "`json:\"result,omitempty\"`" + `
	Error   string ` + "`json:\"error,omitempty\"`" + `
}

type ToolCallRequest struct {
	ToolName string      ` + "`json:\"tool_name\"`" + `
	Input    interface{} ` + "`json:\"input\"`" + `
}

type ToolCallResponse struct {
	Success bool   ` + "`json:\"success\"`" + `
	Result  string ` + "`json:\"result,omitempty\"`" + `
	Error   string ` + "`json:\"error,omitempty\"`" + `
}

// callToolServer calls a tool through the internal tool server
func callToolServer(ctx context.Context, toolName string, input string) (string, error) {
	requestBody := ToolCallRequest{
		ToolName: toolName,
		Input:    input,
	}

	jsonData, err := json.Marshal(requestBody)
	if err != nil {
		return "", fmt.Errorf("failed to marshal request: %%w", err)
	}

	req, err := http.NewRequestWithContext(ctx, "POST", toolServerURL+"/call", bytes.NewBuffer(jsonData))
	if err != nil {
		return "", fmt.Errorf("failed to create request: %%w", err)
	}
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return "", fmt.Errorf("failed to call tool server: %%w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("failed to read response: %%w", err)
	}

	var result ToolCallResponse
	if err := json.Unmarshal(body, &result); err != nil {
		return "", fmt.Errorf("failed to unmarshal response: %%w", err)
	}

	if !result.Success {
		return "", fmt.Errorf("tool execution failed: %%s", result.Error)
	}

	return result.Result, nil
}

%s

func main() {
	if len(os.Args) < 2 {
		respondError("missing request argument")
		return
	}

	var req Request
	if err := json.Unmarshal([]byte(os.Args[1]), &req); err != nil {
		respondError("invalid request: " + err.Error())
		return
	}

	ctx := context.Background()
	var result string
	var err error

	switch req.ToolName {
%s
	default:
		respondError("unknown tool: " + req.ToolName)
		return
	}

	if err != nil {
		respondError(err.Error())
		return
	}

	respondSuccess(result)
}

func respondSuccess(result string) {
	resp := Response{
		Success: true,
		Result:  result,
	}
	json.NewEncoder(os.Stdout).Encode(resp)
}

func respondError(errMsg string) {
	resp := Response{
		Success: false,
		Error:   errMsg,
	}
	json.NewEncoder(os.Stdout).Encode(resp)
}
`, serverURL, strings.Join(toolFuncs, "\n"), strings.Join(toolCases, "\n"))

	return source
}

// sanitizeFunctionName converts a tool name to a valid function name
func sanitizeFunctionName(name string) string {
	// Replace invalid characters with underscores
	name = strings.ReplaceAll(name, "-", "_")
	name = strings.ReplaceAll(name, " ", "_")
	name = strings.ReplaceAll(name, ".", "_")

	// Ensure it starts with a letter
	if len(name) > 0 && name[0] >= '0' && name[0] <= '9' {
		name = "tool_" + name
	}

	return name
}

// GetToolDefinitions returns tool definitions for LLM prompting
func (ce *CodeExecutor) GetToolDefinitions() string {
	var defs []string

	defs = append(defs, "# Available Tools\n")
	defs = append(defs, "You have access to the following tools that you can call in your code:\n")

	for _, tool := range ce.Tools {
		def := fmt.Sprintf("\n## %s\n", tool.Name())
		def += fmt.Sprintf("Description: %s\n", tool.Description())
		def += fmt.Sprintf("Usage: %s(input_string)\n", sanitizeFunctionName(tool.Name()))
		defs = append(defs, def)
	}

	return strings.Join(defs, "")
}
