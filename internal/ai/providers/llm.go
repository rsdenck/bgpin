package providers

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
	"sync"
	"time"
)

const (
	DefaultTimeout   = 60 * time.Second
	MaxTokens        = 2048
	MaxRetries       = 3
	BaseRetryDelay   = time.Second
	RateLimitWindow  = time.Minute
	RateLimitRequests = 10
)

type LLMProvider interface {
	Name() string
	Analyze(ctx context.Context, prompt string, data interface{}) (string, error)
}

type BaseProvider struct {
	APIKey     string
	BaseURL    string
	Model      string
	Timeout    time.Duration
	rateLimiter *RateLimiter
}

type RateLimiter struct {
	requests []time.Time
	mutex    sync.Mutex
	limit    int
	window   time.Duration
}

func NewRateLimiter(limit int, window time.Duration) *RateLimiter {
	return &RateLimiter{
		requests: make([]time.Time, 0),
		limit:    limit,
		window:   window,
	}
}

func (rl *RateLimiter) Allow() bool {
	rl.mutex.Lock()
	defer rl.mutex.Unlock()

	now := time.Now()
	cutoff := now.Add(-rl.window)

	// Remove old requests
	filtered := make([]time.Time, 0)
	for _, req := range rl.requests {
		if req.After(cutoff) {
			filtered = append(filtered, req)
		}
	}
	rl.requests = filtered

	if len(rl.requests) >= rl.limit {
		return false
	}

	rl.requests = append(rl.requests, now)
	return true
}

func NewBaseProvider(timeout time.Duration) *BaseProvider {
	if timeout == 0 {
		timeout = DefaultTimeout
	}
	return &BaseProvider{
		Timeout:     timeout,
		Model:       "gpt-4",
		rateLimiter: NewRateLimiter(RateLimitRequests, RateLimitWindow),
	}
}

func (bp *BaseProvider) retryWithBackoff(ctx context.Context, fn func() (*http.Response, error)) (*http.Response, error) {
	var lastErr error
	
	for attempt := 0; attempt < MaxRetries; attempt++ {
		if attempt > 0 {
			delay := BaseRetryDelay * time.Duration(1<<uint(attempt-1))
			select {
			case <-ctx.Done():
				return nil, ctx.Err()
			case <-time.After(delay):
			}
		}

		resp, err := fn()
		if err == nil && resp.StatusCode < 500 {
			return resp, nil
		}
		
		if resp != nil {
			resp.Body.Close()
		}
		lastErr = err
	}
	
	return nil, fmt.Errorf("max retries exceeded: %w", lastErr)
}

func (bp *BaseProvider) checkRateLimit() error {
	if !bp.rateLimiter.Allow() {
		return fmt.Errorf("rate limit exceeded: max %d requests per %v", RateLimitRequests, RateLimitWindow)
	}
	return nil
}

func GetProvider(providerName string) (LLMProvider, error) {
	apiKey := os.Getenv("BGPIN_LLM_API_KEY")

	switch strings.ToLower(providerName) {
	case "openai":
		return NewOpenAIProvider(apiKey, DefaultTimeout), nil
	case "claude":
		return NewClaudeProvider(apiKey, DefaultTimeout), nil
	case "gemini":
		return NewGeminiProvider(apiKey, DefaultTimeout), nil
	case "ollama":
		return NewOllamaProvider(DefaultTimeout), nil
	default:
		return NewOpenAIProvider(apiKey, DefaultTimeout), nil
	}
}

type OpenAIProvider struct {
	BaseProvider
}

func NewOpenAIProvider(apiKey string, timeout time.Duration) *OpenAIProvider {
	if apiKey == "" {
		apiKey = os.Getenv("OPENAI_API_KEY")
	}
	return &OpenAIProvider{
		BaseProvider: *NewBaseProvider(timeout),
	}
}

func (p *OpenAIProvider) Name() string {
	return "OpenAI"
}

func (p *OpenAIProvider) Analyze(ctx context.Context, systemPrompt string, data interface{}) (string, error) {
	if err := p.checkRateLimit(); err != nil {
		return "", err
	}

	if p.APIKey == "" {
		return "", fmt.Errorf("OpenAI API key is required")
	}

	dataJSON, err := json.Marshal(data)
	if err != nil {
		return "", fmt.Errorf("failed to marshal data: %w", err)
	}

	messages := []map[string]interface{}{
		{"role": "system", "content": systemPrompt},
		{"role": "user", "content": string(dataJSON)},
	}

	payload := map[string]interface{}{
		"model":      p.Model,
		"messages":   messages,
		"max_tokens": MaxTokens,
	}

	reqBody, err := json.Marshal(payload)
	if err != nil {
		return "", fmt.Errorf("failed to marshal request: %w", err)
	}

	resp, err := p.retryWithBackoff(ctx, func() (*http.Response, error) {
		req, err := http.NewRequestWithContext(ctx, "POST", p.BaseURL+"/chat/completions", strings.NewReader(string(reqBody)))
		if err != nil {
			return nil, fmt.Errorf("failed to create request: %w", err)
		}

		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("Authorization", "Bearer "+p.APIKey)

		client := &http.Client{Timeout: p.Timeout}
		return client.Do(req)
	})

	if err != nil {
		return "", fmt.Errorf("OpenAI API request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return "", fmt.Errorf("OpenAI API error %d: %s", resp.StatusCode, string(body))
	}

	var result map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return "", fmt.Errorf("failed to decode response: %w", err)
	}

	choices, ok := result["choices"].([]interface{})
	if !ok || len(choices) == 0 {
		return "", fmt.Errorf("invalid response format: no choices")
	}

	firstChoice, ok := choices[0].(map[string]interface{})
	if !ok {
		return "", fmt.Errorf("invalid response format: invalid choice")
	}

	message, ok := firstChoice["message"].(map[string]interface{})
	if !ok {
		return "", fmt.Errorf("invalid response format: no message")
	}

	content, ok := message["content"].(string)
	if !ok {
		return "", fmt.Errorf("invalid response format: no content")
	}

	return content, nil
}

type ClaudeProvider struct {
	BaseProvider
}

func NewClaudeProvider(apiKey string, timeout time.Duration) *ClaudeProvider {
	if apiKey == "" {
		apiKey = os.Getenv("ANTHROPIC_API_KEY")
	}
	provider := &ClaudeProvider{
		BaseProvider: *NewBaseProvider(timeout),
	}
	provider.APIKey = apiKey
	provider.BaseURL = "https://api.anthropic.com/v1"
	provider.Model = "claude-3-sonnet-20240229"
	return provider
}

func (p *ClaudeProvider) Name() string {
	return "Claude"
}

func (p *ClaudeProvider) Analyze(ctx context.Context, systemPrompt string, data interface{}) (string, error) {
	if err := p.checkRateLimit(); err != nil {
		return "", err
	}

	if p.APIKey == "" {
		return "", fmt.Errorf("Claude API key is required")
	}

	dataJSON, err := json.Marshal(data)
	if err != nil {
		return "", fmt.Errorf("failed to marshal data: %w", err)
	}

	messages := []map[string]interface{}{
		{"role": "user", "content": string(dataJSON)},
	}

	payload := map[string]interface{}{
		"model":      p.Model,
		"messages":   messages,
		"max_tokens": MaxTokens,
		"system":     systemPrompt,
	}

	reqBody, err := json.Marshal(payload)
	if err != nil {
		return "", fmt.Errorf("failed to marshal request: %w", err)
	}

	resp, err := p.retryWithBackoff(ctx, func() (*http.Response, error) {
		req, err := http.NewRequestWithContext(ctx, "POST", p.BaseURL+"/messages", strings.NewReader(string(reqBody)))
		if err != nil {
			return nil, fmt.Errorf("failed to create request: %w", err)
		}

		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("x-api-key", p.APIKey)
		req.Header.Set("anthropic-version", "2023-06-01")

		client := &http.Client{Timeout: p.Timeout}
		return client.Do(req)
	})

	if err != nil {
		return "", fmt.Errorf("Claude API request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return "", fmt.Errorf("Claude API error %d: %s", resp.StatusCode, string(body))
	}

	var result map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return "", fmt.Errorf("failed to decode response: %w", err)
	}

	content, ok := result["content"].([]interface{})
	if !ok || len(content) == 0 {
		return "", fmt.Errorf("invalid response format: no content")
	}

	firstContent, ok := content[0].(map[string]interface{})
	if !ok {
		return "", fmt.Errorf("invalid response format: invalid content")
	}

	text, ok := firstContent["text"].(string)
	if !ok {
		return "", fmt.Errorf("invalid response format: no text")
	}

	return text, nil
}

type GeminiProvider struct {
	BaseProvider
}

func NewGeminiProvider(apiKey string, timeout time.Duration) *GeminiProvider {
	if apiKey == "" {
		apiKey = os.Getenv("GOOGLE_API_KEY")
	}
	provider := &GeminiProvider{
		BaseProvider: *NewBaseProvider(timeout),
	}
	provider.APIKey = apiKey
	provider.BaseURL = "https://generativelanguage.googleapis.com/v1beta"
	provider.Model = "gemini-pro"
	return provider
}

func (p *GeminiProvider) Name() string {
	return "Gemini"
}

func (p *GeminiProvider) Analyze(ctx context.Context, systemPrompt string, data interface{}) (string, error) {
	if err := p.checkRateLimit(); err != nil {
		return "", err
	}

	if p.APIKey == "" {
		return "", fmt.Errorf("Gemini API key is required")
	}

	dataJSON, err := json.Marshal(data)
	if err != nil {
		return "", fmt.Errorf("failed to marshal data: %w", err)
	}

	prompt := systemPrompt + "\n\n" + string(dataJSON)
	url := fmt.Sprintf("%s/models/%s:generateContent?key=%s", p.BaseURL, p.Model, p.APIKey)

	payload := map[string]interface{}{
		"contents": []map[string]interface{}{
			{
				"parts": []map[string]interface{}{
					{"text": prompt},
				},
			},
		},
	}

	reqBody, err := json.Marshal(payload)
	if err != nil {
		return "", fmt.Errorf("failed to marshal request: %w", err)
	}

	resp, err := p.retryWithBackoff(ctx, func() (*http.Response, error) {
		req, err := http.NewRequestWithContext(ctx, "POST", url, strings.NewReader(string(reqBody)))
		if err != nil {
			return nil, fmt.Errorf("failed to create request: %w", err)
		}

		req.Header.Set("Content-Type", "application/json")

		client := &http.Client{Timeout: p.Timeout}
		return client.Do(req)
	})

	if err != nil {
		return "", fmt.Errorf("Gemini API request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return "", fmt.Errorf("Gemini API error %d: %s", resp.StatusCode, string(body))
	}

	var result map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return "", fmt.Errorf("failed to decode response: %w", err)
	}

	candidates, ok := result["candidates"].([]interface{})
	if !ok || len(candidates) == 0 {
		return "", fmt.Errorf("invalid response format: no candidates")
	}

	firstCandidate, ok := candidates[0].(map[string]interface{})
	if !ok {
		return "", fmt.Errorf("invalid response format: invalid candidate")
	}

	content, ok := firstCandidate["content"].(map[string]interface{})
	if !ok {
		return "", fmt.Errorf("invalid response format: no content")
	}

	parts, ok := content["parts"].([]interface{})
	if !ok || len(parts) == 0 {
		return "", fmt.Errorf("invalid response format: no parts")
	}

	firstPart, ok := parts[0].(map[string]interface{})
	if !ok {
		return "", fmt.Errorf("invalid response format: invalid part")
	}

	text, ok := firstPart["text"].(string)
	if !ok {
		return "", fmt.Errorf("invalid response format: no text")
	}

	return text, nil
}

type OllamaProvider struct {
	BaseProvider
}

func NewOllamaProvider(timeout time.Duration) *OllamaProvider {
	ollamaURL := os.Getenv("OLLAMA_HOST")
	if ollamaURL == "" {
		ollamaURL = "http://10.1.254.32:11434"
	}

	provider := &OllamaProvider{
		BaseProvider: *NewBaseProvider(timeout),
	}
	provider.BaseURL = ollamaURL
	provider.Model = "llama3:latest"
	// Ollama pode ser mais lento, aumentar timeout
	if provider.Timeout < 180*time.Second {
		provider.Timeout = 180 * time.Second
	}
	return provider
}

func (p *OllamaProvider) Name() string {
	return "Ollama"
}

func (p *OllamaProvider) Analyze(ctx context.Context, systemPrompt string, data interface{}) (string, error) {
	if err := p.checkRateLimit(); err != nil {
		return "", err
	}

	dataJSON, err := json.Marshal(data)
	if err != nil {
		return "", fmt.Errorf("failed to marshal data: %w", err)
	}

	prompt := systemPrompt + "\n\n" + string(dataJSON)

	payload := map[string]interface{}{
		"model":  p.Model,
		"prompt": prompt,
		"stream": false,
	}

	reqBody, err := json.Marshal(payload)
	if err != nil {
		return "", fmt.Errorf("failed to marshal request: %w", err)
	}

	url := p.BaseURL + "/api/generate"

	resp, err := p.retryWithBackoff(ctx, func() (*http.Response, error) {
		req, err := http.NewRequestWithContext(ctx, "POST", url, strings.NewReader(string(reqBody)))
		if err != nil {
			return nil, fmt.Errorf("failed to create request: %w", err)
		}

		req.Header.Set("Content-Type", "application/json")

		client := &http.Client{Timeout: p.Timeout}
		return client.Do(req)
	})

	if err != nil {
		return "", fmt.Errorf("Ollama connection failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return "", fmt.Errorf("Ollama API error %d: %s", resp.StatusCode, string(body))
	}

	var result map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return "", fmt.Errorf("failed to decode response: %w", err)
	}

	response, ok := result["response"].(string)
	if !ok {
		return "", fmt.Errorf("invalid response format: no response field")
	}

	return response, nil
}
