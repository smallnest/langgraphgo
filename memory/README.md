# Memory Management Strategies

This package provides various memory management strategies for AI agents, optimized for different use cases and token efficiency.

## Overview

Memory management is crucial for AI agents to maintain context while controlling token costs. This package implements multiple strategies based on research in optimizing AI agent memory.

## Strategies

### 1. Sequential Memory (Keep-It-All)

**Use Case**: When you need perfect recall and token cost is not a concern

**Pros**:
- Perfect recall of all interactions
- Simple implementation
- No information loss

**Cons**:
- Unbounded token growth
- Can become very expensive
- No optimization

**Example**:
```go
mem := memory.NewSequentialMemory()

msg := memory.NewMessage("user", "Hello, AI!")
mem.AddMessage(ctx, msg)

response := memory.NewMessage("assistant", "Hello! How can I help?")
mem.AddMessage(ctx, response)

// Get all messages
messages, _ := mem.GetContext(ctx, "")
```

### 2. Sliding Window Memory

**Use Case**: Maintaining recent conversation context with bounded size

**Pros**:
- Prevents unbounded growth
- Maintains recent conversation flow
- Simple and predictable

**Cons**:
- Loses older context
- May forget important earlier information

**Example**:
```go
// Keep only last 10 messages
mem := memory.NewSlidingWindowMemory(10)

for i := 0; i < 20; i++ {
    msg := memory.NewMessage("user", fmt.Sprintf("Message %d", i))
    mem.AddMessage(ctx, msg)
}

// Only last 10 messages retained
messages, _ := mem.GetContext(ctx, "")
```

### 3. Summarization-Based Memory

**Use Case**: Long conversations where historical context matters but needs compression

**Pros**:
- Maintains historical awareness
- Reduces token consumption
- Preserves important information

**Cons**:
- Requires LLM calls for summarization
- May lose specific details
- Summary quality depends on LLM

**Example**:
```go
mem := memory.NewSummarizationMemory(&memory.SummarizationConfig{
    RecentWindowSize: 10,   // Keep last 10 messages full
    SummarizeAfter:   20,   // Summarize when exceeds 20
    Summarizer: func(ctx context.Context, messages []*Message) (string, error) {
        // Call your LLM to generate summary
        return llm.Summarize(messages)
    },
})

// As messages accumulate, older ones are automatically summarized
```

### 4. Retrieval-Based Memory

**Use Case**: Large conversation histories where only relevant context is needed

**Pros**:
- Highly efficient token usage
- Retrieves only relevant information
- Scales well with large histories

**Cons**:
- Requires embedding model
- May miss chronologically important context
- Additional latency for embedding generation

**Example**:
```go
mem := memory.NewRetrievalMemory(&memory.RetrievalConfig{
    TopK: 5, // Retrieve top 5 most relevant messages
    EmbeddingFunc: func(ctx context.Context, text string) ([]float64, error) {
        // Call embedding API (e.g., OpenAI embeddings)
        return openai.CreateEmbedding(text)
    },
})

// Add many messages
for _, msg := range manyMessages {
    mem.AddMessage(ctx, msg)
}

// Retrieve only relevant ones
relevantMessages, _ := mem.GetContext(ctx, "Tell me about pricing")
```

### 5. Hierarchical Memory

**Use Case**: Complex conversations with varying importance levels

**Pros**:
- Balances recency and importance
- Flexible prioritization
- Maintains critical information

**Cons**:
- More complex management
- Requires importance scoring
- Higher implementation complexity

**Example**:
```go
mem := memory.NewHierarchicalMemory(&memory.HierarchicalConfig{
    RecentLimit:    10,  // Recent messages
    ImportantLimit: 20,  // Important messages
    ImportanceScorer: func(msg *Message) float64 {
        // Custom scoring logic
        if strings.Contains(msg.Content, "IMPORTANT") {
            return 0.9
        }
        return 0.5
    },
})

// Mark important messages
importantMsg := memory.NewMessage("user", "IMPORTANT: Remember this rule")
importantMsg.Metadata["importance"] = 0.95
mem.AddMessage(ctx, importantMsg)
```

### 6. Buffer Memory

**Use Case**: General-purpose memory with flexible limits (similar to LangChain)

**Pros**:
- Flexible configuration
- Optional auto-summarization
- Can limit by messages or tokens

**Cons**:
- May need tuning for optimal performance

**Example**:
```go
mem := memory.NewBufferMemory(&memory.BufferConfig{
    MaxMessages:   50,    // Limit to 50 messages
    MaxTokens:     2000,  // Or 2000 tokens, whichever comes first
    AutoSummarize: true,  // Auto-summarize when limits exceeded
})

// Messages automatically managed
```

## Interface

All strategies implement the `Strategy` interface:

```go
type Strategy interface {
    // Add a message to memory
    AddMessage(ctx context.Context, msg *Message) error

    // Get relevant context for current query
    GetContext(ctx context.Context, query string) ([]*Message, error)

    // Clear all memory
    Clear(ctx context.Context) error

    // Get statistics
    GetStats(ctx context.Context) (*Stats, error)
}
```

## Message Structure

```go
type Message struct {
    ID         string                 // Unique identifier
    Role       string                 // "user", "assistant", "system"
    Content    string                 // Message content
    Timestamp  time.Time              // Creation time
    Metadata   map[string]interface{} // Additional metadata
    TokenCount int                    // Estimated tokens
}
```

## Statistics

All strategies provide statistics:

```go
stats, _ := mem.GetStats(ctx)
fmt.Printf("Total Messages: %d\n", stats.TotalMessages)
fmt.Printf("Active Messages: %d\n", stats.ActiveMessages)
fmt.Printf("Total Tokens: %d\n", stats.TotalTokens)
fmt.Printf("Compression Rate: %.2f\n", stats.CompressionRate)
```

## Choosing a Strategy

| Scenario | Recommended Strategy |
|----------|---------------------|
| Short conversations, low cost concern | Sequential |
| Chat with bounded history | Sliding Window |
| Long conversations, need compression | Summarization |
| Large knowledge base, query-driven | Retrieval |
| Complex multi-topic conversations | Hierarchical |
| General purpose, flexible | Buffer |

## Integration Example

```go
// Create your preferred strategy
strategy := memory.NewSlidingWindowMemory(20)

// Add messages as conversation progresses
userMsg := memory.NewMessage("user", "What's the weather?")
strategy.AddMessage(ctx, userMsg)

// Get context for LLM
messages, _ := strategy.GetContext(ctx, "current query")

// Format for your LLM
prompt := formatMessagesForLLM(messages)
response := llm.Generate(prompt)

// Add response to memory
assistantMsg := memory.NewMessage("assistant", response)
strategy.AddMessage(ctx, assistantMsg)
```

## Advanced Usage

### Custom Importance Scorer

```go
scorer := func(msg *Message) float64 {
    score := 0.5

    // Boost system messages
    if msg.Role == "system" {
        score += 0.3
    }

    // Boost messages with keywords
    if strings.Contains(msg.Content, "remember") {
        score += 0.2
    }

    return math.Min(score, 1.0)
}

mem := memory.NewHierarchicalMemory(&memory.HierarchicalConfig{
    ImportanceScorer: scorer,
})
```

### Custom Summarizer

```go
summarizer := func(ctx context.Context, messages []*Message) (string, error) {
    // Use your LLM
    prompt := "Summarize the following conversation:\n\n"
    for _, msg := range messages {
        prompt += fmt.Sprintf("%s: %s\n", msg.Role, msg.Content)
    }

    return llm.Complete(ctx, prompt)
}

mem := memory.NewSummarizationMemory(&memory.SummarizationConfig{
    Summarizer: summarizer,
})
```

### Custom Embeddings

```go
embedder := func(ctx context.Context, text string) ([]float64, error) {
    // Use OpenAI, Cohere, or your embedding model
    return openai.CreateEmbedding(ctx, text)
}

mem := memory.NewRetrievalMemory(&memory.RetrievalConfig{
    EmbeddingFunc: embedder,
})
```

## Testing

Run tests:
```bash
go test ./memory -v
```

## References

- Based on research from [optimize-ai-agent-memory](https://github.com/FareedKhan-dev/optimize-ai-agent-memory)
- Implements patterns similar to LangChain's memory systems
- Optimized for Go and LangGraphGo integration
