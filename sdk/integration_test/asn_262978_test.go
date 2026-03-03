package integration_test

import (
	"context"
	"testing"
	"time"

	"github.com/bgpin/bgpin/sdk"
)

const (
	// ASN 262978 - Used for all integration tests
	testASN = 262978
)

func TestGetASNInfo_262978(t *testing.T) {
	client := sdk.NewDefaultClient()
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	info, err := client.GetASNInfo(ctx, testASN)
	if err != nil {
		t.Fatalf("Failed to get ASN info: %v", err)
	}

	if info.ASN != testASN {
		t.Errorf("Expected ASN %d, got %d", testASN, info.ASN)
	}

	t.Logf("ASN Info for AS%d:", testASN)
	t.Logf("  Holder: %s", info.Holder)
	t.Logf("  Announced: %v", info.Announced)
	t.Logf("  Block: %s", info.Block)
	t.Logf("  Description: %s", info.Description)
}

func TestGetASNNeighbors_262978(t *testing.T) {
	client := sdk.NewDefaultClient()
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	neighbors, err := client.GetASNNeighbors(ctx, testASN)
	if err != nil {
		t.Fatalf("Failed to get ASN neighbors: %v", err)
	}

	if neighbors.ASN != testASN {
		t.Errorf("Expected ASN %d, got %d", testASN, neighbors.ASN)
	}

	t.Logf("ASN Neighbors for AS%d:", testASN)
	t.Logf("  Total neighbors: %d", len(neighbors.Neighbors))
	
	if len(neighbors.Neighbors) > 0 {
		t.Logf("  Sample neighbors:")
		for i, neighbor := range neighbors.Neighbors {
			if i >= 5 { // Show only first 5
				break
			}
			t.Logf("    AS%d (type: %s, power: %d)", neighbor.ASN, neighbor.Type, neighbor.Power)
		}
	}
}

func TestGetAnnouncedPrefixes_262978(t *testing.T) {
	client := sdk.NewDefaultClient()
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	prefixes, err := client.GetAnnouncedPrefixes(ctx, testASN)
	if err != nil {
		t.Fatalf("Failed to get announced prefixes: %v", err)
	}

	if prefixes.ASN != testASN {
		t.Errorf("Expected ASN %d, got %d", testASN, prefixes.ASN)
	}

	t.Logf("Announced Prefixes for AS%d:", testASN)
	t.Logf("  Total prefixes: %d", len(prefixes.Prefixes))
	
	if len(prefixes.Prefixes) > 0 {
		t.Logf("  Sample prefixes:")
		for i, prefix := range prefixes.Prefixes {
			if i >= 5 { // Show only first 5
				break
			}
			t.Logf("    %s", prefix.Prefix)
		}
	}
}

func TestGetPrefixOverview_262978(t *testing.T) {
	client := sdk.NewDefaultClient()
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// First get a prefix from AS262978
	prefixes, err := client.GetAnnouncedPrefixes(ctx, testASN)
	if err != nil {
		t.Fatalf("Failed to get announced prefixes: %v", err)
	}

	if len(prefixes.Prefixes) == 0 {
		t.Skip("No prefixes announced by AS262978, skipping prefix overview test")
	}

	testPrefix := prefixes.Prefixes[0].Prefix
	t.Logf("Testing prefix overview for: %s", testPrefix)

	overview, err := client.GetPrefixOverview(ctx, testPrefix)
	if err != nil {
		t.Fatalf("Failed to get prefix overview: %v", err)
	}

	t.Logf("Prefix Overview for %s:", testPrefix)
	t.Logf("  Actual Prefix: %s", overview.ActualPrefix)
	t.Logf("  Is Less Specific: %v", overview.IsLessSpec)
	t.Logf("  ASNs: %v", overview.ASNs)

	// Verify AS262978 is in the ASN list
	found := false
	for _, asn := range overview.ASNs {
		if asn == testASN {
			found = true
			break
		}
	}

	if !found {
		t.Errorf("Expected AS%d to be in the ASN list for prefix %s", testASN, testPrefix)
	}
}

func TestGetRISPeers_262978(t *testing.T) {
	client := sdk.NewDefaultClient()
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	peers, err := client.GetRISPeers(ctx, testASN)
	if err != nil {
		t.Fatalf("Failed to get RIS peers: %v", err)
	}

	t.Logf("RIS Peers for AS%d:", testASN)
	t.Logf("  Total peers: %d", len(peers))
	
	if len(peers) > 0 {
		t.Logf("  Sample peers:")
		for i, peer := range peers {
			if i >= 5 { // Show only first 5
				break
			}
			t.Logf("    %s", peer)
		}
	}
}

func TestRateLimiting(t *testing.T) {
	config := sdk.DefaultConfig()
	config.RateLimit = 2 // 2 requests per second
	client := sdk.NewClient(config)

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	start := time.Now()
	
	// Make 5 requests
	for i := 0; i < 5; i++ {
		_, err := client.GetASNInfo(ctx, testASN)
		if err != nil {
			t.Fatalf("Request %d failed: %v", i+1, err)
		}
	}
	
	elapsed := time.Since(start)
	
	// With rate limit of 2 req/s, 5 requests should take at least 2 seconds
	minExpected := 2 * time.Second
	if elapsed < minExpected {
		t.Errorf("Rate limiting not working properly. Expected at least %v, got %v", minExpected, elapsed)
	}
	
	t.Logf("Rate limiting test: 5 requests took %v (expected >= %v)", elapsed, minExpected)
}

func TestRetryOnError(t *testing.T) {
	config := sdk.DefaultConfig()
	config.RetryMax = 2
	config.RetryWaitMin = 100 * time.Millisecond
	config.RetryWaitMax = 500 * time.Millisecond
	client := sdk.NewClient(config)

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// Test with invalid ASN to trigger potential errors
	_, err := client.GetASNInfo(ctx, -1)
	if err == nil {
		t.Error("Expected error for invalid ASN, got nil")
	}
	
	t.Logf("Error handling test passed: %v", err)
}

func TestContextTimeout(t *testing.T) {
	client := sdk.NewDefaultClient()
	
	// Create a context with very short timeout
	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Nanosecond)
	defer cancel()
	
	time.Sleep(10 * time.Millisecond) // Ensure context is expired
	
	_, err := client.GetASNInfo(ctx, testASN)
	if err == nil {
		t.Error("Expected timeout error, got nil")
	}
	
	t.Logf("Context timeout test passed: %v", err)
}

func TestConcurrentRequests(t *testing.T) {
	client := sdk.NewDefaultClient()
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// Make 3 concurrent requests
	type result struct {
		info *sdk.ASNInfo
		err  error
	}
	
	results := make(chan result, 3)
	
	for i := 0; i < 3; i++ {
		go func() {
			info, err := client.GetASNInfo(ctx, testASN)
			results <- result{info, err}
		}()
	}
	
	// Collect results
	for i := 0; i < 3; i++ {
		res := <-results
		if res.err != nil {
			t.Errorf("Concurrent request %d failed: %v", i+1, res.err)
		}
		if res.info != nil && res.info.ASN != testASN {
			t.Errorf("Expected ASN %d, got %d", testASN, res.info.ASN)
		}
	}
	
	t.Log("Concurrent requests test passed")
}
