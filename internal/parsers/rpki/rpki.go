package rpki

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/bgpin/bgpin/internal/core/bgp"
)

type RIPKIValidator struct {
	client *http.Client
}

type Config struct {
	Timeout time.Duration
}

func NewRIPKIValidator(config Config) *RIPKIValidator {
	timeout := config.Timeout
	if timeout == 0 {
		timeout = 30 * time.Second
	}

	return &RIPKIValidator{
		client: &http.Client{
			Timeout: timeout,
		},
	}
}

type ValidationResult struct {
	Prefix      string    `json:"prefix"`
	ASN         int       `json:"asn"`
	State       string    `json:"state"`
	Description string    `json:"description"`
	MatchedVRPs []VRP     `json:"matched_vrps"`
	Timestamp   time.Time `json:"timestamp"`
}

type VRP struct {
	ASN       string `json:"asn"`
	Prefix    string `json:"prefix"`
	MaxLength string `json:"max_length"`
}

func (v *RIPKIValidator) ValidateRoute(ctx context.Context, asn int, prefix string) (*ValidationResult, error) {
	url := fmt.Sprintf("https://rpki-validator.ripe.net/api/v1/validity/%d/%s", asn, prefix)

	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return nil, err
	}

	resp, err := v.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	var rpkiResp struct {
		ValidatedRoute struct {
			Route struct {
				OriginASN string `json:"origin_asn"`
				Prefix    string `json:"prefix"`
			} `json:"route"`
			Validity struct {
				State       string `json:"state"`
				Description string `json:"description"`
				VRPs        struct {
					Matched         []VRP `json:"matched"`
					UnmatchedAS     []VRP `json:"unmatched_as"`
					UnmatchedLength []VRP `json:"unmatched_length"`
				} `json:"VRPs"`
			} `json:"validity"`
		} `json:"validated_route"`
		GeneratedTime string `json:"generatedTime"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&rpkiResp); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	result := &ValidationResult{
		Prefix:      rpkiResp.ValidatedRoute.Route.Prefix,
		ASN:         asn,
		State:       rpkiResp.ValidatedRoute.Validity.State,
		Description: rpkiResp.ValidatedRoute.Validity.Description,
		MatchedVRPs: rpkiResp.ValidatedRoute.Validity.VRPs.Matched,
	}

	if result.State == "" {
		result.State = "unknown"
		result.Description = "No RPKI validation data available"
	}

	if t, err := time.Parse("2006-01-02T15:04:05Z", rpkiResp.GeneratedTime); err == nil {
		result.Timestamp = t
	} else {
		result.Timestamp = time.Now()
	}

	return result, nil
}

func (v *RIPKIValidator) GetROAs(ctx context.Context, prefix string) ([]ROA, error) {
	url := fmt.Sprintf("https://rpki-validator.ripe.net/api/v1/prefixes/%s", prefix)

	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return nil, err
	}

	resp, err := v.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	var rpkiResp struct {
		ROAs []ROA `json:"roas"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&rpkiResp); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	return rpkiResp.ROAs, nil
}

type ROA struct {
	ASN       string `json:"asn"`
	Prefix    string `json:"prefix"`
	MaxLength int    `json:"max_length"`
	URI       string `json:"uri"`
	Trusted   bool   `json:"trusted"`
}

func (v *RIPKIValidator) ValidateASN(ctx context.Context, asn int) (*bgp.ASnInfo, error) {
	url := fmt.Sprintf("https://rpki-validator.ripe.net/api/v1/asn/%d", asn)

	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return nil, err
	}

	resp, err := v.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	var rpkiResp struct {
		ASNs []struct {
			ASN  int   `json:"asn"`
			ROAs []ROA `json:"roas"`
		} `json:"asns"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&rpkiResp); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	if len(rpkiResp.ASNs) == 0 {
		return &bgp.ASnInfo{
			ASN:      asn,
			Valid:    false,
			ROACount: 0,
		}, nil
	}

	return &bgp.ASnInfo{
		ASN:      asn,
		Valid:    true,
		ROACount: len(rpkiResp.ASNs[0].ROAs),
	}, nil
}
