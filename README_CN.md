# 🦜️🔗 LangGraphGo

[![go.dev reference](https://img.shields.io/badge/go.dev-reference-007d9c?logo=go&logoColor=white&style=flat-square)](https://pkg.go.dev/github.com/smallnest/langgraphgo)

[English](./README.md) | [简体中文](./README_CN.md)

> 🔀 **Fork 自 [paulnegz/langgraphgo](https://github.com/paulnegz/langgraphgo)** - 增强了流式传输、可视化、可观测性和生产就绪特性。
>
> 本分支旨在**实现与 Python LangGraph 库的功能对齐**，增加了对并行执行、持久化、高级状态管理、预构建 Agent 和人在回路（HITL）工作流的支持。

## 📦 安装

```bash
go get github.com/smallnest/langgraphgo
```

## 🚀 特性

- **核心运行时**:
    - **并行执行**: 支持节点的并发执行（扇出），并具备线程安全的状态合并。
    - **运行时配置**: 通过 `RunnableConfig` 传播回调、标签和元数据。
    - **LangChain 兼容**: 与 `langchaingo` 无缝协作。

- **持久化与可靠性**:
    - **Checkpointers**: 提供 Redis、Postgres 和 SQLite 实现，用于持久化状态。
    - **状态恢复**: 支持从 Checkpoint 暂停和恢复执行。

- **高级能力**:
    - **状态 Schema**: 支持细粒度的状态更新和自定义 Reducer（例如 `AppendReducer`）。
    - **智能消息**: 支持基于 ID 更新 (Upsert) 的智能消息合并 (`AddMessages`)。
    - **Command API**: 节点级的动态流控制和状态更新。
    - **临时通道**: 管理每步后自动清除的临时状态。
    - **子图**: 通过嵌套图来构建复杂的 Agent。
    - **增强流式传输**: 支持多种模式 (`updates`, `values`, `messages`) 的实时事件流。
    - **预构建 Agent**: 开箱即用的 `ReAct`, `CreateAgent` 和 `Supervisor` Agent 工厂。

- **开发者体验**:
    - **可视化**: 支持导出为 Mermaid、DOT 和 ASCII 图表，并支持条件边。
    - **人在回路 (HITL)**: 中断执行、检查状态、编辑历史 (`UpdateState`) 并恢复。
    - **可观测性**: 内置追踪和指标支持。
    - **工具**: 集成了 `Tavily` 和 `Exa` 搜索工具。

## 🎯 快速开始

```go
package main

import (
	"context"
	"fmt"
	"log"

	"github.com/smallnest/langgraphgo/graph"
	"github.com/tmc/langchaingo/llms"
	"github.com/tmc/langchaingo/llms/openai"
)

func main() {
	ctx := context.Background()
	model, _ := openai.New()

	// 1. 创建图
	g := graph.NewMessageGraph()

	// 2. 添加节点
	g.AddNode("generate", func(ctx context.Context, state interface{}) (interface{}, error) {
		messages := state.([]llms.MessageContent)
		response, _ := model.GenerateContent(ctx, messages)
		return append(messages, llms.TextParts("ai", response.Choices[0].Content)), nil
	})

	// 3. 定义边
	g.AddEdge("generate", graph.END)
	g.SetEntryPoint("generate")

	// 4. 编译
	runnable, _ := g.Compile()

	// 5. 调用
	initialState := []llms.MessageContent{
		llms.TextParts("human", "Hello, LangGraphGo!"),
	}
	result, _ := runnable.Invoke(ctx, initialState)
	
	fmt.Println(result)
}
```

## 📚 示例

- **[基础 LLM](./examples/basic_llm/)** - 简单的 LangChain 集成
- **[RAG 流程](./examples/rag_pipeline/)** - 完整的检索增强生成
- **[RAG 与 LangChain](./examples/rag_with_langchain/)** - LangChain 组件集成
- **[RAG 与 VectorStores](./examples/rag_langchain_vectorstore_example/)** - LangChain VectorStore 集成 (新增!)
- **[RAG 与 Chroma](./examples/rag_chroma_example/)** - Chroma 向量数据库集成 (新增!)
- **[Tavily 搜索](./examples/tool_tavily/)** - Tavily 搜索工具集成 (新增!)
- **[Exa 搜索](./examples/tool_exa/)** - Exa 搜索工具集成 (新增!)
- **[流式传输](./examples/streaming_pipeline/)** - 实时进度更新
- **[条件路由](./examples/conditional_routing/)** - 动态路径选择
- **[Checkpointing](./examples/checkpointing/)** - 保存和恢复状态
- **[可视化](./examples/visualization/)** - 导出图表
- **[监听器](./examples/listeners/)** - 进度、指标和日志
- **[子图](./examples/subgraphs/)** - 嵌套图组合
- **[Swarm](./examples/swarm/)** - 多 Agent 协作
- **[Create Agent](./examples/create_agent/)** - 使用选项灵活创建 Agent (新增!)
- **[动态技能代理 (Dynamic Skill Agent)](./examples/dynamic_skill_agent/)** - 具有动态技能发现和选择功能的代理 (新增!)
- **[State Schema](./examples/state_schema/)** - 使用 Reducer 进行复杂状态管理
- **[智能消息](./examples/smart_messages/)** - 智能消息合并 (Upserts)
- **[Command API](./examples/command_api/)** - 动态流控制
- **[临时通道](./examples/ephemeral_channels/)** - 临时状态管理
- **[流式模式](./examples/streaming_modes/)** - 高级流式模式
- **[Time Travel / HITL](./examples/time_travel/)** - 检查、编辑和分叉状态历史
- **[Dynamic Interrupt](./examples/dynamic_interrupt/)** - 在节点内部暂停执行
- **[Durable Execution](./examples/durable_execution/)** - 崩溃恢复和从检查点恢复执行
- **[GoSkills 集成](./examples/goskills_example/)** - GoSkills 集成 (新增!)

## 🔧 核心概念

### 并行执行
当多个节点共享同一个起始节点时，LangGraphGo 会自动并行执行它们。结果将使用图的状态合并器或 Schema 进行合并。

```go
g.AddEdge("start", "branch_a")
g.AddEdge("start", "branch_b")
// branch_a 和 branch_b 将并发运行
```

### 人在回路 (HITL)
暂停执行以允许人工批准或输入。

```go
config := &graph.Config{
    InterruptBefore: []string{"human_review"},
}

// 执行在 "human_review" 节点前停止
state, err := runnable.InvokeWithConfig(ctx, input, config)

// 恢复执行
resumeConfig := &graph.Config{
    ResumeFrom: []string{"human_review"},
}
runnable.InvokeWithConfig(ctx, state, resumeConfig)
```

### 预构建 Agent
使用工厂函数快速创建复杂的 Agent。

```go
// 创建 ReAct Agent
agent, err := prebuilt.CreateReactAgent(model, tools)

// 使用选项创建 Agent
agent, err := prebuilt.CreateAgent(model, tools, prebuilt.WithSystemMessage("System prompt"))

// 创建 Supervisor Agent
supervisor, err := prebuilt.CreateSupervisor(model, agents)
```

## 🎨 图可视化

```go
exporter := runnable.GetGraph()
fmt.Println(exporter.DrawMermaid()) // 生成 Mermaid 流程图
```

## 📈 性能

- **图操作**: ~14-94μs (取决于格式)
- **追踪开销**: ~4μs / 次执行
- **事件处理**: 1000+ 事件/秒
- **流式延迟**: <100ms

## 🧪 测试

```bash
go test ./... -v
```

## 🤝 贡献

本项目欢迎贡献！请首选创建feature issues，然后提交PR。

## 📄 许可证

MIT License - 详情请见原始仓库。
