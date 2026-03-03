package aspath

import (
	"fmt"
	"regexp"
	"strings"
)

type ASPath struct {
	Segments []Segment
	Original string
}

type Segment struct {
	Type int
	ASNs []int
}

func Parse(input string) (*ASPath, error) {
	input = strings.TrimSpace(input)
	if input == "" || input == "NULL" || input == "incomplete" {
		return &ASPath{Original: input}, nil
	}

	parts := strings.Fields(input)
	var asns []int

	for _, part := range parts {
		part = strings.Trim(part, "{}")
		if part == "" {
			continue
		}
		asn, err := parseAS(part)
		if err != nil {
			return nil, fmt.Errorf("invalid AS number: %s", part)
		}
		asns = append(asns, asn)
	}

	return &ASPath{
		Segments: []Segment{{Type: 2, ASNs: asns}},
		Original: input,
	}, nil
}

func parseAS(s string) (int, error) {
	re := regexp.MustCompile(`^(\d+)(_\d+)*$`)
	s = strings.ReplaceAll(s, "_", "")

	if !re.MatchString(s) {
		return 0, fmt.Errorf("invalid AS format: %s", s)
	}

	var n int
	_, err := fmt.Sscanf(s, "%d", &n)
	return n, err
}

func (p *ASPath) GetASNs() []int {
	if len(p.Segments) == 0 {
		return nil
	}
	return p.Segments[0].ASNs
}

func (p *ASPath) Length() int {
	return len(p.GetASNs())
}

func (p *ASPath) Contains(asn int) bool {
	for _, as := range p.GetASNs() {
		if as == asn {
			return true
		}
	}
	return false
}

func (p *ASPath) HasLoop() bool {
	seen := make(map[int]bool)
	for _, as := range p.GetASNs() {
		if seen[as] {
			return true
		}
		seen[as] = true
	}
	return false
}

func (p *ASPath) OriginAS() int {
	asns := p.GetASNs()
	if len(asns) > 0 {
		return asns[len(asns)-1]
	}
	return 0
}

func (p *ASPath) FirstAS() int {
	asns := p.GetASNs()
	if len(asns) > 0 {
		return asns[0]
	}
	return 0
}
