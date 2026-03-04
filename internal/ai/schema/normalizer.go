package schema

import (
	"encoding/json"
	"time"
)

type BGPDataType string

const (
	BGPDataTypePrefixAnalysis BGPDataType = "bgp_prefix_analysis"
	BGPDataTypeASNAnalysis    BGPDataType = "bgp_asn_analysis"
	BGPDataTypeRouteAnalysis  BGPDataType = "bgp_route_analysis"
	BGPDataTypeNeighborStatus BGPDataType = "bgp_neighbor_status"
	BGPDataTypeRPKIValidation BGPDataType = "bgp_rpki_validation"
)

type BGPData struct {
	Type       BGPDataType    `json:"type"`
	Timestamp  time.Time      `json:"timestamp"`
	Source     string         `json:"source"`
	ASN        int            `json:"asn,omitempty"`
	Prefix     string         `json:"prefix,omitempty"`
	Attributes BGPAttributes  `json:"attributes,omitempty"`
	RPKIStatus string         `json:"rpki_status,omitempty"`
	Anomalies  []AnomalyInfo  `json:"anomalies,omitempty"`
	Neighbors  []NeighborInfo `json:"neighbors,omitempty"`
	Routes     []RouteInfo    `json:"routes,omitempty"`
}

type BGPAttributes struct {
	ASPath      []int    `json:"as_path"`
	LocalPref   int      `json:"local_pref"`
	MED         int      `json:"med"`
	NextHop     string   `json:"next_hop"`
	Origin      string   `json:"origin"`
	Communities []string `json:"communities"`
	Best        bool     `json:"best"`
}

type AnomalyInfo struct {
	Type     string `json:"type"`
	Severity string `json:"severity"`
	Message  string `json:"message"`
}

type NeighborInfo struct {
	IP           string `json:"ip"`
	ASN          int    `json:"asn"`
	State        string `json:"state"`
	UpTime       string `json:"uptime"`
	MessagesSent uint64 `json:"messages_sent"`
	MessagesRecv uint64 `json:"messages_recv"`
}

type RouteInfo struct {
	Prefix      string   `json:"prefix"`
	ASPath      []int    `json:"as_path"`
	NextHop     string   `json:"next_hop"`
	LocalPref   int      `json:"local_pref"`
	MED         int      `json:"med"`
	Origin      string   `json:"origin"`
	Communities []string `json:"communities"`
	Best        bool     `json:"best"`
}

func NormalizePrefixAnalysis(prefix string, asn int, asPath []int, communities []string, rpkiStatus string) *BGPData {
	return &BGPData{
		Type:      BGPDataTypePrefixAnalysis,
		Timestamp: time.Now(),
		Source:    "RIPE RIS",
		ASN:       asn,
		Prefix:    prefix,
		Attributes: BGPAttributes{
			ASPath:      asPath,
			Communities: communities,
		},
		RPKIStatus: rpkiStatus,
	}
}

func NormalizeRoute(route interface{}) (*BGPData, error) {
	data := &BGPData{
		Type:      BGPDataTypeRouteAnalysis,
		Timestamp: time.Now(),
		Source:    "RIPE RIS",
	}

	return data, nil
}

func (d *BGPData) ToJSON() ([]byte, error) {
	return json.MarshalIndent(d, "", "  ")
}

func (d *BGPData) ToYAML() ([]byte, error) {
	return json.Marshal(d)
}

func (d *BGPData) ToCompactJSON() ([]byte, error) {
	return json.Marshal(d)
}

type LLMPrompt struct {
	SystemPrompt string   `json:"system_prompt"`
	UserPrompt   string   `json:"user_prompt"`
	Data         *BGPData `json:"data"`
}

func (p *LLMPrompt) BuildPrefixAnalysisPrompt(data *BGPData) string {
	systemPrompt := `You are a senior BGP network engineer. Analyze the following BGP route data and provide technical insights.`

	userPrompt := `Analyze the following BGP prefix:

Prefix: ` + data.Prefix + `
Origin ASN: AS` + string(rune(data.ASN)) + `
AS Path: ` + formatASPath(data.Attributes.ASPath) + `
Communities: ` + formatCommunities(data.Attributes.Communities) + `
RPKI Status: ` + data.RPKIStatus + `

Identify:
1. Hijack risk assessment
2. Route leak indicators
3. Path anomalies
4. Optimization suggestions
5. Security concerns

Return a structured technical analysis.`

	p.SystemPrompt = systemPrompt
	p.UserPrompt = userPrompt
	p.Data = data

	return systemPrompt + "\n\n" + userPrompt
}

func formatASPath(asPath []int) string {
	result := ""
	for i, as := range asPath {
		if i > 0 {
			result += " "
		}
		result += "AS" + string(rune(as))
	}
	return result
}

func formatCommunities(communities []string) string {
	if len(communities) == 0 {
		return "None"
	}
	result := ""
	for i, c := range communities {
		if i > 0 {
			result += ", "
		}
		result += c
	}
	return result
}
