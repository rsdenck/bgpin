package arista

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

type Parser struct {
	client *ssh.Client
}

type Config struct {
	Host     string
	Port     int
	Username string
	Password string
}

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

	return &Parser{
		client: client,
	}, nil
}

func (p *Parser) Connect(ctx context.Context) error {
	return p.client.Connect(ctx)
}

func (p *Parser) Close() error {
	return p.client.Close()
}

func (p *Parser) GetBGPNeighbors(ctx context.Context) ([]bgp.Neighbor, error) {
	cmd := "show bgp neighbors"
	output, err := p.client.ExecuteCommand(ctx, cmd)
	if err != nil {
		return nil, fmt.Errorf("failed to execute command: %w", err)
	}

	return p.parseBGPNeighbors(output), nil
}

func (p *Parser) parseBGPNeighbors(output string) []bgp.Neighbor {
	neighbors := make([]bgp.Neighbor, 0)

	neighborPattern := regexp.MustCompile(`BGP neighbor is ([0-9.]+), Remote AS (\d+)`)
	statePattern := regexp.MustCompile(`BGP state = (\w+)`)
	prefixPattern := regexp.MustCompile(`(\d+) received prefixes`)

	scanner := bufio.NewScanner(strings.NewReader(output))
	var currentNeighbor *bgp.Neighbor

	for scanner.Scan() {
		line := scanner.Text()

		if matches := neighborPattern.FindStringSubmatch(line); matches != nil {
			if currentNeighbor != nil {
				neighbors = append(neighbors, *currentNeighbor)
			}

			asn, _ := strconv.Atoi(matches[2])
			currentNeighbor = &bgp.Neighbor{
				IP: matches[1],
				AS: asn,
			}
		}

		if currentNeighbor == nil {
			continue
		}

		if matches := statePattern.FindStringSubmatch(line); matches != nil {
			currentNeighbor.State = matches[1]
		}

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

func (p *Parser) GetBGPRoutes(ctx context.Context, prefix string) ([]bgp.Route, error) {
	cmd := fmt.Sprintf("show ip bgp %s", prefix)
	output, err := p.client.ExecuteCommand(ctx, cmd)
	if err != nil {
		return nil, fmt.Errorf("failed to execute command: %w", err)
	}

	return p.parseBGPRoutes(output), nil
}

func (p *Parser) parseBGPRoutes(output string) []bgp.Route {
	routes := make([]bgp.Route, 0)

	routePattern := regexp.MustCompile(`([*> ]+)\s+([0-9./]+)\s+([0-9.]+)\s+(\d+)\s+(\d+)`)
	asPathPattern := regexp.MustCompile(`(\d+(?:\s+\d+)*)`)

	scanner := bufio.NewScanner(strings.NewReader(output))

	for scanner.Scan() {
		line := scanner.Text()

		if matches := routePattern.FindStringSubmatch(line); matches != nil {
			route := bgp.Route{
				Prefix:  matches[2],
				NextHop: matches[3],
			}

			localPref, _ := strconv.Atoi(matches[4])
			route.LocalPref = localPref

			med, _ := strconv.Atoi(matches[5])
			route.MED = med

			if strings.Contains(matches[1], "*") {
				route.Best = true
			}

			if asMatches := asPathPattern.FindStringSubmatch(line); asMatches != nil {
				asPath := parseASPath(asMatches[1])
				route.ASPath = asPath
			}

			routes = append(routes, route)
		}
	}

	return routes
}

func parseASPath(asPath string) []int {
	parts := strings.Fields(asPath)
	path := make([]int, 0, len(parts))

	for _, part := range parts {
		asn, err := strconv.Atoi(part)
		if err == nil && asn > 0 {
			path = append(path, asn)
		}
	}

	return path
}

func (p *Parser) GetBGPSummary(ctx context.Context) (string, error) {
	return p.client.ExecuteCommand(ctx, "show ip bgp summary")
}

func (p *Parser) GetVersion(ctx context.Context) (string, error) {
	return p.client.ExecuteCommand(ctx, "show version")
}

func (p *Parser) GetVrfList(ctx context.Context) ([]string, error) {
	output, err := p.client.ExecuteCommand(ctx, "show vrf")
	if err != nil {
		return nil, err
	}

	vrfs := make([]string, 0)
	scanner := bufio.NewScanner(strings.NewReader(output))

	for scanner.Scan() {
		line := scanner.Text()
		if strings.Contains(line, "default") {
			vrfs = append(vrfs, "default")
		}
	}

	return vrfs, nil
}
