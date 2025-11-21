package main

import (
	"context"
	"fmt"

	"github.com/smallnest/langgraphgo/graph"
	"github.com/tmc/langchaingo/llms"
	"github.com/tmc/langchaingo/llms/openai"
)

func main() {
	// Create LLM using LangChain
	model, err := openai.New()
	if err != nil {
		panic(err)
	}

	g := graph.NewMessageGraph()

	// Add node that uses LangChain LLM
	g.AddNode("generate", func(ctx context.Context, state interface{}) (interface{}, error) {
		messages := state.([]llms.MessageContent)

		// Use LangChain to generate response
		response, err := model.GenerateContent(ctx, messages,
			llms.WithTemperature(0.7),
		)
		if err != nil {
			return nil, err
		}

		// Append AI response to messages
		return append(messages,
			llms.TextParts("ai", response.Choices[0].Content),
		), nil
	})

	g.AddEdge("generate", graph.END)
	g.SetEntryPoint("generate")

	runnable, err := g.Compile()
	if err != nil {
		panic(err)
	}

	ctx := context.Background()
	initialMessages := []llms.MessageContent{
		llms.TextParts("human", "What is 1 + 1?"),
	}

	res, err := runnable.Invoke(ctx, initialMessages)
	if err != nil {
		panic(err)
	}

	messages := res.([]llms.MessageContent)
	fmt.Println("AI Response:", messages[len(messages)-1].Parts[0])
}
