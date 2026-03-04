package tui

import (
	"context"
	"fmt"
	"time"

	"github.com/bgpin/bgpin/internal/tui/panels"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// Panel represents different TUI panels
type Panel int

const (
	SummaryPanel Panel = iota
	RoutesPanel
	NeighborsPanel
	TrafficPanel
)

// Model represents the main TUI model
type Model struct {
	config          Config
	refreshInterval time.Duration
	activePanel     Panel
	width           int
	height          int
	
	// Panel models
	summary   panels.SummaryModel
	routes    panels.RoutesModel
	neighbors panels.NeighborsModel
	traffic   panels.TrafficModel
	
	// State
	lastUpdate time.Time
	loading    bool
	err        error
}

// DataUpdateMsg represents a data update message
type DataUpdateMsg struct {
	Panel Panel
	Data  interface{}
	Error error
}

// TickMsg represents a tick for refresh
type TickMsg time.Time

// NewModel creates a new TUI model
func NewModel(config Config, refreshInterval time.Duration) *Model {
	activePanel := SummaryPanel
	if config.StartWithFlows {
		activePanel = TrafficPanel
	}

	return &Model{
		config:          config,
		refreshInterval: refreshInterval,
		activePanel:     activePanel,
		summary:         panels.NewSummaryModel(config.FocusASN),
		routes:          panels.NewRoutesModel(config.FocusASN),
		neighbors:       panels.NewNeighborsModel(config.FocusASN),
		traffic:         panels.NewTrafficModel(config.FocusASN),
		lastUpdate:      time.Now(),
	}
}

// Init initializes the model
func (m *Model) Init() tea.Cmd {
	return tea.Batch(
		m.summary.Init(),
		m.routes.Init(),
		m.neighbors.Init(),
		m.traffic.Init(),
		tea.Tick(m.refreshInterval, func(t time.Time) tea.Msg {
			return TickMsg(t)
		}),
	)
}

// Update handles messages
func (m *Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmds []tea.Cmd

	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		
		// Update panel sizes
		panelWidth := m.width
		panelHeight := m.height - 4 // Reserve space for header and footer
		
		m.summary.SetSize(panelWidth, panelHeight)
		m.routes.SetSize(panelWidth, panelHeight)
		m.neighbors.SetSize(panelWidth, panelHeight)
		m.traffic.SetSize(panelWidth, panelHeight)

	case tea.KeyMsg:
		switch msg.String() {
		case "q", "ctrl+c":
			return m, tea.Quit
		case "tab":
			m.nextPanel()
		case "shift+tab":
			m.prevPanel()
		case "r":
			m.loading = true
			return m, m.refreshData()
		case "h", "?":
			// TODO: Show help
		case "1":
			m.activePanel = SummaryPanel
		case "2":
			m.activePanel = RoutesPanel
		case "3":
			m.activePanel = NeighborsPanel
		case "4":
			m.activePanel = TrafficPanel
		}

	case TickMsg:
		m.loading = true
		cmds = append(cmds, m.refreshData())
		cmds = append(cmds, tea.Tick(m.refreshInterval, func(t time.Time) tea.Msg {
			return TickMsg(t)
		}))

	case DataUpdateMsg:
		m.loading = false
		m.lastUpdate = time.Now()
		if msg.Error != nil {
			m.err = msg.Error
		} else {
			m.err = nil
			// Update the appropriate panel
			switch msg.Panel {
			case SummaryPanel:
				m.summary.UpdateData(msg.Data)
			case RoutesPanel:
				m.routes.UpdateData(msg.Data)
			case NeighborsPanel:
				m.neighbors.UpdateData(msg.Data)
			case TrafficPanel:
				m.traffic.UpdateData(msg.Data)
			}
		}
	}

	// Update active panel
	switch m.activePanel {
	case SummaryPanel:
		newModel, cmd := m.summary.Update(msg)
		m.summary = newModel.(panels.SummaryModel)
		if cmd != nil {
			cmds = append(cmds, cmd)
		}
	case RoutesPanel:
		newModel, cmd := m.routes.Update(msg)
		m.routes = newModel.(panels.RoutesModel)
		if cmd != nil {
			cmds = append(cmds, cmd)
		}
	case NeighborsPanel:
		newModel, cmd := m.neighbors.Update(msg)
		m.neighbors = newModel.(panels.NeighborsModel)
		if cmd != nil {
			cmds = append(cmds, cmd)
		}
	case TrafficPanel:
		newModel, cmd := m.traffic.Update(msg)
		m.traffic = newModel.(panels.TrafficModel)
		if cmd != nil {
			cmds = append(cmds, cmd)
		}
	}

	return m, tea.Batch(cmds...)
}

// View renders the TUI
func (m *Model) View() string {
	if m.width == 0 {
		return "Initializing bgptop..."
	}

	// Header
	header := m.renderHeader()
	
	// Active panel content
	var content string
	switch m.activePanel {
	case SummaryPanel:
		content = m.summary.View()
	case RoutesPanel:
		content = m.routes.View()
	case NeighborsPanel:
		content = m.neighbors.View()
	case TrafficPanel:
		content = m.traffic.View()
	}

	// Footer
	footer := m.renderFooter()

	return lipgloss.JoinVertical(lipgloss.Left, header, content, footer)
}

// renderHeader renders the header with navigation tabs
func (m *Model) renderHeader() string {
	titleStyle := lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color("#FAFAFA")).
		Background(lipgloss.Color("#7D56F4")).
		Padding(0, 1)

	tabStyle := lipgloss.NewStyle().
		Padding(0, 1).
		Margin(0, 1, 0, 0)

	activeTabStyle := tabStyle.Copy().
		Bold(true).
		Foreground(lipgloss.Color("#FAFAFA")).
		Background(lipgloss.Color("#7D56F4"))

	inactiveTabStyle := tabStyle.Copy().
		Foreground(lipgloss.Color("#666666"))

	title := titleStyle.Render("bgptop")
	
	tabs := []string{
		"1:Summary",
		"2:Routes", 
		"3:Neighbors",
		"4:Traffic",
	}

	var renderedTabs []string
	for i, tab := range tabs {
		if Panel(i) == m.activePanel {
			renderedTabs = append(renderedTabs, activeTabStyle.Render(tab))
		} else {
			renderedTabs = append(renderedTabs, inactiveTabStyle.Render(tab))
		}
	}

	tabsStr := lipgloss.JoinHorizontal(lipgloss.Top, renderedTabs...)
	
	// Status indicator
	status := "●"
	statusColor := lipgloss.Color("#00FF00") // Green
	if m.loading {
		status = "◐"
		statusColor = lipgloss.Color("#FFFF00") // Yellow
	}
	if m.err != nil {
		status = "●"
		statusColor = lipgloss.Color("#FF0000") // Red
	}

	statusStyle := lipgloss.NewStyle().Foreground(statusColor)
	statusStr := statusStyle.Render(status)

	// Time since last update
	timeSince := time.Since(m.lastUpdate).Truncate(time.Second)
	timeStr := fmt.Sprintf("Updated: %s ago", timeSince)
	if m.loading {
		timeStr = "Updating..."
	}

	timeStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("#666666")).
		Align(lipgloss.Right)

	headerLeft := lipgloss.JoinHorizontal(lipgloss.Top, title, " ", tabsStr)
	headerRight := lipgloss.JoinHorizontal(lipgloss.Top, statusStr, " ", timeStr)

	headerStyle := lipgloss.NewStyle().
		Width(m.width).
		Padding(0, 1).
		Border(lipgloss.NormalBorder(), false, false, true, false)

	header := lipgloss.JoinHorizontal(
		lipgloss.Top,
		headerLeft,
		timeStyle.Width(m.width-lipgloss.Width(headerLeft)).Render(headerRight),
	)

	return headerStyle.Render(header)
}

// renderFooter renders the footer with help text
func (m *Model) renderFooter() string {
	footerStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("#666666")).
		Padding(0, 1).
		Border(lipgloss.NormalBorder(), true, false, false, false)

	help := "Tab: Next Panel • Shift+Tab: Prev Panel • r: Refresh • q: Quit • h: Help"
	if m.err != nil {
		help = fmt.Sprintf("Error: %s • %s", m.err.Error(), help)
	}

	return footerStyle.Width(m.width).Render(help)
}

// nextPanel switches to the next panel
func (m *Model) nextPanel() {
	m.activePanel = (m.activePanel + 1) % 4
}

// prevPanel switches to the previous panel
func (m *Model) prevPanel() {
	if m.activePanel == 0 {
		m.activePanel = 3
	} else {
		m.activePanel--
	}
}

// refreshData triggers data refresh for all panels
func (m *Model) refreshData() tea.Cmd {
	return tea.Batch(
		m.fetchSummaryData(),
		m.fetchRoutesData(),
		m.fetchNeighborsData(),
		m.fetchTrafficData(),
	)
}

// Data fetching commands
func (m *Model) fetchSummaryData() tea.Cmd {
	return tea.Cmd(func() tea.Msg {
		// TODO: Implement actual data fetching
		data := map[string]interface{}{
			"asn":         m.config.FocusASN,
			"routes":      12543,
			"neighbors":   8,
			"traffic":     "1.2 Gbps",
			"status":      "Active",
		}
		return DataUpdateMsg{Panel: SummaryPanel, Data: data}
	})
}

func (m *Model) fetchRoutesData() tea.Cmd {
	return tea.Cmd(func() tea.Msg {
		// TODO: Implement actual data fetching
		data := []map[string]interface{}{
			{"prefix": "8.8.8.0/24", "asn": "AS15169", "status": "Valid"},
			{"prefix": "1.1.1.0/24", "asn": "AS13335", "status": "Valid"},
			{"prefix": "192.168.1.0/24", "asn": "AS64512", "status": "Invalid"},
		}
		return DataUpdateMsg{Panel: RoutesPanel, Data: data}
	})
}

func (m *Model) fetchNeighborsData() tea.Cmd {
	return tea.Cmd(func() tea.Msg {
		// TODO: Implement actual data fetching
		data := []map[string]interface{}{
			{"asn": "AS15169", "type": "Upstream", "status": "Established"},
			{"asn": "AS13335", "type": "Peer", "status": "Established"},
			{"asn": "AS64512", "type": "Downstream", "status": "Idle"},
		}
		return DataUpdateMsg{Panel: NeighborsPanel, Data: data}
	})
}

func (m *Model) fetchTrafficData() tea.Cmd {
	return tea.Cmd(func() tea.Msg {
		// TODO: Implement actual data fetching
		data := []map[string]interface{}{
			{"src": "192.168.1.100", "dst": "8.8.8.8", "protocol": "UDP", "bytes": 1024, "status": "Normal"},
			{"src": "10.0.0.50", "dst": "1.1.1.1", "protocol": "TCP", "bytes": 15360, "status": "Suspicious"},
			{"src": "172.16.0.200", "dst": "203.0.113.10", "protocol": "TCP", "bytes": 512000, "status": "DDoS"},
		}
		return DataUpdateMsg{Panel: TrafficPanel, Data: data}
	})
}

// startDataFetching starts background data fetching
func (m *Model) startDataFetching(ctx context.Context) {
	// This would be used for real-time data streaming
	// For now, we rely on the tick-based updates
}