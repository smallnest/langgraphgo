package memory

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/tmc/langchaingo/llms"
	langchainmemory "github.com/tmc/langchaingo/memory"
)

func TestLangChainMemory_ConversationBuffer(t *testing.T) {
	ctx := context.Background()

	// Create a conversation buffer memory with return messages enabled
	mem := NewConversationBufferMemory(
		langchainmemory.WithReturnMessages(true),
	)

	// Test SaveContext
	err := mem.SaveContext(ctx, map[string]any{
		"input": "Hello, my name is Alice",
	}, map[string]any{
		"output": "Hi Alice! Nice to meet you.",
	})
	require.NoError(t, err)

	err = mem.SaveContext(ctx, map[string]any{
		"input": "What's my name?",
	}, map[string]any{
		"output": "Your name is Alice.",
	})
	require.NoError(t, err)

	// Test LoadMemoryVariables
	memVars, err := mem.LoadMemoryVariables(ctx, map[string]any{})
	require.NoError(t, err)
	assert.Contains(t, memVars, "history")

	// Test GetMessages
	messages, err := mem.GetMessages(ctx)
	require.NoError(t, err)
	assert.Len(t, messages, 4) // 2 user messages + 2 AI messages

	// Verify message content
	assert.Equal(t, llms.ChatMessageTypeHuman, messages[0].GetType())
	assert.Equal(t, "Hello, my name is Alice", messages[0].GetContent())
	assert.Equal(t, llms.ChatMessageTypeAI, messages[1].GetType())
	assert.Equal(t, "Hi Alice! Nice to meet you.", messages[1].GetContent())

	// Test Clear
	err = mem.Clear(ctx)
	require.NoError(t, err)

	messages, err = mem.GetMessages(ctx)
	require.NoError(t, err)
	assert.Len(t, messages, 0)
}

func TestLangChainMemory_ConversationWindowBuffer(t *testing.T) {
	ctx := context.Background()

	// Create a conversation window buffer that keeps only the last 2 turns (4 messages)
	mem := NewConversationWindowBufferMemory(2,
		langchainmemory.WithReturnMessages(true),
	)

	// Add 3 conversation turns
	err := mem.SaveContext(ctx, map[string]any{
		"input": "First message",
	}, map[string]any{
		"output": "First response",
	})
	require.NoError(t, err)

	err = mem.SaveContext(ctx, map[string]any{
		"input": "Second message",
	}, map[string]any{
		"output": "Second response",
	})
	require.NoError(t, err)

	err = mem.SaveContext(ctx, map[string]any{
		"input": "Third message",
	}, map[string]any{
		"output": "Third response",
	})
	require.NoError(t, err)

	// Should only keep the last 2 turns (4 messages)
	messages, err := mem.GetMessages(ctx)
	require.NoError(t, err)
	assert.Len(t, messages, 4)

	// Verify it kept the last 2 turns
	assert.Equal(t, "Second message", messages[0].GetContent())
	assert.Equal(t, "Third response", messages[3].GetContent())
}

func TestChatMessageHistory(t *testing.T) {
	ctx := context.Background()

	// Create a new chat message history
	history := NewChatMessageHistory()

	// Test AddUserMessage
	err := history.AddUserMessage(ctx, "Hello!")
	require.NoError(t, err)

	// Test AddAIMessage
	err = history.AddAIMessage(ctx, "Hi there!")
	require.NoError(t, err)

	// Test AddMessage with custom message
	err = history.AddMessage(ctx, llms.SystemChatMessage{
		Content: "You are a helpful assistant.",
	})
	require.NoError(t, err)

	// Test Messages
	messages, err := history.Messages(ctx)
	require.NoError(t, err)
	assert.Len(t, messages, 3)

	assert.Equal(t, llms.ChatMessageTypeHuman, messages[0].GetType())
	assert.Equal(t, "Hello!", messages[0].GetContent())
	assert.Equal(t, llms.ChatMessageTypeAI, messages[1].GetType())
	assert.Equal(t, "Hi there!", messages[1].GetContent())
	assert.Equal(t, llms.ChatMessageTypeSystem, messages[2].GetType())
	assert.Equal(t, "You are a helpful assistant.", messages[2].GetContent())

	// Test Clear
	err = history.Clear(ctx)
	require.NoError(t, err)

	messages, err = history.Messages(ctx)
	require.NoError(t, err)
	assert.Len(t, messages, 0)
}

func TestChatMessageHistory_WithPreviousMessages(t *testing.T) {
	ctx := context.Background()

	// Create history with previous messages
	previousMessages := []llms.ChatMessage{
		llms.HumanChatMessage{Content: "Previous message 1"},
		llms.AIChatMessage{Content: "Previous response 1"},
	}

	history := NewChatMessageHistory(
		langchainmemory.WithPreviousMessages(previousMessages),
	)

	// Verify previous messages are loaded
	messages, err := history.Messages(ctx)
	require.NoError(t, err)
	assert.Len(t, messages, 2)
	assert.Equal(t, "Previous message 1", messages[0].GetContent())
	assert.Equal(t, "Previous response 1", messages[1].GetContent())

	// Add new message
	err = history.AddUserMessage(ctx, "New message")
	require.NoError(t, err)

	messages, err = history.Messages(ctx)
	require.NoError(t, err)
	assert.Len(t, messages, 3)
}

func TestLangChainMemory_CustomKeys(t *testing.T) {
	ctx := context.Background()

	// Create memory with custom input/output keys
	mem := NewConversationBufferMemory(
		langchainmemory.WithInputKey("user_input"),
		langchainmemory.WithOutputKey("ai_output"),
		langchainmemory.WithMemoryKey("chat_history"),
		langchainmemory.WithReturnMessages(true),
	)

	// Save context with custom keys
	err := mem.SaveContext(ctx, map[string]any{
		"user_input": "What's the weather?",
	}, map[string]any{
		"ai_output": "It's sunny today!",
	})
	require.NoError(t, err)

	// Load memory variables
	memVars, err := mem.LoadMemoryVariables(ctx, map[string]any{})
	require.NoError(t, err)
	assert.Contains(t, memVars, "chat_history")

	// Verify messages
	messages, err := mem.GetMessages(ctx)
	require.NoError(t, err)
	assert.Len(t, messages, 2)
	assert.Equal(t, "What's the weather?", messages[0].GetContent())
	assert.Equal(t, "It's sunny today!", messages[1].GetContent())
}

func TestLangChainMemory_WithChatHistory(t *testing.T) {
	ctx := context.Background()

	// Create a custom chat history
	chatHistory := NewChatMessageHistory()
	err := chatHistory.AddUserMessage(ctx, "Initial message")
	require.NoError(t, err)

	// Create memory with the custom chat history
	mem := NewConversationBufferMemory(
		langchainmemory.WithChatHistory(chatHistory.GetHistory()),
		langchainmemory.WithReturnMessages(true),
	)

	// Verify initial message is present
	messages, err := mem.GetMessages(ctx)
	require.NoError(t, err)
	assert.Len(t, messages, 1)
	assert.Equal(t, "Initial message", messages[0].GetContent())

	// Add more messages
	err = mem.SaveContext(ctx, map[string]any{
		"input": "Follow-up message",
	}, map[string]any{
		"output": "Follow-up response",
	})
	require.NoError(t, err)

	messages, err = mem.GetMessages(ctx)
	require.NoError(t, err)
	assert.Len(t, messages, 3)
}
