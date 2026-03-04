package main

import (
	"context"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/bgpin/bgpin/internal/core/bgp"
	"github.com/bgpin/bgpin/internal/parsers/http"
	"github.com/bgpin/bgpin/pkg/config"
	"github.com/jedib0t/go-pretty/v6/table"
	"github.com/jedib0t/go-pretty/v6/text"
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

	cmd.Flags().StringVarP(&lgName, "lg", "l", "", "Looking glass name (not used, uses RIPE RIS)")
	cmd.Flags().StringVarP(&outputFormat, "output", "o", "table", "Output format: table, json, yaml")

	return cmd
}

func runLookup(cmd *cobra.Command, args []string) error {
	prefix = args[0]
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	parser := http.NewRIPEParser()

	result, err := parser.QueryBGP(ctx, prefix)
	if err != nil {
		return fmt.Errorf("lookup failed: %w", err)
	}

	result.QueryLG = "RIPE RIS"

	return printResult(result, outputFormat)
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
	t := table.NewWriter()
	t.SetOutputMirror(os.Stdout)
	t.SetTitle(fmt.Sprintf("BGP Lookup: %s", result.Prefix))
	t.Style().Title.Align = text.AlignCenter
	t.SetStyle(table.StyleRounded)
	t.Style().Options.SeparateRows = false

	t.AppendHeader(table.Row{"#", "AS Path", "Origin", "Communities", "Best"})

	for i, route := range result.Routes {

		asPathStr := formatASPath(route.ASPath)
		communities := ""
		if len(route.Community) > 0 {
			if len(route.Community) > 3 {
				communities = fmt.Sprintf("%s (+%d)", strings.Join(route.Community[:3], ", "), len(route.Community)-3)
			} else {
				communities = strings.Join(route.Community, ", ")
			}
		}

		best := ""
		if route.Best {
			best = "Yes"
		}

		t.AppendRow(table.Row{
			i + 1,
			asPathStr,
			route.Origin,
			communities,
			best,
		})
	}

	t.Render()

	fmt.Printf("\nSource: %s | Query Time: %s\n", result.QueryLG, result.Timestamp.Format("2006-01-02 15:04:05"))

	return nil
}

func formatASPath(asPath []int) string {
	if len(asPath) == 0 {
		return "N/A"
	}

	parts := make([]string, len(asPath))
	for i, as := range asPath {
		parts[i] = fmt.Sprintf("AS%d", as)
	}
	return strings.Join(parts, " ")
}
