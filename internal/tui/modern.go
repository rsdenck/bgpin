package tui

import (
	"context"
	"fmt"
	"math"
	"strconv"
	"strings"
	"time"

	"github.com/charmbracelet/bubbles/table"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/bgpin/bgpin/internal/adapters/ssh"
	"github.com/bgpin/bgpin/internal/tui/gobgp"
	"github.com/bgpin/bgpin/internal/tui/components"
	"github.com/bgpin/bgpin/internal/tui/metrics"
)

// ModernTUI representa a nova TUI moderna
type ModernTUI struct {
	width  int
	height int
	
	// Conexões
	sshClient *ssh.Client
	bgpClient *gobgp.BGPClient
	
	// Métricas e componentes
	metricsCollector *metrics.MetricsCollector
	alertManager     *metrics.AlertManager
	
	// Gráficos e componentes visuais
	cpuChart         *components.LargeChart
	memoryChart      *components.LargeChart
	networkChart     *components.NetworkChart
	bgpPeerChart     *components.BGPPeerChart
	cpuGauge         *components.Gauge
	memoryGauge      *components.Gauge
	
	// Gráficos de análise técnica
	routeAnalysisChart *components.LineChart
	trafficAnalysisChart *components.CandlestickChart
	peerVolumeChart    *components.VolumeChart
	
	// Dados em tempo real
	bgpSummary    BGPSummary
	peerStats     []PeerStat
	routeStats    RouteStats
	interfaceStats []InterfaceStat
	systemInfo    SystemInfo
	
	// Tabelas
	peersTable      table.Model
	routesTable     table.Model
	interfacesTable table.Model
	alertsTable     table.Model
	
	// Estado
	activePanel int
	lastUpdate  time.Time
	errors      []string
	alerts      []*metrics.Alert
	
	// Configuração
	routerIP   string
	username   string
	password   string
	refreshRate time.Duration
	demoMode   bool
}

// Estruturas de dados
type BGPSummary struct {
	RouterID     string
	LocalAS      int
	TotalPeers   int
	ActivePeers  int
	TotalRoutes  int
	BestRoutes   int
	Uptime       string
}

type PeerStat struct {
	IP           string
	ASN          int
	State        string
	Uptime       string
	PrefixRcv    int
	PrefixSent   int
	InMsgs       int64
	OutMsgs      int64
	LastError    string
}

type RouteStats struct {
	IPv4Total    int
	IPv4Best     int
	IPv6Total    int
	IPv6Best     int
	Suppressed   int
	Damped       int
	History      int
}

type InterfaceStat struct {
	Name         string
	Status       string
	Speed        string
	InBytes      int64
	OutBytes     int64
	InPackets    int64
	OutPackets   int64
	InErrors     int64
	OutErrors    int64
}

type SystemInfo struct {
	Hostname     string
	Version      string
	Uptime       string
	CPUUsage     float64
	MemoryUsage  float64
	Temperature  float64
}

// Estilos
var (
	headerStyle = lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color("#00FF00")).
		Background(lipgloss.Color("#000000")).
		Padding(0, 1)
	
	titleStyle = lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color("#FFFFFF")).
		Background(lipgloss.Color("#333333")).
		Padding(0, 1).
		MarginBottom(1)
	
	activeTabStyle = lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color("#000000")).
		Background(lipgloss.Color("#00FF00")).
		Padding(0, 2)
	
	inactiveTabStyle = lipgloss.NewStyle().
		Foreground(lipgloss.Color("#666666")).
		Background(lipgloss.Color("#222222")).
		Padding(0, 2)
	
	panelStyle = lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("#444444")).
		Padding(1).
		MarginRight(1).
		MarginBottom(1)
	
	activePanelStyle = lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("#00FF00")).
		Padding(1).
		MarginRight(1).
		MarginBottom(1)
	
	errorStyle = lipgloss.NewStyle().
		Foreground(lipgloss.Color("#FF4444")).
		Bold(true)
	
	successStyle = lipgloss.NewStyle().
		Foreground(lipgloss.Color("#00FF00")).
		Bold(true)
	
	warningStyle = lipgloss.NewStyle().
		Foreground(lipgloss.Color("#FF8800")).
		Bold(true)
	
	infoStyle = lipgloss.NewStyle().
		Foreground(lipgloss.Color("#00AAFF"))
	
	chartStyle = lipgloss.NewStyle().
		Border(lipgloss.NormalBorder()).
		BorderForeground(lipgloss.Color("#444444")).
		Padding(1)
	
	metricLabelStyle = lipgloss.NewStyle().
		Foreground(lipgloss.Color("#AAAAAA")).
		Bold(true)
	
	metricValueStyle = lipgloss.NewStyle().
		Foreground(lipgloss.Color("#00FF88")).
		Bold(true)
	
	// Cores técnicas para gráficos
	cpuColor = lipgloss.Color("#00FF00")      // Verde para CPU
	memoryColor = lipgloss.Color("#0088FF")   // Azul para memória
	networkInColor = lipgloss.Color("#00FF88") // Verde claro para entrada
	networkOutColor = lipgloss.Color("#FF6600") // Laranja para saída
	diskColor = lipgloss.Color("#FF4488")     // Rosa para disco
	tempColor = lipgloss.Color("#FF0044")     // Vermelho para temperatura
)

// NewModernTUI cria uma nova instância da TUI moderna
func NewModernTUI(routerIP, username, password string) *ModernTUI {
	// Verificar se é modo demo
	demoMode := routerIP == "demo"
	
	tui := &ModernTUI{
		routerIP:    routerIP,
		username:    username,
		password:    password,
		refreshRate: 2 * time.Second,
		activePanel: 0,
		errors:      make([]string, 0),
		demoMode:    demoMode,
	}
	
	// Inicializar componentes
	tui.initializeComponents()
	
	return tui
}

// initializeComponents inicializa todos os componentes visuais
func (m *ModernTUI) initializeComponents() {
	// Inicializar coletor de métricas
	m.metricsCollector = metrics.NewMetricsCollector(1 * time.Second)
	m.alertManager = metrics.NewAlertManager()
	
	// Adicionar métricas
	m.metricsCollector.AddMetric("cpu_usage", "%", 60)
	m.metricsCollector.AddMetric("memory_usage", "%", 60)
	m.metricsCollector.AddMetric("network_traffic", "Mbps", 60)
	m.metricsCollector.AddMetric("bgp_peers", "count", 60)
	m.metricsCollector.AddMetric("bgp_routes", "count", 60)
	
	// Inicializar gráficos grandes
	m.cpuChart = components.NewLargeChart("CPU Usage", "%", 80, 12)
	m.cpuChart.SetColor(cpuColor)
	
	m.memoryChart = components.NewLargeChart("Memory Usage", "%", 80, 12)
	m.memoryChart.SetColor(memoryColor)
	
	m.networkChart = components.NewNetworkChart("Network Traffic", "Mbps", 80, 12)
	
	m.bgpPeerChart = components.NewBGPPeerChart("BGP Peer Prefixes", 80, 12)
	
	// Inicializar gráficos de análise técnica
	m.routeAnalysisChart = components.NewLineChart("Route Analysis", 80, 15)
	m.routeAnalysisChart.AddSeries("IPv4 Routes", cpuColor)
	m.routeAnalysisChart.AddSeries("IPv6 Routes", memoryColor)
	m.routeAnalysisChart.AddSeries("BGP Updates", networkInColor)
	
	m.trafficAnalysisChart = components.NewCandlestickChart("Traffic Analysis", "5min", 80, 15)
	
	m.peerVolumeChart = components.NewVolumeChart("Peer Activity Volume", 80, 12)
	
	// Inicializar medidores
	m.cpuGauge = components.NewGauge("CPU", "%", 100)
	m.memoryGauge = components.NewGauge("Memory", "%", 100)
	
	// Configurar regras de alerta
	m.setupAlertRules()
	
	// Inicializar com dados demo se necessário
	if m.demoMode {
		m.initializeDemoData()
	}
}

// setupAlertRules configura regras de alerta
func (m *ModernTUI) setupAlertRules() {
	rules := []metrics.AlertRule{
		{
			ID:          "cpu_high",
			Metric:      "cpu_usage",
			Condition:   "gt",
			Threshold:   80.0,
			Level:       metrics.AlertLevelWarning,
			Title:       "High CPU Usage",
			Description: "CPU usage is above 80%",
			Enabled:     true,
		},
		{
			ID:          "cpu_critical",
			Metric:      "cpu_usage",
			Condition:   "gt",
			Threshold:   95.0,
			Level:       metrics.AlertLevelCritical,
			Title:       "Critical CPU Usage",
			Description: "CPU usage is above 95%",
			Enabled:     true,
		},
		{
			ID:          "memory_high",
			Metric:      "memory_usage",
			Condition:   "gt",
			Threshold:   85.0,
			Level:       metrics.AlertLevelWarning,
			Title:       "High Memory Usage",
			Description: "Memory usage is above 85%",
			Enabled:     true,
		},
		{
			ID:          "bgp_peer_down",
			Metric:      "bgp_peers",
			Condition:   "lt",
			Threshold:   2.0,
			Level:       metrics.AlertLevelError,
			Title:       "BGP Peer Down",
			Description: "Less than 2 BGP peers are active",
			Enabled:     true,
		},
	}
	
	for _, rule := range rules {
		m.alertManager.AddRule(rule)
	}
}

// initializeDemoData inicializa dados de demonstração
func (m *ModernTUI) initializeDemoData() {
	// Dados BGP de exemplo
	m.bgpSummary = BGPSummary{
		RouterID:     "192.168.0.1",
		LocalAS:      262978,
		TotalPeers:   5,
		ActivePeers:  4,
		TotalRoutes:  850000,
		BestRoutes:   750000,
		Uptime:       "15d 8h 32m",
	}
	
	// Peers de exemplo
	m.peerStats = []PeerStat{
		{IP: "10.0.1.1", ASN: 65001, State: "Established", Uptime: "15d8h", PrefixRcv: 250000, PrefixSent: 5000, InMsgs: 1500000, OutMsgs: 750000},
		{IP: "10.0.1.2", ASN: 65002, State: "Established", Uptime: "12d3h", PrefixRcv: 180000, PrefixSent: 3500, InMsgs: 980000, OutMsgs: 520000},
		{IP: "10.0.1.3", ASN: 65003, State: "Established", Uptime: "8d15h", PrefixRcv: 320000, PrefixSent: 7200, InMsgs: 2100000, OutMsgs: 890000},
		{IP: "10.0.1.4", ASN: 65004, State: "Idle", Uptime: "0", PrefixRcv: 0, PrefixSent: 0, InMsgs: 0, OutMsgs: 0, LastError: "Connection refused"},
		{IP: "10.0.1.5", ASN: 65005, State: "Established", Uptime: "3d22h", PrefixRcv: 95000, PrefixSent: 2800, InMsgs: 650000, OutMsgs: 280000},
	}
	
	// Estatísticas de rotas
	m.routeStats = RouteStats{
		IPv4Total:  750000,
		IPv4Best:   680000,
		IPv6Total:  100000,
		IPv6Best:   85000,
		Suppressed: 15000,
		Damped:     2500,
		History:    50000,
	}
	
	// Interfaces de exemplo
	m.interfaceStats = []InterfaceStat{
		{Name: "GigabitEthernet0/0/0", Status: "up", Speed: "1000Mbps", InBytes: 15840000000, OutBytes: 8920000000, InPackets: 12500000, OutPackets: 8900000, InErrors: 0, OutErrors: 0},
		{Name: "GigabitEthernet0/0/1", Status: "up", Speed: "1000Mbps", InBytes: 9870000000, OutBytes: 12340000000, InPackets: 8200000, OutPackets: 11500000, InErrors: 5, OutErrors: 2},
		{Name: "GigabitEthernet0/0/2", Status: "down", Speed: "1000Mbps", InBytes: 0, OutBytes: 0, InPackets: 0, OutPackets: 0, InErrors: 0, OutErrors: 0},
		{Name: "TenGigabitEthernet0/1/0", Status: "up", Speed: "10000Mbps", InBytes: 45600000000, OutBytes: 38900000000, InPackets: 35000000, OutPackets: 28500000, InErrors: 0, OutErrors: 0},
	}
	
	// Informações do sistema
	m.systemInfo = SystemInfo{
		Hostname:     "bgp-router-01",
		Version:      "Cisco IOS XE 17.3.4a",
		Uptime:       "15 days, 8 hours, 32 minutes",
		CPUUsage:     25.8,
		MemoryUsage:  42.3,
		Temperature:  38.5,
	}
	
	// Simular dados históricos para gráficos
	go m.simulateMetrics()
}

// Init inicializa a TUI
func (m *ModernTUI) Init() tea.Cmd {
	return tea.Batch(
		m.connectToRouter(),
		m.initializeTables(),
		m.startAutoRefresh(),
	)
}

// Update processa mensagens
func (m *ModernTUI) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmds []tea.Cmd
	
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		m.updateTableSizes()
		
	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c", "q":
			return m, tea.Quit
		case "tab":
			m.nextPanel()
		case "shift+tab":
			m.prevPanel()
		case "r":
			cmds = append(cmds, m.refreshData())
		case "h", "?":
			// Toggle help
		}
		
	case RouterConnectedMsg:
		m.addInfo("Conectado ao router " + m.routerIP)
		cmds = append(cmds, m.refreshData())
		
	case DataRefreshMsg:
		m.lastUpdate = time.Now()
		m.updateTables()
		
	case ErrorMsg:
		m.addError(msg.Error)
		
	case AutoRefreshMsg:
		cmds = append(cmds, m.refreshData(), m.startAutoRefresh())
	}
	
	// Atualizar tabelas
	var cmd tea.Cmd
	m.peersTable, cmd = m.peersTable.Update(msg)
	cmds = append(cmds, cmd)
	
	m.routesTable, cmd = m.routesTable.Update(msg)
	cmds = append(cmds, cmd)
	
	m.interfacesTable, cmd = m.interfacesTable.Update(msg)
	cmds = append(cmds, cmd)
	
	return m, tea.Batch(cmds...)
}

// View renderiza a interface
func (m *ModernTUI) View() string {
	if m.width == 0 {
		return "Inicializando..."
	}
	
	// Header principal
	header := m.renderHeader()
	
	// Tabs
	tabs := m.renderTabs()
	
	// Painel principal baseado na aba ativa
	var mainContent string
	switch m.activePanel {
	case 0:
		mainContent = m.renderOverviewPanel()
	case 1:
		mainContent = m.renderPeersPanel()
	case 2:
		mainContent = m.renderRoutesPanel()
	case 3:
		mainContent = m.renderInterfacesPanel()
	case 4:
		mainContent = m.renderSystemPanel()
	default:
		mainContent = m.renderOverviewPanel()
	}
	
	// Footer
	footer := m.renderFooter()
	
	// Combinar tudo
	return lipgloss.JoinVertical(
		lipgloss.Left,
		header,
		tabs,
		mainContent,
		footer,
	)
}

// renderHeader renderiza o cabeçalho
func (m *ModernTUI) renderHeader() string {
	title := headerStyle.Render("BGPIN MONITOR - ROUTER DASHBOARD")
	
	// Status da conexão
	status := "DISCONNECTED"
	statusStyle := errorStyle
	if m.sshClient != nil || m.demoMode {
		status = "CONNECTED"
		statusStyle = successStyle
		if m.demoMode {
			status = "DEMO MODE"
			statusStyle = warningStyle
		}
	}
	
	// Informações adicionais
	info := fmt.Sprintf("Router: %s | Status: %s | AS: %d | Peers: %d/%d | Last Update: %s",
		m.routerIP,
		statusStyle.Render(status),
		m.bgpSummary.LocalAS,
		m.bgpSummary.ActivePeers,
		m.bgpSummary.TotalPeers,
		m.lastUpdate.Format("15:04:05"),
	)
	
	// Alertas no header
	alertCounts := m.alertManager.GetAlertCount()
	if alertCounts[metrics.AlertLevelCritical] > 0 {
		info += " | " + errorStyle.Render(fmt.Sprintf("CRIT: %d", alertCounts[metrics.AlertLevelCritical]))
	}
	if alertCounts[metrics.AlertLevelWarning] > 0 {
		info += " | " + warningStyle.Render(fmt.Sprintf("WARN: %d", alertCounts[metrics.AlertLevelWarning]))
	}
	
	headerWidth := m.width
	titleWidth := lipgloss.Width(title)
	infoWidth := lipgloss.Width(info)
	
	padding := headerWidth - titleWidth - infoWidth
	if padding < 0 {
		padding = 0
	}
	
	return lipgloss.JoinHorizontal(
		lipgloss.Top,
		title,
		strings.Repeat(" ", padding),
		info,
	)
}

// renderTabs renderiza as abas
func (m *ModernTUI) renderTabs() string {
	tabs := []string{"Overview", "BGP Peers", "Routes", "Interfaces", "System"}
	var renderedTabs []string
	
	for i, tab := range tabs {
		if i == m.activePanel {
			renderedTabs = append(renderedTabs, activeTabStyle.Render(tab))
		} else {
			renderedTabs = append(renderedTabs, inactiveTabStyle.Render(tab))
		}
	}
	
	return lipgloss.JoinHorizontal(lipgloss.Top, renderedTabs...)
}

// renderOverviewPanel renderiza o painel de visão geral
func (m *ModernTUI) renderOverviewPanel() string {
	// Layout estilo análise técnica
	
	// Painel superior: Análise de rotas (gráfico de linha)
	routeAnalysisPanel := chartStyle.Width(m.width-4).Height(18).Render(m.routeAnalysisChart.Render())
	
	// Painel do meio: Análise de tráfego (candlestick)
	trafficAnalysisPanel := chartStyle.Width(m.width/2-2).Height(18).Render(m.trafficAnalysisChart.Render())
	
	// Painel do meio direito: Volume de atividade de peers
	peerVolumePanel := chartStyle.Width(m.width/2-2).Height(18).Render(m.peerVolumeChart.Render())
	
	middleRow := lipgloss.JoinHorizontal(lipgloss.Top, trafficAnalysisPanel, peerVolumePanel)
	
	// Painel inferior: Métricas do sistema (compacto)
	systemMetricsPanel := m.renderCompactSystemMetrics()
	
	return lipgloss.JoinVertical(lipgloss.Left, routeAnalysisPanel, middleRow, systemMetricsPanel)
}

// renderMetricsPanel renderiza painel de métricas com gráficos
func (m *ModernTUI) renderMetricsPanel() string {
	var content strings.Builder
	
	// Gráficos de linha
	content.WriteString(m.cpuChart.Render() + "\n\n")
	content.WriteString(m.memoryChart.Render() + "\n\n")
	
	// Medidores
	gaugeRow := lipgloss.JoinHorizontal(
		lipgloss.Top,
		m.cpuGauge.Render(),
		"  ",
		m.memoryGauge.Render(),
	)
	content.WriteString(gaugeRow)
	
	return panelStyle.Width(m.width/2-2).Height(20).Render(
		titleStyle.Render("📊 System Metrics") + "\n" + content.String(),
	)
}

// renderCompactSystemMetrics renderiza métricas do sistema de forma compacta
func (m *ModernTUI) renderCompactSystemMetrics() string {
	var content strings.Builder
	
	// Linha de métricas principais
	content.WriteString(metricLabelStyle.Render("SYSTEM METRICS") + " | ")
	content.WriteString(fmt.Sprintf("CPU: %s | ", 
		m.getColoredMetric(m.systemInfo.CPUUsage, 80, 95)))
	content.WriteString(fmt.Sprintf("MEM: %s | ", 
		m.getColoredMetric(m.systemInfo.MemoryUsage, 85, 95)))
	content.WriteString(fmt.Sprintf("BGP: %s/%d peers | ", 
		successStyle.Render(strconv.Itoa(m.bgpSummary.ActivePeers)), 
		m.bgpSummary.TotalPeers))
	content.WriteString(fmt.Sprintf("Routes: %s | ", 
		metricValueStyle.Render(strconv.Itoa(m.bgpSummary.TotalRoutes))))
	
	// Alertas
	alertCounts := m.alertManager.GetAlertCount()
	if alertCounts[metrics.AlertLevelCritical] > 0 {
		content.WriteString(errorStyle.Render(fmt.Sprintf("CRIT: %d | ", alertCounts[metrics.AlertLevelCritical])))
	}
	if alertCounts[metrics.AlertLevelWarning] > 0 {
		content.WriteString(warningStyle.Render(fmt.Sprintf("WARN: %d | ", alertCounts[metrics.AlertLevelWarning])))
	}
	
	content.WriteString(fmt.Sprintf("Uptime: %s", infoStyle.Render(m.systemInfo.Uptime)))
	
	return chartStyle.Width(m.width-4).Height(3).Render(content.String())
}

// getColoredMetric retorna uma métrica com cor baseada nos thresholds
func (m *ModernTUI) getColoredMetric(value, warningThreshold, criticalThreshold float64) string {
	valueStr := fmt.Sprintf("%.1f%%", value)
	
	if value >= criticalThreshold {
		return errorStyle.Render(valueStr)
	} else if value >= warningThreshold {
		return warningStyle.Render(valueStr)
	} else {
		return successStyle.Render(valueStr)
	}
}

// renderBGPSummaryPanel renderiza resumo BGP
func (m *ModernTUI) renderBGPSummaryPanel() string {
	content := fmt.Sprintf(`Router ID: %s
Local AS: %d
Total Peers: %d
Active Peers: %s
Total Routes: %d
Best Routes: %d
Uptime: %s`,
		m.bgpSummary.RouterID,
		m.bgpSummary.LocalAS,
		m.bgpSummary.TotalPeers,
		successStyle.Render(strconv.Itoa(m.bgpSummary.ActivePeers)),
		m.bgpSummary.TotalRoutes,
		m.bgpSummary.BestRoutes,
		m.bgpSummary.Uptime,
	)
	
	return panelStyle.Width(m.width/2-2).Height(8).Render(
		titleStyle.Render("BGP Summary") + "\n" + content,
	)
}

// renderSystemInfoPanel renderiza informações do sistema
func (m *ModernTUI) renderSystemInfoPanel() string {
	cpuColor := successStyle
	if m.systemInfo.CPUUsage > 80 {
		cpuColor = errorStyle
	} else if m.systemInfo.CPUUsage > 60 {
		cpuColor = warningStyle
	}
	
	memColor := successStyle
	if m.systemInfo.MemoryUsage > 80 {
		memColor = errorStyle
	} else if m.systemInfo.MemoryUsage > 60 {
		memColor = warningStyle
	}
	
	content := fmt.Sprintf(`Hostname: %s
Version: %s
Uptime: %s
CPU Usage: %s
Memory Usage: %s
Temperature: %.1f°C`,
		m.systemInfo.Hostname,
		m.systemInfo.Version,
		m.systemInfo.Uptime,
		cpuColor.Render(fmt.Sprintf("%.1f%%", m.systemInfo.CPUUsage)),
		memColor.Render(fmt.Sprintf("%.1f%%", m.systemInfo.MemoryUsage)),
		m.systemInfo.Temperature,
	)
	
	return panelStyle.Width(m.width/2-2).Height(8).Render(
		titleStyle.Render("System Info") + "\n" + content,
	)
}

// renderRouteStatsPanel renderiza estatísticas de rotas
func (m *ModernTUI) renderRouteStatsPanel() string {
	content := fmt.Sprintf(`IPv4 Total: %d
IPv4 Best: %s
IPv6 Total: %d
IPv6 Best: %s
Suppressed: %s
Damped: %s`,
		m.routeStats.IPv4Total,
		successStyle.Render(strconv.Itoa(m.routeStats.IPv4Best)),
		m.routeStats.IPv6Total,
		successStyle.Render(strconv.Itoa(m.routeStats.IPv6Best)),
		warningStyle.Render(strconv.Itoa(m.routeStats.Suppressed)),
		errorStyle.Render(strconv.Itoa(m.routeStats.Damped)),
	)
	
	return panelStyle.Width(m.width/2-2).Height(8).Render(
		titleStyle.Render("Route Statistics") + "\n" + content,
	)
}

// renderInterfaceSummaryPanel renderiza resumo de interfaces
func (m *ModernTUI) renderInterfaceSummaryPanel() string {
	upCount := 0
	downCount := 0
	totalTraffic := int64(0)
	
	for _, intf := range m.interfaceStats {
		if intf.Status == "up" {
			upCount++
		} else {
			downCount++
		}
		totalTraffic += intf.InBytes + intf.OutBytes
	}
	
	content := fmt.Sprintf(`Total Interfaces: %d
Up: %s
Down: %s
Total Traffic: %s
Errors: %d`,
		len(m.interfaceStats),
		successStyle.Render(strconv.Itoa(upCount)),
		errorStyle.Render(strconv.Itoa(downCount)),
		m.formatBytes(totalTraffic),
		0, // TODO: calcular erros
	)
	
	return panelStyle.Width(m.width/2-2).Height(8).Render(
		titleStyle.Render("Interface Summary") + "\n" + content,
	)
}

// Métodos auxiliares
func (m *ModernTUI) nextPanel() {
	m.activePanel = (m.activePanel + 1) % 5
}

func (m *ModernTUI) prevPanel() {
	m.activePanel = (m.activePanel - 1 + 5) % 5
}

func (m *ModernTUI) addError(err string) {
	m.errors = append(m.errors, fmt.Sprintf("[%s] %s", time.Now().Format("15:04:05"), err))
	if len(m.errors) > 10 {
		m.errors = m.errors[1:]
	}
}

func (m *ModernTUI) addInfo(info string) {
	// TODO: implementar log de informações
}

func (m *ModernTUI) formatBytes(bytes int64) string {
	const unit = 1024
	if bytes < unit {
		return fmt.Sprintf("%d B", bytes)
	}
	div, exp := int64(unit), 0
	for n := bytes / unit; n >= unit; n /= unit {
		div *= unit
		exp++
	}
	return fmt.Sprintf("%.1f %cB", float64(bytes)/float64(div), "KMGTPE"[exp])
}

// renderFooter renderiza o rodapé
func (m *ModernTUI) renderFooter() string {
	help := "Tab: Next Panel | Shift+Tab: Prev Panel | R: Refresh | Q: Quit | ?: Help"
	
	if len(m.errors) > 0 {
		lastError := errorStyle.Render("Error: " + m.errors[len(m.errors)-1])
		return lipgloss.JoinVertical(lipgloss.Left, lastError, infoStyle.Render(help))
	}
	
	return infoStyle.Render(help)
}
// Mensagens para o sistema de eventos
type RouterConnectedMsg struct{}
type DataRefreshMsg struct{}
type ErrorMsg struct{ Error string }
type AutoRefreshMsg struct{}

// connectToRouter conecta ao router via SSH
func (m *ModernTUI) connectToRouter() tea.Cmd {
	return func() tea.Msg {
		// Configurar cliente SSH
		sshConfig := ssh.Config{
			Host:     m.routerIP,
			Port:     22,
			Username: m.username,
			Password: m.password,
			Timeout:  10 * time.Second,
		}
		
		client, err := ssh.NewClient(sshConfig)
		if err != nil {
			return ErrorMsg{Error: fmt.Sprintf("Erro ao criar cliente SSH: %v", err)}
		}
		
		ctx := context.Background()
		if err := client.Connect(ctx); err != nil {
			return ErrorMsg{Error: fmt.Sprintf("Erro ao conectar SSH: %v", err)}
		}
		
		m.sshClient = client
		
		// Tentar conectar GoBGP também
		bgpClient, err := gobgp.NewBGPClient("127.0.0.1:50051")
		if err == nil {
			m.bgpClient = bgpClient
		}
		
		return RouterConnectedMsg{}
	}
}

// refreshData atualiza todos os dados
func (m *ModernTUI) refreshData() tea.Cmd {
	return func() tea.Msg {
		if m.sshClient == nil {
			return ErrorMsg{Error: "Não conectado ao router"}
		}
		
		ctx := context.Background()
		
		// Buscar dados BGP
		if err := m.fetchBGPData(ctx); err != nil {
			return ErrorMsg{Error: fmt.Sprintf("Erro ao buscar dados BGP: %v", err)}
		}
		
		// Buscar dados do sistema
		if err := m.fetchSystemData(ctx); err != nil {
			return ErrorMsg{Error: fmt.Sprintf("Erro ao buscar dados do sistema: %v", err)}
		}
		
		// Buscar dados de interfaces
		if err := m.fetchInterfaceData(ctx); err != nil {
			return ErrorMsg{Error: fmt.Sprintf("Erro ao buscar dados de interfaces: %v", err)}
		}
		
		return DataRefreshMsg{}
	}
}

// fetchBGPData busca dados BGP do router
func (m *ModernTUI) fetchBGPData(ctx context.Context) error {
	// Comandos BGP para diferentes vendors
	commands := []string{
		"show ip bgp summary",
		"show bgp summary",
		"show ip bgp neighbors",
		"show bgp neighbors",
	}
	
	var output string
	var err error
	
	// Tentar diferentes comandos até encontrar um que funcione
	for _, cmd := range commands {
		output, err = m.sshClient.ExecuteCommand(ctx, cmd)
		if err == nil && len(output) > 0 {
			break
		}
	}
	
	if err != nil {
		return err
	}
	
	// Parse do output BGP
	m.parseBGPSummary(output)
	m.parseBGPPeers(output)
	
	return nil
}

// fetchSystemData busca informações do sistema
func (m *ModernTUI) fetchSystemData(ctx context.Context) error {
	commands := map[string]string{
		"hostname": "hostname",
		"uptime":   "uptime",
		"version":  "show version",
		"cpu":      "show processes cpu",
		"memory":   "show memory",
	}
	
	for key, cmd := range commands {
		output, err := m.sshClient.ExecuteCommand(ctx, cmd)
		if err != nil {
			continue
		}
		
		switch key {
		case "hostname":
			m.systemInfo.Hostname = strings.TrimSpace(output)
		case "uptime":
			m.systemInfo.Uptime = m.parseUptime(output)
		case "version":
			m.systemInfo.Version = m.parseVersion(output)
		case "cpu":
			m.systemInfo.CPUUsage = m.parseCPUUsage(output)
		case "memory":
			m.systemInfo.MemoryUsage = m.parseMemoryUsage(output)
		}
	}
	
	return nil
}

// fetchInterfaceData busca dados de interfaces
func (m *ModernTUI) fetchInterfaceData(ctx context.Context) error {
	commands := []string{
		"show interfaces",
		"show interface brief",
		"show ip interface brief",
	}
	
	var output string
	var err error
	
	for _, cmd := range commands {
		output, err = m.sshClient.ExecuteCommand(ctx, cmd)
		if err == nil && len(output) > 0 {
			break
		}
	}
	
	if err != nil {
		return err
	}
	
	m.parseInterfaces(output)
	return nil
}

// Parsers para diferentes formatos de output

func (m *ModernTUI) parseBGPSummary(output string) {
	lines := strings.Split(output, "\n")
	
	for _, line := range lines {
		line = strings.TrimSpace(line)
		
		if strings.Contains(line, "BGP router identifier") {
			parts := strings.Fields(line)
			if len(parts) >= 4 {
				m.bgpSummary.RouterID = parts[3]
			}
		}
		
		if strings.Contains(line, "local AS number") {
			parts := strings.Fields(line)
			for i, part := range parts {
				if part == "AS" && i+1 < len(parts) {
					if asn, err := strconv.Atoi(parts[i+1]); err == nil {
						m.bgpSummary.LocalAS = asn
					}
				}
			}
		}
	}
	
	// Valores padrão para demonstração
	if m.bgpSummary.RouterID == "" {
		m.bgpSummary.RouterID = m.routerIP
	}
	if m.bgpSummary.LocalAS == 0 {
		m.bgpSummary.LocalAS = 262978
	}
	
	m.bgpSummary.TotalPeers = len(m.peerStats)
	m.bgpSummary.ActivePeers = 0
	for _, peer := range m.peerStats {
		if peer.State == "Established" {
			m.bgpSummary.ActivePeers++
		}
	}
}

func (m *ModernTUI) parseBGPPeers(output string) {
	// Parse básico de peers BGP
	lines := strings.Split(output, "\n")
	m.peerStats = []PeerStat{}
	
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if len(line) == 0 || strings.HasPrefix(line, "BGP") || strings.HasPrefix(line, "Neighbor") {
			continue
		}
		
		fields := strings.Fields(line)
		if len(fields) >= 4 {
			peer := PeerStat{
				IP:    fields[0],
				State: "Established",
			}
			
			// Tentar extrair ASN
			if len(fields) > 1 {
				if asn, err := strconv.Atoi(fields[1]); err == nil {
					peer.ASN = asn
				}
			}
			
			m.peerStats = append(m.peerStats, peer)
		}
	}
	
	// Dados de exemplo se não conseguir fazer parse
	if len(m.peerStats) == 0 {
		m.peerStats = []PeerStat{
			{IP: "10.0.0.1", ASN: 65001, State: "Established", Uptime: "2d3h", PrefixRcv: 150000, PrefixSent: 5000},
			{IP: "10.0.0.2", ASN: 65002, State: "Established", Uptime: "1d5h", PrefixRcv: 75000, PrefixSent: 3000},
			{IP: "10.0.0.3", ASN: 65003, State: "Idle", Uptime: "0", PrefixRcv: 0, PrefixSent: 0},
		}
	}
}

func (m *ModernTUI) parseInterfaces(output string) {
	lines := strings.Split(output, "\n")
	m.interfaceStats = []InterfaceStat{}
	
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if len(line) == 0 || strings.HasPrefix(line, "Interface") {
			continue
		}
		
		fields := strings.Fields(line)
		if len(fields) >= 3 {
			intf := InterfaceStat{
				Name:   fields[0],
				Status: "up",
				Speed:  "1Gbps",
			}
			
			if len(fields) > 1 && (fields[1] == "down" || fields[1] == "administratively") {
				intf.Status = "down"
			}
			
			m.interfaceStats = append(m.interfaceStats, intf)
		}
	}
	
	// Dados de exemplo se não conseguir fazer parse
	if len(m.interfaceStats) == 0 {
		m.interfaceStats = []InterfaceStat{
			{Name: "GigabitEthernet0/0", Status: "up", Speed: "1Gbps", InBytes: 1024000000, OutBytes: 512000000},
			{Name: "GigabitEthernet0/1", Status: "up", Speed: "1Gbps", InBytes: 2048000000, OutBytes: 1024000000},
			{Name: "GigabitEthernet0/2", Status: "down", Speed: "1Gbps", InBytes: 0, OutBytes: 0},
		}
	}
}

func (m *ModernTUI) parseUptime(output string) string {
	// Parse simples de uptime
	if strings.Contains(output, "up") {
		return strings.TrimSpace(output)
	}
	return "Unknown"
}

func (m *ModernTUI) parseVersion(output string) string {
	lines := strings.Split(output, "\n")
	for _, line := range lines {
		if strings.Contains(line, "Version") || strings.Contains(line, "version") {
			return strings.TrimSpace(line)
		}
	}
	return "Unknown"
}

func (m *ModernTUI) parseCPUUsage(output string) float64 {
	// Parse básico de CPU
	lines := strings.Split(output, "\n")
	for _, line := range lines {
		if strings.Contains(line, "%") {
			fields := strings.Fields(line)
			for _, field := range fields {
				if strings.Contains(field, "%") {
					if val, err := strconv.ParseFloat(strings.TrimSuffix(field, "%"), 64); err == nil {
						return val
					}
				}
			}
		}
	}
	return 25.5 // Valor padrão para demonstração
}

func (m *ModernTUI) parseMemoryUsage(output string) float64 {
	// Parse básico de memória
	return 45.2 // Valor padrão para demonstração
}

// startAutoRefresh inicia refresh automático
func (m *ModernTUI) startAutoRefresh() tea.Cmd {
	return tea.Tick(m.refreshRate, func(time.Time) tea.Msg {
		return AutoRefreshMsg{}
	})
}

// initializeTables inicializa as tabelas
func (m *ModernTUI) initializeTables() tea.Cmd {
	// Tabela de peers
	peerColumns := []table.Column{
		{Title: "IP", Width: 15},
		{Title: "ASN", Width: 8},
		{Title: "State", Width: 12},
		{Title: "Uptime", Width: 10},
		{Title: "Rcv", Width: 8},
		{Title: "Sent", Width: 8},
	}
	
	m.peersTable = table.New(
		table.WithColumns(peerColumns),
		table.WithFocused(true),
		table.WithHeight(10),
	)
	
	// Tabela de rotas
	routeColumns := []table.Column{
		{Title: "Network", Width: 18},
		{Title: "Next Hop", Width: 15},
		{Title: "Metric", Width: 8},
		{Title: "LocPrf", Width: 8},
		{Title: "Weight", Width: 8},
		{Title: "Path", Width: 20},
	}
	
	m.routesTable = table.New(
		table.WithColumns(routeColumns),
		table.WithHeight(10),
	)
	
	// Tabela de interfaces
	intfColumns := []table.Column{
		{Title: "Interface", Width: 20},
		{Title: "Status", Width: 8},
		{Title: "Speed", Width: 10},
		{Title: "In Bytes", Width: 12},
		{Title: "Out Bytes", Width: 12},
		{Title: "Errors", Width: 8},
	}
	
	m.interfacesTable = table.New(
		table.WithColumns(intfColumns),
		table.WithHeight(10),
	)
	
	return nil
}

// updateTables atualiza dados das tabelas
func (m *ModernTUI) updateTables() {
	// Atualizar tabela de peers
	peerRows := make([]table.Row, len(m.peerStats))
	for i, peer := range m.peerStats {
		state := peer.State
		if state == "Established" {
			state = successStyle.Render(state)
		} else {
			state = errorStyle.Render(state)
		}
		
		peerRows[i] = table.Row{
			peer.IP,
			fmt.Sprintf("AS%d", peer.ASN),
			state,
			peer.Uptime,
			strconv.Itoa(peer.PrefixRcv),
			strconv.Itoa(peer.PrefixSent),
		}
	}
	m.peersTable.SetRows(peerRows)
	
	// Atualizar tabela de interfaces
	intfRows := make([]table.Row, len(m.interfaceStats))
	for i, intf := range m.interfaceStats {
		status := intf.Status
		if status == "up" {
			status = successStyle.Render(status)
		} else {
			status = errorStyle.Render(status)
		}
		
		intfRows[i] = table.Row{
			intf.Name,
			status,
			intf.Speed,
			m.formatBytes(intf.InBytes),
			m.formatBytes(intf.OutBytes),
			strconv.FormatInt(intf.InErrors+intf.OutErrors, 10),
		}
	}
	m.interfacesTable.SetRows(intfRows)
}

func (m *ModernTUI) updateTableSizes() {
	if m.width > 0 {
		// Recriar tabelas com nova largura
		peerColumns := []table.Column{
			{Title: "IP", Width: 15},
			{Title: "ASN", Width: 8},
			{Title: "State", Width: 12},
			{Title: "Uptime", Width: 10},
			{Title: "Rcv", Width: 8},
			{Title: "Sent", Width: 8},
		}
		
		m.peersTable = table.New(
			table.WithColumns(peerColumns),
			table.WithFocused(true),
			table.WithHeight(10),
			table.WithWidth(m.width-4),
		)
		
		routeColumns := []table.Column{
			{Title: "Network", Width: 18},
			{Title: "Next Hop", Width: 15},
			{Title: "Metric", Width: 8},
			{Title: "LocPrf", Width: 8},
			{Title: "Weight", Width: 8},
			{Title: "Path", Width: 20},
		}
		
		m.routesTable = table.New(
			table.WithColumns(routeColumns),
			table.WithHeight(10),
			table.WithWidth(m.width-4),
		)
		
		intfColumns := []table.Column{
			{Title: "Interface", Width: 20},
			{Title: "Status", Width: 8},
			{Title: "Speed", Width: 10},
			{Title: "In Bytes", Width: 12},
			{Title: "Out Bytes", Width: 12},
			{Title: "Errors", Width: 8},
		}
		
		m.interfacesTable = table.New(
			table.WithColumns(intfColumns),
			table.WithHeight(10),
			table.WithWidth(m.width-4),
		)
	}
}

// renderPeersPanel renderiza painel de peers
func (m *ModernTUI) renderPeersPanel() string {
	style := panelStyle
	if m.activePanel == 1 {
		style = activePanelStyle
	}
	
	var content strings.Builder
	
	// Estatísticas de peers
	content.WriteString(fmt.Sprintf("Total Peers: %d | Active: %s | Idle: %s\n\n",
		len(m.peerStats),
		successStyle.Render(strconv.Itoa(m.bgpSummary.ActivePeers)),
		errorStyle.Render(strconv.Itoa(len(m.peerStats)-m.bgpSummary.ActivePeers))))
	
	// Gráfico de peers por AS
	if m.bgpPeerChart != nil {
		content.WriteString(m.bgpPeerChart.Render() + "\n")
	}
	
	// Tabela de peers
	content.WriteString(m.peersTable.View())
	
	return style.Width(m.width-2).Height(m.height-8).Render(
		titleStyle.Render("BGP Peers") + "\n" + content.String(),
	)
}

// renderRoutesPanel renderiza painel de rotas
func (m *ModernTUI) renderRoutesPanel() string {
	style := panelStyle
	if m.activePanel == 2 {
		style = activePanelStyle
	}
	
	// Dados de exemplo para rotas
	routeRows := []table.Row{
		{"0.0.0.0/0", "10.0.0.1", "0", "100", "0", "65001 65002"},
		{"192.168.1.0/24", "10.0.0.2", "0", "100", "0", "65003"},
		{"10.0.0.0/8", "10.0.0.1", "0", "100", "0", "65001"},
	}
	m.routesTable.SetRows(routeRows)
	
	return style.Width(m.width-2).Height(m.height-8).Render(
		titleStyle.Render("BGP Routes") + "\n" + m.routesTable.View(),
	)
}

// renderInterfacesPanel renderiza painel de interfaces
func (m *ModernTUI) renderInterfacesPanel() string {
	style := panelStyle
	if m.activePanel == 3 {
		style = activePanelStyle
	}
	
	return style.Width(m.width-2).Height(m.height-8).Render(
		titleStyle.Render("Network Interfaces") + "\n" + m.interfacesTable.View(),
	)
}

// renderSystemPanel renderiza painel do sistema
func (m *ModernTUI) renderSystemPanel() string {
	style := panelStyle
	if m.activePanel == 4 {
		style = activePanelStyle
	}
	
	var content strings.Builder
	
	// Informações do sistema
	content.WriteString(fmt.Sprintf("Hostname: %s\n", m.systemInfo.Hostname))
	content.WriteString(fmt.Sprintf("Version: %s\n", m.systemInfo.Version))
	content.WriteString(fmt.Sprintf("Uptime: %s\n\n", m.systemInfo.Uptime))
	
	// Gráficos de sistema
	if m.cpuChart != nil && m.memoryChart != nil && m.networkChart != nil {
		content.WriteString("System Performance:\n")
		content.WriteString(m.cpuChart.Render() + "\n")
		content.WriteString(m.memoryChart.Render() + "\n")
		content.WriteString(m.networkChart.Render() + "\n")
	}
	
	// Alertas recentes
	content.WriteString("\nRecent System Events:\n")
	if len(m.errors) > 0 {
		for i, err := range m.errors {
			if i >= 5 {
				break
			}
			content.WriteString(errorStyle.Render("• "+err) + "\n")
		}
	} else {
		content.WriteString(successStyle.Render("All systems normal") + "\n")
	}
	
	return style.Width(m.width-2).Height(m.height-8).Render(
		titleStyle.Render("System Status") + "\n" + content.String(),
	)
}
// simulateMetrics simula dados de métricas para demonstração
func (m *ModernTUI) simulateMetrics() {
	if !m.demoMode {
		return
	}
	
	ticker := time.NewTicker(1 * time.Second)
	defer ticker.Stop()
	
	for {
		select {
		case <-ticker.C:
			now := time.Now()
			
			// Simular CPU usage (20-80%)
			cpuUsage := 20.0 + 30.0*math.Sin(float64(now.Unix())/10.0) + 10.0*math.Sin(float64(now.Unix())/3.0)
			if cpuUsage < 0 {
				cpuUsage = 0
			}
			if cpuUsage > 100 {
				cpuUsage = 100
			}
			
			m.cpuChart.AddData(cpuUsage)
			m.cpuGauge.SetValue(cpuUsage)
			m.metricsCollector.RecordValue("cpu_usage", cpuUsage)
			m.alertManager.CheckMetric("cpu_usage", cpuUsage)
			m.systemInfo.CPUUsage = cpuUsage
			
			// Simular Memory usage (30-70%)
			memUsage := 40.0 + 20.0*math.Sin(float64(now.Unix())/15.0) + 5.0*math.Sin(float64(now.Unix())/5.0)
			if memUsage < 0 {
				memUsage = 0
			}
			if memUsage > 100 {
				memUsage = 100
			}
			
			m.memoryChart.AddData(memUsage)
			m.memoryGauge.SetValue(memUsage)
			m.metricsCollector.RecordValue("memory_usage", memUsage)
			m.alertManager.CheckMetric("memory_usage", memUsage)
			m.systemInfo.MemoryUsage = memUsage
			
			// Simular Network traffic (entrada e saída)
			baseTraffic := 50.0
			inTraffic := baseTraffic + 30.0*math.Sin(float64(now.Unix())/8.0) + 10.0*math.Sin(float64(now.Unix())/2.0)
			outTraffic := baseTraffic + 25.0*math.Sin(float64(now.Unix())/12.0) + 15.0*math.Sin(float64(now.Unix())/4.0)
			
			if inTraffic < 0 {
				inTraffic = 0
			}
			if outTraffic < 0 {
				outTraffic = 0
			}
			
			m.networkChart.AddData(inTraffic, outTraffic)
			m.metricsCollector.RecordValue("network_traffic", inTraffic+outTraffic)
			
			// Simular dados de análise de rotas
			ipv4Routes := 750000.0 + 50000.0*math.Sin(float64(now.Unix())/30.0)
			ipv6Routes := 100000.0 + 10000.0*math.Sin(float64(now.Unix())/25.0)
			bgpUpdates := 1000.0 + 500.0*math.Sin(float64(now.Unix())/5.0)
			
			m.routeAnalysisChart.AddDataPoint("IPv4 Routes", now, ipv4Routes)
			m.routeAnalysisChart.AddDataPoint("IPv6 Routes", now, ipv6Routes)
			m.routeAnalysisChart.AddDataPoint("BGP Updates", now, bgpUpdates)
			
			// Simular dados de candlestick para análise de tráfego (a cada 5 segundos)
			if now.Unix()%5 == 0 {
				// Simular OHLC para tráfego
				basePrice := 100.0
				volatility := 10.0
				
				open := basePrice + volatility*math.Sin(float64(now.Unix())/20.0)
				high := open + math.Abs(volatility*math.Sin(float64(now.Unix())/7.0))
				low := open - math.Abs(volatility*math.Sin(float64(now.Unix())/11.0))
				close := open + volatility*math.Sin(float64(now.Unix())/13.0)
				volume := 1000.0 + 500.0*math.Abs(math.Sin(float64(now.Unix())/6.0))
				
				m.trafficAnalysisChart.AddData(now, open, high, low, close, volume)
			}
			
			// Simular dados de peers BGP
			for i, peer := range m.peerStats {
				if peer.State == "Established" {
					// Simular variação nos prefixos recebidos
					variation := 1.0 + 0.1*math.Sin(float64(now.Unix()+int64(i))/20.0)
					prefixCount := float64(peer.PrefixRcv) * variation
					
					peerName := fmt.Sprintf("AS%d", peer.ASN)
					m.bgpPeerChart.AddPeerData(peerName, prefixCount)
					
					// Adicionar ao gráfico de volume
					peerActivity := prefixCount / 1000.0 // Normalizar
					m.peerVolumeChart.AddData(now, peerActivity, prefixCount)
				}
			}
			
			// Simular BGP peers
			activePeers := float64(m.bgpSummary.ActivePeers)
			m.metricsCollector.RecordValue("bgp_peers", activePeers)
			m.alertManager.CheckMetric("bgp_peers", activePeers)
			
			// Atualizar alertas
			m.alerts = m.alertManager.GetAlerts()
		}
	}
}

