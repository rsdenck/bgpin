package http

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"time"
)

// HTTPAdapter handles HTTP-based Looking Glass queries
type HTTPAdapter struct {
	baseURL string
	client  *http.Client
}

// NewHTTPAdapter creates a new HTTP adapter
func NewHTTPAdapter(baseURL string, timeout time.Duration) *HTTPAdapter {
	return &HTTPAdapter{
		baseURL: baseURL,
		client: &http.Client{
			Timeout: timeout,
		},
	}
}

// QueryBGP queries a BGP prefix via HTTP
func (a *HTTPAdapter) QueryBGP(ctx context.Context, prefix string) (string, error) {
	url := fmt.Sprintf("%s/bgp?prefix=%s", a.baseURL, prefix)
	
	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return "", fmt.Errorf("failed to create request: %w", err)
	}
	
	resp, err := a.client.Do(req)
	if err != nil {
		return "", fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()
	
	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}
	
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("failed to read response: %w", err)
	}
	
	return string(body), nil
}
