package memory

import (
	"context"
	"sync"
)

// SlidingWindowMemory maintains only the most recent N messages
// Pros: Bounded context size, prevents unbounded token growth
// Cons: Loses older context, may forget important earlier information
type SlidingWindowMemory struct {
	messages   []*Message
	windowSize int // Maximum number of messages to retain
	mu         sync.RWMutex
}

// NewSlidingWindowMemory creates a new sliding window memory strategy
// windowSize determines how many recent messages to keep
func NewSlidingWindowMemory(windowSize int) *SlidingWindowMemory {
	if windowSize <= 0 {
		windowSize = 10 // Default window size
	}

	return &SlidingWindowMemory{
		messages:   make([]*Message, 0, windowSize),
		windowSize: windowSize,
	}
}

// AddMessage adds a new message, removing oldest if window is full
func (s *SlidingWindowMemory) AddMessage(ctx context.Context, msg *Message) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.messages = append(s.messages, msg)

	// If we exceed window size, remove oldest messages
	if len(s.messages) > s.windowSize {
		// Keep only the most recent windowSize messages
		s.messages = s.messages[len(s.messages)-s.windowSize:]
	}

	return nil
}

// GetContext returns messages within the sliding window
func (s *SlidingWindowMemory) GetContext(ctx context.Context, query string) ([]*Message, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	// Return a copy
	result := make([]*Message, len(s.messages))
	copy(result, s.messages)
	return result, nil
}

// Clear removes all messages from memory
func (s *SlidingWindowMemory) Clear(ctx context.Context) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.messages = make([]*Message, 0, s.windowSize)
	return nil
}

// GetStats returns statistics about the sliding window memory
func (s *SlidingWindowMemory) GetStats(ctx context.Context) (*Stats, error) {
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
		CompressionRate: 1.0,
	}, nil
}

// SetWindowSize updates the window size
// If the new size is smaller than current messages, oldest are removed
func (s *SlidingWindowMemory) SetWindowSize(size int) {
	s.mu.Lock()
	defer s.mu.Unlock()

	if size <= 0 {
		size = 10
	}

	s.windowSize = size

	// Trim if necessary
	if len(s.messages) > size {
		s.messages = s.messages[len(s.messages)-size:]
	}
}

// GetWindowSize returns the current window size
func (s *SlidingWindowMemory) GetWindowSize() int {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.windowSize
}
