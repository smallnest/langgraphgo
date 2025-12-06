package memory

import (
	"context"
	"sync"
)

// SequentialMemory implements the "Keep-It-All" strategy
// Stores complete conversation history in chronological order
// Pros: Perfect recall of all interactions
// Cons: Token costs grow unbounded with conversation length
type SequentialMemory struct {
	messages []*Message
	mu       sync.RWMutex
}

// NewSequentialMemory creates a new sequential memory strategy
func NewSequentialMemory() *SequentialMemory {
	return &SequentialMemory{
		messages: make([]*Message, 0),
	}
}

// AddMessage appends a new message to the conversation history
func (s *SequentialMemory) AddMessage(ctx context.Context, msg *Message) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.messages = append(s.messages, msg)
	return nil
}

// GetContext returns all messages in chronological order
// The query parameter is ignored for sequential memory
func (s *SequentialMemory) GetContext(ctx context.Context, query string) ([]*Message, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	// Return a copy to prevent external modification
	result := make([]*Message, len(s.messages))
	copy(result, s.messages)
	return result, nil
}

// Clear removes all messages from memory
func (s *SequentialMemory) Clear(ctx context.Context) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.messages = make([]*Message, 0)
	return nil
}

// GetStats returns statistics about the sequential memory
func (s *SequentialMemory) GetStats(ctx context.Context) (*Stats, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	totalTokens := 0
	for _, msg := range s.messages {
		totalTokens += msg.TokenCount
	}

	return &Stats{
		TotalMessages:   len(s.messages),
		TotalTokens:     totalTokens,
		ActiveMessages:  len(s.messages),
		ActiveTokens:    totalTokens,
		CompressionRate: 1.0, // No compression
	}, nil
}
