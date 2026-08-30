# Reflection Agent 示例

本示例演示 **Reflection Agent（反思代理）** 设计模式 - 一个通过自我反思和修订来迭代改进响应的 AI 代理。

## 什么是 Reflection 模式？

Reflection 模式是一种技术，AI 代理会：
1. **生成**初始响应
2. **反思**自己的输出，批评质量并识别改进点
3. **修订**基于反思的响应
4. **重复**直到响应令人满意或达到最大迭代次数

这创建了一个反馈循环，产生比单次生成更高质量的输出。

## 工作原理

```
用户查询 → 生成 → 反思 → 满意？
            ↑                ↓ 否
            └──── 修订 ←──────┘
                   ↓ 是
              最终响应
```

### 工作流步骤

1. **生成节点**：创建初始响应或修订版本
2. **反思节点**：评估响应质量并提出改进建议
3. **路由逻辑**：决定继续改进还是完成
4. **最大迭代次数**：防止无限循环

## 核心特性

- ✅ **迭代改进**：多轮精炼
- ✅ **自我批评**：AI 反思自己的输出
- ✅ **可自定义**：针对不同用例的灵活提示
- ✅ **独立模型**：可选择使用不同模型进行生成和反思
- ✅ **智能停止**：自动检测输出何时令人满意

## 安装

```bash
go get github.com/smallnest/langgraphgo
```

## 使用方法

### 基本示例

```go
package main

import (
    "context"
    "fmt"
    "log"

    "github.com/smallnest/langgraphgo/prebuilt"
    "github.com/tmc/langchaingo/llms"
    "github.com/tmc/langchaingo/llms/openai"
)

func main() {
    // 创建 LLM
    model, err := openai.New(openai.WithModel("gpt-4"))
    if err != nil {
        log.Fatal(err)
    }

    // 配置 Reflection Agent
    config := prebuilt.ReflectionAgentConfig{
        Model:         model,
        MaxIterations: 3,
        Verbose:       true,
    }

    // 创建代理
    agent, err := prebuilt.CreateReflectionAgentMap(config)
    if err != nil {
        log.Fatal(err)
    }

    // 准备查询
    initialState := map[string]any{
        "messages": []llms.MessageContent{
            {
                Role:  llms.ChatMessageTypeHuman,
                Parts: []llms.ContentPart{
                    llms.TextPart("解释 CAP 定理"),
                },
            },
        },
    }

    // 调用代理
    result, err := agent.Invoke(context.Background(), initialState)
    if err != nil {
        log.Fatal(err)
    }

    // 提取最终响应
    finalState := result.(map[string]any)
    draft := finalState["draft"].(string)
    fmt.Println(draft)
}
```

### 高级配置

```go
config := prebuilt.ReflectionAgentConfig{
    Model:           generationModel,
    ReflectionModel: reflectionModel,  // 使用独立模型进行反思
    MaxIterations:   3,
    Verbose:         true,

    // 自定义生成提示
    SystemMessage: "你是一位专业的技术写作专家。",

    // 自定义反思提示
    ReflectionPrompt: `评估以下方面：
    1. 技术准确性
    2. 清晰度和组织
    3. 完整性
    4. 示例的使用

    在反馈中要具体。`,
}
```

## 配置选项

| 选项 | 类型 | 默认值 | 描述 |
|------|------|--------|------|
| `Model` | `llms.Model` | *必需* | 用于生成和反思的 LLM |
| `ReflectionModel` | `llms.Model` | `nil` | 可选的独立反思模型 |
| `MaxIterations` | `int` | `3` | 最大精炼循环次数 |
| `SystemMessage` | `string` | 默认值 | 生成步骤的提示 |
| `ReflectionPrompt` | `string` | 默认值 | 反思步骤的提示 |
| `Verbose` | `bool` | `false` | 启用详细日志 |

## 示例输出

```
🎨 生成初始响应...
📝 草稿已生成（1247 字符）

🤔 反思响应...
💭 反思：
**优点：**
- CAP 定理的清晰定义
- 良好的示例使用

**缺点：**
- 缺少权衡讨论
- 可以更好地解释实际含义

**改进建议：**
- 添加实际应用部分
- 包含不同数据库选择的比较

🔄 修订响应（迭代 2）...
📝 草稿已生成（1856 字符）

🤔 反思响应...
💭 反思：
**优点：**
- 全面且结构良好
- 出色的示例和实际应用
- 清晰的权衡讨论

**缺点：**
- 没有主要问题

**改进建议：**
- 不需要改进

✅ 响应令人满意，正在完成

=== 最终响应 ===
[改进的 CAP 定理解释...]
```

## 使用场景

### 1. 技术写作
生成经过多轮精炼的高质量文档。

```go
config := prebuilt.ReflectionAgentConfig{
    SystemMessage: "你是一位创建清晰文档的专业技术写作者。",
    ReflectionPrompt: "评估清晰度、完整性、示例和结构。",
}
```

### 2. 代码审查
提供周到、建设性的代码审查反馈。

```go
config := prebuilt.ReflectionAgentConfig{
    SystemMessage: "你是一位提供代码审查的经验丰富的软件工程师。",
    ReflectionPrompt: "评估建设性、完整性和专业性。",
}
```

### 3. 内容创作
创建经过迭代改进的博客文章、文章或教育内容。

```go
config := prebuilt.ReflectionAgentConfig{
    SystemMessage: "你是一位撰写引人入胜文章的熟练内容创作者。",
    ReflectionPrompt: "评估吸引力、准确性、结构和可读性。",
}
```

### 4. 问题解决
为复杂问题生成经过充分推理的解决方案。

```go
config := prebuilt.ReflectionAgentConfig{
    SystemMessage: "你是一位提供全面分析的问题解决专家。",
    ReflectionPrompt: "评估逻辑推理、完整性和清晰度。",
}
```

## 运行示例

```bash
export OPENAI_API_KEY=your_key
cd examples/reflection_agent
go run main.go
```

示例演示了三种场景：
1. **基本反思**：解释 CAP 定理
2. **技术写作**：创建 API 文档
3. **代码审查**：审查 Go 代码

## 状态结构

代理维护以下状态：

```go
{
    "messages": []llms.MessageContent,    // 对话历史
    "draft": string,                      // 当前响应草稿
    "reflection": string,                 // 最新反思反馈
    "iteration": int,                     // 当前迭代次数
    "is_satisfactory": bool,              // 响应是否足够好
}
```

## 优势

1. **更高质量**：通过多次迭代精炼输出
2. **自我纠正**：代理识别并修复自己的错误
3. **一致性**：系统地遵循评估标准
4. **灵活性**：可针对不同领域和用例进行自定义
5. **透明性**：详细模式显示改进过程

## 与其他模式的比较

| 特性 | 单次生成 | ReAct | Reflection |
|------|---------|-------|------------|
| **质量** | 可变 | 好 | 高 |
| **迭代次数** | 1 | 可变 | 可控 |
| **自我批评** | 否 | 否 | 是 |
| **精炼** | 否 | 有限 | 是 |
| **使用场景** | 简单任务 | 工具使用 | 高质量写作 |
| **延迟** | 低 | 中 | 高 |

## 最佳实践

### 1. 选择合适的最大迭代次数
- 简单任务：2-3 次迭代
- 复杂任务：3-5 次迭代
- 关键内容：5+ 次迭代

### 2. 制定好的反思提示
专注于具体的评估标准：
```go
ReflectionPrompt: `评估以下方面：
1. [具体标准 1]
2. [具体标准 2]
3. [具体标准 3]

在反馈中要具体且可操作。`
```

### 3. 策略性地使用独立模型
- 相同模型：更简单、更一致
- 不同模型：可以利用不同的优势
  - GPT-4 用于生成，GPT-3.5 用于反思（成本优化）
  - 针对特定领域反思的专门模型

### 4. 使用详细模式进行监控
在开发期间启用详细模式以了解反思过程。

### 5. 设置合理的迭代限制
太少：可能无法达到质量阈值
太多：收益递减，成本更高

## 故障排除

### 问题：代理总是在第一次迭代后停止

**解决方案**：检查反思提示。确保它不太宽松。反思应该在早期草稿中识别改进领域。

### 问题：代理每次都达到最大迭代次数

**解决方案**：
1. 检查反思提示是否太挑剔
2. 验证满意度检测逻辑
3. 考虑为复杂任务增加最大迭代次数

### 问题：反思太笼统

**解决方案**：在反思提示中提供更具体的评估标准，并给出要查找的具体示例。

## 高级：自定义满意度检测

您可以修改 `isResponseSatisfactory` 函数以实现自定义逻辑：

```go
// 在您自己的实现中，您可以：
// - 调用 LLM 为响应质量评分（1-10）
// - 对反思使用情感分析
// - 检查特定关键字或模式
// - 将草稿长度与需求进行比较
```

## 下一步

- 针对您的用例尝试不同的提示
- 尝试使用不同的模型进行生成和反思
- 集成到您的应用程序工作流中
- 添加自定义评估指标
- 与其他代理模式结合

## 参考

- [LangGraph 中的 Reflection 模式](https://langchain-ai.github.io/langgraph/concepts/reflection/)
- [Constitutional AI](https://arxiv.org/abs/2212.08073)
- [Self-Refine](https://arxiv.org/abs/2303.17651)
