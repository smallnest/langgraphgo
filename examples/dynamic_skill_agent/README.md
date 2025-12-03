# Dynamic Skill Agent Example

This example demonstrates how to create an agent that can dynamically discover and select skills based on user input using `langgraphgo` and `goskills`.

## Overview

The agent is configured with a `skillDir`. When it receives a user request, it performs the following steps:
1.  **Discover**: Scans the specified directory for available skills.
2.  **Select**: Uses the LLM to select the most relevant skill for the user's request.
3.  **Execute**: Loads the selected skill's tools and executes the agent logic, which may involve calling the tools.

## Prerequisites

- Go 1.22+
- OpenAI API Key (set as `OPENAI_API_KEY` environment variable)

## How to Run

```bash
export OPENAI_API_KEY="your-api-key"
go run examples/dynamic_skill_agent/main.go
```

The example will:
1.  Create a dummy "hello_world" skill in a `skills` directory.
2.  Initialize the agent with the `skills` directory.
3.  Send a request "Please run the hello world script." to the agent.
4.  The agent will discover the skill, select it, and execute the script.
