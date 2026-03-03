package bgp

import "time"

type Route struct {
	Prefix     string   `json:"prefix"`
	ASPath     []int    `json:"as_path"`
	NextHop    string   `json:"next_hop"`
	LocalPref  int      `json:"local_pref"`
	MED        int      `json:"med"`
	Valid      bool     `json:"valid"`
	Best       bool     `json:"best"`
	Origin     string   `json:"origin"`
	Community  []string `json:"community,omitempty"`
	AtomicAgg  bool     `json:"atomic_aggregate"`
	Aggregator string   `json:"aggregator,omitempty"`
}

type Neighbor struct {
	IP           string        `json:"ip"`
	AS           int           `json:"asn"`
	State        string        `json:"state"`
	UpTime       time.Duration `json:"uptime"`
	MessagesSent uint64        `json:"messages_sent"`
	MessagesRecv uint64        `json:"messages_recv"`
	LastError    string        `json:"last_error,omitempty"`
}

type BGPTable struct {
	RouterID  string    `json:"router_id"`
	LocalAS   int       `json:"local_as"`
	Vrf       string    `json:"vrf,omitempty"`
	Routes    []Route   `json:"routes"`
	Timestamp time.Time `json:"timestamp"`
}

type LookupResult struct {
	Prefix    string    `json:"prefix"`
	QueryLG   string    `json:"looking_glass"`
	Timestamp time.Time `json:"timestamp"`
	Routes    []Route   `json:"routes"`
	Anomalies []Anomaly `json:"anomalies,omitempty"`
}

type Anomaly struct {
	Type     string `json:"type"`
	Severity string `json:"severity"`
	Message  string `json:"message"`
}

func (r *Route) HasASInPath(asn int) bool {
	for _, as := range r.ASPath {
		if as == asn {
			return true
		}
	}
	return false
}

func (r *Route) ASPathLength() int {
	return len(r.ASPath)
}

func (r *Route) DetectAnomalies() []Anomaly {
	var anomalies []Anomaly

	if len(r.ASPath) > 7 {
		anomalies = append(anomalies, Anomaly{
			Type:     "excessive_prepend",
			Severity: "medium",
			Message:  "AS path length is unusually long",
		})
	}

	for i := 0; i < len(r.ASPath)-1; i++ {
		if r.ASPath[i] == r.ASPath[i+1] {
			anomalies = append(anomalies, Anomaly{
				Type:     "as_path_loop",
				Severity: "high",
				Message:  "Duplicate AS in path detected",
			})
		}
	}

	if r.LocalPref == 0 && r.Best {
		anomalies = append(anomalies, Anomaly{
			Type:     "missing_localpref",
			Severity: "low",
			Message:  "Route has no local preference set",
		})
	}

	return anomalies
}
