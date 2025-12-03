# LangGraphGo Examples

This directory contains various examples demonstrating the features of LangGraphGo.

## Basic Concepts
- **[Basic Example](basic_example/README.md)**: Simple graph with hardcoded steps.
- **[Basic LLM](basic_llm/README.md)**: Integration with LLMs.
- **[Conditional Routing](conditional_routing/README.md)**: Dynamic routing based on state.
- **[Conditional Edges](conditional_edges_example/README.md)**: Using conditional edges.

## Advanced Features
- **[Parallel Execution](parallel_execution/README.md)**: Fan-out/Fan-in execution with state merging.
- **[Configuration](configuration/README.md)**: Using runtime configuration to pass metadata and settings.
- **[Custom Reducer](custom_reducer/README.md)**: Defining custom state reducers for complex merge logic.
- **[State Schema](state_schema/README.md)**: Managing complex state updates with Schema and Reducers.
- **[Subgraphs](subgraphs/README.md)**: Composing graphs within graphs (New).
- **[Streaming Modes](streaming_modes/README.md)**: Advanced streaming with updates, values, and messages modes.
- **[Smart Messages](smart_messages/README.md)**: Intelligent message merging with ID-based upserts.
- **[Command API](command_api/README.md)**: Dynamic control flow and state updates from nodes.
- **[Ephemeral Channels](ephemeral_channels/README.md)**: Managing temporary state that clears after each step.
- **[Listeners](listeners/README.md)**: Attaching event listeners to the graph.

## Persistence (Checkpointing)
- **[Memory](checkpointing/main.go)**: In-memory checkpointing.
- **[PostgreSQL](checkpointing/postgres/)**: Persistent state using PostgreSQL.
- **[SQLite](checkpointing/sqlite/)**: Persistent state using SQLite.
- **[Redis](checkpointing/redis/)**: Persistent state using Redis.
- **[Durable Execution](durable_execution/README.md)**: Crash recovery and resuming execution from checkpoints.

## Human-in-the-loop
- **[Human Approval](human_in_the_loop/README.md)**: Workflow with interrupts and human approval steps.
- **[Time Travel / HITL](time_travel/README.md)**: Inspecting, modifying state history, and forking execution (UpdateState).
- **[Dynamic Interrupt](dynamic_interrupt/README.md)**: Pausing execution from within a node using `graph.Interrupt`.

## Pre-built Agents
- **[Create Agent](create_agent/README.md)**: Easy way to create an agent with options.
- **[Dynamic Skill Agent](dynamic_skill_agent/README.md)**: Agent with dynamic skill discovery and selection.
- **[ReAct Agent](react_agent/README.md)**: Reason and Action agent using tools.
- **[Supervisor](supervisor/README.md)**: Multi-agent orchestration using a supervisor.
- **[Swarm](swarm/README.md)**: Multi-agent collaboration using handoffs.

## Memory
- **[Memory Basic](memory_basic/README.md)**: Basic usage of LangChain memory.
- **[Memory Chatbot](memory_chatbot/README.md)**: Chatbot with memory integration.

## RAG (Retrieval Augmented Generation)
- **[RAG Basic](rag_basic/README.md)**: Basic RAG implementation.
- **[RAG Pipeline](rag_pipeline/README.md)**: Complete RAG pipeline.
- **[RAG Advanced](rag_advanced/README.md)**: Advanced RAG techniques.
- **[RAG Conditional](rag_conditional/README.md)**: Conditional RAG workflow.
- **[RAG with Embeddings](rag_with_embeddings/README.md)**: RAG using embeddings.
- **[RAG with LangChain](rag_with_langchain/README.md)**: RAG using LangChain components.
- **[RAG with VectorStores](rag_langchain_vectorstore_example/README.md)**: RAG using LangChain VectorStores.
- **[RAG with Chroma](rag_chroma_example/README.md)**: RAG using Chroma database.

## Other
- **[Visualization](visualization/README.md)**: Generating Mermaid diagrams for graphs.
- **[LangChain Integration](langchain_example/README.md)**: Using LangChain tools and models.
- **[Tavily Search](tool_tavily/README.md)**: Using Tavily search tool with ReAct agent.
- **[Exa Search](tool_exa/README.md)**: Using Exa search tool with ReAct agent.
- **[GoSkills Integration](goskills_example/README.md)**: Integrating GoSkills as tools for agents.
