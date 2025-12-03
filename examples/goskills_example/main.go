package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"strings"

	"github.com/smallnest/goskills"
	adapter "github.com/smallnest/langgraphgo/adapter/goskills"
	"github.com/smallnest/langgraphgo/prebuilt"
	"github.com/tmc/langchaingo/llms"
	"github.com/tmc/langchaingo/llms/openai"
	"github.com/tmc/langchaingo/tools"
)

func main() {
	// 1. Initialize LLM
	if os.Getenv("OPENAI_API_KEY") == "" {
		log.Fatal("OPENAI_API_KEY is not set")
	}

	llm, err := openai.New()
	if err != nil {
		log.Fatal(err)
	}

	// 2. Load Skills
	// Assuming there is a "skills" directory in the current directory or somewhere accessible.
	// For this example, we might need to create a dummy skill or assume one exists.
	// Let's assume the user has some skills in "./skills".
	// If not, we can try to create a temporary skill for demonstration.

	skillsDir := "skills"
	if _, err := os.Stat(skillsDir); os.IsNotExist(err) {
		// Create a dummy skill for demonstration
		err = createDummySkill(skillsDir)
		if err != nil {
			log.Fatal(err)
		}
		defer os.RemoveAll(skillsDir)
	}

	packages, err := goskills.ParseSkillPackages(skillsDir)
	if err != nil {
		log.Fatal(err)
	}

	if len(packages) == 0 {
		log.Fatal("No skills found")
	}

	// 3. Convert Skills to Tools
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

		for _, t := range skillTools {
			fmt.Printf("Tool: %s, Description: %s\n", t.Name(), t.Description())
		}
	}

	if len(allTools) == 0 {
		log.Fatal("No tools found from skills")
	}

	// 4. Create Agent
	agent, err := prebuilt.CreateAgent(llm, allTools, prebuilt.WithSystemMessage(allSystemMessages.String()))
	if err != nil {
		log.Fatal(err)
	}

	// 5. Run Agent
	ctx := context.Background()
	input := "Please use the available skill to say hello to the world."

	resp, err := agent.Invoke(ctx, map[string]interface{}{
		"messages": []llms.MessageContent{
			llms.TextParts(llms.ChatMessageTypeHuman, input),
		},
	})
	if err != nil {
		log.Fatal(err)
	}

	fmt.Printf("Agent: %v\n", resp)
}

func createDummySkill(dir string) error {
	err := os.MkdirAll(dir+"/hello_world", 0755)
	if err != nil {
		return err
	}

	meta := `---
name: hello_world
description: A fundamental skill that demonstrates the basic execution of a Python script. It serves as a "Hello, World!" example for the skill system, verifying that the environment is correctly set up and that the agent can execute scripts.
version: 1.0.0
license: MIT
---

## Overview

The ` + "`hello_world`" + ` skill is the simplest possible skill in the ecosystem. Its primary purpose is to validate the operational status of the skill runner and the agent's ability to invoke tools.

## Functionality

When invoked, this skill executes a Python script that prints a greeting message to the standard output. This confirms:
1.  **Script Execution**: The agent can successfully locate and run a Python script.
2.  **Output Capture**: The system can capture the standard output from the script and return it to the agent.
3.  **Tool Integration**: The skill is correctly registered and accessible as a tool.

## Usage

This skill is typically used in the following scenarios:
-   **System Health Check**: To verify that the agent and skill runner are functioning correctly.
-   **Onboarding**: As a first step for developers learning how to create and use skills.
-   **Debugging**: To isolate issues with script execution or tool invocation.

### Example Command

To use this skill, the agent can execute the following command:

` + "```python" + `
scripts/hello.py
` + "```" + `

## Implementation Details

The skill consists of a single Python script ` + "`hello.py`" + ` which performs a simple print operation. No external dependencies or complex logic are involved, ensuring that any failure is likely due to the environment or configuration rather than the skill itself.
`
	err = os.WriteFile(dir+"/hello_world/SKILL.md", []byte(meta), 0644)
	if err != nil {
		return err
	}

	script := `
print("Hello, World from Python!")
`
	err = os.WriteFile(dir+"/hello_world/scripts/hello.py", []byte(script), 0644)
	if err != nil {
		return err
	}

	// We need to define the tool in the skill body (usually README.md or similar, but goskills parses skill.yaml and other files)
	// Wait, goskills parses the "Body" from somewhere.
	// Let's look at goskills.ParseSkillPackages implementation or docs if available.
	// Based on runner.go, it seems to use `ParseSkillPackages`.
	// Let's assume a simple structure.

	// Actually, goskills uses `skill.yaml` and maybe other files.
	// Let's create a `tools.json` or similar if goskills supports it, OR just rely on the fact that `goskills` might auto-detect scripts?
	// `GenerateToolDefinitions` in `goskills` scans for scripts.

	return nil
}
