package main

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/bgpin/bgpin/sdk"
)

func main() {
	fmt.Println("=== RIPE RIS SDK Demo ===")
	fmt.Println("Testing with ASN 262978\n")

	// Create client with default configuration
	client := sdk.NewDefaultClient()

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	asn := 262978

	// 1. Get ASN Information
	fmt.Println("1. Getting ASN Information...")
	info, err := client.GetASNInfo(ctx, asn)
	if err != nil {
		log.Fatalf("Failed to get ASN info: %v", err)
	}
	fmt.Printf("   ASN: %d\n", info.ASN)
	fmt.Printf("   Holder: %s\n", info.Holder)
	fmt.Printf("   Announced: %v\n", info.Announced)
	fmt.Printf("   Block: %s\n\n", info.Block)

	// 2. Get ASN Neighbors
	fmt.Println("2. Getting ASN Neighbors...")
	neighbors, err := client.GetASNNeighbors(ctx, asn)
	if err != nil {
		log.Fatalf("Failed to get neighbors: %v", err)
	}
	fmt.Printf("   Total neighbors: %d\n", len(neighbors.Neighbors))
	fmt.Println("   Top 3 neighbors:")
	for i, neighbor := range neighbors.Neighbors {
		if i >= 3 {
			break
		}
		fmt.Printf("     - AS%d (type: %s, power: %d)\n", neighbor.ASN, neighbor.Type, neighbor.Power)
	}
	fmt.Println()

	// 3. Get Announced Prefixes
	fmt.Println("3. Getting Announced Prefixes...")
	prefixes, err := client.GetAnnouncedPrefixes(ctx, asn)
	if err != nil {
		log.Fatalf("Failed to get prefixes: %v", err)
	}
	fmt.Printf("   Total prefixes: %d\n", len(prefixes.Prefixes))
	fmt.Println("   First 3 prefixes:")
	for i, prefix := range prefixes.Prefixes {
		if i >= 3 {
			break
		}
		fmt.Printf("     - %s\n", prefix.Prefix)
	}
	fmt.Println()

	// 4. Get Prefix Overview
	if len(prefixes.Prefixes) > 0 {
		fmt.Println("4. Getting Prefix Overview...")
		testPrefix := prefixes.Prefixes[0].Prefix
		overview, err := client.GetPrefixOverview(ctx, testPrefix)
		if err != nil {
			log.Printf("   Failed to get prefix overview: %v", err)
		} else {
			fmt.Printf("   Prefix: %s\n", overview.Prefix)
			fmt.Printf("   ASNs announcing: %v\n", overview.ASNs)
			fmt.Printf("   Is Less Specific: %v\n", overview.IsLessSpec)
		}
		fmt.Println()
	}

	// 5. Get RIS Peers
	fmt.Println("5. Getting RIS Peers...")
	peers, err := client.GetRISPeers(ctx, asn)
	if err != nil {
		log.Fatalf("Failed to get RIS peers: %v", err)
	}
	fmt.Printf("   Total peers: %d\n", len(peers))
	fmt.Println("   First 3 peers:")
	for i, peer := range peers {
		if i >= 3 {
			break
		}
		fmt.Printf("     - %s\n", peer)
	}
	fmt.Println()

	fmt.Println("=== Demo completed successfully! ===")
}
