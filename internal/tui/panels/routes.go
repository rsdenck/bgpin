package panels

import (
	"fmt"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// RoutesModel represents the routes panel
type RoutesModel struct {
	focusASN int
	width    int
	height   int
	data     []map[string]interface{}
	selected int
	offset   int
}

// NewRoutesModel creates a new routes model
func NewRoutesModel(focusASN int) RoutesModel {
	return RoutesModel{
		focusASN: focusASN,
		data:     make([]map[string]interface{}, 0),
	}
}

// Init initializes the routes model
func (m RoutesModel) Init() tea.Cmd {
	return nil
}

// Update handles messages for the routes panel
func (m RoutesModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
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
				maxVisible := m.height - 6 // Account for header and borders
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

// View renders the routes panel
func (m RoutesModel) View() string {
	if len(m.data) == 0 {
		return m.renderLoading()
	}

	return m.renderRoutes()
}

// SetSize sets the panel size
func (m *RoutesModel) SetSize(width, height int) {
	m.width = width
	m.height = height
}

// UpdateData updates the panel data
func (m *RoutesModel) UpdateData(data interface{}) {
	if d, ok := data.([]map[string]interface{}); ok {
		m.data = d
		// Reset selection if out of bounds
		if m.selected >= len(m.data) {
			m.selected = 0
			m.offset = 0
		}
	}
}

// renderLoading renders loading state
func (m RoutesModel) renderLoading() string {
	style := lipgloss.NewStyle().
		Width(m.width).
		Height(m.height).
		Align(lipgloss.Center, lipgloss.Center).
		Foreground(lipgloss.Color("#666666"))

	return style.Render("Loading routes data...")
}

// renderRoutes renders the routes content
func (m RoutesModel) renderRoutes() string {
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

	validStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("#00FF00")).
		Bold(true)

	invalidStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("#FF0000")).
		Bold(true)

	// Build header
	title := "BGP Routes"
	if m.focusASN > 0 {
		title = fmt.Sprintf("BGP Routes - AS%d", m.focusASN)
	}
	header := headerStyle.Render(title)

	// Table header
	tableHeader := lipgloss.JoinHorizontal(lipgloss.Top,
		tableHeaderStyle.Width(20).Render("PREFIX"),
		tableHeaderStyle.Width(12).Render("ASN"),
		tableHeaderStyle.Width(10).Render("STATUS"),
		tableHeaderStyle.Width(15).Render("NEXT HOP"),
		tableHeaderStyle.Width(8).Render("MED"),
		tableHeaderStyle.Width(m.width-67).Render("PATH"),
	)

	// Table rows
	var rows []string
	maxVisible := m.height - 6 // Account for header, table header, and borders
	
	for i := m.offset; i < len(m.data) && i < m.offset+maxVisible; i++ {
		route := m.data[i]
		
		prefix := m.getStringValue(route, "prefix", "N/A")
		asn := m.getStringValue(route, "asn", "N/A")
		status := m.getStringValue(route, "status", "Unknown")
		nextHop := m.getStringValue(route, "next_hop", "N/A")
		med := m.getStringValue(route, "med", "0")
		path := m.getStringValue(route, "path", "N/A")

		// Status styling
		var statusRendered string
		if status == "Valid" {
			statusRendered = validStyle.Render(status)
		} else {
			statusRendered = invalidStyle.Render(status)
		}

		var rowStyle lipgloss.Style
		if i == m.selected {
			rowStyle = selectedRowStyle
		} else {
			rowStyle = normalRowStyle
		}

		row := lipgloss.JoinHorizontal(lipgloss.Top,
			rowStyle.Width(20).Render(prefix),
			rowStyle.Width(12).Render(asn),
			rowStyle.Width(10).Render(statusRendered),
			rowStyle.Width(15).Render(nextHop),
			rowStyle.Width(8).Render(med),
			rowStyle.Width(m.width-67).Render(path),
		)

		rows = append(rows, row)
	}

	// Build table
	table := lipgloss.JoinVertical(lipgloss.Left, tableHeader)
	for _, row := range rows {
		table = lipgloss.JoinVertical(lipgloss.Left, table, row)
	}

	// Info footer
	infoStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("#666666")).
		Padding(0, 1)

	info := infoStyle.Render(fmt.Sprintf(
		"Showing %d-%d of %d routes | Selected: %d | ↑↓: Navigate | Home/End: Jump",
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

	content := lipgloss.JoinVertical(lipgloss.Left, header, table, info)
	
	return containerStyle.Render(content)
}

// getStringValue safely gets a string value from a map
func (m RoutesModel) getStringValue(data map[string]interface{}, key, defaultValue string) string {
	if val, ok := data[key]; ok {
		return fmt.Sprintf("%v", val)
	}
	return defaultValue
}

// min returns the minimum of two integers
func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}