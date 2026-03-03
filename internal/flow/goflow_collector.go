package flow

import (
	"context"
	"fmt"
	"log"
	"net"
	"sync"
	"time"
)

// GoFlowCollector wraps goflow2 for NetFlow/sFlow/IPFIX collection
type GoFlowCollector struct {
	mu             sync.RWMutex
	config         GoFlowConfig
	flowChan       chan *FlowMessage
	stopCh         chan struct{}
	stats          CollectorStats
	aggregator     *Collector
	bgpCorrelation *BGPCorrelation
}

// GoFlowConfig holds goflow collector configuration
type GoFlowConfig struct {
	NetFlowEnabled bool
	NetFlowAddr    string
	NetFlowPort    int

	SFlowEnabled bool
	SFlowAddr    string
	SFlowPort    int

	IPFIXEnabled bool
	IPFIXAddr    string
	IPFIXPort    int

	Workers       int
	BufferSize    int
	EnableBGPCorr bool
}

// FlowMessage represents a decoded flow message
type FlowMessage struct {
	Type      string    // netflow, sflow, ipfix
	Version   uint16
	SrcAddr   net.IP
	DstAddr   net.IP
	SrcPort   uint16
	DstPort   uint16
	Protocol  uint8
	Bytes     uint64
	Packets   uint64
	SrcAS     uint32
	DstAS     uint32
	NextHop   net.IP
	Timestamp time.Time
	Exporter  net.IP
}

// CollectorStats holds collector statistics
type CollectorStats struct {
	NetFlowPackets   uint64
	SFlowPackets     uint64
	IPFIXPackets     uint64
	TotalFlows       uint64
	DroppedFlows     uint64
	ProcessingErrors uint64
	LastUpdate       time.Time
}

// BGPCorrelation correlates flow data with BGP information
type BGPCorrelation struct {
	mu            sync.RWMutex
	asnPrefixes   map[uint32][]string // ASN -> Prefixes
	prefixASN     map[string]uint32   // Prefix -> ASN
	asnNeighbors  map[uint32][]uint32 // ASN -> Neighbors
	lastUpdate    time.Time
}

// NewGoFlowCollector creates a new goflow-based collector
func NewGoFlowCollector(config GoFlowConfig) (*GoFlowCollector, error) {
	collector := &GoFlowCollector{
		config:   config,
		flowChan: make(chan *FlowMessage, config.BufferSize),
		stopCh:   make(chan struct{}),
		aggregator: NewCollector(CollectorConfig{
			AggregateWindow: 60 * time.Second,
			MaxFlows:        100000,
			EnableAnomaly:   true,
			AnomalyWindow:   300 * time.Second,
		}),
	}

	if config.EnableBGPCorr {
		collector.bgpCorrelation = &BGPCorrelation{
			asnPrefixes:  make(map[uint32][]string),
			prefixASN:    make(map[string]uint32),
			asnNeighbors: make(map[uint32][]uint32),
		}
	}

	return collector, nil
}

// Start starts the goflow collector
func (c *GoFlowCollector) Start(ctx context.Context) error {
	// Start aggregator
	if err := c.aggregator.Start(ctx); err != nil {
		return fmt.Errorf("failed to start aggregator: %w", err)
	}

	// Start NetFlow listener
	if c.config.NetFlowEnabled {
		go c.startNetFlowListener(ctx)
	}

	// Start sFlow listener
	if c.config.SFlowEnabled {
		go c.startSFlowListener(ctx)
	}

	// Start IPFIX listener (uses NetFlow decoder)
	if c.config.IPFIXEnabled {
		go c.startIPFIXListener(ctx)
	}

	// Start flow processor workers
	for i := 0; i < c.config.Workers; i++ {
		go c.processFlows(ctx)
	}

	log.Printf("GoFlow collector started (NetFlow: %v, sFlow: %v, IPFIX: %v)",
		c.config.NetFlowEnabled, c.config.SFlowEnabled, c.config.IPFIXEnabled)

	return nil
}

// Stop stops the collector
func (c *GoFlowCollector) Stop() {
	close(c.stopCh)
	c.aggregator.Stop()
	log.Println("GoFlow collector stopped")
}

// startNetFlowListener starts NetFlow/IPFIX listener
func (c *GoFlowCollector) startNetFlowListener(ctx context.Context) {
	addr := fmt.Sprintf("%s:%d", c.config.NetFlowAddr, c.config.NetFlowPort)
	conn, err := net.ListenUDP("udp", &net.UDPAddr{
		IP:   net.ParseIP(c.config.NetFlowAddr),
		Port: c.config.NetFlowPort,
	})
	if err != nil {
		log.Printf("Failed to start NetFlow listener: %v", err)
		return
	}
	defer conn.Close()

	log.Printf("NetFlow listener started on %s", addr)

	buffer := make([]byte, 9000) // MTU size

	for {
		select {
		case <-ctx.Done():
			return
		case <-c.stopCh:
			return
		default:
			n, remoteAddr, err := conn.ReadFromUDP(buffer)
			if err != nil {
				c.incrementError()
				continue
			}

			c.incrementNetFlowPackets()

			// Decode NetFlow packet
			go c.decodeNetFlow(buffer[:n], remoteAddr.IP)
		}
	}
}

// startSFlowListener starts sFlow listener
func (c *GoFlowCollector) startSFlowListener(ctx context.Context) {
	addr := fmt.Sprintf("%s:%d", c.config.SFlowAddr, c.config.SFlowPort)
	conn, err := net.ListenUDP("udp", &net.UDPAddr{
		IP:   net.ParseIP(c.config.SFlowAddr),
		Port: c.config.SFlowPort,
	})
	if err != nil {
		log.Printf("Failed to start sFlow listener: %v", err)
		return
	}
	defer conn.Close()

	log.Printf("sFlow listener started on %s", addr)

	buffer := make([]byte, 9000)

	for {
		select {
		case <-ctx.Done():
			return
		case <-c.stopCh:
			return
		default:
			n, remoteAddr, err := conn.ReadFromUDP(buffer)
			if err != nil {
				c.incrementError()
				continue
			}

			c.incrementSFlowPackets()

			// Decode sFlow packet
			go c.decodeSFlow(buffer[:n], remoteAddr.IP)
		}
	}
}

// startIPFIXListener starts IPFIX listener
func (c *GoFlowCollector) startIPFIXListener(ctx context.Context) {
	// IPFIX uses same decoder as NetFlow v10
	addr := fmt.Sprintf("%s:%d", c.config.IPFIXAddr, c.config.IPFIXPort)
	conn, err := net.ListenUDP("udp", &net.UDPAddr{
		IP:   net.ParseIP(c.config.IPFIXAddr),
		Port: c.config.IPFIXPort,
	})
	if err != nil {
		log.Printf("Failed to start IPFIX listener: %v", err)
		return
	}
	defer conn.Close()

	log.Printf("IPFIX listener started on %s", addr)

	buffer := make([]byte, 9000)

	for {
		select {
		case <-ctx.Done():
			return
		case <-c.stopCh:
			return
		default:
			n, remoteAddr, err := conn.ReadFromUDP(buffer)
			if err != nil {
				c.incrementError()
				continue
			}

			c.incrementIPFIXPackets()

			// Decode IPFIX packet (same as NetFlow v10)
			go c.decodeNetFlow(buffer[:n], remoteAddr.IP)
		}
	}
}

// decodeNetFlow decodes NetFlow packets
func (c *GoFlowCollector) decodeNetFlow(data []byte, exporter net.IP) {
	// Simplified NetFlow parsing
	// In production, use proper goflow2 decoder or implement full NetFlow parser
	
	// For now, create a basic flow record from the packet
	// This is a placeholder - real implementation would parse NetFlow v5/v9/v10 formats
	
	flowMsg := &FlowMessage{
		Type:      "netflow",
		Version:   5, // Assume v5 for simplicity
		Timestamp: time.Now(),
		Exporter:  exporter,
		// Fields would be extracted from actual NetFlow packet
	}

	select {
	case c.flowChan <- flowMsg:
		c.incrementTotalFlows()
	default:
		c.incrementDropped()
	}
}

// decodeSFlow decodes sFlow packets
func (c *GoFlowCollector) decodeSFlow(data []byte, exporter net.IP) {
	// Simplified sFlow parsing
	// In production, use proper goflow2 decoder or implement full sFlow parser
	
	flowMsg := &FlowMessage{
		Type:      "sflow",
		Version:   5, // Assume v5 for simplicity
		Timestamp: time.Now(),
		Exporter:  exporter,
		// Fields would be extracted from actual sFlow packet
	}

	select {
	case c.flowChan <- flowMsg:
		c.incrementTotalFlows()
	default:
		c.incrementDropped()
	}
}

// processFlows processes flow messages
func (c *GoFlowCollector) processFlows(ctx context.Context) {
	for {
		select {
		case <-ctx.Done():
			return
		case <-c.stopCh:
			return
		case flowMsg := <-c.flowChan:
			c.processFlowMessage(flowMsg)
		}
	}
}

// processFlowMessage processes a single flow message
func (c *GoFlowCollector) processFlowMessage(flowMsg *FlowMessage) {
	// Convert to FlowRecord
	record := FlowRecord{
		SrcAddr:   flowMsg.SrcAddr,
		SrcPort:   flowMsg.SrcPort,
		SrcAS:     flowMsg.SrcAS,
		DstAddr:   flowMsg.DstAddr,
		DstPort:   flowMsg.DstPort,
		DstAS:     flowMsg.DstAS,
		Bytes:     flowMsg.Bytes,
		Packets:   flowMsg.Packets,
		Protocol:  flowMsg.Protocol,
		StartTime: flowMsg.Timestamp,
		EndTime:   flowMsg.Timestamp,
		NextHop:   flowMsg.NextHop,
	}

	// Add to aggregator
	c.aggregator.AddFlow(record)

	// BGP correlation if enabled
	if c.config.EnableBGPCorr && c.bgpCorrelation != nil {
		c.correlateBGP(&record)
	}
}

// correlateBGP correlates flow with BGP data
func (c *GoFlowCollector) correlateBGP(record *FlowRecord) {
	c.bgpCorrelation.mu.RLock()
	defer c.bgpCorrelation.mu.RUnlock()

	// Check if destination ASN has BGP data
	if prefixes, ok := c.bgpCorrelation.asnPrefixes[record.DstAS]; ok {
		// ASN is known in BGP
		_ = prefixes // Use for correlation
	}
}

// GetStats returns collector statistics
func (c *GoFlowCollector) GetStats() CollectorStats {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.stats
}

// Helper methods for statistics
func (c *GoFlowCollector) incrementNetFlowPackets() {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.stats.NetFlowPackets++
	c.stats.LastUpdate = time.Now()
}

func (c *GoFlowCollector) incrementSFlowPackets() {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.stats.SFlowPackets++
	c.stats.LastUpdate = time.Now()
}

func (c *GoFlowCollector) incrementIPFIXPackets() {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.stats.IPFIXPackets++
	c.stats.LastUpdate = time.Now()
}

func (c *GoFlowCollector) incrementTotalFlows() {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.stats.TotalFlows++
}

func (c *GoFlowCollector) incrementDropped() {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.stats.DroppedFlows++
}

func (c *GoFlowCollector) incrementError() {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.stats.ProcessingErrors++
}

// UpdateBGPData updates BGP correlation data
func (c *GoFlowCollector) UpdateBGPData(asn uint32, prefixes []string, neighbors []uint32) {
	if c.bgpCorrelation == nil {
		return
	}

	c.bgpCorrelation.mu.Lock()
	defer c.bgpCorrelation.mu.Unlock()

	c.bgpCorrelation.asnPrefixes[asn] = prefixes
	for _, prefix := range prefixes {
		c.bgpCorrelation.prefixASN[prefix] = asn
	}
	c.bgpCorrelation.asnNeighbors[asn] = neighbors
	c.bgpCorrelation.lastUpdate = time.Now()
}

// GetAggregator returns the underlying aggregator
func (c *GoFlowCollector) GetAggregator() *Collector {
	return c.aggregator
}
