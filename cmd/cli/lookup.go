package main

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/bgpin/bgpin/internal/adapters/http"
	"github.com/bgpin/bgpin/internal/core/bgp"
	"github.com/bgpin/bgpin/internal/parsers/cisco"
	"github.com/bgpin/bgpin/internal/parsers/junos"
	"github.com/bgpin/bgpin/pkg/config"
	"github.com/spf13/cobra"
)

var prefix string
var lgName string
var outputFormat string

func newLookupCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "lookup [prefix]",
		Short: "Look up a prefix in BGP tables",
		Args:  cobra.ExactArgs(1),
		RunE:  runLookup,
	}

	cmd.Flags().StringVarP(&lgName, "lg", "l", "", "Looking glass name")
	cmd.Flags().StringVarP(&outputFormat, "output", "o", "table", "Output format: table, json, yaml")

	return cmd
}

func runLookup(cmd *cobra.Command, args []string) error {
	prefix = args[0]
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	cfg := GetConfig()

	var targetLG *config.LookingGlass
	if lgName != "" {
		for _, lg := range cfg.LookingGlasses {
			if strings.EqualFold(lg.Name, lgName) {
				targetLG = &lg
				break
			}
		}
		if targetLG == nil {
			return fmt.Errorf("looking glass not found: %s", lgName)
		}
	} else {
		targetLG = &cfg.LookingGlasses[0]
	}

	var result *bgp.LookupResult
	var err error

	switch targetLG.Vendor {
	case "cisco":
		result, err = queryCiscoLG(ctx, targetLG, prefix)
	case "juniper":
		result, err = queryJuniperLG(ctx, targetLG, prefix)
	default:
		result, err = queryCiscoLG(ctx, targetLG, prefix)
	}

	if err != nil {
		return fmt.Errorf("lookup failed: %w", err)
	}

	return printResult(result, outputFormat)
}

func queryCiscoLG(ctx context.Context, lg *config.LookingGlass, prefix string) (*bgp.LookupResult, error) {
	adapter := http.NewHTTPAdapter(lg.URL, 30*time.Second)
	parser := cisco.NewParser()

	output, err := adapter.QueryBGP(ctx, prefix)
	if err != nil {
		return nil, err
	}

	routes, err := parser.ParseRoutes(output)
	if err != nil {
		return nil, err
	}

	return &bgp.LookupResult{
		Prefix:    prefix,
		QueryLG:   lg.Name,
		Timestamp: time.Now(),
		Routes:    routes,
	}, nil
}

func queryJuniperLG(ctx context.Context, lg *config.LookingGlass, prefix string) (*bgp.LookupResult, error) {
	adapter := http.NewHTTPAdapter(lg.URL, 30*time.Second)
	parser := junos.NewParser()

	output, err := adapter.QueryBGP(ctx, prefix)
	if err != nil {
		return nil, err
	}

	routes, err := parser.ParseRoutes(output)
	if err != nil {
		return nil, err
	}

	return &bgp.LookupResult{
		Prefix:    prefix,
		QueryLG:   lg.Name,
		Timestamp: time.Now(),
		Routes:    routes,
	}, nil
}

func printResult(result *bgp.LookupResult, format string) error {
	switch format {
	case "json":
		return printJSON(result)
	case "yaml":
		return printYAML(result)
	default:
		return printTable(result)
	}
}

func printJSON(result *bgp.LookupResult) error {
	data, err := config.MarshalJSON(result)
	if err != nil {
		return err
	}
	fmt.Println(string(data))
	return nil
}

func printYAML(result *bgp.LookupResult) error {
	data, err := config.MarshalYAML(result)
	if err != nil {
		return err
	}
	fmt.Println(string(data))
	return nil
}

func printTable(result *bgp.LookupResult) error {
	fmt.Printf("Prefix: %s\n", result.Prefix)
	fmt.Printf("Looking Glass: %s\n", result.QueryLG)
	fmt.Printf("Timestamp: %s\n\n", result.Timestamp.Format(time.RFC3339))

	if len(result.Routes) == 0 {
		fmt.Println("No routes found")
		return nil
	}

	fmt.Println("Routes:")
	fmt.Println("--------")
	for _, route := range result.Routes {
		fmt.Printf("  Prefix: %s\n", route.Prefix)
		fmt.Printf("  Next Hop: %s\n", route.NextHop)
		fmt.Printf("  AS Path: %v\n", route.ASPath)
		fmt.Printf("  Local Pref: %d\n", route.LocalPref)
		fmt.Printf("  MED: %d\n", route.MED)
		fmt.Printf("  Best: %v\n", route.Best)
		fmt.Println()
	}

	return nil
}
