package memory

import (
	"context"
	"fmt"
	"strings"
	"sync"
)

// SummarizationMemory condenses older messages into summaries
// Pros: Maintains historical context while reducing token count
// Cons: May lose specific details in summarization
type SummarizationMemory struct {
	recentMessages   []*Message // Recent messages kept verbatim
	summaries        []string   // Condensed summaries of older conversations
	recentWindowSize int        // How many recent messages to keep full
	summarizeAfter   int        // Summarize when recent messages exceed this
	mu               sync.RWMutex

	// Summarizer is a function that takes messages and returns a summary
	// In production, this would call an LLM
	Summarizer func(ctx context.Context, messages []*Message) (string, error)
}

// SummarizationConfig holds configuration for summarization memory
type SummarizationConfig struct {
	RecentWindowSize int                                                       // Number of recent messages to keep
	SummarizeAfter   int                                                       // Trigger summarization after this many messages
	Summarizer       func(ctx context.Context, messages []*Message) (string, error) // Custom summarizer
}

// NewSummarizationMemory creates a new summarization-based memory strategy
func NewSummarizationMemory(config *SummarizationConfig) *SummarizationMemory {
	if config == nil {
		config = &SummarizationConfig{
			RecentWindowSize: 10,
			SummarizeAfter:   20,
		}
	}

	if config.RecentWindowSize <= 0 {
		config.RecentWindowSize = 10
	}
	if config.SummarizeAfter <= 0 {
		config.SummarizeAfter = 20
	}

	summarizer := config.Summarizer
	if summarizer == nil {
		summarizer = defaultSummarizer
	}

	return &SummarizationMemory{
		recentMessages:   make([]*Message, 0),
		summaries:        make([]string, 0),
		recentWindowSize: config.RecentWindowSize,
		summarizeAfter:   config.SummarizeAfter,
		Summarizer:       summarizer,
	}
}

// AddMessage adds a new message and triggers summarization if needed
func (s *SummarizationMemory) AddMessage(ctx context.Context, msg *Message) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.recentMessages = append(s.recentMessages, msg)

	// Check if we need to summarize
	if len(s.recentMessages) > s.summarizeAfter {
		return s.triggerSummarization(ctx)
	}

	return nil
}

// triggerSummarization creates a summary of older messages
// Must be called with lock held
func (s *SummarizationMemory) triggerSummarization(ctx context.Context) error {
	// Determine how many messages to summarize
	toSummarize := len(s.recentMessages) - s.recentWindowSize
	if toSummarize <= 0 {
		return nil
	}

	// Get messages to summarize
	messagesToSummarize := s.recentMessages[:toSummarize]

	// Generate summary
	summary, err := s.Summarizer(ctx, messagesToSummarize)
	if err != nil {
		return fmt.Errorf("summarization failed: %w", err)
	}

	// Store summary and keep only recent messages
	s.summaries = append(s.summaries, summary)
	s.recentMessages = s.recentMessages[toSummarize:]

	return nil
}

// GetContext returns summaries plus recent messages
func (s *SummarizationMemory) GetContext(ctx context.Context, query string) ([]*Message, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	result := make([]*Message, 0)

	// Add summaries as system messages
	for i, summary := range s.summaries {
		summaryMsg := &Message{
			ID:         fmt.Sprintf("summary_%d", i),
			Role:       "system",
			Content:    fmt.Sprintf("[Summary of earlier conversation]: %s", summary),
			Timestamp:  s.recentMessages[0].Timestamp, // Use first recent message timestamp
			TokenCount: estimateTokens(summary),
		}
		result = append(result, summaryMsg)
	}

	// Add recent messages
	for _, msg := range s.recentMessages {
		result = append(result, msg)
	}

	return result, nil
}

// Clear removes all messages and summaries
func (s *SummarizationMemory) Clear(ctx context.Context) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.recentMessages = make([]*Message, 0)
	s.summaries = make([]string, 0)
	return nil
}

// GetStats returns statistics about the summarization memory
func (s *SummarizationMemory) GetStats(ctx context.Context) (*Stats, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	// Calculate tokens in recent messages
	recentTokens := 0
	for _, msg := range s.recentMessages {
		recentTokens += msg.TokenCount
	}

	// Calculate tokens in summaries
	summaryTokens := 0
	for _, summary := range s.summaries {
		summaryTokens += estimateTokens(summary)
	}

	totalTokens := recentTokens + summaryTokens

	// Estimate compression rate
	// Assume each summary represents ~summarizeAfter messages
	estimatedOriginalTokens := len(s.summaries)*s.summarizeAfter*100 + recentTokens
	compressionRate := 1.0
	if estimatedOriginalTokens > 0 {
		compressionRate = float64(totalTokens) / float64(estimatedOriginalTokens)
	}

	return &Stats{
		TotalMessages:   len(s.summaries) + len(s.recentMessages),
		TotalTokens:     totalTokens,
		ActiveMessages:  len(s.recentMessages),
		ActiveTokens:    totalTokens,
		CompressionRate: compressionRate,
	}, nil
}

// defaultSummarizer provides a simple summarization function
// In production, this should call an LLM
func defaultSummarizer(ctx context.Context, messages []*Message) (string, error) {
	if len(messages) == 0 {
		return "", nil
	}

	// Simple concatenation with role prefixes
	var parts []string
	for _, msg := range messages {
		// Truncate long messages
		content := msg.Content
		if len(content) > 200 {
			content = content[:200] + "..."
		}
		parts = append(parts, fmt.Sprintf("%s: %s", msg.Role, content))
	}

	summary := fmt.Sprintf("Conversation with %d exchanges covering: %s",
		len(messages), strings.Join(parts, "; "))

	return summary, nil
}
