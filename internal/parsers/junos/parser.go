package junos

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"

	"github.com/bgpin/bgpin/internal/core/bgp"
)

// Parser handles Juniper JunOS BGP output parsing
type Parser struct {
	routeRegex *regexp.Regexp
}

// NewParser creates a new Juniper parser
func NewParser() *Parser {
	return &Parser{
		routeRegex: regexp.MustCompile(`(\d+\.\d+\.\d+\.\d+/\d+)\s+\*?\[BGP.*?\]\s+(\d+\.\d+\.\d+\.\d+)`),
	}
}

// ParseRoutes parses Juniper BGP output into routes
func (p *Parser) ParseRoutes(output string) ([]bgp.Route, error) {
	var routes []bgp.Route
	
	lines := strings.Split(output, "\n")
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" || strings.HasPrefix(line, "inet.0:") {
			continue
		}
		
		matches := p.routeRegex.FindStringSubmatch(line)
		if len(matches) >= 3 {
			route := bgp.Route{
				Prefix:  matches[1],
				NextHop: matches[2],
				Best:    strings.Contains(line, "*"),
				Valid:   true,
			}
			
			// Parse AS path if present
			if asPath := p.extractASPath(line); len(asPath) > 0 {
				route.ASPath = asPath
			}
			
			routes = append(routes, route)
		}
	}
	
	if len(routes) == 0 {
		return nil, fmt.Errorf("no routes found in output")
	}
	
	return routes, nil
}

func (p *Parser) extractASPath(line string) []int {
	// Extract AS path from Juniper format
	asPathRegex := regexp.MustCompile(`AS path:\s+(\d+(?:\s+\d+)*)`)
	matches := asPathRegex.FindStringSubmatch(line)
	if len(matches) < 2 {
		return nil
	}
	
	asns := strings.Fields(matches[1])
	var path []int
	for _, asn := range asns {
		if num, err := strconv.Atoi(asn); err == nil {
			path = append(path, num)
		}
	}
	
	return path
}
