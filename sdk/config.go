package sdk

import "time"

// Config holds the SDK configuration
type Config struct {
	// Timeout for HTTP requests
	Timeout time.Duration
	
	// RateLimit defines max requests per second
	RateLimit int
	
	// RetryMax is the maximum number of retry attempts
	RetryMax int
	
	// RetryWaitMin is the minimum wait time between retries
	RetryWaitMin time.Duration
	
	// RetryWaitMax is the maximum wait time between retries
	RetryWaitMax time.Duration
	
	// UserAgent for HTTP requests
	UserAgent string
	
	// BaseURL for RIPE RIS API
	BaseURL string
}

// DefaultConfig returns a Config with sensible defaults
func DefaultConfig() Config {
	return Config{
		Timeout:      30 * time.Second,
		RateLimit:    10, // 10 requests per second
		RetryMax:     3,
		RetryWaitMin: 1 * time.Second,
		RetryWaitMax: 10 * time.Second,
		UserAgent:    "bgpin-sdk/1.0",
		BaseURL:      "https://stat.ripe.net/data",
	}
}
