package components

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/lipgloss"
)

// LargeChart representa um gráfico grande estilo btop
type LargeChart struct {
	data     []float64
	width    int
	height   int
	title    string
	unit     string
	color    lipgloss.Color
	min      float64
	max      float64
	showGrid bool
	showAxis bool
}

// NewLargeChart cria um novo gráfico grande
func NewLargeChart(title, unit string, width, height int) *LargeChart {
	return &LargeChart{
		data:     make([]float64, 0),
		width:    width,
		height:   height,
		title:    title,
		unit:     unit,
		color:    lipgloss.Color("#00FF00"),
		showGrid: true,
		showAxis: true,
	}
}

// AddData adiciona dados ao gráfico
func (c *LargeChart) AddData(value float64) {
	c.data = append(c.data, value)
	
	// Manter apenas os dados que cabem na largura
	maxPoints := c.width - 10 // Reservar espaço para eixo Y
	if len(c.data) > maxPoints {
		c.data = c.data[len(c.data)-maxPoints:]
	}
	
	c.updateMinMax()
}

// updateMinMax atualiza valores mínimo e máximo
func (c *LargeChart) updateMinMax() {
	if len(c.data) == 0 {
		c.min = 0
		c.max = 100
		return
	}
	
	c.min = c.data[0]
	c.max = c.data[0]
	
	for _, v := range c.data {
		if v < c.min {
			c.min = v
		}
		if v > c.max {
			c.max = v
		}
	}
	
	// Adicionar margem
	range_ := c.max - c.min
	if range_ == 0 {
		range_ = 1
	}
	margin := range_ * 0.1
	c.min -= margin
	c.max += margin
	
	if c.min < 0 {
		c.min = 0
	}
}

// Render renderiza o gráfico grande
func (c *LargeChart) Render() string {
	if len(c.data) == 0 {
		return c.renderEmpty()
	}
	
	var result strings.Builder
	
	// Título
	titleLine := fmt.Sprintf("%s (%s)", c.title, c.unit)
	result.WriteString(lipgloss.NewStyle().Bold(true).Foreground(c.color).Render(titleLine) + "\n")
	
	// Estatísticas na linha do título
	current := c.data[len(c.data)-1]
	avg := c.getAverage()
	statsLine := fmt.Sprintf("Current: %.1f %s | Avg: %.1f %s | Min: %.1f | Max: %.1f",
		current, c.unit, avg, c.unit, c.min, c.max)
	result.WriteString(lipgloss.NewStyle().Foreground(lipgloss.Color("#666666")).Render(statsLine) + "\n")
	
	// Gráfico principal
	chart := c.renderChart()
	result.WriteString(chart)
	
	return result.String()
}

// renderChart renderiza o gráfico principal
func (c *LargeChart) renderChart() string {
	chartHeight := c.height - 3 // Reservar espaço para título e stats
	chartWidth := c.width - 8   // Reservar espaço para eixo Y
	
	if chartHeight < 3 || chartWidth < 10 {
		return "Chart too small"
	}
	
	var lines []string
	
	// Renderizar linha por linha (de cima para baixo)
	for row := 0; row < chartHeight; row++ {
		line := c.renderChartLine(row, chartHeight, chartWidth)
		lines = append(lines, line)
	}
	
	// Adicionar eixo X
	xAxis := c.renderXAxis(chartWidth)
	lines = append(lines, xAxis)
	
	return strings.Join(lines, "\n")
}

// renderChartLine renderiza uma linha do gráfico
func (c *LargeChart) renderChartLine(row, totalHeight, chartWidth int) string {
	var line strings.Builder
	
	// Eixo Y (valores)
	yValue := c.max - (float64(row)/float64(totalHeight-1))*(c.max-c.min)
	yLabel := fmt.Sprintf("%6.1f", yValue)
	line.WriteString(lipgloss.NewStyle().Foreground(lipgloss.Color("#666666")).Render(yLabel))
	line.WriteString(" ")
	
	// Linha do gráfico
	threshold := yValue
	
	for col := 0; col < chartWidth; col++ {
		if col < len(c.data) {
			value := c.data[col]
			
			if value >= threshold {
				// Ponto do gráfico
				if row == 0 || (row > 0 && c.data[col] < (c.max - (float64(row-1)/float64(totalHeight-1))*(c.max-c.min))) {
					line.WriteString(lipgloss.NewStyle().Foreground(c.color).Render("█"))
				} else {
					line.WriteString(lipgloss.NewStyle().Foreground(c.color).Render("█"))
				}
			} else {
				// Grid ou espaço vazio
				if c.showGrid && (row%2 == 0 || col%10 == 0) {
					line.WriteString(lipgloss.NewStyle().Foreground(lipgloss.Color("#333333")).Render("·"))
				} else {
					line.WriteString(" ")
				}
			}
		} else {
			// Espaço vazio
			if c.showGrid && (row%2 == 0) {
				line.WriteString(lipgloss.NewStyle().Foreground(lipgloss.Color("#333333")).Render("·"))
			} else {
				line.WriteString(" ")
			}
		}
	}
	
	return line.String()
}

// renderXAxis renderiza o eixo X
func (c *LargeChart) renderXAxis(chartWidth int) string {
	var line strings.Builder
	
	// Espaço para alinhamento com eixo Y
	line.WriteString("       ")
	
	// Linha do eixo
	for i := 0; i < chartWidth; i++ {
		if i%10 == 0 {
			line.WriteString(lipgloss.NewStyle().Foreground(lipgloss.Color("#666666")).Render("┼"))
		} else {
			line.WriteString(lipgloss.NewStyle().Foreground(lipgloss.Color("#666666")).Render("─"))
		}
	}
	
	return line.String()
}

// renderEmpty renderiza gráfico vazio
func (c *LargeChart) renderEmpty() string {
	return lipgloss.NewStyle().Foreground(lipgloss.Color("#666666")).Render(
		fmt.Sprintf("%s - No data available", c.title))
}

// getAverage calcula a média dos dados
func (c *LargeChart) getAverage() float64 {
	if len(c.data) == 0 {
		return 0
	}
	
	var sum float64
	for _, v := range c.data {
		sum += v
	}
	
	return sum / float64(len(c.data))
}

// SetColor define a cor do gráfico
func (c *LargeChart) SetColor(color lipgloss.Color) {
	c.color = color
}

// NetworkChart representa um gráfico de rede com múltiplas séries
type NetworkChart struct {
	inData    []float64
	outData   []float64
	width     int
	height    int
	title     string
	unit      string
	maxValue  float64
}

// NewNetworkChart cria um novo gráfico de rede
func NewNetworkChart(title, unit string, width, height int) *NetworkChart {
	return &NetworkChart{
		inData:   make([]float64, 0),
		outData:  make([]float64, 0),
		width:    width,
		height:   height,
		title:    title,
		unit:     unit,
		maxValue: 100,
	}
}

// AddData adiciona dados de entrada e saída
func (nc *NetworkChart) AddData(inValue, outValue float64) {
	nc.inData = append(nc.inData, inValue)
	nc.outData = append(nc.outData, outValue)
	
	maxPoints := nc.width - 10
	if len(nc.inData) > maxPoints {
		nc.inData = nc.inData[1:]
		nc.outData = nc.outData[1:]
	}
	
	// Atualizar valor máximo
	for _, v := range nc.inData {
		if v > nc.maxValue {
			nc.maxValue = v
		}
	}
	for _, v := range nc.outData {
		if v > nc.maxValue {
			nc.maxValue = v
		}
	}
}

// Render renderiza o gráfico de rede
func (nc *NetworkChart) Render() string {
	if len(nc.inData) == 0 {
		return lipgloss.NewStyle().Foreground(lipgloss.Color("#666666")).Render(
			fmt.Sprintf("%s - No data available", nc.title))
	}
	
	var result strings.Builder
	
	// Título e legenda
	result.WriteString(lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("#FFFFFF")).Render(nc.title) + "\n")
	
	legend := lipgloss.NewStyle().Foreground(lipgloss.Color("#00FF88")).Render("▲ In") + " " +
		lipgloss.NewStyle().Foreground(lipgloss.Color("#FF6600")).Render("▼ Out") + " " +
		fmt.Sprintf("Max: %.1f %s", nc.maxValue, nc.unit)
	result.WriteString(legend + "\n")
	
	// Gráfico
	chartHeight := nc.height - 3
	chartWidth := nc.width - 8
	
	for row := 0; row < chartHeight; row++ {
		line := nc.renderNetworkLine(row, chartHeight, chartWidth)
		result.WriteString(line + "\n")
	}
	
	return result.String()
}

// renderNetworkLine renderiza uma linha do gráfico de rede
func (nc *NetworkChart) renderNetworkLine(row, totalHeight, chartWidth int) string {
	var line strings.Builder
	
	// Eixo Y
	yValue := nc.maxValue * (1.0 - float64(row)/float64(totalHeight-1))
	yLabel := fmt.Sprintf("%6.1f", yValue)
	line.WriteString(lipgloss.NewStyle().Foreground(lipgloss.Color("#666666")).Render(yLabel))
	line.WriteString(" ")
	
	// Dados
	for col := 0; col < chartWidth && col < len(nc.inData); col++ {
		inValue := nc.inData[col]
		outValue := nc.outData[col]
		
		inThreshold := yValue
		outThreshold := yValue
		
		inActive := inValue >= inThreshold
		outActive := outValue >= outThreshold
		
		if inActive && outActive {
			// Ambos ativos - usar cor mista
			line.WriteString(lipgloss.NewStyle().Foreground(lipgloss.Color("#00AAAA")).Render("█"))
		} else if inActive {
			// Apenas entrada ativa
			line.WriteString(lipgloss.NewStyle().Foreground(lipgloss.Color("#00FF88")).Render("▲"))
		} else if outActive {
			// Apenas saída ativa
			line.WriteString(lipgloss.NewStyle().Foreground(lipgloss.Color("#FF6600")).Render("▼"))
		} else {
			if row%3 == 0 && col%5 == 0 {
				line.WriteString(lipgloss.NewStyle().Foreground(lipgloss.Color("#333333")).Render("·"))
			} else {
				line.WriteString(" ")
			}
		}
	}
	
	return line.String()
}

// BGPPeerChart representa um gráfico de peers BGP
type BGPPeerChart struct {
	peerData map[string][]float64
	width    int
	height   int
	title    string
	colors   []lipgloss.Color
}

// NewBGPPeerChart cria um novo gráfico de peers BGP
func NewBGPPeerChart(title string, width, height int) *BGPPeerChart {
	return &BGPPeerChart{
		peerData: make(map[string][]float64),
		width:    width,
		height:   height,
		title:    title,
		colors: []lipgloss.Color{
			lipgloss.Color("#00FF00"), // Verde
			lipgloss.Color("#0088FF"), // Azul
			lipgloss.Color("#FF6600"), // Laranja
			lipgloss.Color("#FF4488"), // Rosa
			lipgloss.Color("#00FF88"), // Verde claro
			lipgloss.Color("#8800FF"), // Roxo
			lipgloss.Color("#FF8844"), // Laranja claro
			lipgloss.Color("#44FF88"), // Verde água
		},
	}
}

// AddPeerData adiciona dados de um peer
func (bpc *BGPPeerChart) AddPeerData(peerName string, value float64) {
	if _, exists := bpc.peerData[peerName]; !exists {
		bpc.peerData[peerName] = make([]float64, 0)
	}
	
	bpc.peerData[peerName] = append(bpc.peerData[peerName], value)
	
	maxPoints := bpc.width - 15
	if len(bpc.peerData[peerName]) > maxPoints {
		bpc.peerData[peerName] = bpc.peerData[peerName][1:]
	}
}

// Render renderiza o gráfico de peers BGP
func (bpc *BGPPeerChart) Render() string {
	if len(bpc.peerData) == 0 {
		return lipgloss.NewStyle().Foreground(lipgloss.Color("#666666")).Render(
			fmt.Sprintf("%s - No peer data available", bpc.title))
	}
	
	var result strings.Builder
	
	// Título
	result.WriteString(lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("#FFFFFF")).Render(bpc.title) + "\n")
	
	// Legenda
	colorIndex := 0
	var legendItems []string
	for peerName := range bpc.peerData {
		color := bpc.colors[colorIndex%len(bpc.colors)]
		legendItems = append(legendItems, lipgloss.NewStyle().Foreground(color).Render("■ "+peerName))
		colorIndex++
	}
	result.WriteString(strings.Join(legendItems, " ") + "\n")
	
	// Encontrar valor máximo
	var maxValue float64
	for _, data := range bpc.peerData {
		for _, value := range data {
			if value > maxValue {
				maxValue = value
			}
		}
	}
	
	if maxValue == 0 {
		maxValue = 1
	}
	
	// Gráfico
	chartHeight := bpc.height - 3
	chartWidth := bpc.width - 10
	
	for row := 0; row < chartHeight; row++ {
		line := bpc.renderPeerLine(row, chartHeight, chartWidth, maxValue)
		result.WriteString(line + "\n")
	}
	
	return result.String()
}

// renderPeerLine renderiza uma linha do gráfico de peers
func (bpc *BGPPeerChart) renderPeerLine(row, totalHeight, chartWidth int, maxValue float64) string {
	var line strings.Builder
	
	// Eixo Y
	yValue := maxValue * (1.0 - float64(row)/float64(totalHeight-1))
	yLabel := fmt.Sprintf("%8.0f", yValue)
	line.WriteString(lipgloss.NewStyle().Foreground(lipgloss.Color("#666666")).Render(yLabel))
	line.WriteString(" ")
	
	// Dados dos peers
	for col := 0; col < chartWidth; col++ {
		var activeChar string
		var activeColor lipgloss.Color
		found := false
		
		colorIndex := 0
		for _, data := range bpc.peerData {
			if col < len(data) && data[col] >= yValue {
				activeChar = "█"
				activeColor = bpc.colors[colorIndex%len(bpc.colors)]
				found = true
				break
			}
			colorIndex++
		}
		
		if found {
			line.WriteString(lipgloss.NewStyle().Foreground(activeColor).Render(activeChar))
		} else {
			if row%2 == 0 && col%5 == 0 {
				line.WriteString(lipgloss.NewStyle().Foreground(lipgloss.Color("#333333")).Render("·"))
			} else {
				line.WriteString(" ")
			}
		}
	}
	
	return line.String()
}