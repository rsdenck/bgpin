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
	if c == nil {
		return nil
	}
	if c.cancel != nil {
		c.cancel()
	}
	if c.conn != nil {
		return c.conn.Close()
	}
	return nil
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
	if c == nil || c.client == nil {
		return fmt.Errorf("BGP client not connected")
	}
	
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
	if c == nil || c.client == nil {
		return fmt.Errorf("BGP client not connected")
	}
	
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

// IsConnected returns true if the BGP client is connected
func (c *BGPClient) IsConnected() bool {
	return c != nil && c.client != nil && c.conn != nil
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

// GetRealPeers returns real peer data from GoBGP daemon
func (c *BGPClient) GetRealPeers() ([]*PeerInfo, error) {
	if c == nil || c.client == nil {
		return nil, fmt.Errorf("BGP client not connected")
	}
	
	req := &api.ListPeerRequest{}
	
	stream, err := c.client.ListPeer(c.ctx, req)
	if err != nil {
		return nil, fmt.Errorf("failed to list peers from GoBGP: %w", err)
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
		
		// Calculate uptime from real data
		var uptime time.Duration
		if peer.Timers != nil && peer.Timers.State != nil && peer.Timers.State.Uptime != nil {
			uptime = time.Since(peer.Timers.State.Uptime.AsTime())
		}
		
		// Get real peer state
		state := "Unknown"
		if peer.State != nil {
			state = peer.State.SessionState.String()
		}
		
		// Get real counters
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
			Accepted:    accepted, // TODO: Get actual accepted count from RIB
			Advertised:  advertised,
			Description: peer.Conf.Description,
			Flaps:       0, // TODO: Get real flap count from statistics
			LastError:   "",
		}
		
		peers = append(peers, peerInfo)
	}
	
	return peers, nil
}

// GetRealRoutes returns real route data from GoBGP RIB
func (c *BGPClient) GetRealRoutes(family api.Family) ([]*RouteInfo, error) {
	if c == nil || c.client == nil {
		return nil, fmt.Errorf("BGP client not connected")
	}
	
	req := &api.ListPathRequest{
		TableType: api.TableType_GLOBAL,
		Family:    &family,
	}
	
	stream, err := c.client.ListPath(c.ctx, req)
	if err != nil {
		return nil, fmt.Errorf("failed to list routes from GoBGP RIB: %w", err)
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
			
			// Parse real NLRI to get prefix
			prefix := ""
			if destination.Prefix != "" {
				prefix = destination.Prefix
			}
			
			// Get real next hop from path attributes
			nextHop := ""
			// TODO: Parse next hop from path attributes
			
			// Get real AS path from path attributes
			var asPath []uint32
			// TODO: Parse AS path from path attributes
			
			// Calculate real age
			var age time.Duration
			if path.Age != nil {
				age = time.Since(path.Age.AsTime())
			}
			
			routeInfo := &RouteInfo{
				Prefix:    prefix,
				NextHop:   nextHop,
				ASPath:    asPath,
				Origin:    "IGP", // TODO: Parse real origin from attributes
				MED:       0,     // TODO: Parse real MED from attributes
				LocalPref: 100,   // TODO: Parse real local preference
				Community: []string{}, // TODO: Parse real communities
				Age:       age,
				Best:      path.Best,
				Valid:     true, // TODO: Determine real validity from RPKI
			}
			
			routes = append(routes, routeInfo)
		}
	}
	
	return routes, nil
}