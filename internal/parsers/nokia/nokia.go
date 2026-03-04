package nokia

import (
	"bufio"
	"context"
	"encoding/xml"
	"fmt"
	"regexp"
	"strconv"
	"strings"

	"github.com/bgpin/bgpin/internal/adapters/netconf"
	"github.com/bgpin/bgpin/internal/core/bgp"
)

type Parser struct {
	client *netconf.Client
}

type Config struct {
	Host     string
	Port     int
	Username string
	Password string
}

func NewParser(config Config) (*Parser, error) {
	client, err := netconf.NewClient(netconf.Config{
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
	rpc := `<rpc><get-bgp-neighbor-information/></rpc>`
	output, err := p.client.ExecuteRPC(ctx, rpc)
	if err != nil {
		return nil, fmt.Errorf("failed to get BGP neighbors: %w", err)
	}

	return p.parseBGPNeighbors(output), nil
}

func (p *Parser) parseBGPNeighbors(output string) []bgp.Neighbor {
	neighbors := make([]bgp.Neighbor, 0)

	neighborPattern := regexp.MustCompile(`peer-address="([^"]+)"`)
	asPattern := regexp.MustCompile(`peer-as="(\d+)"`)
	statePattern := regexp.MustCompile(`peer-state="([^"]+)"`)

	scanner := bufio.NewScanner(strings.NewReader(output))

	for scanner.Scan() {
		line := scanner.Text()

		ipMatch := neighborPattern.FindStringSubmatch(line)
		asMatch := asPattern.FindStringSubmatch(line)
		stateMatch := statePattern.FindStringSubmatch(line)

		if ipMatch != nil {
			neighbor := bgp.Neighbor{
				IP: ipMatch[1],
			}

			if asMatch != nil {
				asn, _ := strconv.Atoi(asMatch[1])
				neighbor.AS = asn
			}

			if stateMatch != nil {
				neighbor.State = stateMatch[1]
			}

			neighbors = append(neighbors, neighbor)
		}
	}

	return neighbors
}

func (p *Parser) GetBGPRoutes(ctx context.Context, prefix string) ([]bgp.Route, error) {
	rpc := fmt.Sprintf(`<rpc><get-route-information><destination>%s</destination></get-route-information></rpc>`, prefix)
	output, err := p.client.ExecuteRPC(ctx, rpc)
	if err != nil {
		return nil, fmt.Errorf("failed to get routes: %w", err)
	}

	return p.parseBGPRoutes(output), nil
}

func (p *Parser) parseBGPRoutes(output string) []bgp.Route {
	routes := make([]bgp.Route, 0)

	routePattern := regexp.MustCompile(`rt-destination="([^"]+)"\s+rt-prefix-length="(\d+)"`)
	nextHopPattern := regexp.MustCompile(`to="([^"]+)"`)
	asPathPattern := regexp.MustCompile(`as-path="([^"]*)"`)
	localPrefPattern := regexp.MustCompile(`local-preference="(\d+)"`)
	medPattern := regexp.MustCompile(`med="(\d+)"`)

	scanner := bufio.NewScanner(strings.NewReader(output))
	var currentRoute *bgp.Route

	for scanner.Scan() {
		line := scanner.Text()

		if routeMatch := routePattern.FindStringSubmatch(line); routeMatch != nil {
			if currentRoute != nil {
				routes = append(routes, *currentRoute)
			}

			currentRoute = &bgp.Route{
				Prefix: routeMatch[1] + "/" + routeMatch[2],
			}
		}

		if currentRoute == nil {
			continue
		}

		if nhMatch := nextHopPattern.FindStringSubmatch(line); nhMatch != nil {
			currentRoute.NextHop = nhMatch[1]
		}

		if asMatch := asPathPattern.FindStringSubmatch(line); asMatch != nil {
			currentRoute.ASPath = parseASPath(asMatch[1])
		}

		if lpMatch := localPrefPattern.FindStringSubmatch(line); lpMatch != nil {
			lp, _ := strconv.Atoi(lpMatch[1])
			currentRoute.LocalPref = lp
		}

		if medMatch := medPattern.FindStringSubmatch(line); medMatch != nil {
			med, _ := strconv.Atoi(medMatch[1])
			currentRoute.MED = med
		}
	}

	if currentRoute != nil {
		routes = append(routes, *currentRoute)
	}

	return routes
}

func parseASPath(asPath string) []int {
	if asPath == "" {
		return nil
	}

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
	rpc := `<rpc><get-bgp-summary/></rpc>`
	return p.client.ExecuteRPC(ctx, rpc)
}

func (p *Parser) GetVersion(ctx context.Context) (string, error) {
	rpc := `<rpc><get-system-information/></rpc>`
	return p.client.ExecuteRPC(ctx, rpc)
}

func (p *Parser) GetVRFList(ctx context.Context) ([]string, error) {
	rpc := `<rpc><get-router-list/></rpc>`
	output, err := p.client.ExecuteRPC(ctx, rpc)
	if err != nil {
		return nil, err
	}

	vrfs := make([]string, 0)
	vrfPattern := regexp.MustCompile(`router-name="([^"]+)"`)

	scanner := bufio.NewScanner(strings.NewReader(output))
	for scanner.Scan() {
		line := scanner.Text()
		if match := vrfPattern.FindStringSubmatch(line); match != nil {
			vrfs = append(vrfs, match[1])
		}
	}

	return vrfs, nil
}

type BGPNeighborXML struct {
	XMLName   xml.Name `xml:"bgp-information"`
	Neighbors []struct {
		PeerAddress string `xml:"peer-address"`
		PeerAS      uint32 `xml:"peer-as"`
		PeerState   string `xml:"peer-state"`
	} `xml:"bgp-peer"`
}
