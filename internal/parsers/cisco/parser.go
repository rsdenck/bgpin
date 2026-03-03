package cisco

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"

	"github.com/bgpin/bgpin/internal/core/bgp"
)

// Parser handles Cisco IOS BGP output parsing
type Parser struct {
	routeRegex *regexp.Regexp
}

// NewParser creates a new Cisco parser
func NewParser() *Parser {
	return &Parser{
		routeRegex: regexp.MustCompile(`\*?\s*(\d+\.\d+\.\d+\.\d+/\d+)\s+(\d+\.\d+\.\d+\.\d+)`),
	}
}

// ParseRoutes parses Cisco BGP output into routes
func (p *Parser) ParseRoutes(output string) ([]bgp.Route, error) {
	var routes []bgp.Route
	
	lines := strings.Split(output, "\n")
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" || strings.HasPrefix(line, "BGP") {
			continue
		}
		
		matches := p.routeRegex.FindStringSubmatch(line)
		if len(matches) >= 3 {
			route := bgp.Route{
				Prefix:  matches[1],
				NextHop: matches[2],
				Best:    strings.HasPrefix(line, "*"),
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
	// Simple AS path extraction - can be enhanced
	asPathRegex := regexp.MustCompile(`\s+(\d+(?:\s+\d+)*)\s*$`)
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
