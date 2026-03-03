package flow

import (
	"net"
	"time"
)

// FlowRecord represents a network flow record
type FlowRecord struct {
	// Source information
	SrcAddr net.IP
	SrcPort uint16
	SrcAS   uint32

	// Destination information
	DstAddr net.IP
	DstPort uint16
	DstAS   uint32

	// Flow metrics
	Bytes      uint64
	Packets    uint64
	Protocol   uint8
	TCPFlags   uint8
	StartTime  time.Time
	EndTime    time.Time
	SampleRate uint32

	// BGP information
	NextHop    net.IP
	ASPath     []uint32
	Community  []uint32
	LocalPref  uint32
	MED        uint32

	// Additional metadata
	InputIface  uint32
	OutputIface uint32
	ExporterIP  net.IP
}

// FlowStats represents aggregated flow statistics
type FlowStats struct {
	Prefix       string
	ASN          uint32
	TotalBytes   uint64
	TotalPackets uint64
	FlowCount    uint64
	BPS          float64 // Bits per second
	PPS          float64 // Packets per second
	TopProtocols map[uint8]uint64
	TopPorts     map[uint16]uint64
	StartTime    time.Time
	EndTime      time.Time
}

// ASNTraffic represents traffic statistics for an ASN
type ASNTraffic struct {
	ASN            uint32
	InboundBPS     float64
	OutboundBPS    float64
	InboundPPS     float64
	OutboundPPS    float64
	InboundBytes   uint64
	OutboundBytes  uint64
	InboundPackets uint64
	OutboundPackets uint64
	InboundFlows   uint64
	OutboundFlows  uint64
	TotalBytes     uint64
	TotalPackets   uint64
	TopPrefixes    []PrefixTraffic
	TopPeers       []PeerTraffic
}

// PrefixTraffic represents traffic for a specific prefix
type PrefixTraffic struct {
	Prefix       string
	Bytes        uint64
	Packets      uint64
	BPS          float64
	PPS          float64
	TopProtocols map[uint8]uint64
}

// PeerTraffic represents traffic with a peer ASN
type PeerTraffic struct {
	PeerASN uint32
	Bytes   uint64
	Packets uint64
	BPS     float64
	PPS     float64
}

// Anomaly represents a detected traffic anomaly
type Anomaly struct {
	Type        string    // ddos, spike, drop, unusual_protocol
	Severity    string    // low, medium, high, critical
	Description string
	Prefix      string
	DstAddr     net.IP
	DstAS       uint32
	ASN         uint32
	Metric      string // bps, pps, flow_count
	Baseline    float64
	Current     float64
	Threshold   float64
	DetectedAt  time.Time
	Duration    time.Duration
}

// UpstreamComparison compares traffic across multiple upstreams
type UpstreamComparison struct {
	Prefix    string
	Upstreams []UpstreamStats
	TotalBPS  float64
	TotalPPS  float64
}

// UpstreamStats represents statistics for a single upstream
type UpstreamStats struct {
	Provider string
	ASN      uint32
	ASPath   []uint32
	BPS      float64
	PPS      float64
	Latency  time.Duration
	Loss     float64 // Packet loss percentage
	Jitter   time.Duration
}

// Protocol names mapping
var ProtocolNames = map[uint8]string{
	1:   "ICMP",
	6:   "TCP",
	17:  "UDP",
	47:  "GRE",
	50:  "ESP",
	51:  "AH",
	58:  "ICMPv6",
	89:  "OSPF",
	132: "SCTP",
}

// GetProtocolName returns the protocol name for a given number
func GetProtocolName(proto uint8) string {
	if name, ok := ProtocolNames[proto]; ok {
		return name
	}
	return "Unknown"
}
