package sdk

import "time"

// ASNNeighbors represents ASN neighbors information
type ASNNeighbors struct {
	ASN       int                `json:"asn"`
	Neighbors []NeighborRelation `json:"neighbours"`
	QueryTime time.Time          `json:"query_time"`
}

// NeighborRelation represents a BGP neighbor relationship
type NeighborRelation struct {
	ASN   int    `json:"asn"`
	Type  string `json:"type"` // "left" or "right"
	Power int    `json:"power"`
}

// PrefixOverview represents prefix information
type PrefixOverview struct {
	Prefix       string    `json:"prefix"`
	ASNs         []int     `json:"asns"`
	IsLessSpec   bool      `json:"is_less_specific"`
	ActualPrefix string    `json:"actual_prefix"`
	QueryTime    time.Time `json:"query_time"`
}

// AnnouncedPrefixes represents all prefixes announced by an ASN
type AnnouncedPrefixes struct {
	ASN      int      `json:"asn"`
	Prefixes []Prefix `json:"prefixes"`
}

// Prefix represents a BGP prefix
type Prefix struct {
	Prefix   string `json:"prefix"`
	Timelines []Timeline `json:"timelines,omitempty"`
}

// Timeline represents prefix announcement timeline
type Timeline struct {
	StartTime time.Time `json:"starttime"`
	EndTime   time.Time `json:"endtime"`
}

// ASPath represents an AS path
type ASPath struct {
	Path      []int     `json:"path"`
	Origin    string    `json:"origin"`
	Prefix    string    `json:"prefix"`
	Peer      string    `json:"peer"`
	Timestamp time.Time `json:"timestamp"`
}

// LookingGlassResult represents a looking glass query result
type LookingGlassResult struct {
	RRCs    []RRCResult `json:"rrcs"`
	Prefix  string      `json:"prefix"`
	ASN     int         `json:"asn,omitempty"`
}

// RRCResult represents results from a specific RRC
type RRCResult struct {
	RRC     string   `json:"rrc"`
	Peers   int      `json:"peers"`
	ASPaths []ASPath `json:"as_paths"`
}

// ASNInfo represents general ASN information
type ASNInfo struct {
	ASN         int      `json:"asn"`
	Holder      string   `json:"holder"`
	Announced   bool     `json:"announced"`
	Block       string   `json:"block"`
	Description string   `json:"description"`
	Country     string   `json:"country"`
}

// RIPEResponse is the generic wrapper for RIPE API responses
type RIPEResponse struct {
	Status       string      `json:"status"`
	StatusCode   int         `json:"status_code"`
	Version      string      `json:"version"`
	Data         interface{} `json:"data"`
	Messages     [][]string  `json:"messages,omitempty"`
	SeeAlso      interface{} `json:"see_also,omitempty"`
	Time         string      `json:"time"`
	QueryID      string      `json:"query_id"`
	ProcessTime  int         `json:"process_time"`
	ServerID     string      `json:"server_id"`
	BuildVersion string      `json:"build_version"`
	Cached       bool        `json:"cached,omitempty"`
}
