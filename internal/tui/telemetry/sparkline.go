package telemetry

import (
	"fmt"
	"math"
	"strings"
	"time"

	"github.com/charmbracelet/lipgloss"
)

// SparklineData represents a data point in a sparkline
type SparklineData struct {
	Value     float64
	Timestamp time.Time
}

// Sparkline represents a sparkline chart
type Sparkline struct {
	data     []SparklineData
	maxSize  int
	minValue float64
	maxValue float64
	width    int
	title    string
	unit     string
}

// NewSparkline creates a new sparkline
func NewSparkline(title, unit string, width, maxSize int) *Sparkline {
	return &Sparkline{
		data:     make([]SparklineData, 0, maxSize),
		maxSize:  maxSize,
		width:    width,
		title:    title,
		unit:     unit,
		minValue: math.Inf(1),
		maxValue: math.Inf(-1),
	}
}

// AddData adds a new data point
func (s *Sparkline) AddData(value float64) {
	now := time.Now()
	
	// Add new data point
	s.data = append(s.data, SparklineData{
		Value:     value,
		Timestamp: now,
	})
	
	// Remove old data if exceeding max size
	if len(s.data) > s.maxSize {
		s.data = s.data[1:]
	}
	
	// Update min/max values
	s.updateMinMax()
}

// updateMinMax updates the min and max values
func (s *Sparkline) updateMinMax() {
	if len(s.data) == 0 {
		return
	}
	
	s.minValue = math.Inf(1)
	s.maxValue = math.Inf(-1)
	
	for _, point := range s.data {
		if point.Value < s.minValue {
			s.minValue = point.Value
		}
		if point.Value > s.maxValue {
			s.maxValue = point.Value
		}
	}
	
	// Ensure we have a range
	if s.minValue == s.maxValue {
		s.minValue -= 1
		s.maxValue += 1
	}
}

// Render renders the sparkline
func (s *Sparkline) Render() string {
	if len(s.data) == 0 {
		return s.renderEmpty()
	}
	
	// Sparkline characters (from low to high)
	sparkChars := []rune{'▁', '▂', '▃', '▄', '▅', '▆', '▇', '█'}
	
	var sparkline strings.Builder
	
	// Calculate how many data points to show
	dataPoints := min(len(s.data), s.width-2) // Reserve space for borders
	if dataPoints == 0 {
		return s.renderEmpty()
	}
	
	// Get the most recent data points
	startIdx := len(s.data) - dataPoints
	
	for i := startIdx; i < len(s.data); i++ {
		value := s.data[i].Value
		
		// Normalize value to 0-1 range
		normalized := (value - s.minValue) / (s.maxValue - s.minValue)
		
		// Map to character index
		charIdx := int(normalized * float64(len(sparkChars)-1))
		if charIdx < 0 {
			charIdx = 0
		}
		if charIdx >= len(sparkChars) {
			charIdx = len(sparkChars) - 1
		}
		
		sparkline.WriteRune(sparkChars[charIdx])
	}
	
	// Style the sparkline
	sparkStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("#00FFFF"))
	
	titleStyle := lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color("#FFFFFF"))
	
	valueStyle := lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color("#00FF00"))
	
	// Get current value
	currentValue := s.data[len(s.data)-1].Value
	
	// Format the output
	title := titleStyle.Render(s.title)
	spark := sparkStyle.Render(sparkline.String())
	value := valueStyle.Render(fmt.Sprintf("%.1f%s", currentValue, s.unit))
	
	return fmt.Sprintf("%s %s %s", title, spark, value)
}

// renderEmpty renders empty sparkline
func (s *Sparkline) renderEmpty() string {
	style := lipgloss.NewStyle().
		Foreground(lipgloss.Color("#666666"))
	
	return style.Render(fmt.Sprintf("%s: No data", s.title))
}

// GetStats returns statistics about the sparkline data
func (s *Sparkline) GetStats() SparklineStats {
	if len(s.data) == 0 {
		return SparklineStats{}
	}
	
	// Calculate average
	sum := 0.0
	for _, point := range s.data {
		sum += point.Value
	}
	avg := sum / float64(len(s.data))
	
	// Calculate trend (simple linear regression slope)
	trend := s.calculateTrend()
	
	return SparklineStats{
		Current: s.data[len(s.data)-1].Value,
		Min:     s.minValue,
		Max:     s.maxValue,
		Average: avg,
		Trend:   trend,
		Points:  len(s.data),
	}
}

// SparklineStats represents statistics about sparkline data
type SparklineStats struct {
	Current float64
	Min     float64
	Max     float64
	Average float64
	Trend   float64 // Positive = increasing, Negative = decreasing
	Points  int
}

// calculateTrend calculates the trend using simple linear regression
func (s *Sparkline) calculateTrend() float64 {
	if len(s.data) < 2 {
		return 0
	}
	
	n := float64(len(s.data))
	sumX := 0.0
	sumY := 0.0
	sumXY := 0.0
	sumX2 := 0.0
	
	for i, point := range s.data {
		x := float64(i)
		y := point.Value
		
		sumX += x
		sumY += y
		sumXY += x * y
		sumX2 += x * x
	}
	
	// Calculate slope (trend)
	slope := (n*sumXY - sumX*sumY) / (n*sumX2 - sumX*sumX)
	
	return slope
}

// TelemetryManager manages multiple sparklines
type TelemetryManager struct {
	sparklines map[string]*Sparkline
	width      int
}

// NewTelemetryManager creates a new telemetry manager
func NewTelemetryManager(width int) *TelemetryManager {
	return &TelemetryManager{
		sparklines: make(map[string]*Sparkline),
		width:      width,
	}
}

// AddSparkline adds a new sparkline
func (tm *TelemetryManager) AddSparkline(key, title, unit string, maxSize int) {
	tm.sparklines[key] = NewSparkline(title, unit, tm.width/4, maxSize)
}

// UpdateData updates data for a sparkline
func (tm *TelemetryManager) UpdateData(key string, value float64) {
	if sparkline, exists := tm.sparklines[key]; exists {
		sparkline.AddData(value)
	}
}

// RenderAll renders all sparklines
func (tm *TelemetryManager) RenderAll() string {
	if len(tm.sparklines) == 0 {
		return "No telemetry data available"
	}
	
	var lines []string
	
	// Render sparklines in a specific order
	order := []string{"traffic", "routes", "neighbors", "latency", "cpu", "memory"}
	
	for _, key := range order {
		if sparkline, exists := tm.sparklines[key]; exists {
			lines = append(lines, sparkline.Render())
		}
	}
	
	// Add any remaining sparklines
	for key, sparkline := range tm.sparklines {
		found := false
		for _, orderedKey := range order {
			if key == orderedKey {
				found = true
				break
			}
		}
		if !found {
			lines = append(lines, sparkline.Render())
		}
	}
	
	return strings.Join(lines, "\n")
}

// GetSparkline returns a sparkline by key
func (tm *TelemetryManager) GetSparkline(key string) *Sparkline {
	return tm.sparklines[key]
}