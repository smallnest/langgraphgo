# Changelog

## [0.3.0] - 2025-12-01

### Core Runtime
- **Parallel Execution**: Implemented fan-out/fan-in execution model with thread-safe state merging.
- **Runtime Configuration**: Added `RunnableConfig` to propagate configuration (like thread IDs, user IDs) through the graph execution context.
- **Command API**: Introduced `Command` struct for dynamic flow control (`Goto`) and state updates (`Update`) directly from nodes.
- **Subgraphs**: Added native support for composing graphs by using compiled graphs as nodes (`AddSubgraph`).

### Persistence & Checkpointing
- **Checkpoint Interface**: Refined `CheckpointSaver` interface for state persistence.
- **Implementations**: Added full support for **Redis**, **PostgreSQL**, and **SQLite** checkpoint stores.

### Advanced State & Streaming
- **State Management**: Introduced `Schema` interface and `Annotated` style reducers (e.g., `AppendMessages`) for complex state updates.
- **Smart Messages**: Implemented `AddMessages` reducer for ID-based message upserts and deduplication.
- **Ephemeral Channels**: Added support for temporary state values (`isEphemeral`) that are automatically cleared after each step.
- **Enhanced Streaming**: Added typed `StreamEvent`s and `CallbackHandler` interface. Implemented multiple streaming modes: `updates`, `values`, `messages`, and `debug`.

### Pre-built Agents
- **ToolExecutor**: Added a dedicated node for executing tools.
- **ReAct Agent**: Implemented a factory for creating ReAct-style agents.
- **Create Agent**: Added `CreateAgent` factory with functional options for flexible agent creation.
- **Supervisor**: Added support for Supervisor agent patterns for multi-agent orchestration.

### Human-in-the-loop (HITL)
- **Interrupts**: Implemented `InterruptBefore` and `InterruptAfter` mechanisms to pause graph execution.
- **Resume & Command**: Added support for resuming execution and updating state via commands.
- **Time Travel**: Implemented `GetState` and `UpdateState` APIs to inspect/modify past checkpoints and fork execution history.

### Visualization
- **Mermaid Export**: Improved graph visualization with better rendering of conditional edges and styling options.

### Experimental & Research
- **Swarm Patterns**: Added prototypes for multi-agent collaboration using subgraphs (`examples/swarm`).
- **Channels RFC**: Added `RFC_CHANNELS.md` proposing a channel-based architecture for future improvements.

### LangChain Integration
- **VectorStore Adapter**: Added `LangChainVectorStore` adapter to integrate any langchaingo vectorstore implementation.
- **Supported Backends**: Full support for Chroma, Weaviate, Pinecone, Qdrant, Milvus, PGVector, and any other langchaingo vectorstore.
- **Unified Interface**: Seamless integration with RAG pipelines through standard `AddDocuments`, `SimilaritySearch`, and `SimilaritySearchWithScore` methods.
- **Complete Adapters**: Now includes adapters for DocumentLoaders, TextSplitters, Embedders, and VectorStores from langchaingo.

### Tools & Integrations
- **Tool Package**: Added a new `tool` package for easy integration of external tools.
- **Search Tools**: Implemented `TavilySearch` and `ExaSearch` tools compatible with `langchaingo` interfaces.
- **Agent Integration**: Updated `ReAct` agent to support tool parameter schema generation and argument parsing for OpenAI-compatible APIs.
- **GoSkills Adapter**: Added `adapter/goskills` to integrate [GoSkills](github.com/smallnest/goskills) as tools.

### Examples
- Added comprehensive examples for:
  - Checkpointing (Postgres, SQLite, Redis)
  - Human-in-the-loop workflows
  - Swarm multi-agent patterns
  - Subgraphs
  - **Smart Messages** (new)
  - **Command API** (new)
  - **Ephemeral Channels** (new)
  - **Streaming Modes** (new)
  - **Time Travel / HITL** (new)
  - **LangChain VectorStore integration** (new)
  - **Chroma vector database integration** (new)
  - **Tavily Search Tool** (new)
  - **Exa Search Tool** (new)
  - **Create Agent** (new)
  - **Durable Execution** (new)
  - **GoSkills Integration** (new)
- **General**: Improved reliability and correctness of all examples.

## [0.1.0] - 2025-01-02

### Added
- Generic state management - works with any type, not just MessageContent
- Performance optimizations for production use
- Support for any LLM client (removed hard dependency on LangChain)

### Changed
- Simplified API for building graphs
- Updated examples to show generic usage

### Fixed
- CI/CD pipeline issues from original repository
- Build errors with recent Go versions

### Removed
- Hard dependency on LangChain - now works with any LLM library