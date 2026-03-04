package panels

import (
	"fmt"
	"strconv"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// SummaryModel represents the summary panel
type SummaryModel struct {
	focusASN int
	width    int
	height   int
	data     map[string]interface{}
}

// NewSummaryModel creates a new summary model
func NewSummaryModel(focusASN int) SummaryModel {
	return SummaryModel{
		focusASN: focusASN,
		data:     make(map[string]interface{}),
	}
}

// Init initializes the summary model
func (m SummaryModel) Init() tea.Cmd {
	return nil
}

// Update handles messages for the summary panel
func (m SummaryModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	return m, nil
}

// View renders the summary panel
func (m SummaryModel) View() string {
	if len(m.data) == 0 {
		return m.renderLoading()
	}

	return m.renderSummary()
}

// SetSize sets the panel size
func (m *SummaryModel) SetSize(width, height int) {
	m.width = width
	m.height = height
}

// UpdateData updates the panel data
func (m *SummaryModel) UpdateData(data interface{}) {
	if d, ok := data.(map[string]interface{}); ok {
		m.data = d
	}
}

// renderLoading renders loading state
func (m SummaryModel) renderLoading() string {
	style := lipgloss.NewStyle().
		Width(m.width).
		Height(m.height).
		Align(lipgloss.Center, lipgloss.Center).
		Foreground(lipgloss.Color("#666666"))

	return style.Render("Loading summary data...")
}

// renderSummary renders the summary content
func (m SummaryModel) renderSummary() string {
	// Header style
	headerStyle := lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color("#FAFAFA")).
		Background(lipgloss.Color("#7D56F4")).
		Padding(0, 1).
		Width(m.width - 2).
		Align(lipgloss.Center)

	// Card style
	cardStyle := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("#7D56F4")).
		Padding(1, 2).
		Margin(1, 1)

	// Metric style
	metricStyle := lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color("#00FF00"))

	labelStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("#CCCCCC"))

	// Extract data
	routes := m.getStringValue("routes", "0")
	neighbors := m.getStringValue("neighbors", "0")
	traffic := m.getStringValue("traffic", "0 bps")
	status := m.getStringValue("status", "Unknown")

	// Status color
	statusColor := lipgloss.Color("#00FF00") // Green
	if status != "Active" {
		statusColor = lipgloss.Color("#FF0000") // Red
	}
	statusStyle := lipgloss.NewStyle().
		Bold(true).
		Foreground(statusColor)

	// Build header
	title := "BGP Summary"
	if m.focusASN > 0 {
		title = fmt.Sprintf("BGP Summary - AS%d", m.focusASN)
	}
	header := headerStyle.Render(title)

	// Build metrics cards
	routesCard := cardStyle.Render(
		labelStyle.Render("Routes Announced") + "\n" +
		metricStyle.Render(routes),
	)

	neighborsCard := cardStyle.Render(
		labelStyle.Render("BGP Neighbors") + "\n" +
		metricStyle.Render(neighbors),
	)

	trafficCard := cardStyle.Render(
		labelStyle.Render("Current Traffic") + "\n" +
		metricStyle.Render(traffic),
	)

	statusCard := cardStyle.Render(
		labelStyle.Render("Status") + "\n" +
		statusStyle.Render(status),
	)

	// Layout cards in a grid
	topRow := lipgloss.JoinHorizontal(lipgloss.Top, routesCard, neighborsCard)
	bottomRow := lipgloss.JoinHorizontal(lipgloss.Top, trafficCard, statusCard)
	
	cards := lipgloss.JoinVertical(lipgloss.Left, topRow, bottomRow)

	// System info section
	sysInfoStyle := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("#666666")).
		Padding(1, 2).
		Margin(1, 1).
		Width(m.width - 4)

	sysInfo := sysInfoStyle.Render(
		labelStyle.Render("System Information") + "\n\n" +
		fmt.Sprintf("%-20s %s", "Platform:", "bgpin v0.3.0") + "\n" +
		fmt.Sprintf("%-20s %s", "Uptime:", "2h 34m") + "\n" +
		fmt.Sprintf("%-20s %s", "Memory Usage:", "45.2 MB") + "\n" +
		fmt.Sprintf("%-20s %s", "Active Connections:", "127") + "\n" +
		fmt.Sprintf("%-20s %s", "Last Update:", "2 seconds ago"),
	)

	content := lipgloss.JoinVertical(lipgloss.Left, header, cards, sysInfo)

	// Center content if there's extra space
	if m.height > lipgloss.Height(content) {
		contentStyle := lipgloss.NewStyle().
			Width(m.width).
			Height(m.height).
			Align(lipgloss.Center, lipgloss.Top)
		return contentStyle.Render(content)
	}

	return content
}

// getStringValue safely gets a string value from data
func (m SummaryModel) getStringValue(key, defaultValue string) string {
	if val, ok := m.data[key]; ok {
		switch v := val.(type) {
		case string:
			return v
		case int:
			return strconv.Itoa(v)
		case float64:
			return fmt.Sprintf("%.0f", v)
		default:
			return fmt.Sprintf("%v", v)
		}
	}
	return defaultValue
}