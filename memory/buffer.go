package memory

import (
	"context"
	"sync"
)

// BufferMemory is a simple buffer-based memory implementation
// Similar to LangChain's ConversationBufferMemory
// Combines sliding window with optional summarization
type BufferMemory struct {
	messages     []*Message
	maxMessages  int  // 0 = unlimited
	maxTokens    int  // 0 = unlimited
	autoSummarize bool // Auto-summarize when limits exceeded
	mu           sync.RWMutex

	// Optional summarizer
	Summarizer func(ctx context.Context, messages []*Message) (string, error)
}

// BufferConfig holds configuration for buffer memory
type BufferConfig struct {
	MaxMessages   int  // Maximum number of messages (0 = unlimited)
	MaxTokens     int  // Maximum total tokens (0 = unlimited)
	AutoSummarize bool // Enable automatic summarization
	Summarizer    func(ctx context.Context, messages []*Message) (string, error)
}

// NewBufferMemory creates a new buffer memory
func NewBufferMemory(config *BufferConfig) *BufferMemory {
	if config == nil {
		config = &BufferConfig{
			MaxMessages:   0, // Unlimited by default
			MaxTokens:     0,
			AutoSummarize: false,
		}
	}

	summarizer := config.Summarizer
	if summarizer == nil {
		summarizer = defaultSummarizer
	}

	return &BufferMemory{
		messages:      make([]*Message, 0),
		maxMessages:   config.MaxMessages,
		maxTokens:     config.MaxTokens,
		autoSummarize: config.AutoSummarize,
		Summarizer:    summarizer,
	}
}

// AddMessage adds a message to the buffer
func (b *BufferMemory) AddMessage(ctx context.Context, msg *Message) error {
	b.mu.Lock()
	defer b.mu.Unlock()

	b.messages = append(b.messages, msg)

	// Check limits and trim if necessary
	if b.maxMessages > 0 && len(b.messages) > b.maxMessages {
		if b.autoSummarize {
			// Summarize oldest messages
			toSummarize := b.messages[:len(b.messages)-b.maxMessages]
			summary, err := b.Summarizer(ctx, toSummarize)
			if err == nil {
				summaryMsg := NewMessage("system", summary)
				b.messages = append([]*Message{summaryMsg}, b.messages[len(b.messages)-b.maxMessages:]...)
			} else {
				// Fallback: just trim
				b.messages = b.messages[len(b.messages)-b.maxMessages:]
			}
		} else {
			// Simple trim
			b.messages = b.messages[len(b.messages)-b.maxMessages:]
		}
	}

	// Check token limit
	if b.maxTokens > 0 {
		totalTokens := 0
		for i := len(b.messages) - 1; i >= 0; i-- {
			totalTokens += b.messages[i].TokenCount
			if totalTokens > b.maxTokens {
				// Remove oldest messages
				if b.autoSummarize && i > 0 {
					toSummarize := b.messages[:i]
					summary, err := b.Summarizer(ctx, toSummarize)
					if err == nil {
						summaryMsg := NewMessage("system", summary)
						b.messages = append([]*Message{summaryMsg}, b.messages[i:]...)
					} else {
						b.messages = b.messages[i:]
					}
				} else {
					b.messages = b.messages[i:]
				}
				break
			}
		}
	}

	return nil
}

// GetContext returns all messages in the buffer
func (b *BufferMemory) GetContext(ctx context.Context, query string) ([]*Message, error) {
	b.mu.RLock()
	defer b.mu.RUnlock()

	result := make([]*Message, len(b.messages))
	copy(result, b.messages)
	return result, nil
}

// Clear removes all messages
func (b *BufferMemory) Clear(ctx context.Context) error {
	b.mu.Lock()
	defer b.mu.Unlock()

	b.messages = make([]*Message, 0)
	return nil
}

// GetStats returns buffer memory statistics
func (b *BufferMemory) GetStats(ctx context.Context) (*Stats, error) {
	b.mu.RLock()
	defer b.mu.RUnlock()

	totalTokens := 0
	for _, msg := range b.messages {
		totalTokens += msg.TokenCount
	}

	return &Stats{
		TotalMessages:   len(b.messages),
		TotalTokens:     totalTokens,
		ActiveMessages:  len(b.messages),
		ActiveTokens:    totalTokens,
		CompressionRate: 1.0,
	}, nil
}

// GetMessages returns a copy of all messages
func (b *BufferMemory) GetMessages() []*Message {
	b.mu.RLock()
	defer b.mu.RUnlock()

	result := make([]*Message, len(b.messages))
	copy(result, b.messages)
	return result
}

// LoadMessages loads messages into the buffer (replaces existing)
func (b *BufferMemory) LoadMessages(messages []*Message) {
	b.mu.Lock()
	defer b.mu.Unlock()

	b.messages = make([]*Message, len(messages))
	copy(b.messages, messages)
}
