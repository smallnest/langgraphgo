package main

import (
	"context"
	"fmt"
	"log"

	"github.com/smallnest/langgraphgo/graph"
	"github.com/tmc/langchaingo/llms"
	"github.com/tmc/langchaingo/llms/openai"
)

// CreateDeepResearcherGraph creates the complete deep research workflow
func CreateDeepResearcherGraph(config *Configuration) (*graph.StateRunnable, error) {
	// Initialize OpenAI model
	model, err := openai.New(
		openai.WithModel(config.ResearchModel),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create model: %w", err)
	}

	// Create researcher subgraph
	researcherGraph, err := CreateResearcherGraph(config, model)
	if err != nil {
		return nil, fmt.Errorf("failed to create researcher graph: %w", err)
	}

	// Create supervisor subgraph
	supervisorGraph, err := CreateSupervisorGraph(config, model, researcherGraph)
	if err != nil {
		return nil, fmt.Errorf("failed to create supervisor graph: %w", err)
	}

	// Create main workflow
	workflow := graph.NewStateGraph()

	// Define state schema
	schema := graph.NewMapSchema()
	schema.RegisterReducer("messages", graph.AppendReducer)
	schema.RegisterReducer("supervisor_messages", graph.AppendReducer)
	schema.RegisterReducer("notes", graph.AppendReducer)
	schema.RegisterReducer("raw_notes", graph.AppendReducer)
	workflow.SetSchema(schema)

	// Initialize research node - creates research brief
	workflow.AddNode("init_research", func(ctx context.Context, state interface{}) (interface{}, error) {
		mState := state.(map[string]interface{})
		messages, ok := mState["messages"].([]llms.MessageContent)
		if !ok || len(messages) == 0 {
			return nil, fmt.Errorf("no messages in state")
		}

		// Get user's query from the first message
		userQuery := ""
		for _, part := range messages[0].Parts {
			if textPart, ok := part.(llms.TextContent); ok {
				userQuery = textPart.Text
				break
			}
		}

		if userQuery == "" {
			return nil, fmt.Errorf("could not extract user query")
		}

		// Create research brief (simplified - in production, could use LLM to refine)
		researchBrief := fmt.Sprintf("Research the following topic: %s", userQuery)

		log.Printf("[Init Research] Research brief created: %s", researchBrief)

		return map[string]interface{}{
			"research_brief":      researchBrief,
			"supervisor_messages": []llms.MessageContent{},
			"notes":               []string{},
			"raw_notes":           []string{},
			"research_iterations": 0,
		}, nil
	})

	// Supervisor subgraph node
	workflow.AddNode("supervisor", func(ctx context.Context, state interface{}) (interface{}, error) {
		mState := state.(map[string]interface{})

		// Compile supervisor graph
		supervisorRunnable, err := supervisorGraph.Compile()
		if err != nil {
			return nil, fmt.Errorf("failed to compile supervisor: %w", err)
		}

		// Prepare supervisor state
		supervisorState := map[string]interface{}{
			"supervisor_messages": mState["supervisor_messages"],
			"research_brief":      mState["research_brief"],
			"notes":               mState["notes"],
			"raw_notes":           mState["raw_notes"],
			"research_iterations": mState["research_iterations"],
		}

		// Invoke supervisor
		result, err := supervisorRunnable.Invoke(ctx, supervisorState)
		if err != nil {
			return nil, fmt.Errorf("supervisor execution failed: %w", err)
		}

		resultState := result.(map[string]interface{})

		return map[string]interface{}{
			"supervisor_messages": resultState["supervisor_messages"],
			"notes":               resultState["notes"],
			"raw_notes":           resultState["raw_notes"],
			"research_iterations": resultState["research_iterations"],
		}, nil
	})

	// Final report generation node
	workflow.AddNode("final_report", func(ctx context.Context, state interface{}) (interface{}, error) {
		mState := state.(map[string]interface{})

		researchBrief, _ := mState["research_brief"].(string)
		notes, _ := mState["notes"].([]string)
		messages, _ := mState["messages"].([]llms.MessageContent)

		if len(notes) == 0 {
			return map[string]interface{}{
				"final_report": "No research findings available to generate report.",
				"messages": []llms.MessageContent{
					llms.TextParts(llms.ChatMessageTypeAI, "No research findings available to generate report."),
				},
			}, nil
		}

		// Create final report prompt
		userMessages := GetMessagesString(messages)
		findings := JoinNotes(notes)
		prompt := GetFinalReportPrompt(researchBrief, userMessages, findings)

		log.Printf("[Final Report] Generating report with %d research findings", len(notes))

		// Generate report using final report model
		reportModel, err := openai.New(
			openai.WithModel(config.FinalReportModel),
		)
		if err != nil {
			return nil, fmt.Errorf("failed to create report model: %w", err)
		}

		resp, err := reportModel.GenerateContent(ctx, []llms.MessageContent{
			llms.TextParts(llms.ChatMessageTypeHuman, prompt),
		}, llms.WithMaxTokens(config.FinalReportModelMaxTokens))

		if err != nil {
			return nil, fmt.Errorf("report generation failed: %w", err)
		}

		finalReport := resp.Choices[0].Content

		log.Printf("[Final Report] Report generated successfully (%d characters)", len(finalReport))

		return map[string]interface{}{
			"final_report": finalReport,
			"messages": []llms.MessageContent{
				llms.TextParts(llms.ChatMessageTypeAI, finalReport),
			},
			"notes": []string{}, // Clear notes after final report
		}, nil
	})

	// Define workflow edges
	workflow.SetEntryPoint("init_research")
	workflow.AddEdge("init_research", "supervisor")
	workflow.AddEdge("supervisor", "final_report")
	workflow.AddEdge("final_report", graph.END)

	// Compile the complete workflow
	return workflow.Compile()
}
