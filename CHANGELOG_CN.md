# 更新日志

## [0.3.0] - 2025-12-01

### 核心运行时 (Core Runtime)
- **并行执行**: 实现了扇出/扇入 (Fan-out/Fan-in) 执行模型，支持线程安全的状态合并。
- **运行时配置**: 添加了 `RunnableConfig`，用于在图执行上下文中传递配置（如线程 ID、用户 ID 等）。
- **Command API**: 引入了 `Command` 结构体，支持直接从节点进行动态流控制 (`Goto`) 和状态更新 (`Update`)。
- **子图 (Subgraphs)**: 添加了原生支持，通过将编译后的图作为节点来组合图 (`AddSubgraph`)。

### 持久化与检查点 (Persistence & Checkpointing)
- **检查点接口**: 优化了 `CheckpointSaver` 接口以支持状态持久化。
- **实现**: 增加了对 **Redis**、**PostgreSQL** 和 **SQLite** 检查点存储的完整支持。

### 高级状态与流式处理 (Advanced State & Streaming)
- **状态管理**: 引入了 `Schema` 接口和 `Annotated` 风格的归约器 (Reducers)（例如 `AppendMessages`），用于复杂的状态更新。
- **智能消息 (Smart Messages)**: 实现了 `AddMessages` 归约器，用于基于 ID 的消息更新 (Upsert) 和去重。
- **临时通道 (Ephemeral Channels)**: 添加了对临时状态值 (`isEphemeral`) 的支持，这些值在每步后自动清除。
- **增强流式处理**: 添加了类型化的 `StreamEvent` 和 `CallbackHandler` 接口。实现了多种流式模式：`updates`, `values`, `messages`, 和 `debug`。

### 预构建代理 (Pre-built Agents)
- **ToolExecutor**: 添加了用于执行工具的专用节点。
- **ReAct Agent**: 实现了用于创建 ReAct 风格代理的工厂方法。
- **Create Agent**: 添加了 `CreateAgent` 工厂，支持函数式选项以灵活创建代理。
- **Supervisor**: 添加了对 Supervisor 代理模式的支持，用于多代理编排。

### 人机交互 (Human-in-the-loop, HITL)
- **中断 (Interrupts)**: 实现了 `InterruptBefore` 和 `InterruptAfter` 机制以暂停图的执行。
- **恢复与命令 (Resume & Command)**: 添加了通过命令恢复执行和更新状态的支持。
- **时间旅行 (Time Travel)**: 实现了 `GetState` 和 `UpdateState` API 以检查/修改过去的检查点并分叉执行历史。

### 可视化 (Visualization)
- **Mermaid 导出**: 改进了图的可视化，优化了条件边和样式的渲染。

### 实验性与研究 (Experimental & Research)
- **Swarm 模式**: 使用子图 (`examples/swarm`) 添加了多代理协作的原型。
- **Channels RFC**: 添加了 `RFC_CHANNELS.md`，提议在未来改进中采用基于 Channel 的架构。

### LangChain 集成 (LangChain Integration)
- **VectorStore 适配器**: 添加了 `LangChainVectorStore` 适配器，可集成任何 langchaingo vectorstore 实现。
- **支持的后端**: 完整支持 Chroma、Weaviate、Pinecone、Qdrant、Milvus、PGVector 以及任何其他 langchaingo vectorstore。
- **统一接口**: 通过标准的 `AddDocuments`、`SimilaritySearch` 和 `SimilaritySearchWithScore` 方法与 RAG 管道无缝集成。
- **完整适配器**: 现在包含 langchaingo 的 DocumentLoaders、TextSplitters、Embedders 和 VectorStores 适配器。

### 工具与集成 (Tools & Integrations)
- **Tool 包**: 添加了新的 `tool` 包，便于集成外部工具。
- **搜索工具**: 实现了与 `langchaingo` 接口兼容的 `TavilySearch` 和 `ExaSearch` 工具。
- **Agent 集成**: 更新了 `ReAct` Agent 以支持为 OpenAI 兼容 API 生成工具参数 Schema 和解析参数。
- **GoSkills 适配器**: 添加了 `adapter/goskills` 以集成 [GoSkills](github.com/smallnest/goskills) 作为工具。

### 示例 (Examples)
- 添加了涵盖以下内容的综合示例：
  - 检查点 (Postgres, SQLite, Redis)
  - 人机交互工作流
  - Swarm 多代理模式
  - 子图
  - **智能消息 (Smart Messages)** (新增)
  - **Command API** (新增)
  - **临时通道 (Ephemeral Channels)** (新增)
  - **流式模式 (Streaming Modes)** (新增)
  - **时间旅行 / HITL** (新增)
  - **LangChain VectorStore 集成** (新增)
  - **Chroma 向量数据库集成** (新增)
  - **Tavily 搜索工具** (新增)
  - **Exa 搜索工具** (新增)
  - **Create Agent** (新增)
  - **动态技能代理 (Dynamic Skill Agent)** (新增)
  - **Durable Execution** (新增)
  - **GoSkills 集成** (新增)
- **通用**: 改进了所有示例的可靠性和正确性。

## [0.1.0] - 2025-01-02

### 新增
- 通用状态管理 - 适用于任何类型，不仅仅是 MessageContent
- 针对生产环境的性能优化
- 支持任何 LLM 客户端（移除了对 LangChain 的硬依赖）

### 变更
- 简化了构建图的 API
- 更新了示例以展示通用用法

### 修复
- 原始仓库中的 CI/CD 流水线问题
- 最新 Go 版本的构建错误

### 移除
- 对 LangChain 的硬依赖 - 现在可以与任何 LLM 库一起工作
