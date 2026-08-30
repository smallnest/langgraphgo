# Planning Agent 示例

本示例演示如何使用 **Planning Agent（规划代理）** - 一个能够根据用户请求使用 LLM 推理动态创建工作流计划的智能代理。

## 1. 背景

传统代理遵循预定义的工作流。**Planning Agent** 不同：
1. **分析**用户的请求
2. **规划**最优工作流，选择和排序可用节点
3. **执行**动态生成的计划

这种方法提供：
- **灵活性**：工作流适应不同的用户请求
- **智能性**：LLM 确定最佳操作序列
- **效率**：仅执行必要的步骤

## 2. 核心概念

- **可用节点**：可以组合成工作流的预定义操作（节点）集合
- **规划节点**：使用 LLM 根据用户请求生成 JSON 格式的工作流计划
- **执行节点**：动态构建并执行计划的工作流
- **工作流计划**：描述节点和边的结构化 JSON（类似 Mermaid 图）

## 3. 工作原理

### 步骤 1：定义可用节点
```go
nodes := []*graph.Node{
    {
        Name:        "fetch_data",
        Description: "从数据库获取用户数据",
        Function:    fetchDataNode,
    },
    {
        Name:        "validate_data",
        Description: "验证数据的完整性和格式",
        Function:    validateDataNode,
    },
    // ... 更多节点
}
```

### 步骤 2：创建 Planning Agent
```go
agent, err := prebuilt.CreatePlanningAgentMap(
    model,
    nodes,
    []tools.Tool{},
    prebuilt.WithVerbose(true), // 可选：显示详细日志
)
```

### 步骤 3：使用用户请求执行
```go
query := "获取用户数据，验证它，并保存结果"
initialState := map[string]any{
    "messages": []llms.MessageContent{
        llms.TextParts(llms.ChatMessageTypeHuman, query),
    },
}
res, err := agent.Invoke(context.Background(), initialState)
```

## 4. 工作流计划格式

LLM 生成以下 JSON 格式的计划：
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

这创建了一个工作流：`START → fetch_data → validate_data → save_results → END`

## 5. 示例场景

### 场景 1：数据处理
**请求**："获取用户数据，验证它，转换为 JSON，并保存结果"

**生成的计划**：
```
START → fetch_data → validate_data → transform_data → save_results → END
```

### 场景 2：数据分析
**请求**："获取数据，分析它，并生成报告"

**生成的计划**：
```
START → fetch_data → analyze_data → generate_report → END
```

### 场景 3：完整管道
**请求**："获取数据，验证和转换它，分析结果，并生成综合报告"

**生成的计划**：
```
START → fetch_data → validate_data → transform_data → analyze_data → generate_report → END
```

## 6. 代码亮点

### 定义节点
```go
func fetchDataNode(ctx context.Context, state any) (any, error) {
    mState := state.(map[string]any)
    messages := mState["messages"].([]llms.MessageContent)

    // 你的业务逻辑
    fmt.Println("📥 从数据库获取数据...")

    msg := llms.MessageContent{
        Role:  llms.ChatMessageTypeAI,
        Parts: []llms.ContentPart{llms.TextPart("数据获取成功")},
    }

    return map[string]any{
        "messages": append(messages, msg),
    }, nil
}
```

### 详细输出
启用 `WithVerbose(true)` 时，你会看到：
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

## 7. 运行示例

```bash
export OPENAI_API_KEY=your_key
go run main.go
```

**预期输出：**
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

## 8. 优势

1. **自适应工作流**：不同的请求自动生成不同的工作流
2. **无需硬编码**：不需要预定义所有可能的工作流组合
3. **智能路由**：LLM 理解意图并创建最优路径
4. **可重用节点**：定义一次节点，以无限方式组合它们
5. **自然语言接口**：用户描述他们想要什么，而不是如何做

## 9. 使用场景

- **数据管道**：动态组合 ETL 工作流
- **业务流程**：自适应审批和处理工作流
- **多步分析**：基于数据特征的灵活分析管道
- **任务自动化**：智能排序自动化任务
- **报告生成**：基于需求的自定义报告工作流

## 10. 与其他代理的比较

| 特性 | ReAct Agent | Supervisor | Planning Agent |
|------|-------------|------------|----------------|
| 工作流 | 固定 | 固定路由逻辑 | 每个请求动态生成 |
| 规划 | 无 | 无 | 是（基于 LLM） |
| 灵活性 | 低 | 中 | 高 |
| 使用场景 | 工具调用 | 多代理编排 | 自适应工作流 |

## 11. 提示

1. **清晰的描述**：编写清晰、描述性的节点描述 - LLM 使用这些来规划
2. **细粒度节点**：保持节点专注于单一职责
3. **错误处理**：在节点函数中实现适当的错误处理
4. **日志记录**：在开发期间使用 `WithVerbose(true)` 来理解规划过程
5. **测试**：使用各种用户请求进行测试以确保稳健的规划

## 12. 下一步

- 尝试不同的节点组合
- 向节点函数添加条件逻辑
- 与真实数据库和 API 集成
- 实现错误恢复策略
- 创建特定领域的节点库
