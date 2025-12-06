# DeepAgents - Filesystem-Aware AI Agent

A Go implementation of an intelligent agent with filesystem access and task management capabilities, built using [langgraphgo](https://github.com/smallnest/langgraphgo) and [langchaingo](https://github.com/tmc/langchaingo).

DeepAgents provides a powerful agent framework that can interact with the filesystem, manage tasks, and delegate work to subagents, making it ideal for automation, file processing, and complex task orchestration.

## Overview

DeepAgents is an AI agent that can:
- **Read and write files**: Full filesystem access within a configured workspace
- **Manage tasks**: Built-in todo list management
- **Delegate work**: Spawn subagents to handle complex subtasks
- **Search files**: Pattern-based file finding with glob support
- **Execute autonomously**: Uses LLM-powered reasoning to complete tasks

## Architecture

```
┌─────────────────────────────────────────────────────────────┐
│                       DeepAgents                            │
├─────────────────────────────────────────────────────────────┤
│                                                             │
│  ┌──────────────┐      ┌──────────────┐      ┌──────────┐ │
│  │  Filesystem  │      │     Todo     │      │ SubAgent │ │
│  │    Tools     │      │   Manager    │      │  System  │ │
│  └──────────────┘      └──────────────┘      └──────────┘ │
│        │                      │                     │      │
│        ▼                      ▼                     ▼      │
│   • ls                   • write_todos          • task    │
│   • read_file            • read_todos                     │
│   • write_file                                            │
│   • glob                                                  │
│                                                             │
│  ┌──────────────────────────────────────────────────────┐ │
│  │           LangGraph Agent (prebuilt)                 │ │
│  │         LLM-powered reasoning and tool use           │ │
│  └──────────────────────────────────────────────────────┘ │
│                                                             │
└─────────────────────────────────────────────────────────────┘
```

## Features

### 🗂️ Filesystem Tools
- **ls**: List files and directories with size information
- **read_file**: Read file contents
- **write_file**: Create or update files
- **glob**: Find files matching patterns

### ✅ Task Management
- **write_todos**: Create and update todo lists
- **read_todos**: Retrieve current tasks
- **Thread-safe**: Concurrent access support with mutex locks

### 🤖 SubAgent System
- **task**: Delegate complex tasks to specialized subagents
- **Configurable handlers**: Custom subagent logic
- **Hierarchical task decomposition**: Break down complex problems

### 🔧 Flexible Configuration
- **Custom workspace**: Configurable root directory
- **System prompts**: Tailor agent behavior
- **Tool integration**: Easy to extend with new tools

## Prerequisites

- **Go**: Version 1.21 or higher
- **API Key**: OpenAI-compatible API (OpenAI, DeepSeek, etc.)

## Installation

```bash
# Navigate to the deepagents directory
cd showcases/deepagents

# Set up environment variables
export OPENAI_API_KEY="your-api-key-here"

# Optional: If using DeepSeek or another provider
export OPENAI_API_BASE="https://api.deepseek.com/v1"

# Run the example
go run main.go
```

## Usage

### Basic Example

```go
package main

import (
    "context"
    "log"

    "github.com/smallnest/langgraphgo/showcases/deepagents/agent"
    "github.com/tmc/langchaingo/llms"
    "github.com/tmc/langchaingo/llms/openai"
)

func main() {
    ctx := context.Background()

    // Initialize LLM
    model, err := openai.New()
    if err != nil {
        log.Fatalf("Failed to create LLM: %v", err)
    }

    // Create Deep Agent
    deepAgent, err := agent.CreateDeepAgent(model,
        agent.WithRootDir("./workspace"),
        agent.WithSystemPrompt("You are a capable assistant with filesystem access."),
    )
    if err != nil {
        log.Fatalf("Failed to create deep agent: %v", err)
    }

    // Run the agent
    inputs := map[string]interface{}{
        "messages": []llms.MessageContent{
            llms.TextParts(llms.ChatMessageTypeHuman,
                "Create a file named 'hello.txt' with content 'Hello, DeepAgents!', then read it back."),
        },
    }

    result, err := deepAgent.Invoke(ctx, inputs)
    if err != nil {
        log.Fatalf("Agent execution failed: %v", err)
    }

    // Process result
    // ...
}
```

### With SubAgent Handler

```go
// Define a custom subagent handler
subAgentHandler := func(ctx context.Context, task string) (string, error) {
    log.Printf("[SubAgent] Processing task: %s", task)

    // Implement custom logic, spawn another agent, etc.
    // For example, delegate to a specialized research agent

    return "Task completed: " + task, nil
}

// Create agent with subagent support
deepAgent, err := agent.CreateDeepAgent(model,
    agent.WithRootDir("./workspace"),
    agent.WithSubAgentHandler(subAgentHandler),
)
```

## Configuration Options

### CreateDeepAgent Options

```go
type DeepAgentOptions struct {
    RootDir         string              // Workspace root directory
    SystemPrompt    string              // Agent system prompt
    SubAgentHandler SubAgentHandler     // Handler for delegated tasks
}
```

**Available Options**:
- `WithRootDir(path string)`: Set the workspace directory (default: ".")
- `WithSystemPrompt(prompt string)`: Customize agent behavior (default: "You are a helpful deep agent.")
- `WithSubAgentHandler(handler func)`: Enable task delegation to subagents

### Environment Variables

| Variable | Description | Default | Required |
|----------|-------------|---------|----------|
| `OPENAI_API_KEY` | OpenAI API key | None | ✅ Yes |
| `OPENAI_API_BASE` | API base URL | OpenAI default | ❌ No |

## Tools Reference

### Filesystem Tools

#### ls
**Description**: List files in a directory
**Input**: Directory path (relative to root)
**Output**: Formatted list with type (D/F), name, and size

```
D subdir 0
F file.txt 1234
```

#### read_file
**Description**: Read file contents
**Input**: File path (relative to root)
**Output**: File content as string

#### write_file
**Description**: Write to a file
**Input**: Path and content separated by newline
```
filename.txt
File content goes here
```
**Output**: Success confirmation

#### glob
**Description**: Find files matching a pattern
**Input**: Glob pattern (e.g., `*.txt`, `**/*.go`)
**Output**: Newline-separated list of matching files

### Task Management Tools

#### write_todos
**Description**: Create or update todo list
**Input**: Newline-separated list of tasks
```
Task 1
Task 2
Task 3
```
**Output**: Success confirmation

#### read_todos
**Description**: Read current todo list
**Input**: Empty (not used)
**Output**: Newline-separated list of todos

### SubAgent Tools

#### task
**Description**: Delegate a task to a subagent
**Input**: Task description
**Output**: Result from subagent handler

**Example Use Cases**:
- Spawn a specialized research agent for complex queries
- Delegate file processing to a batch processor
- Create hierarchical task workflows

## Project Structure

```
deepagents/
├── main.go                 # Example application
├── agent/
│   └── agent.go           # Agent creation and configuration
└── tools/
    ├── filesystem.go      # File operations (ls, read, write, glob)
    ├── todo.go            # Task management (write_todos, read_todos)
    └── subagent.go        # Subagent delegation (task)
```

## How It Works

### 1. Agent Initialization

```go
// The CreateDeepAgent function:
// 1. Creates workspace directory if needed
// 2. Initializes TodoManager
// 3. Creates all tools with configured options
// 4. Builds agent using prebuilt.CreateAgent
```

### 2. Tool Registration

All tools implement the `Tool` interface:
```go
type Tool interface {
    Name() string
    Description() string
    Call(ctx context.Context, input string) (string, error)
}
```

Tools are registered with the agent and exposed to the LLM for function calling.

### 3. Agent Execution

The agent uses LangGraph's prebuilt agent pattern:
1. **Receive message**: User provides natural language instruction
2. **Reason**: LLM decides which tools to use
3. **Act**: Execute tool calls
4. **Loop**: Continue until task is complete
5. **Respond**: Return final result

### 4. SubAgent Delegation

When the agent encounters a complex task:
1. Agent calls the `task` tool with task description
2. SubAgentHandler receives the task
3. Handler can spawn new agents, make API calls, etc.
4. Result is returned to the main agent
5. Main agent continues with the result

## Example Use Cases

### 1. File Organization

```go
inputs := map[string]interface{}{
    "messages": []llms.MessageContent{
        llms.TextParts(llms.ChatMessageTypeHuman,
            "Organize all .txt files into a 'docs' directory"),
    },
}
```

**Agent Actions**:
1. Uses `glob` to find all .txt files
2. Creates docs directory
3. Uses `read_file` and `write_file` to move files
4. Confirms completion

### 2. Task Management

```go
inputs := map[string]interface{}{
    "messages": []llms.MessageContent{
        llms.TextParts(llms.ChatMessageTypeHuman,
            "Create a todo list for setting up a new Go project"),
    },
}
```

**Agent Actions**:
1. Generates todo list
2. Uses `write_todos` to save tasks
3. Can later use `read_todos` to track progress

### 3. Batch Processing

```go
subAgentHandler := func(ctx context.Context, task string) (string, error) {
    // Process individual file
    return processFile(task)
}

inputs := map[string]interface{}{
    "messages": []llms.MessageContent{
        llms.TextParts(llms.ChatMessageTypeHuman,
            "Process all JSON files in the data directory"),
    },
}
```

**Agent Actions**:
1. Uses `glob` to find JSON files
2. Delegates each file to subagent via `task` tool
3. Aggregates results

### 4. Code Generation

```go
inputs := map[string]interface{}{
    "messages": []llms.MessageContent{
        llms.TextParts(llms.ChatMessageTypeHuman,
            "Create a simple Go HTTP server in server.go"),
    },
}
```

**Agent Actions**:
1. Generates code content
2. Uses `write_file` to create server.go
3. Confirms creation

## Best Practices

### 1. Workspace Isolation

✅ **Good**:
```go
agent.WithRootDir("./workspace")  // Isolated workspace
```

❌ **Avoid**:
```go
agent.WithRootDir("/")  // System-wide access (security risk)
```

### 2. Clear System Prompts

✅ **Good**:
```go
agent.WithSystemPrompt("You are a file organizer. Create directories when needed and move files systematically.")
```

❌ **Too Vague**:
```go
agent.WithSystemPrompt("You are helpful.")
```

### 3. Structured SubAgent Handlers

```go
subAgentHandler := func(ctx context.Context, task string) (string, error) {
    // Parse task
    taskType, params := parseTask(task)

    // Route to specialized handler
    switch taskType {
    case "research":
        return researchAgent.Handle(ctx, params)
    case "analyze":
        return analysisAgent.Handle(ctx, params)
    default:
        return "", fmt.Errorf("unknown task type: %s", taskType)
    }
}
```

### 4. Error Handling

```go
result, err := deepAgent.Invoke(ctx, inputs)
if err != nil {
    log.Printf("Agent failed: %v", err)
    // Implement retry logic or fallback
}
```

## Advanced Usage

### Custom Tool Integration

Add new tools by implementing the `Tool` interface:

```go
type CustomTool struct {
    // Configuration
}

func (t *CustomTool) Name() string {
    return "custom_tool"
}

func (t *CustomTool) Description() string {
    return "Description for the LLM"
}

func (t *CustomTool) Call(ctx context.Context, input string) (string, error) {
    // Implementation
    return result, nil
}
```

Then register with the agent:
```go
agentTools := []ltools.Tool{
    // ... existing tools
    &CustomTool{},
}

agent, err := prebuilt.CreateAgent(model, agentTools, ...)
```

### Hierarchical Agent Systems

```go
// Create specialized agents
researchAgent := createResearchAgent(model)
analysisAgent := createAnalysisAgent(model)

// Main agent delegates to specialized agents
mainHandler := func(ctx context.Context, task string) (string, error) {
    if strings.Contains(task, "research") {
        return researchAgent.Invoke(ctx, taskInputs)
    }
    if strings.Contains(task, "analyze") {
        return analysisAgent.Invoke(ctx, taskInputs)
    }
    return "", fmt.Errorf("unknown task type")
}

mainAgent := agent.CreateDeepAgent(model,
    agent.WithSubAgentHandler(mainHandler),
)
```

### Persistent Todo Management

```go
// Save todos to file
todoManager := tools.NewTodoManager()

// Load from file on startup
if data, err := os.ReadFile("todos.txt"); err == nil {
    // Initialize from file
}

// Save to file after updates
// (implement in WriteTodosTool.Call)
```

## Troubleshooting

### API Key Not Set

```
OPENAI_API_KEY not set, skipping example execution
```

**Solution**:
```bash
export OPENAI_API_KEY="sk-..."
```

### Permission Denied

If filesystem operations fail:
- Check workspace directory permissions
- Ensure root directory exists and is writable
- Verify user has access to the workspace

### Tool Call Failures

If tools are not being called correctly:
- Check tool descriptions are clear
- Verify LLM model supports function calling
- Review system prompt for clarity
- Enable verbose logging to see LLM reasoning

### SubAgent Not Called

If task delegation doesn't work:
- Verify SubAgentHandler is configured
- Check task descriptions are specific
- Ensure handler is not nil
- Review error logs from handler

## Performance Considerations

### Filesystem Operations

- **Read file**: Fast for small files (<1MB)
- **Write file**: Atomic operations, safe for concurrent access
- **Glob**: Performance depends on directory size
- **Ls**: Efficient for reasonable directory sizes

### LLM Calls

Each agent invocation may involve:
- 1-5 LLM calls (depending on task complexity)
- Tool calls are executed sequentially
- Consider using faster models (gpt-3.5-turbo) for simple tasks

### Optimization Tips

1. **Use specific prompts**: Reduce reasoning iterations
2. **Batch operations**: Group file operations when possible
3. **Cache results**: Implement caching in SubAgentHandler
4. **Limit tool set**: Only include necessary tools

## Security Considerations

### Workspace Isolation

Always use a dedicated workspace directory:
```go
agent.WithRootDir("./safe_workspace")
```

### Input Validation

The agent executes LLM-generated tool calls. To enhance security:
- Use a restricted workspace
- Implement tool input validation
- Review generated code before execution
- Monitor filesystem operations

### API Key Protection

```bash
# Use environment variables
export OPENAI_API_KEY="..."

# Don't hardcode in source
# ❌ apiKey := "sk-..."
```

## Future Enhancements

Planned features:
- [ ] File watching and event-driven execution
- [ ] Database integration tools
- [ ] HTTP request tools
- [ ] Git operations support
- [ ] Archive/compression tools
- [ ] Image processing tools
- [ ] Persistent state management
- [ ] Multi-agent collaboration
- [ ] Workflow templates

## License

MIT License - Same as the parent langgraphgo project

## References

- [LangGraph Go](https://github.com/smallnest/langgraphgo) - Graph-based agent framework
- [LangChain Go](https://github.com/tmc/langchaingo) - LLM integration library
- [OpenAI Function Calling](https://platform.openai.com/docs/guides/function-calling) - Tool use documentation

## Contributing

Contributions are welcome! Areas for improvement:
- Additional filesystem tools
- Enhanced error handling
- Tool input validation
- Performance optimizations
- Documentation improvements
- Example applications

## Support

For issues and questions:
- Review this README
- Check the example in main.go
- Open an issue on the langgraphgo GitHub repository

---

**Built with**:
- [langgraphgo](https://github.com/smallnest/langgraphgo) - Agent orchestration
- [langchaingo](https://github.com/tmc/langchaingo) - LLM integration
- Standard Go libraries for filesystem operations
