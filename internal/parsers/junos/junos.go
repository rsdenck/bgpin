package junos

import (
	"context"
	"encoding/xml"
	"fmt"
	"strings"

	"github.com/bgpin/bgpin/internal/adapters/netconf"
	"github.com/bgpin/bgpin/internal/core/bgp"
)

// Parser handles Juniper JunOS parsing
type Parser struct {
	client *netconf.Client
}

// Config holds JunOS parser configuration
type Config struct {
	Host     string
	Port     int
	Username string
	Password string
}

// NewParser creates a new JunOS parser
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

// Connect establishes connection to JunOS device
func (p *Parser) Connect(ctx context.Context) error {
	return p.client.Connect(ctx)
}

// Close closes the connection
func (p *Parser) Close() error {
	return p.client.Close()
}

// BGPNeighborInformation represents JunOS BGP neighbor XML structure
type BGPNeighborInformation struct {
	XMLName   xml.Name    `xml:"bgp-information"`
	Neighbors []BGPPeer   `xml:"bgp-peer"`
}

// BGPPeer represents a BGP peer
type BGPPeer struct {
	PeerAddress    string `xml:"peer-address"`
	PeerAS         uint32 `xml:"peer-as"`
	PeerState      string `xml:"peer-state"`
	LocalAddress   string `xml:"local-address"`
	LocalAS        uint32 `xml:"local-as"`
	PeerType       string `xml:"peer-type"`
	ActivePrefixes int    `xml:"bgp-rib>active-prefix-count"`
	ReceivedPrefixes int  `xml:"bgp-rib>received-prefix-count"`
}

// GetBGPNeighbors retrieves BGP neighbors from JunOS
func (p *Parser) GetBGPNeighbors(ctx context.Context) ([]bgp.Neighbor, error) {
	rpc := `<rpc><get-bgp-neighbor-information/></rpc>`
	output, err := p.client.ExecuteRPC(ctx, rpc)
	if err != nil {
		return nil, fmt.Errorf("failed to get BGP neighbors: %w", err)
	}

	var info BGPNeighborInformation
	if err := xml.Unmarshal([]byte(output), &info); err != nil {
		return nil, fmt.Errorf("failed to parse XML: %w", err)
	}

	neighbors := make([]bgp.Neighbor, 0, len(info.Neighbors))
	for _, peer := range info.Neighbors {
		neighbor := bgp.Neighbor{
			IP:           peer.PeerAddress,
			AS:           int(peer.PeerAS),
			State:        peer.PeerState,
			MessagesRecv: uint64(peer.ReceivedPrefixes),
		}
		neighbors = append(neighbors, neighbor)
	}

	return neighbors, nil
}

// RouteInformation represents JunOS route XML structure
type RouteInformation struct {
	XMLName xml.Name      `xml:"route-information"`
	Routes  []RouteEntry  `xml:"route-table>rt"`
}

// RouteEntry represents a route entry
type RouteEntry struct {
	Destination string        `xml:"rt-destination"`
	Prefix      string        `xml:"rt-prefix-length"`
	Entries     []RouteDetail `xml:"rt-entry"`
}

// RouteDetail represents route details
type RouteDetail struct {
	Protocol   string   `xml:"protocol-name"`
	Preference int      `xml:"preference"`
	ASPath     string   `xml:"as-path"`
	NextHop    string   `xml:"nh>to"`
	LocalPref  int      `xml:"local-preference"`
	MED        int      `xml:"med"`
	Communities []string `xml:"communities>community"`
}

// GetBGPRoutes retrieves BGP routes for a prefix
func (p *Parser) GetBGPRoutes(ctx context.Context, prefix string) ([]bgp.Route, error) {
	rpc := fmt.Sprintf(`<rpc><get-route-information><destination>%s</destination></get-route-information></rpc>`, prefix)
	output, err := p.client.ExecuteRPC(ctx, rpc)
	if err != nil {
		return nil, fmt.Errorf("failed to get routes: %w", err)
	}

	var info RouteInformation
	if err := xml.Unmarshal([]byte(output), &info); err != nil {
		return nil, fmt.Errorf("failed to parse XML: %w", err)
	}

	routes := make([]bgp.Route, 0)
	for _, routeEntry := range info.Routes {
		for _, detail := range routeEntry.Entries {
			if detail.Protocol != "BGP" {
				continue
			}

			asPath := p.parseASPath(detail.ASPath)
			asPathInt := make([]int, len(asPath))
			for i, as := range asPath {
				asPathInt[i] = int(as)
			}
			
			route := bgp.Route{
				Prefix:    fmt.Sprintf("%s/%s", routeEntry.Destination, routeEntry.Prefix),
				NextHop:   detail.NextHop,
				ASPath:    asPathInt,
				LocalPref: detail.LocalPref,
				MED:       detail.MED,
				Community: detail.Communities,
			}
			routes = append(routes, route)
		}
	}

	return routes, nil
}

// parseASPath parses AS_PATH string to slice
func (p *Parser) parseASPath(asPath string) []uint32 {
	parts := strings.Fields(asPath)
	path := make([]uint32, 0, len(parts))
	
	for _, part := range parts {
		var asn uint32
		fmt.Sscanf(part, "%d", &asn)
		if asn > 0 {
			path = append(path, asn)
		}
	}
	
	return path
}

// determinePeerType determines peer relationship type
func (p *Parser) determinePeerType(peerType string) string {
	switch strings.ToLower(peerType) {
	case "internal":
		return "ibgp"
	case "external":
		return "ebgp"
	default:
		return peerType
	}
}

// GetVersion retrieves JunOS version
func (p *Parser) GetVersion(ctx context.Context) (string, error) {
	rpc := `<rpc><get-software-information/></rpc>`
	return p.client.ExecuteRPC(ctx, rpc)
}
