# Memory-Powered Agent Example

This example demonstrates how to integrate memory strategies with real AI agents, showing practical applications in different scenarios.

## Overview

A simple chat agent that uses different memory strategies to handle conversations. This example shows:

- **How to integrate memory strategies** into an agent
- **Real conversation scenarios** demonstrating strategy differences
- **Practical comparison** of memory behavior
- **Interactive mode** to experiment with strategies

## What This Example Shows

Unlike the basic memory_strategies example that just demonstrates the APIs, this example shows **real agent behavior** with memory:

1. **Agent remembers** user information (names, preferences)
2. **Agent recalls** previous topics and context
3. **Different strategies** affect what the agent "remembers"
4. **Practical scenarios** like customer support and consultations

## Running the Example

### Automated Demonstrations

Run all scenarios to see different strategies in action:

```bash
cd examples/memory_agent
go run main.go
```

This will run 5 different scenarios:
1. **Customer Support** - Using Sliding Window
2. **Long Consultation** - Using Summarization
3. **Knowledge Base** - Using Retrieval
4. **Important Info Tracking** - Using Hierarchical
5. **Topic Relationships** - Using Graph-Based

### Interactive Mode

Chat with the agent using your chosen strategy:

```bash
go run main.go interactive
```

You can choose from 9 strategies and have a real conversation with the agent.

## Scenarios Explained

### 1. Customer Support (Sliding Window)

```
Strategy: Sliding Window (keeps last 5 messages)
Use case: Recent conversation context is most important

User: Hello!
Agent: Hello! I'm your assistant. How can I help you today?

User: What's the price of your product?
Agent: Our premium product is priced at $99, which includes free shipping!

User: Can you remind me of the price again?
Agent: As I mentioned before, the product is priced at $99.
```

**Why Sliding Window?**
- Customer support needs recent context
- Keeps conversation focused
- Predictable memory usage

### 2. Long Consultation (Summarization)

```
Strategy: Summarization (summarizes old, keeps recent)
Use case: Long sessions where history matters

User: Hi, I'm John
Agent: Nice to meet you, John! I'll remember your name.

User: IMPORTANT: I need it to be waterproof
Agent: I understand...

[... many messages later ...]

User: Do you remember my name?
Agent: Of course I remember you, John!
```

**Why Summarization?**
- Long conversations would overflow memory
- Old messages summarized, recent kept in full
- Maintains context awareness

### 3. Knowledge Base (Retrieval)

```
Strategy: Retrieval (retrieves relevant messages)
Use case: Large history, query-driven retrieval

User: What's the price?
Agent: Our premium product is priced at $99...

[... many other topics discussed ...]

User: Tell me about the price again
Agent: As I mentioned before, the product is priced at $99.
```

**Why Retrieval?**
- Finds relevant information from large history
- Query-based memory access
- Efficient for knowledge-intensive conversations

### 4. Important Information (Hierarchical)

```
Strategy: Hierarchical (keeps important + recent)
Use case: Some messages more important than others

User: IMPORTANT: Remember I'm allergic to latex
Agent: I understand...

[... several messages later ...]

User: Any latex in the materials?
Agent: [Remembers the allergy from earlier]

User: Just to confirm - you remember my allergy?
Agent: [Successfully recalls the important information]
```

**Why Hierarchical?**
- Some information is critical (allergies, requirements)
- Separates important from routine messages
- Ensures critical info isn't forgotten

### 5. Topic Relationships (Graph-Based)

```
Strategy: Graph-Based (tracks topic relationships)
Use case: Related topics and cross-references

User: What's the price of the product?
Agent: Our premium product is priced at $99...

User: Does the price include warranty?
Agent: [Connects price and warranty topics]

User: What features justify the price?
Agent: [Connects features and price topics]
```

**Why Graph-Based?**
- Tracks how topics relate to each other
- Better context when topics are interconnected
- Helps agent make connections

## Code Structure

### ChatAgent

The core agent class that integrates memory:

```go
type ChatAgent struct {
    memory   memory.Strategy  // Memory strategy
    strategy string            // Strategy name
    ctx      context.Context
}

func NewChatAgent(strategyName string) *ChatAgent {
    // Creates agent with chosen memory strategy
}

func (a *ChatAgent) ProcessMessage(userMsg string) (string, error) {
    // 1. Add user message to memory
    // 2. Retrieve relevant context
    // 3. Generate response using context
    // 4. Add response to memory
}
```

### Memory Integration Pattern

```go
// 1. Create message
msg := memory.NewMessage("user", userInput)

// 2. Mark important messages
if strings.Contains(userInput, "important") {
    msg.Metadata["importance"] = 0.9
}

// 3. Add to memory
agent.memory.AddMessage(ctx, msg)

// 4. Get relevant context
context, _ := agent.memory.GetContext(ctx, userInput)

// 5. Use context to generate response
response := generateResponse(userInput, context)

// 6. Store response
agent.memory.AddMessage(ctx, memory.NewMessage("assistant", response))
```

## Key Insights

### Memory Affects Agent Behavior

Different strategies make the agent "remember" differently:

| Strategy | Remembers | Forgets | Best For |
|----------|-----------|---------|----------|
| Sequential | Everything | Nothing | Short conversations |
| Sliding Window | Last N turns | Old messages | Customer support |
| Summarization | Summary + recent | Old details | Long consultations |
| Retrieval | Relevant to query | Irrelevant | Knowledge bases |
| Hierarchical | Important + recent | Unimportant old | Critical info tracking |
| Graph | Related topics | Unrelated | Topic navigation |

### Choosing Strategy for Your Agent

**Customer Support Bot**
→ Use **Sliding Window** or **Buffer**
- Recent context is what matters
- Conversations are typically short
- Predictable memory usage

**Consultation/Advisory Bot**
→ Use **Summarization** or **Hierarchical**
- Long sessions need history
- Some info more important
- Can't remember everything verbatim

**Knowledge Base/FAQ Bot**
→ Use **Retrieval** or **Graph-Based**
- Large knowledge base
- Query-driven access
- Need to find relevant info

**Multi-topic Discussion Bot**
→ Use **Graph-Based** or **Hierarchical**
- Topics interconnect
- Importance varies
- Need to track relationships

## Interactive Mode Guide

1. Run: `go run main.go interactive`
2. Choose a strategy (1-9)
3. Chat with the agent
4. Type 'quit' to exit

**Try these experiments:**

1. **Test Memory Limits**
   - Choose Sliding Window (2)
   - Send 10 messages
   - Ask agent to recall first message
   - Notice it's forgotten

2. **Test Important Messages**
   - Choose Hierarchical (6)
   - Say "IMPORTANT: My name is X"
   - Send several other messages
   - Ask "Do you remember my name?"
   - Notice it remembers

3. **Test Retrieval**
   - Choose Retrieval (5)
   - Discuss multiple topics
   - Ask about a specific topic later
   - Notice it retrieves relevant messages

## Extending This Example

### Add Real LLM

Replace the `generateResponse` function with actual LLM calls:

```go
func (a *ChatAgent) generateResponse(input string, context []*memory.Message) string {
    // Build prompt from context
    prompt := buildPrompt(context, input)

    // Call LLM (OpenAI, Anthropic, etc.)
    response := llm.Complete(prompt)

    return response
}
```

### Add Custom Strategies

Create your own memory strategy:

```go
type CustomMemory struct {
    // Your implementation
}

func (c *CustomMemory) AddMessage(ctx context.Context, msg *memory.Message) error {
    // Custom logic
}

func (c *CustomMemory) GetContext(ctx context.Context, query string) ([]*memory.Message, error) {
    // Custom retrieval
}
```

### Integrate with LangGraph

The example includes a `CreateAgentGraph` function showing how to integrate with LangGraph workflows.

## Related Examples

- [memory_strategies](../memory_strategies/) - API demonstrations for all strategies
- [memory_chatbot](../memory_chatbot/) - Basic chatbot with memory
- [memory_basic](../memory_basic/) - Simple memory usage

## Further Reading

- [Memory Package Documentation](../../memory/README.md)
- [Memory Package Documentation (中文)](../../memory/README_CN.md)
- [LangGraph Documentation](../../README.md)
