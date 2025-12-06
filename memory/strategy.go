package memory

import (
	"context"
	"time"
)

// Message represents a single conversation message
type Message struct {
	ID        string                 // Unique identifier
	Role      string                 // "user", "assistant", "system"
	Content   string                 // Message content
	Timestamp time.Time              // When the message was created
	Metadata  map[string]interface{} // Additional metadata
	TokenCount int                   // Approximate token count
}

// Strategy defines the interface for memory management strategies
// All memory strategies must implement these methods
type Strategy interface {
	// AddMessage adds a new message to memory
	AddMessage(ctx context.Context, msg *Message) error

	// GetContext retrieves relevant context for the current conversation
	// Returns messages that should be included in the LLM prompt
	GetContext(ctx context.Context, query string) ([]*Message, error)

	// Clear removes all messages from memory
	Clear(ctx context.Context) error

	// GetStats returns statistics about the current memory state
	GetStats(ctx context.Context) (*Stats, error)
}

// Stats contains statistics about memory usage
type Stats struct {
	TotalMessages   int     // Total number of messages stored
	TotalTokens     int     // Total tokens across all messages
	ActiveMessages  int     // Messages currently in active context
	ActiveTokens    int     // Tokens in active context
	CompressionRate float64 // Compression rate (if applicable)
}

// NewMessage creates a new message with the given role and content
func NewMessage(role, content string) *Message {
	return &Message{
		ID:        generateID(),
		Role:      role,
		Content:   content,
		Timestamp: time.Now(),
		Metadata:  make(map[string]interface{}),
		TokenCount: estimateTokens(content),
	}
}

// estimateTokens provides a rough estimate of token count
// In production, use a proper tokenizer like tiktoken
func estimateTokens(text string) int {
	// Rough approximation: ~4 characters per token
	return len(text) / 4
}

// generateID generates a unique ID for a message
func generateID() string {
	return time.Now().Format("20060102150405.000000")
}
