package rpki

import (
	"context"
	"fmt"
	"net"
	"sync"
	"time"
)

type ValidationResult struct {
	Prefix      string     `json:"prefix"`
	ASN         int        `json:"asn"`
	Status      RPKIStatus `json:"status"`
	ValidLength bool       `json:"valid_length"`
	ValidASN    bool       `json:"valid_asn"`
	Expires     time.Time  `json:"expires"`
	Timestamp   time.Time  `json:"timestamp"`
}

type RPKIStatus string

const (
	StatusValid    RPKIStatus = "valid"
	StatusInvalid  RPKIStatus = "invalid"
	StatusNotFound RPKIStatus = "not_found"
	StatusError    RPKIStatus = "error"
)

type Validator interface {
	Validate(ctx context.Context, prefix string, asn int) (*ValidationResult, error)
	ValidateBatch(ctx context.Context, routes []RouteInput) ([]ValidationResult, error)
	Close() error
}

type RouteInput struct {
	Prefix string
	ASN    int
}

type RPKIValidator struct {
	mu         sync.RWMutex
	cache      map[string]*ValidationResult
	cacheTTL   time.Duration
	serverAddr string
	timeout    time.Duration
}

func NewRPKIValidator(serverAddr string, cacheTTL time.Duration) *RPKIValidator {
	return &RPKIValidator{
		serverAddr: serverAddr,
		cacheTTL:   cacheTTL,
		cache:      make(map[string]*ValidationResult),
		timeout:    10 * time.Second,
	}
}

func (v *RPKIValidator) Validate(ctx context.Context, prefix string, asn int) (*ValidationResult, error) {
	cacheKey := fmt.Sprintf("%s_%d", prefix, asn)

	v.mu.RLock()
	if cached, ok := v.cache[cacheKey]; ok {
		if time.Since(cached.Timestamp) < v.cacheTTL {
			v.mu.RUnlock()
			return cached, nil
		}
	}
	v.mu.RUnlock()

	_, _, err := net.ParseCIDR(prefix)
	if err != nil {
		return &ValidationResult{
			Prefix:    prefix,
			ASN:       asn,
			Status:    StatusError,
			Timestamp: time.Now(),
		}, fmt.Errorf("invalid prefix: %w", err)
	}

	result := &ValidationResult{
		Prefix:      prefix,
		ASN:         asn,
		Status:      StatusNotFound,
		ValidLength: false,
		ValidASN:    false,
		Timestamp:   time.Now(),
		Expires:     time.Now().Add(v.cacheTTL),
	}

	v.mu.Lock()
	v.cache[cacheKey] = result
	v.mu.Unlock()

	return result, nil
}

func (v *RPKIValidator) ValidateBatch(ctx context.Context, routes []RouteInput) ([]ValidationResult, error) {
	results := make([]ValidationResult, 0, len(routes))

	var wg sync.WaitGroup
	var mu sync.Mutex
	errs := make([]error, 0)

	for _, route := range routes {
		wg.Add(1)
		go func(r RouteInput) {
			defer wg.Done()
			result, err := v.Validate(ctx, r.Prefix, r.ASN)
			mu.Lock()
			if err != nil {
				errs = append(errs, err)
			} else {
				results = append(results, *result)
			}
			mu.Unlock()
		}(route)
	}

	wg.Wait()

	if len(errs) > 0 {
		return results, fmt.Errorf("validation errors: %v", errs)
	}

	return results, nil
}

func (v *RPKIValidator) Close() error {
	return nil
}

type MockValidator struct{}

func NewMockValidator() *MockValidator {
	return &MockValidator{}
}

func (m *MockValidator) Validate(ctx context.Context, prefix string, asn int) (*ValidationResult, error) {
	return &ValidationResult{
		Prefix:      prefix,
		ASN:         asn,
		Status:      StatusNotFound,
		ValidLength: false,
		ValidASN:    false,
		Timestamp:   time.Now(),
	}, nil
}

func (m *MockValidator) ValidateBatch(ctx context.Context, routes []RouteInput) ([]ValidationResult, error) {
	results := make([]ValidationResult, len(routes))
	for i, r := range routes {
		results[i] = ValidationResult{
			Prefix:      r.Prefix,
			ASN:         r.ASN,
			Status:      StatusNotFound,
			ValidLength: false,
			ValidASN:    false,
			Timestamp:   time.Now(),
		}
	}
	return results, nil
}

func (m *MockValidator) Close() error {
	return nil
}
