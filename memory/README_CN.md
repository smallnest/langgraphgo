# 内存管理策略

本包为 AI 代理提供多种内存管理策略，针对不同使用场景和 token 效率进行优化。

## 概述

内存管理对于 AI 代理在控制 token 成本的同时保持上下文至关重要。本包基于优化 AI 代理内存的研究，实现了多种策略。

## 策略列表

### 1. 顺序内存（Sequential Memory - 保留全部）

**使用场景**：需要完美回忆且 token 成本不是问题时

**优点**：
- 完美回忆所有交互
- 实现简单
- 无信息损失

**缺点**：
- token 无限增长
- 可能变得非常昂贵
- 无优化

**示例**：
```go
mem := memory.NewSequentialMemory()

msg := memory.NewMessage("user", "你好，AI！")
mem.AddMessage(ctx, msg)

response := memory.NewMessage("assistant", "你好！有什么可以帮助的？")
mem.AddMessage(ctx, response)

// 获取所有消息
messages, _ := mem.GetContext(ctx, "")
```

### 2. 滑动窗口内存（Sliding Window）

**使用场景**：维持最近对话上下文，限定大小

**优点**：
- 防止无限增长
- 维持最近对话流
- 简单且可预测

**缺点**：
- 丢失旧上下文
- 可能忘记重要的早期信息

**示例**：
```go
// 仅保留最近 10 条消息
mem := memory.NewSlidingWindowMemory(10)

for i := 0; i < 20; i++ {
    msg := memory.NewMessage("user", fmt.Sprintf("消息 %d", i))
    mem.AddMessage(ctx, msg)
}

// 仅保留最后 10 条消息
messages, _ := mem.GetContext(ctx, "")
```

### 3. 摘要式内存（Summarization-Based）

**使用场景**：长对话中历史上下文重要但需要压缩时

**优点**：
- 维持历史意识
- 减少 token 消耗
- 保留重要信息

**缺点**：
- 需要调用 LLM 进行摘要
- 可能丢失具体细节
- 摘要质量取决于 LLM

**示例**：
```go
mem := memory.NewSummarizationMemory(&memory.SummarizationConfig{
    RecentWindowSize: 10,   // 保留最近 10 条完整消息
    SummarizeAfter:   20,   // 超过 20 条时进行摘要
    Summarizer: func(ctx context.Context, messages []*Message) (string, error) {
        // 调用你的 LLM 生成摘要
        return llm.Summarize(messages)
    },
})

// 随着消息累积，旧消息会自动被摘要
```

### 4. 检索式内存（Retrieval-Based）

**使用场景**：大量对话历史，只需要相关上下文时

**优点**：
- 高效的 token 使用
- 仅检索相关信息
- 在大型历史记录中扩展良好

**缺点**：
- 需要嵌入模型
- 可能错过按时间顺序重要的上下文
- 嵌入生成增加延迟

**示例**：
```go
mem := memory.NewRetrievalMemory(&memory.RetrievalConfig{
    TopK: 5, // 检索最相关的 5 条消息
    EmbeddingFunc: func(ctx context.Context, text string) ([]float64, error) {
        // 调用嵌入 API（如 OpenAI embeddings）
        return openai.CreateEmbedding(text)
    },
})

// 添加多条消息
for _, msg := range manyMessages {
    mem.AddMessage(ctx, msg)
}

// 仅检索相关消息
relevantMessages, _ := mem.GetContext(ctx, "告诉我关于定价的信息")
```

### 5. 分层内存（Hierarchical）

**使用场景**：复杂对话，具有不同重要性级别

**优点**：
- 平衡最近性和重要性
- 灵活的优先级
- 维持关键信息

**缺点**：
- 管理更复杂
- 需要重要性评分
- 实现复杂度更高

**示例**：
```go
mem := memory.NewHierarchicalMemory(&memory.HierarchicalConfig{
    RecentLimit:    10,  // 最近消息
    ImportantLimit: 20,  // 重要消息
    ImportanceScorer: func(msg *Message) float64 {
        // 自定义评分逻辑
        if strings.Contains(msg.Content, "重要") {
            return 0.9
        }
        return 0.5
    },
})

// 标记重要消息
importantMsg := memory.NewMessage("user", "重要：记住这条规则")
importantMsg.Metadata["importance"] = 0.95
mem.AddMessage(ctx, importantMsg)
```

### 6. 缓冲内存（Buffer Memory）

**使用场景**：通用内存，灵活限制（类似 LangChain）

**优点**：
- 灵活配置
- 可选自动摘要
- 可按消息数或 token 数限制

**缺点**：
- 可能需要调优以获得最佳性能

**示例**：
```go
mem := memory.NewBufferMemory(&memory.BufferConfig{
    MaxMessages:   50,    // 限制为 50 条消息
    MaxTokens:     2000,  // 或 2000 个 tokens，以先达到者为准
    AutoSummarize: true,  // 超过限制时自动摘要
})

// 消息自动管理
```

## 接口

所有策略都实现 `Strategy` 接口：

```go
type Strategy interface {
    // 添加消息到内存
    AddMessage(ctx context.Context, msg *Message) error

    // 获取当前查询的相关上下文
    GetContext(ctx context.Context, query string) ([]*Message, error)

    // 清除所有内存
    Clear(ctx context.Context) error

    // 获取统计信息
    GetStats(ctx context.Context) (*Stats, error)
}
```

## 消息结构

```go
type Message struct {
    ID         string                 // 唯一标识符
    Role       string                 // "user"、"assistant"、"system"
    Content    string                 // 消息内容
    Timestamp  time.Time              // 创建时间
    Metadata   map[string]interface{} // 附加元数据
    TokenCount int                    // 估计的 token 数
}
```

## 统计信息

所有策略都提供统计信息：

```go
stats, _ := mem.GetStats(ctx)
fmt.Printf("总消息数: %d\n", stats.TotalMessages)
fmt.Printf("活跃消息数: %d\n", stats.ActiveMessages)
fmt.Printf("总 Tokens: %d\n", stats.TotalTokens)
fmt.Printf("压缩率: %.2f\n", stats.CompressionRate)
```

## 选择策略

| 场景 | 推荐策略 |
|------|---------|
| 短对话，成本不是问题 | Sequential（顺序） |
| 有界历史的聊天 | Sliding Window（滑动窗口） |
| 长对话，需要压缩 | Summarization（摘要） |
| 大型知识库，查询驱动 | Retrieval（检索） |
| 复杂多主题对话 | Hierarchical（分层） |
| 通用目的，灵活 | Buffer（缓冲） |

## 集成示例

```go
// 创建首选策略
strategy := memory.NewSlidingWindowMemory(20)

// 随着对话进行添加消息
userMsg := memory.NewMessage("user", "天气怎么样？")
strategy.AddMessage(ctx, userMsg)

// 获取 LLM 的上下文
messages, _ := strategy.GetContext(ctx, "当前查询")

// 格式化为 LLM 使用
prompt := formatMessagesForLLM(messages)
response := llm.Generate(prompt)

// 将响应添加到内存
assistantMsg := memory.NewMessage("assistant", response)
strategy.AddMessage(ctx, assistantMsg)
```

## 高级用法

### 自定义重要性评分器

```go
scorer := func(msg *Message) float64 {
    score := 0.5

    // 提升系统消息
    if msg.Role == "system" {
        score += 0.3
    }

    // 提升包含关键词的消息
    if strings.Contains(msg.Content, "记住") {
        score += 0.2
    }

    return math.Min(score, 1.0)
}

mem := memory.NewHierarchicalMemory(&memory.HierarchicalConfig{
    ImportanceScorer: scorer,
})
```

### 自定义摘要器

```go
summarizer := func(ctx context.Context, messages []*Message) (string, error) {
    // 使用你的 LLM
    prompt := "总结以下对话：\n\n"
    for _, msg := range messages {
        prompt += fmt.Sprintf("%s: %s\n", msg.Role, msg.Content)
    }

    return llm.Complete(ctx, prompt)
}

mem := memory.NewSummarizationMemory(&memory.SummarizationConfig{
    Summarizer: summarizer,
})
```

### 自定义嵌入

```go
embedder := func(ctx context.Context, text string) ([]float64, error) {
    // 使用 OpenAI、Cohere 或你的嵌入模型
    return openai.CreateEmbedding(ctx, text)
}

mem := memory.NewRetrievalMemory(&memory.RetrievalConfig{
    EmbeddingFunc: embedder,
})
```

## 测试

运行测试：
```bash
go test ./memory -v
```

## 参考

- 基于 [optimize-ai-agent-memory](https://github.com/FareedKhan-dev/optimize-ai-agent-memory) 的研究
- 实现类似 LangChain 内存系统的模式
- 针对 Go 和 LangGraphGo 集成进行优化
