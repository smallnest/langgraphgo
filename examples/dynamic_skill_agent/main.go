package main

import (
	"context"
	"fmt"
	"log"
	"os"

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

	// 2. Setup Skills Directory
	skillsDir := "skills"
	if _, err := os.Stat(skillsDir); os.IsNotExist(err) {
		err = createDummySkill(skillsDir)
		if err != nil {
			log.Fatal(err)
		}
		defer os.RemoveAll(skillsDir)
	}

	// 3. Create Agent with Skill Selection
	// We pass an empty list of initial tools, as we rely on dynamic skill selection
	agent, err := prebuilt.CreateAgent(llm, []tools.Tool{},
		prebuilt.WithSkillDir(skillsDir),
		prebuilt.WithVerbose(true),
		prebuilt.WithSystemMessage("You are a helpful assistant."),
	)
	if err != nil {
		log.Fatal(err)
	}

	// 4. Run Agent
	ctx := context.Background()
	// Input that should trigger the hello_world skill
	input := "Please run the hello world script."

	fmt.Println("User:", input)
	resp, err := agent.Invoke(ctx, map[string]interface{}{
		"messages": []llms.MessageContent{
			llms.TextParts(llms.ChatMessageTypeHuman, input),
		},
	})
	if err != nil {
		log.Fatal(err)
	}

	fmt.Printf("Agent Response: %v\n", resp)
}

func createDummySkill(dir string) error {
	err := os.MkdirAll(dir+"/hello_world", 0755)
	if err != nil {
		return err
	}

	meta := `---
name: hello_world
description: A skill that prints hello world.
version: 1.0.0
---

## Usage

` + "```python" + `
scripts/hello.py
` + "```" + `
`
	err = os.WriteFile(dir+"/hello_world/SKILL.md", []byte(meta), 0644)
	if err != nil {
		return err
	}

	script := `
print("Hello, World from Python Skill!")
`
	err = os.MkdirAll(dir+"/hello_world/scripts", 0755)
	if err != nil {
		return err
	}

	err = os.WriteFile(dir+"/hello_world/scripts/hello.py", []byte(script), 0644)
	if err != nil {
		return err
	}

	return nil
}
