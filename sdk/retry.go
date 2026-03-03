package sdk

import (
	"context"
	"math"
	"time"
)

// RetryPolicy defines retry behavior
type RetryPolicy struct {
	MaxRetries  int
	MinWait     time.Duration
	MaxWait     time.Duration
}

// ShouldRetry determines if a request should be retried based on status code
func ShouldRetry(statusCode int) bool {
	// Retry on 5xx errors and 429 (rate limit)
	return statusCode >= 500 || statusCode == 429
}

// CalculateBackoff calculates exponential backoff duration
func CalculateBackoff(attempt int, minWait, maxWait time.Duration) time.Duration {
	// Exponential backoff: min * 2^attempt
	backoff := time.Duration(float64(minWait) * math.Pow(2, float64(attempt)))
	
	if backoff > maxWait {
		return maxWait
	}
	
	return backoff
}

// RetryWithBackoff executes a function with exponential backoff
func RetryWithBackoff(ctx context.Context, policy RetryPolicy, fn func() error) error {
	var err error
	
	for attempt := 0; attempt <= policy.MaxRetries; attempt++ {
		err = fn()
		if err == nil {
			return nil
		}
		
		// Don't wait after the last attempt
		if attempt == policy.MaxRetries {
			break
		}
		
		backoff := CalculateBackoff(attempt, policy.MinWait, policy.MaxWait)
		
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-time.After(backoff):
			// Continue to next attempt
		}
	}
	
	return err
}
