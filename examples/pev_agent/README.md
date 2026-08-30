# PEV Agent Example

## Overview

This example demonstrates the **PEV (Plan, Execute, Verify)** agent pattern, a robust and self-correcting architecture for reliable task execution. PEV is particularly valuable in high-stakes automation scenarios where accuracy and reliability are critical.

## What is PEV?

PEV is an agentic architecture that implements a three-phase workflow with built-in error detection and recovery:

1. **Plan**: Break down the user's request into concrete, executable steps
2. **Execute**: Run each step using available tools
3. **Verify**: Check if the execution was successful

If verification fails, the agent triggers a re-planning cycle, creating an improved plan based on the failure context. This creates a "quality assurance checkpoint" that ensures only valid data flows into the final synthesis phase.

## Architecture

```
┌─────────────────────────────────────────────────────────────┐
│                        PEV Workflow                         │
└─────────────────────────────────────────────────────────────┘

   User Request
        ↓
   ┌─────────┐
   │ Planner │──────┐ (Re-plan on failure)
   └─────────┘      │
        ↓           │
   ┌──────────┐     │
   │ Executor │     │
   └──────────┘     │
        ↓           │
   ┌──────────┐     │
   │ Verifier │     │
   └──────────┘     │
        ↓           │
   Verification     │
   Successful?      │
        ├─No───────┘
        │
       Yes
        ↓
   ┌─────────────┐
   │ Synthesizer │
   └─────────────┘
        ↓
   Final Answer
```

## Key Features

- **Self-Correction**: Automatically retries failed operations with improved plans
- **Verification at Each Step**: Every execution is validated before proceeding
- **Error Recovery**: Learns from failures to create better plans
- **Configurable Retries**: Control the maximum number of retry attempts
- **Tool Agnostic**: Works with any tools that implement the `tools.Tool` interface

## When to Use PEV

PEV is ideal for:

- **High-Stakes Automation**: Financial systems, healthcare, legal processing
- **Unreliable External Tools**: APIs with intermittent failures, network issues
- **Complex Multi-Step Tasks**: Tasks requiring multiple tool calls in sequence
- **Quality-Critical Applications**: Where accuracy must be verified before proceeding

## State Schema

The PEV agent maintains the following state:

```go
{
    "messages": []llms.MessageContent,      // Conversation history
    "plan": []string,                       // Current execution plan
    "current_step": int,                    // Index of current step
    "last_tool_result": string,             // Result from last tool execution
    "intermediate_steps": []string,         // History of all steps
    "retries": int,                         // Current retry count
    "verification_result": VerificationResult, // Last verification outcome
    "final_answer": string,                 // Synthesized final response
}
```

## Configuration

```go
type PEVAgentConfig struct {
    Model              llms.Model    // LLM for planning and verification
    Tools              []tools.Tool  // Available tools for execution
    MaxRetries         int           // Maximum retry attempts (default: 3)
    SystemMessage      string        // Custom planner prompt (optional)
    VerificationPrompt string        // Custom verifier prompt (optional)
    Verbose            bool          // Enable detailed logging
}
```

## Examples

### Example 1: Simple Calculation

Demonstrates basic PEV operation with a reliable calculator tool:

```go
config := prebuilt.PEVAgentConfig{
    Model:      model,
    Tools:      []tools.Tool{CalculatorTool{}},
    MaxRetries: 3,
    Verbose:    true,
}

agent, err := prebuilt.CreatePEVAgentMap(config)
```

**Query**: "Calculate the result of 15 multiplied by 8"

**Expected Flow**:
1. Plan: "Multiply 15 by 8"
2. Execute: Use calculator tool → "120.00"
3. Verify: ✅ Success
4. Synthesize: "The result is 120"

### Example 2: Unreliable Weather API

Demonstrates self-correction with a tool that has 40% failure rate:

```go
config := prebuilt.PEVAgentConfig{
    Model: model,
    Tools: []tools.Tool{
        WeatherTool{FailureRate: 0.4}, // 40% chance of failure
    },
    MaxRetries: 3,
    Verbose:    true,
}
```

**Query**: "What's the weather like in Tokyo?"

**Possible Flow**:
1. Plan: "Get weather for Tokyo"
2. Execute: Call weather API → "Error: Connection timeout"
3. Verify: ❌ Failure detected
4. Re-plan: "Retry weather query for Tokyo with proper city name"
5. Execute: Call weather API → "Weather in Tokyo: 22°C, Sunny"
6. Verify: ✅ Success
7. Synthesize: "The weather in Tokyo is 22°C and sunny"

### Example 3: Multi-Step Task

Demonstrates PEV with multiple steps and different tools:

```go
config := prebuilt.PEVAgentConfig{
    Model: model,
    Tools: []tools.Tool{
        CalculatorTool{},
        WeatherTool{FailureRate: 0.2},
        DatabaseTool{FailureRate: 0.3},
    },
    MaxRetries: 3,
    Verbose:    true,
}
```

**Query**: "First, calculate 25 times 4. Then, check the weather in Paris."

## Running the Example

1. Set your OpenAI API key:
```bash
export OPENAI_API_KEY=your-api-key-here
```

2. Run the example:
```bash
cd examples/pev_agent
go run main.go
```

## Implementation Details

### Planner Node

- Analyzes user request and breaks it into steps
- On failure, receives verification feedback to create improved plan
- Returns numbered list of actionable steps

### Executor Node

- Executes current step using appropriate tool
- Handles tool errors gracefully
- Returns execution result for verification

### Verifier Node

- Uses LLM to analyze execution results
- Returns structured verification result:
  ```go
  type VerificationResult struct {
      IsSuccessful bool   `json:"is_successful"`
      Reasoning    string `json:"reasoning"`
  }
  ```
- Looks for indicators of success/failure in tool output

### Synthesizer Node

- Combines all successful intermediate steps
- Generates coherent final answer
- Invoked only after all steps succeed or max retries reached

## Comparison with Other Patterns

| Pattern | Self-Correction | Verification | Use Case |
|---------|----------------|--------------|----------|
| **ReAct** | No | No | Quick, simple tasks |
| **Reflection** | Yes | Post-generation | Content quality improvement |
| **PEV** | Yes | Per-step | Reliable execution with tools |

## Best Practices

1. **Tool Design**: Create tools that return clear error messages
2. **Retry Limits**: Set `MaxRetries` based on tool reliability
3. **Verbose Mode**: Enable during development to understand failures
4. **Verification Prompt**: Customize for domain-specific validation
5. **Step Granularity**: Keep steps atomic and independently verifiable

## Limitations

- **Increased Latency**: Multiple verification steps add processing time
- **Higher Cost**: More LLM calls compared to simpler patterns
- **Complexity**: Overkill for simple, reliable operations

## Trade-offs

**Advantages**:
- High reliability and accuracy
- Automatic error recovery
- Clear execution audit trail

**Disadvantages**:
- More expensive (additional LLM calls)
- Slower than single-pass patterns
- Requires well-designed tools with clear success/failure indicators


## License

This implementation is part of the langgraphgo project.
