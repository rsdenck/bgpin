package graph

import (
	"fmt"
	"math"
	"strings"

	"github.com/charmbracelet/lipgloss"
)

// Node represents a BGP AS node in the graph
type Node struct {
	ASN         int
	Name        string
	Status      NodeStatus
	Traffic     float64 // Mbps
	Prefixes    int
	Latency     float64 // ms
	Position    Position
	Connections []int // Connected ASNs
}

// NodeStatus represents the status of a BGP node
type NodeStatus int

const (
	StatusEstablished NodeStatus = iota
	StatusIdle
	StatusConnect
	StatusDown
)

// Position represents a node position in the graph
type Position struct {
	X, Y int
}

// ASPathGraph represents the AS-PATH visualization graph
type ASPathGraph struct {
	nodes       map[int]*Node
	centerASN   int
	width       int
	height      int
	selectedASN int
	zoom        float64
}

// NewASPathGraph creates a new AS-PATH graph
func NewASPathGraph(centerASN int, width, height int) *ASPathGraph {
	return &ASPathGraph{
		nodes:       make(map[int]*Node),
		centerASN:   centerASN,
		width:       width,
		height:      height,
		selectedASN: centerASN,
		zoom:        1.0,
	}
}

// AddNode adds a node to the graph
func (g *ASPathGraph) AddNode(asn int, name string, status NodeStatus, traffic float64, prefixes int, latency float64) {
	g.nodes[asn] = &Node{
		ASN:      asn,
		Name:     name,
		Status:   status,
		Traffic:  traffic,
		Prefixes: prefixes,
		Latency:  latency,
	}
	g.calculatePositions()
}

// AddConnection adds a connection between two ASNs
func (g *ASPathGraph) AddConnection(from, to int) {
	if node, exists := g.nodes[from]; exists {
		node.Connections = append(node.Connections, to)
	}
}

// SetSelected sets the selected ASN
func (g *ASPathGraph) SetSelected(asn int) {
	if _, exists := g.nodes[asn]; exists {
		g.selectedASN = asn
	}
}

// Render renders the AS-PATH graph
func (g *ASPathGraph) Render() string {
	if len(g.nodes) == 0 {
		return g.renderEmpty()
	}

	// Create canvas
	canvas := make([][]rune, g.height)
	for i := range canvas {
		canvas[i] = make([]rune, g.width)
		for j := range canvas[i] {
			canvas[i][j] = ' '
		}
	}

	// Draw connections first (behind nodes)
	g.drawConnections(canvas)

	// Draw nodes
	g.drawNodes(canvas)

	// Convert canvas to string with styling
	return g.canvasToString(canvas)
}

// calculatePositions calculates node positions in a circular layout
func (g *ASPathGraph) calculatePositions() {
	if len(g.nodes) == 0 {
		return
	}

	centerX := g.width / 2
	centerY := g.height / 2

	// Place center ASN in the middle
	if centerNode, exists := g.nodes[g.centerASN]; exists {
		centerNode.Position = Position{X: centerX, Y: centerY}
	}

	// Place other nodes in concentric circles
	otherNodes := make([]*Node, 0)
	for asn, node := range g.nodes {
		if asn != g.centerASN {
			otherNodes = append(otherNodes, node)
		}
	}

	if len(otherNodes) == 0 {
		return
	}

	// Calculate radius based on available space
	radius := int(float64(min(g.width, g.height)) * 0.3 * g.zoom)
	
	// Distribute nodes in a circle
	angleStep := 2 * math.Pi / float64(len(otherNodes))
	
	for i, node := range otherNodes {
		angle := float64(i) * angleStep
		x := centerX + int(float64(radius)*math.Cos(angle))
		y := centerY + int(float64(radius)*math.Sin(angle))
		
		// Ensure positions are within bounds
		x = max(2, min(g.width-10, x))
		y = max(1, min(g.height-2, y))
		
		node.Position = Position{X: x, Y: y}
	}
}

// drawConnections draws connections between nodes
func (g *ASPathGraph) drawConnections(canvas [][]rune) {
	for _, node := range g.nodes {
		for _, connectedASN := range node.Connections {
			if connectedNode, exists := g.nodes[connectedASN]; exists {
				g.drawLine(canvas, node.Position, connectedNode.Position, node.Traffic)
			}
		}
	}
}

// drawLine draws a line between two positions with thickness based on traffic
func (g *ASPathGraph) drawLine(canvas [][]rune, from, to Position, traffic float64) {
	// Simple line drawing algorithm
	dx := abs(to.X - from.X)
	dy := abs(to.Y - from.Y)
	
	x, y := from.X, from.Y
	xInc := 1
	yInc := 1
	
	if to.X < from.X {
		xInc = -1
	}
	if to.Y < from.Y {
		yInc = -1
	}
	
	// Choose line character based on traffic volume
	var lineChar rune
	if traffic > 1000 { // > 1 Gbps
		lineChar = '━' // Thick line
	} else if traffic > 100 { // > 100 Mbps
		lineChar = '─' // Medium line
	} else {
		lineChar = '·' // Thin line
	}
	
	if dx > dy {
		err := dx / 2
		for x != to.X {
			if y >= 0 && y < g.height && x >= 0 && x < g.width {
				canvas[y][x] = lineChar
			}
			err -= dy
			if err < 0 {
				y += yInc
				err += dx
			}
			x += xInc
		}
	} else {
		err := dy / 2
		for y != to.Y {
			if y >= 0 && y < g.height && x >= 0 && x < g.width {
				canvas[y][x] = lineChar
			}
			err -= dx
			if err < 0 {
				x += xInc
				err += dy
			}
			y += yInc
		}
	}
}

// drawNodes draws nodes on the canvas
func (g *ASPathGraph) drawNodes(canvas [][]rune) {
	for asn, node := range g.nodes {
		g.drawNode(canvas, node, asn == g.selectedASN, asn == g.centerASN)
	}
}

// drawNode draws a single node
func (g *ASPathGraph) drawNode(canvas [][]rune, node *Node, selected, center bool) {
	x, y := node.Position.X, node.Position.Y
	
	// Choose node character based on status and type
	var nodeChar rune
	if center {
		nodeChar = '◉' // Center node
	} else if selected {
		nodeChar = '◎' // Selected node
	} else {
		switch node.Status {
		case StatusEstablished:
			nodeChar = '●' // Established
		case StatusIdle:
			nodeChar = '○' // Idle
		case StatusConnect:
			nodeChar = '◐' // Connecting
		case StatusDown:
			nodeChar = '✕' // Down
		}
	}
	
	// Draw the node
	if y >= 0 && y < g.height && x >= 0 && x < g.width {
		canvas[y][x] = nodeChar
	}
	
	// Draw ASN label
	asnLabel := fmt.Sprintf("AS%d", node.ASN)
	labelX := x - len(asnLabel)/2
	labelY := y + 1
	
	if labelY < g.height && labelX >= 0 {
		for i, char := range asnLabel {
			if labelX+i < g.width && labelX+i >= 0 {
				canvas[labelY][labelX+i] = char
			}
		}
	}
}

// canvasToString converts canvas to styled string
func (g *ASPathGraph) canvasToString(canvas [][]rune) string {
	var result strings.Builder
	
	// Styles for different elements
	establishedStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("#00FF00"))
	idleStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("#FF0000"))
	connectStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("#FFFF00"))
	selectedStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("#FF00FF")).Bold(true)
	centerStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("#00FFFF")).Bold(true)
	lineStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("#666666"))
	
	for y, row := range canvas {
		for x, char := range row {
			if char == ' ' {
				result.WriteRune(char)
				continue
			}
			
			// Find which node this character belongs to
			var nodeStyle lipgloss.Style
			found := false
			
			for asn, node := range g.nodes {
				if (node.Position.X == x && node.Position.Y == y) ||
				   (node.Position.Y+1 == y && x >= node.Position.X-5 && x <= node.Position.X+5) {
					
					if asn == g.centerASN {
						nodeStyle = centerStyle
					} else if asn == g.selectedASN {
						nodeStyle = selectedStyle
					} else {
						switch node.Status {
						case StatusEstablished:
							nodeStyle = establishedStyle
						case StatusIdle:
							nodeStyle = idleStyle
						case StatusConnect:
							nodeStyle = connectStyle
						default:
							nodeStyle = idleStyle
						}
					}
					found = true
					break
				}
			}
			
			if !found {
				nodeStyle = lineStyle
			}
			
			result.WriteString(nodeStyle.Render(string(char)))
		}
		if y < len(canvas)-1 {
			result.WriteRune('\n')
		}
	}
	
	return result.String()
}

// renderEmpty renders empty graph message
func (g *ASPathGraph) renderEmpty() string {
	style := lipgloss.NewStyle().
		Width(g.width).
		Height(g.height).
		Align(lipgloss.Center, lipgloss.Center).
		Foreground(lipgloss.Color("#666666"))
	
	return style.Render("No BGP peers available\nConnecting to GoBGP...")
}

// GetSelectedNode returns the currently selected node
func (g *ASPathGraph) GetSelectedNode() *Node {
	if node, exists := g.nodes[g.selectedASN]; exists {
		return node
	}
	return nil
}

// GetNodeDetails returns detailed information about a node
func (g *ASPathGraph) GetNodeDetails(asn int) string {
	node, exists := g.nodes[asn]
	if !exists {
		return "Node not found"
	}
	
	statusStr := "Unknown"
	switch node.Status {
	case StatusEstablished:
		statusStr = "Established"
	case StatusIdle:
		statusStr = "Idle"
	case StatusConnect:
		statusStr = "Connect"
	case StatusDown:
		statusStr = "Down"
	}
	
	return fmt.Sprintf(`AS%d - %s
Status: %s
Traffic: %.1f Mbps
Prefixes: %d
Latency: %.1f ms
Connections: %d`,
		node.ASN, node.Name, statusStr, node.Traffic, 
		node.Prefixes, node.Latency, len(node.Connections))
}

// Helper functions
func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}

func abs(x int) int {
	if x < 0 {
		return -x
	}
	return x
}