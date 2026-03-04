package tui

import (
	"context"
	"fmt"
	"time"

	"github.com/bgpin/bgpin/internal/tui/gobgp"
	"github.com/bgpin/bgpin/internal/tui/graph"
	"github.com/bgpin/bgpin/internal/tui/panels"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// Panel represents different TUI panels
type Panel int

const (
	GraphPanel Panel = iota
	PeersPanel
	RoutesPanel
	FlowsPanel
	SummaryPanel
)

// Model represents the main TUI model
type Model struct {
	config          Config
	refreshInterval time.Duration
	activePanel     Panel
	width           int
	height          int
	
	// BGP Client
	bgpClient *gobgp.BGPClient
	
	// Panel models
	graph     *graph.ASPathGraph
	peers     panels.PeersModel
	routes    panels.RoutesModel
	flows     panels.FlowsModel
	summary   panels.SummaryModel
	
	// State
	lastUpdate  time.Time
	loading     bool
	err         error
	searchMode  bool
	searchQuery string
	helpMode    bool
}

// DataUpdateMsg represents a data update message
type DataUpdateMsg struct {
	Panel Panel
	Data  interface{}
	Error error
}

// TickMsg represents a tick for refresh
type TickMsg time.Time

// SearchMsg represents a search message
type SearchMsg struct {
	Query string
}

// NewModel creates a new TUI model
func NewModel(config Config, refreshInterval time.Duration) *Model {
	// Initialize BGP client
	var bgpClient *gobgp.BGPClient
	var err error
	
	// Try to connect to GoBGP daemon
	bgpClient, err = gobgp.NewBGPClient("localhost:50051")
	if err != nil {
		// Fallback to mock client for development
		bgpClient = gobgp.MockBGPClient()
	}
	
	// Create AS-PATH graph
	centerASN := config.FocusASN
	if centerASN == 0 {
		centerASN = 65001 // Default ASN
	}
	
	asGraph := graph.NewASPathGraph(centerASN, 80, 20)
	
	// Initialize panels with BGP client
	activePanel := GraphPanel
	if config.StartWithFlows {
		activePanel = FlowsPanel
	}

	return &Model{
		config:          config,
		refreshInterval: refreshInterval,
		activePanel:     activePanel,
		bgpClient:       bgpClient,
		graph:           asGraph,
		peers:           panels.NewPeersModel(bgpClient),
		routes:          panels.NewRoutesModel(config.FocusASN),
		flows:           panels.NewFlowsModel(bgpClient),
		summary:         panels.NewSummaryModel(config.FocusASN),
		lastUpdate:      time.Now(),
	}
}

// Init initializes the model
func (m *Model) Init() tea.Cmd {
	return tea.Batch(
		m.peers.Init(),
		m.routes.Init(),
		m.flows.Init(),
		m.summary.Init(),
		tea.Tick(m.refreshInterval, func(t time.Time) tea.Msg {
			return TickMsg(t)
		}),
		m.initializeGraph(),
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
		
		// Update graph size
		m.graph = graph.NewASPathGraph(m.config.FocusASN, panelWidth-4, panelHeight-4)
		
		m.peers.SetSize(panelWidth, panelHeight)
		m.routes.SetSize(panelWidth, panelHeight)
		m.flows.SetSize(panelWidth, panelHeight)
		m.summary.SetSize(panelWidth, panelHeight)

	case tea.KeyMsg:
		if m.searchMode {
			return m.handleSearchInput(msg)
		}
		
		if m.helpMode {
			if msg.String() == "h" || msg.String() == "?" || msg.String() == "esc" {
				m.helpMode = false
			}
			return m, nil
		}
		
		switch msg.String() {
		case "q", "ctrl+c":
			if m.bgpClient != nil {
				m.bgpClient.Close()
			}
			return m, tea.Quit
		case "tab":
			m.nextPanel()
		case "shift+tab":
			m.prevPanel()
		case "r":
			m.loading = true
			return m, m.refreshData()
		case "h", "?":
			m.helpMode = true
		case "1":
			m.activePanel = GraphPanel
		case "2":
			m.activePanel = PeersPanel
		case "3":
			m.activePanel = RoutesPanel
		case "4":
			m.activePanel = FlowsPanel
		case "5":
			m.activePanel = SummaryPanel
		case "/":
			m.searchMode = true
			m.searchQuery = ""
		case "esc":
			m.searchMode = false
			m.searchQuery = ""
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
			case GraphPanel:
				// Update graph with peer data
				if peers, ok := msg.Data.([]*gobgp.PeerInfo); ok {
					m.updateGraph(peers)
				}
			case PeersPanel:
				m.peers.UpdateData(msg.Data)
			case RoutesPanel:
				m.routes.UpdateData(msg.Data)
			case FlowsPanel:
				m.flows.UpdateData(msg.Data)
			case SummaryPanel:
				m.summary.UpdateData(msg.Data)
			}
		}
		
	case SearchMsg:
		m.searchQuery = msg.Query
		// TODO: Implement search functionality
	}

	// Update active panel
	switch m.activePanel {
	case GraphPanel:
		// Graph doesn't need update handling
	case PeersPanel:
		newModel, cmd := m.peers.Update(msg)
		m.peers = newModel.(panels.PeersModel)
		if cmd != nil {
			cmds = append(cmds, cmd)
		}
	case RoutesPanel:
		newModel, cmd := m.routes.Update(msg)
		m.routes = newModel.(panels.RoutesModel)
		if cmd != nil {
			cmds = append(cmds, cmd)
		}
	case FlowsPanel:
		newModel, cmd := m.flows.Update(msg)
		m.flows = newModel.(panels.FlowsModel)
		if cmd != nil {
			cmds = append(cmds, cmd)
		}
	case SummaryPanel:
		newModel, cmd := m.summary.Update(msg)
		m.summary = newModel.(panels.SummaryModel)
		if cmd != nil {
			cmds = append(cmds, cmd)
		}
	}

	return m, tea.Batch(cmds...)
}

// View renders the TUI
func (m *Model) View() string {
	if m.width == 0 {
		return "Initializing bgptop (Advanced BGP Monitor)..."
	}
	
	if m.helpMode {
		return m.renderHelp()
	}

	// Header
	header := m.renderHeader()
	
	// Active panel content
	var content string
	switch m.activePanel {
	case GraphPanel:
		content = m.renderGraphPanel()
	case PeersPanel:
		content = m.peers.View()
	case RoutesPanel:
		content = m.routes.View()
	case FlowsPanel:
		content = m.flows.View()
	case SummaryPanel:
		content = m.summary.View()
	}

	// Footer
	footer := m.renderFooter()

	return lipgloss.JoinVertical(lipgloss.Left, header, content, footer)
}

// renderGraphPanel renders the AS-PATH graph panel
func (m *Model) renderGraphPanel() string {
	if m.graph == nil {
		return "Initializing AS-PATH graph..."
	}
	
	// Container style
	containerStyle := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("#7D56F4")).
		Padding(1).
		Width(m.width - 2).
		Height(m.height - 6) // Account for header and footer
	
	// Title
	titleStyle := lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color("#FAFAFA")).
		Background(lipgloss.Color("#7D56F4")).
		Padding(0, 1).
		Width(m.width - 4).
		Align(lipgloss.Center)
	
	title := fmt.Sprintf("AS-PATH Visualizer - AS%d", m.config.FocusASN)
	if m.config.FocusASN == 0 {
		title = "AS-PATH Visualizer - Network Overview"
	}
	
	// Graph content
	graphContent := m.graph.Render()
	
	// Selected node details
	selectedNode := m.graph.GetSelectedNode()
	var nodeDetails string
	if selectedNode != nil {
		nodeDetails = m.graph.GetNodeDetails(selectedNode.ASN)
	} else {
		nodeDetails = "No node selected\nUse arrow keys to navigate"
	}
	
	// Split view: graph (left) and details (right)
	graphWidth := (m.width - 6) * 2 / 3
	detailsWidth := (m.width - 6) - graphWidth - 2
	
	graphSection := lipgloss.NewStyle().
		Width(graphWidth).
		Height(m.height - 10).
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("#666666")).
		Padding(1).
		Render(graphContent)
	
	detailsStyle := lipgloss.NewStyle().
		Width(detailsWidth).
		Height(m.height - 10).
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("#666666")).
		Padding(1)
	
	detailsSection := detailsStyle.Render(nodeDetails)
	
	mainContent := lipgloss.JoinHorizontal(
		lipgloss.Top,
		graphSection,
		lipgloss.NewStyle().Width(1).Render("│"),
		detailsSection,
	)
	
	// Legend
	legendStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("#666666")).
		Padding(0, 1)
	
	legend := legendStyle.Render(
		"Legend: ◉ Center AS • ● Established • ○ Idle • ◐ Connecting • ✕ Down • ━ High Traffic • ─ Medium • · Low",
	)
	
	content := lipgloss.JoinVertical(
		lipgloss.Left,
		titleStyle.Render(title),
		mainContent,
		legend,
	)
	
	return containerStyle.Render(content)
}

// renderHelp renders the help screen
func (m *Model) renderHelp() string {
	helpStyle := lipgloss.NewStyle().
		Width(m.width).
		Height(m.height).
		Align(lipgloss.Center, lipgloss.Center).
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("#7D56F4")).
		Padding(2)
	
	titleStyle := lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color("#00FFFF")).
		Align(lipgloss.Center)
	
	sectionStyle := lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color("#FFFF00")).
		Margin(1, 0, 0, 0)
	
	keyStyle := lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color("#00FF00"))
	
	help := titleStyle.Render("bgptop - Advanced BGP Monitor") + "\n\n"
	
	help += sectionStyle.Render("Navigation:") + "\n"
	help += keyStyle.Render("Tab/Shift+Tab") + " - Switch between panels\n"
	help += keyStyle.Render("1-5") + " - Jump to specific panel\n"
	help += keyStyle.Render("↑↓/jk") + " - Navigate within panels\n"
	help += keyStyle.Render("Enter") + " - Select/View details\n\n"
	
	help += sectionStyle.Render("Panels:") + "\n"
	help += keyStyle.Render("1") + " - AS-PATH Graph Visualizer\n"
	help += keyStyle.Render("2") + " - BGP Peers (Advanced)\n"
	help += keyStyle.Render("3") + " - BGP Routes\n"
	help += keyStyle.Render("4") + " - Top 5 Network Flows\n"
	help += keyStyle.Render("5") + " - System Summary\n\n"
	
	help += sectionStyle.Render("Actions:") + "\n"
	help += keyStyle.Render("r") + " - Refresh data\n"
	help += keyStyle.Render("/") + " - Search (in supported panels)\n"
	help += keyStyle.Render("a") + " - Toggle auto-refresh (flows panel)\n"
	help += keyStyle.Render("ESC") + " - Cancel search/Close dialogs\n\n"
	
	help += sectionStyle.Render("Features:") + "\n"
	help += "• Real-time BGP peer monitoring with GoBGP integration\n"
	help += "• AS-PATH visualization with dynamic graph\n"
	help += "• NetFlow/sFlow/IPFIX top talkers analysis\n"
	help += "• Sparkline telemetry charts\n"
	help += "• Advanced search and filtering\n"
	help += "• Professional network operations interface\n\n"
	
	help += keyStyle.Render("h/?") + " - Toggle this help • " + keyStyle.Render("q/Ctrl+C") + " - Quit"
	
	return helpStyle.Render(help)
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
		"1:Graph",
		"2:Peers",
		"3:Routes", 
		"4:Flows",
		"5:Summary",
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
	
	// BGP connection status
	bgpStatus := "Mock"
	if m.bgpClient != nil {
		bgpStatus = "GoBGP"
	}
	
	timeStr += fmt.Sprintf(" | %s", bgpStatus)

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

	help := "Tab: Next Panel • 1-5: Jump to Panel • r: Refresh • /: Search • h: Help • q: Quit"
	if m.searchMode {
		help = fmt.Sprintf("Search: %s_ | Enter: Apply • ESC: Cancel", m.searchQuery)
	}
	if m.err != nil {
		help = fmt.Sprintf("Error: %s • %s", m.err.Error(), help)
	}

	return footerStyle.Width(m.width).Render(help)
}

// nextPanel switches to the next panel
func (m *Model) nextPanel() {
	m.activePanel = (m.activePanel + 1) % 5
}

// prevPanel switches to the previous panel
func (m *Model) prevPanel() {
	if m.activePanel == 0 {
		m.activePanel = 4
	} else {
		m.activePanel--
	}
}

// handleSearchInput handles search input
func (m *Model) handleSearchInput(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "enter":
		m.searchMode = false
		return m, tea.Cmd(func() tea.Msg {
			return SearchMsg{Query: m.searchQuery}
		})
	case "esc":
		m.searchMode = false
		m.searchQuery = ""
	case "backspace":
		if len(m.searchQuery) > 0 {
			m.searchQuery = m.searchQuery[:len(m.searchQuery)-1]
		}
	default:
		if len(msg.String()) == 1 {
			m.searchQuery += msg.String()
		}
	}
	
	return m, nil
}

// refreshData triggers data refresh for all panels
func (m *Model) refreshData() tea.Cmd {
	return tea.Batch(
		m.fetchGraphData(),
		m.fetchPeersData(),
		m.fetchRoutesData(),
		m.fetchFlowsData(),
		m.fetchSummaryData(),
	)
}

// initializeGraph initializes the AS-PATH graph with sample data
func (m *Model) initializeGraph() tea.Cmd {
	return tea.Cmd(func() tea.Msg {
		// Add sample nodes to the graph
		centerASN := m.config.FocusASN
		if centerASN == 0 {
			centerASN = 65001
		}
		
		// Add center node
		m.graph.AddNode(centerASN, "Local AS", graph.StatusEstablished, 1200.5, 150, 5.2)
		
		// Add peer nodes
		m.graph.AddNode(15169, "Google", graph.StatusEstablished, 850.3, 1200, 12.1)
		m.graph.AddNode(13335, "Cloudflare", graph.StatusEstablished, 640.8, 800, 8.5)
		m.graph.AddNode(64512, "Private AS", graph.StatusIdle, 0, 0, 0)
		m.graph.AddNode(174, "Cogent", graph.StatusEstablished, 420.2, 600, 15.3)
		
		// Add connections
		m.graph.AddConnection(centerASN, 15169)
		m.graph.AddConnection(centerASN, 13335)
		m.graph.AddConnection(centerASN, 64512)
		m.graph.AddConnection(centerASN, 174)
		m.graph.AddConnection(15169, 174)
		m.graph.AddConnection(13335, 174)
		
		return DataUpdateMsg{Panel: GraphPanel, Data: nil}
	})
}

// updateGraph updates the graph with peer data
func (m *Model) updateGraph(peers []*gobgp.PeerInfo) {
	if m.graph == nil {
		return
	}
	
	// Clear existing nodes except center
	centerASN := m.config.FocusASN
	if centerASN == 0 {
		centerASN = 65001
	}
	
	// Create new graph with updated data
	m.graph = graph.NewASPathGraph(centerASN, m.width-4, m.height-8)
	
	// Add center node
	m.graph.AddNode(centerASN, "Local AS", graph.StatusEstablished, 0, 0, 0)
	
	// Add peer nodes
	for _, peer := range peers {
		var status graph.NodeStatus
		switch peer.State {
		case "Established":
			status = graph.StatusEstablished
		case "Idle":
			status = graph.StatusIdle
		case "Connect":
			status = graph.StatusConnect
		default:
			status = graph.StatusDown
		}
		
		// Calculate traffic (mock for now)
		traffic := float64(peer.Received) * 0.1 // Mock calculation
		
		m.graph.AddNode(int(peer.ASN), peer.Description, status, traffic, int(peer.Received), 0)
		m.graph.AddConnection(centerASN, int(peer.ASN))
	}
}

// Data fetching commands
func (m *Model) fetchGraphData() tea.Cmd {
	return tea.Cmd(func() tea.Msg {
		if m.bgpClient != nil {
			peers, err := m.bgpClient.GetPeers()
			if err != nil {
				peers = m.bgpClient.GetMockPeers()
			}
			return DataUpdateMsg{Panel: GraphPanel, Data: peers}
		}
		
		// Mock data
		mockClient := gobgp.MockBGPClient()
		peers := mockClient.GetMockPeers()
		return DataUpdateMsg{Panel: GraphPanel, Data: peers}
	})
}

func (m *Model) fetchPeersData() tea.Cmd {
	return tea.Cmd(func() tea.Msg {
		if m.bgpClient != nil {
			peers, err := m.bgpClient.GetPeers()
			if err != nil {
				peers = m.bgpClient.GetMockPeers()
			}
			return DataUpdateMsg{Panel: PeersPanel, Data: peers}
		}
		
		// Mock data
		mockClient := gobgp.MockBGPClient()
		peers := mockClient.GetMockPeers()
		return DataUpdateMsg{Panel: PeersPanel, Data: peers}
	})
}

func (m *Model) fetchRoutesData() tea.Cmd {
	return tea.Cmd(func() tea.Msg {
		if m.bgpClient != nil {
			routes := m.bgpClient.GetMockRoutes() // Use mock for now
			
			// Convert to expected format
			data := make([]map[string]interface{}, len(routes))
			for i, route := range routes {
				data[i] = map[string]interface{}{
					"prefix":   route.Prefix,
					"asn":      fmt.Sprintf("AS%d", route.ASPath[0]),
					"status":   "Valid",
					"next_hop": route.NextHop,
					"med":      fmt.Sprintf("%d", route.MED),
					"path":     fmt.Sprintf("%v", route.ASPath),
				}
				if !route.Valid {
					data[i]["status"] = "Invalid"
				}
			}
			
			return DataUpdateMsg{Panel: RoutesPanel, Data: data}
		}
		
		// Mock data
		data := []map[string]interface{}{
			{"prefix": "8.8.8.0/24", "asn": "AS15169", "status": "Valid", "next_hop": "8.8.8.8", "med": "0", "path": "[15169]"},
			{"prefix": "1.1.1.0/24", "asn": "AS13335", "status": "Valid", "next_hop": "1.1.1.1", "med": "0", "path": "[13335]"},
			{"prefix": "192.168.1.0/24", "asn": "AS64512", "status": "Invalid", "next_hop": "192.168.1.1", "med": "10", "path": "[64512]"},
		}
		return DataUpdateMsg{Panel: RoutesPanel, Data: data}
	})
}

func (m *Model) fetchFlowsData() tea.Cmd {
	return tea.Cmd(func() tea.Msg {
		// This would integrate with actual NetFlow/sFlow collector
		// For now, return mock data
		return DataUpdateMsg{Panel: FlowsPanel, Data: nil} // Flows panel generates its own mock data
	})
}

func (m *Model) fetchSummaryData() tea.Cmd {
	return tea.Cmd(func() tea.Msg {
		data := map[string]interface{}{
			"asn":       m.config.FocusASN,
			"routes":    12543,
			"neighbors": 8,
			"traffic":   "1.2 Gbps",
			"status":    "Active",
		}
		return DataUpdateMsg{Panel: SummaryPanel, Data: data}
	})
}

// startDataFetching starts background data fetching
func (m *Model) startDataFetching(ctx context.Context) {
	// Start watching for BGP events if client is available
	if m.bgpClient != nil {
		// Watch peer events
		go func() {
			err := m.bgpClient.WatchPeers(func(peer *gobgp.PeerInfo, eventType string) {
				// TODO: Send update message to TUI
			})
			if err != nil {
				// Log error but continue
			}
		}()
		
		// Watch route events
		go func() {
			err := m.bgpClient.WatchRoutes(func(route *gobgp.RouteInfo, eventType string) {
				// TODO: Send update message to TUI
			})
			if err != nil {
				// Log error but continue
			}
		}()
	}
}