package cisco

import (
	"bufio"
	"context"
	"fmt"
	"regexp"
	"strconv"
	"strings"

	"github.com/bgpin/bgpin/internal/adapters/ssh"
	"github.com/bgpin/bgpin/internal/core/bgp"
)

// Parser handles Cisco IOS/IOS-XE/IOS-XR parsing
type Parser struct {
	client *ssh.Client
	vendor string // ios, ios-xe, ios-xr, nxos
}

// Config holds Cisco parser configuration
type Config struct {
	Host     string
	Port     int
	Username string
	Password string
	Vendor   string // ios, ios-xe, ios-xr, nxos
}

// NewParser creates a new Cisco parser
func NewParser(config Config) (*Parser, error) {
	client, err := ssh.NewClient(ssh.Config{
		Host:     config.Host,
		Port:     config.Port,
		Username: config.Username,
		Password: config.Password,
	})
	if err != nil {
		return nil, err
	}

	vendor := config.Vendor
	if vendor == "" {
		vendor = "ios" // Default
	}

	return &Parser{
		client: client,
		vendor: vendor,
	}, nil
}

// Connect establishes SSH connection
func (p *Parser) Connect(ctx context.Context) error {
	return p.client.Connect(ctx)
}

// Close closes the connection
func (p *Parser) Close() error {
	return p.client.Close()
}

// GetBGPNeighbors retrieves BGP neighbors
func (p *Parser) GetBGPNeighbors(ctx context.Context) ([]bgp.Neighbor, error) {
	var cmd string
	switch p.vendor {
	case "ios-xr":
		cmd = "show bgp neighbors"
	case "nxos":
		cmd = "show bgp all neighbors"
	default:
		cmd = "show ip bgp neighbors"
	}

	output, err := p.client.ExecuteCommand(ctx, cmd)
	if err != nil {
		return nil, fmt.Errorf("failed to execute command: %w", err)
	}

	return p.parseBGPNeighbors(output), nil
}

// parseBGPNeighbors parses BGP neighbors from CLI output
func (p *Parser) parseBGPNeighbors(output string) []bgp.Neighbor {
	neighbors := make([]bgp.Neighbor, 0)
	
	// Regex patterns for different vendors
	var neighborPattern *regexp.Regexp
	switch p.vendor {
	case "ios-xr":
		neighborPattern = regexp.MustCompile(`BGP neighbor is ([0-9.]+)`)
	default:
		neighborPattern = regexp.MustCompile(`BGP neighbor is ([0-9.]+),\s+remote AS (\d+)`)
	}

	statePattern := regexp.MustCompile(`BGP state = (\w+)`)
	prefixPattern := regexp.MustCompile(`Prefixes Total:\s+(\d+)`)
	
	scanner := bufio.NewScanner(strings.NewReader(output))
	var currentNeighbor *bgp.Neighbor
	
	for scanner.Scan() {
		line := scanner.Text()
		
		// Match neighbor line
		if matches := neighborPattern.FindStringSubmatch(line); matches != nil {
			if currentNeighbor != nil {
				neighbors = append(neighbors, *currentNeighbor)
			}
			
			currentNeighbor = &bgp.Neighbor{
				IP: matches[1],
			}
			
			if len(matches) > 2 {
				asn, _ := strconv.Atoi(matches[2])
				currentNeighbor.AS = asn
			}
		}
		
		if currentNeighbor == nil {
			continue
		}
		
		// Match state
		if matches := statePattern.FindStringSubmatch(line); matches != nil {
			currentNeighbor.State = matches[1]
		}
		
		// Match prefix count
		if matches := prefixPattern.FindStringSubmatch(line); matches != nil {
			count, _ := strconv.ParseUint(matches[1], 10, 64)
			currentNeighbor.MessagesRecv = count
		}
	}
	
	if currentNeighbor != nil {
		neighbors = append(neighbors, *currentNeighbor)
	}
	
	return neighbors
}

// GetBGPRoutes retrieves BGP routes for a prefix
func (p *Parser) GetBGPRoutes(ctx context.Context, prefix string) ([]bgp.Route, error) {
	var cmd string
	switch p.vendor {
	case "ios-xr":
		cmd = fmt.Sprintf("show bgp %s", prefix)
	case "nxos":
		cmd = fmt.Sprintf("show bgp all %s", prefix)
	default:
		cmd = fmt.Sprintf("show ip bgp %s", prefix)
	}

	output, err := p.client.ExecuteCommand(ctx, cmd)
	if err != nil {
		return nil, fmt.Errorf("failed to execute command: %w", err)
	}

	return p.parseBGPRoutes(output), nil
}

// parseBGPRoutes parses BGP routes from CLI output
func (p *Parser) parseBGPRoutes(output string) []bgp.Route {
	routes := make([]bgp.Route, 0)
	
	// Pattern for route line: *> prefix via nexthop, metric, locpref, path
	routePattern := regexp.MustCompile(`([*>si ]+)\s+([0-9./]+)\s+([0-9.]+)`)
	asPathPattern := regexp.MustCompile(`\s+(\d+(?:\s+\d+)*)$`)
	
	scanner := bufio.NewScanner(strings.NewReader(output))
	
	for scanner.Scan() {
		line := scanner.Text()
		
		if matches := routePattern.FindStringSubmatch(line); matches != nil {
			route := bgp.Route{
				Prefix:  matches[2],
				NextHop: matches[3],
			}
			
			// Extract AS_PATH
			if asMatches := asPathPattern.FindStringSubmatch(line); asMatches != nil {
				asPath := p.parseASPath(asMatches[1])
				route.ASPath = make([]int, len(asPath))
				for i, as := range asPath {
					route.ASPath[i] = int(as)
				}
			}
			
			routes = append(routes, route)
		}
	}
	
	return routes
}

// parseASPath parses AS_PATH string to slice
func (p *Parser) parseASPath(asPath string) []uint32 {
	parts := strings.Fields(asPath)
	path := make([]uint32, 0, len(parts))
	
	for _, part := range parts {
		asn, err := strconv.ParseUint(part, 10, 32)
		if err == nil {
			path = append(path, uint32(asn))
		}
	}
	
	return path
}

// GetBGPSummary retrieves BGP summary
func (p *Parser) GetBGPSummary(ctx context.Context) (string, error) {
	var cmd string
	switch p.vendor {
	case "ios-xr":
		cmd = "show bgp summary"
	case "nxos":
		cmd = "show bgp all summary"
	default:
		cmd = "show ip bgp summary"
	}

	return p.client.ExecuteCommand(ctx, cmd)
}

// GetVersion retrieves device version
func (p *Parser) GetVersion(ctx context.Context) (string, error) {
	return p.client.ExecuteCommand(ctx, "show version")
}
