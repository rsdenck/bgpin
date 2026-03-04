package components

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/lipgloss"
)

// SparklineChart representa um gráfico de linha simples
type SparklineChart struct {
	data   []float64
	width  int
	height int
	max    float64
	min    float64
	title  string
	color  lipgloss.Color
}

// NewSparklineChart cria um novo gráfico sparkline
func NewSparklineChart(title string, width, height int) *SparklineChart {
	return &SparklineChart{
		data:   make([]float64, 0),
		width:  width,
		height: height,
		title:  title,
		color:  lipgloss.Color("#00FF00"),
	}
}

// AddData adiciona um ponto de dados
func (s *SparklineChart) AddData(value float64) {
	s.data = append(s.data, value)
	
	// Manter apenas os últimos pontos que cabem na largura
	if len(s.data) > s.width {
		s.data = s.data[1:]
	}
	
	// Atualizar min/max
	s.updateMinMax()
}

// updateMinMax atualiza valores mínimo e máximo
func (s *SparklineChart) updateMinMax() {
	if len(s.data) == 0 {
		return
	}
	
	s.min = s.data[0]
	s.max = s.data[0]
	
	for _, v := range s.data {
		if v < s.min {
			s.min = v
		}
		if v > s.max {
			s.max = v
		}
	}
}

// Render renderiza o gráfico
func (s *SparklineChart) Render() string {
	if len(s.data) == 0 {
		return s.renderEmpty()
	}
	
	var result strings.Builder
	
	// Título
	titleStyle := lipgloss.NewStyle().Bold(true).Foreground(s.color)
	result.WriteString(titleStyle.Render(s.title) + "\n")
	
	// Gráfico
	chart := s.renderChart()
	result.WriteString(chart)
	
	// Estatísticas
	stats := fmt.Sprintf("Min: %.1f | Max: %.1f | Current: %.1f", 
		s.min, s.max, s.data[len(s.data)-1])
	result.WriteString("\n" + lipgloss.NewStyle().Foreground(lipgloss.Color("#666666")).Render(stats))
	
	return result.String()
}

// renderChart renderiza o gráfico principal
func (s *SparklineChart) renderChart() string {
	if s.max == s.min {
		// Dados constantes
		line := strings.Repeat("─", s.width)
		return lipgloss.NewStyle().Foreground(s.color).Render(line)
	}
	
	var chart strings.Builder
	
	// Caracteres para diferentes alturas
	chars := []string{"▁", "▂", "▃", "▄", "▅", "▆", "▇", "█"}
	
	for _, value := range s.data {
		// Normalizar valor para 0-7 (índices dos caracteres)
		normalized := (value - s.min) / (s.max - s.min)
		index := int(normalized * 7)
		if index > 7 {
			index = 7
		}
		if index < 0 {
			index = 0
		}
		
		chart.WriteString(chars[index])
	}
	
	// Preencher espaços vazios
	for len(chart.String()) < s.width {
		chart.WriteString(" ")
	}
	
	return lipgloss.NewStyle().Foreground(s.color).Render(chart.String())
}

// renderEmpty renderiza gráfico vazio
func (s *SparklineChart) renderEmpty() string {
	titleStyle := lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("#666666"))
	emptyLine := strings.Repeat("─", s.width)
	
	return titleStyle.Render(s.title) + "\n" + 
		lipgloss.NewStyle().Foreground(lipgloss.Color("#333333")).Render(emptyLine) + "\n" +
		lipgloss.NewStyle().Foreground(lipgloss.Color("#666666")).Render("No data")
}

// SetColor define a cor do gráfico
func (s *SparklineChart) SetColor(color lipgloss.Color) {
	s.color = color
}

// BarChart representa um gráfico de barras
type BarChart struct {
	data   map[string]float64
	width  int
	height int
	title  string
	colors []lipgloss.Color
}

// NewBarChart cria um novo gráfico de barras
func NewBarChart(title string, width, height int) *BarChart {
	return &BarChart{
		data:   make(map[string]float64),
		width:  width,
		height: height,
		title:  title,
		colors: []lipgloss.Color{
			lipgloss.Color("#00FF00"),
			lipgloss.Color("#0080FF"),
			lipgloss.Color("#FF8000"),
			lipgloss.Color("#FF0080"),
			lipgloss.Color("#80FF00"),
			lipgloss.Color("#8000FF"),
		},
	}
}

// SetData define os dados do gráfico
func (b *BarChart) SetData(data map[string]float64) {
	b.data = data
}

// Render renderiza o gráfico de barras
func (b *BarChart) Render() string {
	if len(b.data) == 0 {
		return b.renderEmpty()
	}
	
	var result strings.Builder
	
	// Título
	titleStyle := lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("#FFFFFF"))
	result.WriteString(titleStyle.Render(b.title) + "\n")
	
	// Encontrar valor máximo
	var maxValue float64
	for _, value := range b.data {
		if value > maxValue {
			maxValue = value
		}
	}
	
	if maxValue == 0 {
		return b.renderEmpty()
	}
	
	// Renderizar barras
	colorIndex := 0
	for label, value := range b.data {
		barWidth := int((value / maxValue) * float64(b.width-20))
		if barWidth < 0 {
			barWidth = 0
		}
		
		color := b.colors[colorIndex%len(b.colors)]
		colorIndex++
		
		// Label (limitado a 15 caracteres)
		labelStr := label
		if len(labelStr) > 15 {
			labelStr = labelStr[:12] + "..."
		}
		labelStr = fmt.Sprintf("%-15s", labelStr)
		
		// Barra
		bar := strings.Repeat("█", barWidth)
		if barWidth == 0 {
			bar = "▏"
		}
		
		// Valor
		valueStr := fmt.Sprintf(" %.1f", value)
		
		line := labelStr + " " + 
			lipgloss.NewStyle().Foreground(color).Render(bar) + 
			valueStr
		
		result.WriteString(line + "\n")
	}
	
	return result.String()
}

// renderEmpty renderiza gráfico vazio
func (b *BarChart) renderEmpty() string {
	titleStyle := lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("#666666"))
	return titleStyle.Render(b.title) + "\n" +
		lipgloss.NewStyle().Foreground(lipgloss.Color("#666666")).Render("No data available")
}

// ProgressBar representa uma barra de progresso
type ProgressBar struct {
	value   float64
	max     float64
	width   int
	label   string
	color   lipgloss.Color
	bgColor lipgloss.Color
}

// NewProgressBar cria uma nova barra de progresso
func NewProgressBar(label string, width int, max float64) *ProgressBar {
	return &ProgressBar{
		label:   label,
		width:   width,
		max:     max,
		color:   lipgloss.Color("#00FF00"),
		bgColor: lipgloss.Color("#333333"),
	}
}

// SetValue define o valor atual
func (p *ProgressBar) SetValue(value float64) {
	p.value = value
	if p.value > p.max {
		p.value = p.max
	}
	if p.value < 0 {
		p.value = 0
	}
}

// SetColor define a cor da barra
func (p *ProgressBar) SetColor(color lipgloss.Color) {
	p.color = color
}

// Render renderiza a barra de progresso
func (p *ProgressBar) Render() string {
	percentage := (p.value / p.max) * 100
	filledWidth := int((p.value / p.max) * float64(p.width))
	
	// Escolher cor baseada na porcentagem
	color := p.color
	if percentage > 80 {
		color = lipgloss.Color("#FF0000") // Vermelho para valores altos
	} else if percentage > 60 {
		color = lipgloss.Color("#FFFF00") // Amarelo para valores médios
	}
	
	// Construir barra
	filled := strings.Repeat("█", filledWidth)
	empty := strings.Repeat("░", p.width-filledWidth)
	
	bar := lipgloss.NewStyle().Foreground(color).Render(filled) +
		lipgloss.NewStyle().Foreground(p.bgColor).Render(empty)
	
	// Label e porcentagem
	label := fmt.Sprintf("%-20s", p.label)
	percent := fmt.Sprintf("%6.1f%%", percentage)
	
	return label + " [" + bar + "] " + percent
}

// Gauge representa um medidor circular (simulado com texto)
type Gauge struct {
	value float64
	max   float64
	label string
	unit  string
	color lipgloss.Color
}

// NewGauge cria um novo medidor
func NewGauge(label, unit string, max float64) *Gauge {
	return &Gauge{
		label: label,
		unit:  unit,
		max:   max,
		color: lipgloss.Color("#00FF00"),
	}
}

// SetValue define o valor atual
func (g *Gauge) SetValue(value float64) {
	g.value = value
}

// Render renderiza o medidor
func (g *Gauge) Render() string {
	percentage := (g.value / g.max) * 100
	
	// Escolher cor baseada na porcentagem
	color := g.color
	if percentage > 80 {
		color = lipgloss.Color("#FF0000")
	} else if percentage > 60 {
		color = lipgloss.Color("#FFFF00")
	}
	
	// Criar representação visual do medidor
	segments := 10
	filled := int((percentage / 100) * float64(segments))
	
	var gauge strings.Builder
	gauge.WriteString("┌")
	for i := 0; i < segments; i++ {
		if i < filled {
			gauge.WriteString("█")
		} else {
			gauge.WriteString("░")
		}
	}
	gauge.WriteString("┐")
	
	valueStr := fmt.Sprintf("%.1f %s", g.value, g.unit)
	percentStr := fmt.Sprintf("(%.1f%%)", percentage)
	
	return lipgloss.NewStyle().Bold(true).Render(g.label) + "\n" +
		lipgloss.NewStyle().Foreground(color).Render(gauge.String()) + "\n" +
		lipgloss.NewStyle().Foreground(color).Render(valueStr) + " " +
		lipgloss.NewStyle().Foreground(lipgloss.Color("#666666")).Render(percentStr)
}

// Table representa uma tabela customizada
type Table struct {
	headers []string
	rows    [][]string
	widths  []int
	title   string
}

// NewTable cria uma nova tabela
func NewTable(title string, headers []string, widths []int) *Table {
	return &Table{
		title:   title,
		headers: headers,
		widths:  widths,
		rows:    make([][]string, 0),
	}
}

// AddRow adiciona uma linha à tabela
func (t *Table) AddRow(row []string) {
	t.rows = append(t.rows, row)
}

// SetRows define todas as linhas da tabela
func (t *Table) SetRows(rows [][]string) {
	t.rows = rows
}

// Render renderiza a tabela
func (t *Table) Render() string {
	if len(t.headers) == 0 {
		return ""
	}
	
	var result strings.Builder
	
	// Título
	if t.title != "" {
		titleStyle := lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("#FFFFFF"))
		result.WriteString(titleStyle.Render(t.title) + "\n")
	}
	
	// Cabeçalho
	headerStyle := lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("#00FF00"))
	var headerLine strings.Builder
	for i, header := range t.headers {
		width := t.widths[i]
		if len(header) > width {
			header = header[:width-3] + "..."
		}
		headerLine.WriteString(fmt.Sprintf("%-*s", width, header))
		if i < len(t.headers)-1 {
			headerLine.WriteString(" │ ")
		}
	}
	result.WriteString(headerStyle.Render(headerLine.String()) + "\n")
	
	// Separador
	var separator strings.Builder
	for i, width := range t.widths {
		separator.WriteString(strings.Repeat("─", width))
		if i < len(t.widths)-1 {
			separator.WriteString("─┼─")
		}
	}
	result.WriteString(lipgloss.NewStyle().Foreground(lipgloss.Color("#444444")).Render(separator.String()) + "\n")
	
	// Linhas
	for _, row := range t.rows {
		var rowLine strings.Builder
		for i, cell := range row {
			if i >= len(t.widths) {
				break
			}
			width := t.widths[i]
			if len(cell) > width {
				cell = cell[:width-3] + "..."
			}
			rowLine.WriteString(fmt.Sprintf("%-*s", width, cell))
			if i < len(t.headers)-1 && i < len(row)-1 {
				rowLine.WriteString(" │ ")
			}
		}
		result.WriteString(rowLine.String() + "\n")
	}
	
	return result.String()
}