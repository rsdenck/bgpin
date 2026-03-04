package gobgp

import (
	"context"
	"fmt"
	"log"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	
	api "github.com/osrg/gobgp/v3/api"
)

// BGPClient represents a GoBGP client
type BGPClient struct {
	conn   *grpc.ClientConn
	client api.GobgpApiClient
	ctx    context.Context
	cancel context.CancelFunc
}

// PeerInfo represents BGP peer information
type PeerInfo struct {
	ASN         uint32
	RouterID    string
	RemoteAddr  string
	State       string
	Uptime      time.Duration
	Received    uint32
	Accepted    uint32
	Advertised  uint32
	Description string
	Flaps       uint32
	LastError   string
}

// RouteInfo represents BGP route information
type RouteInfo struct {
	Prefix     string
	NextHop    string
	ASPath     []uint32
	Origin     string
	MED        uint32
	LocalPref  uint32
	Community  []string
	Age        time.Duration
	Best       bool
	Valid      bool
}

// FlowInfo represents flow information
type FlowInfo struct {
	SrcAddr    string
	DstAddr    string
	SrcPort    uint32
	DstPort    uint32
	Protocol   string
	Bytes      uint64
	Packets    uint64
	Duration   time.Duration
	Flags      []string
}

// NewBGPClient creates a new GoBGP client
func NewBGPClient(address string) (*BGPClient, error) {
	ctx, cancel := context.WithCancel(context.Background())
	
	// Connect to GoBGP daemon
	conn, err := grpc.DialContext(
		ctx,
		address,
		grpc.WithTransportCredentials(insecure.NewCredentials()),
		grpc.WithBlock(),
		grpc.WithTimeout(5*time.Second),
	)
	if err != nil {
		cancel()
		return nil, fmt.Errorf("failed to connect to GoBGP: %w", err)
	}
	
	client := api.NewGobgpApiClient(conn)
	
	return &BGPClient{
		conn:   conn,
		client: client,
		ctx:    ctx,
		cancel: cancel,
	}, nil
}

// Close closes the BGP client connection
func (c *BGPClient) Close() error {
	c.cancel()
	return c.conn.Close()
}

// GetPeers returns all BGP peers
func (c *BGPClient) GetPeers() ([]*PeerInfo, error) {
	req := &api.ListPeerRequest{}
	
	stream, err := c.client.ListPeer(c.ctx, req)
	if err != nil {
		return nil, fmt.Errorf("failed to list peers: %w", err)
	}
	
	var peers []*PeerInfo
	
	for {
		resp, err := stream.Recv()
		if err != nil {
			break
		}
		
		peer := resp.Peer
		if peer == nil {
			continue
		}
		
		// Calculate uptime
		var uptime time.Duration
		if peer.Timers != nil && peer.Timers.State != nil && peer.Timers.State.Uptime != nil {
			uptime = time.Since(peer.Timers.State.Uptime.AsTime())
		}
		
		// Get peer state
		state := "Unknown"
		if peer.State != nil {
			state = peer.State.SessionState.String()
		}
		
		// Get counters
		var received, accepted, advertised uint32
		if peer.State != nil && peer.State.Messages != nil {
			if peer.State.Messages.Received != nil {
				received = uint32(peer.State.Messages.Received.Update)
			}
			if peer.State.Messages.Sent != nil {
				advertised = uint32(peer.State.Messages.Sent.Update)
			}
		}
		
		peerInfo := &PeerInfo{
			ASN:         peer.Conf.PeerAsn,
			RouterID:    peer.State.RouterId,
			RemoteAddr:  peer.State.NeighborAddress,
			State:       state,
			Uptime:      uptime,
			Received:    received,
			Accepted:    accepted, // TODO: Get actual accepted count
			Advertised:  advertised,
			Description: peer.Conf.Description,
			Flaps:       0, // TODO: Get flap count
			LastError:   "",
		}
		
		peers = append(peers, peerInfo)
	}
	
	return peers, nil
}

// GetRoutes returns BGP routes
func (c *BGPClient) GetRoutes(family api.Family) ([]*RouteInfo, error) {
	req := &api.ListPathRequest{
		TableType: api.TableType_GLOBAL,
		Family:    &family,
	}
	
	stream, err := c.client.ListPath(c.ctx, req)
	if err != nil {
		return nil, fmt.Errorf("failed to list routes: %w", err)
	}
	
	var routes []*RouteInfo
	
	for {
		resp, err := stream.Recv()
		if err != nil {
			break
		}
		
		destination := resp.Destination
		if destination == nil {
			continue
		}
		
		for _, path := range destination.Paths {
			if path == nil {
				continue
			}
			
			// Parse NLRI to get prefix
			prefix := ""
			if destination.Prefix != "" {
				prefix = destination.Prefix
			}
			
			// Get next hop
			nextHop := ""
			// TODO: Parse next hop from path attributes
			
			// Get AS path
			var asPath []uint32
			// TODO: Parse AS path from path attributes
			
			// Calculate age
			var age time.Duration
			if path.Age != nil {
				age = time.Since(path.Age.AsTime())
			}
			
			routeInfo := &RouteInfo{
				Prefix:    prefix,
				NextHop:   nextHop,
				ASPath:    asPath,
				Origin:    "IGP", // TODO: Parse origin
				MED:       0,     // TODO: Parse MED
				LocalPref: 100,   // TODO: Parse local preference
				Community: []string{}, // TODO: Parse communities
				Age:       age,
				Best:      path.Best,
				Valid:     true, // TODO: Determine validity
			}
			
			routes = append(routes, routeInfo)
		}
	}
	
	return routes, nil
}

// WatchPeers watches for peer state changes
func (c *BGPClient) WatchPeers(callback func(*PeerInfo, string)) error {
	req := &api.WatchEventRequest{
		Peer: &api.WatchEventRequest_Peer{},
	}
	
	stream, err := c.client.WatchEvent(c.ctx, req)
	if err != nil {
		return fmt.Errorf("failed to watch peers: %w", err)
	}
	
	go func() {
		for {
			resp, err := stream.Recv()
			if err != nil {
				log.Printf("Error receiving peer event: %v", err)
				return
			}
			
			if peerEvent := resp.GetPeer(); peerEvent != nil {
				peer := peerEvent.Peer
				if peer == nil {
					continue
				}
				
				// Convert to PeerInfo
				peerInfo := &PeerInfo{
					ASN:        peer.Conf.PeerAsn,
					RouterID:   peer.State.RouterId,
					RemoteAddr: peer.State.NeighborAddress,
					State:      peer.State.SessionState.String(),
				}
				
				// Determine event type
				eventType := "state_change"
				
				callback(peerInfo, eventType)
			}
		}
	}()
	
	return nil
}

// WatchRoutes watches for route changes
func (c *BGPClient) WatchRoutes(callback func(*RouteInfo, string)) error {
	req := &api.WatchEventRequest{
		Table: &api.WatchEventRequest_Table{
			Filters: []*api.WatchEventRequest_Table_Filter{
				{
					Type: api.WatchEventRequest_Table_Filter_BEST,
				},
			},
		},
	}
	
	stream, err := c.client.WatchEvent(c.ctx, req)
	if err != nil {
		return fmt.Errorf("failed to watch routes: %w", err)
	}
	
	go func() {
		for {
			resp, err := stream.Recv()
			if err != nil {
				log.Printf("Error receiving route event: %v", err)
				return
			}
			
			if tableEvent := resp.GetTable(); tableEvent != nil {
				// TODO: Parse table event and convert to RouteInfo
				routeInfo := &RouteInfo{
					Prefix: "example", // TODO: Parse from event
				}
				
				eventType := "route_change"
				callback(routeInfo, eventType)
			}
		}
	}()
	
	return nil
}

// GetGlobalConfig returns global BGP configuration
func (c *BGPClient) GetGlobalConfig() (*api.Global, error) {
	req := &api.GetBgpRequest{}
	
	resp, err := c.client.GetBgp(c.ctx, req)
	if err != nil {
		return nil, fmt.Errorf("failed to get global config: %w", err)
	}
	
	return resp.Global, nil
}

// SearchPath searches for paths containing a specific IP
func (c *BGPClient) SearchPath(ip string) ([]*RouteInfo, error) {
	// TODO: Implement path search functionality
	// This would search through all routes to find paths that contain the specified IP
	
	routes, err := c.GetRoutes(api.Family{
		Afi:  api.Family_AFI_IP,
		Safi: api.Family_SAFI_UNICAST,
	})
	if err != nil {
		return nil, err
	}
	
	var matchingRoutes []*RouteInfo
	for _, route := range routes {
		// TODO: Implement IP matching logic
		if route.Prefix == ip || route.NextHop == ip {
			matchingRoutes = append(matchingRoutes, route)
		}
	}
	
	return matchingRoutes, nil
}

// MockBGPClient creates a mock BGP client for testing
func MockBGPClient() *BGPClient {
	return &BGPClient{
		// Mock implementation for development/testing
	}
}

// GetMockPeers returns mock peer data for testing
func (c *BGPClient) GetMockPeers() []*PeerInfo {
	return []*PeerInfo{
		{
			ASN:         15169,
			RouterID:    "8.8.8.8",
			RemoteAddr:  "8.8.8.8",
			State:       "Established",
			Uptime:      2 * time.Hour,
			Received:    1500,
			Accepted:    1450,
			Advertised:  800,
			Description: "Google Public DNS",
			Flaps:       0,
		},
		{
			ASN:         13335,
			RouterID:    "1.1.1.1",
			RemoteAddr:  "1.1.1.1",
			State:       "Established",
			Uptime:      4 * time.Hour,
			Received:    2200,
			Accepted:    2100,
			Advertised:  1200,
			Description: "Cloudflare DNS",
			Flaps:       1,
		},
		{
			ASN:         64512,
			RouterID:    "192.168.1.1",
			RemoteAddr:  "192.168.1.1",
			State:       "Idle",
			Uptime:      0,
			Received:    0,
			Accepted:    0,
			Advertised:  0,
			Description: "Local Peer",
			Flaps:       5,
		},
	}
}

// GetMockRoutes returns mock route data for testing
func (c *BGPClient) GetMockRoutes() []*RouteInfo {
	return []*RouteInfo{
		{
			Prefix:    "8.8.8.0/24",
			NextHop:   "8.8.8.8",
			ASPath:    []uint32{15169},
			Origin:    "IGP",
			MED:       0,
			LocalPref: 100,
			Community: []string{"15169:1000"},
			Age:       30 * time.Minute,
			Best:      true,
			Valid:     true,
		},
		{
			Prefix:    "1.1.1.0/24",
			NextHop:   "1.1.1.1",
			ASPath:    []uint32{13335},
			Origin:    "IGP",
			MED:       0,
			LocalPref: 100,
			Community: []string{"13335:1000"},
			Age:       45 * time.Minute,
			Best:      true,
			Valid:     true,
		},
		{
			Prefix:    "192.168.1.0/24",
			NextHop:   "192.168.1.1",
			ASPath:    []uint32{64512},
			Origin:    "IGP",
			MED:       10,
			LocalPref: 90,
			Community: []string{},
			Age:       10 * time.Minute,
			Best:      false,
			Valid:     false,
		},
	}
}