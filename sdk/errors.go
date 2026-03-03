package sdk

import (
	"errors"
	"fmt"
)

var (
	// ErrInvalidASN is returned when an invalid ASN is provided
	ErrInvalidASN = errors.New("invalid ASN")
	
	// ErrInvalidPrefix is returned when an invalid prefix is provided
	ErrInvalidPrefix = errors.New("invalid prefix")
	
	// ErrRateLimitExceeded is returned when rate limit is exceeded
	ErrRateLimitExceeded = errors.New("rate limit exceeded")
	
	// ErrTimeout is returned when a request times out
	ErrTimeout = errors.New("request timeout")
	
	// ErrAPIError is returned when the API returns an error
	ErrAPIError = errors.New("API error")
)

// APIError represents an error from the RIPE RIS API
type APIError struct {
	StatusCode int
	Message    string
	Endpoint   string
}

func (e *APIError) Error() string {
	return fmt.Sprintf("API error [%d] on %s: %s", e.StatusCode, e.Endpoint, e.Message)
}

// WrapAPIError creates a new APIError
func WrapAPIError(statusCode int, endpoint, message string) error {
	return &APIError{
		StatusCode: statusCode,
		Message:    message,
		Endpoint:   endpoint,
	}
}
