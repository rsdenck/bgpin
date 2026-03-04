package http

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/bgpin/bgpin/internal/core/bgp"
)

type RIPEParser struct {
	client *http.Client
}

func NewRIPEParser() *RIPEParser {
	return &RIPEParser{
		client: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

func (p *RIPEParser) QueryBGP(ctx context.Context, prefix string) (*bgp.LookupResult, error) {
	result := &bgp.LookupResult{
		Prefix:    prefix,
		Timestamp: time.Now(),
		Routes:    []bgp.Route{},
	}

	routes, err := p.getRoutesFromRIPE(ctx, prefix)
	if err != nil {
		return result, err
	}

	result.Routes = routes
	return result, nil
}

func (p *RIPEParser) getRoutesFromRIPE(ctx context.Context, prefix string) ([]bgp.Route, error) {
	url := fmt.Sprintf("https://stat.ripe.net/data/bgp-updates/data.json?resource=%s&starttime=%s&endtime=%s",
		prefix,
		time.Now().Add(-24*time.Hour).Format("2006-01-02T15:04:05Z"),
		time.Now().Format("2006-01-02T15:04:05Z"),
	)

	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return nil, err
	}

	resp, err := p.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	var ripeResp struct {
		Data struct {
			Updates []struct {
				Timestamp string `json:"timestamp"`
				Attrs     struct {
					SourceID     string   `json:"source_id"`
					TargetPrefix string   `json:"target_prefix"`
					Path         []int    `json:"path"`
					NextHop      string   `json:"next_hop"`
					Community    []string `json:"community"`
				} `json:"attrs"`
			} `json:"updates"`
		} `json:"data"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&ripeResp); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	routes := make([]bgp.Route, 0)
	seenRoutes := make(map[string]bool)

	for _, update := range ripeResp.Data.Updates {
		if update.Attrs.TargetPrefix == "" {
			continue
		}

		key := update.Attrs.TargetPrefix
		if update.Attrs.NextHop != "" {
			key = key + update.Attrs.NextHop
		} else if len(update.Attrs.Path) > 0 {
			key = key + fmt.Sprintf("%v", update.Attrs.Path)
		}

		if seenRoutes[key] {
			continue
		}
		seenRoutes[key] = true

		nextHop := update.Attrs.NextHop
		if nextHop == "" {
			nextHop = "N/A (RIPE RIS)"
		}

		route := bgp.Route{
			Prefix:    update.Attrs.TargetPrefix,
			NextHop:   nextHop,
			ASPath:    update.Attrs.Path,
			Best:      len(routes) == 0,
			Community: update.Attrs.Community,
		}

		if len(update.Attrs.Path) > 0 {
			route.Origin = "IGP"
		}

		routes = append(routes, route)
	}

	return routes, nil
}

func parseASPath(asPath string) []int {
	parts := make([]int, 0)
	for _, s := range splitAndTrim(asPath) {
		if asn, err := strconv.Atoi(s); err == nil && asn > 0 {
			parts = append(parts, asn)
		}
	}
	return parts
}

func splitAndTrim(s string) []string {
	var result []string
	var current string
	for _, c := range s {
		if c == ' ' || c == ',' {
			if current != "" {
				result = append(result, current)
				current = ""
			}
		} else {
			current += string(c)
		}
	}
	if current != "" {
		result = append(result, current)
	}
	return result
}
