package main

import (
	"context"
	"fmt"
	"os"

	"github.com/bgpin/bgpin/internal/flow"
	"github.com/jedib0t/go-pretty/v6/table"
	"github.com/jedib0t/go-pretty/v6/text"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var (
	flowOutputFormat string
	flowLimit        int
	flowCollector    *flow.GoFlowCollector
)

func newFlowCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "flow",
		Short: "Flow telemetry and traffic analysis",
		Long:  "Analyze network flow data (NetFlow/sFlow/IPFIX) and correlate with BGP",
		PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
			return initFlowCollector()
		},
	}

	cmd.PersistentFlags().StringVarP(&flowOutputFormat, "output", "o", "table", "Output format: table, json, yaml")
	cmd.PersistentFlags().IntVarP(&flowLimit, "limit", "l", 10, "Limit number of results")

	cmd.AddCommand(newFlowTopCommand())
	cmd.AddCommand(newFlowASNCommand())
	cmd.AddCommand(newFlowAnomalyCommand())
	cmd.AddCommand(newFlowUpstreamCommand())
	cmd.AddCommand(newFlowStatsCommand())

	return cmd
}

func initFlowCollector() error {
	// Check if flow collection is enabled
	if !viper.GetBool("flow.enabled") {
		return nil
	}

	// Create collector config from viper
	config := flow.GoFlowConfig{
		NetFlowEnabled: viper.GetBool("flow.netflow.enabled"),
		NetFlowAddr:    viper.GetString("flow.netflow.addr"),
		NetFlowPort:    viper.GetInt("flow.netflow.port"),
		SFlowEnabled:   viper.GetBool("flow.sflow.enabled"),
		SFlowAddr:      viper.GetString("flow.sflow.addr"),
		SFlowPort:      viper.GetInt("flow.sflow.port"),
		IPFIXEnabled:   viper.GetBool("flow.ipfix.enabled"),
		IPFIXAddr:      viper.GetString("flow.ipfix.addr"),
		IPFIXPort:      viper.GetInt("flow.ipfix.port"),
		Workers:        viper.GetInt("flow.workers"),
		BufferSize:     viper.GetInt("flow.buffer_size"),
		EnableBGPCorr:  viper.GetBool("flow.bgp_correlation"),
	}

	var err error
	flowCollector, err = flow.NewGoFlowCollector(config)
	if err != nil {
		return fmt.Errorf("failed to create flow collector: %w", err)
	}

	// Start collector in background
	ctx := context.Background()
	if err := flowCollector.Start(ctx); err != nil {
		return fmt.Errorf("failed to start flow collector: %w", err)
	}

	return nil
}

func newFlowTopCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "top",
		Short: "Show top prefixes by traffic",
		Long:  "Display top prefixes sorted by traffic volume",
		RunE:  runFlowTop,
		Example: `  bgpin flow top
  bgpin flow top --limit 20
  bgpin flow top --prefix 8.8.8.0/24`,
	}
}

func newFlowASNCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "asn [asn]",
		Short: "Show traffic statistics for an ASN",
		Long:  "Display detailed traffic statistics for a specific ASN",
		Args:  cobra.ExactArgs(1),
		RunE:  runFlowASN,
		Example: `  bgpin flow asn 15169
  bgpin flow asn AS262978`,
	}
}

func newFlowAnomalyCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "anomaly",
		Short: "Detect traffic anomalies",
		Long:  "Detect and display traffic anomalies (DDoS, spikes, unusual patterns)",
		RunE:  runFlowAnomaly,
		Example: `  bgpin flow anomaly
  bgpin flow anomaly --severity high`,
	}
}

func newFlowUpstreamCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "upstream-compare",
		Short: "Compare traffic across upstreams",
		Long:  "Compare traffic patterns and performance across multiple upstream providers",
		RunE:  runFlowUpstream,
		Example: `  bgpin flow upstream-compare
  bgpin flow upstream-compare --prefix 8.8.8.0/24`,
	}
}

func runFlowTop(cmd *cobra.Command, args []string) error {
	t := table.NewWriter()
	t.SetOutputMirror(os.Stdout)
	t.SetTitle("Top Prefixes by Traffic")
	t.Style().Title.Align = text.AlignCenter
	t.SetStyle(table.StyleRounded)
	t.Style().Options.SeparateRows = false

	t.AppendHeader(table.Row{"#", "Prefix", "ASN", "Traffic", "PPS", "Top Protocol"})

	if flowCollector != nil {
		// Get real data from collector
		aggregator := flowCollector.GetAggregator()
		topFlows := aggregator.GetTopFlows(flowLimit)

		if len(topFlows) > 0 {
			for i, f := range topFlows {
				prefix := fmt.Sprintf("%s/24", f.DstAddr.String())
				asn := fmt.Sprintf("AS%d", f.DstAS)
				traffic := formatBytes(f.Bytes)
				pps := formatPackets(f.Packets)
				proto := formatProtocol(f.Protocol, f.DstPort)

				t.AppendRow(table.Row{i + 1, prefix, asn, traffic, pps, proto})
			}

			t.Render()
			return nil
		}
	}

	// Show example data if no collector or no data
	examples := []table.Row{
		{1, "8.8.8.0/24", "AS15169", "850 Mbps", "120k", "TCP (443)"},
		{2, "1.1.1.0/24", "AS13335", "640 Mbps", "98k", "UDP (53)"},
		{3, "208.67.222.0/24", "AS36692", "1.2 Gbps", "150k", "TCP (80)"},
	}

	for _, row := range examples {
		t.AppendRow(row)
	}

	t.AppendFooter(table.Row{"", "", "", "Total: 2.69 Gbps", "368k", ""})
	t.Render()

	fmt.Println("\nâš ï¸  Note: Flow collection not enabled or no data available.")
	fmt.Println("To enable: configure flow settings in bgpin.yaml and restart exporters.")

	return nil
}

func runFlowASN(cmd *cobra.Command, args []string) error {
	asn := args[0]
	asnNum := parseASNNumber(asn)

	t := table.NewWriter()
	t.SetOutputMirror(os.Stdout)
	t.SetTitle(fmt.Sprintf("Traffic Statistics for %s", asn))
	t.Style().Title.Align = text.AlignCenter
	t.SetStyle(table.StyleRounded)
	t.Style().Options.SeparateRows = false

	t.AppendHeader(table.Row{"Metric", "Inbound", "Outbound"})

	if flowCollector != nil {
		aggregator := flowCollector.GetAggregator()
		stats := aggregator.GetASNStats(asnNum)

		if stats != nil {
			t.AppendRow(table.Row{"Traffic", formatBytes(stats.InboundBytes), formatBytes(stats.OutboundBytes)})
			t.AppendRow(table.Row{"Packets", formatPackets(stats.InboundPackets), formatPackets(stats.OutboundPackets)})
			t.AppendRow(table.Row{"Flows", fmt.Sprintf("%d", stats.InboundFlows), fmt.Sprintf("%d", stats.OutboundFlows)})
			t.Render()
			return nil
		}
	}

	// Example data
	t.AppendRow(table.Row{"Traffic", "850 Mbps", "640 Mbps"})
	t.AppendRow(table.Row{"Packets", "120k pps", "98k pps"})
	t.AppendRow(table.Row{"Flows", "15,234", "12,891"})
	t.Render()

	fmt.Println("\nâš ï¸  Note: Flow collection not enabled or no data for this ASN.")

	return nil
}

func runFlowAnomaly(cmd *cobra.Command, args []string) error {
	t := table.NewWriter()
	t.SetOutputMirror(os.Stdout)
	t.SetTitle("Detected Traffic Anomalies")
	t.Style().Title.Align = text.AlignCenter
	t.SetStyle(table.StyleRounded)
	t.Style().Options.SeparateRows = false

	t.AppendHeader(table.Row{"Time", "Type", "Severity", "Prefix", "ASN", "Description"})

	if flowCollector != nil {
		aggregator := flowCollector.GetAggregator()
		anomalies := aggregator.GetAnomalies()

		if len(anomalies) > 0 {
			for _, a := range anomalies {
				timeStr := a.DetectedAt.Format("15:04:05")
				prefix := fmt.Sprintf("%s/24", a.DstAddr.String())
				asn := fmt.Sprintf("AS%d", a.DstAS)

				t.AppendRow(table.Row{timeStr, a.Type, a.Severity, prefix, asn, a.Description})
			}
			t.Render()
			return nil
		}
	}

	// Example data
	examples := []table.Row{
		{"11:45:23", "DDoS", "CRITICAL", "8.8.8.0/24", "AS15169", "High PPS detected (250k)"},
		{"11:42:15", "Spike", "HIGH", "1.1.1.0/24", "AS13335", "Traffic spike (2.5 Gbps)"},
		{"11:38:07", "Drop", "MEDIUM", "208.67.222.0/24", "AS36692", "Traffic drop (80%)"},
	}

	for _, row := range examples {
		t.AppendRow(row)
	}
	t.Render()

	fmt.Println("\nâš ï¸  Note: Anomaly detection not enabled or no anomalies detected.")

	return nil
}

func runFlowUpstream(cmd *cobra.Command, args []string) error {
	t := table.NewWriter()
	t.SetOutputMirror(os.Stdout)
	t.SetTitle("Upstream Provider Comparison")
	t.Style().Title.Align = text.AlignCenter
	t.SetStyle(table.StyleRounded)
	t.Style().Options.SeparateRows = false

	t.AppendHeader(table.Row{"Provider", "ASN", "AS Path", "Traffic", "PPS", "Latency", "Loss"})

	// Example data
	examples := []table.Row{
		{"Telia", "AS1299", "1299 15169", "850 Mbps", "120k", "42ms", "0.1%"},
		{"Level3", "AS3356", "3356 15169", "640 Mbps", "98k", "38ms", "0.2%"},
		{"GTT", "AS3257", "3257 15169", "1.2 Gbps", "150k", "55ms", "0.3%"},
	}

	for _, row := range examples {
		t.AppendRow(row)
	}

	t.AppendFooter(table.Row{"", "", "Total", "2.69 Gbps", "368k", "Avg: 45ms", "Avg: 0.2%"})

	t.Render()

	fmt.Println("\nâš ï¸  Note: Upstream comparison not yet implemented. This is example output.")
	fmt.Println("This feature will correlate BGP data with actual traffic flows.")

	return nil
}

func newFlowStatsCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "stats",
		Short: "Show flow collector statistics",
		Long:  "Display statistics about the flow collector (packets, flows, errors)",
		RunE:  runFlowStats,
		Example: `  bgpin flow stats`,
	}
}

func runFlowStats(cmd *cobra.Command, args []string) error {
	if flowCollector == nil {
		fmt.Println("âš ï¸  Flow collector not enabled.")
		fmt.Println("To enable: set flow.enabled=true in bgpin.yaml")
		return nil
	}

	stats := flowCollector.GetStats()

	t := table.NewWriter()
	t.SetOutputMirror(os.Stdout)
	t.SetTitle("Flow Collector Statistics")
	t.Style().Title.Align = text.AlignCenter
	t.SetStyle(table.StyleRounded)
	t.Style().Options.SeparateRows = false

	t.AppendHeader(table.Row{"Metric", "Value"})
	t.AppendRow(table.Row{"NetFlow Packets", fmt.Sprintf("%d", stats.NetFlowPackets)})
	t.AppendRow(table.Row{"sFlow Packets", fmt.Sprintf("%d", stats.SFlowPackets)})
	t.AppendRow(table.Row{"IPFIX Packets", fmt.Sprintf("%d", stats.IPFIXPackets)})
	t.AppendRow(table.Row{"Total Flows", fmt.Sprintf("%d", stats.TotalFlows)})
	t.AppendRow(table.Row{"Dropped Flows", fmt.Sprintf("%d", stats.DroppedFlows)})
	t.AppendRow(table.Row{"Processing Errors", fmt.Sprintf("%d", stats.ProcessingErrors)})
	t.AppendRow(table.Row{"Last Update", stats.LastUpdate.Format("2006-01-02 15:04:05")})

	t.Render()

	return nil
}

// Helper functions
func parseASNNumber(asn string) uint32 {
	var asnNum uint32
	fmt.Sscanf(asn, "AS%d", &asnNum)
	if asnNum == 0 {
		fmt.Sscanf(asn, "%d", &asnNum)
	}
	return asnNum
}

func formatBytes(bytes uint64) string {
	const unit = 1000
	if bytes < unit {
		return fmt.Sprintf("%d B", bytes)
	}
	div, exp := uint64(unit), 0
	for n := bytes / unit; n >= unit; n /= unit {
		div *= unit
		exp++
	}
	return fmt.Sprintf("%.1f %cB", float64(bytes)/float64(div), "kMGTPE"[exp])
}

func formatPackets(packets uint64) string {
	if packets < 1000 {
		return fmt.Sprintf("%d", packets)
	}
	if packets < 1000000 {
		return fmt.Sprintf("%.1fk", float64(packets)/1000)
	}
	return fmt.Sprintf("%.1fM", float64(packets)/1000000)
}

func formatProtocol(proto uint8, port uint16) string {
	protoName := "Unknown"
	switch proto {
	case 6:
		protoName = "TCP"
	case 17:
		protoName = "UDP"
	case 1:
		protoName = "ICMP"
	}

	portName := ""
	switch port {
	case 80:
		portName = "HTTP"
	case 443:
		portName = "HTTPS"
	case 53:
		portName = "DNS"
	case 22:
		portName = "SSH"
	case 25:
		portName = "SMTP"
	default:
		portName = fmt.Sprintf("%d", port)
	}

	return fmt.Sprintf("%s (%s)", protoName, portName)
}
