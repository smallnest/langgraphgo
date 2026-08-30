# Planning Agent Example

This example demonstrates how to use the **Planning Agent** - an intelligent agent that dynamically creates workflow plans based on user requests using LLM reasoning.

## 1. Background

Traditional agents follow predefined workflows. The **Planning Agent** is different:
1. **Analyzes** the user's request
2. **Plans** an optimal workflow by selecting and ordering available nodes
3. **Executes** the dynamically generated plan

This approach provides:
- **Flexibility**: The workflow adapts to different user requests
- **Intelligence**: LLM determines the best sequence of operations
- **Efficiency**: Only executes necessary steps

## 2. Key Concepts

- **Available Nodes**: A collection of predefined operations (nodes) that can be composed into workflows
- **Planner Node**: Uses LLM to generate a workflow plan in JSON format based on the user request
- **Executor Node**: Dynamically builds and executes the planned workflow
- **Workflow Plan**: A structured JSON describing nodes and edges (similar to a Mermaid diagram)

## 3. How It Works

### Step 1: Define Available Nodes
```go
nodes := []*graph.Node{
    {
        Name:        "fetch_data",
        Description: "Fetch user data from the database",
        Function:    fetchDataNode,
    },
    {
        Name:        "validate_data",
        Description: "Validate the integrity and format of the data",
        Function:    validateDataNode,
    },
    // ... more nodes
}
```

### Step 2: Create the Planning Agent
```go
agent, err := prebuilt.CreatePlanningAgentMap(
    model,
    nodes,
    []tools.Tool{},
    prebuilt.WithVerbose(true), // Optional: show detailed logs
)
```

### Step 3: Execute with User Request
```go
query := "Fetch user data, validate it, and save the results"
initialState := map[string]any{
    "messages": []llms.MessageContent{
        llms.TextParts(llms.ChatMessageTypeHuman, query),
    },
}
res, err := agent.Invoke(context.Background(), initialState)
```

## 4. Workflow Plan Format

The LLM generates a plan in this JSON format:
```json
{
  "nodes": [
    {"name": "fetch_data", "type": "process"},
    {"name": "validate_data", "type": "process"},
    {"name": "save_results", "type": "process"}
  ],
  "edges": [
    {"from": "START", "to": "fetch_data"},
    {"from": "fetch_data", "to": "validate_data"},
    {"from": "validate_data", "to": "save_results"},
    {"from": "save_results", "to": "END"}
  ]
}
```

This creates a workflow: `START → fetch_data → validate_data → save_results → END`

## 5. Example Scenarios

### Scenario 1: Data Processing
**Request**: "Fetch user data, validate it, transform it to JSON, and save the results"

**Generated Plan**:
```
START → fetch_data → validate_data → transform_data → save_results → END
```

### Scenario 2: Data Analysis
**Request**: "Fetch data, analyze it, and generate a report"

**Generated Plan**:
```
START → fetch_data → analyze_data → generate_report → END
```

### Scenario 3: Complete Pipeline
**Request**: "Fetch data, validate and transform it, analyze the results, and generate a comprehensive report"

**Generated Plan**:
```
START → fetch_data → validate_data → transform_data → analyze_data → generate_report → END
```

## 6. Code Highlights

### Defining a Node
```go
func fetchDataNode(ctx context.Context, state any) (any, error) {
    mState := state.(map[string]any)
    messages := mState["messages"].([]llms.MessageContent)

    // Your business logic here
    fmt.Println("📥 Fetching data from database...")

    msg := llms.MessageContent{
        Role:  llms.ChatMessageTypeAI,
        Parts: []llms.ContentPart{llms.TextPart("Data fetched successfully")},
    }

    return map[string]any{
        "messages": append(messages, msg),
    }, nil
}
```

### Verbose Output
When `WithVerbose(true)` is enabled, you'll see:
```
🤔 Planning workflow...
📋 Generated plan:
{
  "nodes": [...],
  "edges": [...]
}

🚀 Executing planned workflow...
  ✓ Added node: fetch_data
  ✓ Added node: validate_data
  ✓ Added edge: fetch_data -> validate_data
  ✓ Added edge: validate_data -> END
✅ Workflow execution completed
```

## 7. Running the Example

```bash
export OPENAI_API_KEY=your_key
go run main.go
```

**Expected Output:**
```text
=== Example 1: Data Processing Workflow ===

User Query: Fetch user data, validate it, transform it to JSON, and save the results

🤔 Planning workflow...
📋 Generated plan: {...}
🚀 Executing planned workflow...
  ✓ Added node: fetch_data
  ✓ Added node: validate_data
  ✓ Added node: transform_data
  ✓ Added node: save_results
📥 Fetching data from database...
✅ Validating data...
🔄 Transforming data...
💾 Saving results...
✅ Workflow execution completed

--- Execution Result ---
Step 1: Workflow plan created with 4 nodes and 5 edges
Step 2: Data fetched: 1000 user records retrieved
Step 3: Data validation passed: all records valid
Step 4: Data transformed to JSON format successfully
Step 5: Results saved to database successfully
------------------------
```

## 8. Advantages

1. **Adaptive Workflows**: Different requests generate different workflows automatically
2. **No Hardcoding**: Don't need to predefine all possible workflow combinations
3. **Intelligent Routing**: LLM understands the intent and creates optimal paths
4. **Reusable Nodes**: Define nodes once, compose them in infinite ways
5. **Natural Language Interface**: Users describe what they want, not how to do it

## 9. Use Cases

- **Data Pipelines**: Dynamically compose ETL workflows
- **Business Processes**: Adaptive approval and processing workflows
- **Multi-step Analysis**: Flexible analysis pipelines based on data characteristics
- **Task Automation**: Intelligently sequence automation tasks
- **Report Generation**: Custom report workflows based on requirements

## 10. Comparison with Other Agents

| Feature | ReAct Agent | Supervisor | Planning Agent |
|---------|-------------|------------|----------------|
| Workflow | Fixed | Fixed routing logic | Dynamic per request |
| Planning | No | No | Yes (LLM-based) |
| Flexibility | Low | Medium | High |
| Use Case | Tool calling | Multi-agent orchestration | Adaptive workflows |

## 11. Tips

1. **Clear Descriptions**: Write clear, descriptive node descriptions - the LLM uses these to plan
2. **Granular Nodes**: Keep nodes focused on single responsibilities
3. **Error Handling**: Implement proper error handling in node functions
4. **Logging**: Use `WithVerbose(true)` during development to understand the planning process
5. **Testing**: Test with various user requests to ensure robust planning

## 12. Next Steps

- Experiment with different node combinations
- Add conditional logic to node functions
- Integrate with real databases and APIs
- Implement error recovery strategies
- Create domain-specific node libraries
