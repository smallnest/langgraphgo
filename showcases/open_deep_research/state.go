package main

import (
	"github.com/tmc/langchaingo/llms"
)

// Structured output types for tool calls

// ConductResearch represents a research task delegation
type ConductResearch struct {
	ResearchTopic string `json:"research_topic" jsonschema:"description=The topic to research. Should be a single topic described in high detail (at least a paragraph)."`
}

// ResearchComplete indicates research is finished
type ResearchComplete struct {
	Complete bool `json:"complete" jsonschema:"description=Set to true when research is complete"`
}

// ThinkTool represents a reflection/thinking step
type ThinkTool struct {
	Reflection string `json:"reflection" jsonschema:"description=Your reflection on the current state and next steps"`
}

// Summary represents compressed research findings
type Summary struct {
	Summary     string `json:"summary" jsonschema:"description=Concise summary of research findings"`
	KeyExcerpts string `json:"key_excerpts" jsonschema:"description=Important excerpts from the research"`
}

// ClarifyWithUser represents a clarification request
type ClarifyWithUser struct {
	NeedClarification bool   `json:"need_clarification" jsonschema:"description=Whether clarification is needed from the user"`
	Question          string `json:"question" jsonschema:"description=Question to ask the user"`
	Verification      string `json:"verification" jsonschema:"description=Verification message after user provides information"`
}

// ResearchQuestion represents the research brief
type ResearchQuestion struct {
	ResearchBrief string `json:"research_brief" jsonschema:"description=Research question that will guide the research"`
}

// State definitions

// AgentState is the main state for the complete research workflow
type AgentState struct {
	Messages           []llms.MessageContent
	SupervisorMessages []llms.MessageContent
	ResearchBrief      string
	RawNotes           []string
	Notes              []string
	FinalReport        string
}

// SupervisorState manages research task delegation
type SupervisorState struct {
	SupervisorMessages []llms.MessageContent
	ResearchBrief      string
	Notes              []string
	ResearchIterations int
	RawNotes           []string
}

// ResearcherState manages individual research execution
type ResearcherState struct {
	ResearcherMessages []llms.MessageContent
	ToolCallIterations int
	ResearchTopic      string
	CompressedResearch string
	RawNotes           []string
}

// ResearcherOutputState is the output from a researcher
type ResearcherOutputState struct {
	CompressedResearch string
	RawNotes           []string
}
