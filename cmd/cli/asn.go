package main

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/bgpin/bgpin/sdk"
	"github.com/jedib0t/go-pretty/v6/table"
	"github.com/jedib0t/go-pretty/v6/text"
	"github.com/spf13/cobra"
	"gopkg.in/yaml.v3"
)

var (
	asnOutputFormat string
	asnTimeout      int
)

func newASNCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "asn",
		Short: "Query ASN information",
		Long:  "Query Autonomous System Number information from RIPE RIS",
	}

	cmd.PersistentFlags().StringVarP(&asnOutputFormat, "output", "o", "table", "Output format: table, json, yaml")
	cmd.PersistentFlags().IntVarP(&asnTimeout, "timeout", "t", 30, "Timeout in seconds")

	cmd.AddCommand(newASNInfoCommand())
	cmd.AddCommand(newASNNeighborsCommand())
	cmd.AddCommand(newASNPrefixesCommand())
	cmd.AddCommand(newASNPeersCommand())

	return cmd
}

func newASNInfoCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "info [asn]",
		Short: "Get ASN information",
		Long:  "Get detailed information about an Autonomous System Number",
		Args:  cobra.ExactArgs(1),
		RunE:  runASNInfo,
		Example: `  bgpin asn info 262978
  bgpin asn info 262978 -o json
  bgpin asn info 262978 --output yaml`,
	}
}

func newASNNeighborsCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "neighbors [asn]",
		Short: "Get ASN neighbors",
		Long:  "Get BGP neighbors of an Autonomous System",
		Args:  cobra.ExactArgs(1),
		RunE:  runASNNeighbors,
		Example: `  bgpin asn neighbors 262978
  bgpin asn neighbors 262978 -o json`,
	}
}

func newASNPrefixesCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "prefixes [asn]",
		Short: "Get announced prefixes",
		Long:  "Get all prefixes announced by an Autonomous System",
		Args:  cobra.ExactArgs(1),
		RunE:  runASNPrefixes,
		Example: `  bgpin asn prefixes 262978
  bgpin asn prefixes 262978 -o json`,
	}
}

func newASNPeersCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "peers [asn]",
		Short: "Get RIS peers",
		Long:  "Get RIPE RIS peers for an Autonomous System",
		Args:  cobra.ExactArgs(1),
		RunE:  runASNPeers,
		Example: `  bgpin asn peers 262978
  bgpin asn peers 262978 -o json`,
	}
}

func runASNInfo(cmd *cobra.Command, args []string) error {
	asn, err := parseASN(args[0])
	if err != nil {
		return err
	}

	client := sdk.NewDefaultClient()
	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(asnTimeout)*time.Second)
	defer cancel()

	info, err := client.GetASNInfo(ctx, asn)
	if err != nil {
		return fmt.Errorf("failed to get ASN info: %w", err)
	}

	return outputASNInfo(info, asnOutputFormat)
}

func runASNNeighbors(cmd *cobra.Command, args []string) error {
	asn, err := parseASN(args[0])
	if err != nil {
		return err
	}

	client := sdk.NewDefaultClient()
	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(asnTimeout)*time.Second)
	defer cancel()

	neighbors, err := client.GetASNNeighbors(ctx, asn)
	if err != nil {
		return fmt.Errorf("failed to get ASN neighbors: %w", err)
	}

	return outputASNNeighbors(neighbors, asnOutputFormat)
}

func runASNPrefixes(cmd *cobra.Command, args []string) error {
	asn, err := parseASN(args[0])
	if err != nil {
		return err
	}

	client := sdk.NewDefaultClient()
	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(asnTimeout)*time.Second)
	defer cancel()

	prefixes, err := client.GetAnnouncedPrefixes(ctx, asn)
	if err != nil {
		return fmt.Errorf("failed to get announced prefixes: %w", err)
	}

	return outputASNPrefixes(prefixes, asnOutputFormat)
}

func runASNPeers(cmd *cobra.Command, args []string) error {
	asn, err := parseASN(args[0])
	if err != nil {
		return err
	}

	client := sdk.NewDefaultClient()
	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(asnTimeout)*time.Second)
	defer cancel()

	peers, err := client.GetRISPeers(ctx, asn)
	if err != nil {
		return fmt.Errorf("failed to get RIS peers: %w", err)
	}

	return outputASNPeers(peers, asnOutputFormat)
}

func parseASN(s string) (int, error) {
	// Remove "AS" prefix if present
	if len(s) > 2 && (s[:2] == "AS" || s[:2] == "as") {
		s = s[2:]
	}

	asn, err := strconv.Atoi(s)
	if err != nil {
		return 0, fmt.Errorf("invalid ASN: %s", s)
	}

	if asn <= 0 {
		return 0, fmt.Errorf("ASN must be positive: %d", asn)
	}

	return asn, nil
}

func outputASNInfo(info *sdk.ASNInfo, format string) error {
	switch format {
	case "json":
		data, err := json.MarshalIndent(info, "", "  ")
		if err != nil {
			return err
		}
		fmt.Println(string(data))
	case "yaml":
		data, err := yaml.Marshal(info)
		if err != nil {
			return err
		}
		fmt.Println(string(data))
	default:
		t := table.NewWriter()
		t.SetOutputMirror(os.Stdout)
		t.SetTitle(fmt.Sprintf("ASN Information: AS%d", info.ASN))
		t.Style().Title.Align = text.AlignCenter
		t.SetStyle(table.StyleRounded)
		t.Style().Options.SeparateRows = false

		t.AppendRow(table.Row{"Holder", info.Holder})
		t.AppendRow(table.Row{"Announced", info.Announced})
		t.AppendRow(table.Row{"Block", info.Block})
		if info.Description != "" {
			t.AppendRow(table.Row{"Description", info.Description})
		}
		if info.Country != "" {
			t.AppendRow(table.Row{"Country", info.Country})
		}

		t.Render()
	}
	return nil
}

func outputASNNeighbors(neighbors *sdk.ASNNeighbors, format string) error {
	switch format {
	case "json":
		data, err := json.MarshalIndent(neighbors, "", "  ")
		if err != nil {
			return err
		}
		fmt.Println(string(data))
	case "yaml":
		data, err := yaml.Marshal(neighbors)
		if err != nil {
			return err
		}
		fmt.Println(string(data))
	default:
		t := table.NewWriter()
		t.SetOutputMirror(os.Stdout)
		t.SetTitle(fmt.Sprintf("BGP Neighbors for AS%d (Total: %d)", neighbors.ASN, len(neighbors.Neighbors)))
		t.Style().Title.Align = text.AlignCenter
		t.SetStyle(table.StyleRounded)
		t.Style().Options.SeparateRows = false

		t.AppendHeader(table.Row{"ASN", "Type", "Power"})

		displayCount := 30
		for i, neighbor := range neighbors.Neighbors {
			if i >= displayCount {
				t.AppendFooter(table.Row{"...", fmt.Sprintf("and %d more neighbors", len(neighbors.Neighbors)-displayCount), ""})
				break
			}
			t.AppendRow(table.Row{
				fmt.Sprintf("AS%d", neighbor.ASN),
				neighbor.Type,
				neighbor.Power,
			})
		}

		t.Render()
	}
	return nil
}

func outputASNPrefixes(prefixes *sdk.AnnouncedPrefixes, format string) error {
	switch format {
	case "json":
		data, err := json.MarshalIndent(prefixes, "", "  ")
		if err != nil {
			return err
		}
		fmt.Println(string(data))
	case "yaml":
		data, err := yaml.Marshal(prefixes)
		if err != nil {
			return err
		}
		fmt.Println(string(data))
	default:
		t := table.NewWriter()
		t.SetOutputMirror(os.Stdout)
		t.SetTitle(fmt.Sprintf("Announced Prefixes for AS%d (Total: %d)", prefixes.ASN, len(prefixes.Prefixes)))
		t.Style().Title.Align = text.AlignCenter
		t.SetStyle(table.StyleRounded)
		t.Style().Options.SeparateRows = false

		t.AppendHeader(table.Row{"#", "Prefix", "Type"})

		displayCount := 30
		for i, prefix := range prefixes.Prefixes {
			if i >= displayCount {
				t.AppendFooter(table.Row{"...", fmt.Sprintf("and %d more prefixes", len(prefixes.Prefixes)-displayCount), ""})
				break
			}

			// Detect IPv4 vs IPv6
			prefixType := "IPv4"
			if strings.Contains(prefix.Prefix, ":") {
				prefixType = "IPv6"
			}

			t.AppendRow(table.Row{
				i + 1,
				prefix.Prefix,
				prefixType,
			})
		}

		t.Render()
	}
	return nil
}

func outputASNPeers(peers []string, format string) error {
	switch format {
	case "json":
		data, err := json.MarshalIndent(peers, "", "  ")
		if err != nil {
			return err
		}
		fmt.Println(string(data))
	case "yaml":
		data, err := yaml.Marshal(peers)
		if err != nil {
			return err
		}
		fmt.Println(string(data))
	default:
		t := table.NewWriter()
		t.SetOutputMirror(os.Stdout)
		t.SetTitle(fmt.Sprintf("RIS Peers (Total: %d)", len(peers)))
		t.Style().Title.Align = text.AlignCenter
		t.SetStyle(table.StyleRounded)
		t.Style().Options.SeparateRows = false

		t.AppendHeader(table.Row{"#", "Peer Information"})

		displayCount := 30
		for i, peer := range peers {
			if i >= displayCount {
				t.AppendFooter(table.Row{"...", fmt.Sprintf("and %d more peers", len(peers)-displayCount)})
				break
			}

			t.AppendRow(table.Row{
				i + 1,
				peer,
			})
		}

		t.Render()
	}
	return nil
}
