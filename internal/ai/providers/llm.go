package providers

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"strings"
	"time"
)

type LLMProvider interface {
	Name() string
	Analyze(ctx context.Context, prompt string, data interface{}) (string, error)
}

type BaseProvider struct {
	APIKey  string
	BaseURL string
	Model   string
	Timeout time.Duration
}

func NewBaseProvider() *BaseProvider {
	return &BaseProvider{
		Timeout: 60 * time.Second,
		Model:   "gpt-4",
	}
}

func GetProvider(providerName string) (LLMProvider, error) {
	apiKey := os.Getenv("BGPIN_LLM_API_KEY")

	switch strings.ToLower(providerName) {
	case "openai":
		return NewOpenAIProvider(apiKey), nil
	case "claude":
		return NewClaudeProvider(apiKey), nil
	case "gemini":
		return NewGeminiProvider(apiKey), nil
	case "ollama":
		return NewOllamaProvider(), nil
	default:
		return NewOpenAIProvider(apiKey), nil
	}
}

type OpenAIProvider struct {
	BaseProvider
}

func NewOpenAIProvider(apiKey string) *OpenAIProvider {
	if apiKey == "" {
		apiKey = os.Getenv("OPENAI_API_KEY")
	}
	return &OpenAIProvider{
		BaseProvider: BaseProvider{
			APIKey:  apiKey,
			BaseURL: "https://api.openai.com/v1",
			Model:   "gpt-4",
		},
	}
}

func (p *OpenAIProvider) Name() string {
	return "OpenAI"
}

func (p *OpenAIProvider) Analyze(ctx context.Context, systemPrompt string, data interface{}) (string, error) {
	dataJSON, _ := json.Marshal(data)

	messages := []map[string]interface{}{
		{"role": "system", "content": systemPrompt},
		{"role": "user", "content": string(dataJSON)},
	}

	payload := map[string]interface{}{
		"model":      p.Model,
		"messages":   messages,
		"max_tokens": 2048,
	}

	reqBody, err := json.Marshal(payload)
	if err != nil {
		return "", err
	}

	req, err := http.NewRequestWithContext(ctx, "POST", p.BaseURL+"/chat/completions", strings.NewReader(string(reqBody)))
	if err != nil {
		return "", err
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+p.APIKey)

	client := &http.Client{Timeout: p.Timeout}
	resp, err := client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("OpenAI API error: %d", resp.StatusCode)
	}

	var result map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return "", err
	}

	choices := result["choices"].([]interface{})
	firstChoice := choices[0].(map[string]interface{})
	message := firstChoice["message"].(map[string]interface{})

	return message["content"].(string), nil
}

type ClaudeProvider struct {
	BaseProvider
}

func NewClaudeProvider(apiKey string) *ClaudeProvider {
	if apiKey == "" {
		apiKey = os.Getenv("ANTHROPIC_API_KEY")
	}
	return &ClaudeProvider{
		BaseProvider: BaseProvider{
			APIKey:  apiKey,
			BaseURL: "https://api.anthropic.com/v1",
			Model:   "claude-3-sonnet-20240229",
		},
	}
}

func (p *ClaudeProvider) Name() string {
	return "Claude"
}

func (p *ClaudeProvider) Analyze(ctx context.Context, systemPrompt string, data interface{}) (string, error) {
	dataJSON, _ := json.Marshal(data)

	messages := []map[string]interface{}{
		{"role": "user", "content": string(dataJSON)},
	}

	payload := map[string]interface{}{
		"model":      p.Model,
		"messages":   messages,
		"max_tokens": 2048,
		"system":     systemPrompt,
	}

	reqBody, _ := json.Marshal(payload)

	req, _ := http.NewRequestWithContext(ctx, "POST", p.BaseURL+"/messages", strings.NewReader(string(reqBody)))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("x-api-key", p.APIKey)
	req.Header.Set("anthropic-version", "2023-06-01")

	client := &http.Client{Timeout: p.Timeout}
	resp, err := client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("Claude API error: %d", resp.StatusCode)
	}

	var result map[string]interface{}
	json.NewDecoder(resp.Body).Decode(&result)

	content := result["content"].([]interface{})
	firstContent := content[0].(map[string]interface{})

	return firstContent["text"].(string), nil
}

type GeminiProvider struct {
	BaseProvider
}

func NewGeminiProvider(apiKey string) *GeminiProvider {
	if apiKey == "" {
		apiKey = os.Getenv("GOOGLE_API_KEY")
	}
	return &GeminiProvider{
		BaseProvider: BaseProvider{
			APIKey:  apiKey,
			BaseURL: "https://generativelanguage.googleapis.com/v1beta",
			Model:   "gemini-pro",
		},
	}
}

func (p *GeminiProvider) Name() string {
	return "Gemini"
}

func (p *GeminiProvider) Analyze(ctx context.Context, systemPrompt string, data interface{}) (string, error) {
	dataJSON, _ := json.Marshal(data)

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

	reqBody, _ := json.Marshal(payload)

	req, _ := http.NewRequestWithContext(ctx, "POST", url, strings.NewReader(string(reqBody)))
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{Timeout: p.Timeout}
	resp, err := client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("Gemini API error: %d", resp.StatusCode)
	}

	var result map[string]interface{}
	json.NewDecoder(resp.Body).Decode(&result)

	candidates := result["candidates"].([]interface{})
	firstCandidate := candidates[0].(map[string]interface{})
	content := firstCandidate["content"].(map[string]interface{})
	parts := content["parts"].([]interface{})
	firstPart := parts[0].(map[string]interface{})

	return firstPart["text"].(string), nil
}

type OllamaProvider struct {
	BaseProvider
}

func NewOllamaProvider() *OllamaProvider {
	ollamaURL := os.Getenv("OLLAMA_HOST")
	if ollamaURL == "" {
		ollamaURL = "http://localhost:11434"
	}

	return &OllamaProvider{
		BaseProvider: BaseProvider{
			BaseURL: ollamaURL,
			Model:   "llama2",
		},
	}
}

func (p *OllamaProvider) Name() string {
	return "Ollama"
}

func (p *OllamaProvider) Analyze(ctx context.Context, systemPrompt string, data interface{}) (string, error) {
	dataJSON, _ := json.Marshal(data)

	prompt := systemPrompt + "\n\n" + string(dataJSON)

	payload := map[string]interface{}{
		"model":  p.Model,
		"prompt": prompt,
		"stream": false,
	}

	reqBody, _ := json.Marshal(payload)

	url := p.BaseURL + "/api/generate"
	req, _ := http.NewRequestWithContext(ctx, "POST", url, strings.NewReader(string(reqBody)))
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{Timeout: p.Timeout}
	resp, err := client.Do(req)
	if err != nil {
		return "", fmt.Errorf("Ollama connection failed: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("Ollama API error: %d", resp.StatusCode)
	}

	var result map[string]interface{}
	json.NewDecoder(resp.Body).Decode(&result)

	return result["response"].(string), nil
}
