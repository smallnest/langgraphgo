# LangGraphGo 示例

本目录包含演示 LangGraphGo 特性的各种示例。

## 基本概念
- **[基本示例 (Basic Example)](basic_example/README_CN.md)**: 带有硬编码步骤的简单图。
- **[基本 LLM (Basic LLM)](basic_llm/README_CN.md)**: 与 LLM 的集成。
- **[条件路由 (Conditional Routing)](conditional_routing/README_CN.md)**: 基于状态的动态路由。
- **[条件边 (Conditional Edges)](conditional_edges_example/README_CN.md)**: 使用条件边。

## 高级特性
- **[并行执行 (Parallel Execution)](parallel_execution/README_CN.md)**: 带有状态合并的扇出/扇入 (Fan-out/Fan-in) 执行。
- **[配置 (Configuration)](configuration/README_CN.md)**: 使用运行时配置传递元数据和设置。
- **[自定义归约器 (Custom Reducer)](custom_reducer/README_CN.md)**: 为复杂的合并逻辑定义自定义状态归约器。
- **[State Schema](state_schema/README_CN.md)**: 使用 Schema 和 Reducer 管理复杂的状态更新。
- **[子图 (Subgraphs)](subgraphs/README_CN.md)**: 在图中组合图 (新)。
- **[流式模式 (Streaming Modes)](streaming_modes/README_CN.md)**: 支持 updates, values, messages 等模式的高级流式处理。
- **[智能消息 (Smart Messages)](smart_messages/README_CN.md)**: 支持基于 ID 更新 (Upsert) 的智能消息合并。
- **[Command API](command_api/README_CN.md)**: 节点级的动态流控制和状态更新。
- **[临时通道 (Ephemeral Channels)](ephemeral_channels/README_CN.md)**: 管理每步后自动清除的临时状态。
- **[监听器 (Listeners)](listeners/README_CN.md)**: 向图添加事件监听器。

## 持久化 (检查点 Checkpointing)
- **[内存 (Memory)](checkpointing/main.go)**: 内存检查点。
- **[PostgreSQL](checkpointing/postgres/)**: 使用 PostgreSQL 的持久化状态。
- **[SQLite](checkpointing/sqlite/)**: 使用 SQLite 的持久化状态。
- **[Redis](checkpointing/redis/)**: 使用 Redis 的持久化状态。
- **[持久化执行 (Durable Execution)](durable_execution/README_CN.md)**: 崩溃恢复和从检查点恢复执行。

## 人机交互 (Human-in-the-loop)
- **[人工审批 (Human Approval)](human_in_the_loop/README_CN.md)**: 包含中断和人工审批步骤的工作流。
- **[时间旅行 / HITL (Time Travel)](time_travel/README_CN.md)**: 检查、修改状态历史并分叉执行 (UpdateState)。
- **[动态中断 (Dynamic Interrupt)](dynamic_interrupt/README_CN.md)**: 使用 `graph.Interrupt` 在节点内部暂停执行。

## 预构建代理 (Pre-built Agents)
- **[Create Agent](create_agent/README_CN.md)**: 使用选项轻松创建代理。
- **[动态技能代理 (Dynamic Skill Agent)](dynamic_skill_agent/README_CN.md)**: 具有动态技能发现和选择功能的代理。
- **[ReAct Agent](react_agent/README_CN.md)**: 使用工具的推理与行动 (Reason and Action) 代理。
- **[Supervisor](supervisor/README_CN.md)**: 使用 Supervisor 进行多代理编排。
- **[Swarm](swarm/README_CN.md)**: 使用切换 (handoffs) 的多代理协作.

## Memory (记忆)
- **[Memory Basic](memory_basic/README_CN.md)**: LangChain Memory 的基本用法。
- **[Memory Chatbot](memory_chatbot/README_CN.md)**: 集成 Memory 的聊天机器人。

## RAG (检索增强生成)
- **[RAG Basic](rag_basic/README_CN.md)**: 基础 RAG 实现。
- **[RAG Pipeline](rag_pipeline/README_CN.md)**: 完整的 RAG 管道。
- **[RAG Advanced](rag_advanced/README_CN.md)**: 高级 RAG 技术。
- **[RAG Conditional](rag_conditional/README_CN.md)**: 条件 RAG 工作流。
- **[RAG with Embeddings](rag_with_embeddings/README_CN.md)**: 使用 Embeddings 的 RAG。
- **[RAG with LangChain](rag_with_langchain/README_CN.md)**: 使用 LangChain 组件的 RAG。
- **[RAG with VectorStores](rag_langchain_vectorstore_example/README_CN.md)**: 使用 LangChain VectorStores 的 RAG。
- **[RAG with Chroma](rag_chroma_example/README_CN.md)**: 使用 Chroma 数据库的 RAG。

## 其他
- **[可视化 (Visualization)](visualization/README_CN.md)**: 生成图的 Mermaid 图表。
- **[LangChain 集成 (LangChain Integration)](langchain_example/README_CN.md)**: 使用 LangChain 工具和模型。
- **[Tavily Search](tool_tavily/README_CN.md)**: 使用 Tavily 搜索工具和 ReAct Agent。
- **[Exa Search](tool_exa/README_CN.md)**: 使用 Exa 搜索工具和 ReAct Agent。
- **[GoSkills 集成 (GoSkills Integration)](goskills_example/README_CN.md)**: 将 GoSkills 作为工具集成到 Agent 中。
