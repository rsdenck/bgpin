package main

import (
	"fmt"
	"strings"

	"github.com/bgpin/bgpin/internal/generators/config"
	"github.com/spf13/cobra"
)

var (
	configVendor   string
	configLocalAS  int
	configNeighbor string
	configASN      int
)

func newConfigCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "config",
		Short: "Generate router configuration",
		Long:  "Generate BGP configuration templates for various router vendors",
	}

	cmd.AddCommand(newConfigGenerateCommand())

	return cmd
}

func newConfigGenerateCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "generate bgp",
		Short: "Generate BGP configuration",
		Long:  "Generate BGP neighbor configuration for Cisco, Juniper, or MikroTik",
		RunE:  runConfigGenerate,
		Example: `  bgpin config generate bgp --vendor cisco --asn 65010 --neighbors 192.0.2.1,192.0.2.2 --remote-as 65020
  bgpin config generate bgp --vendor juniper --asn 65010 --neighbors 192.0.2.1 --remote-as 65020`,
	}

	cmd.Flags().StringVarP(&configVendor, "vendor", "", "cisco", "Vendor: cisco, juniper, mikrotik")
	cmd.Flags().IntVarP(&configLocalAS, "asn", "a", 0, "Local AS number")
	cmd.Flags().StringVarP(&configNeighbor, "neighbors", "n", "", "Neighbor IP addresses (comma-separated)")
	cmd.Flags().IntVarP(&configASN, "remote-as", "r", 0, "Remote AS number")
	cmd.Flags().StringVarP(&configOutputFormat, "output", "o", "text", "Output format: text, json")

	return cmd
}

func runConfigGenerate(cmd *cobra.Command, args []string) error {
	if configLocalAS == 0 {
		return fmt.Errorf("local AS number is required (--asn)")
	}

	if configNeighbor == "" {
		return fmt.Errorf("neighbor IP is required (--neighbors)")
	}

	if configASN == 0 {
		return fmt.Errorf("remote AS number is required (--remote-as)")
	}

	neighborIPs := strings.Split(configNeighbor, ",")

	neighbors := make([]config.NeighborConfig, 0, len(neighborIPs))
	for _, ip := range neighborIPs {
		ip = strings.TrimSpace(ip)
		if ip == "" {
			continue
		}
		neighbors = append(neighbors, config.NeighborConfig{
			NeighborIP:  ip,
			RemoteAS:    configASN,
			Description: fmt.Sprintf("Peer AS%d", configASN),
		})
	}

	gen := config.NewGenerator(configVendor, configLocalAS)
	bgpConfig := gen.GenerateBGPConfig(neighbors)

	switch configOutputFormat {
	case "json":
		fmt.Printf("{\n")
		fmt.Printf("  \"vendor\": \"%s\",\n", configVendor)
		fmt.Printf("  \"local_as\": %d,\n", configLocalAS)
		fmt.Printf("  \"neighbors\": %d,\n", len(neighbors))
		fmt.Printf("  \"config\": %q\n", bgpConfig)
		fmt.Printf("}\n")
	default:
		fmt.Println(bgpConfig)
	}

	return nil
}

var configOutputFormat string
