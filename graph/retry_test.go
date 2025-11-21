package graph_test

import (
	"context"
	"errors"
	"sync/atomic"
	"testing"
	"time"

	"github.com/smallnest/langgraphgo/graph"
)

const (
	successResult = "success"
)

//nolint:gocognit,cyclop // Comprehensive retry logic test with multiple scenarios
func TestRetryNode(t *testing.T) {
	t.Parallel()

	t.Run("SuccessOnFirstAttempt", func(t *testing.T) {
		g := graph.NewMessageGraph()
		callCount := int32(0)

		g.AddNodeWithRetry("retry_node",
			func(ctx context.Context, state interface{}) (interface{}, error) {
				atomic.AddInt32(&callCount, 1)
				return successResult, nil
			},
			&graph.RetryConfig{
				MaxAttempts:   3,
				InitialDelay:  10 * time.Millisecond,
				BackoffFactor: 2.0,
			},
		)

		g.AddEdge("retry_node", graph.END)
		g.SetEntryPoint("retry_node")

		runnable, err := g.Compile()
		if err != nil {
			t.Fatalf("Failed to compile: %v", err)
		}

		result, err := runnable.Invoke(context.Background(), "input")
		if err != nil {
			t.Fatalf("Execution failed: %v", err)
		}

		if result != successResult {
			t.Errorf("Expected success, got %v", result)
		}

		if atomic.LoadInt32(&callCount) != 1 {
			t.Errorf("Expected 1 call, got %d", callCount)
		}
	})

	t.Run("RetryOnTransientFailure", func(t *testing.T) {
		g := graph.NewMessageGraph()
		callCount := int32(0)

		g.AddNodeWithRetry("retry_node",
			func(ctx context.Context, state interface{}) (interface{}, error) {
				count := atomic.AddInt32(&callCount, 1)
				if count < 3 {
					return nil, errors.New("transient error")
				}
				return successResult, nil
			},
			&graph.RetryConfig{
				MaxAttempts:   5,
				InitialDelay:  5 * time.Millisecond,
				BackoffFactor: 1.5,
			},
		)

		g.AddEdge("retry_node", graph.END)
		g.SetEntryPoint("retry_node")

		runnable, err := g.Compile()
		if err != nil {
			t.Fatalf("Failed to compile: %v", err)
		}

		result, err := runnable.Invoke(context.Background(), "input")
		if err != nil {
			t.Fatalf("Execution failed: %v", err)
		}

		if result != successResult {
			t.Errorf("Expected success, got %v", result)
		}

		if atomic.LoadInt32(&callCount) != 3 {
			t.Errorf("Expected 3 calls, got %d", callCount)
		}
	})

	t.Run("MaxAttemptsExceeded", func(t *testing.T) {
		g := graph.NewMessageGraph()
		callCount := int32(0)

		g.AddNodeWithRetry("retry_node",
			func(ctx context.Context, state interface{}) (interface{}, error) {
				atomic.AddInt32(&callCount, 1)
				return nil, errors.New("persistent error")
			},
			&graph.RetryConfig{
				MaxAttempts:   3,
				InitialDelay:  5 * time.Millisecond,
				BackoffFactor: 2.0,
			},
		)

		g.AddEdge("retry_node", graph.END)
		g.SetEntryPoint("retry_node")

		runnable, err := g.Compile()
		if err != nil {
			t.Fatalf("Failed to compile: %v", err)
		}

		_, err = runnable.Invoke(context.Background(), "input")
		if err == nil {
			t.Error("Expected error for max retries exceeded")
		}

		if atomic.LoadInt32(&callCount) != 3 {
			t.Errorf("Expected 3 attempts, got %d", callCount)
		}
	})

	t.Run("NonRetryableError", func(t *testing.T) {
		g := graph.NewMessageGraph()
		callCount := int32(0)

		g.AddNodeWithRetry("retry_node",
			func(ctx context.Context, state interface{}) (interface{}, error) {
				atomic.AddInt32(&callCount, 1)
				return nil, errors.New("critical error")
			},
			&graph.RetryConfig{
				MaxAttempts:   3,
				InitialDelay:  5 * time.Millisecond,
				BackoffFactor: 2.0,
				RetryableErrors: func(err error) bool {
					// Only retry if error message contains "transient"
					return err != nil && err.Error() == "transient"
				},
			},
		)

		g.AddEdge("retry_node", graph.END)
		g.SetEntryPoint("retry_node")

		runnable, err := g.Compile()
		if err != nil {
			t.Fatalf("Failed to compile: %v", err)
		}

		_, err = runnable.Invoke(context.Background(), "input")
		if err == nil {
			t.Error("Expected error for non-retryable error")
		}

		// Should only try once for non-retryable errors
		if atomic.LoadInt32(&callCount) != 1 {
			t.Errorf("Expected 1 attempt for non-retryable error, got %d", callCount)
		}
	})
}

func TestTimeoutNode(t *testing.T) {
	t.Parallel()

	t.Run("SuccessWithinTimeout", func(t *testing.T) {
		g := graph.NewMessageGraph()

		g.AddNodeWithTimeout("timeout_node",
			func(ctx context.Context, state interface{}) (interface{}, error) {
				time.Sleep(10 * time.Millisecond)
				return successResult, nil
			},
			100*time.Millisecond,
		)

		g.AddEdge("timeout_node", graph.END)
		g.SetEntryPoint("timeout_node")

		runnable, err := g.Compile()
		if err != nil {
			t.Fatalf("Failed to compile: %v", err)
		}

		result, err := runnable.Invoke(context.Background(), "input")
		if err != nil {
			t.Fatalf("Execution failed: %v", err)
		}

		if result != successResult {
			t.Errorf("Expected success, got %v", result)
		}
	})

	t.Run("TimeoutExceeded", func(t *testing.T) {
		g := graph.NewMessageGraph()

		g.AddNodeWithTimeout("timeout_node",
			func(ctx context.Context, state interface{}) (interface{}, error) {
				time.Sleep(100 * time.Millisecond)
				return successResult, nil
			},
			20*time.Millisecond,
		)

		g.AddEdge("timeout_node", graph.END)
		g.SetEntryPoint("timeout_node")

		runnable, err := g.Compile()
		if err != nil {
			t.Fatalf("Failed to compile: %v", err)
		}

		_, err = runnable.Invoke(context.Background(), "input")
		if err == nil {
			t.Error("Expected timeout error")
		}
	})

	t.Run("RespectContextCancellation", func(t *testing.T) {
		g := graph.NewMessageGraph()

		g.AddNodeWithTimeout("timeout_node",
			func(ctx context.Context, _ interface{}) (interface{}, error) {
				select {
				case <-ctx.Done():
					return nil, ctx.Err()
				case <-time.After(100 * time.Millisecond):
					return successResult, nil
				}
			},
			200*time.Millisecond,
		)

		g.AddEdge("timeout_node", graph.END)
		g.SetEntryPoint("timeout_node")

		runnable, err := g.Compile()
		if err != nil {
			t.Fatalf("Failed to compile: %v", err)
		}

		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Millisecond)
		defer cancel()

		_, err = runnable.Invoke(ctx, "input")
		if err == nil {
			t.Error("Expected context cancellation error")
		}
	})
}

func TestCircuitBreaker(t *testing.T) {
	t.Parallel()

	t.Run("CircuitOpensAfterFailures", func(t *testing.T) {
		g := graph.NewMessageGraph()
		callCount := int32(0)

		g.AddNodeWithCircuitBreaker("cb_node",
			func(ctx context.Context, state interface{}) (interface{}, error) {
				atomic.AddInt32(&callCount, 1)
				return nil, errors.New("service unavailable")
			},
			graph.CircuitBreakerConfig{
				FailureThreshold: 2,
				SuccessThreshold: 2,
				Timeout:          50 * time.Millisecond,
				HalfOpenMaxCalls: 1,
			},
		)

		g.AddEdge("cb_node", graph.END)
		g.SetEntryPoint("cb_node")

		runnable, err := g.Compile()
		if err != nil {
			t.Fatalf("Failed to compile: %v", err)
		}

		// First call - should fail
		_, _ = runnable.Invoke(context.Background(), "input")

		// Second call - should fail and open circuit
		_, _ = runnable.Invoke(context.Background(), "input")

		// Third call - circuit should be open
		_, err = runnable.Invoke(context.Background(), "input")
		if err == nil {
			t.Error("Expected circuit breaker open error")
		}

		// Should have only 2 actual calls (third blocked by circuit breaker)
		if atomic.LoadInt32(&callCount) != 2 {
			t.Errorf("Expected 2 calls, got %d", callCount)
		}
	})

	t.Run("CircuitClosesAfterSuccess", func(t *testing.T) {
		g := graph.NewMessageGraph()
		callCount := int32(0)

		g.AddNodeWithCircuitBreaker("cb_node",
			func(ctx context.Context, state interface{}) (interface{}, error) {
				count := atomic.AddInt32(&callCount, 1)
				// Fail first 2 calls, succeed afterwards
				if count <= 2 {
					return nil, errors.New("service unavailable")
				}
				return successResult, nil
			},
			graph.CircuitBreakerConfig{
				FailureThreshold: 2,
				SuccessThreshold: 1,
				Timeout:          10 * time.Millisecond,
				HalfOpenMaxCalls: 2,
			},
		)

		g.AddEdge("cb_node", graph.END)
		g.SetEntryPoint("cb_node")

		runnable, err := g.Compile()
		if err != nil {
			t.Fatalf("Failed to compile: %v", err)
		}

		// First two calls fail and open circuit
		_, _ = runnable.Invoke(context.Background(), "input")
		_, _ = runnable.Invoke(context.Background(), "input")

		// Wait for timeout to move to half-open
		time.Sleep(15 * time.Millisecond)

		// This call should succeed and close circuit
		result, err := runnable.Invoke(context.Background(), "input")
		if err != nil {
			t.Fatalf("Expected success after circuit recovery: %v", err)
		}

		if result != successResult {
			t.Errorf("Expected success, got %v", result)
		}
	})
}

//nolint:gocognit,cyclop // Comprehensive rate limiter test with multiple scenarios
func TestRateLimiter(t *testing.T) {
	t.Parallel()

	t.Run("AllowsCallsWithinLimit", func(t *testing.T) {
		g := graph.NewMessageGraph()
		callCount := int32(0)

		g.AddNodeWithRateLimit("rate_limited",
			func(ctx context.Context, state interface{}) (interface{}, error) {
				atomic.AddInt32(&callCount, 1)
				return successResult, nil
			},
			3,
			100*time.Millisecond,
		)

		g.AddEdge("rate_limited", graph.END)
		g.SetEntryPoint("rate_limited")

		runnable, err := g.Compile()
		if err != nil {
			t.Fatalf("Failed to compile: %v", err)
		}

		// Make 3 calls within the window - all should succeed
		for i := 0; i < 3; i++ {
			result, err := runnable.Invoke(context.Background(), "input")
			if err != nil {
				t.Fatalf("Call %d failed: %v", i+1, err)
			}
			if result != successResult {
				t.Errorf("Expected success, got %v", result)
			}
		}

		if atomic.LoadInt32(&callCount) != 3 {
			t.Errorf("Expected 3 calls, got %d", callCount)
		}
	})

	t.Run("BlocksCallsExceedingLimit", func(t *testing.T) {
		g := graph.NewMessageGraph()
		callCount := int32(0)

		g.AddNodeWithRateLimit("rate_limited",
			func(ctx context.Context, state interface{}) (interface{}, error) {
				atomic.AddInt32(&callCount, 1)
				return successResult, nil
			},
			2,
			100*time.Millisecond,
		)

		g.AddEdge("rate_limited", graph.END)
		g.SetEntryPoint("rate_limited")

		runnable, err := g.Compile()
		if err != nil {
			t.Fatalf("Failed to compile: %v", err)
		}

		// Make 2 calls - should succeed
		_, _ = runnable.Invoke(context.Background(), "input")
		_, _ = runnable.Invoke(context.Background(), "input")

		// Third call should be rate limited
		_, err = runnable.Invoke(context.Background(), "input")
		if err == nil {
			t.Error("Expected rate limit error")
		}

		if atomic.LoadInt32(&callCount) != 2 {
			t.Errorf("Expected 2 calls, got %d", callCount)
		}
	})

	t.Run("AllowsCallsAfterWindowExpires", func(t *testing.T) {
		g := graph.NewMessageGraph()
		callCount := int32(0)

		g.AddNodeWithRateLimit("rate_limited",
			func(ctx context.Context, state interface{}) (interface{}, error) {
				atomic.AddInt32(&callCount, 1)
				return successResult, nil
			},
			2,
			50*time.Millisecond,
		)

		g.AddEdge("rate_limited", graph.END)
		g.SetEntryPoint("rate_limited")

		runnable, err := g.Compile()
		if err != nil {
			t.Fatalf("Failed to compile: %v", err)
		}

		// Make 2 calls
		_, _ = runnable.Invoke(context.Background(), "input")
		_, _ = runnable.Invoke(context.Background(), "input")

		// Wait for window to expire
		time.Sleep(60 * time.Millisecond)

		// Should be able to make more calls
		result, err := runnable.Invoke(context.Background(), "input")
		if err != nil {
			t.Fatalf("Call after window expiry failed: %v", err)
		}

		if result != successResult {
			t.Errorf("Expected success, got %v", result)
		}

		if atomic.LoadInt32(&callCount) != 3 {
			t.Errorf("Expected 3 calls, got %d", callCount)
		}
	})
}

func TestExponentialBackoffRetry(t *testing.T) {
	t.Parallel()

	t.Run("SuccessAfterRetries", func(t *testing.T) {
		attempts := int32(0)

		fn := func() (interface{}, error) {
			count := atomic.AddInt32(&attempts, 1)
			if count < 3 {
				return nil, errors.New("temporary failure")
			}
			return successResult, nil
		}

		result, err := graph.ExponentialBackoffRetry(
			context.Background(),
			fn,
			5,
			5*time.Millisecond,
		)

		if err != nil {
			t.Fatalf("Expected success, got error: %v", err)
		}

		if result != successResult {
			t.Errorf("Expected success, got %v", result)
		}

		if atomic.LoadInt32(&attempts) != 3 {
			t.Errorf("Expected 3 attempts, got %d", attempts)
		}
	})

	t.Run("MaxAttemptsReached", func(t *testing.T) {
		attempts := int32(0)

		fn := func() (interface{}, error) {
			atomic.AddInt32(&attempts, 1)
			return nil, errors.New("persistent failure")
		}

		_, err := graph.ExponentialBackoffRetry(
			context.Background(),
			fn,
			3,
			5*time.Millisecond,
		)

		if err == nil {
			t.Error("Expected max attempts error")
		}

		if atomic.LoadInt32(&attempts) != 3 {
			t.Errorf("Expected 3 attempts, got %d", attempts)
		}
	})

	t.Run("ContextCancellation", func(t *testing.T) {
		attempts := int32(0)

		fn := func() (interface{}, error) {
			atomic.AddInt32(&attempts, 1)
			return nil, errors.New("failure")
		}

		ctx, cancel := context.WithTimeout(context.Background(), 20*time.Millisecond)
		defer cancel()

		_, err := graph.ExponentialBackoffRetry(
			ctx,
			fn,
			10,
			10*time.Millisecond,
		)

		if !errors.Is(err, context.DeadlineExceeded) {
			t.Errorf("Expected context deadline exceeded, got %v", err)
		}

		// Should have made at least 1 attempt but not all 10
		count := atomic.LoadInt32(&attempts)
		if count < 1 || count >= 10 {
			t.Errorf("Expected 1-3 attempts, got %d", count)
		}
	})
}
