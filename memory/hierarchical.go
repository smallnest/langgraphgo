package memory

import (
	"context"
	"sync"
	"time"
)

// HierarchicalMemory organizes messages in layers based on importance and recency
// Pros: Balances recent context with important historical information
// Cons: More complex management, requires importance scoring
type HierarchicalMemory struct {
	// Layer 1: Recent messages (always included)
	recentMessages []*Message
	recentLimit    int

	// Layer 2: Important messages (high priority)
	importantMessages []*Message
	importantLimit    int

	// Layer 3: Archived messages (low priority, rarely accessed)
	archivedMessages []*Message

	mu sync.RWMutex

	// ImportanceScorer determines message importance (0.0 to 1.0)
	// Higher scores = more important
	ImportanceScorer func(msg *Message) float64
}

// HierarchicalConfig holds configuration for hierarchical memory
type HierarchicalConfig struct {
	RecentLimit      int                        // Max recent messages
	ImportantLimit   int                        // Max important messages
	ImportanceScorer func(msg *Message) float64 // Custom importance scorer
}

// NewHierarchicalMemory creates a new hierarchical memory strategy
func NewHierarchicalMemory(config *HierarchicalConfig) *HierarchicalMemory {
	if config == nil {
		config = &HierarchicalConfig{
			RecentLimit:    10,
			ImportantLimit: 20,
		}
	}

	if config.RecentLimit <= 0 {
		config.RecentLimit = 10
	}
	if config.ImportantLimit <= 0 {
		config.ImportantLimit = 20
	}

	scorer := config.ImportanceScorer
	if scorer == nil {
		scorer = defaultImportanceScorer
	}

	return &HierarchicalMemory{
		recentMessages:    make([]*Message, 0),
		importantMessages: make([]*Message, 0),
		archivedMessages:  make([]*Message, 0),
		recentLimit:       config.RecentLimit,
		importantLimit:    config.ImportantLimit,
		ImportanceScorer:  scorer,
	}
}

// AddMessage adds a message and organizes it into appropriate layer
func (h *HierarchicalMemory) AddMessage(ctx context.Context, msg *Message) error {
	h.mu.Lock()
	defer h.mu.Unlock()

	// Always add to recent messages
	h.recentMessages = append(h.recentMessages, msg)

	// Score the message for importance
	// Check if there's an explicit importance in metadata first
	if importance, ok := msg.Metadata["importance"].(float64); ok && importance > 0.7 {
		h.importantMessages = append(h.importantMessages, msg)
	} else {
		// Use scorer to determine importance
		score := h.ImportanceScorer(msg)
		if score > 0.7 {
			h.importantMessages = append(h.importantMessages, msg)
		}
	}

	// Manage recent layer size
	if len(h.recentMessages) > h.recentLimit {
		// Move oldest recent message to archive
		oldest := h.recentMessages[0]
		h.recentMessages = h.recentMessages[1:]

		// Only archive if not already in important layer
		if !h.isInImportant(oldest) {
			h.archivedMessages = append(h.archivedMessages, oldest)
		}
	}

	// Manage important layer size
	if len(h.importantMessages) > h.importantLimit {
		// Move lowest importance message to archive
		lowestIdx := h.findLowestImportance()
		if lowestIdx >= 0 {
			archived := h.importantMessages[lowestIdx]
			h.importantMessages = append(h.importantMessages[:lowestIdx], h.importantMessages[lowestIdx+1:]...)
			h.archivedMessages = append(h.archivedMessages, archived)
		}
	}

	return nil
}

// GetContext returns messages from all layers, prioritized by importance
func (h *HierarchicalMemory) GetContext(ctx context.Context, query string) ([]*Message, error) {
	h.mu.RLock()
	defer h.mu.RUnlock()

	result := make([]*Message, 0)

	// Layer 1: Important messages (highest priority)
	result = append(result, h.importantMessages...)

	// Layer 2: Recent messages (medium priority)
	// Avoid duplicates with important layer
	for _, msg := range h.recentMessages {
		if !h.containsMessage(result, msg) {
			result = append(result, msg)
		}
	}

	// Note: Archived messages are only included if specifically queried
	// This implementation doesn't include them by default

	return result, nil
}

// Clear removes all messages from all layers
func (h *HierarchicalMemory) Clear(ctx context.Context) error {
	h.mu.Lock()
	defer h.mu.Unlock()

	h.recentMessages = make([]*Message, 0)
	h.importantMessages = make([]*Message, 0)
	h.archivedMessages = make([]*Message, 0)
	return nil
}

// GetStats returns statistics about hierarchical memory
func (h *HierarchicalMemory) GetStats(ctx context.Context) (*Stats, error) {
	h.mu.RLock()
	defer h.mu.RUnlock()

	allMessages := len(h.recentMessages) + len(h.importantMessages) + len(h.archivedMessages)

	totalTokens := 0
	activeTokens := 0

	// Count tokens in all layers
	for _, msg := range h.recentMessages {
		totalTokens += msg.TokenCount
		activeTokens += msg.TokenCount
	}
	for _, msg := range h.importantMessages {
		totalTokens += msg.TokenCount
		activeTokens += msg.TokenCount
	}
	for _, msg := range h.archivedMessages {
		totalTokens += msg.TokenCount
	}

	activeMessages := len(h.recentMessages) + len(h.importantMessages)

	compressionRate := 1.0
	if totalTokens > 0 {
		compressionRate = float64(activeTokens) / float64(totalTokens)
	}

	return &Stats{
		TotalMessages:   allMessages,
		TotalTokens:     totalTokens,
		ActiveMessages:  activeMessages,
		ActiveTokens:    activeTokens,
		CompressionRate: compressionRate,
	}, nil
}

// isInImportant checks if a message is in the important layer
func (h *HierarchicalMemory) isInImportant(msg *Message) bool {
	for _, m := range h.importantMessages {
		if m.ID == msg.ID {
			return true
		}
	}
	return false
}

// findLowestImportance finds the index of the least important message
func (h *HierarchicalMemory) findLowestImportance() int {
	if len(h.importantMessages) == 0 {
		return -1
	}

	lowestIdx := 0
	lowestScore := h.ImportanceScorer(h.importantMessages[0])

	for i, msg := range h.importantMessages {
		score := h.ImportanceScorer(msg)
		if score < lowestScore {
			lowestScore = score
			lowestIdx = i
		}
	}

	return lowestIdx
}

// containsMessage checks if a message is already in the result set
func (h *HierarchicalMemory) containsMessage(messages []*Message, target *Message) bool {
	for _, msg := range messages {
		if msg.ID == target.ID {
			return true
		}
	}
	return false
}

// defaultImportanceScorer provides a simple importance scoring function
// Scores based on: message length, role, and recency
func defaultImportanceScorer(msg *Message) float64 {
	score := 0.5 // Base score

	// Boost for system messages
	if msg.Role == "system" {
		score += 0.2
	}

	// Boost for longer messages (more content = potentially more important)
	if msg.TokenCount > 100 {
		score += 0.2
	}

	// Boost for very recent messages
	age := time.Since(msg.Timestamp)
	if age < time.Minute*5 {
		score += 0.1
	}

	// Check metadata for explicit importance
	if importance, ok := msg.Metadata["importance"].(float64); ok {
		score = importance
	}

	// Clamp to [0, 1]
	if score > 1.0 {
		score = 1.0
	}
	if score < 0.0 {
		score = 0.0
	}

	return score
}
