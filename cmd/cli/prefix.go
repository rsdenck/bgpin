package main

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"time"

	"github.com/bgpin/bgpin/sdk"
	"github.com/jedib0t/go-pretty/v6/table"
	"github.com/jedib0t/go-pretty/v6/text"
	"github.com/spf13/cobra"
	"gopkg.in/yaml.v3"
)

var (
	prefixOutputFormat string
	prefixTimeout      int
)

func newPrefixCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "prefix",
		Short: "Query prefix information",
		Long:  "Query IP prefix information from RIPE RIS",
	}

	cmd.PersistentFlags().StringVarP(&prefixOutputFormat, "output", "o", "table", "Output format: table, json, yaml")
	cmd.PersistentFlags().IntVarP(&prefixTimeout, "timeout", "t", 30, "Timeout in seconds")

	cmd.AddCommand(newPrefixOverviewCommand())

	return cmd
}

func newPrefixOverviewCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "overview [prefix]",
		Short: "Get prefix overview",
		Long:  "Get detailed overview of an IP prefix",
		Args:  cobra.ExactArgs(1),
		RunE:  runPrefixOverview,
		Example: `  bgpin prefix overview 200.160.0.0/20
  bgpin prefix overview 2804:4d44::/32 -o json
  bgpin prefix overview 186.250.184.0/24 --output yaml`,
	}
}

func runPrefixOverview(cmd *cobra.Command, args []string) error {
	prefix := args[0]

	client := sdk.NewDefaultClient()
	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(prefixTimeout)*time.Second)
	defer cancel()

	overview, err := client.GetPrefixOverview(ctx, prefix)
	if err != nil {
		return fmt.Errorf("failed to get prefix overview: %w", err)
	}

	return outputPrefixOverview(overview, prefixOutputFormat)
}

func outputPrefixOverview(overview *sdk.PrefixOverview, format string) error {
	switch format {
	case "json":
		data, err := json.MarshalIndent(overview, "", "  ")
		if err != nil {
			return err
		}
		fmt.Println(string(data))
	case "yaml":
		data, err := yaml.Marshal(overview)
		if err != nil {
			return err
		}
		fmt.Println(string(data))
	default:
		t := table.NewWriter()
		t.SetOutputMirror(os.Stdout)
		t.SetTitle(fmt.Sprintf("Prefix Overview: %s", overview.Prefix))
		t.Style().Title.Align = text.AlignCenter
		t.SetStyle(table.StyleRounded)
		t.Style().Options.SeparateRows = false

		if overview.ActualPrefix != "" && overview.ActualPrefix != overview.Prefix {
			t.AppendRow(table.Row{"Actual Prefix", overview.ActualPrefix})
		}
		t.AppendRow(table.Row{"Is Less Specific", overview.IsLessSpec})

		// Format ASNs
		asnsStr := "None"
		if len(overview.ASNs) > 0 {
			asnsStr = ""
			for i, asn := range overview.ASNs {
				if i > 0 {
					asnsStr += ", "
				}
				asnsStr += fmt.Sprintf("AS%d", asn)
			}
		}
		t.AppendRow(table.Row{"Announcing ASNs", asnsStr})
		t.AppendRow(table.Row{"Query Time", overview.QueryTime.Format(time.RFC3339)})

		t.Render()
	}
	return nil
}
