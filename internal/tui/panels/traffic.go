package panels

import (
	"fmt"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// TrafficModel represents the traffic panel
type TrafficModel struct {
	focusASN int
	width    int
	height   int
	data     []map[string]interface{}
	selected int
	offset   int
}

// NewTrafficModel creates a new traffic model
func NewTrafficModel(focusASN int) TrafficModel {
	return TrafficModel{
		focusASN: focusASN,
		data:     make([]map[string]interface{}, 0),
	}
}

// Init initializes the traffic model
func (m TrafficModel) Init() tea.Cmd {
	return nil
}

// Update handles messages for the traffic panel
func (m TrafficModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "up", "k":
			if m.selected > 0 {
				m.selected--
				if m.selected < m.offset {
					m.offset--
				}
			}
		case "down", "j":
			if m.selected < len(m.data)-1 {
				m.selected++
				maxVisible := m.height - 8 // Account for header, stats, and borders
				if m.selected >= m.offset+maxVisible {
					m.offset++
				}
			}
		case "home":
			m.selected = 0
			m.offset = 0
		case "end":
			m.selected = len(m.data) - 1
			maxVisible := m.height - 8
			if len(m.data) > maxVisible {
				m.offset = len(m.data) - maxVisible
			}
		}
	}
	return m, nil
}

// View renders the traffic panel
func (m TrafficModel) View() string {
	if len(m.data) == 0 {
		return m.renderLoading()
	}

	return m.renderTraffic()
}

// SetSize sets the panel size
func (m *TrafficModel) SetSize(width, height int) {
	m.width = width
	m.height = height
}

// UpdateData updates the panel data
func (m *TrafficModel) UpdateData(data interface{}) {
	if d, ok := data.([]map[string]interface{}); ok {
		m.data = d
		if m.selected >= len(m.data) {
			m.selected = 0
			m.offset = 0
		}
	}
}

// renderLoading renders loading state
func (m TrafficModel) renderLoading() string {
	style := lipgloss.NewStyle().
		Width(m.width).
		Height(m.height).
		Align(lipgloss.Center, lipgloss.Center).
		Foreground(lipgloss.Color("#666666"))

	return style.Render("Loading traffic data...")
}

// renderTraffic renders the traffic content
func (m TrafficModel) renderTraffic() string {
	// Header style
	headerStyle := lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color("#FAFAFA")).
		Background(lipgloss.Color("#7D56F4")).
		Padding(0, 1).
		Width(m.width - 2).
		Align(lipgloss.Center)

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

	// Status styles
	normalStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("#00FF00")).
		Bold(true)

	suspiciousStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("#FFFF00")).
		Bold(true)

	ddosStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("#FF0000")).
		Bold(true)

	// Build header
	title := "Network Traffic (NetFlow/sFlow)"
	if m.focusASN > 0 {
		title = fmt.Sprintf("Network Traffic - AS%d", m.focusASN)
	}
	header := headerStyle.Render(title)

	// Traffic statistics bar
	statsBarStyle := lipgloss.NewStyle().
		Background(lipgloss.Color("#333333")).
		Foreground(lipgloss.Color("#FFFFFF")).
		Padding(0, 1).
		Width(m.width - 2)

	totalBytes := 0
	normalFlows := 0
	suspiciousFlows := 0
	ddosFlows := 0

	for _, flow := range m.data {
		if bytes, ok := flow["bytes"].(int); ok {
			totalBytes += bytes
		}
		status := m.getStringValue(flow, "status", "Normal")
		switch status {
		case "Normal":
			normalFlows++
		case "Suspicious":
			suspiciousFlows++
		case "DDoS":
			ddosFlows++
		}
	}

	// Format bytes
	bytesStr := m.formatBytes(totalBytes)
	
	statsBar := statsBarStyle.Render(fmt.Sprintf(
		"Total: %s | Flows: %d | Normal: %d | Suspicious: %d | DDoS: %d | Rate: %.1f Mbps",
		bytesStr, len(m.data), normalFlows, suspiciousFlows, ddosFlows, float64(totalBytes)*8/1000000,
	))

	// Table header
	tableHeader := lipgloss.JoinHorizontal(lipgloss.Top,
		tableHeaderStyle.Width(16).Render("SOURCE IP"),
		tableHeaderStyle.Width(16).Render("DEST IP"),
		tableHeaderStyle.Width(8).Render("PROTO"),
		tableHeaderStyle.Width(10).Render("BYTES"),
		tableHeaderStyle.Width(8).Render("PACKETS"),
		tableHeaderStyle.Width(12).Render("STATUS"),
		tableHeaderStyle.Width(8).Render("DURATION"),
		tableHeaderStyle.Width(m.width-80).Render("FLAGS"),
	)

	// Table rows
	var rows []string
	maxVisible := m.height - 8 // Account for header, stats, and borders

	for i := m.offset; i < len(m.data) && i < m.offset+maxVisible; i++ {
		flow := m.data[i]

		src := m.getStringValue(flow, "src", "N/A")
		dst := m.getStringValue(flow, "dst", "N/A")
		protocol := m.getStringValue(flow, "protocol", "N/A")
		bytes := m.getStringValue(flow, "bytes", "0")
		packets := m.getStringValue(flow, "packets", "0")
		status := m.getStringValue(flow, "status", "Normal")
		duration := m.getStringValue(flow, "duration", "N/A")
		flags := m.getStringValue(flow, "flags", "N/A")

		// Format bytes for display
		if bytesInt, ok := flow["bytes"].(int); ok {
			bytes = m.formatBytes(bytesInt)
		}

		// Status styling
		var statusRendered string
		switch status {
		case "Normal":
			statusRendered = normalStyle.Render(status)
		case "Suspicious":
			statusRendered = suspiciousStyle.Render(status)
		case "DDoS":
			statusRendered = ddosStyle.Render(status)
		default:
			statusRendered = status
		}

		var rowStyle lipgloss.Style
		if i == m.selected {
			rowStyle = selectedRowStyle
		} else {
			rowStyle = normalRowStyle
		}

		row := lipgloss.JoinHorizontal(lipgloss.Top,
			rowStyle.Width(16).Render(src),
			rowStyle.Width(16).Render(dst),
			rowStyle.Width(8).Render(protocol),
			rowStyle.Width(10).Render(bytes),
			rowStyle.Width(8).Render(packets),
			rowStyle.Width(12).Render(statusRendered),
			rowStyle.Width(8).Render(duration),
			rowStyle.Width(m.width-80).Render(flags),
		)

		rows = append(rows, row)
	}

	// Build table
	table := lipgloss.JoinVertical(lipgloss.Left, tableHeader)
	for _, row := range rows {
		table = lipgloss.JoinVertical(lipgloss.Left, table, row)
	}

	// Traffic visualization (simple bar chart)
	chartStyle := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("#666666")).
		Padding(0, 1).
		Margin(1, 0, 0, 0)

	chart := m.renderTrafficChart()
	chartSection := chartStyle.Render("Traffic Pattern:\n" + chart)

	// Info footer
	infoStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("#666666")).
		Padding(0, 1)

	info := infoStyle.Render(fmt.Sprintf(
		"Showing %d-%d of %d flows | Selected: %d | ↑↓: Navigate | Real-time monitoring",
		m.offset+1,
		min(m.offset+maxVisible, len(m.data)),
		len(m.data),
		m.selected+1,
	))

	// Container style
	containerStyle := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("#7D56F4")).
		Padding(1).
		Width(m.width - 2).
		Height(m.height - 2)

	content := lipgloss.JoinVertical(lipgloss.Left, header, statsBar, table, chartSection, info)

	return containerStyle.Render(content)
}

// renderTrafficChart renders a simple ASCII traffic chart
func (m TrafficModel) renderTrafficChart() string {
	if len(m.data) == 0 {
		return "No data available"
	}

	// Create a simple bar chart showing traffic distribution
	chartWidth := 50
	maxBytes := 0
	
	// Find max bytes for scaling
	for _, flow := range m.data {
		if bytes, ok := flow["bytes"].(int); ok && bytes > maxBytes {
			maxBytes = bytes
		}
	}

	if maxBytes == 0 {
		return "No traffic data"
	}

	var chart strings.Builder
	chart.WriteString("Traffic Distribution (last 10 flows):\n")

	// Show up to 10 flows in the chart
	displayCount := min(len(m.data), 10)
	for i := 0; i < displayCount; i++ {
		flow := m.data[i]
		bytes, ok := flow["bytes"].(int)
		if !ok {
			bytes = 0
		}

		// Calculate bar length
		barLength := int(float64(bytes) / float64(maxBytes) * float64(chartWidth))
		
		// Create bar
		bar := strings.Repeat("█", barLength)
		if barLength < chartWidth {
			bar += strings.Repeat("░", chartWidth-barLength)
		}

		// Color based on status
		status := m.getStringValue(flow, "status", "Normal")
		var barColored string
		switch status {
		case "Normal":
			barColored = lipgloss.NewStyle().Foreground(lipgloss.Color("#00FF00")).Render(bar)
		case "Suspicious":
			barColored = lipgloss.NewStyle().Foreground(lipgloss.Color("#FFFF00")).Render(bar)
		case "DDoS":
			barColored = lipgloss.NewStyle().Foreground(lipgloss.Color("#FF0000")).Render(bar)
		default:
			barColored = bar
		}

		src := m.getStringValue(flow, "src", "N/A")
		chart.WriteString(fmt.Sprintf("%-15s %s %s\n", src, barColored, m.formatBytes(bytes)))
	}

	return chart.String()
}

// formatBytes formats bytes into human readable format
func (m TrafficModel) formatBytes(bytes int) string {
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

// getStringValue safely gets a string value from a map
func (m TrafficModel) getStringValue(data map[string]interface{}, key, defaultValue string) string {
	if val, ok := data[key]; ok {
		return fmt.Sprintf("%v", val)
	}
	return defaultValue
}