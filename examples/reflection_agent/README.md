# Reflection Agent Example

This example demonstrates the **Reflection Agent** design pattern - an AI agent that iteratively improves its responses through self-reflection and revision.

## What is the Reflection Pattern?

The Reflection pattern is a technique where an AI agent:
1. **Generates** an initial response
2. **Reflects** on its own output, critiquing quality and identifying improvements
3. **Revises** the response based on the reflection
4. **Repeats** until the response is satisfactory or max iterations are reached

This creates a feedback loop that produces higher-quality outputs than single-pass generation.

## How It Works

```
User Query → Generate → Reflect → Satisfied?
                ↑                      ↓ No
                └──── Revise ←─────────┘
                         ↓ Yes
                   Final Response
```

### Workflow Steps

1. **Generate Node**: Creates initial response or revised version
2. **Reflect Node**: Evaluates the response quality and suggests improvements
3. **Routing Logic**: Decides whether to continue improving or finalize
4. **Max Iterations**: Prevents infinite loops

## Key Features

- ✅ **Iterative Improvement**: Multiple rounds of refinement
- ✅ **Self-Critique**: AI reflects on its own outputs
- ✅ **Customizable**: Flexible prompts for different use cases
- ✅ **Separate Models**: Optionally use different models for generation vs reflection
- ✅ **Smart Stopping**: Automatically detects when output is satisfactory

## Installation

```bash
go get github.com/smallnest/langgraphgo
```

## Usage

### Basic Example

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
    // Create LLM
    model, err := openai.New(openai.WithModel("gpt-4"))
    if err != nil {
        log.Fatal(err)
    }

    // Configure Reflection Agent
    config := prebuilt.ReflectionAgentConfig{
        Model:         model,
        MaxIterations: 3,
        Verbose:       true,
    }

    // Create agent
    agent, err := prebuilt.CreateReflectionAgentMap(config)
    if err != nil {
        log.Fatal(err)
    }

    // Prepare query
    initialState := map[string]any{
        "messages": []llms.MessageContent{
            {
                Role:  llms.ChatMessageTypeHuman,
                Parts: []llms.ContentPart{
                    llms.TextPart("Explain the CAP theorem"),
                },
            },
        },
    }

    // Invoke agent
    result, err := agent.Invoke(context.Background(), initialState)
    if err != nil {
        log.Fatal(err)
    }

    // Extract final response
    finalState := result.(map[string]any)
    draft := finalState["draft"].(string)
    fmt.Println(draft)
}
```

### Advanced Configuration

```go
config := prebuilt.ReflectionAgentConfig{
    Model:         generationModel,
    ReflectionModel: reflectionModel,  // Use separate model for reflection
    MaxIterations: 3,
    Verbose:       true,

    // Custom generation prompt
    SystemMessage: "You are an expert technical writer.",

    // Custom reflection prompt
    ReflectionPrompt: `Evaluate for:
    1. Technical accuracy
    2. Clarity and organization
    3. Completeness
    4. Use of examples

    Be specific in your feedback.`,
}
```

## Configuration Options

| Option | Type | Default | Description |
|--------|------|---------|-------------|
| `Model` | `llms.Model` | *required* | LLM for generation and reflection |
| `ReflectionModel` | `llms.Model` | `nil` | Optional separate model for reflection |
| `MaxIterations` | `int` | `3` | Maximum refinement cycles |
| `SystemMessage` | `string` | Default | Prompt for generation step |
| `ReflectionPrompt` | `string` | Default | Prompt for reflection step |
| `Verbose` | `bool` | `false` | Enable detailed logging |

## Example Output

```
🎨 Generating initial response...
📝 Draft generated (1247 chars)

🤔 Reflecting on the response...
💭 Reflection:
**Strengths:**
- Clear definition of CAP theorem
- Good use of examples

**Weaknesses:**
- Missing trade-offs discussion
- Could explain practical implications better

**Suggestions for improvement:**
- Add section on real-world applications
- Include comparison of different database choices

🔄 Revising response (iteration 2)...
📝 Draft generated (1856 chars)

🤔 Reflecting on the response...
💭 Reflection:
**Strengths:**
- Comprehensive and well-structured
- Excellent examples and practical applications
- Clear trade-off discussions

**Weaknesses:**
- No major issues

**Suggestions for improvement:**
- No improvements needed

✅ Response is satisfactory, finalizing

=== Final Response ===
[Improved CAP theorem explanation...]
```

## Use Cases

### 1. Technical Writing
Generate high-quality documentation with multiple rounds of refinement.

```go
config := prebuilt.ReflectionAgentConfig{
    SystemMessage: "You are an expert technical writer creating clear documentation.",
    ReflectionPrompt: "Evaluate for clarity, completeness, examples, and structure.",
}
```

### 2. Code Review
Provide thoughtful, constructive code review feedback.

```go
config := prebuilt.ReflectionAgentConfig{
    SystemMessage: "You are an experienced software engineer providing code review.",
    ReflectionPrompt: "Evaluate for constructiveness, completeness, and professionalism.",
}
```

### 3. Content Creation
Create blog posts, articles, or educational content with iterative improvement.

```go
config := prebuilt.ReflectionAgentConfig{
    SystemMessage: "You are a skilled content creator writing engaging articles.",
    ReflectionPrompt: "Evaluate for engagement, accuracy, structure, and readability.",
}
```

### 4. Problem Solving
Generate well-reasoned solutions to complex problems.

```go
config := prebuilt.ReflectionAgentConfig{
    SystemMessage: "You are a problem-solving expert providing thorough analysis.",
    ReflectionPrompt: "Evaluate for logical reasoning, completeness, and clarity.",
}
```

## Running the Examples

```bash
export OPENAI_API_KEY=your_key
cd examples/reflection_agent
go run main.go
```

The example demonstrates three scenarios:
1. **Basic Reflection**: Explaining CAP theorem
2. **Technical Writing**: Creating API documentation
3. **Code Review**: Reviewing Go code

## State Structure

The agent maintains the following state:

```go
{
    "messages": []llms.MessageContent,    // Conversation history
    "draft": string,                      // Current response draft
    "reflection": string,                 // Latest reflection feedback
    "iteration": int,                     // Current iteration count
    "is_satisfactory": bool,              // Whether response is good enough
}
```

## Advantages

1. **Higher Quality**: Outputs are refined through multiple iterations
2. **Self-Correcting**: Agent identifies and fixes its own mistakes
3. **Consistent**: Follows evaluation criteria systematically
4. **Flexible**: Customizable for different domains and use cases
5. **Transparent**: Verbose mode shows the improvement process

## Comparison with Other Patterns

| Feature | Single-Pass | ReAct | Reflection |
|---------|------------|-------|------------|
| **Quality** | Variable | Good | High |
| **Iterations** | 1 | Variable | Controlled |
| **Self-Critique** | No | No | Yes |
| **Refinement** | No | Limited | Yes |
| **Use Case** | Simple tasks | Tool use | Quality writing |
| **Latency** | Low | Medium | High |

## Best Practices

### 1. Choose Appropriate Max Iterations
- Simple tasks: 2-3 iterations
- Complex tasks: 3-5 iterations
- Critical content: 5+ iterations

### 2. Craft Good Reflection Prompts
Focus on specific evaluation criteria:
```go
ReflectionPrompt: `Evaluate for:
1. [Specific criterion 1]
2. [Specific criterion 2]
3. [Specific criterion 3]

Be concrete and actionable in feedback.`
```

### 3. Use Separate Models Strategically
- Same model: Simpler, more consistent
- Different models: Can leverage different strengths
  - GPT-4 for generation, GPT-3.5 for reflection (cost optimization)
  - Specialized models for domain-specific reflection

### 4. Monitor with Verbose Mode
Enable verbose mode during development to understand the reflection process.

### 5. Set Realistic Iteration Limits
Too few: May not reach quality threshold
Too many: Diminishing returns, higher cost

## Troubleshooting

### Issue: Agent Always Stops After First Iteration

**Solution**: Check reflection prompt. Make sure it's not too lenient. The reflection should identify areas for improvement in early drafts.

### Issue: Agent Reaches Max Iterations Every Time

**Solution**:
1. Check if reflection prompt is too critical
2. Verify the satisfactory detection logic
3. Consider increasing max iterations for complex tasks

### Issue: Reflections Are Too Generic

**Solution**: Provide more specific evaluation criteria in the reflection prompt with concrete examples of what to look for.

## Advanced: Custom Satisfactory Detection

You can modify the `isResponseSatisfactory` function for custom logic:

```go
// In your own implementation, you could:
// - Call LLM to rate response quality (1-10)
// - Use sentiment analysis on reflection
// - Check for specific keywords or patterns
// - Compare draft length with requirements
```

## Next Steps

- Experiment with different prompts for your use case
- Try using different models for generation vs reflection
- Integrate with your application's workflow
- Add custom evaluation metrics
- Combine with other agent patterns

## References

- [Reflection Pattern in LangGraph](https://langchain-ai.github.io/langgraph/concepts/reflection/)
- [Constitutional AI](https://arxiv.org/abs/2212.08073)
- [Self-Refine](https://arxiv.org/abs/2303.17651)
