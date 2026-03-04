package main

import (
	"context"
	"fmt"
	"os"
	"time"

	"github.com/bgpin/bgpin/internal/core/bgp"
	"github.com/bgpin/bgpin/internal/parsers/http"
	"github.com/jedib0t/go-pretty/v6/table"
	"github.com/jedib0t/go-pretty/v6/text"
	"github.com/spf13/cobra"
)

var analyzeOutputFormat string

func newAnalyzeCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "analyze",
		Short: "Analyze BGP routes for anomalies",
		Long:  "Analyze BGP routes for anomalies and security issues",
	}

	cmd.PersistentFlags().StringVarP(&analyzeOutputFormat, "output", "o", "table", "Output format: table, json, yaml")

	cmd.AddCommand(newAnalyzeRouteCommand())
	cmd.AddCommand(newAnalyzeASNCommand())

	return cmd
}

func newAnalyzeRouteCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "route [prefix]",
		Short: "Analyze routes for a prefix",
		Long:  "Analyze BGP routes for a prefix, showing detailed attributes and anomalies",
		Args:  cobra.ExactArgs(1),
		RunE:  runAnalyzeRoute,
		Example: `  bgpin analyze route 8.8.8.0/24
  bgpin analyze route 1.1.1.0/24 -o json`,
	}
}

func runAnalyzeRoute(cmd *cobra.Command, args []string) error {
	prefix := args[0]
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	parser := http.NewRIPEParser()
	result, err := parser.QueryBGP(ctx, prefix)
	if err != nil {
		return fmt.Errorf("lookup failed: %w", err)
	}

	result.QueryLG = "RIPE RIS"

	var allAnomalies []bgp.Anomaly

	for _, route := range result.Routes {
		anomalies := route.DetectAnomalies()
		allAnomalies = append(allAnomalies, anomalies...)
	}

	return outputAnalyzeRoute(prefix, result.Routes, allAnomalies, analyzeOutputFormat)
}

func outputAnalyzeRoute(prefix string, routes []bgp.Route, anomalies []bgp.Anomaly, format string) error {
	switch format {
	case "json":
		fmt.Printf("{\n")
		fmt.Printf("  \"prefix\": \"%s\",\n", prefix)
		fmt.Printf("  \"route_count\": %d,\n", len(routes))
		fmt.Printf("  \"anomaly_count\": %d,\n", len(anomalies))
		fmt.Printf("  \"routes\": [\n")
		for i, r := range routes {
			fmt.Printf("    {\n")
			fmt.Printf("      \"prefix\": \"%s\",\n", r.Prefix)
			fmt.Printf("      \"next_hop\": \"%s\",\n", r.NextHop)
			fmt.Printf("      \"as_path\": %v,\n", r.ASPath)
			fmt.Printf("      \"local_pref\": %d,\n", r.LocalPref)
			fmt.Printf("      \"med\": %d,\n", r.MED)
			fmt.Printf("      \"origin\": \"%s\",\n", r.Origin)
			fmt.Printf("      \"communities\": %v,\n", r.Community)
			fmt.Printf("      \"best\": %v\n", r.Best)
			fmt.Printf("    }")
			if i < len(routes)-1 {
				fmt.Printf(",")
			}
			fmt.Printf("\n")
		}
		fmt.Printf("  ]\n")
		fmt.Printf("}\n")
	default:
		t := table.NewWriter()
		t.SetOutputMirror(os.Stdout)
		t.SetTitle(fmt.Sprintf("BGP Route Analysis: %s", prefix))
		t.Style().Title.Align = text.AlignCenter
		t.SetStyle(table.StyleRounded)
		t.Style().Options.SeparateRows = false

		t.AppendHeader(table.Row{"#", "AS Path", "Next Hop", "Local Pref", "MED", "Origin", "Communities", "Best"})

		for i, route := range routes {
			asPathStr := formatASPath(route.ASPath)
			communities := ""
			if len(route.Community) > 0 {
				if len(route.Community) > 2 {
					communities = fmt.Sprintf("%s (+%d)", route.Community[0], len(route.Community)-1)
				} else {
					communities = route.Community[0]
				}
			}

			best := ""
			if route.Best {
				best = "Yes"
			}

			t.AppendRow(table.Row{
				i + 1,
				asPathStr,
				route.NextHop,
				route.LocalPref,
				route.MED,
				route.Origin,
				communities,
				best,
			})
		}

		t.Render()

		if len(anomalies) > 0 {
			fmt.Println()
			t2 := table.NewWriter()
			t2.SetOutputMirror(os.Stdout)
			t2.SetTitle(fmt.Sprintf("Detected Anomalies (%d)", len(anomalies)))
			t2.Style().Title.Align = text.AlignCenter
			t2.SetStyle(table.StyleRounded)
			t2.Style().Options.SeparateRows = false

			t2.AppendHeader(table.Row{"Type", "Severity", "Message"})

			for _, a := range anomalies {
				t2.AppendRow(table.Row{
					a.Type,
					a.Severity,
					a.Message,
				})
			}

			t2.Render()
		}
	}

	return nil
}

func newAnalyzeASNCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "asn [asn]",
		Short: "Analyze routes from an AS",
		Long:  "Analyze all routes originated by an Autonomous System",
		Args:  cobra.ExactArgs(1),
		RunE:  runAnalyzeASN,
		Example: `  bgpin analyze asn 13335
  bgpin analyze asn 15169 -o json`,
	}
}

func runAnalyzeASN(cmd *cobra.Command, args []string) error {
	asn, err := parseASN(args[0])
	if err != nil {
		return err
	}

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	parser := http.NewRIPEParser()
	result, err := parser.QueryBGP(ctx, fmt.Sprintf("AS%d", asn))
	if err != nil {
		return fmt.Errorf("lookup failed: %w", err)
	}

	var allAnomalies []bgp.Anomaly

	for _, route := range result.Routes {
		anomalies := route.DetectAnomalies()
		allAnomalies = append(allAnomalies, anomalies...)
	}

	return outputAnalyzeRoute(fmt.Sprintf("AS%d", asn), result.Routes, allAnomalies, analyzeOutputFormat)
}
