package panels

import (
	"fmt"
	"strings"
	"time"

	"github.com/bgpin/bgpin/internal/tui/gobgp"
	"github.com/bgpin/bgpin/internal/tui/telemetry"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// PeersModel represents the advanced peers panel
type PeersModel struct {
	width       int
	height      int
	peers       []*gobgp.PeerInfo
	selected    int
	offset      int
	bgpClient   *gobgp.BGPClient
	telemetry   *telemetry.TelemetryManager
	searchMode  bool
	searchQuery string
	filteredPeers []*gobgp.PeerInfo
}

// NewPeersModel creates a new advanced peers model
func NewPeersModel(bgpClient *gobgp.BGPClient) PeersModel {
	tm := telemetry.NewTelemetryManager(80)
	
	// Initialize sparklines for peer telemetry
	tm.AddSparkline("peer_traffic", "Traffic", "Mbps", 60)
	tm.AddSparkline("peer_routes", "Routes", "", 60)
	tm.AddSparkline("peer_flaps", "Flaps", "/min", 60)
	tm.AddSparkline("peer_latency", "Latency", "ms", 60)
	
	return PeersModel{
		bgpClient: bgpClient,
		telemetry: tm,
		peers:     make([]*gobgp.PeerInfo, 0),
		filteredPeers: make([]*gobgp.PeerInfo, 0),
	}
}

// Init initializes the peers model
func (m PeersModel) Init() tea.Cmd {
	return m.fetchPeersData()
}

// Update handles messages for the peers panel
func (m PeersModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		if m.searchMode {
			return m.handleSearchInput(msg)
		}
		
		switch msg.String() {
		case "up", "k":
			if m.selected > 0 {
				m.selected--
				if m.selected < m.offset {
					m.offset--
				}
			}
		case "down", "j":
			maxItems := len(m.getDisplayPeers()) - 1
			if m.selected < maxItems {
				m.selected++
				maxVisible := m.getMaxVisible()
				if m.selected >= m.offset+maxVisible {
					m.offset++
				}
			}
		case "home":
			m.selected = 0
			m.offset = 0
		case "end":
			peers := m.getDisplayPeers()
			m.selected = len(peers) - 1
			maxVisible := m.getMaxVisible()
			if len(peers) > maxVisible {
				m.offset = len(peers) - maxVisible
			}
		case "/":
			m.searchMode = true
			m.searchQuery = ""
		case "enter":
			// Show detailed view of selected peer
			return m, m.showPeerDetails()
		case "r":
			return m, m.fetchPeersData()
		case "esc":
			if m.searchMode {
				m.searchMode = false
				m.searchQuery = ""
				m.updateFilter()
			}
		}
		
	case PeersDataMsg:
		m.peers = msg.Peers
		m.updateFilter()
		m.updateTelemetry()
	}
	
	return m, nil
}

// handleSearchInput handles search input
func (m PeersModel) handleSearchInput(msg tea.KeyMsg) (PeersModel, tea.Cmd) {
	switch msg.String() {
	case "enter":
		m.searchMode = false
		m.updateFilter()
	case "esc":
		m.searchMode = false
		m.searchQuery = ""
		m.updateFilter()
	case "backspace":
		if len(m.searchQuery) > 0 {
			m.searchQuery = m.searchQuery[:len(m.searchQuery)-1]
			m.updateFilter()
		}
	default:
		if len(msg.String()) == 1 {
			m.searchQuery += msg.String()
			m.updateFilter()
		}
	}
	
	return m, nil
}

// View renders the peers panel
func (m PeersModel) View() string {
	if len(m.peers) == 0 {
		return m.renderLoading()
	}
	
	return m.renderPeers()
}

// SetSize sets the panel size
func (m *PeersModel) SetSize(width, height int) {
	m.width = width
	m.height = height
	m.telemetry = telemetry.NewTelemetryManager(width)
}

// UpdateData updates the panel data
func (m *PeersModel) UpdateData(data interface{}) {
	if peers, ok := data.([]*gobgp.PeerInfo); ok {
		m.peers = peers
		m.updateFilter()
		m.updateTelemetry()
	}
}

// renderLoading renders loading state
func (m PeersModel) renderLoading() string {
	style := lipgloss.NewStyle().
		Width(m.width).
		Height(m.height).
		Align(lipgloss.Center, lipgloss.Center).
		Foreground(lipgloss.Color("#666666"))
	
	return style.Render("Connecting to GoBGP daemon...\nLoading peer information...")
}

// renderPeers renders the peers content
func (m PeersModel) renderPeers() string {
	// Header
	headerStyle := lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color("#FAFAFA")).
		Background(lipgloss.Color("#7D56F4")).
		Padding(0, 1).
		Width(m.width - 2).
		Align(lipgloss.Center)
	
	title := "BGP Peers (Advanced)"
	if m.searchMode {
		title = fmt.Sprintf("BGP Peers - Search: %s_", m.searchQuery)
	} else if m.searchQuery != "" {
		title = fmt.Sprintf("BGP Peers - Filter: %s", m.searchQuery)
	}
	
	header := headerStyle.Render(title)
	
	// Main content area
	contentHeight := m.height - 8 // Reserve space for header, telemetry, and footer
	
	// Split into two sections: peer list (left) and details (right)
	leftWidth := m.width * 2 / 3
	rightWidth := m.width - leftWidth - 2
	
	peerList := m.renderPeerList(leftWidth, contentHeight)
	peerDetails := m.renderPeerDetails(rightWidth, contentHeight)
	
	mainContent := lipgloss.JoinHorizontal(
		lipgloss.Top,
		peerList,
		lipgloss.NewStyle().Width(1).Render("│"),
		peerDetails,
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

// renderPeerList renders the peer list table
func (m PeersModel) renderPeerList(width, height int) string {
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
	establishedStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("#00FF00")).Bold(true)
	idleStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("#FF0000")).Bold(true)
	connectStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("#FFFF00")).Bold(true)
	
	// Table header
	tableHeader := lipgloss.JoinHorizontal(lipgloss.Top,
		tableHeaderStyle.Width(8).Render("ASN"),
		tableHeaderStyle.Width(15).Render("REMOTE IP"),
		tableHeaderStyle.Width(12).Render("STATE"),
		tableHeaderStyle.Width(8).Render("UPTIME"),
		tableHeaderStyle.Width(6).Render("RX"),
		tableHeaderStyle.Width(6).Render("TX"),
		tableHeaderStyle.Width(6).Render("FLAPS"),
		tableHeaderStyle.Width(width-63).Render("DESCRIPTION"),
	)
	
	// Table rows
	var rows []string
	peers := m.getDisplayPeers()
	maxVisible := height - 2 // Account for header
	
	for i := m.offset; i < len(peers) && i < m.offset+maxVisible; i++ {
		peer := peers[i]
		
		// Format uptime
		uptime := m.formatDuration(peer.Uptime)
		
		// Status styling
		var statusRendered string
		switch peer.State {
		case "Established":
			statusRendered = establishedStyle.Render(peer.State)
		case "Idle":
			statusRendered = idleStyle.Render(peer.State)
		case "Connect":
			statusRendered = connectStyle.Render(peer.State)
		default:
			statusRendered = peer.State
		}
		
		var rowStyle lipgloss.Style
		if i == m.selected {
			rowStyle = selectedRowStyle
		} else {
			rowStyle = normalRowStyle
		}
		
		row := lipgloss.JoinHorizontal(lipgloss.Top,
			rowStyle.Width(8).Render(fmt.Sprintf("AS%d", peer.ASN)),
			rowStyle.Width(15).Render(peer.RemoteAddr),
			rowStyle.Width(12).Render(statusRendered),
			rowStyle.Width(8).Render(uptime),
			rowStyle.Width(6).Render(fmt.Sprintf("%d", peer.Received)),
			rowStyle.Width(6).Render(fmt.Sprintf("%d", peer.Advertised)),
			rowStyle.Width(6).Render(fmt.Sprintf("%d", peer.Flaps)),
			rowStyle.Width(width-63).Render(m.truncateString(peer.Description, width-65)),
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

// renderPeerDetails renders detailed information about the selected peer
func (m PeersModel) renderPeerDetails(width, height int) string {
	peers := m.getDisplayPeers()
	if len(peers) == 0 || m.selected >= len(peers) {
		return m.renderNoSelection(width, height)
	}
	
	peer := peers[m.selected]
	
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
	
	statusStyle := lipgloss.NewStyle().
		Bold(true)
	
	// Status color
	var statusColor lipgloss.Color
	switch peer.State {
	case "Established":
		statusColor = lipgloss.Color("#00FF00")
	case "Idle":
		statusColor = lipgloss.Color("#FF0000")
	case "Connect":
		statusColor = lipgloss.Color("#FFFF00")
	default:
		statusColor = lipgloss.Color("#666666")
	}
	statusStyle = statusStyle.Foreground(statusColor)
	
	// Build details
	var details strings.Builder
	
	details.WriteString(titleStyle.Render(fmt.Sprintf("AS%d Details", peer.ASN)))
	details.WriteString("\n")
	
	details.WriteString(labelStyle.Render("Router ID: "))
	details.WriteString(valueStyle.Render(peer.RouterID))
	details.WriteString("\n")
	
	details.WriteString(labelStyle.Render("Remote Address: "))
	details.WriteString(valueStyle.Render(peer.RemoteAddr))
	details.WriteString("\n")
	
	details.WriteString(labelStyle.Render("State: "))
	details.WriteString(statusStyle.Render(peer.State))
	details.WriteString("\n")
	
	details.WriteString(labelStyle.Render("Uptime: "))
	details.WriteString(valueStyle.Render(m.formatDuration(peer.Uptime)))
	details.WriteString("\n\n")
	
	details.WriteString(labelStyle.Render("Messages:"))
	details.WriteString("\n")
	details.WriteString(fmt.Sprintf("  Received: %s\n", valueStyle.Render(fmt.Sprintf("%d", peer.Received))))
	details.WriteString(fmt.Sprintf("  Accepted: %s\n", valueStyle.Render(fmt.Sprintf("%d", peer.Accepted))))
	details.WriteString(fmt.Sprintf("  Advertised: %s\n", valueStyle.Render(fmt.Sprintf("%d", peer.Advertised))))
	details.WriteString("\n")
	
	details.WriteString(labelStyle.Render("Statistics:"))
	details.WriteString("\n")
	details.WriteString(fmt.Sprintf("  Flaps: %s\n", valueStyle.Render(fmt.Sprintf("%d", peer.Flaps))))
	
	if peer.LastError != "" {
		details.WriteString("\n")
		details.WriteString(labelStyle.Render("Last Error: "))
		errorStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("#FF0000"))
		details.WriteString(errorStyle.Render(peer.LastError))
	}
	
	if peer.Description != "" {
		details.WriteString("\n\n")
		details.WriteString(labelStyle.Render("Description: "))
		details.WriteString(valueStyle.Render(peer.Description))
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

// renderNoSelection renders message when no peer is selected
func (m PeersModel) renderNoSelection(width, height int) string {
	style := lipgloss.NewStyle().
		Width(width).
		Height(height).
		Align(lipgloss.Center, lipgloss.Center).
		Foreground(lipgloss.Color("#666666"))
	
	return style.Render("Select a peer to view details")
}

// renderTelemetry renders the telemetry sparklines
func (m PeersModel) renderTelemetry() string {
	telemetryStyle := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("#666666")).
		Padding(0, 1).
		Margin(1, 0, 0, 0)
	
	telemetryContent := m.telemetry.RenderAll()
	
	return telemetryStyle.Render("Real-time Telemetry:\n" + telemetryContent)
}

// renderFooter renders the footer with help and statistics
func (m PeersModel) renderFooter() string {
	footerStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("#666666")).
		Padding(0, 1)
	
	peers := m.getDisplayPeers()
	established := 0
	for _, peer := range m.peers {
		if peer.State == "Established" {
			established++
		}
	}
	
	help := "↑↓: Navigate • Enter: Details • /: Search • r: Refresh • ESC: Cancel"
	stats := fmt.Sprintf("Total: %d | Established: %d | Showing: %d", 
		len(m.peers), established, len(peers))
	
	return footerStyle.Render(fmt.Sprintf("%s | %s", help, stats))
}

// Helper methods

func (m PeersModel) getDisplayPeers() []*gobgp.PeerInfo {
	if m.searchQuery != "" {
		return m.filteredPeers
	}
	return m.peers
}

func (m PeersModel) getMaxVisible() int {
	return m.height - 8 // Account for header, telemetry, footer
}

func (m *PeersModel) updateFilter() {
	if m.searchQuery == "" {
		m.filteredPeers = m.peers
		return
	}
	
	m.filteredPeers = make([]*gobgp.PeerInfo, 0)
	query := strings.ToLower(m.searchQuery)
	
	for _, peer := range m.peers {
		if strings.Contains(strings.ToLower(peer.RemoteAddr), query) ||
		   strings.Contains(strings.ToLower(peer.Description), query) ||
		   strings.Contains(strings.ToLower(fmt.Sprintf("AS%d", peer.ASN)), query) ||
		   strings.Contains(strings.ToLower(peer.State), query) {
			m.filteredPeers = append(m.filteredPeers, peer)
		}
	}
	
	// Reset selection if out of bounds
	if m.selected >= len(m.filteredPeers) {
		m.selected = 0
		m.offset = 0
	}
}

func (m *PeersModel) updateTelemetry() {
	// Update telemetry with current peer data
	totalTraffic := 0.0
	totalRoutes := 0
	totalFlaps := 0
	avgLatency := 0.0
	
	for _, peer := range m.peers {
		if peer.State == "Established" {
			totalRoutes += int(peer.Received)
			totalFlaps += int(peer.Flaps)
			// TODO: Calculate actual traffic and latency
		}
	}
	
	m.telemetry.UpdateData("peer_traffic", totalTraffic)
	m.telemetry.UpdateData("peer_routes", float64(totalRoutes))
	m.telemetry.UpdateData("peer_flaps", float64(totalFlaps))
	m.telemetry.UpdateData("peer_latency", avgLatency)
}

func (m PeersModel) formatDuration(d time.Duration) string {
	if d == 0 {
		return "0s"
	}
	
	if d < time.Minute {
		return fmt.Sprintf("%ds", int(d.Seconds()))
	} else if d < time.Hour {
		return fmt.Sprintf("%dm", int(d.Minutes()))
	} else if d < 24*time.Hour {
		return fmt.Sprintf("%dh", int(d.Hours()))
	} else {
		return fmt.Sprintf("%dd", int(d.Hours()/24))
	}
}

func (m PeersModel) truncateString(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	return s[:maxLen-3] + "..."
}

// Commands

func (m PeersModel) fetchPeersData() tea.Cmd {
	return tea.Cmd(func() tea.Msg {
		if m.bgpClient != nil {
			peers, err := m.bgpClient.GetPeers()
			if err != nil {
				// Fallback to mock data
				peers = m.bgpClient.GetMockPeers()
			}
			return PeersDataMsg{Peers: peers}
		}
		
		// Mock data for development
		mockClient := gobgp.MockBGPClient()
		peers := mockClient.GetMockPeers()
		return PeersDataMsg{Peers: peers}
	})
}

func (m PeersModel) showPeerDetails() tea.Cmd {
	return tea.Cmd(func() tea.Msg {
		// TODO: Implement detailed peer view
		return nil
	})
}

// Messages

type PeersDataMsg struct {
	Peers []*gobgp.PeerInfo
}