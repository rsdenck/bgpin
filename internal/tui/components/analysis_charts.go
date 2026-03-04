package components

import (
	"fmt"
	"math"
	"strings"
	"time"

	"github.com/charmbracelet/lipgloss"
)

// CandlestickChart representa um gráfico de candlestick para análise técnica
type CandlestickChart struct {
	data     []CandlestickData
	width    int
	height   int
	title    string
	timeframe string
	maxValue float64
	minValue float64
}

// CandlestickData representa dados OHLC
type CandlestickData struct {
	Timestamp time.Time
	Open      float64
	High      float64
	Low       float64
	Close     float64
	Volume    float64
}

// NewCandlestickChart cria um novo gráfico candlestick
func NewCandlestickChart(title, timeframe string, width, height int) *CandlestickChart {
	return &CandlestickChart{
		data:      make([]CandlestickData, 0),
		width:     width,
		height:    height,
		title:     title,
		timeframe: timeframe,
	}
}

// AddData adiciona dados OHLC
func (c *CandlestickChart) AddData(timestamp time.Time, open, high, low, close, volume float64) {
	data := CandlestickData{
		Timestamp: timestamp,
		Open:      open,
		High:      high,
		Low:       low,
		Close:     close,
		Volume:    volume,
	}
	
	c.data = append(c.data, data)
	
	// Manter apenas os dados que cabem na largura
	maxCandles := (c.width - 15) / 3 // Cada candle ocupa ~3 caracteres
	if len(c.data) > maxCandles {
		c.data = c.data[len(c.data)-maxCandles:]
	}
	
	c.updateMinMax()
}

// updateMinMax atualiza valores mínimo e máximo
func (c *CandlestickChart) updateMinMax() {
	if len(c.data) == 0 {
		return
	}
	
	c.minValue = c.data[0].Low
	c.maxValue = c.data[0].High
	
	for _, candle := range c.data {
		if candle.Low < c.minValue {
			c.minValue = candle.Low
		}
		if candle.High > c.maxValue {
			c.maxValue = candle.High
		}
	}
	
	// Adicionar margem
	range_ := c.maxValue - c.minValue
	if range_ == 0 {
		range_ = 1
	}
	margin := range_ * 0.05
	c.minValue -= margin
	c.maxValue += margin
}

// Render renderiza o gráfico candlestick
func (c *CandlestickChart) Render() string {
	if len(c.data) == 0 {
		return c.renderEmpty()
	}
	
	var result strings.Builder
	
	// Título e informações
	titleLine := fmt.Sprintf("%s (%s)", c.title, c.timeframe)
	result.WriteString(lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("#FFFFFF")).Render(titleLine) + "\n")
	
	// Estatísticas da última vela
	lastCandle := c.data[len(c.data)-1]
	change := lastCandle.Close - lastCandle.Open
	changePercent := (change / lastCandle.Open) * 100
	
	var changeColor lipgloss.Color
	var changeSymbol string
	if change >= 0 {
		changeColor = lipgloss.Color("#00FF88")
		changeSymbol = "▲"
	} else {
		changeColor = lipgloss.Color("#FF4444")
		changeSymbol = "▼"
	}
	
	statsLine := fmt.Sprintf("O: %.2f H: %.2f L: %.2f C: %.2f %s %.2f (%.2f%%)",
		lastCandle.Open, lastCandle.High, lastCandle.Low, lastCandle.Close,
		changeSymbol, change, changePercent)
	result.WriteString(lipgloss.NewStyle().Foreground(changeColor).Render(statsLine) + "\n")
	
	// Gráfico
	chart := c.renderCandlesticks()
	result.WriteString(chart)
	
	return result.String()
}

// renderCandlesticks renderiza as velas
func (c *CandlestickChart) renderCandlesticks() string {
	chartHeight := c.height - 3
	chartWidth := c.width - 10
	
	var lines []string
	
	// Renderizar linha por linha
	for row := 0; row < chartHeight; row++ {
		line := c.renderCandlestickLine(row, chartHeight, chartWidth)
		lines = append(lines, line)
	}
	
	// Eixo X com timestamps
	xAxis := c.renderTimeAxis(chartWidth)
	lines = append(lines, xAxis)
	
	return strings.Join(lines, "\n")
}

// renderCandlestickLine renderiza uma linha do gráfico
func (c *CandlestickChart) renderCandlestickLine(row, totalHeight, chartWidth int) string {
	var line strings.Builder
	
	// Eixo Y
	yValue := c.maxValue - (float64(row)/float64(totalHeight-1))*(c.maxValue-c.minValue)
	yLabel := fmt.Sprintf("%8.2f", yValue)
	line.WriteString(lipgloss.NewStyle().Foreground(lipgloss.Color("#666666")).Render(yLabel))
	line.WriteString(" ")
	
	// Velas
	candleWidth := 3
	for i, candle := range c.data {
		col := i * candleWidth
		if col >= chartWidth {
			break
		}
		
		// Determinar se é vela verde (alta) ou vermelha (baixa)
		isGreen := candle.Close >= candle.Open
		var bodyColor, wickColor lipgloss.Color
		
		if isGreen {
			bodyColor = lipgloss.Color("#00FF88")
			wickColor = lipgloss.Color("#00AA66")
		} else {
			bodyColor = lipgloss.Color("#FF4444")
			wickColor = lipgloss.Color("#AA2222")
		}
		
		// Verificar se o preço atual está na linha
		bodyTop := math.Max(candle.Open, candle.Close)
		bodyBottom := math.Min(candle.Open, candle.Close)
		
		var char string
		var color lipgloss.Color
		
		if yValue <= candle.High && yValue >= bodyTop {
			// Pavio superior
			char = "│"
			color = wickColor
		} else if yValue <= bodyTop && yValue >= bodyBottom {
			// Corpo da vela
			if candleWidth >= 3 {
				if col == i*candleWidth {
					char = "█"
				} else if col == i*candleWidth+1 {
					char = "█"
				} else {
					char = "█"
				}
			} else {
				char = "█"
			}
			color = bodyColor
		} else if yValue <= bodyBottom && yValue >= candle.Low {
			// Pavio inferior
			char = "│"
			color = wickColor
		} else {
			char = " "
			color = lipgloss.Color("#000000")
		}
		
		line.WriteString(lipgloss.NewStyle().Foreground(color).Render(char))
		
		// Espaçamento entre velas
		if col+1 < chartWidth {
			line.WriteString(" ")
		}
	}
	
	return line.String()
}

// renderTimeAxis renderiza o eixo de tempo
func (c *CandlestickChart) renderTimeAxis(chartWidth int) string {
	var line strings.Builder
	line.WriteString("         ") // Alinhamento com eixo Y
	
	candleWidth := 3
	for i, candle := range c.data {
		col := i * candleWidth
		if col >= chartWidth {
			break
		}
		
		if i%5 == 0 { // Mostrar timestamp a cada 5 velas
			timeStr := candle.Timestamp.Format("15:04")
			line.WriteString(lipgloss.NewStyle().Foreground(lipgloss.Color("#666666")).Render(timeStr))
		} else {
			line.WriteString("     ")
		}
	}
	
	return line.String()
}

// renderEmpty renderiza gráfico vazio
func (c *CandlestickChart) renderEmpty() string {
	return lipgloss.NewStyle().Foreground(lipgloss.Color("#666666")).Render(
		fmt.Sprintf("%s - No candlestick data available", c.title))
}

// LineChart representa um gráfico de linha para análise técnica
type LineChart struct {
	series   map[string]*TimeSeries
	width    int
	height   int
	title    string
	colors   map[string]lipgloss.Color
	maxValue float64
	minValue float64
}

// TimeSeries representa uma série de dados temporais
type TimeSeries struct {
	Name   string
	Data   []TimePoint
	Color  lipgloss.Color
}

// TimePoint representa um ponto no tempo
type TimePoint struct {
	Timestamp time.Time
	Value     float64
}

// NewLineChart cria um novo gráfico de linha
func NewLineChart(title string, width, height int) *LineChart {
	return &LineChart{
		series: make(map[string]*TimeSeries),
		width:  width,
		height: height,
		title:  title,
		colors: make(map[string]lipgloss.Color),
	}
}

// AddSeries adiciona uma nova série
func (lc *LineChart) AddSeries(name string, color lipgloss.Color) {
	lc.series[name] = &TimeSeries{
		Name:  name,
		Data:  make([]TimePoint, 0),
		Color: color,
	}
	lc.colors[name] = color
}

// AddDataPoint adiciona um ponto de dados a uma série
func (lc *LineChart) AddDataPoint(seriesName string, timestamp time.Time, value float64) {
	series, exists := lc.series[seriesName]
	if !exists {
		return
	}
	
	point := TimePoint{
		Timestamp: timestamp,
		Value:     value,
	}
	
	series.Data = append(series.Data, point)
	
	// Manter apenas os pontos que cabem na largura
	maxPoints := lc.width - 15
	if len(series.Data) > maxPoints {
		series.Data = series.Data[len(series.Data)-maxPoints:]
	}
	
	lc.updateMinMax()
}

// updateMinMax atualiza valores mínimo e máximo
func (lc *LineChart) updateMinMax() {
	first := true
	
	for _, series := range lc.series {
		for _, point := range series.Data {
			if first {
				lc.minValue = point.Value
				lc.maxValue = point.Value
				first = false
			} else {
				if point.Value < lc.minValue {
					lc.minValue = point.Value
				}
				if point.Value > lc.maxValue {
					lc.maxValue = point.Value
				}
			}
		}
	}
	
	if first {
		lc.minValue = 0
		lc.maxValue = 100
	}
	
	// Adicionar margem
	range_ := lc.maxValue - lc.minValue
	if range_ == 0 {
		range_ = 1
	}
	margin := range_ * 0.05
	lc.minValue -= margin
	lc.maxValue += margin
}

// Render renderiza o gráfico de linha
func (lc *LineChart) Render() string {
	if len(lc.series) == 0 {
		return lc.renderEmpty()
	}
	
	var result strings.Builder
	
	// Título
	result.WriteString(lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("#FFFFFF")).Render(lc.title) + "\n")
	
	// Legenda
	var legendItems []string
	for name, series := range lc.series {
		if len(series.Data) > 0 {
			lastValue := series.Data[len(series.Data)-1].Value
			legendItems = append(legendItems, 
				lipgloss.NewStyle().Foreground(series.Color).Render(fmt.Sprintf("■ %s: %.2f", name, lastValue)))
		}
	}
	result.WriteString(strings.Join(legendItems, " ") + "\n")
	
	// Gráfico
	chart := lc.renderLines()
	result.WriteString(chart)
	
	return result.String()
}

// renderLines renderiza as linhas
func (lc *LineChart) renderLines() string {
	chartHeight := lc.height - 3
	chartWidth := lc.width - 10
	
	var lines []string
	
	for row := 0; row < chartHeight; row++ {
		line := lc.renderLineRow(row, chartHeight, chartWidth)
		lines = append(lines, line)
	}
	
	return strings.Join(lines, "\n")
}

// renderLineRow renderiza uma linha do gráfico
func (lc *LineChart) renderLineRow(row, totalHeight, chartWidth int) string {
	var line strings.Builder
	
	// Eixo Y
	yValue := lc.maxValue - (float64(row)/float64(totalHeight-1))*(lc.maxValue-lc.minValue)
	yLabel := fmt.Sprintf("%8.2f", yValue)
	line.WriteString(lipgloss.NewStyle().Foreground(lipgloss.Color("#666666")).Render(yLabel))
	line.WriteString(" ")
	
	// Dados das séries
	for col := 0; col < chartWidth; col++ {
		var activeChar string
		var activeColor lipgloss.Color
		found := false
		
		// Verificar todas as séries para este ponto
		for _, series := range lc.series {
			if col < len(series.Data) {
				point := series.Data[col]
				
				// Verificar se o valor está próximo da linha atual
				tolerance := (lc.maxValue - lc.minValue) / float64(totalHeight) * 0.5
				if math.Abs(point.Value-yValue) <= tolerance {
					activeChar = "●"
					activeColor = series.Color
					found = true
					break
				}
			}
		}
		
		if found {
			line.WriteString(lipgloss.NewStyle().Foreground(activeColor).Render(activeChar))
		} else {
			// Grid
			if row%3 == 0 && col%10 == 0 {
				line.WriteString(lipgloss.NewStyle().Foreground(lipgloss.Color("#333333")).Render("·"))
			} else {
				line.WriteString(" ")
			}
		}
	}
	
	return line.String()
}

// renderEmpty renderiza gráfico vazio
func (lc *LineChart) renderEmpty() string {
	return lipgloss.NewStyle().Foreground(lipgloss.Color("#666666")).Render(
		fmt.Sprintf("%s - No line data available", lc.title))
}

// VolumeChart representa um gráfico de volume
type VolumeChart struct {
	data     []VolumeData
	width    int
	height   int
	title    string
	maxVolume float64
}

// VolumeData representa dados de volume
type VolumeData struct {
	Timestamp time.Time
	Volume    float64
	Price     float64
}

// NewVolumeChart cria um novo gráfico de volume
func NewVolumeChart(title string, width, height int) *VolumeChart {
	return &VolumeChart{
		data:   make([]VolumeData, 0),
		width:  width,
		height: height,
		title:  title,
	}
}

// AddData adiciona dados de volume
func (vc *VolumeChart) AddData(timestamp time.Time, volume, price float64) {
	data := VolumeData{
		Timestamp: timestamp,
		Volume:    volume,
		Price:     price,
	}
	
	vc.data = append(vc.data, data)
	
	// Manter apenas os dados que cabem na largura
	maxBars := vc.width - 15
	if len(vc.data) > maxBars {
		vc.data = vc.data[len(vc.data)-maxBars:]
	}
	
	// Atualizar volume máximo
	vc.maxVolume = 0
	for _, d := range vc.data {
		if d.Volume > vc.maxVolume {
			vc.maxVolume = d.Volume
		}
	}
}

// Render renderiza o gráfico de volume
func (vc *VolumeChart) Render() string {
	if len(vc.data) == 0 {
		return vc.renderEmpty()
	}
	
	var result strings.Builder
	
	// Título
	result.WriteString(lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("#FFFFFF")).Render(vc.title) + "\n")
	
	// Estatísticas
	totalVolume := 0.0
	for _, d := range vc.data {
		totalVolume += d.Volume
	}
	avgVolume := totalVolume / float64(len(vc.data))
	
	statsLine := fmt.Sprintf("Total: %.0f | Avg: %.0f | Max: %.0f", totalVolume, avgVolume, vc.maxVolume)
	result.WriteString(lipgloss.NewStyle().Foreground(lipgloss.Color("#AAAAAA")).Render(statsLine) + "\n")
	
	// Gráfico
	chart := vc.renderVolumeBars()
	result.WriteString(chart)
	
	return result.String()
}

// renderVolumeBars renderiza as barras de volume
func (vc *VolumeChart) renderVolumeBars() string {
	chartHeight := vc.height - 3
	chartWidth := vc.width - 10
	
	var lines []string
	
	for row := 0; row < chartHeight; row++ {
		line := vc.renderVolumeRow(row, chartHeight, chartWidth)
		lines = append(lines, line)
	}
	
	return strings.Join(lines, "\n")
}

// renderVolumeRow renderiza uma linha do gráfico de volume
func (vc *VolumeChart) renderVolumeRow(row, totalHeight, chartWidth int) string {
	var line strings.Builder
	
	// Eixo Y
	yValue := vc.maxVolume * (1.0 - float64(row)/float64(totalHeight-1))
	yLabel := fmt.Sprintf("%8.0f", yValue)
	line.WriteString(lipgloss.NewStyle().Foreground(lipgloss.Color("#666666")).Render(yLabel))
	line.WriteString(" ")
	
	// Barras de volume
	for i, data := range vc.data {
		if i >= chartWidth {
			break
		}
		
		if data.Volume >= yValue {
			// Determinar cor baseada no preço (verde se subiu, vermelho se desceu)
			var color lipgloss.Color
			if i > 0 && data.Price >= vc.data[i-1].Price {
				color = lipgloss.Color("#00FF88")
			} else {
				color = lipgloss.Color("#FF4444")
			}
			
			line.WriteString(lipgloss.NewStyle().Foreground(color).Render("█"))
		} else {
			line.WriteString(" ")
		}
	}
	
	return line.String()
}

// renderEmpty renderiza gráfico vazio
func (vc *VolumeChart) renderEmpty() string {
	return lipgloss.NewStyle().Foreground(lipgloss.Color("#666666")).Render(
		fmt.Sprintf("%s - No volume data available", vc.title))
}