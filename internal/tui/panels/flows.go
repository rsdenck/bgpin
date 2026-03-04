package panels

import (
	"fmt"
	"sort"
	"strings"
	"time"

	"github.com/bgpin/bgpin/internal/tui/gobgp"
	"github.com/bgpin/bgpin/internal/tui/telemetry"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// FlowEntry represents a network flow
type FlowEntry struct {
	SrcIP       string
	DstIP       string
	SrcPort     uint16
	DstPort     uint16
	Protocol    string
	Bytes       uint64
	Packets     uint64
	Mbps        float64
	SrcASN      uint32
	DstASN      uint32
	FirstSeen   time.Time
	LastSeen    time.Time
	Flags       []string
	GeoSrc      string
	GeoDst      string
}

// FlowsModel represents the Top 5 Flows panel
type FlowsModel struct {
	width       int
	height      int
	flows       []*FlowEntry
	topFlows    []*FlowEntry
	selected    int
	bgpClient   *gobgp.BGPClient
	telemetry   *telemetry.TelemetryManager
	sortBy      string // "bytes", "packets", "mbps"
	filterASN   uint32
	searchQuery string
	autoRefresh bool
}

// NewFlowsModel creates a new flows model
func NewFlowsModel(bgpClient *gobgp.BGPClient) FlowsModel {
	tm := telemetry.NewTelemetryManager(80)
	
	// Initialize sparklines for flow telemetry
	tm.AddSparkline("total_flows", "Flows", "/s", 60)
	tm.AddSparkline("total_bandwidth", "Bandwidth", "Gbps", 60)
	tm.AddSparkline("top_talkers", "Top Talkers", "", 60)
	tm.AddSparkline("protocols", "Protocols", "", 60)
	
	return FlowsModel{
		bgpClient:   bgpClient,
		telemetry:   tm,
		flows:       make([]*FlowEntry, 0),
		topFlows:    make([]*FlowEntry, 0),
		sortBy:      "mbps",
		autoRefresh: true,
	}
}

// Init initializes the flows model
func (m FlowsModel) Init() tea.Cmd {
	return tea.Batch(
		m.fetchFlowsData(),
		tea.Tick(5*time.Second, func(t time.Time) tea.Msg {
			return FlowTickMsg(t)
		}),
	)
}

// Update handles messages for the flows panel
func (m FlowsModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "up", "k":
			if m.selected > 0 {
				m.selected--
			}
		case "down", "j":
			if m.selected < len(m.topFlows)-1 {
				m.selected++
			}
		case "1":
			m.sortBy = "mbps"
			m.updateTopFlows()
		case "2":
			m.sortBy = "bytes"
			m.updateTopFlows()
		case "3":
			m.sortBy = "packets"
			m.updateTopFlows()
		case "r":
			return m, m.fetchFlowsData()
		case "a":
			m.autoRefresh = !m.autoRefresh
		case "f":
			// TODO: Implement filter dialog
		}
		
	case FlowsDataMsg:
		m.flows = msg.Flows
		m.updateTopFlows()
		m.updateTelemetry()
		
	case FlowTickMsg:
		var cmd tea.Cmd
		if m.autoRefresh {
			cmd = m.fetchFlowsData()
		}
		return m, tea.Batch(
			cmd,
			tea.Tick(5*time.Second, func(t time.Time) tea.Msg {
				return FlowTickMsg(t)
			}),
		)
	}
	
	return m, nil
}

// View renders the flows panel
func (m FlowsModel) View() string {
	if len(m.flows) == 0 {
		return m.renderLoading()
	}
	
	return m.renderFlows()
}

// SetSize sets the panel size
func (m *FlowsModel) SetSize(width, height int) {
	m.width = width
	m.height = height
	m.telemetry = telemetry.NewTelemetryManager(width)
}

// UpdateData updates the panel data
func (m *FlowsModel) UpdateData(data interface{}) {
	if flows, ok := data.([]*FlowEntry); ok {
		m.flows = flows
		m.updateTopFlows()
		m.updateTelemetry()
	}
}

// renderLoading renders loading state
func (m FlowsModel) renderLoading() string {
	style := lipgloss.NewStyle().
		Width(m.width).
		Height(m.height).
		Align(lipgloss.Center, lipgloss.Center).
		Foreground(lipgloss.Color("#666666"))
	
	return style.Render("Loading NetFlow/sFlow/IPFIX data...\nConnecting to flow collector...")
}

// renderFlows renders the flows content
func (m FlowsModel) renderFlows() string {
	// Header
	headerStyle := lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color("#FAFAFA")).
		Background(lipgloss.Color("#7D56F4")).
		Padding(0, 1).
		Width(m.width - 2).
		Align(lipgloss.Center)
	
	sortIndicator := ""
	switch m.sortBy {
	case "mbps":
		sortIndicator = " (by Bandwidth)"
	case "bytes":
		sortIndicator = " (by Bytes)"
	case "packets":
		sortIndicator = " (by Packets)"
	}
	
	autoRefreshIndicator := ""
	if m.autoRefresh {
		autoRefreshIndicator = " 🔄"
	}
	
	title := fmt.Sprintf("Top 5 Network Flows%s%s", sortIndicator, autoRefreshIndicator)
	header := headerStyle.Render(title)
	
	// Main content split: flows table (left) and details (right)
	leftWidth := m.width * 3 / 4
	rightWidth := m.width - leftWidth - 2
	contentHeight := m.height - 6 // Reserve space for header, telemetry, footer
	
	flowsTable := m.renderFlowsTable(leftWidth, contentHeight)
	flowDetails := m.renderFlowDetails(rightWidth, contentHeight)
	
	mainContent := lipgloss.JoinHorizontal(
		lipgloss.Top,
		flowsTable,
		lipgloss.NewStyle().Width(1).Render("│"),
		flowDetails,
	)
	
	// Telemetry section
	telemetrySection := m.renderTelemetry()
	
	// Footer
	footer := m.renderFooter()
	
	// Combine all sections
	content := lipgloss.JoinVertical(
		lipgloss.Left,
		header,
		mainContent,
		telemetrySection,
		footer,
	)
	
	// Container
	containerStyle := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("#7D56F4")).
		Width(m.width).
		Height(m.height)
	
	return containerStyle.Render(content)
}

// renderFlowsTable renders the top flows table
func (m FlowsModel) renderFlowsTable(width, height int) string {
	// Table styles
	tableHeaderStyle := lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color("#FAFAFA")).
		Background(lipgloss.Color("#444444")).
		Padding(0, 1)
	
	selectedRowStyle := lipgloss.NewStyle().
		Background(lipgloss.Color("#7D56F4")).
		Foreground(lipgloss.Color("#FAFAFA")).
		Padding(0, 1)
	
	normalRowStyle := lipgloss.NewStyle().
		Padding(0, 1)
	
	// Protocol styles
	tcpStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("#00FF00")).Bold(true)
	udpStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("#00FFFF")).Bold(true)
	icmpStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("#FFFF00")).Bold(true)
	
	// Table header
	tableHeader := lipgloss.JoinHorizontal(lipgloss.Top,
		tableHeaderStyle.Width(3).Render("#"),
		tableHeaderStyle.Width(16).Render("SOURCE"),
		tableHeaderStyle.Width(16).Render("DESTINATION"),
		tableHeaderStyle.Width(8).Render("PROTO"),
		tableHeaderStyle.Width(10).Render("BANDWIDTH"),
		tableHeaderStyle.Width(10).Render("BYTES"),
		tableHeaderStyle.Width(8).Render("PACKETS"),
		tableHeaderStyle.Width(8).Render("SRC ASN"),
		tableHeaderStyle.Width(8).Render("DST ASN"),
		tableHeaderStyle.Width(width-89).Render("GEO"),
	)
	
	// Table rows
	var rows []string
	
	for i, flow := range m.topFlows {
		if i >= 5 { // Only show top 5
			break
		}
		
		// Format source and destination
		src := fmt.Sprintf("%s:%d", flow.SrcIP, flow.SrcPort)
		dst := fmt.Sprintf("%s:%d", flow.DstIP, flow.DstPort)
		
		// Protocol styling
		var protoRendered string
		switch flow.Protocol {
		case "TCP":
			protoRendered = tcpStyle.Render(flow.Protocol)
		case "UDP":
			protoRendered = udpStyle.Render(flow.Protocol)
		case "ICMP":
			protoRendered = icmpStyle.Render(flow.Protocol)
		default:
			protoRendered = flow.Protocol
		}
		
		// Format bandwidth
		bandwidth := m.formatBandwidth(flow.Mbps)
		
		// Format bytes
		bytes := m.formatBytes(flow.Bytes)
		
		// Geography
		geo := fmt.Sprintf("%s→%s", flow.GeoSrc, flow.GeoDst)
		
		var rowStyle lipgloss.Style
		if i == m.selected {
			rowStyle = selectedRowStyle
		} else {
			rowStyle = normalRowStyle
		}
		
		row := lipgloss.JoinHorizontal(lipgloss.Top,
			rowStyle.Width(3).Render(fmt.Sprintf("%d", i+1)),
			rowStyle.Width(16).Render(m.truncateString(src, 14)),
			rowStyle.Width(16).Render(m.truncateString(dst, 14)),
			rowStyle.Width(8).Render(protoRendered),
			rowStyle.Width(10).Render(bandwidth),
			rowStyle.Width(10).Render(bytes),
			rowStyle.Width(8).Render(fmt.Sprintf("%d", flow.Packets)),
			rowStyle.Width(8).Render(fmt.Sprintf("AS%d", flow.SrcASN)),
			rowStyle.Width(8).Render(fmt.Sprintf("AS%d", flow.DstASN)),
			rowStyle.Width(width-89).Render(m.truncateString(geo, width-91)),
		)
		
		rows = append(rows, row)
	}
	
	// Build table
	table := lipgloss.JoinVertical(lipgloss.Left, tableHeader)
	for _, row := range rows {
		table = lipgloss.JoinVertical(lipgloss.Left, table, row)
	}
	
	return table
}

// renderFlowDetails renders detailed information about the selected flow
func (m FlowsModel) renderFlowDetails(width, height int) string {
	if len(m.topFlows) == 0 || m.selected >= len(m.topFlows) {
		return m.renderNoFlowSelection(width, height)
	}
	
	flow := m.topFlows[m.selected]
	
	// Detail styles
	titleStyle := lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color("#00FFFF")).
		Margin(0, 0, 1, 0)
	
	labelStyle := lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color("#CCCCCC"))
	
	valueStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("#FFFFFF"))
	
	// Build details
	var details strings.Builder
	
	details.WriteString(titleStyle.Render("Flow Details"))
	details.WriteString("\n")
	
	details.WriteString(labelStyle.Render("Source: "))
	details.WriteString(valueStyle.Render(fmt.Sprintf("%s:%d", flow.SrcIP, flow.SrcPort)))
	details.WriteString("\n")
	
	details.WriteString(labelStyle.Render("Destination: "))
	details.WriteString(valueStyle.Render(fmt.Sprintf("%s:%d", flow.DstIP, flow.DstPort)))
	details.WriteString("\n")
	
	details.WriteString(labelStyle.Render("Protocol: "))
	details.WriteString(valueStyle.Render(flow.Protocol))
	details.WriteString("\n\n")
	
	details.WriteString(labelStyle.Render("Traffic:"))
	details.WriteString("\n")
	details.WriteString(fmt.Sprintf("  Bandwidth: %s\n", valueStyle.Render(m.formatBandwidth(flow.Mbps))))
	details.WriteString(fmt.Sprintf("  Bytes: %s\n", valueStyle.Render(m.formatBytes(flow.Bytes))))
	details.WriteString(fmt.Sprintf("  Packets: %s\n", valueStyle.Render(fmt.Sprintf("%d", flow.Packets))))
	details.WriteString("\n")
	
	details.WriteString(labelStyle.Render("ASN Information:"))
	details.WriteString("\n")
	details.WriteString(fmt.Sprintf("  Source ASN: %s\n", valueStyle.Render(fmt.Sprintf("AS%d", flow.SrcASN))))
	details.WriteString(fmt.Sprintf("  Dest ASN: %s\n", valueStyle.Render(fmt.Sprintf("AS%d", flow.DstASN))))
	details.WriteString("\n")
	
	details.WriteString(labelStyle.Render("Geography:"))
	details.WriteString("\n")
	details.WriteString(fmt.Sprintf("  Source: %s\n", valueStyle.Render(flow.GeoSrc)))
	details.WriteString(fmt.Sprintf("  Destination: %s\n", valueStyle.Render(flow.GeoDst)))
	details.WriteString("\n")
	
	details.WriteString(labelStyle.Render("Timing:"))
	details.WriteString("\n")
	details.WriteString(fmt.Sprintf("  First Seen: %s\n", valueStyle.Render(flow.FirstSeen.Format("15:04:05"))))
	details.WriteString(fmt.Sprintf("  Last Seen: %s\n", valueStyle.Render(flow.LastSeen.Format("15:04:05"))))
	details.WriteString(fmt.Sprintf("  Duration: %s\n", valueStyle.Render(flow.LastSeen.Sub(flow.FirstSeen).String())))
	
	if len(flow.Flags) > 0 {
		details.WriteString("\n")
		details.WriteString(labelStyle.Render("Flags: "))
		flagsStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("#FFFF00"))
		details.WriteString(flagsStyle.Render(strings.Join(flow.Flags, ", ")))
	}
	
	// Wrap in container
	containerStyle := lipgloss.NewStyle().
		Width(width).
		Height(height).
		Padding(1).
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("#666666"))
	
	return containerStyle.Render(details.String())
}

// renderNoFlowSelection renders message when no flow is selected
func (m FlowsModel) renderNoFlowSelection(width, height int) string {
	style := lipgloss.NewStyle().
		Width(width).
		Height(height).
		Align(lipgloss.Center, lipgloss.Center).
		Foreground(lipgloss.Color("#666666"))
	
	return style.Render("Select a flow to view details")
}

// renderTelemetry renders the telemetry sparklines
func (m FlowsModel) renderTelemetry() string {
	telemetryStyle := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("#666666")).
		Padding(0, 1).
		Margin(1, 0, 0, 0)
	
	telemetryContent := m.telemetry.RenderAll()
	
	return telemetryStyle.Render("Flow Telemetry:\n" + telemetryContent)
}

// renderFooter renders the footer with help and statistics
func (m FlowsModel) renderFooter() string {
	footerStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("#666666")).
		Padding(0, 1)
	
	totalBandwidth := 0.0
	for _, flow := range m.flows {
		totalBandwidth += flow.Mbps
	}
	
	help := "↑↓: Navigate • 1: Sort by Bandwidth • 2: Sort by Bytes • 3: Sort by Packets • a: Auto-refresh • r: Refresh"
	stats := fmt.Sprintf("Total Flows: %d | Total Bandwidth: %.1f Gbps | Top 5 Shown", 
		len(m.flows), totalBandwidth/1000)
	
	return footerStyle.Render(fmt.Sprintf("%s\n%s", help, stats))
}

// Helper methods

func (m *FlowsModel) updateTopFlows() {
	// Sort flows based on current sort criteria
	sortedFlows := make([]*FlowEntry, len(m.flows))
	copy(sortedFlows, m.flows)
	
	switch m.sortBy {
	case "mbps":
		sort.Slice(sortedFlows, func(i, j int) bool {
			return sortedFlows[i].Mbps > sortedFlows[j].Mbps
		})
	case "bytes":
		sort.Slice(sortedFlows, func(i, j int) bool {
			return sortedFlows[i].Bytes > sortedFlows[j].Bytes
		})
	case "packets":
		sort.Slice(sortedFlows, func(i, j int) bool {
			return sortedFlows[i].Packets > sortedFlows[j].Packets
		})
	}
	
	// Take top 5
	if len(sortedFlows) > 5 {
		m.topFlows = sortedFlows[:5]
	} else {
		m.topFlows = sortedFlows
	}
	
	// Reset selection if out of bounds
	if m.selected >= len(m.topFlows) {
		m.selected = 0
	}
}

func (m *FlowsModel) updateTelemetry() {
	// Calculate telemetry metrics
	totalFlows := float64(len(m.flows))
	totalBandwidth := 0.0
	protocolCount := make(map[string]int)
	
	for _, flow := range m.flows {
		totalBandwidth += flow.Mbps
		protocolCount[flow.Protocol]++
	}
	
	// Update sparklines
	m.telemetry.UpdateData("total_flows", totalFlows)
	m.telemetry.UpdateData("total_bandwidth", totalBandwidth/1000) // Convert to Gbps
	m.telemetry.UpdateData("top_talkers", float64(len(m.topFlows)))
	m.telemetry.UpdateData("protocols", float64(len(protocolCount)))
}

func (m FlowsModel) formatBandwidth(mbps float64) string {
	if mbps < 1 {
		return fmt.Sprintf("%.1f Kbps", mbps*1000)
	} else if mbps < 1000 {
		return fmt.Sprintf("%.1f Mbps", mbps)
	} else {
		return fmt.Sprintf("%.1f Gbps", mbps/1000)
	}
}

func (m FlowsModel) formatBytes(bytes uint64) string {
	if bytes < 1024 {
		return fmt.Sprintf("%dB", bytes)
	} else if bytes < 1024*1024 {
		return fmt.Sprintf("%.1fKB", float64(bytes)/1024)
	} else if bytes < 1024*1024*1024 {
		return fmt.Sprintf("%.1fMB", float64(bytes)/(1024*1024))
	} else {
		return fmt.Sprintf("%.1fGB", float64(bytes)/(1024*1024*1024))
	}
}

func (m FlowsModel) truncateString(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	return s[:maxLen-3] + "..."
}

// Commands

func (m FlowsModel) fetchFlowsData() tea.Cmd {
	return tea.Cmd(func() tea.Msg {
		// Generate mock flow data for development
		flows := m.generateMockFlows()
		return FlowsDataMsg{Flows: flows}
	})
}

func (m FlowsModel) generateMockFlows() []*FlowEntry {
	now := time.Now()
	
	return []*FlowEntry{
		{
			SrcIP:     "192.168.1.100",
			DstIP:     "8.8.8.8",
			SrcPort:   12345,
			DstPort:   53,
			Protocol:  "UDP",
			Bytes:     1024000,
			Packets:   800,
			Mbps:      850.5,
			SrcASN:    65001,
			DstASN:    15169,
			FirstSeen: now.Add(-5 * time.Minute),
			LastSeen:  now,
			Flags:     []string{"DNS"},
			GeoSrc:    "US",
			GeoDst:    "US",
		},
		{
			SrcIP:     "10.0.0.50",
			DstIP:     "1.1.1.1",
			SrcPort:   443,
			DstPort:   443,
			Protocol:  "TCP",
			Bytes:     15360000,
			Packets:   12000,
			Mbps:      640.2,
			SrcASN:    65002,
			DstASN:    13335,
			FirstSeen: now.Add(-10 * time.Minute),
			LastSeen:  now,
			Flags:     []string{"HTTPS", "Encrypted"},
			GeoSrc:    "US",
			GeoDst:    "US",
		},
		{
			SrcIP:     "172.16.0.200",
			DstIP:     "203.0.113.10",
			SrcPort:   80,
			DstPort:   80,
			Protocol:  "TCP",
			Bytes:     512000000,
			Packets:   400000,
			Mbps:      520.8,
			SrcASN:    65003,
			DstASN:    64512,
			FirstSeen: now.Add(-2 * time.Minute),
			LastSeen:  now,
			Flags:     []string{"HTTP", "Suspicious"},
			GeoSrc:    "US",
			GeoDst:    "Unknown",
		},
		{
			SrcIP:     "192.168.2.50",
			DstIP:     "185.199.108.153",
			SrcPort:   443,
			DstPort:   443,
			Protocol:  "TCP",
			Bytes:     8500000,
			Packets:   6800,
			Mbps:      420.1,
			SrcASN:    65004,
			DstASN:    36459,
			FirstSeen: now.Add(-15 * time.Minute),
			LastSeen:  now,
			Flags:     []string{"HTTPS", "GitHub"},
			GeoSrc:    "US",
			GeoDst:    "US",
		},
		{
			SrcIP:     "10.1.1.100",
			DstIP:     "142.250.191.14",
			SrcPort:   443,
			DstPort:   443,
			Protocol:  "TCP",
			Bytes:     6200000,
			Packets:   4960,
			Mbps:      310.5,
			SrcASN:    65005,
			DstASN:    15169,
			FirstSeen: now.Add(-8 * time.Minute),
			LastSeen:  now,
			Flags:     []string{"HTTPS", "Google"},
			GeoSrc:    "US",
			GeoDst:    "US",
		},
	}
}

// Messages

type FlowsDataMsg struct {
	Flows []*FlowEntry
}

type FlowTickMsg time.Time