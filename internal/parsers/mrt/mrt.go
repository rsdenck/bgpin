package mrt

import (
	"encoding/binary"
	"fmt"
	"io"
	"os"
	"time"

	"github.com/bgpin/bgpin/internal/core/bgp"
)

type MRTParser struct {
	Filename string
}

type MRTHeader struct {
	Timestamp uint32
	Type      uint16
	SubType   uint16
	Length    uint32
}

type MRTRecord struct {
	Timestamp   time.Time
	Type        string
	PeerIP      string
	PeerAS      uint32
	Prefix      string
	ASPath      []uint32
	NextHop     string
	LocalPref   uint32
	MED         uint32
	Origin      string
	Communities []string
	PathAttr    []PathAttribute
}

type PathAttribute struct {
	Type   uint8
	Flags  uint8
	Length uint16
	Value  []byte
}

const (
	TABLE_DUMP_V2 = 13
	TABLE_DUMP    = 12
	BGP4MP        = 16
	BGP4MP_ET     = 17

	TYPE_PEER_INDEX_TABLE = 0
	TYPE_RIB_IPV4_UNICAST = 1
	TYPE_RIB_IPV6_UNICAST = 2
	TYPE_BGP_UPDATE       = 3
	TYPE_BGP_KEEPALIVE    = 4
	TYPE_BGP_NOTIFICATION = 5
)

func NewMRTParser(filename string) *MRTParser {
	return &MRTParser{
		Filename: filename,
	}
}

func (p *MRTParser) Parse() ([]MRTRecord, error) {
	file, err := os.Open(p.Filename)
	if err != nil {
		return nil, fmt.Errorf("failed to open MRT file: %w", err)
	}
	defer file.Close()

	records := make([]MRTRecord, 0)

	for {
		header, err := p.readHeader(file)
		if err == io.EOF {
			break
		}
		if err != nil {
			return records, err
		}

		record, err := p.readRecord(file, header)
		if err != nil {
			continue
		}

		if record != nil {
			records = append(records, *record)
		}
	}

	return records, nil
}

func (p *MRTParser) readHeader(file *os.File) (*MRTHeader, error) {
	header := &MRTHeader{}

	err := binary.Read(file, binary.BigEndian, header)
	if err != nil {
		return nil, err
	}

	return header, nil
}

func (p *MRTParser) readRecord(file *os.File, header *MRTHeader) (*MRTRecord, error) {
	data := make([]byte, header.Length)
	_, err := io.ReadFull(file, data)
	if err != nil {
		return nil, err
	}

	record := &MRTRecord{
		Timestamp: time.Unix(int64(header.Timestamp), 0),
	}

	switch header.Type {
	case TABLE_DUMP_V2:
		record.parseTableDumpV2(header.SubType, data)
	case BGP4MP:
		record.parseBGP4MP(header.SubType, data)
	default:
		return nil, nil
	}

	return record, nil
}

func (r *MRTRecord) parseTableDumpV2(subType uint16, data []byte) {
	switch subType {
	case TYPE_PEER_INDEX_TABLE:
		r.Type = "PEER_INDEX"
	case TYPE_RIB_IPV4_UNICAST:
		r.Type = "RIB_IPV4"
		r.parseIPv4Unicast(data)
	case TYPE_RIB_IPV6_UNICAST:
		r.Type = "RIB_IPV6"
		r.parseIPv6Unicast(data)
	case TYPE_BGP_UPDATE:
		r.Type = "UPDATE"
		r.parseBGPUpdate(data)
	}
}

func (r *MRTRecord) parseIPv4Unicast(data []byte) {
	if len(data) < 9 {
		return
	}

	prefixLen := int(data[0])
	prefixBytes := (prefixLen + 7) / 8

	if len(data) < 1+prefixBytes+2 {
		return
	}

	prefix := make([]byte, 4)
	copy(prefix, data[1:1+prefixBytes])
	r.Prefix = fmt.Sprintf("%d.%d.%d.%d/%d", prefix[0], prefix[1], prefix[2], prefix[3], prefixLen)

	r.parsePathAttributes(data[1+prefixBytes+2:])
}

func (r *MRTRecord) parseIPv6Unicast(data []byte) {
	if len(data) < 17 {
		return
	}

	prefixLen := int(data[0])

	r.Prefix = fmt.Sprintf("ipv6/%d", prefixLen)

	r.parsePathAttributes(data[17:])
}

func (r *MRTRecord) parseBGP4MP(subType uint16, data []byte) {
	r.Type = "BGP4MP"

	switch subType {
	case 0:
		r.Type = "BGP4MP_STATE_CHANGE"
	case 1:
		r.Type = "BGP4MP_MESSAGE"
	case 2:
		r.Type = "BGP4MP_MESSAGE_AS4"
	case 3:
		r.Type = "BGP4MP_STATE_CHANGE_AS4"
	}
}

func (r *MRTRecord) parseBGPUpdate(data []byte) {
	if len(data) < 2 {
		return
	}

	withdrawnLen := int(binary.BigEndian.Uint16(data[0:2]))

	if len(data) < 2+withdrawnLen {
		return
	}

	pathAttrOffset := 2 + withdrawnLen
	r.parsePathAttributes(data[pathAttrOffset:])
}

func (r *MRTRecord) parsePathAttributes(data []byte) {
	if len(data) < 2 {
		return
	}

	offset := 0
	for offset < len(data)-2 {
		if offset+2 > len(data) {
			break
		}

		attrFlags := data[offset]
		attrType := data[offset+1]

		offset += 2

		attrLen := uint16(attrFlags & 0x10)
		if attrLen == 0 {
			if offset+1 > len(data) {
				break
			}
			attrLen = uint16(data[offset])
			offset++
		} else {
			if offset+2 > len(data) {
				break
			}
			attrLen = binary.BigEndian.Uint16(data[offset : offset+2])
			offset += 2
		}

		if offset+int(attrLen) > len(data) {
			break
		}

		attrValue := data[offset : offset+int(attrLen)]
		offset += int(attrLen)

		switch attrType {
		case 1:
			r.Origin = string(attrValue[0])
		case 2:
			r.ASPath = parseASPathAttr(attrValue)
		case 3:
			r.NextHop = parseNextHop(attrValue)
		case 4:
			r.MED = binary.BigEndian.Uint32(attrValue)
		case 5:
			r.LocalPref = binary.BigEndian.Uint32(attrValue)
		case 8:
			r.Communities = parseCommunities(attrValue)
		}
	}
}

func parseASPathAttr(data []byte) []uint32 {
	if len(data) < 2 {
		return nil
	}

	asPath := make([]uint32, 0)
	offset := 1

	numSegs := int(data[0])
	for i := 0; i < numSegs && offset < len(data); i++ {
		if offset+1 >= len(data) {
			break
		}

		segType := data[offset]
		_ = segType
		segLen := int(data[offset+1])
		offset += 2

		for j := 0; j < segLen && offset+3 <= len(data); j++ {
			asn := binary.BigEndian.Uint32([]byte{0, data[offset], data[offset+1], data[offset+2]})
			asPath = append(asPath, asn)
			offset += 4
		}
	}

	return asPath
}

func parseNextHop(data []byte) string {
	if len(data) == 4 {
		return fmt.Sprintf("%d.%d.%d.%d", data[0], data[1], data[2], data[3])
	}
	return ""
}

func parseCommunities(data []byte) []string {
	if len(data) < 4 {
		return nil
	}

	communities := make([]string, 0)
	for i := 0; i+4 <= len(data); i += 4 {
		comm := binary.BigEndian.Uint32(data[i : i+4])
		asn := comm >> 16
		value := comm & 0xFFFF
		communities = append(communities, fmt.Sprintf("%d:%d", asn, value))
	}

	return communities
}

func (p *MRTParser) GetStats() (map[string]interface{}, error) {
	records, err := p.Parse()
	if err != nil {
		return nil, err
	}

	stats := make(map[string]interface{})
	stats["total_records"] = len(records)

	prefixes := make(map[string]int)
	peerAS := make(map[uint32]int)

	for _, r := range records {
		if r.Prefix != "" {
			prefixes[r.Prefix]++
		}
		if r.PeerAS > 0 {
			peerAS[r.PeerAS]++
		}
	}

	stats["unique_prefixes"] = len(prefixes)
	stats["unique_peer_as"] = len(peerAS)

	if len(records) > 0 {
		stats["first_timestamp"] = records[0].Timestamp
		stats["last_timestamp"] = records[len(records)-1].Timestamp
	}

	return stats, nil
}

func (p *MRTParser) ExportToBGPTable() (*bgp.BGPTable, error) {
	records, err := p.Parse()
	if err != nil {
		return nil, err
	}

	table := &bgp.BGPTable{
		Routes:    make([]bgp.Route, 0),
		Timestamp: time.Now(),
	}

	for _, r := range records {
		if r.Prefix == "" {
			continue
		}

		asPath := make([]int, len(r.ASPath))
		for i, as := range r.ASPath {
			asPath[i] = int(as)
		}

		route := bgp.Route{
			Prefix:    r.Prefix,
			ASPath:    asPath,
			NextHop:   r.NextHop,
			LocalPref: int(r.LocalPref),
			MED:       int(r.MED),
			Origin:    r.Origin,
			Community: r.Communities,
		}

		table.Routes = append(table.Routes, route)
	}

	return table, nil
}
