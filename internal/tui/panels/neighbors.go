package panels

import (
	"fmt"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// NeighborsModel represents the neighbors panel
type NeighborsModel struct {
	focusASN int
	width    int
	height   int
	data     []map[string]interface{}
	selected int
	offset   int
}

// NewNeighborsModel creates a new neighbors model
func NewNeighborsModel(focusASN int) NeighborsModel {
	return NeighborsModel{
		focusASN: focusASN,
		data:     make([]map[string]interface{}, 0),
	}
}

// Init initializes the neighbors model
func (m NeighborsModel) Init() tea.Cmd {
	return nil
}

// Update handles messages for the neighbors panel
func (m NeighborsModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
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
				maxVisible := m.height - 6
				if m.selected >= m.offset+maxVisible {
					m.offset++
				}
			}
		case "home":
			m.selected = 0
			m.offset = 0
		case "end":
			m.selected = len(m.data) - 1
			maxVisible := m.height - 6
			if len(m.data) > maxVisible {
				m.offset = len(m.data) - maxVisible
			}
		}
	}
	return m, nil
}

// View renders the neighbors panel
func (m NeighborsModel) View() string {
	if len(m.data) == 0 {
		return m.renderLoading()
	}

	return m.renderNeighbors()
}

// SetSize sets the panel size
func (m *NeighborsModel) SetSize(width, height int) {
	m.width = width
	m.height = height
}

// UpdateData updates the panel data
func (m *NeighborsModel) UpdateData(data interface{}) {
	if d, ok := data.([]map[string]interface{}); ok {
		m.data = d
		if m.selected >= len(m.data) {
			m.selected = 0
			m.offset = 0
		}
	}
}

// renderLoading renders loading state
func (m NeighborsModel) renderLoading() string {
	style := lipgloss.NewStyle().
		Width(m.width).
		Height(m.height).
		Align(lipgloss.Center, lipgloss.Center).
		Foreground(lipgloss.Color("#666666"))

	return style.Render("Loading neighbors data...")
}

// renderNeighbors renders the neighbors content
func (m NeighborsModel) renderNeighbors() string {
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
	establishedStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("#00FF00")).
		Bold(true)

	idleStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("#FF0000")).
		Bold(true)

	connectStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("#FFFF00")).
		Bold(true)

	// Type styles
	upstreamStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("#00FFFF")).
		Bold(true)

	peerStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("#FF00FF")).
		Bold(true)

	downstreamStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("#FFFF00")).
		Bold(true)

	// Build header
	title := "BGP Neighbors"
	if m.focusASN > 0 {
		title = fmt.Sprintf("BGP Neighbors - AS%d", m.focusASN)
	}
	header := headerStyle.Render(title)

	// Table header
	tableHeader := lipgloss.JoinHorizontal(lipgloss.Top,
		tableHeaderStyle.Width(15).Render("ASN"),
		tableHeaderStyle.Width(12).Render("TYPE"),
		tableHeaderStyle.Width(15).Render("STATUS"),
		tableHeaderStyle.Width(18).Render("REMOTE IP"),
		tableHeaderStyle.Width(10).Render("UPTIME"),
		tableHeaderStyle.Width(8).Render("ROUTES"),
		tableHeaderStyle.Width(m.width-80).Render("DESCRIPTION"),
	)

	// Table rows
	var rows []string
	maxVisible := m.height - 6

	for i := m.offset; i < len(m.data) && i < m.offset+maxVisible; i++ {
		neighbor := m.data[i]

		asn := m.getStringValue(neighbor, "asn", "N/A")
		neighborType := m.getStringValue(neighbor, "type", "Unknown")
		status := m.getStringValue(neighbor, "status", "Unknown")
		remoteIP := m.getStringValue(neighbor, "remote_ip", "N/A")
		uptime := m.getStringValue(neighbor, "uptime", "N/A")
		routes := m.getStringValue(neighbor, "routes", "0")
		description := m.getStringValue(neighbor, "description", "N/A")

		// Status styling
		var statusRendered string
		switch status {
		case "Established":
			statusRendered = establishedStyle.Render(status)
		case "Idle":
			statusRendered = idleStyle.Render(status)
		case "Connect":
			statusRendered = connectStyle.Render(status)
		default:
			statusRendered = status
		}

		// Type styling
		var typeRendered string
		switch neighborType {
		case "Upstream":
			typeRendered = upstreamStyle.Render(neighborType)
		case "Peer":
			typeRendered = peerStyle.Render(neighborType)
		case "Downstream":
			typeRendered = downstreamStyle.Render(neighborType)
		default:
			typeRendered = neighborType
		}

		var rowStyle lipgloss.Style
		if i == m.selected {
			rowStyle = selectedRowStyle
		} else {
			rowStyle = normalRowStyle
		}

		row := lipgloss.JoinHorizontal(lipgloss.Top,
			rowStyle.Width(15).Render(asn),
			rowStyle.Width(12).Render(typeRendered),
			rowStyle.Width(15).Render(statusRendered),
			rowStyle.Width(18).Render(remoteIP),
			rowStyle.Width(10).Render(uptime),
			rowStyle.Width(8).Render(routes),
			rowStyle.Width(m.width-80).Render(description),
		)

		rows = append(rows, row)
	}

	// Build table
	table := lipgloss.JoinVertical(lipgloss.Left, tableHeader)
	for _, row := range rows {
		table = lipgloss.JoinVertical(lipgloss.Left, table, row)
	}

	// Statistics section
	statsStyle := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("#666666")).
		Padding(0, 1).
		Margin(1, 0, 0, 0)

	established := 0
	idle := 0
	upstream := 0
	peer := 0
	downstream := 0

	for _, neighbor := range m.data {
		status := m.getStringValue(neighbor, "status", "Unknown")
		neighborType := m.getStringValue(neighbor, "type", "Unknown")

		if status == "Established" {
			established++
		} else if status == "Idle" {
			idle++
		}

		switch neighborType {
		case "Upstream":
			upstream++
		case "Peer":
			peer++
		case "Downstream":
			downstream++
		}
	}

	stats := statsStyle.Render(fmt.Sprintf(
		"Total: %d | Established: %d | Idle: %d | Upstream: %d | Peers: %d | Downstream: %d",
		len(m.data), established, idle, upstream, peer, downstream,
	))

	// Info footer
	infoStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("#666666")).
		Padding(0, 1)

	info := infoStyle.Render(fmt.Sprintf(
		"Showing %d-%d of %d neighbors | Selected: %d | ↑↓: Navigate",
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

	content := lipgloss.JoinVertical(lipgloss.Left, header, table, stats, info)

	return containerStyle.Render(content)
}

// getStringValue safely gets a string value from a map
func (m NeighborsModel) getStringValue(data map[string]interface{}, key, defaultValue string) string {
	if val, ok := data[key]; ok {
		return fmt.Sprintf("%v", val)
	}
	return defaultValue
}