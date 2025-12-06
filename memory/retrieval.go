package memory

import (
	"context"
	"fmt"
	"math"
	"sort"
	"sync"
)

// RetrievalMemory uses vector embeddings to retrieve relevant past messages
// Pros: Only fetches contextually relevant history, efficient token usage
// Cons: Requires embedding model, may miss chronologically important context
type RetrievalMemory struct {
	messages   []*Message
	embeddings map[string][]float64 // Message ID -> embedding vector
	topK       int                  // Number of most relevant messages to retrieve
	mu         sync.RWMutex

	// EmbeddingFunc generates embeddings for text
	// In production, this would call an embedding API like OpenAI embeddings
	EmbeddingFunc func(ctx context.Context, text string) ([]float64, error)
}

// RetrievalConfig holds configuration for retrieval-based memory
type RetrievalConfig struct {
	TopK          int                                                        // Number of messages to retrieve
	EmbeddingFunc func(ctx context.Context, text string) ([]float64, error) // Custom embedding function
}

// NewRetrievalMemory creates a new retrieval-based memory strategy
func NewRetrievalMemory(config *RetrievalConfig) *RetrievalMemory {
	if config == nil {
		config = &RetrievalConfig{
			TopK: 5,
		}
	}

	if config.TopK <= 0 {
		config.TopK = 5
	}

	embeddingFunc := config.EmbeddingFunc
	if embeddingFunc == nil {
		embeddingFunc = defaultEmbeddingFunc
	}

	return &RetrievalMemory{
		messages:      make([]*Message, 0),
		embeddings:    make(map[string][]float64),
		topK:          config.TopK,
		EmbeddingFunc: embeddingFunc,
	}
}

// AddMessage adds a message and generates its embedding
func (r *RetrievalMemory) AddMessage(ctx context.Context, msg *Message) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	// Generate embedding for the message
	embedding, err := r.EmbeddingFunc(ctx, msg.Content)
	if err != nil {
		return fmt.Errorf("failed to generate embedding: %w", err)
	}

	r.messages = append(r.messages, msg)
	r.embeddings[msg.ID] = embedding

	return nil
}

// GetContext retrieves the most semantically similar messages to the query
func (r *RetrievalMemory) GetContext(ctx context.Context, query string) ([]*Message, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	if len(r.messages) == 0 {
		return []*Message{}, nil
	}

	// Generate embedding for query
	queryEmbedding, err := r.EmbeddingFunc(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("failed to generate query embedding: %w", err)
	}

	// Calculate similarity scores
	type scoredMessage struct {
		message *Message
		score   float64
	}

	scores := make([]scoredMessage, 0, len(r.messages))
	for _, msg := range r.messages {
		msgEmbedding := r.embeddings[msg.ID]
		similarity := cosineSimilarity(queryEmbedding, msgEmbedding)
		scores = append(scores, scoredMessage{
			message: msg,
			score:   similarity,
		})
	}

	// Sort by similarity (highest first)
	sort.Slice(scores, func(i, j int) bool {
		return scores[i].score > scores[j].score
	})

	// Return top K messages
	k := r.topK
	if k > len(scores) {
		k = len(scores)
	}

	result := make([]*Message, k)
	for i := 0; i < k; i++ {
		result[i] = scores[i].message
	}

	return result, nil
}

// Clear removes all messages and embeddings
func (r *RetrievalMemory) Clear(ctx context.Context) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	r.messages = make([]*Message, 0)
	r.embeddings = make(map[string][]float64)
	return nil
}

// GetStats returns statistics about retrieval memory
func (r *RetrievalMemory) GetStats(ctx context.Context) (*Stats, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	totalTokens := 0
	for _, msg := range r.messages {
		totalTokens += msg.TokenCount
	}

	// Active tokens = tokens in topK messages (approximate)
	activeTokens := 0
	if len(r.messages) > 0 {
		k := r.topK
		if k > len(r.messages) {
			k = len(r.messages)
		}
		for i := 0; i < k; i++ {
			activeTokens += r.messages[i].TokenCount
		}
	}

	return &Stats{
		TotalMessages:   len(r.messages),
		TotalTokens:     totalTokens,
		ActiveMessages:  r.topK,
		ActiveTokens:    activeTokens,
		CompressionRate: float64(activeTokens) / float64(totalTokens),
	}, nil
}

// SetTopK updates the number of messages to retrieve
func (r *RetrievalMemory) SetTopK(k int) {
	r.mu.Lock()
	defer r.mu.Unlock()

	if k > 0 {
		r.topK = k
	}
}

// cosineSimilarity calculates cosine similarity between two vectors
func cosineSimilarity(a, b []float64) float64 {
	if len(a) != len(b) {
		return 0.0
	}

	var dotProduct, normA, normB float64
	for i := range a {
		dotProduct += a[i] * b[i]
		normA += a[i] * a[i]
		normB += b[i] * b[i]
	}

	if normA == 0 || normB == 0 {
		return 0.0
	}

	return dotProduct / (math.Sqrt(normA) * math.Sqrt(normB))
}

// defaultEmbeddingFunc provides a simple embedding function
// In production, use a proper embedding model (e.g., OpenAI embeddings)
func defaultEmbeddingFunc(ctx context.Context, text string) ([]float64, error) {
	// Simple word-frequency based embedding (for demonstration)
	// This is NOT a proper embedding - use an actual model in production

	words := make(map[string]int)
	for _, char := range text {
		word := string(char)
		words[word]++
	}

	// Create a fixed-size vector
	embedding := make([]float64, 128)
	for word, count := range words {
		// Hash word to index
		hash := 0
		for _, c := range word {
			hash = (hash*31 + int(c)) % 128
		}
		embedding[hash] += float64(count)
	}

	// Normalize
	var norm float64
	for _, val := range embedding {
		norm += val * val
	}
	norm = math.Sqrt(norm)

	if norm > 0 {
		for i := range embedding {
			embedding[i] /= norm
		}
	}

	return embedding, nil
}
