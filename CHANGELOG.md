# Changelog

## [0.5.0] - 2025-12-06 (In Development)

### Programmatic Tool Calling (PTC)
- **PTC Package**: Added new `ptc` package for programmatic tool calling (#31).
  - LLM generates code that calls tools directly instead of requiring API round-trips
  - Supports both Python and Go code execution
  - Two execution modes: `ModeServer` (HTTP server, default) and `ModeDirect` (subprocess, experimental)
  - Reduces latency and token usage by up to 10x for complex tool chains
  - Multi-LLM support (OpenAI, Gemini, Claude, any langchaingo-compatible model)

### PTC Features
- **Code Executor**: Executes LLM-generated Python/Go code with tool access in sandboxed environment
- **Tool Server**: HTTP server exposing tools via REST API for secure code execution
- **Smart Code Generation**: Automatic tool wrapper generation for both Python and Go
- **Error Handling**: Robust error reporting with execution output and debugging information
- **Documentation**: Complete bilingual documentation (English & Chinese) with Mermaid flow diagrams

### PTC Examples
- **[PTC Basic](./examples/ptc_basic/)**: Introduction to PTC with calculator, weather, and data processing tools
- **[PTC Simple](./examples/ptc_simple/)**: Simple calculator example demonstrating basic PTC usage
- **[PTC Expense Analysis](./examples/ptc_expense_analysis/)**: Complex scenario based on Anthropic PTC Cookbook, showing data filtering and aggregation

### Design Patterns
- **Planning Pattern**: Added planning mode for task decomposition and execution planning (#24)
- **Reflection Agent**: Implemented reflection-action loop pattern for self-assessment and quality improvement (#32)

### Showcases & Documentation
- **GPT Researcher**: Complete replication of assafelovic/gpt-researcher (#34)
  - Automated research and report generation
  - Multi-source information integration
- **Trading Agents**: Merged documentation files for comprehensive README (#39)
  - Integrated PROJECT_SUMMARY.md and USAGE.md into README.md
  - Added detailed usage guide, verbose mode examples, and API reference
- **Open Deep Research**: Merged WORKFLOW.md into README files (#38)
  - Added 5 detailed Mermaid workflow diagrams
  - Included key concepts: state accumulation, message sequence, parallel execution
- **Health Insights Agent**: Merged PROJECT_SUMMARY_CN.md into README_CN.md (#37)
  - Added technical architecture, performance metrics, and security considerations
- **DeepAgents**: Added comprehensive documentation (#36)
  - Complete tool reference and best practices guide
- **DeerFlow & BettaFish**: Updated documentation for both showcases (#35)

### Agent Documentation
- **CreateAgent & CreateReactAgent**: Added comprehensive comparison documentation (#33)
  - Detailed API reference and usage examples
  - Best practices and use case guidance

### Website & Knowledge Base
- **Official Website**: http://lango.rpcx.io (source: https://github.com/smallnest/lango-website)
  - 233 HTML pages with bilingual support
  - 16+ detailed guides (Getting Started, Advanced Features, State Management, etc.)
  - Showcase gallery with 6 complete projects
  - Examples page with 20+ code examples
- **Wiki Knowledge Base**: 193 Markdown documents covering:
  - Advanced features (HITL, Visualization, Subgraphs, Parallel Execution)
  - Checkpoint storage (SQLite, Redis, PostgreSQL)
  - Tool integration guides
  - Pre-built components and RAG guides

### Documentation Consolidation
- Simplified documentation structure with clearer naming conventions
- Merged scattered documentation into comprehensive README files
- Improved navigation and discoverability across all showcases

## [0.4.0] - 2025-12-04

### Core & Agents
- **MCP Support**: Added support for Model Context Protocol (MCP) (#21).
- **Skills Integration**:
  - Added support for **Claude Skills** (#20).
  - Updated `CreateAgent` to support dynamic skill loading (#20).
- **LLM Providers**: Added support for alternative OpenAI-compatible LLM providers in BettaFish showcase.

### Tools
- **Search Tools**:
  - Added **Brave Search** API support.
  - Added **Bocha Search** tool (#22).

### Showcases
- **DeerFlow**: Updated DeerFlow showcase.
- **BettaFish**: Added new showcase replicating BettaFish (https://github.com/666ghj/BettaFish) (#19).

### Documentation & Website
- **Website**: Moved website content to https://github.com/smallnest/lango-website.
- **DIFF.md**: Added DIFF.md for showcases (#19).

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
  - **Dynamic Skill Agent** (new)
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