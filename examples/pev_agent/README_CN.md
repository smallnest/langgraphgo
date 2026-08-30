# PEV Agent 示例

## 概述

本示例演示了 **PEV (Plan, Execute, Verify，计划-执行-验证)** 代理模式，这是一种用于可靠任务执行的健壮且具有自我纠正能力的架构。PEV 在准确性和可靠性至关重要的高风险自动化场景中特别有价值。

## 什么是 PEV？

PEV 是一种代理架构，实现了具有内置错误检测和恢复功能的三阶段工作流：

1. **Plan (计划)**：将用户请求分解为具体的可执行步骤
2. **Execute (执行)**：使用可用工具运行每个步骤
3. **Verify (验证)**：检查执行是否成功

如果验证失败，代理会触发重新规划循环，根据失败上下文创建改进的计划。这创建了一个"质量保证检查点"，确保只有有效数据流入最终合成阶段。

## 架构

```
┌─────────────────────────────────────────────────────────────┐
│                        PEV Workflow                         │
└─────────────────────────────────────────────────────────────┘

   User Request
        ↓
   ┌─────────┐
   │ Planner │──────┐ (Re-plan on failure)
   └─────────┘      │
        ↓           │
   ┌──────────┐     │
   │ Executor │     │
   └──────────┘     │
        ↓           │
   ┌──────────┐     │
   │ Verifier │     │
   └──────────┘     │
        ↓           │
   Verification     │
   Successful?      │
        ├─No───────┘
        │
       Yes
        ↓
   ┌─────────────┐
   │ Synthesizer │
   └─────────────┘
        ↓
   Final Answer
```

## 核心特性

- **自我纠正**：自动重试失败的操作并使用改进的计划
- **逐步验证**：在继续之前验证每次执行
- **错误恢复**：从失败中学习以创建更好的计划
- **可配置重试**：控制最大重试次数
- **工具无关**：适用于任何实现 `tools.Tool` 接口的工具

## 使用场景

PEV 非常适合：

- **高风险自动化**：金融系统、医疗保健、法律处理
- **不可靠的外部工具**：具有间歇性故障、网络问题的 API
- **复杂的多步骤任务**：需要按顺序进行多次工具调用的任务
- **质量关键应用**：必须在继续之前验证准确性的场景

## 状态模式

PEV 代理维护以下状态：

```go
{
    "messages": []llms.MessageContent,      // 对话历史
    "plan": []string,                       // 当前执行计划
    "current_step": int,                    // 当前步骤索引
    "last_tool_result": string,             // 上次工具执行的结果
    "intermediate_steps": []string,         // 所有步骤的历史
    "retries": int,                         // 当前重试计数
    "verification_result": VerificationResult, // 上次验证结果
    "final_answer": string,                 // 合成的最终响应
}
```

## 配置

```go
type PEVAgentConfig struct {
    Model              llms.Model    // 用于规划和验证的 LLM
    Tools              []tools.Tool  // 可用于执行的工具
    MaxRetries         int           // 最大重试次数（默认：3）
    SystemMessage      string        // 自定义规划器提示词（可选）
    VerificationPrompt string        // 自定义验证器提示词（可选）
    Verbose            bool          // 启用详细日志
}
```

## 示例

### 示例 1：简单计算

演示使用可靠计算器工具的基本 PEV 操作：

```go
config := prebuilt.PEVAgentConfig{
    Model:      model,
    Tools:      []tools.Tool{CalculatorTool{}},
    MaxRetries: 3,
    Verbose:    true,
}

agent, err := prebuilt.CreatePEVAgentMap(config)
```

**查询**："计算 15 乘以 8 的结果"

**预期流程**：
1. 计划："将 15 乘以 8"
2. 执行：使用计算器工具 → "120.00"
3. 验证：✅ 成功
4. 合成："结果是 120"

### 示例 2：不可靠的天气 API

演示使用具有 40% 失败率的工具进行自我纠正：

```go
config := prebuilt.PEVAgentConfig{
    Model: model,
    Tools: []tools.Tool{
        WeatherTool{FailureRate: 0.4}, // 40% 失败概率
    },
    MaxRetries: 3,
    Verbose:    true,
}
```

**查询**："东京的天气如何？"

**可能的流程**：
1. 计划："获取东京的天气"
2. 执行：调用天气 API → "错误：连接超时"
3. 验证：❌ 检测到失败
4. 重新规划："使用正确的城市名称重试东京的天气查询"
5. 执行：调用天气 API → "东京天气：22°C，晴朗"
6. 验证：✅ 成功
7. 合成："东京的天气是 22°C，晴朗"

### 示例 3：多步骤任务

演示使用多个步骤和不同工具的 PEV：

```go
config := prebuilt.PEVAgentConfig{
    Model: model,
    Tools: []tools.Tool{
        CalculatorTool{},
        WeatherTool{FailureRate: 0.2},
        DatabaseTool{FailureRate: 0.3},
    },
    MaxRetries: 3,
    Verbose:    true,
}
```

**查询**："首先，计算 25 乘以 4。然后，查询巴黎的天气。"

## 运行示例

1. 设置 OpenAI API 密钥：
```bash
export OPENAI_API_KEY=your-api-key-here
```

2. 运行示例：
```bash
cd examples/pev_agent
go run main.go
```

## 实现细节

### 规划器节点

- 分析用户请求并将其分解为步骤
- 失败时，接收验证反馈以创建改进的计划
- 返回可操作步骤的编号列表

### 执行器节点

- 使用适当的工具执行当前步骤
- 优雅地处理工具错误
- 返回执行结果以供验证

### 验证器节点

- 使用 LLM 分析执行结果
- 返回结构化验证结果：
  ```go
  type VerificationResult struct {
      IsSuccessful bool   `json:"is_successful"`
      Reasoning    string `json:"reasoning"`
  }
  ```
- 在工具输出中查找成功/失败指标

### 合成器节点

- 组合所有成功的中间步骤
- 生成连贯的最终答案
- 仅在所有步骤成功或达到最大重试次数后调用

## 与其他模式的比较

| 模式 | 自我纠正 | 验证 | 使用场景 |
|---------|----------------|--------------|----------|
| **ReAct** | 否 | 否 | 快速、简单的任务 |
| **Reflection** | 是 | 生成后 | 内容质量改进 |
| **PEV** | 是 | 每步 | 使用工具的可靠执行 |

## 最佳实践

1. **工具设计**：创建返回清晰错误消息的工具
2. **重试限制**：根据工具可靠性设置 `MaxRetries`
3. **详细模式**：在开发期间启用以了解失败原因
4. **验证提示词**：针对特定领域的验证进行自定义
5. **步骤粒度**：保持步骤原子化和可独立验证

## 限制

- **增加延迟**：多个验证步骤增加处理时间
- **更高成本**：与更简单的模式相比，LLM 调用更多
- **复杂性**：对于简单、可靠的操作来说过度设计

## 权衡

**优势**：
- 高可靠性和准确性
- 自动错误恢复
- 清晰的执行审计跟踪

**劣势**：
- 更昂贵（额外的 LLM 调用）
- 比单次通过模式慢
- 需要设计良好的工具，具有清晰的成功/失败指标


## 许可证

此实现是 langgraphgo 项目的一部分。
