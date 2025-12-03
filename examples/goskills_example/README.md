# GoSkills Integration Example

This example demonstrates how to integrate [GoSkills](https://github.com/smallnest/goskills) with LangGraphGo, allowing agents to use skills as tools.

## 1. Background

GoSkills is a framework for creating and managing reusable skills that can execute Python scripts, shell commands, and provide various built-in tools like web search and file operations. By integrating GoSkills with LangGraphGo, you can easily extend your agents with powerful, pre-packaged capabilities without writing custom tool implementations.

## 2. Key Concepts

- **GoSkills**: A skill management framework that packages scripts and tools into reusable skill packages.
- **Skill Package**: A directory containing a `SKILL.md` file with metadata and optional scripts in a `scripts/` subdirectory.
- **SkillsToTools Adapter**: A convenience function (`adapter.SkillsToTools`) that converts GoSkills skill packages into `tools.Tool` interfaces compatible with LangGraphGo.
- **Multi-Skill Support**: The adapter can load and register multiple skills from a directory, aggregating all their tools.

## 3. How It Works

1.  **Load Skills**: Use `goskills.ParseSkillPackages` to scan a directory (e.g., `./skills`) for skill packages.
2.  **Convert to Tools**: For each skill package, call `adapter.SkillsToTools` to convert the skill's capabilities into `tools.Tool` instances.
3.  **Aggregate Tools**: Combine tools from all skills into a single slice.
4.  **Create Agent**: Use `prebuilt.CreateAgent` with the aggregated tools and a system message that includes all skill descriptions.
5.  **Invoke**: Run the agent with a user query. The agent will automatically select and use the appropriate tools from the loaded skills.

## 4. Code Highlights

### Loading All Skills

```go
skillsDir := "skills"
packages, err := goskills.ParseSkillPackages(skillsDir)
if err != nil {
    log.Fatal(err)
}
```

### Converting Skills to Tools

```go
var allTools []tools.Tool
var allSystemMessages strings.Builder

allSystemMessages.WriteString("You are a helpful assistant that can use skills.\n\n")

for _, skill := range packages {
    fmt.Printf("Loading skill: %s\n", skill.Meta.Name)
    skillTools, err := adapter.SkillsToTools(*skill)
    if err != nil {
        log.Printf("Failed to convert skill %s to tools: %v", skill.Meta.Name, err)
        continue
    }
    allTools = append(allTools, skillTools...)
    allSystemMessages.WriteString(fmt.Sprintf("Skill: %s\n%s\n\n", skill.Meta.Name, skill.Body))
}
```

### Creating the Agent

```go
agent, err := prebuilt.CreateAgent(llm, allTools, 
    prebuilt.WithSystemMessage(allSystemMessages.String()))
if err != nil {
    log.Fatal(err)
}
```

## 5. Skill Structure

A basic skill package structure:

```
skills/
└── hello_world/
    ├── SKILL.md          # Skill metadata and description
    └── scripts/
        └── hello.py      # Python script
```

**SKILL.md** example:
```markdown
---
name: hello_world
description: A simple skill that prints a greeting
version: 1.0.0
---

## Overview
This skill demonstrates basic script execution.
```

## 6. Running the Example

```bash
# Ensure you have skills in the ./skills directory
# The example will create a dummy skill if none exists

export OPENAI_API_KEY=your_api_key
go run main.go
```

**Expected Output:**
```text
Loading skill: hello_world
Tool: hello, Description: Run hello.py script
User: Please use the available skill to say hello to the world.
Agent: [Agent uses the hello_world skill and returns the greeting]
```

## 7. Available Skill Tools

GoSkills provides several built-in tool types that can be used in skills:

- **Script Execution**: `run_python_script`, `run_shell_script`
- **Code Execution**: `run_python_code`, `run_shell_code`
- **File Operations**: `read_file`, `write_file`
- **Web Search**: `duckduckgo_search`, `wikipedia_search`, `tavily_search`
- **Web Fetching**: `web_fetch`

## 8. Learn More

- [GoSkills Documentation](https://github.com/smallnest/goskills)
- [LangGraphGo Prebuilt Agents](../../prebuilt/)
- [Create Agent Example](../create_agent/)
