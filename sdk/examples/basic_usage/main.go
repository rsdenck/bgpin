package main

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/bgpin/bgpin/sdk"
)

func main() {
	// Create client with default configuration
	client := sdk.NewDefaultClient()

	// Or create with custom configuration
	customConfig := sdk.Config{
		Timeout:      30 * time.Second,
		RateLimit:    5, // 5 requests per second
		RetryMax:     3,
		RetryWaitMin: 1 * time.Second,
		RetryWaitMax: 10 * time.Second,
		UserAgent:    "my-app/1.0",
		BaseURL:      "https://stat.ripe.net/data",
	}
	customClient := sdk.NewClient(customConfig)
	_ = customClient // Use custom client if needed

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// Example ASN
	asn := 262978

	// 1. Get ASN Information
	fmt.Println("=== ASN Information ===")
	info, err := client.GetASNInfo(ctx, asn)
	if err != nil {
		log.Fatalf("Failed to get ASN info: %v", err)
	}
	fmt.Printf("ASN: %d\n", info.ASN)
	fmt.Printf("Holder: %s\n", info.Holder)
	fmt.Printf("Announced: %v\n", info.Announced)
	fmt.Printf("Block: %s\n", info.Block)
	fmt.Printf("Description: %s\n\n", info.Description)

	// 2. Get ASN Neighbors
	fmt.Println("=== ASN Neighbors ===")
	neighbors, err := client.GetASNNeighbors(ctx, asn)
	if err != nil {
		log.Fatalf("Failed to get neighbors: %v", err)
	}
	fmt.Printf("Total neighbors: %d\n", len(neighbors.Neighbors))
	for i, neighbor := range neighbors.Neighbors {
		if i >= 5 { // Show only first 5
			break
		}
		fmt.Printf("  AS%d (type: %s, power: %d)\n", neighbor.ASN, neighbor.Type, neighbor.Power)
	}
	fmt.Println()

	// 3. Get Announced Prefixes
	fmt.Println("=== Announced Prefixes ===")
	prefixes, err := client.GetAnnouncedPrefixes(ctx, asn)
	if err != nil {
		log.Fatalf("Failed to get prefixes: %v", err)
	}
	fmt.Printf("Total prefixes: %d\n", len(prefixes.Prefixes))
	for i, prefix := range prefixes.Prefixes {
		if i >= 5 { // Show only first 5
			break
		}
		fmt.Printf("  %s\n", prefix.Prefix)
	}
	fmt.Println()

	// 4. Get Prefix Overview (if prefixes exist)
	if len(prefixes.Prefixes) > 0 {
		fmt.Println("=== Prefix Overview ===")
		testPrefix := prefixes.Prefixes[0].Prefix
		overview, err := client.GetPrefixOverview(ctx, testPrefix)
		if err != nil {
			log.Printf("Failed to get prefix overview: %v", err)
		} else {
			fmt.Printf("Prefix: %s\n", overview.Prefix)
			fmt.Printf("Actual Prefix: %s\n", overview.ActualPrefix)
			fmt.Printf("Is Less Specific: %v\n", overview.IsLessSpec)
			fmt.Printf("ASNs: %v\n", overview.ASNs)
		}
		fmt.Println()
	}

	// 5. Get RIS Peers
	fmt.Println("=== RIS Peers ===")
	peers, err := client.GetRISPeers(ctx, asn)
	if err != nil {
		log.Fatalf("Failed to get RIS peers: %v", err)
	}
	fmt.Printf("Total peers: %d\n", len(peers))
	for i, peer := range peers {
		if i >= 5 { // Show only first 5
			break
		}
		fmt.Printf("  %s\n", peer)
	}
}
