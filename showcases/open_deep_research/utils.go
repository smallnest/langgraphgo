package main

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/tmc/langchaingo/llms"
)

// TavilySearchResult represents a single search result from Tavily
type TavilySearchResult struct {
	Title   string  `json:"title"`
	URL     string  `json:"url"`
	Content string  `json:"content"`
	Score   float64 `json:"score"`
}

// TavilySearchResponse represents the response from Tavily API
type TavilySearchResponse struct {
	Results []TavilySearchResult `json:"results"`
}

// TavilySearchTool implements web search using Tavily API
type TavilySearchTool struct {
	APIKey string
}

// Name returns the tool name
func (t *TavilySearchTool) Name() string {
	return "tavily_search"
}

// Description returns the tool description
func (t *TavilySearchTool) Description() string {
	return "Search the web for information using Tavily search API. Input should be a search query string."
}

// Call executes the search
func (t *TavilySearchTool) Call(ctx context.Context, input string) (string, error) {
	if t.APIKey == "" {
		return "", fmt.Errorf("TAVILY_API_KEY not set")
	}

	// Prepare request
	reqBody := map[string]interface{}{
		"api_key":     t.APIKey,
		"query":       input,
		"max_results": 5,
	}

	jsonData, err := json.Marshal(reqBody)
	if err != nil {
		return "", fmt.Errorf("failed to marshal request: %w", err)
	}

	// Make HTTP request
	resp, err := http.Post(
		"https://api.tavily.com/search",
		"application/json",
		strings.NewReader(string(jsonData)),
	)
	if err != nil {
		return "", fmt.Errorf("search request failed: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("failed to read response: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("search API returned status %d: %s", resp.StatusCode, string(body))
	}

	// Parse response
	var searchResp TavilySearchResponse
	if err := json.Unmarshal(body, &searchResp); err != nil {
		return "", fmt.Errorf("failed to parse response: %w", err)
	}

	// Format results
	var results []string
	for i, result := range searchResp.Results {
		results = append(results, fmt.Sprintf(
			"[Result %d]\nTitle: %s\nURL: %s\nContent: %s\n",
			i+1, result.Title, result.URL, result.Content,
		))
	}

	return strings.Join(results, "\n---\n"), nil
}

// ThinkToolImpl implements the reflection/thinking tool
type ThinkToolImpl struct{}

// Name returns the tool name
func (t *ThinkToolImpl) Name() string {
	return "think_tool"
}

// Description returns the tool description
func (t *ThinkToolImpl) Description() string {
	return "Use this tool to reflect on your progress and plan next steps. Input should be your reflection."
}

// Call executes the thinking/reflection
func (t *ThinkToolImpl) Call(ctx context.Context, input string) (string, error) {
	return fmt.Sprintf("Reflection recorded: %s", input), nil
}

// Helper functions

// GetMessagesString converts messages to a string representation
func GetMessagesString(messages []llms.MessageContent) string {
	var parts []string
	for _, msg := range messages {
		role := string(msg.Role)
		var content string
		for _, part := range msg.Parts {
			if textPart, ok := part.(llms.TextContent); ok {
				content += textPart.Text
			}
		}
		parts = append(parts, fmt.Sprintf("%s: %s", role, content))
	}
	return strings.Join(parts, "\n")
}

// GetLastAIMessage returns the last AI message from the message list
func GetLastAIMessage(messages []llms.MessageContent) *llms.MessageContent {
	for i := len(messages) - 1; i >= 0; i-- {
		if messages[i].Role == llms.ChatMessageTypeAI {
			return &messages[i]
		}
	}
	return nil
}

// HasToolCalls checks if a message has tool calls
func HasToolCalls(msg llms.MessageContent) bool {
	for _, part := range msg.Parts {
		if _, ok := part.(llms.ToolCall); ok {
			return true
		}
	}
	return false
}

// GetToolCalls extracts tool calls from a message
func GetToolCalls(msg llms.MessageContent) []llms.ToolCall {
	var toolCalls []llms.ToolCall
	for _, part := range msg.Parts {
		if tc, ok := part.(llms.ToolCall); ok {
			toolCalls = append(toolCalls, tc)
		}
	}
	return toolCalls
}

// ExtractToolCallsByName filters tool calls by name
func ExtractToolCallsByName(toolCalls []llms.ToolCall, name string) []llms.ToolCall {
	var filtered []llms.ToolCall
	for _, tc := range toolCalls {
		if tc.FunctionCall.Name == name {
			filtered = append(filtered, tc)
		}
	}
	return filtered
}

// CreateToolMessage creates a tool response message
func CreateToolMessage(toolCallID, name, content string) llms.MessageContent {
	return llms.MessageContent{
		Role: llms.ChatMessageTypeTool,
		Parts: []llms.ContentPart{
			llms.ToolCallResponse{
				ToolCallID: toolCallID,
				Name:       name,
				Content:    content,
			},
		},
	}
}

// JoinNotes joins a list of notes into a single string
func JoinNotes(notes []string) string {
	return strings.Join(notes, "\n\n")
}

// GetTodayString returns today's date as a string
func GetTodayString() string {
	return time.Now().Format("2006-01-02")
}

// ParseToolArguments parses tool call arguments
func ParseToolArguments(tc llms.ToolCall) (map[string]interface{}, error) {
	var args map[string]interface{}
	if err := json.Unmarshal([]byte(tc.FunctionCall.Arguments), &args); err != nil {
		return nil, fmt.Errorf("failed to parse tool arguments: %w", err)
	}
	return args, nil
}

// URLEncode encodes a string for use in URLs
func URLEncode(s string) string {
	return url.QueryEscape(s)
}

// TruncateString truncates a string to a maximum length
func TruncateString(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	return s[:maxLen] + "..."
}

// MergeStates merges two state maps
func MergeStates(base, update map[string]interface{}) map[string]interface{} {
	result := make(map[string]interface{})

	// Copy base
	for k, v := range base {
		result[k] = v
	}

	// Apply updates
	for k, v := range update {
		result[k] = v
	}

	return result
}

// AppendMessages appends new messages to existing messages
func AppendMessages(existing, new []llms.MessageContent) []llms.MessageContent {
	return append(existing, new...)
}

// AppendNotes appends new notes to existing notes
func AppendNotes(existing, new []string) []string {
	return append(existing, new...)
}
