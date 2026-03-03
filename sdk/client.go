package sdk

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

// Client is the main SDK client for RIPE RIS API
type Client struct {
	httpClient  *http.Client
	rateLimiter *RateLimiter
	config      Config
}

// NewClient creates a new RIPE RIS SDK client
func NewClient(config Config) *Client {
	return &Client{
		httpClient: &http.Client{
			Timeout: config.Timeout,
		},
		rateLimiter: NewRateLimiter(config.RateLimit),
		config:      config,
	}
}

// NewDefaultClient creates a client with default configuration
func NewDefaultClient() *Client {
	return NewClient(DefaultConfig())
}

// doRequest performs an HTTP request with rate limiting and retry logic
func (c *Client) doRequest(ctx context.Context, endpoint string) (*RIPEResponse, error) {
	// Apply rate limiting
	if err := c.rateLimiter.Wait(ctx); err != nil {
		return nil, fmt.Errorf("rate limit wait failed: %w", err)
	}
	
	url := fmt.Sprintf("%s/%s", c.config.BaseURL, endpoint)
	
	var response *RIPEResponse
	policy := RetryPolicy{
		MaxRetries: c.config.RetryMax,
		MinWait:    c.config.RetryWaitMin,
		MaxWait:    c.config.RetryWaitMax,
	}
	
	err := RetryWithBackoff(ctx, policy, func() error {
		req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
		if err != nil {
			return fmt.Errorf("failed to create request: %w", err)
		}
		
		req.Header.Set("User-Agent", c.config.UserAgent)
		req.Header.Set("Accept", "application/json")
		
		resp, err := c.httpClient.Do(req)
		if err != nil {
			return fmt.Errorf("request failed: %w", err)
		}
		defer resp.Body.Close()
		
		body, err := io.ReadAll(resp.Body)
		if err != nil {
			return fmt.Errorf("failed to read response body: %w", err)
		}
		
		if resp.StatusCode != http.StatusOK {
			if ShouldRetry(resp.StatusCode) {
				return WrapAPIError(resp.StatusCode, endpoint, string(body))
			}
			return WrapAPIError(resp.StatusCode, endpoint, string(body))
		}
		
		var ripeResp RIPEResponse
		if err := json.Unmarshal(body, &ripeResp); err != nil {
			return fmt.Errorf("failed to unmarshal response: %w", err)
		}
		
		response = &ripeResp
		return nil
	})
	
	if err != nil {
		return nil, err
	}
	
	return response, nil
}

// GetASNNeighbors retrieves ASN neighbors information
func (c *Client) GetASNNeighbors(ctx context.Context, asn int) (*ASNNeighbors, error) {
	if asn <= 0 {
		return nil, ErrInvalidASN
	}
	
	endpoint := fmt.Sprintf("asn-neighbours/data.json?resource=AS%d", asn)
	resp, err := c.doRequest(ctx, endpoint)
	if err != nil {
		return nil, err
	}
	
	// Parse the data field
	dataBytes, err := json.Marshal(resp.Data)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal data: %w", err)
	}
	
	var result struct {
		Neighbours []NeighborRelation `json:"neighbours"`
	}
	
	if err := json.Unmarshal(dataBytes, &result); err != nil {
		return nil, fmt.Errorf("failed to unmarshal neighbors: %w", err)
	}
	
	return &ASNNeighbors{
		ASN:       asn,
		Neighbors: result.Neighbours,
		QueryTime: time.Now(),
	}, nil
}

// GetAnnouncedPrefixes retrieves all prefixes announced by an ASN
func (c *Client) GetAnnouncedPrefixes(ctx context.Context, asn int) (*AnnouncedPrefixes, error) {
	if asn <= 0 {
		return nil, ErrInvalidASN
	}
	
	endpoint := fmt.Sprintf("announced-prefixes/data.json?resource=AS%d", asn)
	resp, err := c.doRequest(ctx, endpoint)
	if err != nil {
		return nil, err
	}
	
	dataBytes, err := json.Marshal(resp.Data)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal data: %w", err)
	}
	
	var result struct {
		Prefixes []struct {
			Prefix string `json:"prefix"`
		} `json:"prefixes"`
	}
	
	if err := json.Unmarshal(dataBytes, &result); err != nil {
		return nil, fmt.Errorf("failed to unmarshal prefixes: %w", err)
	}
	
	prefixes := make([]Prefix, len(result.Prefixes))
	for i, p := range result.Prefixes {
		prefixes[i] = Prefix{Prefix: p.Prefix}
	}
	
	return &AnnouncedPrefixes{
		ASN:      asn,
		Prefixes: prefixes,
	}, nil
}

// GetPrefixOverview retrieves information about a specific prefix
func (c *Client) GetPrefixOverview(ctx context.Context, prefix string) (*PrefixOverview, error) {
	if prefix == "" {
		return nil, ErrInvalidPrefix
	}
	
	endpoint := fmt.Sprintf("prefix-overview/data.json?resource=%s", prefix)
	resp, err := c.doRequest(ctx, endpoint)
	if err != nil {
		return nil, err
	}
	
	dataBytes, err := json.Marshal(resp.Data)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal data: %w", err)
	}
	
	var result struct {
		Resource string `json:"resource"`
		ASNs     []struct {
			ASN int `json:"asn"`
		} `json:"asns"`
		IsLessSpecific bool   `json:"is_less_specific"`
		ActualPrefix   string `json:"actual_prefix"`
	}
	
	if err := json.Unmarshal(dataBytes, &result); err != nil {
		return nil, fmt.Errorf("failed to unmarshal prefix overview: %w", err)
	}
	
	asns := make([]int, len(result.ASNs))
	for i, a := range result.ASNs {
		asns[i] = a.ASN
	}
	
	return &PrefixOverview{
		Prefix:       result.Resource,
		ASNs:         asns,
		IsLessSpec:   result.IsLessSpecific,
		ActualPrefix: result.ActualPrefix,
		QueryTime:    time.Now(),
	}, nil
}

// GetASNInfo retrieves general information about an ASN
func (c *Client) GetASNInfo(ctx context.Context, asn int) (*ASNInfo, error) {
	if asn <= 0 {
		return nil, ErrInvalidASN
	}
	
	endpoint := fmt.Sprintf("as-overview/data.json?resource=AS%d", asn)
	resp, err := c.doRequest(ctx, endpoint)
	if err != nil {
		return nil, err
	}
	
	dataBytes, err := json.Marshal(resp.Data)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal data: %w", err)
	}
	
	var result struct {
		Holder      string `json:"holder"`
		Announced   bool   `json:"announced"`
		Block       struct {
			Resource string `json:"resource"`
		} `json:"block"`
		Description string `json:"description"`
	}
	
	if err := json.Unmarshal(dataBytes, &result); err != nil {
		return nil, fmt.Errorf("failed to unmarshal ASN info: %w", err)
	}
	
	return &ASNInfo{
		ASN:         asn,
		Holder:      result.Holder,
		Announced:   result.Announced,
		Block:       result.Block.Resource,
		Description: result.Description,
	}, nil
}

// GetRISPeers retrieves RIS peers for a specific resource
func (c *Client) GetRISPeers(ctx context.Context, asn int) ([]string, error) {
	if asn <= 0 {
		return nil, ErrInvalidASN
	}
	
	endpoint := fmt.Sprintf("ris-peers/data.json?resource=AS%d", asn)
	resp, err := c.doRequest(ctx, endpoint)
	if err != nil {
		return nil, err
	}
	
	dataBytes, err := json.Marshal(resp.Data)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal data: %w", err)
	}
	
	// The data is a map of RRC names to peer arrays
	var result struct {
		Peers map[string][]struct {
			ASN            string `json:"asn"`
			IP             string `json:"ip"`
			V4PrefixCount  int    `json:"v4_prefix_count"`
			V6PrefixCount  int    `json:"v6_prefix_count"`
		} `json:"peers"`
	}
	
	if err := json.Unmarshal(dataBytes, &result); err != nil {
		return nil, fmt.Errorf("failed to unmarshal RIS peers: %w", err)
	}
	
	var peers []string
	for rrc, peerList := range result.Peers {
		for _, p := range peerList {
			peers = append(peers, fmt.Sprintf("[%s] AS%s - %s (v4: %d, v6: %d)", 
				rrc, p.ASN, p.IP, p.V4PrefixCount, p.V6PrefixCount))
		}
	}
	
	return peers, nil
}
