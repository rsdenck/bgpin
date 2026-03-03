package flow

import (
	"context"
	"fmt"
	"sync"
	"time"
)

// Collector collects and aggregates flow data
type Collector struct {
	mu            sync.RWMutex
	flows         []FlowRecord
	stats         map[string]*FlowStats // key: prefix
	asnStats      map[uint32]*ASNTraffic
	anomalies     []Anomaly
	config        CollectorConfig
	stopCh        chan struct{}
	aggregateCh   chan FlowRecord
}

// CollectorConfig holds collector configuration
type CollectorConfig struct {
	ListenAddr      string
	AggregateWindow time.Duration
	MaxFlows        int
	EnableAnomaly   bool
	AnomalyWindow   time.Duration
}

// NewCollector creates a new flow collector
func NewCollector(config CollectorConfig) *Collector {
	return &Collector{
		flows:       make([]FlowRecord, 0),
		stats:       make(map[string]*FlowStats),
		asnStats:    make(map[uint32]*ASNTraffic),
		anomalies:   make([]Anomaly, 0),
		config:      config,
		stopCh:      make(chan struct{}),
		aggregateCh: make(chan FlowRecord, 1000),
	}
}

// Start starts the collector
func (c *Collector) Start(ctx context.Context) error {
	// Start aggregation goroutine
	go c.aggregateLoop(ctx)

	// Start anomaly detection if enabled
	if c.config.EnableAnomaly {
		go c.anomalyDetectionLoop(ctx)
	}

	return nil
}

// Stop stops the collector
func (c *Collector) Stop() {
	close(c.stopCh)
}

// AddFlow adds a flow record to the collector
func (c *Collector) AddFlow(flow FlowRecord) {
	select {
	case c.aggregateCh <- flow:
	default:
		// Channel full, drop flow
	}
}

// aggregateLoop aggregates flows periodically
func (c *Collector) aggregateLoop(ctx context.Context) {
	ticker := time.NewTicker(c.config.AggregateWindow)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-c.stopCh:
			return
		case flow := <-c.aggregateCh:
			c.processFlow(flow)
		case <-ticker.C:
			c.aggregateStats()
		}
	}
}

// processFlow processes a single flow record
func (c *Collector) processFlow(flow FlowRecord) {
	c.mu.Lock()
	defer c.mu.Unlock()

	// Add to flows list
	c.flows = append(c.flows, flow)

	// Limit flows in memory
	if len(c.flows) > c.config.MaxFlows {
		c.flows = c.flows[len(c.flows)-c.config.MaxFlows:]
	}
}

// aggregateStats aggregates flow statistics
func (c *Collector) aggregateStats() {
	c.mu.Lock()
	defer c.mu.Unlock()

	// Reset stats
	c.stats = make(map[string]*FlowStats)
	c.asnStats = make(map[uint32]*ASNTraffic)

	// Aggregate by prefix and ASN
	for _, flow := range c.flows {
		// Aggregate by destination prefix (simplified)
		prefix := flow.DstAddr.String() + "/32"
		if stats, ok := c.stats[prefix]; ok {
			stats.TotalBytes += flow.Bytes
			stats.TotalPackets += flow.Packets
			stats.FlowCount++
			stats.TopProtocols[flow.Protocol]++
		} else {
			c.stats[prefix] = &FlowStats{
				Prefix:       prefix,
				ASN:          flow.DstAS,
				TotalBytes:   flow.Bytes,
				TotalPackets: flow.Packets,
				FlowCount:    1,
				TopProtocols: map[uint8]uint64{flow.Protocol: 1},
				TopPorts:     make(map[uint16]uint64),
				StartTime:    flow.StartTime,
				EndTime:      flow.EndTime,
			}
		}

		// Aggregate by ASN
		if asnStats, ok := c.asnStats[flow.DstAS]; ok {
			asnStats.TotalBytes += flow.Bytes
			asnStats.TotalPackets += flow.Packets
		} else {
			c.asnStats[flow.DstAS] = &ASNTraffic{
				ASN:          flow.DstAS,
				TotalBytes:   flow.Bytes,
				TotalPackets: flow.Packets,
				TopPrefixes:  make([]PrefixTraffic, 0),
				TopPeers:     make([]PeerTraffic, 0),
			}
		}
	}

	// Calculate rates
	for _, stats := range c.stats {
		duration := stats.EndTime.Sub(stats.StartTime).Seconds()
		if duration > 0 {
			stats.BPS = float64(stats.TotalBytes*8) / duration
			stats.PPS = float64(stats.TotalPackets) / duration
		}
	}

	for _, asnStats := range c.asnStats {
		// Calculate rates (simplified)
		asnStats.InboundBPS = float64(asnStats.TotalBytes * 8)
		asnStats.InboundPPS = float64(asnStats.TotalPackets)
	}
}

// anomalyDetectionLoop detects traffic anomalies
func (c *Collector) anomalyDetectionLoop(ctx context.Context) {
	ticker := time.NewTicker(c.config.AnomalyWindow)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-c.stopCh:
			return
		case <-ticker.C:
			c.detectAnomalies()
		}
	}
}

// detectAnomalies detects traffic anomalies
func (c *Collector) detectAnomalies() {
	c.mu.RLock()
	defer c.mu.RUnlock()

	// Simple anomaly detection based on thresholds
	for prefix, stats := range c.stats {
		// Detect high traffic
		if stats.BPS > 1e9 { // > 1 Gbps
			anomaly := Anomaly{
				Type:        "spike",
				Severity:    "high",
				Description: fmt.Sprintf("High traffic detected on prefix %s", prefix),
				Prefix:      prefix,
				ASN:         stats.ASN,
				Metric:      "bps",
				Current:     stats.BPS,
				Threshold:   1e9,
				DetectedAt:  time.Now(),
			}
			c.anomalies = append(c.anomalies, anomaly)
		}

		// Detect high packet rate
		if stats.PPS > 100000 { // > 100k pps
			anomaly := Anomaly{
				Type:        "ddos",
				Severity:    "critical",
				Description: fmt.Sprintf("Potential DDoS detected on prefix %s", prefix),
				Prefix:      prefix,
				ASN:         stats.ASN,
				Metric:      "pps",
				Current:     stats.PPS,
				Threshold:   100000,
				DetectedAt:  time.Now(),
			}
			c.anomalies = append(c.anomalies, anomaly)
		}
	}
}

// GetStats returns current flow statistics
func (c *Collector) GetStats() map[string]*FlowStats {
	c.mu.RLock()
	defer c.mu.RUnlock()

	// Return copy
	stats := make(map[string]*FlowStats)
	for k, v := range c.stats {
		stats[k] = v
	}
	return stats
}

// GetASNStats returns ASN statistics
func (c *Collector) GetASNStats(asn uint32) *ASNTraffic {
	c.mu.RLock()
	defer c.mu.RUnlock()

	if stats, ok := c.asnStats[asn]; ok {
		return stats
	}
	return nil
}

// GetAnomalies returns detected anomalies
func (c *Collector) GetAnomalies() []Anomaly {
	c.mu.RLock()
	defer c.mu.RUnlock()

	// Return copy
	anomalies := make([]Anomaly, len(c.anomalies))
	copy(anomalies, c.anomalies)
	return anomalies
}

// GetTopPrefixes returns top prefixes by traffic
func (c *Collector) GetTopPrefixes(limit int) []FlowStats {
	c.mu.RLock()
	defer c.mu.RUnlock()

	// Convert map to slice
	stats := make([]FlowStats, 0, len(c.stats))
	for _, s := range c.stats {
		stats = append(stats, *s)
	}

	// Sort by bytes (simplified - should use sort.Slice)
	// TODO: Implement proper sorting

	if len(stats) > limit {
		stats = stats[:limit]
	}

	return stats
}
